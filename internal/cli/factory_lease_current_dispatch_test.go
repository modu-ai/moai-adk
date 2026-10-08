package cli

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// The queue's current-dispatch record follows EVERY path that moves a card's
// factory binding onto a run (turn-end gate relays #2-#3, card t1538, V2/F2):
// otherwise a retried older dispatch reads the stale record as the card's
// current engagement and drags the binding back. The paths only REFRESH a
// record the dispatch hook wrote — a queue that never had one keeps its
// schema untouched (AC-FR-021).

// fcQueueCurrent reads the queue's current-dispatch (run, owner) for a card.
func fcQueueCurrent(t *testing.T, store *factory.BacklogStore, cardID string) (run, owner string, ok bool) {
	t.Helper()
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range rec.Runtime.DispatchCurrent {
		if c.CardID == cardID {
			return c.RunID, c.OwnerLabel, true
		}
	}
	return "", "", false
}

func fcWantCurrent(t *testing.T, store *factory.BacklogStore, cardID, wantRun, wantOwner string) {
	t.Helper()
	run, owner, ok := fcQueueCurrent(t, store, cardID)
	if !ok || run != wantRun || owner != wantOwner {
		t.Fatalf("queue current dispatch for %s = (run %q, owner %q, recorded %v), want (run %q, owner %q)", cardID, run, owner, ok, wantRun, wantOwner)
	}
}

// seedOlderDispatch records an earlier engagement (run-old, lane-1) as the
// queue's current dispatch for a card, the way the dispatch hook does.
func seedOlderDispatch(t *testing.T, root string, store *factory.BacklogStore, cardID string) {
	t.Helper()
	if err := factory.RecordFactoryCardAssignment(root, "run-old", cardID, "lane-1", ""); err != nil {
		t.Fatal(err)
	}
	fcWantCurrent(t, store, cardID, "run-old", "lane-1")
}

// V2: a card ALREADY assigned in a run and leased through T3 re-points the
// factory binding to that run (the transition does it in its own
// transaction), so the queue's record must follow.
func TestFactoryNextLeaseOfAssignedCardRefreshesCurrentDispatch(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-1")
	// run-old: the card's earlier engagement — the binding and the queue's
	// current-dispatch record both name it.
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")
	// run-new already holds the card ASSIGNED to lane-1, placed ahead of its lease.
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-new", State: homestate.CardAssigned, OwnerLabel: "lane-1", Version: 2})
	ctx := context.Background()
	db := fcOpen(t, root)
	cur, err := db.LoadCard(ctx, "run-new", "t1")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, raced, err := factoryNextClaim(ctx, db, root, "run-new", cur, "lane-1"); err != nil || !ok || raced {
		t.Fatalf("lease of the assigned card: ok=%v raced=%v err=%v", ok, raced, err)
	}
	row, linked, err := db.RecordedCardRowReadonly(ctx, "t1")
	if err != nil || !linked || row.RunID != "run-new" {
		t.Fatalf("binding after the lease: run=%q linked=%v err=%v, want run-new", row.RunID, linked, err)
	}
	fcWantCurrent(t, store, "t1", "run-new", "lane-1")
}

// F2: every `factory assign` branch refreshes the current dispatch — the
// picked -> assigned edge, the --to-less record, and the state-preserving
// re-bind of a row past picked.
func TestFactoryAssignRecordsCurrentDispatchOnEveryBranch(t *testing.T) {
	t.Run("picked_to_assigned", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked)
		seedOlderDispatch(t, root, store, "t1")
		if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-2", "--run", "run-a"); err != nil {
			t.Fatal(err)
		}
		fcWantCurrent(t, store, "t1", "run-a", "lane-2")
	})
	t.Run("reassign_to_a_new_run", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked)
		seedOlderDispatch(t, root, store, "t1")
		if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-a"); err != nil {
			t.Fatal(err)
		}
		if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-2", "--run", "run-b"); err != nil {
			t.Fatal(err)
		}
		fcWantCurrent(t, store, "t1", "run-b", "lane-2")
	})
	t.Run("to_less_record", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked)
		seedOlderDispatch(t, root, store, "t1")
		if _, _, err := runFactory(t, "assign", "t1", "--run", "run-b"); err != nil {
			t.Fatal(err)
		}
		run, _, ok := fcQueueCurrent(t, store, "t1")
		if !ok || run != "run-b" {
			t.Fatalf("queue current dispatch run = %q (recorded %v), want run-b", run, ok)
		}
	})
	t.Run("state_preserving_rebind", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked)
		seedOlderDispatch(t, root, store, "t1")
		fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-c", State: homestate.CardAssigned, OwnerLabel: "lane-1", Version: 2})
		if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-c"); err != nil {
			t.Fatal(err)
		}
		fcWantCurrent(t, store, "t1", "run-c", "lane-1")
	})
	t.Run("no_record_stays_absent", func(t *testing.T) {
		_, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked)
		if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-a"); err != nil {
			t.Fatal(err)
		}
		if run, owner, ok := fcQueueCurrent(t, store, "t1"); ok {
			t.Fatalf("assign created a current-dispatch record (run %q, owner %q) in a queue the dispatch hook never recorded", run, owner)
		}
	})
}
