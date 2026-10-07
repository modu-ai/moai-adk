package homestate

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestNotGitRepositoryAcceptsOnlyOrdinaryDiscovery(t *testing.T) {
	cmd := exec.Command("git", "-C", t.TempDir(), "rev-parse", "--absolute-git-dir")
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANGUAGE=C")
	_, err := cmd.Output()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 128 {
		t.Fatalf("native nonGit positive control failed: %v", err)
	}
	for _, tc := range []struct {
		name, diagnostic string
		want             bool
	}{
		{"ordinary", "fatal: not a git repository (or any of the parent directories): .git", true},
		{"boundary", "fatal: not a git repository (or any parent up to mount point /tmp)\nStopping at filesystem boundary (GIT_DISCOVERY_ACROSS_FILESYSTEM not set).", true},
		{"corrupt-null", "fatal: not a git repository: (null)", false},
		{"corrupt-path", "fatal: not a git repository: /missing/gitdir", false},
		{"unknown-with-substring", "unknown failure: not a git repository", false},
		{"extra-error", "fatal: not a git repository (or any of the parent directories): .git\ncontrolled failure", false},
		{"incomplete-boundary", "fatal: not a git repository (or any parent up to mount point /tmp)", false},
		{"localized", "fatal: pas un dépôt git", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			controlled := &exec.ExitError{ProcessState: exit.ProcessState, Stderr: []byte(tc.diagnostic)}
			if got := notGitRepository(controlled); got != tc.want {
				t.Fatalf("classification=%t want=%t diagnostic=%q", got, tc.want, tc.diagnostic)
			}
		})
	}
}
