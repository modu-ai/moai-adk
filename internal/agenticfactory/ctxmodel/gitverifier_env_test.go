package ctxmodel

// Card t2136 (follow-up of t2126): GitVerifier.Lookup ran `git cat-file` with
// the caller's environment, so an inherited GIT_DIR (git hook context) made it
// look the blob up in another repository. Not parallel: t.Setenv.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func lookupFixtureGit(t *testing.T, dir string, args ...string) string {
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

// Two real repositories: the verifier's clone (clone, which does not hold the
// commit) and an unrelated one (other, which does). With the unrelated
// repository's GIT_DIR inherited, the lookup must still read the verifier's own
// clone and report the commit as not found.
func TestGitVerifierLookupIgnoresInheritedGitDir(t *testing.T) {
	root := t.TempDir()
	clone := filepath.Join(root, "clone")
	other := filepath.Join(root, "other")
	for _, p := range []string{clone, other} {
		lookupFixtureGit(t, root, "init", "-q", p)
	}
	if err := os.WriteFile(filepath.Join(other, "f.md"), []byte("only in other"), 0o644); err != nil {
		t.Fatal(err)
	}
	lookupFixtureGit(t, other, "add", "f.md")
	lookupFixtureGit(t, other, "commit", "-q", "-m", "other")
	commit := lookupFixtureGit(t, other, "rev-parse", "HEAD")

	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	sum, found, err := GitVerifier{Dir: clone}.Lookup(commit, "f.md")
	if err != nil {
		t.Fatalf("Lookup err = %v, want nil (not found)", err)
	}
	if found {
		t.Errorf("Lookup found the blob (sha256 %s) in another repository through an inherited GIT_DIR", sum)
	}
}
