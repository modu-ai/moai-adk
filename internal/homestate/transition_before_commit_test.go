package homestate

import (
	"context"
	"errors"
	"testing"
)

// BeforeCommit lets a caller order a write to ANOTHER store ahead of this
// commit: it runs inside the transition's transaction, once every guard has
// passed and the new row is written, right before the commit (card t1538,
// turn-end gate). The queue's current-dispatch record is that write — it must
// name the run before the binding moves, only for a transition that will land.

func TestTransitionBeforeCommitSeesTheRowTheTransitionWrites(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, Card{RunID: frRun, CardID: "bc1", State: CardPicked, Version: 1})
	var seen []Card
	got, err := db.Transition(context.Background(), TransitionRequest{
		RunID: frRun, CardID: "bc1", To: CardAssigned, ExpectedVersion: 1, Actor: "assign", Owner: "worker-1", Now: frNow,
		BeforeCommit: func(next Card) error { seen = append(seen, next); return nil },
	})
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if len(seen) != 1 {
		t.Fatalf("BeforeCommit ran %d times, want once", len(seen))
	}
	if seen[0].State != CardAssigned || seen[0].OwnerLabel != "worker-1" || seen[0].Version != got.Version {
		t.Fatalf("BeforeCommit saw %+v, want the assigned row at version %d owned by worker-1", seen[0], got.Version)
	}
}

func TestTransitionBeforeCommitErrorRollsTheTransitionBack(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, Card{RunID: frRun, CardID: "bc2", State: CardPicked, Version: 1})
	before := frRowDump(t, db, frRun, "bc2")
	boom := errors.New("record write failed")
	_, err := db.Transition(context.Background(), TransitionRequest{
		RunID: frRun, CardID: "bc2", To: CardAssigned, ExpectedVersion: 1, Actor: "assign", Owner: "worker-1", Now: frNow,
		BeforeCommit: func(Card) error { return boom },
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the hook's error", err)
	}
	if after := frRowDump(t, db, frRun, "bc2"); after != before {
		t.Fatalf("a rolled-back transition changed the row or the log:\nbefore %s\nafter  %s", before, after)
	}
	// T2 writes the dispatch binding in the same transaction: it rolls back too.
	var bindings int
	if err := db.DB.QueryRow(`SELECT count(*) FROM card_dispatch WHERE card_id=?`, "bc2").Scan(&bindings); err != nil {
		t.Fatal(err)
	}
	if bindings != 0 {
		t.Fatalf("a rolled-back T2 left %d dispatch binding row(s)", bindings)
	}
}

func TestTransitionBeforeCommitIsNotCalledOnARefusal(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, Card{RunID: frRun, CardID: "bc3", State: CardPicked, Version: 1})
	called := false
	_, err := db.Transition(context.Background(), TransitionRequest{
		RunID: frRun, CardID: "bc3", To: CardAssigned, ExpectedVersion: 9, Actor: "assign", Owner: "worker-1", Now: frNow,
		BeforeCommit: func(Card) error { called = true; return nil },
	})
	if !errors.Is(err, ErrStaleVersion) {
		t.Fatalf("err = %v, want ErrStaleVersion", err)
	}
	if called {
		t.Fatal("BeforeCommit ran for a refused transition")
	}
}
