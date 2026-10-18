package quality

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// affected_scope_adapters_test.go — the REAL go list and git adapters against
// throwaway modules and repositories (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 M3:
// AC-PAS-005 adapter half, AC-PAS-008 `go list` half). `go` and `git` are
// necessarily present wherever these tests run, so there is no guard.

const scopeGoMod = "module example.com/m\n\ngo 1.23\n"

// hermeticGoEnv keeps the go command off the network and away from any outer
// workspace, so a throwaway module is judged on its own.
func hermeticGoEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
}

// writeScopeFiles writes files (slash paths) under root.
func writeScopeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
}

func scopeCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	return ctx
}

// The go list adapter reports every dependency edge kind, the embed lists in
// module-root-relative form, and does not invent a package out of testdata.
func TestAffectedScope_GoListGraph(t *testing.T) {
	hermeticGoEnv(t)
	root := t.TempDir()
	writeScopeFiles(t, root, map[string]string{
		"go.mod":            scopeGoMod,
		"a/a.go":            "package a\n\nfunc A() int { return 1 }\n",
		"b/b.go":            "package b\n\nimport \"example.com/m/a\"\n\nfunc B() int { return a.A() }\n",
		"b/b_test.go":       "package b\n\nimport \"testing\"\n\nfunc TestB(t *testing.T) {}\n",
		"e/e.go":            "package e\n",
		"e/e_test.go":       "package e\n\nimport (\n\t\"testing\"\n\n\t_ \"example.com/m/a\"\n)\n\nfunc TestE(t *testing.T) {}\n",
		"f/f.go":            "package f\n",
		"f/f_ext_test.go":   "package f_test\n\nimport (\n\t\"testing\"\n\n\t_ \"example.com/m/a\"\n)\n\nfunc TestF(t *testing.T) {}\n",
		"g/g.go":            "package g\n\nimport _ \"embed\"\n\n//go:embed tpl/t.txt\nvar T string\n",
		"g/tpl/t.txt":       "t\n",
		"h/h.go":            "package h\n",
		"h/testdata/x/x.go": "package x\n",
	})

	g, err := listModuleGraph(scopeCtx(t), root, "")
	if err != nil {
		t.Fatalf("listModuleGraph: %v", err)
	}
	if g.ModulePath != scopeTestModule {
		t.Errorf("module path = %q, want %q", g.ModulePath, scopeTestModule)
	}
	pkg := func(dir string) *scopePackage {
		t.Helper()
		p := g.Packages[scopeImportPath(dir)]
		if p == nil {
			t.Fatalf("package %s missing from the graph (have %d packages)", dir, len(g.Packages))
		}
		return p
	}
	if pkg("a").HasTests {
		t.Error("package a has no test files, HasTests = true")
	}
	if !pkg("b").HasTests || !slices.Contains(pkg("b").Imports, scopeImportPath("a")) {
		t.Errorf("package b: HasTests = %v, imports = %v, want tests and an import of a", pkg("b").HasTests, pkg("b").Imports)
	}
	if !slices.Contains(pkg("e").TestImports, scopeImportPath("a")) || slices.Contains(pkg("e").Imports, scopeImportPath("a")) {
		t.Errorf("package e: TestImports = %v, Imports = %v, want a only among the in-package test imports", pkg("e").TestImports, pkg("e").Imports)
	}
	if !slices.Contains(pkg("f").XTestImports, scopeImportPath("a")) || !pkg("f").HasTests {
		t.Errorf("package f: XTestImports = %v, HasTests = %v, want a among the external-test imports", pkg("f").XTestImports, pkg("f").HasTests)
	}
	if !slices.Contains(pkg("g").EmbedFiles, "g/tpl/t.txt") {
		t.Errorf("package g embed files = %v, want the module-root-relative g/tpl/t.txt", pkg("g").EmbedFiles)
	}
	if _, ok := g.Packages[scopeImportPath("h/testdata/x")]; ok {
		t.Error("a testdata directory was reported as a package")
	}
}

// AC-PAS-008 (go list half): a package importing a nonexistent dependency makes
// `go list -e` exit 0 while marking the package incomplete; the adapter must
// not trust the exit status.
func TestAffectedScope_IncompleteGraphFallsBack(t *testing.T) {
	hermeticGoEnv(t)

	good := t.TempDir()
	writeScopeFiles(t, good, map[string]string{
		"go.mod": scopeGoMod,
		"a/a.go": "package a\n",
	})
	if _, err := listModuleGraph(scopeCtx(t), good, ""); err != nil {
		t.Fatalf("control: a clean module was refused: %v", err)
	}

	bad := t.TempDir()
	writeScopeFiles(t, bad, map[string]string{
		"go.mod":     scopeGoMod,
		"a/a.go":     "package a\n",
		"bad/bad.go": "package bad\n\nimport _ \"example.com/nonexistent/dep\"\n",
	})
	_, err := listModuleGraph(scopeCtx(t), bad, "")
	if err == nil {
		t.Fatal("an incomplete package graph was accepted; the adapter trusted go list's zero exit status")
	}
	if !strings.Contains(err.Error(), scopeImportPath("bad")) {
		t.Errorf("error %q does not name the incomplete package %s", err, scopeImportPath("bad"))
	}
}

// newScopeRepo creates a repository holding a Go module and one commit.
func newScopeRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	repo := newRepo(t, filepath.Join(t.TempDir(), "repo"))
	writeScopeFiles(t, repo, files)
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-q", "-m", "module")
	return repo
}

