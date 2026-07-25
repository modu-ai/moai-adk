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
//
// TWO SIGNALS, TWO MUTATIONS.
//
// The guard observes two independent signals because each one alone leaves a
// mutation uncovered:
//
//   - warmUpDone (execution) catches CALL-DELETION — removing the
//     warmUpCommandTree call from TestMain also makes the assignment
//     unreachable, so the flag stays false.
//   - warmUpVisited (traversal depth) catches BODY-GUTTING — keeping the
//     TestMain call but deleting the recursive loop still sets warmUpDone,
//     so warmUpDone alone passes. That mutation was measured to reintroduce
//     the race, which is why depth must be observed separately.
//
// The node count the guard compares warmUpVisited against is produced by
// countCommandTree, a DELIBERATELY SEPARATE traversal (see its doc comment).
// Do not merge the two traversals into a shared helper.

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

// warmUpVisited counts the command nodes the warm-up recursion actually visited.
// It is the traversal-depth signal the guard test asserts, and it is additive to
// warmUpDone rather than a replacement for it.
//
// Like warmUpDone, the increment lives INSIDE warmUpCommandTree (never in
// TestMain). It sits ABOVE the recursive loop, so gutting only the loop body
// leaves the counter at exactly 1 — the root was still visited — while the
// independently computed tree size stays greater than 1. That divergence is
// what makes body-gutting deterministically detectable.
var warmUpVisited int

// warmUpTreeSize is the node count of the command tree AS IT STOOD when the
// warm-up ran, produced by countCommandTree — a separate traversal.
//
// It is captured in TestMain, immediately after the warm-up, rather than
// recomputed inside the guard test, because a shared cobra root can GROW during
// a run: cobra's Execute path calls InitDefaultHelpCmd / InitDefaultCompletionCmd,
// which append `help` and `completion` (plus its four shell children) to the
// global command. That drift was measured in the sibling internal/cli package,
// where recounting inside the guard test failed a CORRECT implementation.
// Comparing two numbers taken at the same instant removes that false positive,
// and keeping both packages symmetric keeps the guard reviewable as one design.
//
// This does NOT weaken the guard. countCommandTree never touches warmUpDone or
// warmUpVisited, so gutting the warm-up recursion still pins warmUpVisited at 1
// while this stays at the full tree size — the two numbers diverge exactly as
// they must. DO NOT "simplify" this back into the guard test body.
var warmUpTreeSize int

// warmUpCommandTree performs a depth-first traversal from c, calling Commands()
// once per node to complete cobra's lazy sort before any test runs. The
// traversal is recursive so that globals added below the root in future are
// covered without editing an enumeration.
func warmUpCommandTree(c *cobra.Command) {
	warmUpDone = true
	warmUpVisited++
	for _, sub := range c.Commands() { // triggers the lazy sort on c
		warmUpCommandTree(sub)
	}
}

// countCommandTree counts the command nodes reachable from c.
//
// DELIBERATE DUPLICATION — DO NOT refactor this to share a traversal helper
// with warmUpCommandTree. The whole point of this function is that it is a
// SEPARATE implementation: if the two traversals shared code, breaking one
// would move both numbers together and the guard could never fail. That is the
// same structural-unfalsifiability trap that retired the original sortedness
// guard. The duplication is the guard, not an oversight.
func countCommandTree(c *cobra.Command) int {
	n := 1
	for _, sub := range c.Commands() {
		n += countCommandTree(sub)
	}
	return n
}

func TestMain(m *testing.M) {
	warmUpCommandTree(PreferenceCmd)
	// Serial, single-goroutine, immediately after the warm-up: these Commands()
	// calls cannot open a race window, and the tree cannot have drifted yet.
	warmUpTreeSize = countCommandTree(PreferenceCmd)
	os.Exit(m.Run())
}

// TestWarmUpReachability is the REQ-CFS-004 guard. It runs serially (no
// t.Parallel) and asserts both warm-up signals: execution (warmUpDone) and
// traversal depth (warmUpVisited).
func TestWarmUpReachability(t *testing.T) {
	if !warmUpDone {
		t.Fatal("cobra lazy-sort warm-up did not run: TestMain must call " +
			"warmUpCommandTree(PreferenceCmd) before m.Run(). Removing it " +
			"reintroduces the data race documented at the top of main_test.go.")
	}

	// warmUpTreeSize comes from countCommandTree, the separate traversal, taken
	// at warm-up time (see its doc comment for why it is not recounted here).
	want := warmUpTreeSize
	if want <= 1 {
		t.Fatalf("guard is vacuous: the command tree rooted at PreferenceCmd holds "+
			"%d node(s), so the traversal-depth assertion below would pass even "+
			"with the warm-up recursion deleted. The root must have at least one "+
			"subcommand for this guard to be falsifiable.", want)
	}
	if warmUpVisited != want {
		t.Fatalf("cobra lazy-sort warm-up did not traverse the whole command tree: "+
			"warmUpCommandTree visited %d node(s) but the tree holds %d. TestMain "+
			"must call warmUpCommandTree(PreferenceCmd) AND that function must "+
			"recurse over Commands() — visiting the root alone leaves every "+
			"subcommand's commandsAreSorted unwritten, which reintroduces the data "+
			"race documented at the top of main_test.go.", warmUpVisited, want)
	}
}
