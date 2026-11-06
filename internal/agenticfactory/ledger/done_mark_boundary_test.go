package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit db0a69698a N32: two rules of the DONE check were not tied to any test, and the
// audit's mutations of them survived the whole suite. These tests state them at the edge.

// Q1: a green about another commit is older than the offer of the landed commit when its id is no
// greater than the offer's evidence_mark, so a green whose id IS the mark is the last piece of
// evidence the card had when it made the offer, and counts as older (left behind). One more and
// it is newer work. The edge is set by writing the mark itself, which is a count in the ledger.
func TestDone_AGreenWithExactlyTheIdOfTheOffersMarkIsOlderWork(t *testing.T) {
	for _, tc := range []struct {
		name string
		move int64 // added to the id of the green about c2 to get the offer's mark
		want bool  // whether the card may finish on c1
	}{
		{"the mark is the id of the green", 0, true},
		{"the mark is one more than the id", 1, true},
		{"the mark is one less than the id", -1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := openTemp(t)
			c := gated(t, s, "a", "c1")
			e := c.Epoch
			if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
				t.Fatal(err)
			}
			landOne(t, s, "a", e)
			greenOn(t, s, e, "c2", ResultPass)
			var id int64
			if err := s.db.QueryRowContext(bg, `SELECT max(id) FROM evidence WHERE card_id='a' AND subject_sha='c2'`).Scan(&id); err != nil {
				t.Fatal(err)
			}
			if _, err := s.db.ExecContext(bg, `UPDATE transitions SET evidence_mark=? WHERE card_id='a' AND to_stage='MERGING'`, id+tc.move); err != nil {
				t.Fatal(err)
			}
			err := finish(s, e)
			if tc.want && err != nil {
				t.Fatalf("a green with id %d, offer mark %d, is taken for newer work: %v", id, id+tc.move, err)
			}
			if !tc.want && !errors.Is(err, core.ErrGuardMissing) {
				t.Fatalf("a green with id %d, offer mark %d, is older work and the card finished: %v", id, id+tc.move, err)
			}
		})
	}
}

// Q3: the rework mark is the one of the LATEST move back into BUILDING. A card taken back twice
// has to have been measured since the second time; a measurement between the two is old.
func TestDone_ACardTakenBackTwiceNeedsAMeasurementAfterTheSecondMoveBack(t *testing.T) {
	back := func(t *testing.T, s *Store, e int64) {
		t.Helper()
		mustMove(t, s, e, posImplement, "lane")
		mustMove(t, s, e, posReviewing, "lane")
	}
	setup := func(t *testing.T) (*Store, int64) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		e := c.Epoch
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatal(err)
		}
		landOne(t, s, "a", e)
		mustMove(t, s, e, posReviewing, "lane")
		back(t, s, e) // the first move back
		return s, e
	}
	done := core.Position{Stage: core.StageDone}
	t.Run("measured between the two, not since the second", func(t *testing.T) {
		s, e := setup(t)
		evidenceOn(t, s, e, "c1")
		back(t, s, e) // the second move back
		if err := s.Move(bg, "a", e, done, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("a measurement made before the second move back let the card finish: %v", err)
		}
	})
	t.Run("measured since the second", func(t *testing.T) {
		s, e := setup(t)
		evidenceOn(t, s, e, "c1")
		back(t, s, e)
		evidenceOn(t, s, e, "c1")
		if err := s.Move(bg, "a", e, done, "lane", "done"); err != nil {
			t.Fatalf("measured again since the second move back, and the card is refused: %v", err)
		}
	})
	t.Run("measured since the first, taken back once", func(t *testing.T) {
		s, e := setup(t)
		evidenceOn(t, s, e, "c1")
		if err := s.Move(bg, "a", e, done, "lane", "done"); err != nil {
			t.Fatalf("measured again since the only move back, and the card is refused: %v", err)
		}
	})
}
