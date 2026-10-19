package cli

import (
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// t1710 — the card-creation step runs its name-uniqueness probe and its
// `git branch -m` on the inherited environment, so an exported GIT_DIR /
// GIT_WORK_TREE points BOTH at the caller's repository: the rename then moves
// the WRONG repository's branch (the GH #1691 write shape) while the new card
// worktree keeps its template name. Both git children of the step must read
// the worktree they were handed.
func TestFactoryCreateAndRenameWorktreeIgnoresInheritedGitEnv(t *testing.T) {
	root := t1710GitRepo(t)
	wt := filepath.Join(t.TempDir(), "wt")
	if out, err := exec.Command("git", "-C", root, "worktree", "add", "-q", "-b", "card-base", wt).CombinedOutput(); err != nil {
		t.Fatalf("git worktree add: %v\n%s", err, out)
	}
	other := t1710GitRepo(t)
	otherBefore := t1710BranchOf(t, other)

	orig := worktree.WorktreeCreator
	worktree.WorktreeCreator = func(cardID string, out io.Writer) (string, error) { return wt, nil }
	t.Cleanup(func() { worktree.WorktreeCreator = orig })

	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	got, err := factoryCreateAndRenameWorktree(root, "t1710", "env-leak", io.Discard)
	if err != nil {
		t.Fatalf("factoryCreateAndRenameWorktree error = %v", err)
	}
	if got != wt {
		t.Fatalf("created worktree = %q, want the stubbed %q", got, wt)
	}
	if branch := t1710BranchOf(t, wt); branch != "WT-env-leak" {
		t.Errorf("worktree branch = %q, want the renamed WT-env-leak (the inherited GIT_DIR redirected the rename)", branch)
	}
	if branch := t1710BranchOf(t, other); branch != otherBefore {
		t.Errorf("repository %s branch moved from %q to %q — the rename wrote into the caller's repository", other, otherBefore, branch)
	}
}

// t1710GitRepo makes a one-commit repository in a temp directory (the t1589
// fixture shape, local to package cli).
func t1710GitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"-c", "user.name=test", "-c", "user.email=test@example.invalid", "commit", "-q", "--allow-empty", "--no-gpg-sign", "-m", "init"},
	} {
		full := append([]string{"-C", dir}, args...)
		if out, err := exec.Command("git", full...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

// t1710BranchOf reads dir's HEAD branch with the git environment scrubbed, so
// the assertion itself cannot be redirected by the variables the test
// exported.
func t1710BranchOf(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "symbolic-ref", "--quiet", "--short", "HEAD")
	cmd.Env = gitenv.Env()
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return "" // detached HEAD
		}
		t.Fatalf("branch of %s: %v", dir, err)
	}
	return strings.TrimSpace(string(out))
}
