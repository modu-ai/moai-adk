package homestate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

	// A FAILED assignment must leave the binding exactly as it was (review
	// round-5 P1): the binding write lives in the T2 transaction, so a
	// transition refusal cannot strand the card on a new run.
	frPlace(t, db, Card{RunID: "run-cur", CardID: "t4", State: CardPicked, OwnerLabel: "worker-1", Version: 1, UpdatedAt: "2026-09-26T02:00:00Z"})
	frBindDispatch(t, db, "t4", "run-cur")
	if _, err := db.Transition(ctx, TransitionRequest{
		RunID: "run-cur", CardID: "t4", To: CardAssigned,
		ExpectedVersion: 1, Actor: "assign", Owner: "", Now: frNow,
	}); !errors.Is(err, ErrInvalidCardInput) {
		t.Fatalf("invalid assignment err = %v, want ErrInvalidCardInput", err)
	}
	row2, linked, err := db.RecordedCardRowReadonly(ctx, "t4")
	if err != nil || !linked || row2.RunID != "run-cur" || row2.Version != 1 {
		t.Fatalf("failed assignment changed the binding: linked=%v row=%+v err=%v", linked, row2, err)
	}

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

// Review round-8 P2 (card t1538): the read-only handle must not CHECKPOINT
// or otherwise physically modify the database WHILE IT IS OPEN — the scan's
// queries run against a hot WAL without merging it — and in the steady
// state (clean store, no WAL) its whole life is byte-for-byte inert. The
// -shm file is excluded from the hashes: it is SQLite's transient
// coordination metadata, recreated by every opener and carrying no durable
// content. Residual: modernc's driver checkpoints a hot WAL when the
// read-only handle CLOSES while it is the able closer — a content-preserving
// merge recorded as §E.2 residual risk, not a content change.
func TestOpenFactoryReadonlyNeverCheckpointsWAL(t *testing.T) {
	db := openSandboxFactory(t)
	path := db.Path
	frPlaceRun(t, db, frRun, "active", "2026-09-25T00:00:00Z")
	frPlace(t, db, Card{RunID: frRun, CardID: "hot", State: CardMergedLocal, OwnerLabel: "worker-1", Version: 1, EvidenceSHA: "sha-hot"})
	frBindDispatch(t, db, "hot", frRun)

	hash := func() string {
		var b strings.Builder
		for _, suffix := range []string{"", "-wal"} {
			raw, err := os.ReadFile(path + suffix)
			if err != nil {
				b.WriteString("absent;")
				continue
			}
			sum := sha256.Sum256(raw)
			b.WriteString(hex.EncodeToString(sum[:8]) + ";")
		}
		return b.String()
	}

	// Hot-WAL arm: a live writer keeps the WAL; the read-only handle's open
	// and queries leave db and wal untouched.
	before := hash()
	ro, err := OpenFactoryReadonly(path)
	if err != nil {
		t.Fatalf("open read-only beside a live writer: %v", err)
	}
	if _, linked, err := ro.RecordedCardRowReadonly(context.Background(), "hot"); err != nil || !linked {
		t.Fatalf("read-only read: linked=%v err=%v", linked, err)
	}
	if during := hash(); during != before {
		_ = ro.Close()
		t.Fatalf("the read-only handle modified the store during its life: before %s during %s", before, during)
	}

	// Steady-state arm: after the writer closes cleanly, the read-only
	// handle's whole life on the clean store is byte-for-byte inert.
	_ = db.Close()
	db2, err := OpenFactoryReadonly(path)
	if err != nil {
		t.Fatalf("open read-only on the clean store: %v", err)
	}
	if _, _, err := db2.RecordedCardRowReadonly(context.Background(), "hot"); err != nil {
		t.Fatalf("clean-store read: %v", err)
	}
	steadyBefore := hash()
	_ = db2.Close()
	if steadyAfter := hash(); steadyAfter != steadyBefore {
		t.Fatalf("the read-only handle modified the clean store: before %s after %s", steadyBefore, steadyAfter)
	}
}
