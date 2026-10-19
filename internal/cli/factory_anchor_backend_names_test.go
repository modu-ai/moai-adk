package cli

import (
	"path/filepath"
	"testing"
)

func TestLauncherAnchorBackendWorktreeNames(t *testing.T) {
	pinSnapshotModel(t)
	for _, tc := range []struct {
		name, value, directory string
	}{
		{"lone_dash", "-", "-"},
		{"slash_normalized", "feature/foo", "feature+foo"},
		{"nested_slash_normalized", "feature/foo/bar", "feature+foo+bar"},
	} {
		for _, arm := range leaderArms {
			t.Run(tc.name+"/"+arm, func(t *testing.T) {
				fx := newLeaderFixture(t)
				target := filepath.Join(fx.root, ".claude", "worktrees", tc.directory)
				anchorGit(t, fx.root, "worktree", "add", "-q", "-b", "WT-backend-name", target, fx.base)
				obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--", "-w", tc.value})
				assertRefusedBeforeMutation(t, arm, obs)
				refusalNames(t, arm, obs, target, "WT-backend-name")
			})
		}
	}
	for _, arm := range leaderArms {
		t.Run("normalized_plain_tree/"+arm, func(t *testing.T) {
			fx := newLeaderFixture(t)
			literal := filepath.Join(fx.root, ".claude", "worktrees", "feature", "foo")
			selected := filepath.Join(fx.root, ".claude", "worktrees", "feature+foo")
			anchorGit(t, fx.root, "worktree", "add", "-q", "-b", "WT-unselected", literal, fx.base)
			anchorGit(t, fx.root, "worktree", "add", "-q", "-b", "plain-selected", selected, fx.base)
			obs := driveLeaderLaunch(t, fx, arm, []string{"-f", "--", "-w", "feature/foo"})
			if obs.err != nil || !obs.launched {
				t.Fatalf("normalized plain target must launch: err=%v launched=%v", obs.err, obs.launched)
			}
		})
	}
}
