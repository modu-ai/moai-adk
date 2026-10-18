package quality

import (
	"slices"
	"strings"
	"testing"
)

// affected_scope_test.go — table tests of the pure affected-set core
// (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 M2: AC-PAS-001 .. -007, core halves).
// Every graph is a hand-written in-memory value; nothing here runs go, git or
// reads the clock.

const scopeTestModule = "example.com/m"

// scopeImportPath returns the import path of a module-root-relative directory.
func scopeImportPath(dir string) string {
	if dir == "." || dir == "" {
		return scopeTestModule
	}
	return scopeTestModule + "/" + dir
}

// scopePkg builds a package at dir. imports are directories, converted to
// import paths.
func scopePkg(dir string, hasTests bool, imports ...string) *scopePackage {
	p := &scopePackage{ImportPath: scopeImportPath(dir), HasTests: hasTests}
	for _, d := range imports {
		p.Imports = append(p.Imports, scopeImportPath(d))
	}
	return p
}

// scopeGraph assembles a graph from packages.
func scopeGraph(pkgs ...*scopePackage) moduleGraph {
	g := moduleGraph{ModulePath: scopeTestModule, Packages: map[string]*scopePackage{}}
	for _, p := range pkgs {
		g.Packages[p.ImportPath] = p
	}
	return g
}

// baseScopeGraph is the AC-PAS-001 graph: a; b importing a; c importing b; d
// standalone, all with test files.
func baseScopeGraph() moduleGraph {
	return scopeGraph(
		scopePkg("a", true),
		scopePkg("b", true, "a"),
		scopePkg("c", true, "b"),
		scopePkg("d", true),
	)
}

// wantSelected asserts the selection equals the given directories exactly.
func wantSelected(t *testing.T, got affectedResult, dirs ...string) {
	t.Helper()
	want := make([]string, 0, len(dirs))
	for _, d := range dirs {
		want = append(want, scopeImportPath(d))
	}
	slices.Sort(want)
	if got.Full {
		t.Fatalf("selection is the full module (trigger %q), want exactly %v", got.Trigger, want)
	}
	if !slices.Equal(got.Selected, want) {
		t.Errorf("selected = %v, want exactly %v", got.Selected, want)
	}
}

// AC-PAS-001: direct dependents only.
func TestAffectedScope_DirectDependents(t *testing.T) {
	res := selectAffected(affectedInput{Changed: []string{"a/x.go"}, Graph: baseScopeGraph()})
	wantSelected(t, res, "a", "b")
	if want := []string{"./a", "./b"}; !slices.Equal(res.Run, want) {
		t.Errorf("run patterns = %v, want %v (module-root-relative directory patterns)", res.Run, want)
	}
}

// AC-PAS-002: an import from an in-package test or an external test package is
// a dependency edge.
func TestAffectedScope_TestImportEdges(t *testing.T) {
	e := scopePkg("e", true)
	e.TestImports = []string{scopeImportPath("a")}
	f := scopePkg("f", true)
	f.XTestImports = []string{scopeImportPath("a")}
	g := baseScopeGraph()
	g.Packages[e.ImportPath] = e
	g.Packages[f.ImportPath] = f

	res := selectAffected(affectedInput{Changed: []string{"a/x.go"}, Graph: g})
	wantSelected(t, res, "a", "b", "e", "f")
}

// AC-PAS-003: ownership and widening per file kind.
func TestAffectedScope_Ownership(t *testing.T) {
	q := scopePkg("q", true)
	q.EmbedFiles = []string{"q/tpl/t.txt"}
	g := baseScopeGraph()
	for _, p := range []*scopePackage{
		scopePkg("p", true),
		scopePkg("pp", true, "p"),
		q,
		scopePkg("r", true, "q"),
	} {
		g.Packages[p.ImportPath] = p
	}

	res := selectAffected(affectedInput{
		Changed: []string{"p/testdata/in.json", "p/testdata/fix.go", "q/tpl/t.txt", "docs/note.md"},
		Graph:   g,
	})
	wantSelected(t, res, "p", "q", "r")
	if want := []string{"docs/note.md"}; !slices.Equal(res.Unowned, want) {
		t.Errorf("unowned = %v, want %v", res.Unowned, want)
	}
	if !slices.ContainsFunc(res.Notes, func(n string) bool {
		return strings.Contains(n, "docs/note.md") && strings.Contains(n, "no package")
	}) {
		t.Errorf("summary does not name docs/note.md as owned by no package:\n%s", strings.Join(res.Notes, "\n"))
	}
}

// AC-PAS-004: a package whose changed files are all test-only is selected
// alone.
func TestAffectedScope_TestOnlyNarrows(t *testing.T) {
	res := selectAffected(affectedInput{
		Changed: []string{"a/x_test.go", "a/testdata/y.json", "a/testdata/f.go"},
		Graph:   baseScopeGraph(),
	})
	wantSelected(t, res, "a")
}

// AC-PAS-005 (core half): a deleted file and a rename attribute to the former
// directory, and importers are found by import path whether or not the former
// package still exists.
func TestAffectedScope_DeletedAndRenamed(t *testing.T) {
	t.Run("deleted last file of a package", func(t *testing.T) {
		g := baseScopeGraph()
		delete(g.Packages, scopeImportPath("a")) // a retains no files
		res := selectAffected(affectedInput{Changed: []string{"a/x.go"}, Graph: g})
		wantSelected(t, res, "b")
	})
	t.Run("rename inside a package", func(t *testing.T) {
		res := selectAffected(affectedInput{Changed: []string{"a/x.go", "a/y.go"}, Graph: baseScopeGraph()})
		wantSelected(t, res, "a", "b")
	})
	t.Run("rename across packages reports both directories", func(t *testing.T) {
		g := baseScopeGraph()
		g.Packages[scopeImportPath("e")] = scopePkg("e", true)
		res := selectAffected(affectedInput{Changed: []string{"a/x.go", "e/x.go"}, Graph: g})
		wantSelected(t, res, "a", "b", "e")
	})
}

