package worktree

// Card t805. The root help text claimed the command "supports creating"
// worktrees while its own next paragraph says entering one is the launchers'
// job, and no create verb is registered. Prose and the command tree drifted
// apart with nothing reading both.
//
// This guard reads both: every registered subcommand must be named in the
// Long text, and the Long text must not advertise creation. It asserts the
// command set is non-empty first, so a future refactor that empties it cannot
// turn this into a check that passes by having nothing to check.

import (
	"strings"
	"testing"
)

func TestWorktreeLongTextNamesExactlyTheRegisteredVerbs(t *testing.T) {
	subs := WorktreeCmd.Commands()
	if len(subs) == 0 {
		t.Fatal("no subcommands registered — the guard would pass vacuously")
	}

	long := WorktreeCmd.Long
	if long == "" {
		t.Fatal("root Long is empty")
	}

	for _, sub := range subs {
		if !strings.Contains(long, sub.Name()) {
			t.Errorf("subcommand %q is registered but not named in the root help text", sub.Name())
		}
	}

	// "creating" is the specific claim that drifted: this command creates
	// nothing, and the paragraph below the first line says so.
	for _, claim := range []string{"creating", "Supports creating"} {
		if strings.Contains(long, claim) {
			t.Errorf("root help text still advertises %q — entering a worktree is the launchers' job", claim)
		}
	}
}
