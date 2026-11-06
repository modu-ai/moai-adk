package mergeq

// Defect answer (lane-19 37th review): a card withdrawn while the gates ran makes the landing write refuse
// (ErrStaleRow). The lock is held at that moment, so the row must be written back to waiting there; left verifying,
// the holder's way back into the queue was an Enqueue that the ledger refuses for a verifying row. Real ledger.

import (
	"context"
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestLand_RealLedger_AWithdrawnCardsRowGoesBackToWaitingAndTheCardCanBeQueuedAgain(t *testing.T) {
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
		return s.Move(ctx, "t1", epoch, core.Position{Stage: core.StageReviewing}, "lane", "withdrawn during the gates")
	}
	if _, err := m.Land(ctx, "t1", cand); !errors.Is(err, ErrStaleRow) {
		t.Fatalf("Land of a card withdrawn during the gates: %v, want ErrStaleRow", err)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueWaiting {
		t.Fatalf("the row is %s, want waiting (a verifying row cannot be queued again)", got)
	}
	// the holder queues the card again: the same candidate, the same epoch
	if err := s.Enqueue(ctx, "t1", epoch, cand); err != nil {
		t.Fatalf("queueing the withdrawn card again: %v", err)
	}
	cards, _ := s.Cards(ctx)
	if len(cards) != 1 || cards[0].Position.Stage != core.StageMerging {
		t.Errorf("cards %+v, want t1 back in MERGING", cards)
	}
}
