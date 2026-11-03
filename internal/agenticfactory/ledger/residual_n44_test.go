package ledger

// N44 rest (audits 46dbcce0d4 and a36ea8c368, residual classification): the mutant N40-bar-green-only (the "measured again after
// other work" rule asks for a passing green after the bar and no longer a passing gate) survived the whole suite. Rule 2 of
// done_results_after_landing_test.go says the landed commit must be measured again with BOTH a passing green and a passing
// gate once the card recorded evidence about another commit; the gate-only half is tested there, this is the green-only half.

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func TestResidualN44_AGreenOfTheLandedCommitAloneDoesNotSayTheOtherWorkWasGivenUp(t *testing.T) {
	s, e := landedThenWithdrawn(t)
	gateOn(t, s, e, "c2", ResultPass)
	greenOn(t, s, e, "c1", ResultPass)
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("DONE accepted with c2 gated after the withdrawal and only a green of c1 recorded since: %v", err)
	}
	gateOn(t, s, e, "c1", ResultPass)
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); err != nil {
		t.Fatalf("c1 measured again with a green and a gate and the card is refused: %v", err)
	}
}
