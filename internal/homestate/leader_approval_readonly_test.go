package homestate

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

// The read-only surface the backlog scan consumes (SPEC-FACTORY-COMPLETION-RECOVERY-001
// review P2-2) and the dispatch-run resolution (reviews P1 round-2 and P1
// round-3), pinned at the homestate level.

func frPlaceRun(t *testing.T, db *FactoryDB, runID, status, created string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO runs(run_id,status,manifest_json,created_at,updated_at) VALUES(?,?,'{}',?,?)`, runID, status, created, created); err != nil {
		t.Fatalf("place run %s: %v", runID, err)
	}
}

// The dispatch binding the runs table carries is the anchor: an active run
// wins over a retired one even when the retired run's row was modified more
// recently, and a store that no longer describes any candidate run refuses
// to guess (ErrApprovalRunUnresolvable).
func TestRecordedCardRowPrefersActiveRun(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	// The card is recorded in TWO runs. run-old's row carries the newest
	// updated_at (a worktree re-record bumped it) and its run is retired;
	// run-cur is active.
	frPlaceRun(t, db, "run-old", "retired", "2026-09-25T00:00:00Z")
	frPlaceRun(t, db, "run-cur", "active", "2026-09-26T00:00:00Z")
	frPlace(t, db, Card{RunID: "run-old", CardID: "t1", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-old", UpdatedAt: "2026-09-26T03:00:00Z"})
	frPlace(t, db, Card{RunID: "run-cur", CardID: "t1", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 2, EvidenceSHA: "sha-cur", UpdatedAt: "2026-09-26T01:00:00Z"})

	row, linked, err := db.RecordedCardRowReadonly(ctx, "t1")
	if err != nil || !linked {
		t.Fatalf("active-run resolution: linked=%v err=%v, want linked", linked, err)
	}
	if row.RunID != "run-cur" || row.Version != 2 || row.EvidenceSHA != "sha-cur" {
		t.Fatalf("row = %s v%d sha=%s, want run-cur v2 sha-cur (the active dispatch)", row.RunID, row.Version, row.EvidenceSHA)
	}

	// A card with no factory row at all is not linked.
	if _, linked, err := db.RecordedCardRowReadonly(ctx, "absent"); err != nil || linked {
		t.Fatalf("absent card: linked=%v err=%v, want not linked", linked, err)
	}

	// Both candidate runs undescribed: the dispatch cannot be determined —
	// refuse rather than guess on modification time.
	frPlace(t, db, Card{RunID: "run-ghost-a", CardID: "t2", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 1, UpdatedAt: "2026-09-26T05:00:00Z"})
	frPlace(t, db, Card{RunID: "run-ghost-b", CardID: "t2", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 1, UpdatedAt: "2026-09-26T04:00:00Z"})
	if _, _, err := db.RecordedCardRowReadonly(ctx, "t2"); !errors.Is(err, ErrApprovalRunUnresolvable) {
		t.Fatalf("undescribed runs err = %v, want ErrApprovalRunUnresolvable", err)
	}
}

func TestVerifyApprovalReadonlyPaths(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	frPlaceRun(t, db, frRun, "active", "2026-09-25T00:00:00Z")
	frPlaceRun(t, db, "run-two", "active", "2026-09-26T00:00:00Z")
	c := Card{RunID: frRun, CardID: "t1", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-t1"}
	frPlace(t, db, c)
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
