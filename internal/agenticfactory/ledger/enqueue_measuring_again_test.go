package ledger

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, audit 46dbcce0d4 N41 (and its parent N27). Enqueue takes a candidate only when it is the
// commit the card measured last, and the way to make an older commit the last one is to measure it
// again. Only a passing green or a passing gate is a measurement: the acceptance tests ran and
// passed, or the build gate did. A review is a reading of the diff, not a run, and a failing record
// of the commit is the opposite of a measurement; neither may hand the offer back to a commit whose
// newer sibling just failed. The audit's reproduction: c1 gated, then c2's green FAILED (Enqueue of
// c1 is refused, as it should be), then a FAILING review of c1 was recorded, and c1 was taken,
// landed, and the card finished with c2 not in the trunk.

func reviewOn(t *testing.T, s *Store, e int64, sha, result string) {
	t.Helper()
	if err := s.AddEvidence(bg, "a", e, Evidence{Kind: KindReview, TreeSHA: "tree-" + sha, SubjectSHA: sha, Result: result}); err != nil {
		t.Fatal(err)
	}
}

func notInTheQueue(t *testing.T, s *Store) {
	t.Helper()
	if it, ok := queueRow(t, s, "a"); ok {
		t.Fatalf("a refused offer left a queue row: %+v", it)
	}
	if got := cardOf(t, s, "a").Position.Stage; got == core.StageMerging || got == core.StageDone {
		t.Fatalf("a refused offer moved the card to %s", got)
	}
}

func TestEnqueue_AReviewOrAFailureOfTheCandidateIsNotMeasuringItAgain(t *testing.T) {
	t.Run("a failing review of c1 after c2's failing green", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		greenOn(t, s, c.Epoch, "c2", ResultFail)
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrEvidenceStale) {
			t.Fatalf("c1 offered with c2's failing green newest: %v", err)
		}
		reviewOn(t, s, c.Epoch, "c1", ResultFail)
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrEvidenceStale) {
			t.Fatalf("a FAILING review of c1 counted as measuring c1 again: %v", err)
		}
		notInTheQueue(t, s)
	})
	t.Run("a passing review of c1 after c2's failing green", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		greenOn(t, s, c.Epoch, "c2", ResultFail)
		reviewOn(t, s, c.Epoch, "c1", ResultPass)
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrEvidenceStale) {
			t.Fatalf("a review of c1 counted as measuring c1 again: %v", err)
		}
		notInTheQueue(t, s)
	})
	t.Run("a failing gate of c1 after c2's failing green", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		greenOn(t, s, c.Epoch, "c2", ResultFail)
		gateOn(t, s, c.Epoch, "c1", ResultFail)
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrEvidenceStale) {
			t.Fatalf("a failing gate of c1 was taken: %v", err)
		}
		notInTheQueue(t, s)
	})
	t.Run("measured again with a passing green and gate, the offer is taken", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		greenOn(t, s, c.Epoch, "c2", ResultFail)
		reviewOn(t, s, c.Epoch, "c1", ResultFail)
		evidenceOn(t, s, c.Epoch, "c1")
		// c1's newest review failed (a review is part of the newest-result judgment since N40), so
		// the way out is the measurement and a review that passes.
		reviewOn(t, s, c.Epoch, "c1", ResultPass)
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatalf("c1 measured again in full and reviewed again is refused: %v", err)
		}
	})
}

// A failing run of the candidate is not the newest evidence either: the refusal names the commit
// the card measured last, not the failed candidate (the failed run is refused on its own account
// only after that).
func TestEnqueue_AFailingRunOfTheCandidateDoesNotHideTheNewerCommit(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	greenOn(t, s, c.Epoch, "c2", ResultPass)
	greenOn(t, s, c.Epoch, "c1", ResultFail)
	err := s.Enqueue(bg, "a", c.Epoch, "c1")
	if !errors.Is(err, ErrEvidenceStale) || !strings.Contains(err.Error(), "about commit c2") {
		t.Fatalf("the refusal must say the newest evidence is about c2: %v", err)
	}
}

// What stays as it was: a review about ANOTHER commit as the newest evidence is still evidence
// the card moved on, so the older candidate has to be measured again.
func TestEnqueue_AReviewOfAnotherCommitIsStillTheNewestEvidence(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	reviewOn(t, s, c.Epoch, "c2", ResultPass)
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrEvidenceStale) {
		t.Fatalf("c1 offered with a review of c2 as the newest evidence: %v", err)
	}
	evidenceOn(t, s, c.Epoch, "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatalf("c1 measured again is refused: %v", err)
	}
}

// A risk card's own flow is untouched: green, gate, then the review of the same commit is the
// newest evidence, and it is taken.
func TestEnqueue_ARiskCardWhoseNewestEvidenceIsItsOwnReviewIsStillTaken(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1", "security")
	reviewOn(t, s, c.Epoch, "c1", ResultPass)
	mustMove(t, s, c.Epoch, posReviewing, "lane")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatalf("green, gate and review of c1: %v", err)
	}
}
