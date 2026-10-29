package homestate

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestReacquireExpiredLeaseRejectsChangedGenerationAndRollsBack(t *testing.T) {
	for _, tc := range []string{"version", "renewed", "other owner", "new generation", "other run", "callback"} {
		t.Run(tc, func(t *testing.T) {
			db := frOpen(t)
			repo := frNewRepo(t, true)
			frRegisterWorker(t, db, "worker-1")
			card := frLeasedCard(repo, "c1", CardRun)
			card.LeaseExpiresAt = frNow.Add(-time.Minute).Format(time.RFC3339Nano)
			frPlace(t, db, card)
			if tc == "other run" {
				other := card
				other.RunID = "other-run"
				other.LeaseExpiresAt = frNow.Add(time.Hour).Format(time.RFC3339Nano)
				frPlace(t, db, other)
			}
			version, lane := card.Version, "worker-1"
			if tc == "version" {
				version++
			}
			if tc == "other owner" {
				lane = "worker-2"
			}
			if tc == "renewed" {
				if _, err := db.DB.Exec(`UPDATE cards SET lease_expires_at=? WHERE run_id=? AND card_id=?`, frNow.Add(time.Hour).Format(time.RFC3339Nano), frRun, card.CardID); err != nil {
					t.Fatal(err)
				}
			}
			if tc == "new generation" {
				if _, err := db.DB.Exec(`UPDATE workers SET registered_at=? WHERE label=?`, frNow.Add(time.Minute).Format(time.RFC3339Nano), "worker-1"); err != nil {
					t.Fatal(err)
				}
			}
			before, err := db.LoadCard(context.Background(), frRun, card.CardID)
			if err != nil {
				t.Fatal(err)
			}
			dump := frRowDump(t, db, frRun, card.CardID)
			var cb func(Card) error
			if tc == "callback" {
				cb = func(Card) error { return errors.New("mirror failure") }
			}
			if _, err := db.ReacquireExpiredLease(context.Background(), frRun, card.CardID, lane, version, frNow, cb); err == nil {
				t.Fatal("unexpected success")
			}
			after, err := db.LoadCard(context.Background(), frRun, card.CardID)
			if err != nil {
				t.Fatal(err)
			}
			if after != before {
				t.Fatalf("refusal changed card: before=%+v after=%+v", before, after)
			}
			if got := frRowDump(t, db, frRun, card.CardID); got != dump {
				t.Fatal("refusal changed card or events")
			}
		})
	}
}
