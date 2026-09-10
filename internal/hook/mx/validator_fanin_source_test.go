package mx

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// M3/M4 acceptance tests (SPEC-MX-TAG-EDGES-001): the fan-in evidence
// source seam. AC-MTE-010 (constructor default = textual), REQ-MTE-013
// (inferred-only tail in the Reason), AC-MTE-011 (labeled fallback).

// fakeSource is a controllable FanInEvidenceSource.
type fakeSource struct {
	evidence int
	inferred int
	label    string
	err      error
	calls    int
	lastFunc string
	lastFile string
}

func (f *fakeSource) EvidenceBacked(_ context.Context, funcName, currentFile string) (int, int, string, error) {
	f.calls++
	f.lastFunc = funcName
	f.lastFile = currentFile
	if f.err != nil {
		return 0, 0, "", f.err
	}
	return f.evidence, f.inferred, f.label, nil
}

// AC-MTE-010 (hook/mx half) — the default constructor keeps the TEXTUAL
// source (nil field): the PostToolUse cost profile is unchanged by
// construction; NewValidatorWithSource injects the batch source.
func TestConstructorDefaultsTextualSource(t *testing.T) {
	root := t.TempDir()
	v := NewValidator(nil, root).(*mxValidator)
	if v.fanInSource != nil {
		t.Fatalf("default constructor fanInSource = %v, want nil (textual by construction)", v.fanInSource)
	}

	src := &fakeSource{evidence: 5, label: "edges"}
	v2 := NewValidatorWithSource(nil, root, src).(*mxValidator)
	if v2.fanInSource == nil {
		t.Fatal("NewValidatorWithSource did not inject the source")
	}
}

// REQ-MTE-013 — an injected source's answer shapes the Reason with the
// inferred-only tail, and the violation carries the source's label.
func TestP1ReasonCarriesInferredTailAndLabel(t *testing.T) {
	root := t.TempDir()
	srcFile := filepath.Join(root, "svc.go")
	src := `package app

import "context"

// @MX:WARN: [AUTO] context pattern present
// @MX:REASON: goroutine lifecycle
func Run(ctx context.Context) {
	go func() { _ = ctx }()
}

// CalledOften has no ANCHOR and a high evidence count.
func CalledOften() {}
`
	if err := os.WriteFile(srcFile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	srcFake := &fakeSource{evidence: 4, inferred: 2, label: "edges"}
	v := NewValidatorWithSource(nil, root, srcFake)
	report, err := v.ValidateFile(context.Background(), srcFile)
	if err != nil {
		t.Fatalf("ValidateFile: %v", err)
	}

	found := false
	for _, viol := range report.Violations {
		if viol.FuncName == "CalledOften" && viol.Priority == P1 {
			found = true
			if viol.FanIn != 4 {
				t.Errorf("FanIn = %d, want the source's evidence count 4", viol.FanIn)
			}
			if !strings.Contains(viol.Reason, "fan_in(graph)=4 evidence-backed") {
				t.Errorf("Reason = %q, want the fan_in(graph)= shape", viol.Reason)
			}
			if !strings.Contains(viol.Reason, "(+2 inferred-only)") {
				t.Errorf("Reason = %q, want the (+2 inferred-only) tail", viol.Reason)
			}
			if viol.Source != "edges" {
				t.Errorf("Source = %q, want the injected source's label", viol.Source)
			}
		}
	}
	if !found {
		t.Fatal("no P1 violation for CalledOften — the source seam did not fire")
	}
	if srcFake.lastFunc != "CalledOften" {
		t.Errorf("source was asked for %q, want CalledOften", srcFake.lastFunc)
	}
}

// AC-MTE-011 — when the injected source cannot answer, the validator falls
// back to the textual index, labels the verdict "textual-fallback", and
// surfaces no error (stale/absent artifacts degrade, never fail).
func TestFanInFallbackLabeled(t *testing.T) {
	root := t.TempDir()
	srcFile := filepath.Join(root, "hot.go")

	// The textual index must exceed the threshold on its own: mention the
	// name in enough places across the file.
	var b strings.Builder
	b.WriteString("package app\n\n")
	b.WriteString("// HotFunc is mentioned here for the textual index.\n")
	b.WriteString("func HotFunc() {}\n")
	b.WriteString("func use() {\n\t_ = HotFunc\n\t_ = HotFunc\n\t_ = HotFunc\n\t_ = HotFunc\n}\n")
	if err := os.WriteFile(srcFile, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	// The zero-value textual baseline: no injected source.
	plain := NewValidator(nil, root)
	plainReport, err := plain.ValidateFile(context.Background(), srcFile)
	if err != nil {
		t.Fatalf("textual ValidateFile: %v", err)
	}

	// Erroring source (absent artifact / CGO-off degrade shape).
	srcFake := &fakeSource{err: errors.New("artifact absent")}
	v := NewValidatorWithSource(nil, root, srcFake)
	report, err := v.ValidateFile(context.Background(), srcFile)
	if err != nil {
		t.Fatalf("ValidateFile with a failing source must not surface the error: %v", err)
	}

	var plainCount, fallbackCount int
	for _, viol := range plainReport.Violations {
		if viol.FuncName == "HotFunc" && viol.Priority == P1 {
			plainCount = viol.FanIn
			if viol.Source != "" {
				t.Errorf("plain textual P1 Source = %q, want empty", viol.Source)
			}
		}
	}
	for _, viol := range report.Violations {
		if viol.FuncName == "HotFunc" && viol.Priority == P1 {
			fallbackCount = viol.FanIn
			if viol.Source != "textual-fallback" {
				t.Errorf("fallback P1 Source = %q, want %q", viol.Source, "textual-fallback")
			}
		}
	}
	if plainCount < 3 {
		t.Fatalf("fixture textual fan-in = %d, want >= 3 so the fallback path is observable", plainCount)
	}
	if fallbackCount != plainCount {
		t.Errorf("fallback fan-in = %d, want the textual count %d", fallbackCount, plainCount)
	}
	if srcFake.calls == 0 {
		t.Error("the erroring source was never consulted — fallback fired without trying")
	}
}

// Acceptance §D.3 layering lock — internal/hook/mx carries neither
// internal/graph nor internal/cli in its transitive dependency set, verified
// by `go list -deps` on this package (mirrors AC-GF-016's pattern). The
// fan-in evidence source is structurally compatible, never imported.
func TestHookMxLayeringLock(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Skipf("go list unavailable in test env: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "github.com/modu-ai/moai-adk/internal/graph") ||
			strings.HasPrefix(line, "github.com/modu-ai/moai-adk/internal/cli") {
			t.Errorf("layering lock violated — internal/hook/mx depends on %s", line)
		}
	}
}
