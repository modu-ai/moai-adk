package cli

// SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-002/REQ-010 (sync-audit finding F2): a
// refused entry launches nothing. The --spawn branch of cc and glm re-issues the
// command in a new tmux window, so an entry the launcher refuses must be refused
// by the parent, where the operator reads it; inside the new window the refusal
// would vanish with the window when it closes.

import (
	"strings"
	"testing"
)

// spawnRefusedShapes are entry shapes the launcher refuses, each written with
// --spawn after them (the flag position the user types it in).
var spawnRefusedShapes = [][]string{
	{"-k"},
	{"--kanban"},
	{"-l", "lane-2"},
	{"--lane", "lane-2"},
	{"-l", "-f"},
	{"-f", "-l"},
	{"-f", "3"},
	{"-f", "lane"},
}

func TestSpawnRefusesBadEntriesBeforeOpeningAWindow(t *testing.T) {
	for _, verb := range laneVerbs {
		for _, shape := range spawnRefusedShapes {
			args := append(append([]string{}, shape...), "--spawn")
			t.Run(verb.name+"_"+strings.Join(args, "_"), func(t *testing.T) {
				spawned := withSpawnStubs(t, true, "%1", nil)
				_ = m2Refusal(t, verb.backend, verb.entry, args)
				if spawned.calls != 0 {
					t.Errorf("%s %v opened %d tmux window(s); a refused entry launches nothing", verb.name, args, spawned.calls)
				}
			})
		}
	}
}

// TestSpawnStillOpensAWindowForAnAcceptedEntry is the positive control: the
// entry pre-check must not turn --spawn off for the entries the launcher
// accepts.
func TestSpawnStillOpensAWindowForAnAcceptedEntry(t *testing.T) {
	for _, verb := range laneVerbs {
		for _, shape := range [][]string{{"-l", "--spawn"}, {"--spawn", "-l"}, {"-f", "--spawn"}} {
			t.Run(verb.name+"_"+strings.Join(shape, "_"), func(t *testing.T) {
				spawned := withSpawnStubs(t, true, "%1", nil)
				root := netLaneFixture(t, verb.backend)
				launch := netDriveLaunch(t, root, verb.entry, shape)
				if launch.err != nil {
					t.Fatalf("%s %v: %v, want the window to open", verb.name, shape, launch.err)
				}
				if spawned.calls != 1 {
					t.Errorf("%s %v opened %d tmux window(s), want exactly 1", verb.name, shape, spawned.calls)
				}
				if launch.launched {
					t.Errorf("%s %v replaced this process; --spawn keeps the current session", verb.name, shape)
				}
			})
		}
	}
}
