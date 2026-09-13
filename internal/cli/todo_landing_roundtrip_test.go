// todo_landing_roundtrip_test.go — card t665: the landing evidence `todo
// landed` records must survive `todo done` AND be readable afterwards.
//
// The reported symptom is an archive that lost traceability: eight cards were
// closed on 2026-09-12 after `landed --sha <sha>` recorded successfully, and
// every one of them archived with `landing=unknown`. The card asks two
// branches apart before any repair:
//
//	(a) does `landed` write somewhere `done` does not read, or
//	(b) does `done` never consult the field at all?
//
// Both are answered AT THE STORE rather than by reading the code: the SQL
// assertions below open the engine database directly and ask what is in the
// column, on the live row before the archive and on the archive row after it.
// A Go-level nil pointer cannot separate "never written" from "never read",
// which is exactly the distinction the two branches turn on.
//
// The round trip is the regression: record a SHA, archive the card, then ask
// the read surface for it. Asserting that "a landing field exists" would pass
// against the defect being reported.
package cli

import (
	"strings"
	"testing"
)

// TestLandingEvidenceSurvivesArchive_StoreLevel answers branch (a): the
// evidence `landed` writes and the row `done` archives are the same record in
// the same store, so a repair belongs on the READ side, not the write side.
func TestLandingEvidenceSurvivesArchive_StoreLevel(t *testing.T) {
	f := newLandedFixture(t)
	if _, _, err := runTodo(t, "add", "card whose landing is recorded"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, _, err := runTodo(t, "landed", "t1", "--sha", f.mentioning[0], "--ref", "HEAD"); err != nil {
		t.Fatalf("landed: %v", err)
	}

	// Before the archive: the live row carries the column.
	db := openQueueDB(t, f.store)
	var liveNull bool
	if err := db.QueryRow(`SELECT landing IS NULL FROM items WHERE id = 't1'`).Scan(&liveNull); err != nil {
		t.Fatalf("read live landing column: %v", err)
	}
	if liveNull {
		t.Fatalf("landing column is NULL on the live row after `landed --sha` — the write never landed")
	}

	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done: %v", err)
	}

	// After the archive: the SAME evidence rides on the archive row. If this
	// is NULL, branch (a) holds and the archive path drops the column.
	var archivedNull bool
	if err := db.QueryRow(`SELECT landing IS NULL FROM archived_items WHERE id = 't1'`).Scan(&archivedNull); err != nil {
		t.Fatalf("read archived landing column: %v", err)
	}
	if archivedNull {
		t.Fatalf("landing column is NULL on the archive row: `done` dropped the evidence `landed` wrote (branch a)")
	}
}

// TestLandingEvidenceRoundTrip is the GREEN criterion the card names:
// `landed --sha X` → `done` → `history` returns X.
func TestLandingEvidenceRoundTrip(t *testing.T) {
	f := newLandedFixture(t)
	if _, _, err := runTodo(t, "add", "card whose landing must round-trip"); err != nil {
		t.Fatalf("add: %v", err)
	}
	sha := f.mentioning[0]
	if _, _, err := runTodo(t, "landed", "t1", "--sha", sha, "--ref", "HEAD"); err != nil {
		t.Fatalf("landed: %v", err)
	}

	// The archiving act must not report `unknown` about a card whose landing
	// the operator recorded: `unknown` is the honest report of "no query
	// ran", and a recorded SHA is not the absence of an answer.
	doneOut, _, err := runTodo(t, "done", "t1")
	if err != nil {
		t.Fatalf("done: %v", err)
	}
	if strings.Contains(doneOut, "landing=unknown") {
		t.Errorf("done reported landing=unknown for a card carrying recorded evidence: %q", strings.TrimSpace(doneOut))
	}

	// The read surface over the archive must return the delivering SHA —
	// without it the archive cannot say which commit delivered the card,
	// which is the reported loss.
	histOut, _, err := runTodo(t, "history", "t1")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if !strings.Contains(histOut, sha) {
		t.Errorf("history t1 does not return the recorded SHA %s: %q", sha, strings.TrimSpace(histOut))
	}
}

// TestLandingBackfillPathForAlreadyClosedCard measures the recovery path for
// the eight cards that were archived before this repair, which is the card's
// third question: is `undone` → `landed` → `done` a path that actually
// restores the evidence, or does a closed card have to stay closed blind?
//
// The measurement is what this test contributes — the procedure it walks is
// the one the verdict records for the operator. Executing that procedure
// against the live queue is the lead's act, not this lane's, so the walk
// happens against the fixture queue only.
func TestLandingBackfillPathForAlreadyClosedCard(t *testing.T) {
	f := newLandedFixture(t)
	if _, _, err := runTodo(t, "add", "card closed before the repair"); err != nil {
		t.Fatalf("add: %v", err)
	}
	// Closed the way the eight were: no evidence recorded first.
	if _, _, err := runTodo(t, "done", "t1"); err != nil {
		t.Fatalf("done (pre-repair shape): %v", err)
	}

	// Step 1 — restore the card to the live queue.
	if _, _, err := runTodo(t, "undone", "t1"); err != nil {
		t.Fatalf("undone: %v — the backfill path does not open", err)
	}
	// Step 2 — record the evidence the closing act never had.
	sha := f.mentioning[1]
	if _, _, err := runTodo(t, "landed", "t1", "--sha", sha, "--ref", "HEAD"); err != nil {
		t.Fatalf("landed after undone: %v", err)
	}
	// Step 3 — close it again.
	doneOut, _, err := runTodo(t, "done", "t1")
	if err != nil {
		t.Fatalf("done after backfill: %v", err)
	}
	if !strings.Contains(doneOut, sha) {
		t.Errorf("re-closing after backfill did not report the recorded SHA: %q", strings.TrimSpace(doneOut))
	}

	histOut, _, err := runTodo(t, "history", "t1")
	if err != nil {
		t.Fatalf("history after backfill: %v", err)
	}
	if !strings.Contains(histOut, sha) {
		t.Errorf("history after backfill does not return %s: %q", sha, strings.TrimSpace(histOut))
	}
}
