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

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// fcSlotFixture builds a run whose queue is t1 (picked) plus `queued` queued
// serial cards, with t1 recorded as the given factory card. Both lanes are
// registered so either can lease.
func fcSlotFixture(t *testing.T, t1 homestate.Card, queued int) string {
	t.Helper()
	root, store := fcFixture(t)
	states := []factory.BacklogState{factory.BacklogStatePicked}
	for i := 0; i < queued; i++ {
		states = append(states, factory.BacklogStateQueued)
	}
	fcQueue(t, store, states...)
	for i := range states {
		fcClassify(t, store, fmt.Sprintf("t%d", i+1), factory.ClassPriorityNormal, false, factory.ClassModeSerial)
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

// TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState — the control for the
// expiry repair, across every lease-holding implementation state: a lease that
// has not expired keeps holding the slot, so a mutation that frees one of these
// states unconditionally fails here instead of passing on a single state.
func TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState(t *testing.T) {
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	for _, state := range []string{
		homestate.CardLeased, homestate.CardPlan, homestate.CardPlanAudit,
		homestate.CardRun, homestate.CardSync, homestate.CardSyncAudit,
	} {
		t.Run(state, func(t *testing.T) {
			root := fcSlotFixture(t, fcDeadLaneCard(state, live), 1)
			got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
			if err != nil {
				t.Fatalf("next: %v", err)
			}
			if owned {
				t.Fatalf("lane-1 leased %s while a live %s lease holds the serial slot", got.CardID, state)
			}
		})
	}
}

// TestFactoryNextFailedSerialRowReleasesSlot — `failed` is a terminal card
// state with no outgoing edge (homestate.IsTerminalCardState), so a serial row
// that ended there is not in flight; REQ-TCD-008 re-admits selection on the
// terminal states. The non-terminal parked states stay held: only `failed`
// changes (the controls pin `blocked` and `needs-decision`).
func TestFactoryNextFailedSerialRowReleasesSlot(t *testing.T) {
	cases := []struct {
		state     string
		wantLease bool
	}{
		{homestate.CardFailed, true},
		{homestate.CardBlocked, false},
		{homestate.CardNeedsDecision, false},
	}
	for _, tc := range cases {
		t.Run(tc.state, func(t *testing.T) {
			root := fcSlotFixture(t, homestate.Card{State: tc.state, OwnerLabel: "lane-9"}, 1)
			got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
			if err != nil {
				t.Fatalf("next: %v", err)
			}
			if owned != tc.wantLease {
				t.Fatalf("t1 at %s: owned=%v (card %q), want owned=%v", tc.state, owned, got.CardID, tc.wantLease)
			}
		})
	}
}

// TestFactoryNextOwnAssignedSerialCardLeasesPastSiblingAssigned — operator
// ruling 2026-10-03 (card t1407, option B): `assigned` still counts as holding
// the serial slot, but a lane leasing a card assigned TO ITSELF ignores sibling
// cards that are merely `assigned`. Two serial cards assigned to two lanes with
// nothing in flight no longer wedge each other (the 2-card shape of the run
// tm9i7y deadlock: 8 leader-assigned serial cards, every `factory next`
// refused). The first lane's lease then holds the slot against the second, so
// serial cards are still served one at a time.
func TestFactoryNextOwnAssignedSerialCardLeasesPastSiblingAssigned(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1"},
		homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-2"},
	)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	ctx := context.Background()

	got, owned, err := factoryNextLeaseOnce(ctx, root, fcRun, "lane-1")
	if err != nil {
		t.Fatalf("lane-1 next: %v", err)
	}
	if !owned || got.CardID != "t1" || got.LeaseHolder != "lane-1" {
		t.Fatalf("lane-1 next = (%s, holder=%q, owned=%v), want its own assigned t1 leased past the sibling assigned t2", got.CardID, got.LeaseHolder, owned)
	}

	// t1's live lease now holds the slot: lane-2's own assigned t2 waits.
	got2, owned2, err := factoryNextLeaseOnce(ctx, root, fcRun, "lane-2")
	if err != nil {
		t.Fatalf("lane-2 next: %v", err)
	}
	if owned2 {
		t.Fatalf("lane-2 leased %s while t1's live lease holds the serial slot", got2.CardID)
	}
}

// TestFactoryNextOwnAssignedSerialCardBlockedByLiveSerialLease — the control
// for the option-B repair: ignoring sibling `assigned` cards does not ignore a
// sibling that is actually in flight. A lane's own assigned serial card is not
// leased while another serial card holds a live lease.
func TestFactoryNextOwnAssignedSerialCardBlockedByLiveSerialLease(t *testing.T) {
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	root := fcSlotFixture(t, fcDeadLaneCard(homestate.CardRun, live), 1)
	fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-1"})

	got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if owned {
		t.Fatalf("lane-1 leased its assigned %s while t1 holds a live serial lease", got.CardID)
	}
}

// TestFactoryNextAssignedSerialCardStillHoldsSlotAgainstNewTakes — the second
// control: an `assigned` serial card keeps holding the slot against a lane that
// has no card of its own, so an unassigned lane still cannot take a NEW serial
// card (arms b/b2/c) while one is assigned.
func TestFactoryNextAssignedSerialCardStillHoldsSlotAgainstNewTakes(t *testing.T) {
	root := fcSlotFixture(t, homestate.Card{State: homestate.CardAssigned, OwnerLabel: "lane-1"}, 1)

	got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-2")
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if owned {
		t.Fatalf("lane-2 (no card of its own) took %s while t1 is assigned and holds the serial slot", got.CardID)
	}
}

// TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms — the third control:
// the option-B exception belongs to arm (a) only. A lane with no card of its
// own is refused when the next serial candidate comes through arm (b) (a
// recorded operator-picked card with no owner) or arm (b2) (a queue-picked
// card with no record row yet) while another serial card is merely assigned.
func TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms(t *testing.T) {
	for _, tc := range []struct {
		name         string
		recordSecond bool
	}{
		{"arm-b-recorded-picked-ownerless", true},
		{"arm-b2-queue-picked-unrecorded", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, store := fcFixture(t)
			fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
			fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
			fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
			rows := []homestate.Card{{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1"}}
			if tc.recordSecond {
				rows = append(rows, homestate.Card{CardID: "t2", State: homestate.CardPicked})
			}
			fcPlace(t, root, rows...)
			sdRegisterLane(t, root, "lane-2")

			got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-2")
			if err != nil {
				t.Fatalf("next: %v", err)
			}
			if owned {
				t.Fatalf("lane-2 took %s through a picked arm while t1 is assigned and holds the serial slot", got.CardID)
			}
		})
	}
}

// TestFactoryNextOwnAssignedSerialCardBlockedByPickedSibling — the option-B
// exception ignores sibling `assigned` rows and nothing else: a `picked` serial
// row still counts in arm (a), so a lane's own assigned serial card is not
// leased past it. (Both rows then wait on each other — arm b cannot take the
// picked row while the assigned one holds the slot; that residual belongs to
// the same ruling and is recorded in the amendment, not repaired here.) A
// mutation that ignores `picked` as well as `assigned` fails this test.
func TestFactoryNextOwnAssignedSerialCardBlockedByPickedSibling(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1"},
		homestate.Card{CardID: "t2", State: homestate.CardPicked},
	)
	sdRegisterLane(t, root, "lane-1")

	got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if owned {
		t.Fatalf("lane-1 leased its assigned %s past a picked serial sibling; only assigned siblings are ignored in arm (a)", got.CardID)
	}
}

// TestFactoryNextParallelizableLeasesBesideLiveSerial — the slot is a
// serial-only exclusivity: while a serial card holds a live lease, a
// parallelizable card is still leasable through arm (a) (assigned to the lane)
// and through arm (c) (promoted from the queue). Dropping the serial-mode check
// in either arm blocks the parallelizable card and fails the matching subtest.
func TestFactoryNextParallelizableLeasesBesideLiveSerial(t *testing.T) {
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	for _, tc := range []struct {
		name        string
		assignedRow bool
	}{
		{"arm-a-assigned-parallelizable", true},
		{"arm-c-queued-parallelizable", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, store := fcFixture(t)
			fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStateQueued)
			fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
			fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
			rows := []homestate.Card{fcDeadLaneCard(homestate.CardRun, live)}
			rows[0].CardID = "t1"
			if tc.assignedRow {
				rows = append(rows, homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-1"})
			}
			fcPlace(t, root, rows...)
			sdRegisterLane(t, root, "lane-1")

			got, owned, err := factoryNextLeaseOnce(context.Background(), root, fcRun, "lane-1")
			if err != nil {
				t.Fatalf("next: %v", err)
			}
			if !owned || got.CardID != "t2" {
				t.Fatalf("lane-1 next = (%s, owned=%v), want the parallelizable t2 leased beside the live serial t1", got.CardID, owned)
			}
		})
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
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, true, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
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
