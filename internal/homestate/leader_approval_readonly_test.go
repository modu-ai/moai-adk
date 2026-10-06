package homestate

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

// The read-only surface the backlog scan consumes (SPEC-FACTORY-COMPLETION-RECOVERY-001
// review P2-2) and the recorded dispatch binding (reviews P1 rounds 2-4),
// pinned at the homestate level.

func frPlaceRun(t *testing.T, db *FactoryDB, runID, status, created string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,manifest_json,created_at,updated_at) VALUES(?,?,'{}',?,?) ON CONFLICT(run_id) DO NOTHING`, runID, status, created, created); err != nil {
		t.Fatalf("place run %s: %v", runID, err)
	}
}

func frBindDispatch(t *testing.T, db *FactoryDB, cardID, runID string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO card_dispatch(card_id,run_id,recorded_at) VALUES(?,?,?) ON CONFLICT(card_id) DO UPDATE SET run_id=excluded.run_id,recorded_at=excluded.recorded_at`, cardID, runID, "2026-09-26T00:00:00Z"); err != nil {
		t.Fatalf("bind dispatch %s -> %s: %v", cardID, runID, err)
	}
}

// The recorded dispatch binding is the only current-run authority: it beats
// a newer modification time on another run's row, and a card with no
// binding is refused rather than guessed from any derivative signal.
func TestRecordedCardRowFollowsDispatchBinding(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	// run-old's row carries the newest updated_at (a worktree re-record
	// bumped it); the binding names run-cur.
	frPlaceRun(t, db, "run-old", "retired", "2026-09-25T00:00:00Z")
	frPlaceRun(t, db, "run-cur", "active", "2026-09-26T00:00:00Z")
	frPlace(t, db, Card{RunID: "run-old", CardID: "t1", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-old", UpdatedAt: "2026-09-26T03:00:00Z"})
	frPlace(t, db, Card{RunID: "run-cur", CardID: "t1", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 2, EvidenceSHA: "sha-cur", UpdatedAt: "2026-09-26T01:00:00Z"})
	frBindDispatch(t, db, "t1", "run-cur")

	row, linked, err := db.RecordedCardRowReadonly(ctx, "t1")
	if err != nil || !linked {
		t.Fatalf("binding resolution: linked=%v err=%v, want linked", linked, err)
	}
	if row.RunID != "run-cur" || row.Version != 2 || row.EvidenceSHA != "sha-cur" {
		t.Fatalf("row = %s v%d sha=%s, want run-cur v2 sha-cur (the recorded dispatch)", row.RunID, row.Version, row.EvidenceSHA)
	}

	// A card with neither factory rows nor a binding is not linked.
	if _, linked, err := db.RecordedCardRowReadonly(ctx, "absent"); err != nil || linked {
		t.Fatalf("absent card: linked=%v err=%v, want not linked", linked, err)
	}

	// Factory rows with no binding: the dispatch cannot be determined —
	// refuse rather than guess on modification time.
	frPlace(t, db, Card{RunID: "run-ghost-a", CardID: "t2", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 1, UpdatedAt: "2026-09-26T05:00:00Z"})
	if _, _, err := db.RecordedCardRowReadonly(ctx, "t2"); !errors.Is(err, ErrApprovalRunUnresolvable) {
		t.Fatalf("unbound rows err = %v, want ErrApprovalRunUnresolvable", err)
	}

	// A dangling binding (the named run lost its row) is also unresolvable.
	frPlace(t, db, Card{RunID: "run-two", CardID: "t3", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 1, UpdatedAt: "2026-09-26T05:00:00Z"})
	frBindDispatch(t, db, "t3", "run-gone")
	if _, _, err := db.RecordedCardRowReadonly(ctx, "t3"); !errors.Is(err, ErrApprovalRunUnresolvable) {
		t.Fatalf("dangling binding err = %v, want ErrApprovalRunUnresolvable", err)
	}
}

func TestVerifyApprovalReadonlyPaths(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	frPlaceRun(t, db, "run-two", "active", "2026-09-26T00:00:00Z")
	c := Card{RunID: frRun, CardID: "t1", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-t1"}
	frPlace(t, db, c)
	frBindDispatch(t, db, "t1", frRun)
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-t1", RunID: frRun, CardID: "t1", FactoryVersion: 1,
		EvidenceHash: "sha-t1", Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
	})

	if err := db.VerifyApprovalReadonly(ctx, "t1", "uuid-t1"); err != nil {
		t.Fatalf("matching receipt refused: %v", err)
	}
	// A receipt is keyed by the card's uuid: a foreign uuid finds no
	// receipt at all — the refusal is absence, which is the fail-closed
	// answer for a receipt that does not bind this card.
	if err := db.VerifyApprovalReadonly(ctx, "t1", "uuid-other"); !errors.Is(err, ErrApprovalMissing) {
		t.Fatalf("foreign uuid err = %v, want ErrApprovalMissing", err)
	}
	if err := db.VerifyApprovalReadonly(ctx, "absent", "uuid-t1"); !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("absent card err = %v, want ErrCardNotFound", err)
	}
	// The stale receipt: a bumped row against the version-1 approval.
	frPlace(t, db, Card{RunID: "run-two", CardID: "t2", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 2, EvidenceSHA: "sha-t2"})
	frBindDispatch(t, db, "t2", "run-two")
	frApprove(t, db, LeaderApproval{
		CardUUID: "uuid-t2", RunID: "run-two", CardID: "t2", FactoryVersion: 1,
		EvidenceHash: "sha-t2", Issuer: "lead", IssuerRole: ApprovalIssuerLeader,
	})
	if err := db.VerifyApprovalReadonly(ctx, "t2", "uuid-t2"); !errors.Is(err, ErrApprovalStale) {
		t.Fatalf("stale receipt err = %v, want ErrApprovalStale", err)
	}
}

// The read-only open never writes: no DDL, no migration, no file creation.
func TestOpenFactoryReadonly(t *testing.T) {
	db := openSandboxFactory(t)
	path := db.Path
	if _, err := OpenFactoryReadonly(filepath.Join(filepath.Dir(path), "missing.db")); err == nil {
		t.Fatal("a missing database opened read-only instead of erroring")
	}

	// A store whose leader_approvals table was dropped (an older schema
	// shape) opens as it stands and reports the table absent; writes are
	// refused by the query_only pragma, and opening creates nothing.
	if _, err := db.DB.Exec(`DROP TABLE leader_approvals`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	ro, err := OpenFactoryReadonly(path)
	if err != nil {
		t.Fatalf("open read-only: %v", err)
	}
	defer func() { _ = ro.Close() }()
	ctx := context.Background()
	present, err := ro.FactoryTablePresent(ctx, "leader_approvals")
	if err != nil {
		t.Fatalf("table present check: %v", err)
	}
	if present {
		t.Fatal("the dropped table read as present")
	}
	if present, err := ro.FactoryTablePresent(ctx, "cards"); err != nil || !present {
		t.Fatalf("cards table present=%v err=%v, want true", present, err)
	}
	if _, err := ro.FactoryTablePresent(ctx, "sqlite_master"); err == nil {
		t.Fatal("an unknown table name was accepted")
	}
	if _, err := ro.DB.ExecContext(ctx, `CREATE TABLE sneaky(x)`); err == nil {
		t.Fatal("the read-only handle accepted a write")
	}
}
