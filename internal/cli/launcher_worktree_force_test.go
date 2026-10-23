package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runGitOrSkip executes a git command in dir, skipping the test when git is
// unavailable and failing it when the command itself errors.
func runGitOrSkip(t *testing.T, dir string, args ...string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// newRepoWithWorkerWorktree creates a throwaway git repo with one committed
// file and a worker-* worktree under .claude/worktrees/, returning the repo
// root and the worktree path.
func newRepoWithWorkerWorktree(t *testing.T) (string, string) {
	t.Helper()

	root := t.TempDir()
	runGitOrSkip(t, root, "init", "-b", "main")
	runGitOrSkip(t, root, "config", "user.email", "test@example.com")
	runGitOrSkip(t, root, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(root, "seed.txt"), []byte("seed\n"), 0o600); err != nil {
		t.Fatalf("write seed file: %v", err)
	}
	runGitOrSkip(t, root, "add", "seed.txt")
	runGitOrSkip(t, root, "commit", "-m", "seed")

	// SPEC-WORKTREE-HOME-ROOT-001 M3 (REQ-WHR-014): the cleanup guard
	// removes a worker tree only when its branch is integrated and landed
	// on the configured integration ref, so the fixture configures the
	// target and lands main on a remote — the worker branch sits at the
	// same seed commit and is therefore in the removable state.
	cfgDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	cfg := "git_strategy:\n  mode: manual\n  manual:\n    workflow: git-flow\n    develop_branch: main\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "git-strategy.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
	remote := filepath.Join(t.TempDir(), "origin.git")
	runGitOrSkip(t, root, "init", "-q", "--bare", remote)
	runGitOrSkip(t, root, "remote", "add", "origin", remote)
	runGitOrSkip(t, root, "push", "-q", "-u", "origin", "main")

	wtPath := filepath.Join(root, ".claude", "worktrees", "worker-SPEC-TEST-001")
	runGitOrSkip(t, root, "worktree", "add", "-b", "worker-branch", wtPath)

	return root, wtPath
}

// TestCleanupMoaiWorktrees_KeepsDirtyWorktree is the falsifiable guard for the
// launch-path removal contract: `moai cc` runs cleanupMoaiWorktrees on every
// launch, so a worktree holding uncommitted work must survive it. Re-adding
// --force to removeWorktree makes this test fail.
func TestCleanupMoaiWorktrees_KeepsDirtyWorktree(t *testing.T) {
	root, wtPath := newRepoWithWorkerWorktree(t)

	// Leave uncommitted work in the worktree: one modified tracked file and
	// one untracked file. git refuses to remove either without --force.
	if err := os.WriteFile(filepath.Join(wtPath, "seed.txt"), []byte("edited\n"), 0o600); err != nil {
		t.Fatalf("modify tracked file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wtPath, "scratch.txt"), []byte("wip\n"), 0o600); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}

	msg := cleanupMoaiWorktrees(root)

	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("worktree with uncommitted work was deleted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wtPath, "scratch.txt")); err != nil {
		t.Fatalf("untracked file in worktree was deleted: %v", err)
	}
	if !strings.Contains(msg, "Kept 1 worktree(s) with local changes") {
		t.Errorf("expected the kept-worktree report, got: %q", msg)
	}
	if !strings.Contains(msg, "worker-SPEC-TEST-001") {
		t.Errorf("expected the report to name the kept worktree, got: %q", msg)
	}
}

// TestCleanupMoaiWorktrees_RemovesCleanWorktree verifies the launch-path cleanup
// still does its job on a worktree that holds nothing to lose — the non-force
// change must not turn cleanup into a no-op.
func TestCleanupMoaiWorktrees_RemovesCleanWorktree(t *testing.T) {
	root, wtPath := newRepoWithWorkerWorktree(t)

	msg := cleanupMoaiWorktrees(root)

	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Fatalf("clean worktree should have been removed, stat err = %v", err)
	}
	if !strings.Contains(msg, "Cleaned up 1 worktree(s)") {
		t.Errorf("expected the cleanup report, got: %q", msg)
	}
}
