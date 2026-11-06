package checkpoint

// Card t2126 (PR #1969 Codex P2): GitRemote ran git with the caller's
// environment, so an inherited GIT_DIR made `git -C <Dir>` read another
// repository's origin and ConfirmRemote could set RemoteACK from the wrong
// remote. Not parallel: t.Setenv.

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureGit runs git in dir with a scrubbed-by-TestMain environment and a
// fixed identity.
func fixtureGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=fixture@example.invalid", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// Two real repositories: the checkpoint's repository (work, whose origin does
// NOT hold the branch) and an unrelated one (other, whose origin holds the
// checkpoint commit). With the unrelated repository's GIT_DIR inherited, the
// ACK must still come from the checkpoint repository's own origin.
func TestConfirmRemoteIgnoresInheritedGitDir(t *testing.T) {
	root := t.TempDir()
	bareWork := filepath.Join(root, "work-remote.git")
	bareOther := filepath.Join(root, "other-remote.git")
	work := filepath.Join(root, "work")
	other := filepath.Join(root, "other")
	for _, p := range []string{bareWork, bareOther} {
		fixtureGit(t, root, "init", "-q", "--bare", p)
	}
	for _, p := range []string{work, other} {
		fixtureGit(t, root, "init", "-q", p)
	}
	fixtureGit(t, work, "remote", "add", "origin", bareWork)
	fixtureGit(t, other, "remote", "add", "origin", bareOther)

	if err := os.WriteFile(filepath.Join(other, "f"), []byte("f"), 0o644); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, other, "add", "f")
	fixtureGit(t, other, "commit", "-q", "-m", "other")
	sha := fixtureGit(t, other, "rev-parse", "HEAD")
	fixtureGit(t, other, "push", "-q", "origin", "HEAD:refs/heads/WT-x")

	// The checkpoint repository never pushed: the honest answer is not-acked.
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	r := &Record{PushSent: true}
	r.Branch, r.SHA = "WT-x", sha
	err := r.ConfirmRemote(GitRemote{Dir: work, Name: "origin"})
	if !errors.Is(err, ErrNotAcked) {
		t.Fatalf("ConfirmRemote err = %v, want ErrNotAcked (the checkpoint repository's origin lacks the branch)", err)
	}
	if r.RemoteACK {
		t.Error("RemoteACK set from another repository's origin via an inherited GIT_DIR")
	}
}
