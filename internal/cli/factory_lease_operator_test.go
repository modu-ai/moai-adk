// factory_lease_operator_test.go — SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458)
// AC-FAL-002 (an operator write cannot interleave with a nominated lease),
// AC-FAL-003 (the same for bare arm (c), and arm (a)'s stated exception) and
// AC-FAL-004 (the compensation keeps the operator's fresh pick).
//
// The operator write is started in a GOROUTINE from inside the seam, so it can
// block on the section's lock without deadlocking the lease, and each test
// records whether that write COMPLETED before the seam returned (the window of
// ledger L11, whose positive control shows 400 ms tells a held lock from a free
// one). The goroutine is always joined after the verb returns.
package cli

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// flWindow is the operator-write window of ledger L11.
const flWindow = 400 * time.Millisecond

// TestFactoryLeaseOperatorWriteWaitsForSection — AC-FAL-002: (i) an operator
// `hold` of the nominee started from the nomination seam has not completed when
// the seam returns; (ii) after the verb returns the hold is applied (queue
// `hold`, not lost); (iii) a hold committed before the verb starts is refused
// `held`, writes no record row and changes nothing in the queue.
func TestFactoryLeaseOperatorWriteWaitsForSection(t *testing.T) {
	t.Run("hold during the section waits and is applied", func(t *testing.T) {
		root, store := nmQueuedNominee(t)
		nmIsolatedWorktrees(t, "t1")
		var op *flOp
		inside := false
		nmSetSeam(t, func(cardID string) error {
			op = flStartState(store, cardID, factory.BacklogStateHold)
			inside = op.within(flWindow)
			return nil
		})
		_, _, verbErr := qasRunNext(t, "--run", fcRun, "--card", "t1")
		if op == nil {
			t.Fatal("the nomination seam was never reached")
		}
		flJoin(t, []*flOp{op})
		q := nmQueueState(t, store, "t1")
		state, holder := flRow(t, root, "t1")
		t.Logf("completed inside section = %v operator-write-err=%v after the verb: queue=%s record=%s holder=%s verb-err=%v", inside, op.err, q, state, holder, verbErr)
		if inside {
			t.Errorf("clause (i) not met: the operator write completed inside the section (the seam returned only after the write finished)")
		}
		if op.err != nil {
			t.Errorf("the operator write failed: %v", op.err)
		}
		if q != factory.BacklogStateHold {
			t.Errorf("clause (ii) not met: queue state after the verb = %s, want hold (the operator's write was lost)", q)
		}
	})
	t.Run("hold before the verb is refused held", func(t *testing.T) {
		root, store := nmQueuedNominee(t)
		nmSetState(t, store, "t1", factory.BacklogStateHold)
		before := nmSnapshot(t, root, store)
		out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
		nmAssertRefused(t, out, stderr, err, "held")
		if n := nmRowCount(t, root, "t1"); n != 0 {
			t.Errorf("a refused nomination wrote %d record rows, want 0", n)
		}
		if after := nmSnapshot(t, root, store); after != before {
			t.Errorf("a refused nomination changed the stores")
		}
	})
}

