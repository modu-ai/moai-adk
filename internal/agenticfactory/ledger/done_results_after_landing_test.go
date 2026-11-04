package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit db0a69698a N28, reopened as a blocker by the operator (repair note 1-B). An
// earlier version of this file, accepted_risk_n28_test.go, recorded three shapes in which a card
// finished although its landed candidate no longer stood: a failure recorded after the landing
// with no move back, a gate-only new commit after a withdrawal to REVIEWING, and a gate-only fix
// after a move back into BUILDING followed by a gate on c1. It pinned the outcome (DONE) on the
// premise that the leader accepted the risk. The premise is withdrawn, so the assertions are
// turned round: the ledger refuses the three shapes, and each has the way out that the rules ask for.
//
// The rules:
//  1. The newest green run and the newest gate run about the landed candidate must not be
//     failures, whether or not the card was ever taken back.
//  2. When the card left MERGING after the candidate was offered and has since recorded evidence
//     about another commit, the candidate counts only if it was measured again after that
//     evidence, with both a passing green and a passing gate. A gate of the landed commit alone
//     does not tell whether the work the card did in between was given up.
// The same rules apply at Enqueue to a commit that was offered before, so the entrance and the
// exit agree and no commit is taken that cannot finish.

// A failure recorded after the landing, the card still in MERGING and never taken back.
func TestDone_AFailureRecordedAfterTheLandingRefusesTheCardWhetherOrNotItWasTakenBack(t *testing.T) {
	for _, tc := range []struct {
		name string
		do   func(t *testing.T, s *Store, e int64)
	}{
		{"the green and the gate failed", func(t *testing.T, s *Store, e int64) {
			greenOn(t, s, e, "c1", ResultFail)
			gateOn(t, s, e, "c1", ResultFail)
		}},
		{"the green failed", func(t *testing.T, s *Store, e int64) { greenOn(t, s, e, "c1", ResultFail) }},
		{"the gate failed", func(t *testing.T, s *Store, e int64) { gateOn(t, s, e, "c1", ResultFail) }},
	} {
		t.Run(tc.name+", still in MERGING", func(t *testing.T) {
			s := openTemp(t)
			c := gated(t, s, "a", "c1")
			if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
				t.Fatal(err)
			}
			landOne(t, s, "a", c.Epoch)
			tc.do(t, s, c.Epoch)
			if err := finish(s, c.Epoch); !errors.Is(err, core.ErrGuardMissing) {
				t.Fatalf("the card finished with a failing newest run of its landed commit: %v", err)
			}
			if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
				t.Error("the card is DONE")
			}
			// Measured again and passing, the failure is repaired.
			evidenceOn(t, s, c.Epoch, "c1")
			if err := finish(s, c.Epoch); err != nil {
				t.Fatalf("c1 passes again and the card is refused: %v", err)
			}
		})
		t.Run(tc.name+", withdrawn to REVIEWING", func(t *testing.T) {
			s, e := landedThenWithdrawn(t)
			tc.do(t, s, e)
			if err := s.Move(bg, "a", e, posDone, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
				t.Fatalf("the card finished with a failing newest run of its landed commit: %v", err)
			}
		})
	}
}

// Withdrawn to REVIEWING (never through BUILDING) and a gate-only new commit that is not offered.
func TestDone_WithdrawnToReviewingThenAGateOnlyNewCommitIsRefused(t *testing.T) {
	t.Run("c2 gated and never offered", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		gateOn(t, s, e, "c2", ResultPass)
		if err := s.Move(bg, "a", e, posDone, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("DONE accepted with c2 gated after the withdrawal and c1 not measured again: %v", err)
		}
		// A gate of c1 alone does not say c2 was given up.
		gateOn(t, s, e, "c1", ResultPass)
		if err := s.Move(bg, "a", e, posDone, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("DONE accepted on a gate of c1 alone after the evidence about c2: %v", err)
		}
		// Measured again in full, c1 is the work the card stands on.
		evidenceOn(t, s, e, "c1")
		if err := s.Move(bg, "a", e, posDone, "lane", "done"); err != nil {
			t.Fatalf("c1 measured again (green and gate) after c2, and the card is refused: %v", err)
		}
	})
	t.Run("c2 has no route through the queue from REVIEWING, and gets one through BUILDING", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		gateOn(t, s, e, "c2", ResultPass)
		if err := s.Enqueue(bg, "a", e, "c2"); !errors.Is(err, ErrInvalid) {
			t.Fatalf("c2 offered from REVIEWING without a move back into BUILDING: %v", err)
		}
		mustMove(t, s, e, posImplement, "lane")
		mustMove(t, s, e, posReviewing, "lane")
		if err := s.Enqueue(bg, "a", e, "c2"); err != nil {
			t.Fatalf("a gate-only c2 offered after the card was taken back is refused (D5): %v", err)
		}
		landOne(t, s, "a", e)
		if err := finish(s, e); err != nil {
			t.Fatalf("c2 landed and the card cannot finish: %v", err)
		}
	})
	t.Run("a gate of another commit while the card stays in MERGING does not block", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		landOne(t, s, "a", c.Epoch)
		gateOn(t, s, c.Epoch, "c2", ResultPass)
		if err := finish(s, c.Epoch); err != nil {
			t.Fatalf("a second candidate was gated and the card never left MERGING, and it is refused: %v", err)
		}
	})
}

// Taken back into BUILDING, a gate-only fix c2 that is never offered, then c1 measured again.
func TestDone_ReworkProducesAGateOnlyFixThatIsNeverOffered(t *testing.T) {
	s, e := landedThenWithdrawn(t)
	mustMove(t, s, e, posImplement, "lane")
	gateOn(t, s, e, "c2", ResultPass)
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("DONE accepted with only c2 recorded since the move back: %v", err)
	}
	gateOn(t, s, e, "c1", ResultPass)
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("DONE accepted on a gate of c1 alone after the fix c2 was gated: %v", err)
	}
	evidenceOn(t, s, e, "c1")
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); err != nil {
		t.Fatalf("c1 measured again in full after the fix, and the card is refused: %v", err)
	}
}
