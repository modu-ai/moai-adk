package ledger

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, the stale receipt: the previous fix looked only at the stage the card is in. A card
// whose candidate c1 landed, which was taken back for rework and then handed off again with
// no new evidence, sits in REVIEWING with c1's old green and gate pair (its receipt) and c1's
// landed row, and finished again on those two alone. The ledger now keeps, on each transition, how much
// evidence the card had when the move was made, so "evidence recorded after the move back" is a
// fact stored in the ledger and not a guess from timestamps or from the position.

func openAt(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "af", "ledger.db")
	s, err := Open(bg, path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s, path
}

// reopen closes the ledger and opens the same file again: a process restart.
func reopen(t *testing.T, s *Store, path string) *Store {
	t.Helper()
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	n, err := Open(bg, path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = n.Close() })
	return n
}

// landedBackedOutAt: c1 is gated, queued and lands while the card is in MERGING; the holder then
// takes the card out of MERGING to REVIEWING (E10).
func landedBackedOutAt(t *testing.T) (*Store, string, int64) {
	t.Helper()
	s, path := openAt(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", c.Epoch)
	backOut(t, s, c.Epoch)
	return s, path, c.Epoch
}

var (
	posImplement = core.Position{Stage: core.StageBuilding, Sub: core.SubImplement}
	posFix       = core.Position{Stage: core.StageBuilding, Sub: core.SubFix}
	posSpec      = core.Position{Stage: core.StageBuilding, Sub: core.SubSpec}
	posReviewing = core.Position{Stage: core.StageReviewing}
	posFixParked = core.Position{Stage: core.StageBuilding, Sub: core.SubFix,
		Park: &core.Park{Reason: core.ReasonContractChange, Evidence: "spec change", Resume: "adopt the revision"}}
)

func walk(t *testing.T, s *Store, epoch int64, route []core.Position) {
	t.Helper()
	for _, to := range route {
		if err := s.Move(bg, "a", epoch, to, "lane", "rework"); err != nil {
			t.Fatalf("move to %+v: %v", to, err)
		}
	}
}

// The return paths after a landing. The card has left MERGING for REVIEWING (E10) and goes into
// BUILDING by each way the table allows, then hands off again with the receipt it already had.
// MERGING cannot go to BUILDING in one move (no edge): that is asserted below.
var reworkRoutes = []struct {
	name  string
	moves []core.Position
}{
	{"IMPLEMENT", []core.Position{posImplement, posReviewing}},
	{"FIX", []core.Position{posFix, posReviewing}},
	{"SPEC through the parked FIX and an adopted revision", []core.Position{posFix, posFixParked, posSpec, posImplement, posReviewing}},
}

func TestMove_DoneNeedsWorkRecordedAfterTheMoveBack(t *testing.T) {
	for _, rt := range reworkRoutes {
		for _, restart := range []bool{false, true} {
			name := rt.name
			if restart {
				name += " after a process restart"
			}
			t.Run(name, func(t *testing.T) {
				s, path, e := landedBackedOutAt(t)
				walk(t, s, e, rt.moves)
				if restart {
					s = reopen(t, s, path)
				}
				err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done")
				if !errors.Is(err, core.ErrGuardMissing) {
					t.Fatalf("DONE on the old receipt and the landed row, nothing recorded since the move back: %v", err)
				}
				if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
					t.Error("the card is DONE")
				}
			})
			// Decision (issue-1948): the SAME commit measured again after the move back finishes the
			// card. c1 is what the trunk holds and what the holder just measured, so nothing is lost,
			// and refusing it would strand a card whose rework changed no commit: Enqueue refuses a
			// landed candidate, so it could never be queued again either.
			t.Run(name+", same commit measured again", func(t *testing.T) {
				s, path, e := landedBackedOutAt(t)
				walk(t, s, e, rt.moves)
				evidenceOn(t, s, e, "c1")
				if restart {
					s = reopen(t, s, path)
				}
				if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done"); err != nil {
					t.Fatalf("DONE after c1 was measured again since the move back: %v", err)
				}
			})
		}
	}
}

// Left MERGING straight to REVIEWING (no rework): nothing says the work was taken back, so the
// landed commit finishes the card, as TestMove_DoneNeedsTheLandedCandidateToBeTheNewestWork
// and the audit's TestA1Done_WithdrawnLanded_Direct require.
func TestMove_DoneAfterLeavingMergingWithoutReworkStillFinishes(t *testing.T) {
	for _, restart := range []bool{false, true} {
		s, path, e := landedBackedOutAt(t)
		if restart {
			s = reopen(t, s, path)
		}
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done"); err != nil {
			t.Errorf("restart=%v: %v", restart, err)
		}
	}
}

// MERGING has no edge into BUILDING: the only way back to rework runs through REVIEWING.
func TestMove_MergingGoesToBuildingOnlyThroughReviewing(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", c.Epoch)
	if err := s.Move(bg, "a", c.Epoch, posImplement, "lane", "rework"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("MERGING straight to BUILDING: %v", err)
	}
}

// A landing that settles after the card already went back for rework (the swap window: the row
// is landing while the holder withdraws) is a landed row older than the rework. It must not
// finish the card either.
func TestMove_DoneRefusedWhenTheRowLandedAfterTheCardWentBackForRework(t *testing.T) {
	for _, restart := range []bool{false, true} {
		t.Run(map[bool]string{false: "same process", true: "after a process restart"}[restart], func(t *testing.T) {
			s, path := openAt(t)
			c := gated(t, s, "a", "c1")
			if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
				t.Fatal(err)
			}
			if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
				t.Fatal(err)
			}
			lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq", PID: 1, PIDStart: "t"}, LandingLockTTL)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanding, "trunk", ""); err != nil {
				t.Fatal(err)
			}
			backOut(t, s, c.Epoch)
			walk(t, s, c.Epoch, []core.Position{posImplement})
			if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanded, "trunk", ""); err != nil {
				t.Fatal(err)
			}
			walk(t, s, c.Epoch, []core.Position{posReviewing})
			if restart {
				s = reopen(t, s, path)
			}
			if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageDone}, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
				t.Errorf("DONE on a landing that settled after the move back: %v", err)
			}
		})
	}
}

