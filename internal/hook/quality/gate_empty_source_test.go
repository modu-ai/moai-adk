package quality

// gate_empty_source_test.go — the source-free scaffold regression.
//
// A project that had declared Python (a pyproject.toml) but not yet written
// any .py failed its first gate: `mypy .` prints "There are no .py[i] files in
// directory" and exits 2. The pytest step next to it already treats an empty
// suite as a pass, so the same project was told its absent tests were fine and
// its absent code was a failure.
//
// The steps below never invoke mypy: an optional step whose binary is missing
// returns early at the LookPath check, which on a machine without mypy would
// make every assertion here pass vacuously. The behavioural tests use a step
// that is guaranteed to run and guaranteed to fail, so a skip and a run are
// distinguishable regardless of what is installed.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// alwaysFailingStep returns a step that runs a command certain to exist in a
// Go test environment and certain to exit non-zero. If the step executes, the
// gate reports failure; if the guard skips it, the gate reports success.
func alwaysFailingStep(name string, sourceExts []string) gateStep {
	return gateStep{
		name:       name,
		binary:     "go",
		args:       []string{"run", "./__t193_no_such_package__"},
		sourceExts: sourceExts,
	}
}

// TestExecuteStep_SkipsWhenProjectHasNoSources is the regression: a scaffold
// carrying only a pyproject.toml must not fail the gate for having no code.
func TestExecuteStep_SkipsWhenProjectHasNoSources(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[project]\nname = \"scaffold\"\n")

	g := &QualityGate{config: &GateConfig{Enabled: true, ProjectDir: dir}}
	ok, msg := g.executeStep(context.Background(), alwaysFailingStep("mypy", []string{".py", ".pyi"}), 30*time.Second)

	if !ok {
		t.Fatalf("source-free scaffold failed the gate: %s", msg)
	}
}

// TestExecuteStep_RunsWhenProjectHasOneSource is the other half: one .py is
// enough to make the step meaningful again, so the guard must not swallow it.
// Without this, "always skip" would pass the test above.
func TestExecuteStep_RunsWhenProjectHasOneSource(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[project]\nname = \"scaffold\"\n")
	writeFile(t, dir, "src/app.py", "x = 1\n")

	g := &QualityGate{config: &GateConfig{Enabled: true, ProjectDir: dir}}
	ok, _ := g.executeStep(context.Background(), alwaysFailingStep("mypy", []string{".py", ".pyi"}), 30*time.Second)

	if ok {
		t.Fatal("step was skipped although the project has a .py source; the guard over-skips")
	}
}

// TestExecuteStep_NoSourceExtsUnaffected asserts the guard is opt-in: a step
// that declares no sourceExts runs exactly as before.
func TestExecuteStep_NoSourceExtsUnaffected(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[project]\nname = \"scaffold\"\n")

	g := &QualityGate{config: &GateConfig{Enabled: true, ProjectDir: dir}}
	ok, _ := g.executeStep(context.Background(), alwaysFailingStep("other", nil), 30*time.Second)

	if ok {
		t.Fatal("a step with no sourceExts was skipped; the guard is not opt-in")
	}
}

// TestExecuteStep_PyiCountsAsSource asserts a stubs-only package is not
// treated as source-free — mypy has plenty to check there.
func TestExecuteStep_PyiCountsAsSource(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[project]\nname = \"stubs\"\n")
	writeFile(t, dir, "stubs/app.pyi", "def f() -> int: ...\n")

	g := &QualityGate{config: &GateConfig{Enabled: true, ProjectDir: dir}}
	if ok, _ := g.executeStep(context.Background(), alwaysFailingStep("mypy", []string{".py", ".pyi"}), 30*time.Second); ok {
		t.Fatal("a .pyi-only project was treated as source-free")
	}
}

// TestProjectHasSourceFile_IgnoresDependencyTrees asserts a .py belonging to a
// dependency does not count as the project's own source. Counting it would
// send mypy back into the failure this guard removes.
func TestProjectHasSourceFile_IgnoresDependencyTrees(t *testing.T) {
	for _, vendorDir := range []string{".venv/lib/python3.12/site-packages", "node_modules/pkg", ".git", "__pycache__"} {
		t.Run(vendorDir, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "pyproject.toml", "[project]\nname = \"scaffold\"\n")
			writeFile(t, dir, filepath.Join(vendorDir, "dep.py"), "x = 1\n")

			found, determined := projectHasSourceFile(dir, []string{".py", ".pyi"})
			if !determined {
				t.Fatal("scan reported undetermined on a tiny tree")
			}
			if found {
				t.Errorf("a .py under %s counted as the project's own source", vendorDir)
			}
		})
	}
}

