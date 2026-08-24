package symbol

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/navigator/astx"
)

// seamFixture: A calls B, B calls C — an undocumented chain.
func seamFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "wire"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package wire

func A() {
	B()
}

func B() {
	C()
}

func C() {}
`
	if err := os.WriteFile(filepath.Join(root, "internal", "wire", "wire.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// The seam extracts the A→B→C chain with callers joined by containment.
func TestExtract_JoinsCallersByContainment(t *testing.T) {
	root := seamFixture(t)
	calls, imports, _, err := Extract(root)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	pair := map[string]string{} // caller→callee
	for _, c := range calls {
		pair[c.Caller] = c.Callee
	}
	if pair["A"] != "B" || pair["B"] != "C" {
		t.Errorf("call chain wrong: %v", pair)
	}
	if len(imports) != 0 {
		t.Errorf("fixture has no imports, got %v", imports)
	}
}

// AC-GF-016 — this package (the graph builder's extraction seam, a
// non-navigator astx consumer) carries NO navigator-tier package in its
// transitive dependency set, verified by `go list -deps` on this package —
// the acceptance-specified method.
func TestSeamCarriesNoNavigatorTierDeps(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Skipf("go list unavailable in test env: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "github.com/modu-ai/moai-adk/internal/navigator/astx" {
			continue // astx itself is the sanctioned seam dep
		}
		if strings.HasPrefix(line, "github.com/modu-ai/moai-adk/internal/navigator/") {
			t.Errorf("navigator-tier dependency leaked into the extraction seam: %s", line)
		}
	}
}

// Matrix validation on the seam's own output.
func TestMatrix_NoEmptyCells(t *testing.T) {
	m := GradeMatrix()
	if len(m) != len(astx.SupportedLanguages()) {
		t.Fatalf("matrix size %d != languages %d", len(m), len(astx.SupportedLanguages()))
	}
	if d := ValidateGradeMatrix(m); len(d) != 0 {
		t.Errorf("own matrix must validate clean: %v", d)
	}
	broken := map[string]string{}
	for k, v := range m {
		broken[k] = v
	}
	delete(broken, "scala")
	d := ValidateGradeMatrix(broken)
	if len(d) == 0 || !strings.Contains(d[0], "scala") {
		t.Errorf("gradeless cell must defect naming the language: %v", d)
	}
}
