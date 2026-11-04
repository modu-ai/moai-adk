package mergeq

// N4 of the final audit of a36ea8c368 (residual classification): the card was withdrawn while the gates ran and another
// lander moved the trunk. The rebuild's verifying write is refused and Land returns without the row going back to waiting,
// so the row stays verifying while the lease is alive: enqueue refuses ("already in the queue") and reconcile leaves it.
// The expectation: after Land returns, a row that did not land is not left in verifying.

import (
	"context"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestResidualN4_AWithdrawnCardWhoseTrunkMovedDuringTheGateIsNotLeftVerifying(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	rows, err := s.Queue(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatalf("queue: %+v %v", rows, err)
	}
	epoch := rows[0].CardEpoch
	m := realMerger(t, dir, s)
	m.Verify = func(ctx context.Context, _ string) error {
		other := candidate(t, dir, "other", "o.txt", "o\n")
		git(t, dir, "update-ref", "refs/heads/trunk", other) // another lander's landing moves the trunk under the gate
		return s.Move(ctx, "t1", epoch, core.Position{Stage: core.StageReviewing}, "lane", "withdrawn")
	}
	res, lerr := m.Land(ctx, "t1", cand)
	state := rowState(t, s, "t1")
	t.Logf("land: outcome=%q err=%v row=%s", res.Outcome, lerr, state)
	eqErr := s.Enqueue(ctx, "t1", epoch, cand)
	t.Logf("enqueue at the same epoch: %v row=%s", eqErr, rowState(t, s, "t1"))
	if state == ledger.QueueVerifying {
		t.Errorf("the row of a withdrawn card is left %s after Land returned (enqueue: %v): with a live lease nothing takes it back", state, eqErr)
	}
}
