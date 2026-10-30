package mergeq

// Defect answer: Reconcile and the already-in-trunk path of Land were written against a fake that accepts any
// write. These tests run them on the real ledger: a verifying row whose card still holds a live lease is left
// alone (its lander may be about to write landing), one failing row does not stop the rest, and a row that is
// already landing is not written verifying again.

import (
	"context"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestReconcile_RealLedger_AVerifyingRowWithALiveLeaseIsLeftAlone(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	rows, _ := s.Queue(ctx)
	if err := s.QueueUpdate(ctx, "t1", ledger.CardEpoch(rows[0].CardEpoch, cand), ledger.QueueVerifying, "t0", ""); err != nil {
		t.Fatal(err)
	}
	q := NewLedgerQueue(s)
	m := realMerger(t, dir, s)

	live, err := q.Items(ctx, time.Now())
	if err != nil || len(live) != 1 || !live[0].LeaseLive {
		t.Fatalf("items %+v err %v, want one row with a live lease", live, err)
	}
	if acts, err := m.Reconcile(ctx, live); err != nil || len(acts) != 0 {
		t.Fatalf("a live lander's row: actions %+v err %v, want none", acts, err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueVerifying {
		t.Fatalf("the row is %s, want verifying", got)
	}

	// Control: once the lease has expired nobody is verifying, and the row goes back to waiting.
	dead, err := q.Items(ctx, time.Now().Add(24*time.Hour))
	if err != nil || len(dead) != 1 || dead[0].LeaseLive {
		t.Fatalf("items %+v err %v, want one row without a live lease", dead, err)
	}
	if acts, err := m.Reconcile(ctx, dead); err != nil || len(acts) != 1 {
		t.Fatalf("an abandoned row: actions %+v err %v, want one", acts, err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueWaiting {
		t.Errorf("the row is %s, want waiting", got)
	}
}

func TestReconcile_RealLedger_OneFailingRowDoesNotStopTheOthers(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	leaveLanding(t, s, cand)
	items := append([]Item{{CardID: "ghost", CandidateSHA: "0000000000000000000000000000000000000000", State: StateVerifying}}, itemsOf(t, s)...)
	acts, err := realMerger(t, dir, s).Reconcile(ctx, items)
	if err == nil {
		t.Error("the failing row was not reported")
	}
	if len(acts) != 1 || acts[0].CardID != "t1" {
		t.Errorf("actions %+v, want the second row handled", acts)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueWaiting {
		t.Errorf("the row after the failing one is %s, want waiting", got)
	}
}

func TestLand_RealLedger_ALandingRowWhoseCandidateIsInTheTrunkIsSettledNotRewritten(t *testing.T) {
	dir := repo(t)
	cand := advanceTrunk(t, dir, "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	leaveLanding(t, s, cand)
	res, err := realMerger(t, dir, s).Land(context.Background(), "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("result %+v err %v, want landed", res, err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed", got)
	}
}