// What does not count as work recorded after the move back.
func TestMove_DoneIgnoresEvidenceThatIsNotTheLandedCommitMeasuredAgain(t *testing.T) {
	cases := []struct {
		name string
		add  func(t *testing.T, s *Store, e int64)
	}{
		{"a gate on another commit", func(t *testing.T, s *Store, e int64) {
			if err := s.AddEvidence(bg, "a", e, Evidence{Kind: KindGate, TreeSHA: "tree-c2", SubjectSHA: "c2", Result: ResultPass}); err != nil {
				t.Fatal(err)
			}
		}},
		{"a failing run of the landed commit", func(t *testing.T, s *Store, e int64) {
			for _, k := range []string{KindGreen, KindGate} {
				if err := s.AddEvidence(bg, "a", e, Evidence{Kind: k, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail}); err != nil {
					t.Fatal(err)
				}
			}
		}},
		{"a review of the landed commit", func(t *testing.T, s *Store, e int64) {
			if err := s.AddEvidence(bg, "a", e, Evidence{Kind: KindReview, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultPass}); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, _, e := landedBackedOutAt(t)
			walk(t, s, e, []core.Position{posImplement})
			tc.add(t, s, e)
			done := core.Position{Stage: core.StageDone}
			// The no-handoff path, straight from the build stage ...
			if err := s.Move(bg, "a", e, done, "lane", "done"); err == nil {
				t.Fatal("DONE from BUILDING/IMPLEMENT on the landed row")
			}
			// ... and the receipt path, where the hand-off itself may be refused (a failing run is no
			// receipt) but must not end in DONE when it is accepted.
			if err := s.Move(bg, "a", e, posReviewing, "lane", "handoff"); err == nil {
				if err := s.Move(bg, "a", e, done, "lane", "done"); err == nil {
					t.Fatal("DONE from REVIEWING on the old receipt and the landed row")
				}
			}
		})
	}
}

// A reworked card is never permanently blocked: a new commit goes through the queue, lands and
// finishes. Before the fix Enqueue refused any card whose row had landed ("has landed"), so
// the card with a new commit had no way to the trunk at all.
func TestRework_ANewCommitGoesThroughTheQueueLandsAndFinishes(t *testing.T) {
	for _, restart := range []bool{false, true} {
		s, path, e := landedBackedOutAt(t)
		walk(t, s, e, []core.Position{posImplement})
		evidenceOn(t, s, e, "c2")
		walk(t, s, e, []core.Position{posReviewing})
		if restart {
			s = reopen(t, s, path)
		}
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("restart=%v: DONE while c2 is the newest work and only c1 landed: %v", restart, err)
		}
		if err := s.Enqueue(bg, "a", e, "c2"); err != nil {
			t.Fatalf("restart=%v: the new commit cannot go through the queue: %v", restart, err)
		}
		if it, ok := queueRow(t, s, "a"); !ok || it.CandidateSHA != "c2" || it.State != QueueWaiting {
			t.Fatalf("the row after Enqueue: %+v", it)
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
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "mergeq", "early"); err == nil {
			t.Fatal("DONE before c2 landed")
		}
		landOne(t, s, "a", e)
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageMerging, RemoteIntegrated: true}, "mergeq", "merge seen"); err != nil {
			t.Fatalf("merge seen: %v", err)
		}
		if err := s.Move(bg, "a", e, core.Position{Stage: core.StageDone}, "mergeq", "landed"); err != nil {
			t.Fatalf("restart=%v: DONE after c2 landed: %v", restart, err)
		}
	}
}

// A landed row is replaced only by new work after a move back into BUILDING: a card that left
// MERGING straight to REVIEWING offering another commit is refused as before.
func TestEnqueue_ALandedRowIsNotReplacedByACardNotTakenBackForRework(t *testing.T) {
	s, _, e := landedBackedOutAt(t)
	evidenceOn(t, s, e, "c2")
	err := s.Enqueue(bg, "a", e, "c2")
	if !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), "has landed") {
		t.Fatalf("Enqueue of c2 from REVIEWING with no move back into BUILDING: %v", err)
	}
	if it, _ := queueRow(t, s, "a"); it.State != QueueLanded || it.CandidateSHA != "c1" {
		t.Errorf("the row changed: %+v", it)
	}
}

// The landed candidate itself is still not queued again: the trunk has it, and the way to finish
// is the landed row, not a second landing.
func TestEnqueue_TheLandedCandidateItselfStaysRefused(t *testing.T) {
	s, _, e := landedBackedOutAt(t)
	walk(t, s, e, []core.Position{posImplement})
	evidenceOn(t, s, e, "c1")
	walk(t, s, e, []core.Position{posReviewing})
	err := s.Enqueue(bg, "a", e, "c1")
	if !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), "has landed") {
		t.Fatalf("Enqueue of the landed candidate: %v", err)
	}
	if it, _ := queueRow(t, s, "a"); it.State != QueueLanded || it.CandidateSHA != "c1" {
		t.Errorf("the row changed: %+v", it)
	}
}
