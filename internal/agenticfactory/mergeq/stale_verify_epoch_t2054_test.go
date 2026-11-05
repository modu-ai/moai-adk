package mergeq

// Card t2054, RC40 review finding P2 (ledgerqueue.go SetState): a lander verifying a candidate presents the row's
// card epoch read at WRITE time, not the epoch it started verifying at. If the card is taken back and a new holder
// enqueues the same commit while the gates run, the old lander's failure then carries the new holder's epoch and
// candidate, passes the ledger fence, and marks the new holder's row failed.

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestLand_RealLedger_AGateFailureOfAFencedLanderDoesNotFailTheNewHoldersRow(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)

	items, err := s.Queue(ctx)
	if err != nil || len(items) != 1 {
		t.Fatalf("queue: %+v %v", items, err)
	}
	oldEpoch := items[0].CardEpoch

	var newEpoch int64
	// Mid-gate: the card is taken back, released, claimed again by another holder, which queues the SAME commit.
	retake := func(context.Context, string) error {
		if err := s.Move(ctx, "t1", oldEpoch, core.Position{Stage: core.StageReviewing}, "lane", "back out of the queue"); err != nil {
			t.Fatalf("leave MERGING: %v", err)
		}
		if err := s.Release(ctx, "t1", oldEpoch); err != nil {
			t.Fatalf("release: %v", err)
		}
		c2, err := s.Claim(ctx, ledger.ClaimRequest{SessionID: "s-new", Worktree: "/wt/new", PID: os.Getpid(), PIDStart: "start-new"})
		if err != nil || c2.CardID != "t1" {
			t.Fatalf("claim again: %+v %v", c2, err)
		}
		newEpoch = c2.Epoch
		for _, e := range []ledger.Evidence{
			{Kind: ledger.KindGreen, TreeSHA: "tree-" + cand, SubjectSHA: cand, Result: ledger.ResultPass},
			{Kind: ledger.KindGate, TreeSHA: "tree-" + cand, SubjectSHA: cand, Command: "go test ./...", Result: ledger.ResultPass},
		} {
			if err := s.AddEvidence(ctx, "t1", newEpoch, e); err != nil {
				t.Fatalf("evidence: %v", err)
			}
		}
		if err := s.Enqueue(ctx, "t1", newEpoch, cand); err != nil {
			t.Fatalf("new holder enqueue: %v", err)
		}
		return errors.New("the old lander's gate failed")
	}

	g := GitCLI{Repo: dir, ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}
	m := &Merger{
		Git: g, Queue: NewLedgerQueue(s), Trunk: "refs/heads/trunk", Verify: retake,
		Holder:  Holder{Session: "s1", PID: os.Getpid(), PIDStart: "start-s1"},
		LockTTL: DefaultLockTTL, MaxRetries: DefaultMaxRetries,
		Message: func(card string) string { return "merge(" + card + "): land" },
	}

	res, landErr := m.Land(ctx, "t1", cand)

	rows, err := s.Queue(ctx)
	if err != nil || len(rows) != 1 {
		t.Fatalf("queue after: %+v %v", rows, err)
	}
	if newEpoch == 0 || rows[0].CardEpoch != newEpoch {
		t.Fatalf("setup: the row is at epoch %d, want the new holder's %d", rows[0].CardEpoch, newEpoch)
	}
	if rows[0].State == ledger.QueueFailed {
		t.Errorf("the old lander's failure marked the new holder's row failed (result %+v, err %v)", res, landErr)
	}
	if !errors.Is(landErr, ErrStaleRow) {
		t.Errorf("Land returned %v, want an error that is ErrStaleRow", landErr)
	}
}
