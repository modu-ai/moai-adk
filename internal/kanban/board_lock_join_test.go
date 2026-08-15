// board_lock_join_test.go — regression coverage for the board-store lock
// release fold (joinBoardReleaseErr). The two guarded board entry points
// (WriteBoardState, RecoverBoard) share the fold, so its four-combination
// contract is exercised directly here rather than through the entry points —
// the release path closes a file descriptor, which a test cannot fail on
// demand (same constraint that shaped the backlog_store precedent).
package kanban

import (
	"errors"
	"strings"
	"testing"
)

// TestJoinBoardReleaseErr_BothFailuresSurvive — the regression this exists
// for: when the mutation fails AND the lock release fails, BOTH must reach the
// caller. The prior guard (`relErr != nil && err == nil`) dropped the release
// error in exactly that case, so a wedged lock — the artifact that blocks
// every later writer on Windows — surfaced as an ordinary mutation fault.
// Both guarded entry points share the fold, so both op prefixes are checked.
func TestJoinBoardReleaseErr_BothFailuresSurvive(t *testing.T) {
	t.Parallel()
	mutErr := errors.New("mutation refused")
	relErr := errors.New("close: bad file descriptor")

	for _, op := range []string{"write board state", "recover board"} {
		got := joinBoardReleaseErr(mutErr, relErr, op)
		if got == nil {
			t.Fatalf("joinBoardReleaseErr(mut, rel, %q) = nil; want both errors", op)
		}
		if !errors.Is(got, mutErr) {
			t.Errorf("%s: mutation error did not survive the join: %v", op, got)
		}
		if !errors.Is(got, relErr) {
			t.Errorf("%s: release error did not survive the join: %v", op, got)
		}
		if !strings.Contains(got.Error(), op+": lock release failed") {
			t.Errorf("%s: joined error does not name the entry point and release failure: %v", op, got)
		}
	}
}

// TestJoinBoardReleaseErr_SingleAndCleanPaths — the other three combinations.
// A clean release must not manufacture an error, and either failure alone must
// pass through identifiably.
func TestJoinBoardReleaseErr_SingleAndCleanPaths(t *testing.T) {
	t.Parallel()
	mutErr := errors.New("mutation refused")
	relErr := errors.New("close: bad file descriptor")

	if got := joinBoardReleaseErr(nil, nil, "write board state"); got != nil {
		t.Errorf("both-clean returned %v; want nil", got)
	}
	if got := joinBoardReleaseErr(mutErr, nil, "write board state"); !errors.Is(got, mutErr) {
		t.Errorf("mutation-only returned %v; want the mutation error", got)
	}
	got := joinBoardReleaseErr(nil, relErr, "recover board")
	if !errors.Is(got, relErr) {
		t.Errorf("release-only returned %v; want the release error", got)
	}
	if !strings.Contains(got.Error(), "recover board: lock release failed") {
		t.Errorf("release-only error does not name the entry point: %v", got)
	}
}