// AC-PAS-006 (core half): dropped packages are counted and named; an empty
// selection carries a skip reason.
func TestAffectedScope_DropAndEmpty(t *testing.T) {
	t.Run("a selection with one untested package", func(t *testing.T) {
		g := scopeGraph(scopePkg("a", true), scopePkg("n", false, "a"))
		res := selectAffected(affectedInput{Changed: []string{"a/x.go"}, Graph: g})
		wantSelected(t, res, "a", "n")
		if want := []string{scopeImportPath("n")}; !slices.Equal(res.Dropped, want) {
			t.Errorf("dropped = %v, want %v", res.Dropped, want)
		}
		if want := []string{"./a"}; !slices.Equal(res.Run, want) {
			t.Errorf("run = %v, want %v", res.Run, want)
		}
		notes := strings.Join(res.Notes, "\n")
		if !strings.Contains(notes, "selected 2, dropped 1") {
			t.Errorf("summary lacks %q:\n%s", "selected 2, dropped 1", notes)
		}
		if !strings.Contains(notes, scopeImportPath("n")+": no test files") {
			t.Errorf("summary does not name the dropped package with the reason %q:\n%s", "no test files", notes)
		}
	})
	t.Run("a changed list holding only a docs file", func(t *testing.T) {
		res := selectAffected(affectedInput{Changed: []string{"docs/note.md"}, Graph: baseScopeGraph()})
		if len(res.Run) != 0 || res.Full {
			t.Fatalf("run = %v, full = %v, want an empty non-full selection", res.Run, res.Full)
		}
		if res.SkipReason == "" {
			t.Error("an empty selection carries no skip reason; it must be an explicit skip, not a pass")
		}
	})
	t.Run("every selected package is untested", func(t *testing.T) {
		g := scopeGraph(scopePkg("a", false))
		res := selectAffected(affectedInput{Changed: []string{"a/x.go"}, Graph: g})
		if len(res.Run) != 0 || res.SkipReason == "" {
			t.Errorf("run = %v, skip reason = %q, want an explicit skip", res.Run, res.SkipReason)
		}
	})
	t.Run("no changed list at all", func(t *testing.T) {
		res := selectAffected(affectedInput{Graph: baseScopeGraph()})
		if len(res.Run) != 0 || res.SkipReason == "" {
			t.Errorf("run = %v, skip reason = %q, want an explicit skip", res.Run, res.SkipReason)
		}
	})
}

// AC-PAS-007 (core half): dependency-graph files widen to the full module, and
// the check runs before files outside the module root are discarded.
func TestAffectedScope_WideningTriggers(t *testing.T) {
	flat := []string{"go.mod", "go.sum", "go.work", "go.work.sum", ".moai/config/build-tags"}
	for _, trigger := range flat {
		t.Run("flat "+trigger, func(t *testing.T) {
			res := selectAffected(affectedInput{Changed: []string{"a/x.go", trigger}, Graph: baseScopeGraph()})
			if !res.Full || res.Trigger != trigger {
				t.Fatalf("full = %v, trigger = %q, want the full module triggered by %q", res.Full, res.Trigger, trigger)
			}
			if !slices.ContainsFunc(res.Notes, func(n string) bool { return strings.Contains(n, trigger) }) {
				t.Errorf("summary does not name the trigger %q:\n%s", trigger, strings.Join(res.Notes, "\n"))
			}
		})
	}
	for _, trigger := range []string{"go.work", ".moai/config/build-tags"} {
		t.Run("nested module, trigger at the repository root: "+trigger, func(t *testing.T) {
			res := selectAffected(affectedInput{
				Changed:   []string{trigger, "m/a/x.go"},
				ModuleRel: "m",
				Graph:     baseScopeGraph(),
			})
			if !res.Full || res.Trigger != trigger {
				t.Fatalf("full = %v, trigger = %q, want the full module triggered by %q (a module-root filter ran first)", res.Full, res.Trigger, trigger)
			}
		})
	}
	t.Run("nested module without a trigger narrows", func(t *testing.T) {
		res := selectAffected(affectedInput{
			Changed:   []string{"m/a/x.go", "other/y.go"},
			ModuleRel: "m",
			Graph:     baseScopeGraph(),
		})
		wantSelected(t, res, "a", "b")
	})
	t.Run("a nested go.mod anywhere widens", func(t *testing.T) {
		res := selectAffected(affectedInput{Changed: []string{"tools/go.mod"}, Graph: baseScopeGraph()})
		if !res.Full || res.Trigger != "tools/go.mod" {
			t.Fatalf("full = %v, trigger = %q, want the full module triggered by tools/go.mod", res.Full, res.Trigger)
		}
	})
}

// A package at the module root maps to the "." pattern, and non-Go files fall
// back to the nearest ancestor package directory.
func TestAffectedScope_RootPackageAndAncestorFallback(t *testing.T) {
	g := scopeGraph(scopePkg(".", true), scopePkg("a", true, "."))
	res := selectAffected(affectedInput{Changed: []string{"README.md"}, Graph: g})
	wantSelected(t, res, ".", "a")
	if want := []string{".", "./a"}; !slices.Equal(res.Run, want) {
		t.Errorf("run patterns = %v, want %v", res.Run, want)
	}

	g2 := scopeGraph(scopePkg("a", true), scopePkg("b", true, "a"))
	res2 := selectAffected(affectedInput{Changed: []string{"a/sub/data.bin"}, Graph: g2})
	wantSelected(t, res2, "a", "b")
}