// TestProjectHasSourceFile_FindsNestedSource asserts the scan descends past
// the root — a src/ layout is the common case, not the exception.
func TestProjectHasSourceFile_FindsNestedSource(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "src/pkg/deep/mod.py", "x = 1\n")

	found, determined := projectHasSourceFile(dir, []string{".py", ".pyi"})
	if !determined || !found {
		t.Fatalf("nested source not found: found=%v determined=%v", found, determined)
	}
}

// TestProjectHasSourceFile_CaseInsensitiveExtension asserts an uppercase
// extension still counts — the staged-file check next to it matches case
// insensitively too.
func TestProjectHasSourceFile_CaseInsensitiveExtension(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "MOD.PY", "x = 1\n")

	if found, determined := projectHasSourceFile(dir, []string{".py"}); !determined || !found {
		t.Fatalf("uppercase extension not matched: found=%v determined=%v", found, determined)
	}
}

// TestProjectHasSourceFile_UndeterminedInputs asserts the scan reports
// undetermined rather than "no sources" when it cannot answer. Undetermined
// runs the step, so an unanswerable scan never suppresses a check.
func TestProjectHasSourceFile_UndeterminedInputs(t *testing.T) {
	cases := map[string]struct {
		dir  string
		exts []string
	}{
		"empty dir":     {dir: "", exts: []string{".py"}},
		"no extensions": {dir: t.TempDir(), exts: nil},
		"missing dir":   {dir: filepath.Join(t.TempDir(), "does-not-exist"), exts: []string{".py"}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if found, determined := projectHasSourceFile(tc.dir, tc.exts); determined || found {
				t.Errorf("found=%v determined=%v, want both false", found, determined)
			}
		})
	}
}

// TestProjectHasSourceFile_EmptyTreeIsDetermined asserts a readable tree with
// no matching source answers "no sources" conclusively — that answer is what
// the skip depends on.
func TestProjectHasSourceFile_EmptyTreeIsDetermined(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", "[project]\nname = \"scaffold\"\n")
	writeFile(t, dir, "README.md", "# scaffold\n")

	found, determined := projectHasSourceFile(dir, []string{".py", ".pyi"})
	if !determined {
		t.Fatal("a readable source-free tree reported undetermined")
	}
	if found {
		t.Error("found a .py in a tree that has none")
	}
}

// TestMypyStep_DeclaresSourceExts pins the wiring: the guard is only reachable
// because the mypy step declares the extensions.
func TestMypyStep_DeclaresSourceExts(t *testing.T) {
	tc := pythonToolchain(t)
	var mypy *gateStep
	for i := range tc.lintSteps {
		if tc.lintSteps[i].name == "mypy" {
			mypy = &tc.lintSteps[i]
		}
	}
	if mypy == nil {
		t.Fatal("mypy step not found on the Python toolchain")
	}
	want := map[string]bool{".py": false, ".pyi": false}
	for _, ext := range mypy.sourceExts {
		if _, ok := want[ext]; ok {
			want[ext] = true
		}
	}
	for ext, seen := range want {
		if !seen {
			t.Errorf("mypy sourceExts = %v, missing %q", mypy.sourceExts, ext)
		}
	}
}

// TestRuffStep_DeclaresNoSourceExts records why ruff has no counterpart:
// `ruff check .` on a source-free tree reports "All checks passed!" and exits
// 0, so it needs no guard. If ruff's behaviour ever changes, this test is the
// place that says the decision was deliberate.
func TestRuffStep_DeclaresNoSourceExts(t *testing.T) {
	tc := pythonToolchain(t)
	for i := range tc.lintSteps {
		if tc.lintSteps[i].name == "ruff" && len(tc.lintSteps[i].sourceExts) != 0 {
			t.Errorf("ruff declares sourceExts = %v; it exits 0 on an empty tree and needs no guard",
				tc.lintSteps[i].sourceExts)
		}
	}
}

// TestSourceScanSkipDirs_CoversTheCommonTrees guards the skip list against
// accidental emptying.
func TestSourceScanSkipDirs_CoversTheCommonTrees(t *testing.T) {
	for _, name := range []string{".git", "node_modules", ".venv", "site-packages", "__pycache__"} {
		if _, ok := sourceScanSkipDirs[name]; !ok {
			t.Errorf("sourceScanSkipDirs is missing %q", name)
		}
	}
}

// TestProjectHasSourceFile_RootIsNotSkippedByName asserts a project that
// happens to live in a directory named like a skipped one (a checkout under
// ~/build, say) is still scanned — the skip applies to descendants only.
func TestProjectHasSourceFile_RootIsNotSkippedByName(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "build")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	writeFile(t, dir, "app.py", "x = 1\n")

	if found, determined := projectHasSourceFile(dir, []string{".py"}); !determined || !found {
		t.Fatalf("root named \"build\" was skipped: found=%v determined=%v", found, determined)
	}
}
