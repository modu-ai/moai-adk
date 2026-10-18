//go:build !windows

package quality

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// affected_scope_untracked_test.go — sync-audit round 2, finding R2-1.
//
// `git diff --name-only` never lists untracked files. An untracked go.mod in a
// dependent package directory makes that directory a separate module in the
// working tree, so the working-tree package graph drops it and a staged change
// to the package it imports selects only the changed package. Untracked module
// files (go.mod, go.work, go.work.sum) therefore join the working-tree-versus-
// index check that sends the Go test step to the full module.
//
// Judged from the code, not guessed:
//   - go.sum carries checksums only; it adds and removes no package and no
//     import edge, so an untracked go.sum does not change the selection.
//   - an untracked plain .go file can only ADD an edge or a package to the
//     working-tree graph (over-selection), and a conflicting package clause makes
//     go list mark the package errored, which is the existing graph fallback.
//     It cannot remove a package, so it does not widen.

func TestAffectedScope_UntrackedModuleFileFallsBackToFullModule(t *testing.T) {
	hermeticGoEnv(t)
	seed := map[string]string{
		"go.mod":      scopeGoMod,
		"a/a.go":      "package a\n\nfunc A() int { return 1 }\n",
		"a/a_test.go": "package a\n\nimport \"testing\"\n\nfunc TestA(t *testing.T) {}\n",
		"b/b.go":      "package b\n\nimport \"example.com/m/a\"\n\nfunc B() int { return a.A() }\n",
		"b/b_test.go": "package b\n\nimport \"testing\"\n\nfunc TestB(t *testing.T) {}\n",
	}
	run := func(t *testing.T, untracked map[string]string) ([]string, string) {
		t.Helper()
		repo := newScopeRepo(t, seed)
		logPath := goTestRecorder(t)
		writeScopeFiles(t, repo, map[string]string{"a/a.go": "package a\n\nfunc A() int { return 2 }\n"})
		gitIn(t, repo, "add", "a/a.go")
		writeScopeFiles(t, repo, untracked)
		return goTestCallsIn(t, repo, logPath)
	}

	for _, tc := range []struct {
		name string
		file string
		body string
	}{
		{"an untracked nested go.mod", "b/go.mod", "module example.com/m/b\n\ngo 1.23\n"},
		{"an untracked go.work", "go.work", "go 1.23\n\nuse .\n"},
		{"an untracked go.work.sum", "go.work.sum", "\n"},
	} {
		t.Run(tc.name+" widens to the full module", func(t *testing.T) {
			calls, out := run(t, map[string]string{tc.file: tc.body})
			if want := []string{"test ./..."}; !slices.Equal(calls, want) {
				t.Errorf("go test calls = %v, want %v\n%s", calls, want, out)
			}
			if !strings.Contains(out, tc.file) || !strings.Contains(out, "untracked") || !strings.Contains(out, "falling back to the full module") {
				t.Errorf("the summary does not name the untracked file and the fallback:\n%s", out)
			}
		})
	}

	for _, tc := range []struct{ name, file, body string }{
		{"an untracked text file", "notes.txt", "n\n"},
		{"an untracked plain Go file", "b/extra.go", "package b\n\nfunc Extra() int { return 1 }\n"},
		{"an untracked go.sum", "go.sum", "\n"},
	} {
		t.Run(tc.name+" keeps the narrow set", func(t *testing.T) {
			calls, out := run(t, map[string]string{tc.file: tc.body})
			if want := []string{"test ./a ./b"}; !slices.Equal(calls, want) {
				t.Errorf("go test calls = %v, want %v\n%s", calls, want, out)
			}
		})
	}
}

// The untracked-module-file query must work under the hook's temporary index
// (`git commit -a`) and from a linked worktree, like the other queries.
func TestAffectedScope_UntrackedModuleFileThroughHook(t *testing.T) {
	seed := map[string]string{"a/x.go": "1\n", "d/z.go": "1\n"}

	shape := func(t *testing.T, dir, out, unstaged string) {
		t.Helper()
		writeScopeFiles(t, dir, map[string]string{"a/x.go": "2\n", "n/go.mod": "module example.com/m/n\n", "n/ok.txt": "x\n"})
		gitIn(t, dir, "add", "a/x.go")
		t.Setenv(scopeHelperUnstagedEnv, unstaged)
		gitIn(t, dir, "commit", "-q", "-a", "-m", "with an untracked module file")
		if got := recordedSet(t, out); got != "a/x.go" {
			t.Errorf("changed set = %q, want %q", got, "a/x.go")
		}
		data, err := os.ReadFile(unstaged)
		if err != nil {
			t.Fatalf("the hook recorded no working-tree list: %v", err)
		}
		if got := strings.TrimSpace(string(data)); got != "n/go.mod" {
			t.Errorf("working-tree-versus-index list = %q, want %q (the untracked go.mod, and no other file)", got, "n/go.mod")
		}
	}

	t.Run("plain repository under the temporary index of commit -a", func(t *testing.T) {
		repo := newScopeRepo(t, seed)
		dir := t.TempDir()
		out := filepath.Join(dir, "recorded.txt")
		installScopeHook(t, repo, out)
		shape(t, repo, out, filepath.Join(dir, "unstaged.txt"))
	})
	t.Run("linked worktree", func(t *testing.T) {
		repo := newScopeRepo(t, seed)
		dir := t.TempDir()
		out := filepath.Join(dir, "recorded.txt")
		installScopeHook(t, repo, out)
		wt := filepath.Join(dir, "linked")
		gitIn(t, repo, "worktree", "add", "-q", wt, "-b", "linked-br")
		shape(t, wt, out, filepath.Join(dir, "unstaged.txt"))
	})
}
