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
	calls, imports, _, _, err := Extract(root)
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

// Module normalization: a go.mod-bearing tree localizes internal imports and
// marks them Local; external imports and absent-module trees pass through.
func TestExtract_ModuleNormalization(t *testing.T) {
	root := seamFixture(t)
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/proj\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "wire", "imp.go"),
		[]byte("package wire\n\nimport (\n\t\"example.com/proj/internal/other\"\n\t\"fmt\"\n)\n\nfunc D() { other.X(); fmt.Println(1) }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal", "other"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "other", "other.go"), []byte("package other\n\nfunc X() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, imports, _, _, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	sawLocal, sawExternal := false, false
	for _, imp := range imports {
		if imp.File != "internal/wire/imp.go" {
			continue
		}
		switch imp.Module {
		case "internal/other":
			if !imp.Local {
				t.Error("module-prefixed import must be Local")
			}
			sawLocal = true
		case "fmt":
			if imp.Local {
				t.Error("stdlib import must not be Local")
			}
			sawExternal = true
		}
	}
	if !sawLocal || !sawExternal {
		t.Errorf("localization split not observed: local=%v external=%v imports=%v", sawLocal, sawExternal, imports)
	}

	// localizeModule unit split: prefix absent → pass-through, not local.
	if m, isLocal := localizeModule("github.com/x/y", ""); isLocal || m != "github.com/x/y" {
		t.Errorf("empty module prefix must pass through: %q %v", m, isLocal)
	}
	if _, isLocal := localizeModule("example.com/proj/internal/z", "example.com/proj"); !isLocal {
		t.Error("prefixed module must localize")
	}
	if m, isLocal := localizeModule("example.com/projish", "example.com/proj"); isLocal || m != "example.com/projish" {
		t.Errorf("partial-prefix must not localize: %q %v", m, isLocal)
	}
}

// REQ-GEC-002..004 — the walk retains each file's declared function/method
// names (sorted, deduplicated), so the mapper can join callee names to
// declaring files without a second parse pass.
func TestExtract_RetainsDeclaredNames(t *testing.T) {
	root := seamFixture(t)
	_, _, decls, _, err := Extract(root)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(decls) != 1 {
		t.Fatalf("decls = %+v, want one record for wire.go", decls)
	}
	if decls[0].File != "internal/wire/wire.go" {
		t.Errorf("decls[0].File = %q, want internal/wire/wire.go", decls[0].File)
	}
	want := []string{"A", "B", "C"}
	if len(decls[0].Names) != len(want) {
		t.Fatalf("names = %v, want %v", decls[0].Names, want)
	}
	for i, name := range want {
		if decls[0].Names[i] != name {
			t.Errorf("names[%d] = %q, want %q (sorted)", i, decls[0].Names[i], name)
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
