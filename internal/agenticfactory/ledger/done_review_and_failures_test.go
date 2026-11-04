package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit 46dbcce0d4 N40. Two holes in the rule that a landed candidate may finish the card.
//
//  1. The newest-result judgment of the candidate read green and gate only. A review is the third
//     kind a card is judged on (a risk card cannot be offered without a passing one), so a review of
//     the landed commit that FAILS after the landing, or while the row waits, must refuse the card
//     the same way a failing gate does. Merge and red evidence are about other commits by design and
//     stay out.
//  2. A FAILURE about a different commit, recorded after the candidate was offered, means the card
//     found something wrong with the work it was doing beside the landed commit; the ledger cannot
//     tell whether that work was given up, so the candidate has to be measured again afterwards,
//     whether the failure came before or after the card left MERGING. A PASSING gate of another
//     commit still blocks nothing (D5, REQ-AFL-008: a gate-only candidate is accepted).

// riskLanded takes a card with a risk through green, gate, a passing review, REVIEWING, Enqueue and
// the landing of c1. The card is still in MERGING.
func riskLanded(t *testing.T) (*Store, int64) {
	t.Helper()
	s := openTemp(t)
	c := gated(t, s, "a", "c1", "security")
	reviewOn(t, s, c.Epoch, "c1", ResultPass)
	mustMove(t, s, c.Epoch, posReviewing, "lane")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", c.Epoch)
	return s, c.Epoch
}

func landedInMerging(t *testing.T) (*Store, int64) {
	t.Helper()
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", c.Epoch)
	return s, c.Epoch
}

func recordOn(t *testing.T, s *Store, e int64, kind, sha, result string) {
	t.Helper()
	if err := s.AddEvidence(bg, "a", e, Evidence{Kind: kind, TreeSHA: "tree-" + sha, SubjectSHA: sha, Result: result}); err != nil {
		t.Fatal(err)
	}
}

// tryDone is the card's way to DONE from where it is: through the integrated move while in MERGING,
// directly once it has been moved out.
func tryDone(s *Store, e int64, withdrawn bool) error {
	if withdrawn {
		return s.Move(bg, "a", e, posDone, "lane", "done")
	}
	return finish(s, e)
}

func notDone(t *testing.T, s *Store) {
	t.Helper()
	if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
		t.Fatal("the card is DONE")
	}
}

func TestDone_AFailingReviewOfTheLandedCommitRefusesTheCard(t *testing.T) {
	t.Run("a risk card, the review fails after the landing", func(t *testing.T) {
		s, e := riskLanded(t)
		reviewOn(t, s, e, "c1", ResultFail)
		if err := finish(s, e); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("the card finished although the review of its landed commit failed: %v", err)
		}
		notDone(t, s)
		reviewOn(t, s, e, "c1", ResultPass)
		if err := finish(s, e); err != nil {
			t.Fatalf("the review passes again and the card is refused: %v", err)
		}
	})
	t.Run("a card withdrawn to REVIEWING, the review fails", func(t *testing.T) {
		s, e := landedThenWithdrawn(t)
		reviewOn(t, s, e, "c1", ResultFail)
		if err := s.Move(bg, "a", e, posDone, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("DONE with a failing review of the landed commit: %v", err)
		}
		notDone(t, s)
		reviewOn(t, s, e, "c1", ResultPass)
		if err := s.Move(bg, "a", e, posDone, "lane", "done"); err != nil {
			t.Fatalf("the review passes again and the card is refused: %v", err)
		}
	})
	t.Run("a review that failed and then passed is repaired, and one that passed and then failed is not", func(t *testing.T) {
		s, e := landedInMerging(t)
		reviewOn(t, s, e, "c1", ResultFail)
		reviewOn(t, s, e, "c1", ResultPass)
		if err := finish(s, e); err != nil {
			t.Fatalf("fail then pass: %v", err)
		}
	})
	t.Run("the row waits and the review fails before the landing", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		reviewOn(t, s, c.Epoch, "c1", ResultFail)
		landOne(t, s, "a", c.Epoch)
		if err := finish(s, c.Epoch); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("the card finished although a review of its candidate failed while the row waited: %v", err)
		}
	})
}

// merge and red evidence are about other commits by design: they do not read as the landed commit's
// result.
func TestDone_AFailingMergeOrRedRecordIsNotTheLandedCommitsResult(t *testing.T) {
	for _, kind := range []string{KindMerge, KindRed} {
		t.Run(kind, func(t *testing.T) {
			s, e := landedInMerging(t)
			recordOn(t, s, e, kind, "c1", ResultFail)
			if err := finish(s, e); err != nil {
				t.Fatalf("a failing %s record about the landed commit refused the card: %v", kind, err)
			}
		})
	}
}

