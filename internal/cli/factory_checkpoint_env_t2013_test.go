package cli

// factory_checkpoint_env_t2013_test.go — card t2013, card-review round 2 P2 #2:
// the checkpoint's git queries must describe the card worktree even when the
// process environment carries GIT_DIR / GIT_WORK_TREE pointing at another
// checkout (an inherited GIT_DIR outranks the query's working directory).

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFactoryCheckpointIgnoresInheritedGitEnvironment(t *testing.T) {
	f := newCheckpointFixture(t, "merge-ready", "merge-ready")
	wantHead := cpGit(t, f.wt, "rev-parse", "HEAD")

	other := t.TempDir()
	cpGit(t, other, "init", "-b", "other-branch")
	cpWrite(t, filepath.Join(other, "o.txt"), "other\n")
	cpGit(t, other, "add", "o.txt")
	cpGit(t, other, "commit", "-m", "other base")
	cpWrite(t, filepath.Join(other, "dirty1.txt"), "x\n")
	cpWrite(t, filepath.Join(other, "dirty2.txt"), "x\n")
	otherHead := cpGit(t, other, "rev-parse", "HEAD")
	if otherHead == wantHead {
		t.Fatalf("fixture error: both repositories share HEAD %s", wantHead)
	}

	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(other, ".git", "index"))

	out, err := runCheckpointCmd(t, cpCard)
	if err != nil {
		t.Fatalf("checkpoint: %v\n%s", err, out)
	}
	cp := parseCheckpoint(out)
	if cp.Head != wantHead || cp.Branch != "main" || cp.Dirty != 0 {
		t.Errorf("checkpoint under an inherited GIT_DIR = head %q branch %q dirty %d; want the card worktree's head %q, branch main, dirty 0 (other repo: head %q):\n%s",
			cp.Head, cp.Branch, cp.Dirty, wantHead, otherHead, out)
	}
	if _, err := os.Stat(f.path()); err != nil {
		t.Errorf("the checkpoint must be written in the card worktree: %v", err)
	}
}
