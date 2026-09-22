package worktree

// Card t805 originally pinned a management-only command set. Card t1070
// deliberately reverses the no-creation doctrine by registering one thin
// creation adapter while keeping entry in the launcher surfaces.
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

	if !strings.Contains(long, "moai worktree new <name>") {
		t.Error("root help text does not document the harness-neutral creation verb")
	}
	if !strings.Contains(long, "Entering an existing worktree remains the launchers' job") {
		t.Error("root help text blurred creation with launcher-owned entry")
	}
}
