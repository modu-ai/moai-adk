package signtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFixtureGitStartsNoBackgroundMaintenance pins the fixture repository as
// quiescent once a git command returns. By default every `git commit` starts a
// detached `git maintenance run --auto`; that process keeps writing the
// fixture's .git (it creates and removes .git/objects/maintenance.lock) after
// the commit has returned, so a Snapshot walking .git can fail with "no such
// file or directory" or see a file the next Snapshot no longer does.
//
// GIT_TRACE records every child git process, so the assertion does not depend
// on winning a timing race against the detached process.
func TestFixtureGitStartsNoBackgroundMaintenance(t *testing.T) {
	trace := filepath.Join(t.TempDir(), "git-trace.txt")
	t.Setenv("GIT_TRACE", trace)

	p := New(t)
	p.WriteFile("f.txt", "second commit")
	p.Git("add", "-A")
	p.Git("commit", "-q", "-m", "second")

	data, err := os.ReadFile(trace)
	if err != nil {
		t.Fatalf("read git trace: %v", err)
	}
	if !strings.Contains(string(data), "built-in: git commit") {
		t.Fatalf("git trace recorded no commit — the probe did not observe git:\n%s", data)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "maintenance run") {
			t.Fatalf("a fixture git command started background maintenance:\n%s", line)
		}
	}
}
