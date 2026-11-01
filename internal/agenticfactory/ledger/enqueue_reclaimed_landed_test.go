package ledger

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audits a1ec370195 / 22a0afd247 N3, reopened as a blocker by the operator (repair note
// 1-B): a landing is a fact about the trunk, not about the generation of the card that caused it. The
// holder of a card whose candidate c1 landed lets go (it died, or was reaped) before the card
// finished; the card is claimed again at the next epoch. The row of c1 still says landed at the
// OLD epoch, so the facts of the new epoch never see it, and Enqueue refused any landed row of
// another epoch outright: the work was in the trunk, the new holder could neither finish the
// card on it nor offer anything else, and only CANCELLED was left. The new holder may take the
// landing over, or replace it with a new commit, once it has measured the commit in its own
// epoch. Without that measurement nothing changes.

// landedThenReclaimed builds the case: c1 landed at epoch 1, the holder took the card out of MERGING
// (and through BUILDING when rework is set), let go, and a new session claimed the card. before runs
// at epoch 1, after the landing and before the holder lets go. It returns the new epoch.
func landedThenReclaimed(t *testing.T, rework bool, before func(t *testing.T, s *Store, e1 int64)) (*Store, int64) {
	t.Helper()
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	e1 := c.Epoch
	if err := s.Enqueue(bg, "a", e1, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", e1)
	backOut(t, s, e1)
	if rework {
		mustMove(t, s, e1, posImplement, "lane")
	}
	if before != nil {
		before(t, s, e1)
	}
	if err := s.Release(bg, "a", e1); err != nil {
		t.Fatal(err)
	}
	c2, err := s.Claim(bg, req("next", 77))
	if err != nil || c2.Epoch != e1+1 {
		t.Fatalf("claimed again: %+v %v", c2, err)
	}
	return s, c2.Epoch
}

// handOff takes a reclaimed card that is still in BUILDING to REVIEWING, where it can be offered.
func handOff(t *testing.T, s *Store, e2 int64, rework bool) {
	t.Helper()
	if rework {
		mustMove(t, s, e2, posReviewing, "lane")
	}
}

func TestEnqueue_AReclaimedCardTakesOverOrReplacesTheLandedRowOnceItMeasuredInItsOwnEpoch(t *testing.T) {
	for _, rework := range []bool{false, true} {
		name := "withdrawn to REVIEWING"
		if rework {
			name = "taken back through BUILDING"
		}
		t.Run(name+", c1 measured again takes the landing over", func(t *testing.T) {
			s, e2 := landedThenReclaimed(t, rework, nil)
			evidenceOn(t, s, e2, "c1")
			handOff(t, s, e2, rework)
			if err := s.Enqueue(bg, "a", e2, "c1", OnTrunk(true)); err != nil {
				t.Fatalf("c1 landed, was measured again by the new holder, and the offer is refused: %v", err)
			}
			if it, ok := queueRow(t, s, "a"); !ok || it.State != QueueLanded || it.CandidateSHA != "c1" || it.CardEpoch != e2 {
				t.Fatalf("the row after the take-over: %+v", it)
			}
			if got := cardOf(t, s, "a").Position.Stage; got != core.StageMerging {
				t.Errorf("stage %s", got)
			}
			if err := finish(s, e2); err != nil {
				t.Fatalf("the card cannot finish on the landing it took over: %v", err)
			}
			if got := cardOf(t, s, "a").Position.Stage; got != core.StageDone {
				t.Errorf("stage %s", got)
			}
		})
		t.Run(name+", a new commit measured in the new epoch replaces the row", func(t *testing.T) {
			s, e2 := landedThenReclaimed(t, rework, nil)
			evidenceOn(t, s, e2, "c2")
			handOff(t, s, e2, rework)
			if err := s.Enqueue(bg, "a", e2, "c2"); err != nil {
				t.Fatalf("a new commit measured by the new holder is refused: %v", err)
			}
			if it, ok := queueRow(t, s, "a"); !ok || it.State != QueueWaiting || it.CandidateSHA != "c2" || it.CardEpoch != e2 {
				t.Fatalf("the row after the replacement: %+v", it)
			}
			notes, err := s.NotesSince(bg, 0)
			if err != nil {
				t.Fatal(err)
			}
			kept := false
			for _, n := range notes {
				kept = kept || (n.CardID == "a" && strings.Contains(n.Text, "c1") && strings.Contains(n.Text, "landed"))
			}
			if !kept {
				t.Errorf("the landing of c1 is not kept as a note: %+v", notes)
			}
			landOne(t, s, "a", e2)
			if err := finish(s, e2); err != nil {
				t.Fatalf("c2 landed and the card cannot finish: %v", err)
			}
		})
		// What stays refused, and changes nothing.
		for _, tc := range []struct {
			name   string
			sha    string
			before func(t *testing.T, s *Store, e1 int64)
			record func(t *testing.T, s *Store, e2 int64)
		}{
			{"nothing measured in the new epoch (c1)", "c1", nil, nil},
			{"a new commit measured only by the old holder", "c2", func(t *testing.T, s *Store, e1 int64) { gateOn(t, s, e1, "c2", ResultPass) }, nil},
			{"c1 measured again with a failing green in the new epoch", "c1", nil, func(t *testing.T, s *Store, e2 int64) {
				greenOn(t, s, e2, "c1", ResultFail)
				gateOn(t, s, e2, "c1", ResultPass)
			}},
		} {
			t.Run(name+", refused: "+tc.name, func(t *testing.T) {
				s, e2 := landedThenReclaimed(t, rework, tc.before)
				if tc.record != nil {
					tc.record(t, s, e2)
				}
				if rework {
					// A card without a receipt of its own may not hand off; it is offered from where it is.
					_ = s.Move(bg, "a", e2, posReviewing, "lane", "rework")
				}
				before, _ := queueRow(t, s, "a")
				if err := s.Enqueue(bg, "a", e2, tc.sha); err == nil {
					t.Fatal("the new holder offered a commit it did not measure in its own epoch")
				} else if !errors.Is(err, ErrInvalid) && !errors.Is(err, ErrEvidenceStale) {
					t.Fatalf("refused for another reason: %v", err)
				}
				if after, _ := queueRow(t, s, "a"); after != before {
					t.Errorf("the refused offer changed the row: %+v -> %+v", before, after)
				}
			})
		}
	}
}
