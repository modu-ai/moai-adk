package mergeq

// N10 of the final audit of a36ea8c368 (residual classification): a card held by the operator while its gates run is refused at the
// verifying write, and the error that reaches the caller begins "stale epoch; the card was claimed again". The card was not claimed
// again; it is parked. The expectation: the error names the park, not a claim.

import (
	"context"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

func TestResidualN10_AHoldDuringTheGateIsRefusedAsAParkNotAsAClaim(t *testing.T) {
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
		return s.Move(ctx, "t1", epoch, core.Position{Stage: core.StageMerging, Park: &core.Park{Reason: core.ReasonOperatorHold, Evidence: "operator", Resume: "operator lifts it"}}, "operator", "hold",
			ledger.Evidence{Kind: ledger.KindReview, TreeSHA: "t", SubjectSHA: cand, Result: ledger.ResultPass, Guards: []string{core.GuardOperatorHold}})
	}
	_, lerr := m.Land(ctx, "t1", cand)
	t.Logf("land: err=%v", lerr)
	if lerr == nil {
		t.Fatal("a card held during its gates landed")
	}
	if strings.Contains(lerr.Error(), "claimed again") {
		t.Errorf("the refusal of a parked card says the card was claimed again: %v", lerr)
	}
}
