package preference

// Package-wide test entry point (SPEC-CI-FLAKY-STABILIZE-001, REQ-CFS-001..004).
//
// WHY THIS FILE EXISTS — cobra lazy-sort data race.
//
// cobra v1.10.2 Commands() (command.go:1332-1339) is not a pure accessor:
//
//	func (c *Command) Commands() []*Command {
//	    if EnableCommandSorting && !c.commandsAreSorted {
//	        sort.Sort(commandSorterByName(c.commands))  // sorts IN PLACE
//	        c.commandsAreSorted = true                  // WRITE at command.go:1336
//	    }
//	    return c.commands                               // READ at command.go:1334
//	}
//
// EnableCommandSorting defaults to true (cobra.go:59), so the FIRST call sorts
// c.commands in place and writes c.commandsAreSorted; later calls only read it.
// TestPreferenceCmd_HasDecayScanChild and TestPreferenceCmd_HasToggleChild both
// declare t.Parallel() and both call PreferenceCmd.Commands() on that shared
// package global — when their first calls interleave, the write races the read.
// That is the exact failure captured in the CI log (command.go:1334 read vs
// command.go:1336 write); it is intermittent because the window closes as soon
// as the flag flips.
//
// The warm-up below closes the window before any test starts, so every
// t.Parallel() test observes read-only state. Deleting the warmUpCommandTree
// call in TestMain reintroduces the race — the guard test in this file fails
// deterministically if that happens.
//
// DO NOT replace the guard with an assertion on Commands() slice sortedness.
// Because Commands() sorts inside the accessor, the returned slice is sorted
// after ANY call, warm-up or not; a sortedness guard can never fail and would
// silently permit the warm-up's removal.

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// warmUpDone records that warmUpCommandTree actually ran. It is the reachability
// signal the guard test asserts. The assignment lives INSIDE warmUpCommandTree
// (never in TestMain) so that commenting out the single warmUpCommandTree call
// also makes the assignment unreachable — that coupling is what makes the guard
// deterministic instead of vacuous.
var warmUpDone bool

// warmUpCommandTree performs a depth-first traversal from c, calling Commands()
// once per node to complete cobra's lazy sort before any test runs. The
// traversal is recursive so that globals added below the root in future are
// covered without editing an enumeration.
func warmUpCommandTree(c *cobra.Command) {
	warmUpDone = true
	for _, sub := range c.Commands() { // triggers the lazy sort on c
		warmUpCommandTree(sub)
	}
}

func TestMain(m *testing.M) {
	warmUpCommandTree(PreferenceCmd)
	os.Exit(m.Run())
}

// TestWarmUpReachability is the REQ-CFS-004 guard. It runs serially (no
// t.Parallel) and asserts only the reachability signal.
func TestWarmUpReachability(t *testing.T) {
	if !warmUpDone {
		t.Fatal("cobra lazy-sort warm-up did not run: TestMain must call " +
			"warmUpCommandTree(PreferenceCmd) before m.Run(). Removing it " +
			"reintroduces the data race documented at the top of main_test.go.")
	}
}