// TestFactoryLeaseArmCOperatorHold — AC-FAL-003 clauses (i) and (ii): (i) an
// operator `hold` of the card arm (c) promoted, started at arm (c)'s claim
// point (the factoryLeaseBeforeClaim seam), has not completed there and is
// applied after the verb returns; (ii) a hold committed before the verb starts
// makes arm (c) skip that card and lease the next one.
func TestFactoryLeaseArmCOperatorHold(t *testing.T) {
	t.Run("hold at the claim point waits and is applied", func(t *testing.T) {
		root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
		nmLaneEnv(t, "lane-1", "")
		nmIsolatedWorktrees(t, "t1", "t2")
		var mu sync.Mutex
		var op *flOp
		inside := false
		prev := factoryLeaseBeforeClaim
		factoryLeaseBeforeClaim = func(arm, cardID string) error {
			mu.Lock()
			defer mu.Unlock()
			if arm == "c" && op == nil {
				op = flStartState(store, cardID, factory.BacklogStateHold)
				inside = op.within(flWindow)
			}
			return nil
		}
		t.Cleanup(func() { factoryLeaseBeforeClaim = prev })
		_, _, verbErr := qasRunNext(t, "--run", fcRun)
		mu.Lock()
		got := op
		mu.Unlock()
		if got == nil {
			t.Fatal("arm (c)'s claim point was never reached")
		}
		flJoin(t, []*flOp{got})
		q := nmQueueState(t, store, "t1")
		state, holder := flRow(t, root, "t1")
		t.Logf("completed at claim point = %v operator-write-err=%v after the verb: queue=%s record=%s holder=%s verb-err=%v", inside, got.err, q, state, holder, verbErr)
		if inside {
			t.Errorf("clause (i) not met: the operator write completed at the claim point (arm (c) held no queue lock there)")
		}
		if q != factory.BacklogStateHold {
			t.Errorf("clause (i) not met: queue state after the verb = %s, want hold (applied after the verb)", q)
		}
	})
	t.Run("hold before the verb makes arm (c) lease the next card", func(t *testing.T) {
		root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateQueued)
		nmSetState(t, store, "t1", factory.BacklogStateHold)
		nmLaneEnv(t, "lane-1", "")
		nmIsolatedWorktrees(t, "t1", "t2")
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err != nil {
			t.Fatalf("bare next: %v (stderr %q)", err, stderr)
		}
		if head := nmLeasedHead(out); !strings.HasPrefix(head, "t2 stage=") {
			t.Errorf("bare next leased %q, want t2 (t1 is held)", head)
		}
		nmAssertLeased(t, root, "t2", "lane-1")
		if fcHasCard(t, root, "t1") {
			t.Errorf("the held card t1 gained a record row")
		}
		if q := nmQueueState(t, store, "t1"); q != factory.BacklogStateHold {
			t.Errorf("t1 queue state = %s, want still hold", q)
		}
	})
}

// TestFactoryLeaseArmAExcludesHeldAssignedCard — card t1516 (leader ruling
// 2026-10-05) supersedes AC-FAL-003 clause (iii)'s pin of spec §F R17: a card
// whose queue item is `hold` and whose row is `assigned` to the lane is no
// longer leased by arm (a) — the queue item's current state gates the row's
// lease edge. The verb ends on the no-card answer with the queue item still
// hold and the row still assigned, version unchanged.
func TestFactoryLeaseArmAExcludesHeldAssignedCard(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStatePicked)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun})
	nmSetState(t, store, "t1", factory.BacklogStateHold)
	before := fcCard(t, root, "t1")
	nmLaneEnv(t, "lane-1", "")

	out, _, err := qasRunNext(t, "--run", fcRun)
	sdExit3(t, "an assigned row under a held queue item", err)
	if !strings.Contains(out, "no card is available") {
		t.Errorf("stdout = %q, want `no card is available`", out)
	}
	if q := nmQueueState(t, store, "t1"); q != factory.BacklogStateHold {
		t.Errorf("t1 queue state = %s, want still hold", q)
	}
	if got := fcCard(t, root, "t1"); got.State != homestate.CardAssigned || got.OwnerLabel != "lane-1" || got.Version != before.Version {
		t.Errorf("t1 = %s owner=%q version=%d, want assigned/lane-1 version=%d unchanged", got.State, got.OwnerLabel, got.Version, before.Version)
	}
}

