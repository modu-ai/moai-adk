package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit db0a69698a N25: a commit that was offered to the queue and failed or was withdrawn, after
// which the card went back into the build stage, needs a passing run recorded since the move back
// before the card can finish on its landing. Enqueue used to take the commit again anyway, so the
// refusal came after the work was in the trunk, and when the holder was gone there was no one
// left to measure it: the card stayed in MERGING for good. The refusal belongs at the entrance,
// where the holder is still there to act on it.

// offeredThenTakenBack leaves the card in REVIEWING after c1 was offered, ended ("failed" or
// "withdrawn") and the card went through BUILDING and back.
func offeredThenTakenBack(t *testing.T, ending string) (*Store, int64) {
	t.Helper()
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	e := c.Epoch
	if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
		t.Fatal(err)
	}
	if ending == "failed" {
		if err := s.QueueUpdate(bg, "a", CardEpoch(e, "c1"), QueueFailed, "", "flaky gate"); err != nil {
			t.Fatal(err)
		}
	}
	mustMove(t, s, e, posReviewing, "lane")
	mustMove(t, s, e, posImplement, "lane")
	mustMove(t, s, e, posReviewing, "lane")
	return s, e
}

func TestEnqueue_ACommitOfferedBeforeTheMoveBackNeedsARunRecordedSince(t *testing.T) {
	for _, ending := range []string{"failed", "withdrawn"} {
		t.Run(ending+", refused, measured again, offered, landed, finished without a holder", func(t *testing.T) {
			s, e := offeredThenTakenBack(t, ending)
			before, _ := queueRow(t, s, "a")
			err := s.Enqueue(bg, "a", e, "c1")
			if !errors.Is(err, ErrEvidenceStale) {
				t.Fatalf("c1, offered before the move back with nothing recorded since, was taken again: %v", err)
			}
			if got := cardOf(t, s, "a").Position.Stage; got != core.StageReviewing {
				t.Errorf("the refused offer moved the card to %s", got)
			}
			if after, _ := queueRow(t, s, "a"); after != before {
				t.Errorf("the refused offer changed the queue row: %+v -> %+v", before, after)
			}
			// The way out: measure it again, offer it, land it, and finish without anyone holding the card.
			evidenceOn(t, s, e, "c1")
			if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
				t.Fatalf("c1 measured again after the move back is refused: %v", err)
			}
			landOne(t, s, "a", e)
			if err := s.Release(bg, "a", e); err != nil {
				t.Fatal(err)
			}
			if err := finish(s, e); err != nil {
				t.Fatalf("no holder, c1 measured again and landed: %v", err)
			}
			if got := cardOf(t, s, "a").Position.Stage; got != core.StageDone {
				t.Errorf("stage %s", got)
			}
		})
	}
}

func TestEnqueue_WhatCountsAsARunRecordedSinceTheMoveBack(t *testing.T) {
	passGate := func(t *testing.T, s *Store, e int64) { gateOn(t, s, e, "c1", ResultPass) }
	passGreen := func(t *testing.T, s *Store, e int64) { greenOn(t, s, e, "c1", ResultPass) }
	failGreen := func(t *testing.T, s *Store, e int64) { greenOn(t, s, e, "c1", ResultFail) }
	cases := []struct {
		name   string
		record []func(t *testing.T, s *Store, e int64)
		want   bool
	}{
		{"a passing gate alone", []func(*testing.T, *Store, int64){passGate}, true},
		{"a passing green alone (the gate before the move back still passes)", []func(*testing.T, *Store, int64){passGreen}, true},
		{"nothing", nil, false},
		{"a failing green and then a passing gate", []func(*testing.T, *Store, int64){failGreen, passGate}, false},
		{"a failing green, a passing gate, a passing green", []func(*testing.T, *Store, int64){failGreen, passGate, passGreen}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, e := offeredThenTakenBack(t, "failed")
			for _, r := range tc.record {
				r(t, s, e)
			}
			err := s.Enqueue(bg, "a", e, "c1")
			switch {
			case tc.want && err != nil:
				t.Fatalf("refused: %v", err)
			case !tc.want && !errors.Is(err, ErrEvidenceStale):
				t.Fatalf("taken: %v", err)
			}
		})
	}
}

// What stays accepted: a candidate the card never offered before (REQ-AFL-008, decision D5), a card not
// taken back (a flaky retry of the same commit), and a commit the card was taken back for and has
// since measured.
func TestEnqueue_NoReworkMarkNoNewDemand(t *testing.T) {
	t.Run("a gate-only candidate never offered before, after a withdrawn c1", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		e := c.Epoch
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatal(err)
		}
		mustMove(t, s, e, posReviewing, "lane")
		mustMove(t, s, e, posImplement, "lane")
		gateOn(t, s, e, "c2", ResultPass)
		mustMove(t, s, e, posReviewing, "lane")
		if err := s.Enqueue(bg, "a", e, "c2"); err != nil {
			t.Fatalf("a gate-only candidate is refused (D5): %v", err)
		}
	})
	t.Run("the same commit again after a failure, the card never taken back", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		e := c.Epoch
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatal(err)
		}
		if err := s.QueueUpdate(bg, "a", CardEpoch(e, "c1"), QueueFailed, "", "flaky gate"); err != nil {
			t.Fatal(err)
		}
		mustMove(t, s, e, posReviewing, "lane")
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatalf("a retry of the same commit with no rework is refused: %v", err)
		}
	})
	t.Run("withdrawn to REVIEWING and offered again, never back in BUILDING", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		e := c.Epoch
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatal(err)
		}
		mustMove(t, s, e, posReviewing, "lane")
		if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
			t.Fatalf("c1 withdrawn and offered again is refused: %v", err)
		}
	})
}