func TestDone_AFailureAboutAnotherCommitAfterTheOfferNeedsTheCandidateMeasuredAgain(t *testing.T) {
	for _, kind := range []string{KindGate, KindGreen, KindReview} {
		for _, withdrawn := range []bool{false, true} {
			name := kind + " of c2 fails, the card stays in MERGING"
			if withdrawn {
				name = kind + " of c2 fails while in MERGING, then the card is withdrawn"
			}
			t.Run(name, func(t *testing.T) {
				s, e := landedInMerging(t)
				recordOn(t, s, e, kind, "c2", ResultFail)
				if withdrawn {
					backOut(t, s, e)
				}
				if err := tryDone(s, e, withdrawn); !errors.Is(err, core.ErrGuardMissing) {
					t.Fatalf("the card finished although %s of c2 failed after c1 was offered: %v", kind, err)
				}
				notDone(t, s)
				evidenceOn(t, s, e, "c1")
				if err := tryDone(s, e, withdrawn); err != nil {
					t.Fatalf("c1 measured again after the failure, and the card is refused: %v", err)
				}
			})
		}
	}
	t.Run("the failure comes after the withdrawal", func(t *testing.T) {
		for _, kind := range []string{KindGate, KindGreen, KindReview} {
			s, e := landedThenWithdrawn(t)
			recordOn(t, s, e, kind, "c2", ResultFail)
			if err := s.Move(bg, "a", e, posDone, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
				t.Fatalf("%s of c2 failed after the withdrawal: %v", kind, err)
			}
			evidenceOn(t, s, e, "c1")
			if err := s.Move(bg, "a", e, posDone, "lane", "done"); err != nil {
				t.Fatalf("%s: c1 measured again and the card is refused: %v", kind, err)
			}
		}
	})
	t.Run("the latest of two failures needs a measurement after it", func(t *testing.T) {
		s, e := landedInMerging(t)
		recordOn(t, s, e, KindGate, "c2", ResultFail)
		evidenceOn(t, s, e, "c1")
		recordOn(t, s, e, KindReview, "c3", ResultFail)
		if err := finish(s, e); !errors.Is(err, core.ErrGuardMissing) {
			t.Fatalf("a measurement of c1 older than the second failure was taken for enough: %v", err)
		}
		evidenceOn(t, s, e, "c1")
		if err := finish(s, e); err != nil {
			t.Fatalf("c1 measured after both failures: %v", err)
		}
	})
}

// Kept: a PASSING record about another commit does not block (D5), whether it was recorded while the
// card was in MERGING or before it was withdrawn, and a failure about the landed commit's own kind of
// record is the landed commit's business (rule 1), not another commit's.
func TestDone_APassingRecordAboutAnotherCommitBeforeTheWithdrawalBlocksNothing(t *testing.T) {
	t.Run("gate of c2 passes, the card stays in MERGING", func(t *testing.T) {
		s, e := landedInMerging(t)
		gateOn(t, s, e, "c2", ResultPass)
		if err := finish(s, e); err != nil {
			t.Fatalf("a passing gate of a second candidate refused the card: %v", err)
		}
	})
	t.Run("gate of c2 passes while in MERGING, then the card is withdrawn", func(t *testing.T) {
		s, e := landedInMerging(t)
		gateOn(t, s, e, "c2", ResultPass)
		backOut(t, s, e)
		if err := s.Move(bg, "a", e, posDone, "lane", "done"); err != nil {
			t.Fatalf("a passing gate of c2 recorded before the withdrawal refused the card: %v", err)
		}
	})
}

// A failure about another commit from BEFORE the candidate was offered is the Enqueue check's
// business (the newest evidence must be the candidate's), not a reason to demand a green the
// gate-only candidate never had (D5).
func TestDone_AFailureAboutAnotherCommitBeforeTheOfferBlocksNoGateOnlyCandidate(t *testing.T) {
	s := openTemp(t)
	c := claimOne(t, s, "a", "internal/a/x.go")
	recordOn(t, s, c.Epoch, KindGreen, "c0", ResultFail)
	recordOn(t, s, c.Epoch, KindGate, "c1", ResultPass)
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	landOne(t, s, "a", c.Epoch)
	if err := finish(s, c.Epoch); err != nil {
		t.Fatalf("a gate-only candidate landed and an older failure of c0 refuses the card: %v", err)
	}
}

// A review of another commit recorded after the card left MERGING is evidence the card moved on, as a
// gate or a green is.
func TestDone_AReviewOfAnotherCommitAfterTheWithdrawalNeedsTheCandidateMeasuredAgain(t *testing.T) {
	s, e := landedThenWithdrawn(t)
	reviewOn(t, s, e, "c2", ResultPass)
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("DONE with a review of c2 recorded after the withdrawal and c1 not measured again: %v", err)
	}
	evidenceOn(t, s, e, "c1")
	if err := s.Move(bg, "a", e, posDone, "lane", "done"); err != nil {
		t.Fatalf("c1 measured again after the review, and the card is refused: %v", err)
	}
}
