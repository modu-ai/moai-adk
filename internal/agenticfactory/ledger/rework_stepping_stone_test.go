package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit 22a0afd247 N15, the stepping stone: c1 landed, the holder took the card back for
// rework and offered a second candidate c2, which was then withdrawn or failed in the queue; the holder
// then offered c1 again. The second offer of c1 used to count as "the candidate was offered after the
// move back", so the rework mark was never looked at and the card finished on the landing of c1
// with nothing about c1 recorded since it was taken back (and c2, the work of the rework, not in the
// trunk). The mark belongs to the first offer of the candidate that stands in the row, not to the
// latest Enqueue: offering the same commit again is not new work.

var posIntegrated = core.Position{Stage: core.StageMerging, RemoteIntegrated: true}

// steppingStone leaves the card in MERGING with c1 landed (again), after c2 came and went in between.
// c2Green says whether c2 carried a green as well as its gate (otherwise the no-handoff path: a gate only).
func steppingStone(t *testing.T, ending string, c2Green bool) (*Store, int64) {
	t.Helper()
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	e := c.Epoch
	if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", e)
	mustMove(t, s, e, posReviewing, "lane")
	mustMove(t, s, e, posImplement, "lane")
	kinds := []string{KindGate}
	if c2Green {
		kinds = []string{KindGreen, KindGate}
	}
	for _, k := range kinds {
		if err := s.AddEvidence(bg, "a", e, Evidence{Kind: k, TreeSHA: "tree-c2", SubjectSHA: "c2", Result: ResultPass}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Enqueue(bg, "a", e, "c2"); err != nil {
		t.Fatalf("enqueue c2: %v", err)
	}
	if ending == "failed" {
		if err := s.QueueUpdate(bg, "a", CardEpoch(e, "c2"), QueueFailed, "", "c2 breaks the build"); err != nil {
			t.Fatal(err)
		}
	}
	mustMove(t, s, e, posReviewing, "lane")
	if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
		t.Fatalf("enqueue c1 again: %v", err)
	}
	landOne(t, s, "a", e)
	return s, e
}

// finish tries the two moves that end a landed card and reports the first refusal.
func finish(s *Store, e int64) error {
	if err := s.Move(bg, "a", e, posIntegrated, "mergeq", "merge seen"); err != nil {
		return err
	}
	return s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "mergeq", "landed")
}

func TestRework_OfferingTheLandedCommitAgainDoesNotWashOutTheReworkMark(t *testing.T) {
	for _, ending := range []string{"failed", "withdrawn"} {
		for _, c2Green := range []bool{false, true} {
			name := "c2 " + ending
			if c2Green {
				name += " with a green"
			} else {
				name += " gate only"
			}
			t.Run(name, func(t *testing.T) {
				s, e := steppingStone(t, ending, c2Green)
				if err := finish(s, e); !errors.Is(err, core.ErrGuardMissing) {
					t.Fatalf("the card finished on c1, offered again, with nothing about c1 recorded since the move back: %v", err)
				}
				if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
					t.Error("the card is DONE")
				}
			})
			// Control: the honest holder measures c1 again after the move back and finishes.
			t.Run(name+", c1 measured again", func(t *testing.T) {
				s, e := steppingStone(t, ending, c2Green)
				evidenceOn(t, s, e, "c1")
				if err := finish(s, e); err != nil {
					t.Fatalf("c1 was measured again after the move back, and the card is refused: %v", err)
				}
			})
		}
	}
}

// The first offer of a commit is where its mark starts, however the card got there: a commit
// offered once, failed in the queue, then offered again after the card went back, is the same stone.
// Nothing in the record says what the rework changed about it, so the second offer lands it on
// the old measurements only if the holder measures it again after the move back.
func TestRework_AFailedCandidateOfferedAgainAfterTheMoveBackNeedsNewEvidence(t *testing.T) {
	for _, measured := range []bool{false, true} {
		name := "nothing recorded since the move back"
		if measured {
			name = "c1 measured again since the move back"
		}
		t.Run(name, func(t *testing.T) {
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
			mustMove(t, s, e, posImplement, "lane")
			if measured {
				evidenceOn(t, s, e, "c1")
			}
			mustMove(t, s, e, posReviewing, "lane")
			if err := s.Enqueue(bg, "a", e, "c1"); err != nil {
				t.Fatal(err)
			}
			landOne(t, s, "a", e)
			err := finish(s, e)
			if measured && err != nil {
				t.Fatalf("c1 was measured again after the move back, and the card is refused: %v", err)
			}
			if !measured && !errors.Is(err, core.ErrGuardMissing) {
				t.Fatalf("the card finished on a commit offered again after the move back, nothing measured since: %v", err)
			}
		})
	}
}
