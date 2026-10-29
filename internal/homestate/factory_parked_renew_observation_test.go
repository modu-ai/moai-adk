package homestate

import (
	"context"
	"testing"
	"time"
)

// This records the runtime behavior separately from watchdog policy.
func TestParkedMergeReadyLeaseCanBeRenewed(t *testing.T) {
	db := frOpen(t)
	repo := frNewRepo(t, true)
	frRegisterWorker(t, db, "worker-1")
	card := frLeasedCard(repo, "parked", CardMergeReady)
	frPlace(t, db, card)
	got, err := db.RenewLease(context.Background(), frRun, "parked", "worker-1", frNow)
	if err != nil {
		t.Fatalf("renew merge-ready lease: %v", err)
	}
	if got.State != CardMergeReady || got.LeaseHolder != "worker-1" {
		t.Fatalf("renewed card: state=%s holder=%q", got.State, got.LeaseHolder)
	}
}

func TestRefusedAbandonPreservesExpiredLease(t *testing.T) {
	for _, decider := range []string{"", DeciderHuman} {
		t.Run("decider="+decider, func(t *testing.T) {
			db := frOpen(t)
			repo := frNewRepo(t, true)
			card := frLeasedCard(repo, "expired", CardRun)
			card.LeaseExpiresAt = frNow.Add(-time.Minute).Format(time.RFC3339Nano)
			frPlace(t, db, card)
			_, err := db.Transition(context.Background(), TransitionRequest{RunID: frRun, CardID: card.CardID, To: CardAbandoned, ExpectedVersion: card.Version, Actor: "worker-1", Decider: decider, Now: frNow})
			if err == nil {
				t.Fatal("abandon without human decider unexpectedly succeeded")
			}
			got, err := db.LoadCard(context.Background(), frRun, card.CardID)
			if err != nil {
				t.Fatal(err)
			}
			if got.State != card.State || got.LeaseHolder != card.LeaseHolder || got.LeaseExpiresAt != card.LeaseExpiresAt || got.Version != card.Version {
				t.Fatalf("refused abandon changed card: state=%s holder=%q expiry=%q version=%d", got.State, got.LeaseHolder, got.LeaseExpiresAt, got.Version)
			}
		})
	}
}
