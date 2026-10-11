package cli

import (
	"path/filepath"
	"testing"
)

// t1628 sync-audit P2 — the launch directory is the git toplevel of the
// directory the operator ran from, even when the inherited GIT_DIR and
// GIT_WORK_TREE name another repository. A plain git call would resolve that
// repository and lane the session into its factory.
func TestCodexLaneLaunchDirIgnoresInheritedGitEnv(t *testing.T) {
	primary, _ := t1628LaneFixture(t)
	other := t.TempDir()
	t1628GitFixture(t, other, "init", "-q")
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)
	t1628Enter(t, primary)

	got := codexLaneLaunchDir()
	if !t1628SameRealPath(got, primary) {
		t.Errorf("codexLaneLaunchDir() = %q, want the directory the operator ran from (%q), not the repository named by the inherited GIT_DIR (%q)", got, primary, other)
	}
}
