package ledger

// N63 (final audit of a36ea8c368): `af land` refuses a row whose newest green, gate or review run about the candidate
// failed (N54), and the refusal says to measure again. A holder that left (released, or reaped after its process went)
// cannot: every write needs the lease, and a card in MERGING with a waiting row cannot be claimed. Before N54 the landing
// went through and the settled row let the card out; with the refusal nothing did. The way out is the one the ledger
// already gives a settled row: a MERGING card with no lease, at its own epoch, may be moved on the authority of the queue
// row alone, and a waiting row whose candidate cannot land is the same kind of fact as a failed one.

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// leftWithAFailingRow is a card whose c1 was measured and enqueued, whose newest run of the kind failed afterwards, and
// whose holder let go.
func leftWithAFailingRow(t *testing.T, s *Store, id, kind string) Claimed {
	t.Helper()
	c := gated(t, s, id, "c1")
	if err := s.Enqueue(bg, id, c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddEvidence(bg, id, c.Epoch, Evidence{Kind: kind, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail}); err != nil {
		t.Fatal(err)
	}
	if err := s.Release(bg, id, c.Epoch); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMove_AMergingCardWhoseWaitingRowCannotLandMovesWithoutALease(t *testing.T) {
	for _, kind := range []string{KindGreen, KindGate, KindReview} {
		t.Run(kind, func(t *testing.T) {
			s := openTemp(t)
			c := leftWithAFailingRow(t, s, "a", kind)
			if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReviewing}, "operator", "take the card back"); err != nil {
				t.Fatalf("the move of a card whose waiting row cannot land, with no holder: %v", err)
			}
			if got := cardOf(t, s, "a").Position.Stage; got != core.StageReviewing {
				t.Errorf("stage %s, want REVIEWING", got)
			}
			// and the next holder claims it and carries it on
			next, err := s.Claim(bg, req("s-next", 10))
			if err != nil || next.CardID != "a" || next.Epoch != c.Epoch+1 {
				t.Fatalf("claimed again: %+v %v", next, err)
			}
		})
	}
}

// The condition is narrow, and each side of it has its own refusal: a row that can land needs its holder, as before; a
// failure about another commit does not stop this one landing; a row that is not waiting is the lander's to settle; and
// the epoch is still the card's own.
func TestMove_TheWayOutIsOnlyForAWaitingRowWhoseCandidateCannotLand(t *testing.T) {
	t.Run("a row that can land", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		if err := s.Release(bg, "a", c.Epoch); err != nil {
			t.Fatal(err)
		}
		err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReviewing}, "operator", "x")
		if !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("a landable row let the card out without its lease: %v", err)
		}
	})
	t.Run("a failure about another commit", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		if err := s.AddEvidence(bg, "a", c.Epoch, Evidence{Kind: KindGate, TreeSHA: "tree-c2", SubjectSHA: "c2", Result: ResultFail}); err != nil {
			t.Fatal(err)
		}
		if err := s.Release(bg, "a", c.Epoch); err != nil {
			t.Fatal(err)
		}
		err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReviewing}, "operator", "x")
		if !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("a failure about c2 opened the way out for c1: %v", err)
		}
	})
	t.Run("a row that is verifying", func(t *testing.T) {
		s := openTemp(t)
		c := leftWithAFailingRow(t, s, "a", KindGate)
		if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
			t.Fatal(err)
		}
		err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReviewing}, "operator", "x")
		if !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("a verifying row (a lander may be at work) let the card out: %v", err)
		}
	})
	t.Run("another epoch", func(t *testing.T) {
		s := openTemp(t)
		c := leftWithAFailingRow(t, s, "a", KindGate)
		err := s.Move(bg, "a", c.Epoch+1, core.Position{Stage: core.StageReviewing}, "operator", "x")
		if !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("an epoch that is not the card's own: %v", err)
		}
		if got := cardOf(t, s, "a").Position.Stage; got != core.StageMerging {
			t.Errorf("stage %s after the refusals, want MERGING", got)
		}
	})
}
