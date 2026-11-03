package mergeq

// Defect answer: lane-18 f052a5af31 makes the ledger refuse a row's advance to verifying, landing or landed while
// its card is outside MERGING. The gates run outside the lock, so a holder can withdraw the card between the
// verification and the landing. This test runs that race on the real ledger and a real git repository: Land must
// stop with ErrStaleRow, move nothing, and release the lock.

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestLand_RealLedger_ACardWithdrawnDuringTheGatesIsNotLanded(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	rows, err := s.Queue(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatalf("queue: %+v %v", rows, err)
	}
	epoch := rows[0].CardEpoch
	before := git(t, dir, "rev-parse", "refs/heads/trunk")

	m := realMerger(t, dir, s)
	m.Verify = func(ctx context.Context, _ string) error {
		// the gates pass, and while they ran the holder pulled the card back out of the merge queue
		return s.Move(ctx, "t1", epoch, core.Position{Stage: core.StageReviewing}, "lane", "withdrawn during the gates")
	}
	_, err = m.Land(ctx, "t1", cand)
	if !errors.Is(err, ErrStaleRow) {
		t.Fatalf("Land of a card withdrawn during the gates: %v, want ErrStaleRow", err)
	}
	if got := git(t, dir, "rev-parse", "refs/heads/trunk"); got != before {
		t.Errorf("the withdrawn work reached the trunk: %s -> %s", before, got)
	}
	if got := rowState(t, s, "t1"); got == ledger.QueueLanded || got == ledger.QueueLanding {
		t.Errorf("the row is %s, want it neither landing nor landed", got)
	}
	// the lock was released: another lander can take it
	lock, err := s.AcquireLock(ctx, ledger.LockMerge, ledger.LockHolder{SessionID: "other", PID: 1, PIDStart: "x"}, DefaultLockTTL)
	if err != nil {
		t.Fatalf("the landing lock was left held: %v", err)
	}
	_ = s.ReleaseLock(ctx, ledger.LockMerge, lock)
}
