package factory

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestLandedRefGitRunIgnoresCallerRepoScopingEnv (card t2009, S11): the level-2
// landed-ref read asks git about a named directory. GIT_DIR and GIT_WORK_TREE
// outrank `git -C`, so the default runner must hand git the gitenv-scrubbed
// environment or it answers about the caller's repository.
func TestLandedRefGitRunIgnoresCallerRepoScopingEnv(t *testing.T) {
	initRepo := func() string {
		dir := t.TempDir()
		cmd := exec.Command("git", "init", "-q", "-b", "main")
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init: %v: %s", err, out)
		}
		return dir
	}
	target := initRepo()
	other := initRepo()

	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	out, err := landedRefGitRun("git", "-C", target, "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatalf("rev-parse under foreign GIT_DIR: %v", err)
	}
	got, err := filepath.EvalSymlinks(strings.TrimSpace(out))
	if err != nil {
		t.Fatalf("resolve answer %q: %v", out, err)
	}
	want, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatalf("resolve target: %v", err)
	}
	if got != want {
		t.Errorf("S11_WRONG_REPO: landedRefGitRun answered about %s, want the target %s", got, want)
	}
}
