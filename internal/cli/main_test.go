package cli

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
// When two tests declaring t.Parallel() make that first call concurrently on a
// SHARED package-level *cobra.Command, the write races the read and the race
// detector fires. The window closes once the flag flips, which is exactly why
// the CI failure was intermittent rather than deterministic.
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
// once per node to complete cobra's lazy sort before any test runs.
//
// The traversal must be recursive: hookCmd and githubCmd are separate package
// globals registered as children of rootCmd (github.go:100, hook.go:38), each
// owning its own commandsAreSorted field. Enumerating individual commands
// instead would silently miss every global added later.
func warmUpCommandTree(c *cobra.Command) {
	warmUpDone = true
	for _, sub := range c.Commands() { // triggers the lazy sort on c
		warmUpCommandTree(sub)
	}
}

func TestMain(m *testing.M) {
	warmUpCommandTree(rootCmd)
	os.Exit(m.Run())
}

// TestWarmUpReachability is the REQ-CFS-004 guard. It runs serially (no
// t.Parallel) and asserts only the reachability signal.
func TestWarmUpReachability(t *testing.T) {
	if !warmUpDone {
		t.Fatal("cobra lazy-sort warm-up did not run: TestMain must call " +
			"warmUpCommandTree(rootCmd) before m.Run(). Removing it reintroduces " +
			"the data race documented at the top of main_test.go.")
	}
}
