package homestate

import (
	"context"
	"errors"
	"testing"
	"time"
)

// SPEC-FACTORY-LEASE-RENEW-001 (card t1663), milestone M2, AC-FLR-004 (REQ-FLR-004; the
// refusal rule is decision-index Q3, ruling d-20261010T095731Z-5ddc). A renewal of an
// expired lease is refused with the expired-lease class and no write: the card keeps its
// state, holder, expiry, and version, and no event row is appended.
func TestRenewRefusesExpiredWithoutWrite(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	ctx := context.Background()
	frRegisterWorker(t, db, "worker-1")
	expired := frLeasedCard(repo, "stale", CardRun)
	expired.EvidenceSHA = repo.Commit
	expired.LeaseExpiresAt = frNow.Add(-2 * time.Minute).Format(time.RFC3339Nano)
	frPlace(t, db, expired)
	before, err := db.LoadCard(ctx, frRun, "stale")
	if err != nil {
		t.Fatal(err)
	}
	eventsBefore := renewEventCount(t, db)
	if _, err := db.RenewLease(ctx, frRun, "stale", "worker-1", frNow); !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("renew expired lease: err = %v, want ErrLeaseExpired", err)
	}
	after, err := db.LoadCard(ctx, frRun, "stale")
	if err != nil {
		t.Fatal(err)
	}
	if after.Version != before.Version || after.State != before.State ||
		after.LeaseHolder != before.LeaseHolder || after.LeaseExpiresAt != before.LeaseExpiresAt {
		t.Fatalf("refused renewal changed the card: v%d %s holder=%q expiry=%s → v%d %s holder=%q expiry=%s",
			before.Version, before.State, before.LeaseHolder, before.LeaseExpiresAt,
			after.Version, after.State, after.LeaseHolder, after.LeaseExpiresAt)
	}
	if n := renewEventCount(t, db); n != eventsBefore {
		t.Fatalf("refused renewal appended %d event rows, want none", n-eventsBefore)
	}
}

// renewEventCount counts the event rows of the factory record.
func renewEventCount(t *testing.T, db *FactoryDB) int {
	t.Helper()
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
