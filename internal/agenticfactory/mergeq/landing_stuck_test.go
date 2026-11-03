package mergeq

// Defect answer: once the row is landing under the lock, every way out of Land must leave it somewhere the next
// Land can start from. A swap that fails for a reason other than "the trunk moved" (a stale .lock file, a full
// disk) used to return with the row still landing, and the next Land's verifying write was refused by the real
// ledger for good. These tests run on a real git repository and the real ledger: a failed swap leaves the row
// waiting; a swap that took effect but reported an error settles as landed; a landing row left by a dead lander
// whose candidate is not in the trunk is reset before the new verification.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestLand_RealLedger_ASwapThatFailsForAnotherReasonLeavesTheRowRetryable(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	m := realMerger(t, dir, s)
	before := git(t, dir, "rev-parse", "refs/heads/trunk")

	lockFile := filepath.Join(dir, ".git", "refs", "heads", "trunk.lock") // git refuses to update a ref whose .lock exists
	if err := os.WriteFile(lockFile, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Land(ctx, "t1", cand); err == nil {
		t.Fatal("a swap blocked by a lock file reported no error")
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueWaiting {
		t.Fatalf("the row is %s after the failed swap, want waiting (not stuck in landing)", got)
	}
	if git(t, dir, "rev-parse", "refs/heads/trunk") != before {
		t.Fatal("the trunk moved although the swap failed")
	}

	if err := os.Remove(lockFile); err != nil {
		t.Fatal(err)
	}
	res, err := m.Land(ctx, "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("the next Land: %+v %v, want landed without an error", res, err)
	}
}

// swapThenFail takes the swap and then reports an error, as a git that was killed after it wrote the ref would.
type swapThenFail struct{ Git }

func (g swapThenFail) UpdateRef(ctx context.Context, ref, newSHA, oldSHA string) error {
	if err := g.Git.UpdateRef(ctx, ref, newSHA, oldSHA); err != nil {
		return err
	}
	return errors.New("git was killed after it wrote the ref")
}

func TestLand_RealLedger_ASwapThatTookEffectButReportedAnErrorSettlesAsLanded(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	m := realMerger(t, dir, s)
	m.Git = swapThenFail{m.Git}

	res, err := m.Land(ctx, "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("result %+v err %v, want landed: the trunk did move", res, err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed", got)
	}
	if git(t, dir, "show", "refs/heads/trunk:a.txt") != "a" {
		t.Error("the card's change is not in the trunk")
	}
}

func TestLand_RealLedger_ALandingRowWhoseCandidateIsNotInTheTrunkIsResetBeforeVerifying(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	leaveLanding(t, s, cand) // a lander took the row to landing and died before the swap

	res, err := realMerger(t, dir, s).Land(ctx, "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("result %+v err %v, want landed: the dead lander's row is reset and verified again", res, err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed", got)
	}
}
