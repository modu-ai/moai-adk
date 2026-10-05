package homestate

import (
	"context"
	"errors"
	"testing"
	"time"
)

// RenewLease refuses degenerate renewals the same way Transition does: an
// unknown card, a card that holds no lease, and — after the clock passes the
// expiry — the lease itself, returning the card to assigned first.
func TestRenewLeaseRefusalsAndExpiry(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	frRegisterWorker(t, db, "worker-1")

	if _, err := db.RenewLease(ctx, frRun, "missing", "worker-1", frNow); !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("renew missing card: err = %v, want ErrCardNotFound", err)
	}

	assigned := Card{RunID: frRun, CardID: "plain", State: CardAssigned, Version: 1, OwnerLabel: "worker-1", WorktreePath: repo.Dir}
	frPlace(t, db, assigned)
	if _, err := db.RenewLease(ctx, frRun, "plain", "worker-1", frNow); !errors.Is(err, ErrLeaseHolder) {
		t.Fatalf("renew unleased card: err = %v, want ErrLeaseHolder", err)
	}

	expired := frLeasedCard(repo, "stale", CardRun)
	expired.EvidenceSHA = repo.Commit
	expired.LeaseExpiresAt = frNow.Add(-time.Minute).Format(time.RFC3339Nano)
	frPlace(t, db, expired)
	if _, err := db.RenewLease(ctx, frRun, "stale", "worker-1", frNow); !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("renew expired lease: err = %v, want ErrLeaseExpired", err)
	}
	got, err := db.LoadCard(ctx, frRun, "stale")
	if err != nil {
		t.Fatal(err)
	}
	if got.State != CardAssigned || got.LeaseHolder != "" {
		t.Fatalf("card after expired renewal = %s holder=%q, want assigned with no holder", got.State, got.LeaseHolder)
	}

	// A zero clock means "now": the renewal still stamps a fresh expiry. The
	// card's lease must not already be expired under the real clock, so it is
	// placed with a future expiry rather than the fixed fixture clock.
	live := frLeasedCard(repo, "fresh", CardRun)
	live.EvidenceSHA = repo.Commit
	live.LeaseExpiresAt = time.Now().Add(time.Hour).Format(time.RFC3339Nano)
	frPlace(t, db, live)
	renewed, err := db.RenewLease(ctx, frRun, "fresh", "worker-1", time.Time{})
	if err != nil {
		t.Fatalf("renew with zero clock: %v", err)
	}
	expiry, perr := time.Parse(time.RFC3339Nano, renewed.LeaseExpiresAt)
	if perr != nil {
		t.Fatalf("zero-clock renewal expiry %q unparseable: %v", renewed.LeaseExpiresAt, perr)
	}
	if !expiry.After(time.Now()) {
		t.Fatalf("zero-clock renewal expiry %q is not in the future", renewed.LeaseExpiresAt)
	}
}

// ListCards fails loudly on a broken database: a closed connection fails the
// query, and a row whose version column holds non-numeric text fails the
// scan instead of being silently skipped.
func TestListCardsFailureModes(t *testing.T) {
	db := frOpen(t)
	if _, err := db.ListCards(context.Background(), frRun); err != nil {
		t.Fatalf("list on empty db: %v", err)
	}
	closed := &FactoryDB{DB: db.DB, Path: db.Path}
	_ = db.Close()
	if _, err := closed.ListCards(context.Background(), frRun); err == nil {
		t.Fatal("list on closed db: err = nil, want query failure")
	}

	db2 := frOpen(t)
	frPlace(t, db2, Card{RunID: frRun, CardID: "ok", State: CardAssigned, Version: 1})
	if _, err := db2.DB.Exec(`INSERT INTO cards (run_id,card_id,owner_label,state,version,evidence_path,updated_at) VALUES (?,?,?,?,?,?,?)`,
		frRun, "bad", "w", "assigned", "not-a-number", "", "2026-01-01T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if _, err := db2.ListCards(context.Background(), frRun); err == nil {
		t.Fatal("list with corrupt version column: err = nil, want scan failure")
	}
	_ = db2.Close()
}