// Re-promoting a queued assigned card reuses its original ordering hints.
// A newly held hub peer must not rewrite that already-assigned record.
func TestFactoryNextQueuedAssignedCardPreservesHubHints(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued, factory.BacklogStateHold)
	for _, id := range []string{"t1", "t2"} {
		fbSeedFiles(t, store, id, "internal/template/catalog.yaml")
	}
	fcPlace(t, root,
		homestate.Card{CardID: "t0", State: homestate.CardMergedLocal},
		homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Stage: homestate.CardRun, HintAfter: "t0"},
		homestate.Card{CardID: "t2", State: homestate.CardPicked},
	)
	before := fcCard(t, root, "t2")
	nmIsolatedWorktrees(t, "t1")
	nmLaneEnv(t, "lane-1", "")
	out, stderr, err := qasRunNext(t, "--run", fcRun)
	if err != nil {
		t.Fatalf("queued assigned card: %v (stderr %q)", err, stderr)
	}
	if head := nmLeasedHead(out); !strings.HasPrefix(head, "t1 stage=") {
		t.Fatalf("leased %q, want assigned t1", head)
	}
	nmAssertLeased(t, root, "t1", "lane-1")
	if got := fcCard(t, root, "t1"); got.HintAfter != "t0" || got.OwnerLabel != "lane-1" {
		t.Errorf("original assignment changed: after=%q owner=%q", got.HintAfter, got.OwnerLabel)
	}
	if got := nmQueueState(t, store, "t1"); got != factory.BacklogStatePicked {
		t.Errorf("promoted card queue=%s, want picked", got)
	}
	if got := nmQueueState(t, store, "t2"); got != factory.BacklogStateHold {
		t.Errorf("held peer queue=%s, want hold", got)
	}
	if got := fcCard(t, root, "t2"); got != before {
		t.Errorf("held peer changed: got %+v, want %+v", got, before)
	}
}

// TestFactoryLeaseCompensationKeepsOperatorPick — AC-FAL-004: a queued nominee,
// a seam that fails the claim, and an operator goroutine started from the seam
// that writes the queue item to `queued` and then to `picked` (an unpick and a
// re-pick, as direct store writes). The invocation fails with the injected
// error (not exit 4); the operator's writes have not completed when the seam
// returns; after the verb the card's queue state is `picked`, the operator's
// last write, and there is no leased row.
func TestFactoryLeaseCompensationKeepsOperatorPick(t *testing.T) {
	root, store := nmQueuedNominee(t)
	nmIsolatedWorktrees(t, "t1")
	var op *flOp
	inside := false
	nmSetSeam(t, func(cardID string) error {
		op = &flOp{ch: make(chan error, 1)}
		go func() {
			if err := store.Mutate(func(r *factory.BacklogRecord) error {
				for i := range r.Items {
					if r.Items[i].ID == cardID {
						r.Items[i].State = factory.BacklogStateQueued
					}
				}
				return nil
			}); err != nil {
				op.ch <- err
				return
			}
			stamp := "2026-10-03T00:00:00Z"
			op.ch <- store.Mutate(func(r *factory.BacklogRecord) error {
				for i := range r.Items {
					if r.Items[i].ID == cardID {
						r.Items[i].State = factory.BacklogStatePicked
						r.Items[i].PickedAt = &stamp
					}
				}
				return nil
			})
		}()
		inside = op.within(flWindow)
		return errors.New(nmInjected)
	})
	_, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	if op == nil {
		t.Fatal("the nomination seam was never reached")
	}
	flJoin(t, []*flOp{op})
	q := nmQueueState(t, store, "t1")
	state, holder := flRow(t, root, "t1")
	t.Logf("operator writes completed inside section = %v operator-err=%v verb-err=%v exit=%d queue=%s record=%s holder=%s", inside, op.err, err, nmExit(err), q, state, holder)
	if err == nil || !strings.Contains(err.Error(), nmInjected) {
		t.Errorf("the invocation returned %v (stderr %q), want the injected seam failure", err, stderr)
	}
	if code := nmExit(err); code == 4 {
		t.Errorf("the invocation exited 4 (a refusal), want the injected error")
	}
	if inside {
		t.Errorf("the operator's writes completed inside the section (the seam returned only after they finished)")
	}
	if q != factory.BacklogStatePicked {
		t.Errorf("the operator's fresh pick was reverted by the compensation: queue=%s want picked", q)
	}
	if state == homestate.CardLeased {
		t.Errorf("the card has a leased row after a failed claim")
	}
}
