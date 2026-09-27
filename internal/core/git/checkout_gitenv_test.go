// checkout_gitenv_test.go — ResolveGitDirs must answer about the directory it
// is given, not about the repository an inherited GIT_DIR names (card t1208,
// t1204 audit N2).
//
// git exports GIT_DIR into every hook it runs; a `moai` process started under a
// hook inherits it. `git -C dir rev-parse` still obeys GIT_DIR over -C, so an
// unscrubbed child reports the caller's repository. The per-worktree gitdir is
// the shape that did damage: CanonicalProjectRoot mapped it to the caller's
// primary checkout and runInit wrote .moai/db/<key>/project.json there.
//
// These tests use t.Setenv and are therefore not parallel. Fixtures are built
// BEFORE the variable is set: the runGit helper inherits the environment.
package git

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// forceFallback makes the first probe fail as an older git would, through the
// ExecCommand indirection, so the fallback probes are the ones that answer.
func forceFallback(t *testing.T) {
	t.Helper()
	orig := ExecCommand
	t.Cleanup(func() { ExecCommand = orig })
	ExecCommand = func(name string, args ...string) *exec.Cmd {
		if strings.Contains(strings.Join(args, " "), "--path-format=absolute") {
			return exec.Command("sh", "-c", "echo 'unknown option: path-format=absolute' >&2; exit 1")
		}
		return orig(name, args...)
	}
}

func TestResolveGitDirs_IgnoresInheritedGitDir(t *testing.T) {
	callerPrimary, callerWorktree := initTestRepoWithWorktree(t)
	callerWorktreeGitDir := runGit(t, callerWorktree, "rev-parse", "--absolute-git-dir")
	victim := initTestRepo(t)
	wantVictim := filepath.Join(victim, ".git")

	cases := []struct {
		name   string
		gitDir string
	}{
		// The shape the N2 reproduction used: a commit made from a linked
		// worktree exports that worktree's own gitdir.
		{"caller_linked_worktree_gitdir", callerWorktreeGitDir},
		// A commit from a primary checkout exports the shared gitdir.
		{"caller_primary_gitdir", filepath.Join(callerPrimary, ".git")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GIT_DIR", tc.gitDir)
			got, err := ResolveGitDirs(victim)
			if err != nil {
				t.Fatalf("ResolveGitDirs(victim) under GIT_DIR=%s: %v", tc.gitDir, err)
			}
			if !samePath(got.GitDir, wantVictim) || !samePath(got.CommonDir, wantVictim) {
				t.Fatalf("ResolveGitDirs(victim) answered about the caller's repository:\n GitDir    = %q\n CommonDir = %q\n want both = %q",
					got.GitDir, got.CommonDir, wantVictim)
			}
		})
	}
}

// The older-git fallback reaches git through the same helper and must be
// scrubbed too; forcing it keeps the fix from covering only the first probe.
func TestResolveGitDirs_FallbackIgnoresInheritedGitDir(t *testing.T) {
	if isWindowsRuntime() {
		t.Skip("fallback mock uses sh -c; skip on windows")
	}
	_, callerWorktree := initTestRepoWithWorktree(t)
	callerWorktreeGitDir := runGit(t, callerWorktree, "rev-parse", "--absolute-git-dir")
	victim := initTestRepo(t)
	wantVictim := filepath.Join(victim, ".git")

	forceFallback(t)
	t.Setenv("GIT_DIR", callerWorktreeGitDir)
	got, err := ResolveGitDirs(victim)
	if err != nil {
		t.Fatalf("fallback ResolveGitDirs(victim): %v", err)
	}
	if !samePath(got.GitDir, wantVictim) || !samePath(got.CommonDir, wantVictim) {
		t.Fatalf("fallback answered about the caller's repository: GitDir=%q CommonDir=%q want %q",
			got.GitDir, got.CommonDir, wantVictim)
	}
}
