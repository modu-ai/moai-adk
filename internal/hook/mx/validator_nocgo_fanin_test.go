package mx

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AC-MTE-014 (SPEC-MX-TAG-EDGES-001) — under a CGO-disabled build the
// edges artifact carries zero code-call evidence, so the graph-backed
// source degrades with an error and the validator falls back to the
// textual index, labeled "textual-fallback", surfacing no error
// (REQ-MTE-015). The behavior is build-mode independent, so this test is
// untagged: it runs in every build, including the AC's CGO_ENABLED=0
// command, and pins the degrade contract against a source that errors.
func TestNoCGOFanInTextualFallback(t *testing.T) {
	root := t.TempDir()
	srcFile := filepath.Join(root, "degraded.go")

	var b strings.Builder
	b.WriteString("package app\n\n")
	b.WriteString("// DegradedFunc is mentioned for the textual index.\n")
	b.WriteString("func DegradedFunc() {}\n")
	b.WriteString("func use() {\n\t_ = DegradedFunc\n\t_ = DegradedFunc\n\t_ = DegradedFunc\n\t_ = DegradedFunc\n}\n")
	if err := os.WriteFile(srcFile, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	// The degrade shape: the artifact has no code evidence to serve.
	src := &fakeSource{err: errors.New("graph: artifact carries no code-call evidence")}
	v := NewValidatorWithSource(nil, root, src)
	report, err := v.ValidateFile(context.Background(), srcFile)
	if err != nil {
		t.Fatalf("ValidateFile must not surface the degrade: %v", err)
	}

	found := false
	for _, viol := range report.Violations {
		if viol.FuncName == "DegradedFunc" && viol.Priority == P1 {
			found = true
			if viol.FanIn < 3 {
				t.Errorf("fallback fan-in = %d, want the textual count >= 3", viol.FanIn)
			}
			if viol.Source != "textual-fallback" {
				t.Errorf("degraded verdict Source = %q, want textual-fallback", viol.Source)
			}
		}
	}
	if !found {
		t.Fatal("no P1 verdict for DegradedFunc — the degrade path did not answer")
	}
}
