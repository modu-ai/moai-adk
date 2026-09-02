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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/profile"
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
// recomputed inside the guard test, because the shared rootCmd tree GROWS
// during the run: help_order_test.go drives runFang(rootCmd), and cobra's
// Execute path calls InitDefaultHelpCmd / InitDefaultCompletionCmd, which
// append `help` and `completion` (plus its four shell children) to the global
// root. Recounting inside the guard test measured 188 nodes against a warm-up
// of 182 and failed a CORRECT implementation. Comparing two numbers taken at
// the same instant removes that false positive.
//
// This does NOT weaken the guard. countCommandTree never touches warmUpDone or
// warmUpVisited, so gutting the warm-up recursion still pins warmUpVisited at 1
// while this stays at the full tree size — the two numbers diverge exactly as
// they must. DO NOT "simplify" this back into the guard test body.
var warmUpTreeSize int

// warmUpCommandTree performs a depth-first traversal from c, calling Commands()
// once per node to complete cobra's lazy sort before any test runs.
//
// The traversal must be recursive: hookCmd and githubCmd are separate package
// globals registered as children of rootCmd (github.go:100, hook.go:38), each
// owning its own commandsAreSorted field. Enumerating individual commands
// instead would silently miss every global added later.
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

// sandboxProfileBaseDir points profile.GetBaseDir at a throwaway directory for
// the whole package run, so no test in internal/cli can reach the developer's
// real ~/.moai/claude-profiles.
//
// WHY THIS IS PACKAGE-WIDE rather than per-test. unifiedLaunch step 5
// (launcher.go) calls profile.RecordLastUsedProfile whenever it is handed a
// named profile, which rewrites launch.yaml — the ledger that carries both the
// per-project projects[] map (the live source this binary reads) and the
// legacy last_profile key. Any test that reaches unifiedLaunch with a name,
// directly (launcher_test.go TestUnifiedLaunch_Claude) or through a cobra
// RunE (cc_test.go `-p work`), therefore writes the real ledger unless the
// base dir is overridden. Observed consequence (historical): a fixture name
// was left recorded with no matching profile directory, the stale-record
// guard returned "", and every subsequent `moai cc` silently launched with no
// --model at all.
//
// Overriding here rather than in each test is deliberate: the leak is caused by
// a call several frames below the test body, so a test author has no local
// signal that their test writes to $HOME. A per-test override only fixes the
// tests that exist today; this covers the ones added tomorrow.
//
// t.TempDir() is unavailable in TestMain (no *testing.T), so the directory is
// created and removed manually.
// profileBaseDirEnv carries the sandbox path from a test process to any test
// binary it re-executes. Several tests in this package spawn os.Args[0] as a
// helper subprocess (exitcode_guard_test.go, todo_test.go,
// launch_session_pid_exec_posix_test.go), every one of them with
// append(os.Environ(), …) — so the child picks this up with no per-call-site
// wiring, and future helpers do too.
const profileBaseDirEnv = "MOAI_CLI_TEST_PROFILE_BASE"

func sandboxProfileBaseDir() func() {
	// A re-executed child adopts the parent's sandbox and removes nothing: the
	// helper bodies end in os.Exit, which returns through neither m.Run() nor
	// the restore call after it, so anything a child created for itself would
	// outlive every process that knew about it. Ownership stays with the
	// process that minted the directory.
	if inherited := os.Getenv(profileBaseDirEnv); inherited != "" {
		orig := profile.BaseDirOverride
		profile.BaseDirOverride = inherited
		return func() { profile.BaseDirOverride = orig }
	}

	dir, err := os.MkdirTemp("", "moai-cli-profiles-")
	if err != nil {
		// Fall back to a path under the OS temp dir rather than silently
		// leaving the real base dir in play.
		dir = filepath.Join(os.TempDir(), "moai-cli-profiles-fallback")
		_ = os.MkdirAll(dir, 0o755)
	}
	orig := profile.BaseDirOverride
	profile.BaseDirOverride = dir
	_ = os.Setenv(profileBaseDirEnv, dir)
	return func() {
		profile.BaseDirOverride = orig
		_ = os.Unsetenv(profileBaseDirEnv)
		_ = os.RemoveAll(dir)
	}
}

