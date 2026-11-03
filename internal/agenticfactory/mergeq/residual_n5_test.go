package mergeq

// N5 of the final audit of a36ea8c368 (residual classification): a verifying row whose candidate is already in the trunk, on a
// withdrawn card, is skipped by reconcile without a word; releasing the lease does not change that. The command succeeds and
// nobody is told. The expectation: after the lease is released a reconcile settles the row (it is in the trunk) or resets it.

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestResidualN5_ReconcileSettlesAVerifyingRowWhoseCandidateIsInTheTrunkOnAWithdrawnCard(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	rows, err := s.Queue(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatalf("queue: %+v %v", rows, err)
	}
	epoch := rows[0].CardEpoch
	if err := NewLedgerQueue(s).SetState(ctx, "t1", cand, 0, StateVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(ctx, "t1", epoch, core.Position{Stage: core.StageReviewing}, "lane", "withdrawn"); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "update-ref", "refs/heads/trunk", cand)
	if err := s.Release(ctx, "t1", epoch); err != nil {
		t.Fatal(err)
	}
	items, err := NewLedgerQueue(s).Items(ctx, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	acts, rerr := realMerger(t, dir, s).Reconcile(ctx, items)
	state := rowState(t, s, "t1")
	t.Logf("reconcile after release: actions=%+v err=%v row=%s", acts, rerr, state)
	if state == ledger.QueueVerifying {
		t.Errorf("reconcile left the verifying row of a released, withdrawn card whose candidate is in the trunk: actions=%+v err=%v", acts, rerr)
	}
}
