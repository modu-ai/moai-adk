package homestate

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// ReassignCard (card t1696, defect D2) moves an `assigned` card to another
// REGISTERED lane when no live lease stands: lanes that relaunch return under
// fresh labels and never reuse the old one, so a card owned by a dead label
// was unreachable — every actor failed the holder guard and `assign --to`
// only moves a picked card.

func raRegisterWorker(t *testing.T, db *FactoryDB, label string) {
	t.Helper()
	stamp := frNow.Format(time.RFC3339Nano)
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES(?,0,?,?)`, label, stamp, stamp); err != nil {
		t.Fatalf("register worker %s: %v", label, err)
	}
}

func TestReassignCardMovesAnAssignedCardToARegisteredLane(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, Card{RunID: frRun, CardID: "ra1", State: CardAssigned, OwnerLabel: "lane-9", Version: 3})
	raRegisterWorker(t, db, "lane-15")
	var mirror Card
	card, err := db.ReassignCard(context.Background(), frRun, "ra1", "lane-15", "assign --reassign", frNow,
		func(next Card) error { mirror = next; return nil })
	if err != nil {
		t.Fatalf("reassign: %v", err)
	}
	if card.OwnerLabel != "lane-15" || card.Version != 4 || card.State != CardAssigned {
		t.Fatalf("card = %+v, want assigned to lane-15 at version 4", card)
	}
	if mirror.OwnerLabel != "lane-15" {
		t.Fatalf("beforeCommit saw %+v, want the reassigned row", mirror)
	}
	var payload string
	if err := db.DB.QueryRow(`SELECT payload_json FROM events WHERE kind='card.reassign' ORDER BY seq DESC LIMIT 1`).Scan(&payload); err != nil {
		t.Fatalf("read the reassign event: %v", err)
	}
	if !strings.Contains(payload, `"prev_owner":"lane-9"`) || !strings.Contains(payload, `"new_owner":"lane-15"`) {
		t.Fatalf("the reassign event lost the owner trail: %s", payload)
	}
}

func TestReassignCardRefusals(t *testing.T) {
	cases := []struct {
		name    string
		card    Card
		target  string
		wantErr error
		wantMsg string
	}{
		{
			name:    "a live lease refuses",
			card:    Card{RunID: frRun, CardID: "ra2", State: CardLeased, OwnerLabel: "lane-9", Version: 1, LeaseHolder: "lane-9", LeaseExpiresAt: frNow.Add(24 * time.Hour).Format(time.RFC3339Nano)},
			target:  "lane-15",
			wantErr: ErrLeaseHolder,
		},
		{
			name:    "an unregistered target refuses",
			card:    Card{RunID: frRun, CardID: "ra3", State: CardAssigned, OwnerLabel: "lane-9", Version: 1},
			target:  "lane-99",
			wantErr: ErrInvalidCardInput,
			wantMsg: "not a registered lane",
		},
		{
			name:    "the same owner refuses",
			card:    Card{RunID: frRun, CardID: "ra4", State: CardAssigned, OwnerLabel: "lane-9", Version: 1},
			target:  "lane-9",
			wantErr: ErrInvalidCardInput,
			wantMsg: "already assigned to",
		},
		{
			name:    "a picked card refuses",
			card:    Card{RunID: frRun, CardID: "ra5", State: CardPicked, Version: 1},
			target:  "lane-15",
			wantErr: ErrIllegalTransition,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := frOpen(t)
			frPlace(t, db, tc.card)
			raRegisterWorker(t, db, "lane-15")
			_, err := db.ReassignCard(context.Background(), frRun, tc.card.CardID, tc.target, "assign --reassign", frNow, nil)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if tc.wantMsg != "" && !strings.Contains(err.Error(), tc.wantMsg) {
				t.Fatalf("err = %v, want it to name %q", err, tc.wantMsg)
			}
		})
	}
}

// An EXPIRED lease on a lease-holding card is the dead-lane shape itself
// (card t1696 defect D4, lane-17's state=run): the reassignment lands the
// expiry like T27/T28 — assigned, holder cleared — under the new owner.
func TestReassignCardAdmitsAnExpiredLease(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, Card{RunID: frRun, CardID: "ra6", State: CardRun, OwnerLabel: "lane-9", Version: 2, LeaseHolder: "lane-9", LeaseExpiresAt: frNow.Add(-24 * time.Hour).Format(time.RFC3339Nano), Stage: CardRun})
	raRegisterWorker(t, db, "lane-15")
	card, err := db.ReassignCard(context.Background(), frRun, "ra6", "lane-15", "assign --reassign", frNow, nil)
	if err != nil {
		t.Fatalf("reassign: %v", err)
	}
	if card.OwnerLabel != "lane-15" || card.State != CardAssigned || card.LeaseHolder != "" {
		t.Fatalf("card = %s/%s holder %q, want assigned/lane-15 with no holder", card.State, card.OwnerLabel, card.LeaseHolder)
	}
	if card.Stage != CardRun {
		t.Fatalf("stage = %q, want it kept", card.Stage)
	}
}

// An assigned card's holder string is stale residue (the state holds no
// lease): the reassignment clears it and swaps the owner.
func TestReassignCardClearsAnAssignedRowsResidueHolder(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, Card{RunID: frRun, CardID: "ra7", State: CardAssigned, OwnerLabel: "lane-9", Version: 2, LeaseHolder: "lane-9", LeaseExpiresAt: frNow.Add(-24 * time.Hour).Format(time.RFC3339Nano)})
	raRegisterWorker(t, db, "lane-15")
	card, err := db.ReassignCard(context.Background(), frRun, "ra7", "lane-15", "assign --reassign", frNow, nil)
	if err != nil {
		t.Fatalf("reassign: %v", err)
	}
	if card.OwnerLabel != "lane-15" || card.LeaseHolder != "" {
		t.Fatalf("card = %s/%s holder %q, want lane-15 with the residue cleared", card.State, card.OwnerLabel, card.LeaseHolder)
	}
}
