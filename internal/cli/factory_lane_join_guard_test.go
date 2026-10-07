// factory_lane_join_guard_test.go — card t1513: the serial slot is held by a
// recorded driver, not by a state alone. A join-time worktree-root refusal
// was designed here and WITHDRAWN (see the verdict §부록 C): the lease layer
// already refuses a tree-rooted lane through factoryAssertParentCheckout on
// every path, and a join-time refusal fires inside this repository's own
// worktree-rooted test runs, breaking the whole join family.
package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactorySerialSlotHeldRequiresADriver — the slot predicate matrix. A row
// without an owner holds nothing (the t1453 wedge: picked, ownerless,
// lease-less, refusing every lease for a day); every driven row keeps the
// pre-existing behavior.
func TestFactorySerialSlotHeldRequiresADriver(t *testing.T) {
	now := time.Now()
	live := now.Add(time.Hour).Format(time.RFC3339Nano)
	expired := now.Add(-time.Minute).Format(time.RFC3339Nano)
	cases := []struct {
		name string
		card homestate.Card
		want bool
	}{
		{"picked without owner or lease holds nothing", homestate.Card{CardID: "t1453", State: homestate.CardPicked}, false},
		{"picked with an owner holds", homestate.Card{CardID: "t1", State: homestate.CardPicked, OwnerLabel: "lane-5"}, true},
		{"assigned with an owner holds", homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-2"}, true},
		{"leased with a live lease holds", homestate.Card{CardID: "t3", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live}, true},
		{"an expired lease frees", homestate.Card{CardID: "t4", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: expired}, false},
		{"merge-ready frees", homestate.Card{CardID: "t5", State: homestate.CardMergeReady, OwnerLabel: "lane-1"}, false},
	}
	for _, tc := range cases {
		if got := factorySerialSlotHeld(tc.card, now); got != tc.want {
			t.Errorf("%s: held=%v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestFactoryNextIgnoresDriverlessPickedRow — the end-to-end release: with a
// picked, ownerless, lease-less serial row sitting in the run record (the
// measured t1453 shape), a lane's `factory next --card` leases instead of
// being refused serial-slot.
func TestFactoryNextIgnoresDriverlessPickedRow(t *testing.T) {
	root, _ := flSerialPair(t)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root, homestate.Card{CardID: "t1453", State: homestate.CardPicked})
	nmLaneEnv(t, "lane-1", "")
	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if err != nil {
		t.Fatalf("factory next refused with a driverless picked row in the record: %v stderr=%q", err, stderr)
	}
	if st, owner := flRow(t, root, "t1"); st != homestate.CardLeased {
		t.Fatalf("t1 state=%s owner=%q, want leased", st, owner)
	}
}

// TestFactoryNextStillRefusedBehindADrivenRow — the control arm: with an
// owned serial row in flight (no lease, owner recorded), the slot still holds
// and `factory next` is refused serial-slot.
func TestFactoryNextStillRefusedBehindADrivenRow(t *testing.T) {
	root, _ := flSerialPair(t)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root, homestate.Card{CardID: "t1453", State: homestate.CardPicked, OwnerLabel: "lane-2"})
	nmLaneEnv(t, "lane-1", "")
	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if err == nil {
		t.Fatalf("factory next leased behind a driven serial row: stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "refused serial-slot") {
		t.Fatalf("refusal is not serial-slot: %v stderr=%q", err, stderr)
	}
}

// TestFactoryNextRefusesAgainAfterLaterAssign — the reversal path the
// predicate comment promises (card t1513 review P3-3): an ownerless picked
// row releases the slot and a lease goes through; a later T2 assign sets the
// owner and the row holds the slot again.
func TestFactoryNextRefusesAgainAfterLaterAssign(t *testing.T) {
	root, _ := flSerialPair(t)
	nmIsolatedWorktrees(t, "t1", "t2")
	ctx := context.Background()
	now := time.Now()

	// The ownerless picked row: recorded the way `factory assign` without
	// --to leaves it.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	card, err := db.RecordPicked(ctx, fcRun, "t1453", homestate.CardFields{}, "assign", now)
	if err != nil {
		t.Fatalf("record picked: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close factory: %v", err)
	}

	nmLaneEnv(t, "lane-1", "")
	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
		t.Fatalf("an ownerless picked row must not refuse the lease: %v stderr=%q", err, stderr)
	}

	// T2 assign sets the owner — the row holds again.
	picked := fcCard(t, root, "t1453")
	db, err = homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("reopen factory: %v", err)
	}
	card, err = db.Transition(ctx, homestate.TransitionRequest{RunID: fcRun, CardID: "t1453", To: homestate.CardAssigned, ExpectedVersion: picked.Version, Actor: "assign", Owner: "lane-2", Now: now})
	if err != nil {
		t.Fatalf("assign transition: %v", err)
	}
	if card.State != homestate.CardAssigned || card.OwnerLabel != "lane-2" {
		t.Fatalf("row after assign: state=%s owner=%q", card.State, card.OwnerLabel)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close factory: %v", err)
	}

	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2"); err == nil {
		t.Fatalf("the assigned row must hold the slot again: stderr=%q", stderr)
	} else if !strings.Contains(stderr, "refused serial-slot") {
		t.Fatalf("refusal is not serial-slot: %v stderr=%q", err, stderr)
	}
}

// TestFactoryNextOwnAssignedCardIgnoresSiblingAssignments — arm (a) through
// the nominated door (card t1542, measured on run tmhxo0 2026-10-07): the
// lane's own card sits `assigned` to it and sibling serial rows are merely
// `assigned` to other lanes; `factory next --card` leases its own card
// instead of being refused serial-slot by rows nothing is driving. The
// unnominated arm has read it this way since t1407; the nominated arm passed
// false unconditionally and re-created the wedge through the --card door.
func TestFactoryNextOwnAssignedCardIgnoresSiblingAssignments(t *testing.T) {
	root, store := flSerialPair(t)
	// A nominated resume requires the operator-picked queue state; queued
	// plus an assigned row is the stale-row shape, refused as recorded.
	if _, _, err := runTodo(t, "next", "1"); err != nil {
		t.Fatalf("operator pick: %v", err)
	}
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1"},
		homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-2"})
	sibling := fcCard(t, root, "t2")
	nmLaneEnv(t, "lane-1", "")
	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
		t.Fatalf("factory next refused the lane's own assigned card behind sibling assignments: %v stderr=%q", err, stderr)
	}
	if st, owner := flRow(t, root, "t1"); st != homestate.CardLeased || owner != "lane-1" {
		t.Fatalf("t1 state=%s owner=%q, want leased by lane-1", st, owner)
	}
	if got := nmQueueState(t, store, "t1"); got != factory.BacklogStatePicked {
		t.Errorf("own resumed card queue=%s, want picked", got)
	}
	if got := nmQueueState(t, store, "t2"); got != factory.BacklogStateQueued {
		t.Errorf("sibling queue=%s, want unchanged queued", got)
	}
	if got := fcCard(t, root, "t2"); got != sibling {
		t.Errorf("sibling assignment changed: got %+v, want %+v", got, sibling)
	}
}

