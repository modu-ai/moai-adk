package homestate

// SPEC-HANDOFF-NEUTRAL-001 M1.2 — consumed-history fallback read tests.
//
// ReadLatestConsumedResume is the fallback source of `moai handoff show`
// (REQ-HN-001): the most recent status='consumed' row ordered by consumed_at
// DESC. Fixtures are set through the factory state machine itself
// (SaveResume → ClaimPendingResume → FinishResume), never by raw SQL inserts.

import (
	"context"
	"testing"
	"time"
)

func TestReadLatestConsumedResume_LatestConsumedAtWins(t *testing.T) {
	root := t.TempDir()
	db, err := OpenFactory(root)
	if err != nil {
		t.Fatalf("OpenFactory: %v", err)
	}
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	// First handoff: save → claim → consume.
	if err := db.SaveResume(ctx, ResumeHandoff{
		SchemaVersion: 1, SavedAt: time.Now().Add(-2 * time.Hour), DirectivesJSON: `{}`, Body: "first",
	}); err != nil {
		t.Fatalf("SaveResume first: %v", err)
	}
	row, ok, err := db.ClaimPendingResume(ctx, "token-a")
	if err != nil || !ok {
		t.Fatalf("claim first: ok=%v err=%v", ok, err)
	}
	if err := db.FinishResume(ctx, row.ID, "token-a", "consumed", "test"); err != nil {
		t.Fatalf("finish first: %v", err)
	}

	// Second handoff consumed strictly later (SQLite timestamp resolution is
	// RFC3339Nano, but a wall-clock gap guarantees ordering).
	time.Sleep(1100 * time.Millisecond)
	if err := db.SaveResume(ctx, ResumeHandoff{
		SchemaVersion: 1, SavedAt: time.Now().Add(-time.Hour), DirectivesJSON: `{}`, Body: "second",
	}); err != nil {
		t.Fatalf("SaveResume second: %v", err)
	}
	row2, ok, err := db.ClaimPendingResume(ctx, "token-b")
	if err != nil || !ok {
		t.Fatalf("claim second: ok=%v err=%v", ok, err)
	}
	if err := db.FinishResume(ctx, row2.ID, "token-b", "consumed", "test"); err != nil {
		t.Fatalf("finish second: %v", err)
	}

	latest, present, err := db.ReadLatestConsumedResume(ctx)
	if err != nil || !present {
		t.Fatalf("ReadLatestConsumedResume: present=%v err=%v", present, err)
	}
	if latest.Body != "second" {
		t.Errorf("latest consumed body: got %q, want %q", latest.Body, "second")
	}
	if latest.ConsumedAt == nil || latest.ConsumedAt.IsZero() {
		t.Error("latest consumed row must carry consumed_at")
	}
}

func TestReadLatestConsumedResume_EmptyAndNonConsumed(t *testing.T) {
	root := t.TempDir()
	db, err := OpenFactory(root)
	if err != nil {
		t.Fatalf("OpenFactory: %v", err)
	}
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	// Empty history: not-present, no error.
	if row, present, err := db.ReadLatestConsumedResume(ctx); err != nil || present || row != nil {
		t.Fatalf("empty db: row=%v present=%v err=%v", row, present, err)
	}

	// A pending (never consumed) row is NOT a fallback source.
	if err := db.SaveResume(ctx, ResumeHandoff{
		SchemaVersion: 1, SavedAt: time.Now(), DirectivesJSON: `{}`, Body: "still pending",
	}); err != nil {
		t.Fatalf("SaveResume: %v", err)
	}
	if row, present, err := db.ReadLatestConsumedResume(ctx); err != nil || present || row != nil {
		t.Fatalf("pending-only db: row=%v present=%v err=%v", row, present, err)
	}

	// A claimed (not yet consumed) row is not a fallback source either.
	if _, ok, err := db.ClaimPendingResume(ctx, "token-c"); err != nil || !ok {
		t.Fatalf("claim: ok=%v err=%v", ok, err)
	}
	if row, present, err := db.ReadLatestConsumedResume(ctx); err != nil || present || row != nil {
		t.Fatalf("claimed-only db: row=%v present=%v err=%v", row, present, err)
	}
}