// RESIDUE GUARD (SPEC-CLI-TEST-CWD-ISOLATION-001 REQ-3).
//
// Go test binaries run with cwd = the package directory, so any state write
// whose project root resolves to an empty or relative value lands INSIDE the
// repository tree as internal/cli/.moai (the name-claim registries
// state/todo/leads.json and state/factory/workers.json were the measured
// producers; the state subdirectory has already drifted once — kanban/ →
// todo/ — so the guard watches the .moai DIRECTORY, never a file list). The
// residue is gitignored and therefore invisible to git status, but it breaks
// every .moai-marker upward walk that later starts below the repository root:
// the walk stops at internal/cli/.moai and misjudges an applicable tree as
// inapplicable — a different answer, not an error (the t317 D9 gate failure).
//
// The guard judges the existence DELTA across the run: .moai absent at
// TestMain entry and present after m.Run() fails the run, naming the detected
// path. A directory that already existed at entry is not a delta this run
// produced, so the guard stays silent for it — the message tells the developer
// to remove it, which re-arms the delta for the next run. Tree-wide detection
// beyond the package directory is the AC's baseline-delta scan; the guard is
// deliberately O(1) at the measured locus.
//
// TestMain always runs, so the guard rides every -run selector — including
// selectors that match no tests, where a named guard test would be filtered
// out and pass vacuously.
func TestMain(m *testing.M) {
	restoreProfileBaseDir := sandboxProfileBaseDir()

	// Pin the watched path from the entry cwd (absolute) so a test that chdirs
	// cannot move the locus out from under the post-run check.
	entryWD, err := os.Getwd()
	if err != nil {
		entryWD = "."
	}
	residueDir := filepath.Join(entryWD, ".moai")
	residueExistedAtStart := residueGuardDirExists(residueDir)

	warmUpCommandTree(rootCmd)
	// Serial, single-goroutine, immediately after the warm-up: these Commands()
	// calls cannot open a race window, and the tree cannot have drifted yet.
	warmUpTreeSize = countCommandTree(rootCmd)

	code := m.Run()

	if !residueExistedAtStart && residueGuardDirExists(residueDir) {
		fmt.Fprintf(os.Stderr, "RESIDUE GUARD FAIL: this test run created %s — "+
			"internal/cli tests must not write .moai state into the package working "+
			"directory (SPEC-CLI-TEST-CWD-ISOLATION-001 REQ-1/REQ-2/REQ-3). Isolate the "+
			"producing test's project root (see the SPEC's mechanism ladder), then remove "+
			"the directory so the guard re-arms for the next run.\n", residueDir)
		if code == 0 {
			code = 1
		}
	}

	restoreProfileBaseDir()
	os.Exit(code)
}

// residueGuardDirExists reports whether path names an existing directory. It is
// deliberately separate from any other helper so the guard's judgment cannot be
// redirected by a change to shared machinery.
func residueGuardDirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// TestProfileBaseDirIsSandboxed is the guard for sandboxProfileBaseDir.
//
// It fails deterministically if the TestMain call is removed: with no override,
// profile.GetBaseDir() resolves to $HOME/.moai/claude-profiles and
// BaseDirOverride is "". Both assertions below then fire.
//
// The second assertion is the load-bearing one. An override set to the real
// base dir would satisfy a non-emptiness check while still writing $HOME, so
// the guard compares against the actual home-derived path rather than merely
// asserting "something was set".
func TestProfileBaseDirIsSandboxed(t *testing.T) {
	if profile.BaseDirOverride == "" {
		t.Fatal("profile.BaseDirOverride is empty: TestMain must call " +
			"sandboxProfileBaseDir() before m.Run(). Without it, any test " +
			"reaching unifiedLaunch with a named profile rewrites the " +
			"developer's real ~/.moai/claude-profiles/launch.yaml.")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory; sandbox comparison unavailable")
	}
	realBase := filepath.Join(home, ".moai", "claude-profiles")
	if got := profile.GetBaseDir(); got == realBase {
		t.Fatalf("profile.GetBaseDir() = %q, which is the real user profile "+
			"base. Tests in this package must never resolve to it.", got)
	}
}

// TestWarmUpReachability is the REQ-CFS-004 guard. It runs serially (no
// t.Parallel) and asserts both warm-up signals: execution (warmUpDone) and
// traversal depth (warmUpVisited).
func TestWarmUpReachability(t *testing.T) {
	if !warmUpDone {
		t.Fatal("cobra lazy-sort warm-up did not run: TestMain must call " +
			"warmUpCommandTree(rootCmd) before m.Run(). Removing it reintroduces " +
			"the data race documented at the top of main_test.go.")
	}

	// warmUpTreeSize comes from countCommandTree, the separate traversal, taken
	// at warm-up time (see its doc comment for why it is not recounted here).
	want := warmUpTreeSize
	if want <= 1 {
		t.Fatalf("guard is vacuous: the command tree rooted at rootCmd holds %d "+
			"node(s), so the traversal-depth assertion below would pass even with "+
			"the warm-up recursion deleted. The root must have at least one "+
			"subcommand for this guard to be falsifiable.", want)
	}
	if warmUpVisited != want {
		t.Fatalf("cobra lazy-sort warm-up did not traverse the whole command tree: "+
			"warmUpCommandTree visited %d node(s) but the tree holds %d. TestMain "+
			"must call warmUpCommandTree(rootCmd) AND that function must recurse "+
			"over Commands() — visiting the root alone leaves every subcommand's "+
			"commandsAreSorted unwritten, which reintroduces the data race "+
			"documented at the top of main_test.go.", warmUpVisited, want)
	}
}