// TestFactoryNextLeasesBehindMergedPRPredecessor — github-flow's terminal
// delivery state counts as predecessor completion (card-review r1): a
// successor whose after-hint names a merged-pr card leases instead of being
// skipped forever. The successor rides the operator-picked shape (a row at
// picked with no owner), the shape a hint-carrying record starts from.
func TestFactoryNextLeasesBehindMergedPRPredecessor(t *testing.T) {
	root, store := flSerialPair(t)
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == "t2" {
				r.Items[i].State = factory.BacklogStatePicked
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("set t2 queue state: %v", err)
	}
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardMergedPR, OwnerLabel: "lane-1"},
		homestate.Card{CardID: "t2", State: homestate.CardPicked, HintAfter: "t1"})
	nmLaneEnv(t, "lane-1", "")
	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t2"); err != nil {
		t.Fatalf("factory next refused behind a merged-pr predecessor: %v stderr=%q", err, stderr)
	}
	if st, owner := flRow(t, root, "t2"); st != homestate.CardLeased || owner != "lane-1" {
		t.Fatalf("t2 state=%s owner=%q, want leased by lane-1", st, owner)
	}
}

// TestFactoryNextForeignNominationStillHoldsBehindAssignments — the control
// arm: the arm-(a) read belongs to the lane's OWN card. A nomination of a
// card that is not assigned to the caller still counts sibling assigned rows
// and is refused serial-slot.
func TestFactoryNextForeignNominationStillHoldsBehindAssignments(t *testing.T) {
	root, _ := flSerialPair(t)
	nmIsolatedWorktrees(t, "t1", "t2")
	fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardAssigned, OwnerLabel: "lane-2"})
	nmLaneEnv(t, "lane-1", "")
	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err == nil {
		t.Fatalf("a foreign nomination leased behind sibling assigned rows: stderr=%q", stderr)
	} else if !strings.Contains(stderr, "refused serial-slot") {
		t.Fatalf("refusal is not serial-slot: %v stderr=%q", err, stderr)
	}
}
