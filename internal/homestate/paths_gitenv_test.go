package homestate_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestCanonicalProjectRootSeparateGitDirIgnoresInheritedGitDir pins the
// `git worktree list` branch, reached only from a linked worktree of a
// --separate-git-dir repository (t1208 audit F1). It asserts invariance — the
// answer under an inherited GIT_DIR equals the answer without one — rather
// than a specific path, because what that branch returns even without GIT_DIR
// is a separate, pre-existing question (t1208 audit F2).
func TestCanonicalProjectRootSeparateGitDirIgnoresInheritedGitDir(t *testing.T) {
	callerRepo, callerWorktree, skip := gitFixtures(t)
	if skip != "" {
		t.Skip(skip)
	}
	out, err := exec.Command("git", "-C", callerWorktree, "rev-parse", "--absolute-git-dir").Output()
	if err != nil {
		t.Fatalf("caller worktree gitdir: %v", err)
	}
	callerWorktreeGitDir := strings.TrimSpace(string(out))

	base := t.TempDir()
	repo, meta, wt := filepath.Join(base, "repo"), filepath.Join(base, "meta"), filepath.Join(base, "wt")
	for _, args := range [][]string{
		{"init", "--initial-branch=main", "--separate-git-dir", meta, repo},
		{"-C", repo, "commit", "--allow-empty", "-m", "seed"},
		{"-C", repo, "worktree", "add", wt, "-b", "linked"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Env = append(cmd.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid")
		if o, err := cmd.CombinedOutput(); err != nil {
			t.Skipf("separate-git-dir fixture: git %v: %v: %s", args, err, o)
		}
	}
	baseline := homestate.CanonicalProjectRoot(wt)
	callerRoot, err := filepath.EvalSymlinks(callerRepo)
	if err != nil {
		t.Fatalf("resolve caller: %v", err)
	}

	for _, withWorkTree := range []bool{false, true} {
		name := "git_dir"
		if withWorkTree {
			name = "git_dir_and_work_tree"
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("GIT_DIR", callerWorktreeGitDir)
			if withWorkTree {
				t.Setenv("GIT_WORK_TREE", callerWorktree)
			}
			got := homestate.CanonicalProjectRoot(wt)
			if got != baseline {
				t.Fatalf("CanonicalProjectRoot(separate-git-dir worktree) = %q under an inherited GIT_DIR, want the uninherited answer %q", got, baseline)
			}
			if got == filepath.Clean(callerRoot) {
				t.Fatalf("answered with the caller's checkout %q", got)
			}
		})
	}
}

// TestCanonicalProjectRootIgnoresInheritedGitDir pins the symptom of t1204
// audit N2 (card t1208): under an inherited GIT_DIR naming another checkout's
// per-worktree gitdir, CanonicalProjectRoot mapped the project to THAT
// checkout's primary root, and runInit then wrote .moai/db/<key>/project.json
// into it. The root must be derived from the directory passed in.
//
// Not parallel: t.Setenv. Fixtures are built before the variable is set.
func TestCanonicalProjectRootIgnoresInheritedGitDir(t *testing.T) {
	_, callerWorktree, skip := gitFixtures(t)
	if skip != "" {
		t.Skip(skip)
	}
	out, err := exec.Command("git", "-C", callerWorktree, "rev-parse", "--absolute-git-dir").Output()
	if err != nil {
		t.Fatalf("caller worktree gitdir: %v", err)
	}
	callerWorktreeGitDir := strings.TrimSpace(string(out))

	victimRepo, victimWorktree, skip := gitFixtures(t)
	if skip != "" {
		t.Skip(skip)
	}
	want, err := filepath.EvalSymlinks(victimRepo)
	if err != nil {
		t.Fatalf("resolve victim: %v", err)
	}

	for _, tc := range []struct {
		name, dir string
		workTree  bool
	}{
		{"victim_primary", victimRepo, false},
		{"victim_linked_worktree", victimWorktree, false},
		// With GIT_WORK_TREE inherited as well, `rev-parse --show-toplevel`
		// (the GitDir == CommonDir branch) no longer lands on -C by accident.
		{"victim_primary_with_work_tree", victimRepo, true},
		{"victim_linked_worktree_with_work_tree", victimWorktree, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GIT_DIR", callerWorktreeGitDir)
			if tc.workTree {
				t.Setenv("GIT_WORK_TREE", callerWorktree)
			}
			if got := homestate.CanonicalProjectRoot(tc.dir); got != filepath.Clean(want) {
				t.Fatalf("CanonicalProjectRoot(%s) = %q under an inherited GIT_DIR, want %q", tc.name, got, want)
			}
		})
	}
}
