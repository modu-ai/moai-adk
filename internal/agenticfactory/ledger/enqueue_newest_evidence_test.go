package ledger

import (
	"testing"
)

// issue-1948, decision D5: a gate-only candidate goes through Enqueue, whatever else the card measured.

// What Enqueue accepts (decision D5, N14): a candidate with a gate, whether or not the card has a green for it.
func TestEnqueue_AGateOnlyCandidateStaysAcceptedWhenItIsTheNewestEvidence(t *testing.T) {
	t.Run("c1 has a green and a gate, c2 only a gate", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		gateOn(t, s, c.Epoch, "c2", ResultPass)
		if err := s.Enqueue(bg, "a", c.Epoch, "c2"); err != nil {
			t.Fatalf("a gate-only candidate is refused (REQ-AFL-008): %v", err)
		}
		landOne(t, s, "a", c.Epoch)
		if err := finish(s, c.Epoch); err != nil {
			t.Fatalf("c2 landed with its passing gate and the card cannot finish: %v", err)
		}
	})
	t.Run("c2 measured twice, the last run a gate", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		greenOn(t, s, c.Epoch, "c2", ResultPass)
		gateOn(t, s, c.Epoch, "c2", ResultPass)
		if err := s.Enqueue(bg, "a", c.Epoch, "c2"); err != nil {
			t.Fatalf("the newest evidence is about c2 and c2 is refused: %v", err)
		}
	})
	t.Run("the only evidence is about the candidate", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatalf("%v", err)
		}
	})
}
