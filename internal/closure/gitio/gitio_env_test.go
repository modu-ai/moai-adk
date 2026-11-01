package gitio

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunIgnoresCallerRepoScopingEnv (card t2009, S11): git environment
// variables that locate a repository (GIT_DIR, GIT_WORK_TREE) outrank the
// working directory, so a runner that sets only cmd.Dir answers about the
// caller's repository instead of the one it was asked about. The runner must
// run git with the gitenv-scrubbed environment, like the other git runners.
func TestRunIgnoresCallerRepoScopingEnv(t *testing.T) {
	target := newRepo(t)
	other := newRepo(t)

	// Give the other repository a HEAD that differs from the target's, so a
	// wrong-repository answer cannot coincide with the right one.
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", "other")
	cmd.Dir = other
	cmd.Env = append(environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("other commit: %v: %s", err, out)
	}

	want, err := Head(target)
	if err != nil {
		t.Fatalf("target head: %v", err)
	}
	otherHead, err := Head(other)
	if err != nil {
		t.Fatalf("other head: %v", err)
	}
	if want == otherHead {
		t.Fatalf("fixture is not discriminating: both repositories report %s", want)
	}

	// The caller's process now carries another repository's location.
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)

	got, err := Head(target)
	if err != nil {
		t.Fatalf("Head under foreign GIT_DIR: %v", err)
	}
	if got != want {
		t.Errorf("S11_WRONG_REPO: Head(%s) = %s, want the target's %s (the foreign GIT_DIR repository's HEAD is %s)", target, got, want, otherHead)
	}

	raw, err := RunRaw(target, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("RunRaw under foreign GIT_DIR: %v", err)
	}
	if gotRaw := strings.TrimSpace(string(raw)); gotRaw != want {
		t.Errorf("S11_WRONG_REPO: RunRaw(%s) = %s, want the target's %s", target, gotRaw, want)
	}
}