// AC-PAS-005 (adapter half): the REAL git adapter reports both paths of a
// rename and the last file of a deleted package, and the selection built from
// them is right end to end. The core is never handed the old path by the test.
func TestAffectedScope_RealGitRenameDelete(t *testing.T) {
	hermeticGoEnv(t)

	t.Run("cross-package rename keeps the old package's importers", func(t *testing.T) {
		repo := newScopeRepo(t, map[string]string{
			"go.mod":      scopeGoMod,
			"a/x.go":      "package a\n\nfunc X() int { return 1 }\n",
			"a/y.go":      "package a\n\nfunc Y() int { return 2 }\n",
			"b/b.go":      "package b\n\nimport \"example.com/m/a\"\n\nfunc B() int { return a.Y() }\n",
			"b/b_test.go": "package b\n\nimport \"testing\"\n\nfunc TestB(t *testing.T) {}\n",
			"e/e.go":      "package e\n",
			"e/e_test.go": "package e\n\nimport \"testing\"\n\nfunc TestE(t *testing.T) {}\n",
			"a/a_test.go": "package a\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) {}\n",
		})
		gitIn(t, repo, "mv", "a/x.go", "e/x.go")
		writeScopeFiles(t, repo, map[string]string{"e/x.go": "package e\n\nfunc X() int { return 1 }\n"})
		gitIn(t, repo, "add", "e/x.go")

		ch, err := queryCommitChanges(scopeCtx(t), repo)
		if err != nil {
			t.Fatalf("queryCommitChanges: %v", err)
		}
		if !slices.Contains(ch.Files, "a/x.go") || !slices.Contains(ch.Files, "e/x.go") {
			t.Fatalf("changed list %v must carry both the old path a/x.go and the new path e/x.go (rename detection off)", ch.Files)
		}
		g, err := listModuleGraph(scopeCtx(t), repo, "")
		if err != nil {
			t.Fatalf("listModuleGraph: %v", err)
		}
		res := selectAffected(affectedInput{Changed: ch.Files, ModuleRel: ch.ModulePrefix, Graph: g})
		wantSelected(t, res, "a", "b", "e")
	})

	t.Run("deleting the last file of a package selects its updated importers", func(t *testing.T) {
		repo := newScopeRepo(t, map[string]string{
			"go.mod":      scopeGoMod,
			"old/o.go":    "package old\n\nfunc O() int { return 1 }\n",
			"u/u.go":      "package u\n\nimport \"example.com/m/old\"\n\nfunc U() int { return old.O() }\n",
			"u/u_test.go": "package u\n\nimport \"testing\"\n\nfunc TestU(t *testing.T) {}\n",
			"w/w.go":      "package w\n\nimport \"example.com/m/u\"\n\nfunc W() int { return u.U() }\n",
			"w/w_test.go": "package w\n\nimport \"testing\"\n\nfunc TestW(t *testing.T) {}\n",
		})
		gitIn(t, repo, "rm", "-q", "old/o.go")
		writeScopeFiles(t, repo, map[string]string{"u/u.go": "package u\n\nfunc U() int { return 1 }\n"})
		gitIn(t, repo, "add", "u/u.go")

		ch, err := queryCommitChanges(scopeCtx(t), repo)
		if err != nil {
			t.Fatalf("queryCommitChanges: %v", err)
		}
		if !slices.Contains(ch.Files, "old/o.go") {
			t.Fatalf("changed list %v lacks the deleted path old/o.go", ch.Files)
		}
		g, err := listModuleGraph(scopeCtx(t), repo, "")
		if err != nil {
			t.Fatalf("listModuleGraph: %v", err)
		}
		if _, ok := g.Packages[scopeImportPath("old")]; ok {
			t.Fatal("the deleted package is still in the graph")
		}
		res := selectAffected(affectedInput{Changed: ch.Files, ModuleRel: ch.ModulePrefix, Graph: g})
		wantSelected(t, res, "u", "w")
	})

	t.Run("a nested module root reports its offset inside the repository", func(t *testing.T) {
		repo := newScopeRepo(t, map[string]string{
			"m/go.mod": scopeGoMod,
			"m/a/a.go": "package a\n",
		})
		writeScopeFiles(t, repo, map[string]string{"m/a/a.go": "package a\n\nvar X = 1\n"})
		gitIn(t, repo, "add", "m/a/a.go")

		ch, err := queryCommitChanges(scopeCtx(t), filepath.Join(repo, "m"))
		if err != nil {
			t.Fatalf("queryCommitChanges: %v", err)
		}
		if ch.ModulePrefix != "m" || !slices.Equal(ch.Files, []string{"m/a/a.go"}) {
			t.Errorf("files = %v, prefix = %q, want [m/a/a.go] and %q", ch.Files, ch.ModulePrefix, "m")
		}
	})

	t.Run("nothing staged is an empty list, not an error", func(t *testing.T) {
		repo := newScopeRepo(t, map[string]string{"go.mod": scopeGoMod, "a/a.go": "package a\n"})
		ch, err := queryCommitChanges(scopeCtx(t), repo)
		if err != nil || len(ch.Files) != 0 {
			t.Errorf("files = %v, err = %v, want an empty list and no error", ch.Files, err)
		}
	})

	t.Run("a directory outside any repository is a query failure", func(t *testing.T) {
		if _, err := queryCommitChanges(scopeCtx(t), t.TempDir()); err == nil {
			t.Error("a non-repository directory returned no error; the caller could not fall back")
		}
	})
}
