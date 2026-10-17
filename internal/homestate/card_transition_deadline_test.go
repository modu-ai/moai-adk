package homestate

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// The claim deadline (card t1696): the factory record's write transactions are
// begun with the caller's bounded context, and database/sql rolls such a
// transaction back the moment its context expires. A BeforeCommit whose write
// to the OTHER store survives past the deadline (the queue mirror carries its
// own connection and timeout) used to leave the mirror moved while this
// transaction's Commit returned sql.ErrTxDone — a refusal reported for a
// transition whose mirror had already moved, in the misleading words
// "transaction has already been committed or rolled back" (observed live:
// lane-20 was refused `factory next` while events carried its committed
// lease).
//
// The contract here: the transaction's lifetime is decoupled from the
// caller's deadline (each statement stays bounded by it), so a BeforeCommit
// that outlives the deadline still commits, and the mirror and the transition
// land together or not at all.
func TestTransitionBeforeCommitOutlivingTheDeadlineCommits(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, Card{RunID: frRun, CardID: "bc4", State: CardPicked, Version: 1})
	mirrorWritten := false
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	card, err := db.Transition(ctx, TransitionRequest{
		RunID: frRun, CardID: "bc4", To: CardAssigned, ExpectedVersion: 1, Actor: "assign", Owner: "worker-1", Now: frNow,
		BeforeCommit: func(Card) error {
			// The queue mirror's own connection and timeout outlive this
			// transaction's deadline (backlogOpTimeout, 30 s): a slow mirror
			// write outlives it too, exactly as under a loaded host.
			time.Sleep(200 * time.Millisecond)
			mirrorWritten = true
			return nil
		},
	})
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if !mirrorWritten {
		t.Fatal("BeforeCommit never ran")
	}
	if card.State != CardAssigned {
		t.Fatalf("card = %s, want assigned", card.State)
	}
	landed, lerr := db.LoadCard(context.Background(), frRun, "bc4")
	if lerr != nil || landed.State != CardAssigned {
		t.Fatalf("the mirror moved but the transition did not land: card=%+v err=%v", landed, lerr)
	}
}

// A transaction begun with a dead context still refuses honestly: the
// statement-level bound fires (context deadline exceeded) and the row and the
// log stay untouched — never the post-rollback "transaction has already been
// committed or rolled back" text of a watcher-killed transaction.
func TestTransitionWithAnExpiredContextRefusesWithoutErrTxDone(t *testing.T) {
	db := frOpen(t)
	frPlace(t, db, Card{RunID: frRun, CardID: "bc5", State: CardPicked, Version: 1})
	before := frRowDump(t, db, frRun, "bc5")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	time.Sleep(20 * time.Millisecond) // the deadline is already past
	_, err := db.Transition(ctx, TransitionRequest{
		RunID: frRun, CardID: "bc5", To: CardAssigned, ExpectedVersion: 1, Actor: "assign", Owner: "worker-1", Now: frNow,
	})
	if err == nil {
		t.Fatal("the transition succeeded with an expired context, want a refusal")
	}
	if strings.Contains(err.Error(), "transaction has already been committed or rolled back") {
		t.Fatalf("the refusal leaked the post-rollback tx-done text: %v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want it to wrap context.DeadlineExceeded", err)
	}
	if after := frRowDump(t, db, frRun, "bc5"); after != before {
		t.Fatalf("the refused transition changed the row or the log:\nbefore %s\nafter  %s", before, after)
	}
}
