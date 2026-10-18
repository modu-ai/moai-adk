//go:build !windows

package quality

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// affected_scope_repair_test.go — regression tests for the three under-selection
// defects the sync audit (round 1) found in the affected scope:
//
//	F1  `diff.relative=true` made the changed-file names cwd-relative in a nested module
//	F2  the changed set comes from the index but the package graph from the working tree
//	F3  a file embedded by two packages had one nondeterministic owner

// F1: the changed-set query must not depend on the user's diff.relative setting.
// In a nested module that setting turns names cwd-relative, every file then reads
// as outside the module root and the Go test step would be silently skipped.
func TestAffectedScope_DiffRelativeConfigIgnored(t *testing.T) {
	hermeticGoEnv(t)
	repo := newScopeRepo(t, map[string]string{
		"svc/go.mod":      scopeGoMod,
		"svc/a/a.go":      "package a\n\nfunc A() int { return 1 }\n",
		"svc/a/a_test.go": "package a\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) {}\n",
	})
	gitIn(t, repo, "config", "diff.relative", "true")
	writeScopeFiles(t, repo, map[string]string{"svc/a/a.go": "package a\n\nfunc A() int { return 2 }\n"})
	gitIn(t, repo, "add", "svc/a/a.go")

	svc := filepath.Join(repo, "svc")
	ch, err := queryCommitChanges(scopeCtx(t), svc)
	if err != nil {
		t.Fatalf("queryCommitChanges: %v", err)
	}
	if !slices.Equal(ch.Files, []string{"svc/a/a.go"}) || ch.ModulePrefix != "svc" {
		t.Fatalf("files = %v, prefix = %q, want [svc/a/a.go] and %q (diff.relative=true must not make names cwd-relative)", ch.Files, ch.ModulePrefix, "svc")
	}
	g, err := listModuleGraph(scopeCtx(t), svc, "")
	if err != nil {
		t.Fatalf("listModuleGraph: %v", err)
	}
	res := selectAffected(affectedInput{Changed: ch.Files, ModuleRel: ch.ModulePrefix, Graph: g})
	if res.SkipReason != "" {
		t.Errorf("the Go test step would be skipped: %s", res.SkipReason)
	}
	wantSelected(t, res, "a")
}

// F3: a file embedded by two packages belongs to both; each is touched, and the
// result is the same on every run (map iteration order must not decide it).
func TestAffectedScope_SharedEmbedFileHasAllOwners(t *testing.T) {
	root := scopePkg(".", true)
	root.EmbedFiles = []string{"assets/data.txt"}
	assets := scopePkg("assets", true)
	assets.EmbedFiles = []string{"assets/data.txt"}
	g := scopeGraph(root, assets, scopePkg("x", true, "assets"))

	for i := 0; i < 50; i++ {
		res := selectAffected(affectedInput{Changed: []string{"assets/data.txt"}, Graph: g})
		wantSelected(t, res, ".", "assets", "x")
		if t.Failed() {
			t.Fatalf("iteration %d", i)
		}
	}
}

// goTestRecorder installs a `go` wrapper that delegates every verb except test to
// the real go (the gate's go list query must work) and records only test calls.
func goTestRecorder(t *testing.T) string {
	t.Helper()
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatalf("go not found: %v", err)
	}
	dir := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "go-test.log")
	script := "#!/bin/sh\ncase \"$1\" in\n  test) echo \"$@\" >> '" + logPath + "'; exit 0 ;;\n  *) exec '" + realGo + "' \"$@\" ;;\nesac\n"
	if err := os.WriteFile(filepath.Join(dir, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

// goTestCallsIn runs the pre-commit gate (real adapters) in repo and returns the
// recorded go test calls and the gate output.
func goTestCallsIn(t *testing.T, repo, logPath string) ([]string, string) {
	t.Helper()
	cfg := preCommitGateConfig(repo)
	cfg.DisabledSteps = map[string]bool{"go vet": false}
	passed, out := runGateWith(t, cfg, nil)
	if !passed {
		t.Fatalf("gate failed:\n%s", out)
	}
	return recordedCalls(t, logPath), out
}

// F2: the changed set is read from the index and the package graph from the
// working tree. An unstaged edit that removes b's import of a must not narrow
// the selection to a alone while the committed b still imports a: when a Go file
// differs between index and working tree the gate runs the full module and says
// why. A clean working tree keeps the narrow set.
func TestAffectedScope_UnstagedGoEditFallsBackToFullModule(t *testing.T) {
	hermeticGoEnv(t)
	seed := map[string]string{
		"go.mod":      scopeGoMod,
		"a/a.go":      "package a\n\nfunc A() int { return 1 }\n",
		"a/a_test.go": "package a\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) {}\n",
		"b/b.go":      "package b\n\nimport \"example.com/m/a\"\n\nfunc B() int { return a.A() }\n",
		"b/b_test.go": "package b\n\nimport \"testing\"\n\nfunc TestB(t *testing.T) {}\n",
	}
	stageA := func(t *testing.T, repo string) {
		t.Helper()
		writeScopeFiles(t, repo, map[string]string{"a/a.go": "package a\n\nfunc A() int { return 2 }\n"})
		gitIn(t, repo, "add", "a/a.go")
	}

	t.Run("a clean working tree keeps the narrow set", func(t *testing.T) {
		repo := newScopeRepo(t, seed)
		logPath := goTestRecorder(t)
		stageA(t, repo)
		calls, out := goTestCallsIn(t, repo, logPath)
		if want := []string{"test ./a ./b"}; !slices.Equal(calls, want) {
			t.Errorf("go test calls = %v, want %v\n%s", calls, want, out)
		}
	})

	t.Run("an unstaged edit removing an import widens to the full module", func(t *testing.T) {
		repo := newScopeRepo(t, seed)
		logPath := goTestRecorder(t)
		stageA(t, repo)
		// Unstaged: b no longer imports a in the working tree, but the committed b still does.
		writeScopeFiles(t, repo, map[string]string{"b/b.go": "package b\n\nfunc B() int { return 1 }\n"})
		calls, out := goTestCallsIn(t, repo, logPath)
		if want := []string{"test ./..."}; !slices.Equal(calls, want) {
			t.Errorf("go test calls = %v, want %v (an unstaged Go edit makes the graph unreliable)\n%s", calls, want, out)
		}
		if !strings.Contains(out, "b/b.go") || !strings.Contains(out, "differs from the index") || !strings.Contains(out, "falling back to the full module") {
			t.Errorf("the summary does not name the reason and the fallback:\n%s", out)
		}
	})

	t.Run("an unstaged non-Go edit does not widen", func(t *testing.T) {
		repo := newScopeRepo(t, seed)
		writeScopeFiles(t, repo, map[string]string{"notes.md": "n\n"})
		gitIn(t, repo, "add", "notes.md")
		gitIn(t, repo, "commit", "-q", "-m", "notes")
		logPath := goTestRecorder(t)
		stageA(t, repo)
		writeScopeFiles(t, repo, map[string]string{"notes.md": "edited\n"})
		calls, out := goTestCallsIn(t, repo, logPath)
		if want := []string{"test ./a ./b"}; !slices.Equal(calls, want) {
			t.Errorf("go test calls = %v, want %v\n%s", calls, want, out)
		}
	})
}
