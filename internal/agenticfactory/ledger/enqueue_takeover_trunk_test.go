package ledger

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit 46dbcce0d4 N39 (ledger side). The take-over of a landed row (N3) believed the row's
// own word that its candidate reached the trunk. That word can be wrong: the row may have landed on
// another trunk (a ledger from before the trunk was fixed), or the trunk may have lost the commit
// since. The ledger never runs git (REQ-AFL-008), so the caller says whether the candidate is on the
// trunk: OnTrunk(true) takes the row over, OnTrunk(false) or no option starts the row over as
// waiting, and the commit has to land for real before the card can finish.

func TestEnqueue_ATakeOverAsksTheCallerWhetherTheCandidateIsOnTheTrunk(t *testing.T) {
	for _, rework := range []bool{false, true} {
		name := "withdrawn to REVIEWING"
		if rework {
			name = "taken back through BUILDING"
		}
		setup := func(t *testing.T) (*Store, int64) {
			t.Helper()
			s, e2 := landedThenReclaimed(t, rework, nil)
			evidenceOn(t, s, e2, "c1")
			handOff(t, s, e2, rework)
			return s, e2
		}
		t.Run(name+", on the trunk: the row is taken over", func(t *testing.T) {
			s, e2 := setup(t)
			if err := s.Enqueue(bg, "a", e2, "c1", OnTrunk(true)); err != nil {
				t.Fatal(err)
			}
			if it, ok := queueRow(t, s, "a"); !ok || it.State != QueueLanded || it.CandidateSHA != "c1" || it.CardEpoch != e2 {
				t.Fatalf("the row after the take-over: %+v", it)
			}
			if err := finish(s, e2); err != nil {
				t.Fatalf("the card cannot finish on the landing it took over: %v", err)
			}
		})
		for _, tc := range []struct {
			name string
			opts []EnqueueOption
		}{
			{"not on the trunk", []EnqueueOption{OnTrunk(false)}},
			{"the caller did not say", nil},
		} {
			t.Run(name+", "+tc.name+": the row starts over as waiting", func(t *testing.T) {
				s, e2 := setup(t)
				if err := s.Enqueue(bg, "a", e2, "c1", tc.opts...); err != nil {
					t.Fatalf("the offer of a measured commit is refused: %v", err)
				}
				it, ok := queueRow(t, s, "a")
				if !ok || it.State != QueueWaiting || it.CandidateSHA != "c1" || it.CardEpoch != e2 || it.TrunkBefore != "" {
					t.Fatalf("the row must wait like a new offer: %+v", it)
				}
				if got := cardOf(t, s, "a").Position.Stage; got != core.StageMerging {
					t.Errorf("stage %s", got)
				}
				// It cannot finish on a landing nobody saw.
				if err := finish(s, e2); !errors.Is(err, core.ErrGuardMissing) {
					t.Fatalf("the card finished on a row that is not landed: %v", err)
				}
				notDone(t, s)
				notes, err := s.NotesSince(bg, 0)
				if err != nil {
					t.Fatal(err)
				}
				said := false
				for _, n := range notes {
					said = said || (n.CardID == "a" && strings.Contains(n.Text, "c1") && strings.Contains(n.Text, "not on the trunk"))
				}
				if !said {
					t.Errorf("no note says the landed row was not on the trunk: %+v", notes)
				}
				// The way out is to land for real.
				landOne(t, s, "a", e2)
				if err := finish(s, e2); err != nil {
					t.Fatalf("c1 landed again and the card cannot finish: %v", err)
				}
			})
		}
	}
}

// What the option does not touch.
func TestEnqueue_OnTrunkChangesNothingOutsideTheTakeOverOfALandedRow(t *testing.T) {
	for _, on := range []bool{true, false} {
		name := "OnTrunk(false)"
		if on {
			name = "OnTrunk(true)"
		}
		t.Run(name+" a new commit replaces the landed row", func(t *testing.T) {
			s, e2 := landedThenReclaimed(t, false, nil)
			evidenceOn(t, s, e2, "c2")
			if err := s.Enqueue(bg, "a", e2, "c2", OnTrunk(on)); err != nil {
				t.Fatal(err)
			}
			if it, ok := queueRow(t, s, "a"); !ok || it.State != QueueWaiting || it.CandidateSHA != "c2" {
				t.Fatalf("%+v", it)
			}
		})
		t.Run(name+" a first offer is an ordinary waiting row", func(t *testing.T) {
			s := openTemp(t)
			c := gated(t, s, "a", "c1")
			if err := s.Enqueue(bg, "a", c.Epoch, "c1", OnTrunk(on)); err != nil {
				t.Fatal(err)
			}
			if it, ok := queueRow(t, s, "a"); !ok || it.State != QueueWaiting {
				t.Fatalf("the option made a first offer something else: %+v", it)
			}
		})
		t.Run(name+" a landed row of the card's own epoch is still history", func(t *testing.T) {
			s, e := landedInMerging(t)
			if err := s.Enqueue(bg, "a", e, "c1", OnTrunk(on)); !errors.Is(err, ErrInvalid) {
				t.Fatalf("a landed row of this epoch was offered again: %v", err)
			}
		})
		t.Run(name+" nothing measured in the new epoch is still refused and changes nothing", func(t *testing.T) {
			s, e2 := landedThenReclaimed(t, false, nil)
			before, _ := queueRow(t, s, "a")
			if err := s.Enqueue(bg, "a", e2, "c1", OnTrunk(on)); !errors.Is(err, ErrInvalid) {
				t.Fatalf("the take-over of an unmeasured commit: %v", err)
			}
			if after, _ := queueRow(t, s, "a"); after != before {
				t.Errorf("the refused offer changed the row: %+v -> %+v", before, after)
			}
		})
	}
}
