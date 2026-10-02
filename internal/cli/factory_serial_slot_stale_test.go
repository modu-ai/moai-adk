// factory_serial_slot_stale_test.go — card t1407: the serial slot read of
// `factory next` (REQ-TCD-008) must not count a card whose lease has expired
// as in flight. An expired lease is only collected lazily — the row keeps its
// lease-holding state until something touches it — so a state-only read held
// the slot for a lane that no longer exists, and every other lane's `next`
// ended on "no card is available" for a serial candidate.
package cli

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// fcSlotFixture builds a run whose queue is t1 (picked) plus `queued` queued
// serial cards, with t1 recorded as the given factory card. Both lanes are
// registered so either can lease.
func fcSlotFixture(t *testing.T, t1 homestate.Card, queued int) string {
	t.Helper()
	root, store := fcFixture(t)
	states := []kanban.BacklogState{kanban.BacklogStatePicked}
	for i := 0; i < queued; i++ {
		states = append(states, kanban.BacklogStateQueued)
	}
	fcQueue(t, store, states...)
	for i := range states {
		fcClassify(t, store, fmt.Sprintf("t%d", i+1), kanban.ClassPriorityNormal, false, kanban.ClassModeSerial)
	}
	t1.CardID = "t1"
	fcPlace(t, root, t1)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	return root
}

// fcDeadLaneCard is a card a vanished lane left in a lease-holding state.
func fcDeadLaneCard(state, expiresAt string) homestate.Card {
	return homestate.Card{
		State: state, OwnerLabel: "lane-9", LeaseHolder: "lane-9",
		Stage: homestate.CardRun, LeaseExpiresAt: expiresAt,
	}
}

// TestFactoryNextExpiredLeaseReleasesSerialSlot — every lease-holding
// implementation state with an expired lease stops holding the serial slot, so
// the next serial card is leased; the live lease that results then holds the
// slot against the third serial card (the fix must not open the slot wide).
func TestFactoryNextExpiredLeaseReleasesSerialSlot(t *testing.T) {
	expired := fcNow.Add(-time.Hour).Format(time.RFC3339Nano)
	for _, state := range []string{
		homestate.CardLeased, homestate.CardPlan, homestate.CardPlanAudit,
		homestate.CardRun, homestate.CardSync, homestate.CardSyncAudit,
	} {
		t.Run(state, func(t *testing.T) {
			root := fcSlotFixture(t, fcDeadLaneCard(state, expired), 2)
			ctx := context.Background()

			got, owned, err := factoryNextLeaseOnce(ctx, root, fcRun, "lane-1")
			if err != nil {
				t.Fatalf("lane-1 next: %v", err)
			}
			if !owned || got.CardID != "t2" || got.LeaseHolder != "lane-1" {
				t.Fatalf("lane-1 next = (%s, holder=%q, owned=%v), want t2 leased to lane-1 — an expired %s lease must not hold the serial slot", got.CardID, got.LeaseHolder, owned, state)
			}

			// t2's live lease holds the slot: t3 stays unleased.
			got2, owned2, err := factoryNextLeaseOnce(ctx, root, fcRun, "lane-2")
			if err != nil {
				t.Fatalf("lane-2 next: %v", err)
			}
			if owned2 {
				t.Fatalf("lane-2 leased %s while t2's live lease holds the serial slot", got2.CardID)
			}
		})
	}
}

// TestFactoryNextSerialSlotLeaseExpiryBoundary — the expiry boundary follows
// Card.LeaseExpired: an expiry at or before now frees the slot; a future
// expiry, and a lease with no expiry recorded, keep it held.
func TestFactoryNextSerialSlotLeaseExpiryBoundary(t *testing.T) {
	cases := []struct {
		name      string
		expiresAt string
		wantLease bool
	}{
		{"past", fcNow.Add(-time.Minute).Format(time.RFC3339Nano), true},
		{"exactly-now", fcNow.Format(time.RFC3339Nano), true},
		{"future", fcNow.Add(time.Minute).Format(time.RFC3339Nano), false},
		{"no-expiry-recorded", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := fcSlotFixture(t, fcDeadLaneCard(homestate.CardRun, tc.expiresAt), 1)
			got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
			if err != nil {
				t.Fatalf("next: %v", err)
			}
			if owned != tc.wantLease {
				t.Fatalf("expiry %q: owned=%v (card %q), want owned=%v", tc.expiresAt, owned, got.CardID, tc.wantLease)
			}
		})
	}
}

// TestFactoryNextAssignedSerialCardsHoldSlot_PendingRuling pins the CURRENT
// reading of a leader-assigned serial card — it holds the slot — as measured
// evidence for the operator ruling this card routes upward, not as endorsed
// behavior. Two serial cards assigned (no lease) to two lanes with nothing in
// flight leave each lane's own `next` refused: the 2-card shape of the run
// tm9i7y wedge (8 leader-assigned serial cards, every `factory next` refused).
// REQ-TCD-008 reads literally this way; narrowing it is the operator's call
// (OD-1), so a change here is a deliberate edit of this test, not a repair.
func TestFactoryNextAssignedSerialCardsHoldSlot_PendingRuling(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
	fcClassify(t, store, "t1", kanban.ClassPriorityNormal, false, kanban.ClassModeSerial)
	fcClassify(t, store, "t2", kanban.ClassPriorityNormal, false, kanban.ClassModeSerial)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1"},
		homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-2"},
	)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")

	for _, lane := range []string{"lane-1", "lane-2"} {
		got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, lane)
		if err != nil {
			t.Fatalf("%s next: %v", lane, err)
		}
		if owned {
			t.Fatalf("%s leased %s; the current reading is that assigned serial cards hold the slot against each other", lane, got.CardID)
		}
	}
}

// TestFactoryNextPickedOwnerlessRowHoldsSlot_OutOfExpiryScope measures the
// boundary of the expiry repair: a `picked` row with no owner and no lease —
// what a failed claim leaves behind, and the shape of t810 in run tm9i7y — is
// not a lease-holding state, so lease expiry never reaches it and it keeps
// holding the serial slot. The queue side of t1 is blocked (the operator-held
// shape), so no lane can take t1 itself and t2 stays refused. Whether such a
// row should release the slot is the same ruling as the assigned case
// (REQ-TCD-008), not part of this card's expiry repair.
func TestFactoryNextPickedOwnerlessRowHoldsSlot_OutOfExpiryScope(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStateQueued, kanban.BacklogStateQueued)
	fcClassify(t, store, "t1", kanban.ClassPriorityNormal, true, kanban.ClassModeSerial)
	fcClassify(t, store, "t2", kanban.ClassPriorityNormal, false, kanban.ClassModeSerial)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
	sdRegisterLane(t, root, "lane-1")

	got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if owned {
		t.Fatalf("lane-1 leased %s; a picked ownerless row is outside lease expiry and keeps holding the slot", got.CardID)
	}
}
