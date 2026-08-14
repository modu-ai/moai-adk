// board_coverage_test.go — branch coverage for the M1 store's error and
// normalization paths (SPEC-KANBAN-BOARD-001 M1; the 85% gate is §E.3's).
package kanban

import (
	"os"
	"testing"
)

// TestBoardRoot_NonGitDirFails — the resolution error path: an unresolvable
// start directory surfaces an error rather than a guessed root.
func TestBoardRoot_NonGitDirFails(t *testing.T) {
	t.Parallel()
	if _, err := BoardRoot(t.TempDir()); err == nil {
		t.Fatal("BoardRoot(non-git) err = nil, want non-nil")
	}
}

// TestTransitionIntoRun_Branches — the transition's three load-bearing
// branches beyond the fresh-append path: an empty spec id is rejected, a card
// already in run re-enters idempotently (no WIP double-count), and a card
// recorded elsewhere MOVES into run in place rather than duplicating.
func TestTransitionIntoRun_Branches(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")

	if err := TransitionIntoRun(root, "lead-sess", ""); err == nil {
		t.Fatal("TransitionIntoRun(empty spec) err = nil, want rejection")
	}

	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{
		{SpecID: "SPEC-KB-0070", Column: "run", Holder: "s1", LastMovedAt: "t0"},
		{SpecID: "SPEC-KB-0071", Column: "run", Holder: "s2", LastMovedAt: "t0"},
		{SpecID: "SPEC-KB-0072", Column: "plan", LastMovedAt: "t0"},
	}})

	// Idempotent re-entry: a card already in run does not refuse as a third
	// occupancy and does not duplicate.
	if err := TransitionIntoRun(root, "lead-sess", "SPEC-KB-0070"); err != nil {
		t.Fatalf("idempotent re-entry refused: %v", err)
	}

	// A third DISTINCT card is still refused (bound 2).
	if err := TransitionIntoRun(root, "lead-sess", "SPEC-KB-0073"); !IsWipLimitExceeded(err) {
		t.Fatalf("third distinct card: err = %v, want WIP refusal", err)
	}

	// Moving an existing plan card into an occupied run is likewise refused —
	// it would be the third occupant.
	if err := TransitionIntoRun(root, "lead-sess", "SPEC-KB-0072"); !IsWipLimitExceeded(err) {
		t.Fatalf("move of a plan card into a full run: err = %v, want WIP refusal", err)
	}

	// With one run slot free, the plan card moves in place — one card, moved.
	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{
		{SpecID: "SPEC-KB-0071", Column: "run", Holder: "s2", LastMovedAt: "t0"},
		{SpecID: "SPEC-KB-0072", Column: "plan", LastMovedAt: "t0"},
	}})
	if err := TransitionIntoRun(root, "lead-sess", "SPEC-KB-0072"); err != nil {
		t.Fatalf("move into a free run slot refused: %v", err)
	}
	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard: %v", err)
	}
	found := map[string]Column{}
	for _, c := range st.Cards {
		found[c.SpecID] = c.Column
	}
	if len(st.Cards) != 2 || found["SPEC-KB-0072"] != "run" || found["SPEC-KB-0071"] != "run" {
		t.Fatalf("move did not happen in place: %v (cards=%d)", found, len(st.Cards))
	}
}

// TestDeclareRole_EmptyRoleRejected — a declaration carrying no role is not a
// declaration.
func TestDeclareRole_EmptyRoleRejected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := DeclareRole(root, "sess-x", "", "label"); err == nil {
		t.Fatal("DeclareRole(empty role) err = nil, want rejection")
	}
}

// TestBoardLock_PathDiagnostics — the lock reports its artifact path.
func TestBoardLock_PathDiagnostics(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	lock, err := AcquireBoardLock(root)
	if err != nil {
		t.Fatalf("AcquireBoardLock: %v", err)
	}
	if lock.Path() != boardLockPath(root) {
		t.Fatalf("Path() = %q, want %q", lock.Path(), boardLockPath(root))
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
	// Idempotent release and nil-receiver safety.
	if err := lock.Release(); err != nil {
		t.Fatalf("second Release: %v", err)
	}
	var nilLock *BoardLock
	if nilLock.Path() != "" || nilLock.Release() != nil {
		t.Fatal("nil receiver safety broken")
	}
}

// TestWriteBoardState_UnwritableDirSurfacesError — the atomic write's temp
// file cannot be created in an unwritable board directory; the error surfaces
// rather than being swallowed, and the board is left without a state file.
func TestWriteBoardState_UnwritableDirSurfacesError(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("chmod-based denial skipped on windows")
	}
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")
	// DeclareRole has already created BoardDir via its roles/ subdirectory,
	// so the permission must be forced explicitly rather than via MkdirAll's
	// create-time perm.
	if err := os.Chmod(BoardDir(root), 0o500); err != nil {
		t.Fatalf("chmod board dir read-only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(BoardDir(root), 0o755) })

	err := WriteBoardState(root, "lead-sess", func(st *BoardState) error { return nil })
	if err == nil {
		t.Fatal("WriteBoardState(read-only board dir) err = nil, want surfaced error")
	}
	if _, statErr := os.Stat(BoardPath(root)); statErr == nil {
		t.Fatal("state file appeared despite the write failure")
	}
}

// TestLoadBoard_NullCardsNormalized — a state document whose cards array is
// null normalizes to an empty slice, not a nil the caller must guard.
func TestLoadBoard_NullCardsNormalized(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(BoardDir(root), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(BoardPath(root), []byte(`{"cards":null}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard(cards:null): %v", err)
	}
	if st.Cards == nil || len(st.Cards) != 0 {
		t.Fatalf("cards = %#v, want normalized empty slice", st.Cards)
	}
}

// TestClearStaleBoardLock_UnixGatedOutOnUnreadableArtifact — on Unix the
// clear is a no-op (finding F5), so even an unreadable artifact yields the
// not-applicable report rather than the Windows-substrate unreadable refusal
// (which lives in the windows-tagged suite).
func TestClearStaleBoardLock_UnixGatedOutOnUnreadableArtifact(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("unix-only gate observation")
	}
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(BoardDir(root), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(boardLockPath(root), 0o755); err != nil {
		t.Fatalf("mkdir lock path: %v", err)
	}
	report, err := ClearStaleBoardLock(root)
	if err != nil {
		t.Fatalf("ClearStaleBoardLock(unix) error = %v, want nil no-op", err)
	}
	if report == nil || report.Removed {
		t.Fatalf("report = %+v, want Removed=false not-applicable on unix", report)
	}
}
