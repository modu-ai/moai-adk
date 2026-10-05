package homestate

// card t1521 — the card↔worktree binding carry-over across a run
// replacement. When a factory run ends and a new run starts, the new run's
// card row is born with an empty worktree_path, so `factory next` refused the
// card's own surviving tree as a foreign tree (REQ-SD-011 firing on the
// card's own tree, observed 2026-10-05 on t1453). The carry-over needs a
// record read: the newest worktree binding recorded for the same card id in
// another run. The refusal itself stays untouched — this read only names
// what a previous run of the SAME card already recorded.

import (
	"context"
	"testing"
	"time"
)

// carryNow is the clock every RecordPicked fixture write below uses, so the
// rows carry deterministic updated_at values.
var carryNow = time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

// carryPick records cardID at picked for runID, with worktree when non-empty.
func carryPick(t *testing.T, db *FactoryDB, runID, cardID, worktree string) {
	t.Helper()
	fields := CardFields{}
	if worktree != "" {
		fields.WorktreePath = &worktree
	}
	if _, err := db.RecordPicked(context.Background(), runID, cardID, fields, "carry-test", carryNow); err != nil {
		t.Fatalf("record picked %s/%s: %v", runID, cardID, err)
	}
}

// TestPreviousCardWorktreeReadsNewestOtherRunBinding pins the read the
// carry-over is built on: the newest binding from another run wins, the
// current run's own row is excluded, and rows without a binding are skipped.
func TestPreviousCardWorktreeReadsNewestOtherRunBinding(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-a", Backend: "claude", LeadPID: 1, LeadProcessStart: "st"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-b", Backend: "claude", LeadPID: 2, LeadProcessStart: "st"})
	mustRecordRun(t, db, FactoryRun{RunID: "run-new", Backend: "claude", LeadPID: 3, LeadProcessStart: "st"})

	carryPick(t, db, "run-a", "t1", "/wt/run-a-t1")
	carryPick(t, db, "run-b", "t1", "/wt/run-b-t1")
	carryPick(t, db, "run-new", "t1", "")

	t.Run("newest other run's binding wins", func(t *testing.T) {
		got, ok, err := db.PreviousCardWorktree(context.Background(), "t1", "run-new")
		if err != nil {
			t.Fatalf("PreviousCardWorktree: %v", err)
		}
		if !ok || got != "/wt/run-b-t1" {
			t.Fatalf("PreviousCardWorktree(t1, run-new) = (%q, %v), want (/wt/run-b-t1, true)", got, ok)
		}
	})

	t.Run("excluded run's binding is skipped", func(t *testing.T) {
		got, ok, err := db.PreviousCardWorktree(context.Background(), "t1", "run-b")
		if err != nil {
			t.Fatalf("PreviousCardWorktree: %v", err)
		}
		if !ok || got != "/wt/run-a-t1" {
			t.Fatalf("PreviousCardWorktree(t1, run-b) = (%q, %v), want (/wt/run-a-t1, true)", got, ok)
		}
	})

	t.Run("card with no binding anywhere reads not-found", func(t *testing.T) {
		carryPick(t, db, "run-a", "t2", "")
		got, ok, err := db.PreviousCardWorktree(context.Background(), "t2", "run-new")
		if err != nil {
			t.Fatalf("PreviousCardWorktree: %v", err)
		}
		if ok || got != "" {
			t.Fatalf("PreviousCardWorktree(t2, run-new) = (%q, %v), want no binding", got, ok)
		}
	})

	t.Run("unknown card reads not-found", func(t *testing.T) {
		got, ok, err := db.PreviousCardWorktree(context.Background(), "t-none", "run-new")
		if err != nil {
			t.Fatalf("PreviousCardWorktree: %v", err)
		}
		if ok || got != "" {
			t.Fatalf("PreviousCardWorktree(t-none, run-new) = (%q, %v), want no binding", got, ok)
		}
	})
}
