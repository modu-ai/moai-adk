package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit db0a69698a N27: decision D5 lets a gate-only candidate through Enqueue, and the
// newest-work check at DONE therefore lets a green about another commit that is OLDER than the
// offer of the landed commit go. That left a way to finish a card on its older commit: measure c1,
// then c2 (green, passing or failing), and offer only c1. c1 is in the trunk with its own
// evidence, but the card's newest work is not, and its green may have failed. The ledger holds
// no git history, so it cannot tell whether c2 contains c1; what it can ask is whether the commit
// being offered is the one the card measured last. Enqueue refuses a candidate whose commit is not
// the subject of the card's most recent evidence. A gate-only c2 is the subject of the newest
// evidence when it is offered, so D5 stands; measuring c1 again makes c1 the newest.

func TestEnqueue_TheCandidateMustBeWhatTheCardMeasuredLast(t *testing.T) {
	cases := []struct {
		name    string
		measure func(t *testing.T, s *Store, e int64)
	}{
		{"c2 has a passing green", func(t *testing.T, s *Store, e int64) { greenOn(t, s, e, "c2", ResultPass) }},
		{"c2 has a failing green", func(t *testing.T, s *Store, e int64) { greenOn(t, s, e, "c2", ResultFail) }},
		{"c2 has a green and a gate", func(t *testing.T, s *Store, e int64) { evidenceOn(t, s, e, "c2") }},
		{"c2 has a gate only", func(t *testing.T, s *Store, e int64) { gateOn(t, s, e, "c2", ResultPass) }},
	}
	for _, tc := range cases {
		t.Run(tc.name+", the older c1 is refused until it is measured again", func(t *testing.T) {
			s := openTemp(t)
			c := gated(t, s, "a", "c1")
			e := c.Epoch
			tc.measure(t, s, e)
			err := s.Enqueue(bg, "a", e, "c1")
			if !errors.Is(err, ErrEvidenceStale) {
				t.Fatalf("c1 was offered while the card's newest evidence is about c2: %v", err)
			}
			if row, ok := queueRow(t, s, "a"); ok {
				t.Errorf("the refused offer left a queue row: %+v", row)
			}
			if got := cardOf(t, s, "a").Position.Stage; got == core.StageMerging {
				t.Errorf("the refused offer moved the card to %s", got)
			}
			// The way out: measure c1 again, and it is the card's newest work.
			evidenceOn(t, s, e, "c1")
			if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
				t.Fatalf("c1 measured again is refused: %v", err)
			}
			landOne(t, s, "a", e)
			if err := finish(s, e); err != nil {
				t.Fatalf("c1 measured again and landed, and the card cannot finish: %v", err)
			}
		})
	}
}

// What stays accepted (decision D5, N14): the newest evidence is about the candidate itself, whether
// or not the card has a green for it.
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
		greenOn(t, s, c.Epoch, "c2", ResultFail)
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
