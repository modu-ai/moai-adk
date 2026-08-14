// board_recover_test.go — the bounded recovery out of the unknown state
// (SPEC-KANBAN-BOARD-001 REQ-KB-013/022, M1; AC-KB-020).
//
// The repeated-read half is the load-bearing one for the auto-repair
// failure: a read path that quietly repairs an unreadable board satisfies
// "the board leaves the unknown state" perfectly while destroying the
// evidence of whatever killed the writer. Conversely an implementation with
// no recovery at all satisfies the repeated-read half; only the pair
// distinguishes a bounded exit from both a brick and a silent auto-repair.
// A replacement that discards an unreadable board without recording what was
// lost is the empty-board harm wearing the word "explicit" (AP-26).
package kanban

import (
	"os"
	"testing"
)

// corruptBoard seeds an unreadable board state file and returns its bytes.
func corruptBoard(t *testing.T, root string) []byte {
	t.Helper()
	if err := os.MkdirAll(BoardDir(root), 0o755); err != nil {
		t.Fatalf("mkdir board dir: %v", err)
	}
	raw := []byte(`{"cards":[{"spec_id":"SPEC-KB-0050","column":"run","hold`)
	if err := os.WriteFile(BoardPath(root), raw, 0o600); err != nil {
		t.Fatalf("seed corrupt board: %v", err)
	}
	return raw
}

// TestRecoverBoard_ReadPathNeverRepairs — AC-KB-020's load-bearing half: from
// the unknown state, repeated reads still report unknown and still leave the
// file byte-unchanged; the read path performs NO repair.
func TestRecoverBoard_ReadPathNeverRepairs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seeded := corruptBoard(t, root)

	for i := 0; i < 5; i++ {
		st, err := LoadBoard(root)
		if err == nil {
			t.Fatalf("read %d: err = nil, want unknown (st=%+v)", i, st)
		}
		if !IsBoardUnknown(err) {
			t.Fatalf("read %d: err = %v, want IsBoardUnknown", i, err)
		}
	}
	got, err := os.ReadFile(BoardPath(root))
	if err != nil {
		t.Fatalf("re-read board file: %v", err)
	}
	if string(got) != string(seeded) {
		t.Fatalf("board file changed under repeated reads — the read path repaired it")
	}
}

// TestRecoverBoard_BoundedRecovery — AC-KB-020's exit, bounded per REQ-KB-022:
// the sole writer, holding the board-wide lock, invokes the recovery
// explicitly; the board leaves the unknown state, subsequent reads succeed,
// a durable record of what could not be recovered exists beneath the board
// directory and preserves the lost content, and the recovery modified the
// state file alone.
func TestRecoverBoard_BoundedRecovery(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")
	if err := DeclareRole(root, "worker-sess", "run", "run-alpha"); err != nil {
		t.Fatalf("DeclareRole(worker): %v", err)
	}
	lostRaw := corruptBoard(t, root)
	roleBytes, err := os.ReadFile(roleDeclarationPath(root, "worker-sess"))
	if err != nil {
		t.Fatalf("read worker declaration: %v", err)
	}

	res, err := RecoverBoard(root, "lead-sess")
	if err != nil {
		t.Fatalf("RecoverBoard(lead) error = %v", err)
	}
	if res.Verdict != RecoveryVerdictReplaced {
		t.Fatalf("verdict = %q, want %q", res.Verdict, RecoveryVerdictReplaced)
	}
	if !res.LostRecorded {
		t.Fatal("LostRecorded = false; an unrecorded replacement is the harm AP-26 names")
	}

	// The board left the unknown state; subsequent reads succeed.
	st, err := LoadBoard(root)
	if err != nil {
		t.Fatalf("LoadBoard after recovery: %v", err)
	}
	if len(st.Cards) != 0 {
		t.Fatalf("recovered board holds %d cards; want the explicit empty replacement", len(st.Cards))
	}

	// The durable record exists and preserves what was lost.
	sidecarRaw, err := os.ReadFile(res.SidecarPath)
	if err != nil {
		t.Fatalf("recovery sidecar unreadable: %v", err)
	}
	if !containsAll(string(sidecarRaw), string(lostRaw)) {
		t.Fatalf("sidecar does not preserve the lost content:\nsidecar: %s", sidecarRaw)
	}

	// Extent: the state file alone — the declaration carrier is untouched.
	gotRole, err := os.ReadFile(roleDeclarationPath(root, "worker-sess"))
	if err != nil || string(gotRole) != string(roleBytes) {
		t.Fatalf("recovery modified the role declaration carrier: err=%v", err)
	}

	// Effect: one invocation, one definite verdict — invoking recovery again
	// on the now-readable board yields the recovered verdict and modifies
	// nothing.
	res2, err := RecoverBoard(root, "lead-sess")
	if err != nil {
		t.Fatalf("second RecoverBoard: %v", err)
	}
	if res2.Verdict != RecoveryVerdictRecovered {
		t.Fatalf("second verdict = %q, want %q — no re-entry, no retry", res2.Verdict, RecoveryVerdictRecovered)
	}
}

