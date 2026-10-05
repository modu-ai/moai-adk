package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// M5 acceptance tests (SPEC-MX-TAG-EDGES-001): AC-MTE-013 — the
// --debt-fanin query surface and the stand-in retirement.

// debtFanInFixture builds a project with two DEBT tags — one function-
// anchored (with evidence-backed callers) and one file-scope — then builds
// the edges artifact. Under CGO the anchored tag joins to its symbol and
// the callers give fan-in 2; the file-scope tag self-edges at fan-in 0.
func debtFanInFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, src string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/debt\n\ngo 1.26\n")
	write("internal/pay/owed.go", `package pay

func Owed() {
	// @MX:DEBT: in-memory ledger, no eviction
	// @MX:CEILING: < 10k entries
	// @MX:UPGRADE: swap to a persistent ledger past the ceiling
	_ = map[string]int{}
}
`)
	write("internal/pay/scope.go", `package pay

// @MX:DEBT: file-scope simplification, no upgrade trigger
// @MX:CEILING: < 10k rows

const LedgerName = "ledger"
`)
	write("internal/a/a.go", `package a

import "example.com/debt/internal/pay"

func A() {
	pay.Owed()
}
`)
	write("internal/b/b.go", `package b

import "example.com/debt/internal/pay"

func B() {
	pay.Owed()
}
`)
	write("internal/c/c.go", `package c

func C() {
	Owed()
}
`)

	build := newGraphCmd()
	build.SetArgs([]string{"build", "--root", root})
	out := &strings.Builder{}
	build.SetOut(out)
	build.SetErr(out)
	if err := build.Execute(); err != nil {
		t.Fatalf("graph build: %v\noutput: %s", err, out.String())
	}
	return root
}

// extractionAvailable probes whether the astx layer can run (tree-sitter
// needs CGO): under a !cgo toolchain the anchored DEBT self-edges and the
// ranking asserts only the shape both modes share.
func extractionAvailable(t *testing.T) bool {
	t.Helper()
	probe := filepath.Join(t.TempDir(), "probe.go")
	if err := os.WriteFile(probe, []byte("package probe\n\nfunc P() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	set, err := astx.Extract("go", probe)
	return err == nil && set.Supported
}

// AC-MTE-013 — --debt-fanin ranks mx-debt targets by evidence-backed
// fan-in, descending, ties by target; the file-scope DEBT appears at 0
// with a (self) marker, listed not omitted.
func TestGraphQueryDebtFanIn(t *testing.T) {
	root := debtFanInFixture(t)
	out := runGraphQuery(t, "--root", root, "--debt-fanin")

	// The file-scope DEBT is always present, at fan-in 0, marked.
	if !strings.Contains(out, "0\tinternal/pay/scope.go\t(self)") {
		t.Errorf("output missing the (self) row:\n%s", out)
	}

	if extractionAvailable(t) {
		// Evidence-backed callers (a.go + b.go) outrank the self row; the
		// inferred-only caller (c.go) is not counted toward the blocking
		// fan-in.
		if !strings.Contains(out, "2\tOwed\tinternal/pay/owed.go") {
			t.Errorf("output missing the ranked symbol row (2\tOwed):\n%s", out)
		}
		owedIdx := strings.Index(out, "2\tOwed\tinternal/pay/owed.go")
		selfIdx := strings.Index(out, "0\tinternal/pay/scope.go\t(self)")
		if owedIdx > selfIdx {
			t.Errorf("ranking violated: the fan-in-2 target must precede the self row:\n%s", out)
		}
	}

	// --fanin keeps its import-ranking behavior.
	fanOut := runGraphQuery(t, "--root", root, "--fanin")
	if !strings.Contains(fanOut, "import fan-in") {
		t.Errorf("--fanin output lost its header:\n%s", fanOut)
	}
}

// AC-MTE-013 stand-in retirement, kept firing as a static guard: neither
// the query.go note nor the --fanin help may carry the stand-in phrasing
// or "no tag-kind edges yet" (the AC's grep, pinned as a test).
func TestGraphStandInRetired(t *testing.T) {
	for _, file := range []string{"../graph/query.go", "graph.go"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		src := strings.ToLower(string(data))
		for _, banned := range []string{"stands in for", "stand-in", "no tag-kind edges yet"} {
			if strings.Contains(src, banned) {
				t.Errorf("%s carries the retired stand-in phrasing %q", file, banned)
			}
		}
	}
}

// AC-MTE-010 negative assertion — moai mx scan constructs NO validator,
// before or after this SPEC: its role is sidecar producer feeding the
// freshness probes, never a validation surface (static source scan).
func TestMxScanConstructsNoValidator(t *testing.T) {
	data, err := os.ReadFile("mx_scan.go")
	if err != nil {
		t.Fatalf("read mx_scan.go: %v", err)
	}
	src := string(data)
	for _, banned := range []string{"NewValidator", "ValidateFile", "ValidateFiles"} {
		if strings.Contains(src, banned) {
			t.Errorf("mx_scan.go constructs or drives a validator (%s) — it must stay a sidecar producer", banned)
		}
	}
	if !strings.Contains(src, "NewScanner") {
		t.Error("mx_scan.go lost its scanner construction — the sidecar producer role moved")
	}
}
