package cli

import (
	"context"
	"io"
	"testing"

	"github.com/spf13/cobra"
)

// TestT957_RefreshChildNeverWalksUpFromCwd pins the second half of the config-root
// alignment: the detached refresh child.
//
// The render path resolves its config root from the state anchor and hands that
// same root to the child as --board-root, so parent and child read one tree. The
// alignment survives only while the child takes the root it is GIVEN. If the
// cwd walk-up ever moves above the refresh branches — or the child grows a
// resolution of its own — the divergence comes back one process over, where the
// render's own output cannot show it: the parent would read the anchored tree
// and the cache the child wrote would describe a different one.
//
// This asserts the property directly: on the refresh entry points, the cwd
// walk-up is never consulted.
func TestT957_RefreshChildNeverWalksUpFromCwd(t *testing.T) {
	boardRoot := t.TempDir()

	for _, tc := range []struct {
		name   string
		github bool
		landed bool
	}{
		{"refresh-github", true, false},
		{"refresh-landed", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			walkedUp := false
			origFn := findProjectRootFn
			findProjectRootFn = func() (string, error) {
				walkedUp = true
				return boardRoot, nil
			}
			t.Cleanup(func() { findProjectRootFn = origFn })

			origGitHub, origLanded, origRoot := statuslineRefreshGitHub, statuslineRefreshLanded, statuslineBoardRoot
			statuslineRefreshGitHub, statuslineRefreshLanded, statuslineBoardRoot = tc.github, tc.landed, boardRoot
			t.Cleanup(func() {
				statuslineRefreshGitHub, statuslineRefreshLanded, statuslineBoardRoot = origGitHub, origLanded, origRoot
			})

			cmd := &cobra.Command{}
			cmd.SetContext(context.Background())
			cmd.SetOut(io.Discard)

			if err := runStatusline(cmd, nil); err != nil {
				t.Fatalf("runStatusline: %v", err)
			}
			if walkedUp {
				t.Errorf("the %s child consulted the cwd walk-up — it must use the --board-root it was given, or parent and child diverge", tc.name)
			}
		})
	}
}