// TestRecoverBoard_ReadableBoardRecoveredUnchanged — REQ-KB-022's extent
// clause: recovery moves no card it can still read. A readable board
// recovers with its bytes untouched.
func TestRecoverBoard_ReadableBoardRecoveredUnchanged(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")
	writeBoardRaw(t, BoardPath(root), &BoardState{Cards: []Card{
		{SpecID: "SPEC-KB-0060", Column: "run", Holder: "run-sess-1", LastMovedAt: "2026-08-14T00:00:00Z"},
	}})
	before := readBoardBytes(t, root)

	res, err := RecoverBoard(root, "lead-sess")
	if err != nil {
		t.Fatalf("RecoverBoard(readable) error = %v", err)
	}
	if res.Verdict != RecoveryVerdictRecovered {
		t.Fatalf("verdict = %q, want %q", res.Verdict, RecoveryVerdictRecovered)
	}
	if string(before) != string(readBoardBytes(t, root)) {
		t.Fatal("recovery rewrote a readable board — it must move no card it can still read")
	}
	if res.LostRecorded {
		t.Fatal("LostRecorded = true for a readable board")
	}
}

// TestRecoverBoard_NonLeadRefused — recovery IS a board write, so the
// sole-writer guard binds it: a non-lead session's recovery is refused and
// the board stays unknown.
func TestRecoverBoard_NonLeadRefused(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")
	if err := DeclareRole(root, "worker-sess", "run", "run-alpha"); err != nil {
		t.Fatalf("DeclareRole(worker): %v", err)
	}
	before := string(corruptBoard(t, root))

	res, err := RecoverBoard(root, "worker-sess")
	if err == nil {
		t.Fatalf("RecoverBoard(non-lead) err = nil, res=%+v — the guard must bind recovery too", res)
	}
	if !IsNotSoleWriter(err) {
		t.Fatalf("err = %v, want ErrNotSoleWriter", err)
	}
	if now := string(readBoardBytes(t, root)); now != before {
		t.Fatal("non-lead recovery modified the board")
	}
	if _, lerr := LoadBoard(root); !IsBoardUnknown(lerr) {
		t.Fatalf("board no longer unknown after refused recovery: %v", lerr)
	}
}

// TestRecoverBoard_AbsentNeedsNoRecovery — REQ-KB-021: an absent state file
// is a legitimately empty board requiring no recovery; invoking the operation
// on one reports recovered without fabricating a loss record.
func TestRecoverBoard_AbsentNeedsNoRecovery(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedLead(t, root, "lead-sess")

	res, err := RecoverBoard(root, "lead-sess")
	if err != nil {
		t.Fatalf("RecoverBoard(absent) error = %v", err)
	}
	if res.Verdict != RecoveryVerdictRecovered {
		t.Fatalf("verdict = %q, want %q — absence is not damage", res.Verdict, RecoveryVerdictRecovered)
	}
	if res.LostRecorded {
		t.Fatal("absent board recorded a loss — a first-run path requiring repair of something never damaged")
	}
}

// containsAll reports whether s contains every rune of sub in order (a
// containment check tolerant of JSON escaping differences for the raw bytes).
func containsAll(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for _, r := range sub {
		idx := indexRune(s, r)
		if idx < 0 {
			return false
		}
		s = s[idx+1:]
	}
	return true
}

// indexRune finds the first occurrence of r in s, or -1.
func indexRune(s string, r rune) int {
	for i, sr := range s {
		if sr == r {
			return i
		}
	}
	return -1
}
