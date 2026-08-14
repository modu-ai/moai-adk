// board_lock_cross_test.go — AC-KB-019 (REQ-KB-019 + REQ-KB-009's bound,
// M1 minimal form): the board-wide lock serializes concurrent mutations of
// two DIFFERENT cards in SEPARATE OS PROCESSES.
//
// The scenario is §D.6's, run for real: two processes concurrently transition
// two different cards into `run` against a board already holding one. Exactly
// one must succeed; the final board must hold two cards in `run`, never
// three. A card-scoped lock passes nothing here — both processes take
// different locks — which is what makes this the criterion distinguishing
// board scope from card scope. The positive control (zero cards in run, same
// two concurrent transitions, both succeed) proves the refusal is conditional
// on the bound rather than on concurrency itself.
package kanban

import (
	"os"
	"os/exec"
	"testing"
)

// countRunCards loads the board and counts cards in the run column.
func countRunCards(t *testing.T, root string) int {
	t.Helper()
	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	n := 0
	for _, c := range st.Cards {
		if c.Column == ColumnRun {
			n++
		}
	}
	return n
}

// TestBoardMutation_SerializedAcrossProcesses — with one card already in run,
// two separate processes concurrently transition two DIFFERENT cards into
// run: exactly one succeeds, the other is refused with the named WIP error,
// and the final board holds two in run — never three.
func TestBoardMutation_SerializedAcrossProcesses(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("helper re-exec plumbing exercised on unix; windows substrate covered by GOOS=windows build")
	}
	root := t.TempDir()
	seedLead(t, root, "lead-sess")
	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{
		{SpecID: "SPEC-KB-0030", Column: "run", Holder: "run-sess-1", LastMovedAt: "2026-08-14T00:00:00Z"},
	}})

	// Fire both processes without waiting for the first — the lock, not the
	// spawn order, must serialize the read-modify-write.
	results := make(chan error, 2)
	go func() { results <- runTransitionHelperUnfatal(t, root, "lead-sess", "SPEC-KB-0031") }()
	go func() { results <- runTransitionHelperUnfatal(t, root, "lead-sess", "SPEC-KB-0032") }()
	first := <-results
	second := <-results

	successes := 0
	refusals := 0
	for _, r := range []error{first, second} {
		switch {
		case r == nil:
			successes++
		case IsWipLimitExceeded(r):
			refusals++
		default:
			t.Fatalf("unexpected transition error: %v", r)
		}
	}
	if successes != 1 || refusals != 1 {
		t.Fatalf("successes=%d refusals=%d, want exactly 1 and 1 — two different cards under a board-wide lock must serialize to the WIP bound", successes, refusals)
	}
	if got := countRunCards(t, root); got != 2 {
		t.Fatalf("cards in run = %d, want 2 — never three", got)
	}
}

// TestBoardMutation_ConcurrencyPositiveControl — with ZERO cards in run, the
// same two concurrent transitions BOTH succeed and the board holds two: the
// refusal above is conditional on the bound, not on concurrency itself.
func TestBoardMutation_ConcurrencyPositiveControl(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("helper re-exec plumbing exercised on unix; windows substrate covered by GOOS=windows build")
	}
	root := t.TempDir()
	seedLead(t, root, "lead-sess")
	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{}})

	results := make(chan error, 2)
	go func() { results <- runTransitionHelperUnfatal(t, root, "lead-sess", "SPEC-KB-0041") }()
	go func() { results <- runTransitionHelperUnfatal(t, root, "lead-sess", "SPEC-KB-0042") }()
	if err := <-results; err != nil {
		t.Fatalf("first zero-baseline transition failed: %v", err)
	}
	if err := <-results; err != nil {
		t.Fatalf("second zero-baseline transition failed: %v", err)
	}
	if got := countRunCards(t, root); got != 2 {
		t.Fatalf("cards in run = %d, want 2 (both succeed from an empty run column)", got)
	}
}

// runTransitionHelperUnfatal is runTransitionHelper without the Fatal on
// unexpected errors, so both concurrent processes' outcomes can be collected.
func runTransitionHelperUnfatal(t *testing.T, root, session, spec string) error {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestKanbanHelperProcess", "--")
	cmd.Env = append(os.Environ(),
		"MOAI_KANBAN_HELPER=transition-run",
		"HELPER_ROOT="+root,
		"HELPER_SESSION="+session,
		"HELPER_SPEC="+spec,
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 3 {
		return ErrWipLimitExceeded
	}
	t.Logf("transition-run helper for %s: %v: %s", spec, err, string(out))
	return err
}
