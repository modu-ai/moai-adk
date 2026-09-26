package homestate_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

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

	for _, tc := range []struct{ name, dir string }{
		{"victim_primary", victimRepo},
		{"victim_linked_worktree", victimWorktree},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GIT_DIR", callerWorktreeGitDir)
			if got := homestate.CanonicalProjectRoot(tc.dir); got != filepath.Clean(want) {
				t.Fatalf("CanonicalProjectRoot(%s) = %q under an inherited GIT_DIR, want %q", tc.name, got, want)
			}
		})
	}
}
