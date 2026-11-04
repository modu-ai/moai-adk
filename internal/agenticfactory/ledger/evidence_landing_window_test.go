package ledger

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948, follow-up of the final code audit of a36ea8c368 (the window left open by N66). `af land` asks again, under
// the landing lock and after the row says landing, whether the candidate has a failing newest green, gate or review
// run, and only then moves the trunk ref. Evidence writes take no landing lock, so a failing run recorded after that
// read and before the ref moved was landed all the same. Both sides write through the ledger, so the ledger closes the
// window: the evidence transaction refuses a FAILING green, gate or review run about the candidate of a row of that card
// that is in landing. The row is set to landing before the check reads the ledger, so either the failing run commits
// first and the check sees it, or it commits after and is refused: no instant remains in which it is recorded and not
// seen. The refusal is retryable: the run is recorded after the landing (where the post-landing rule refuses DONE), or
// after the row has left landing.

// evidenceCount is the number of evidence rows of the card.
func evidenceCount(t *testing.T, s *Store, card string) int {
	t.Helper()
	ev, err := s.EvidenceFor(bg, card)
	if err != nil {
		t.Fatal(err)
	}
	return len(ev)
}

// rowInLanding is a card "a" in MERGING whose row for c1 is in landing, held by a lander that has not finished: the lock
// is still held and its epoch is returned.
func rowInLanding(t *testing.T) (s *Store, cardEpoch, lock int64) {
	t.Helper()
	s = openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	var err error
	lock, err = s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq", PID: 1, PIDStart: "t"}, LandingLockTTL)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		epoch QueueEpoch
		state QueueState
	}{{CardEpoch(c.Epoch, "c1"), QueueVerifying}, {LockEpoch(lock, "c1"), QueueLanding}} {
		if err := s.QueueUpdate(bg, "a", step.epoch, step.state, "trunk", ""); err != nil {
			t.Fatalf("%s: %v", step.state, err)
		}
	}
	return s, c.Epoch, lock
}

func TestLandingWindow_AFailingRunAboutTheCandidateOfALandingRowIsRefused(t *testing.T) {
	for _, kind := range []string{KindGreen, KindGate, KindReview} {
		t.Run(kind, func(t *testing.T) {
			s, e, _ := rowInLanding(t)
			before := evidenceCount(t, s, "a")
			err := s.AddEvidence(bg, "a", e, Evidence{Kind: kind, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail})
			if err == nil {
				t.Fatalf("a failing %s run about the candidate of a landing row was recorded", kind)
			}
			for _, want := range []string{"being landed", "reconcile", "moai af land"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the refusal does not say %q: %v", want, err)
				}
			}
			if got := evidenceCount(t, s, "a"); got != before {
				t.Errorf("a refused run changed the evidence: %d rows, was %d", got, before)
			}
			if got, ok := queueRow(t, s, "a"); !ok || got.State != QueueLanding {
				t.Errorf("the refusal moved the row: %+v", got)
			}
		})
	}
}

// What the refusal is not: it does not touch a passing run, a run about another commit, a kind the landing does not read,
// or another card.
func TestLandingWindow_TheRefusalIsAsNarrowAsTheCheckItClosesTheWindowOf(t *testing.T) {
	t.Run("a passing run about the candidate", func(t *testing.T) {
		s, e, _ := rowInLanding(t)
		for _, kind := range []string{KindGreen, KindGate, KindReview} {
			if err := s.AddEvidence(bg, "a", e, Evidence{Kind: kind, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultPass}); err != nil {
				t.Errorf("a passing %s run about the candidate was refused: %v", kind, err)
			}
		}
	})
	t.Run("a failing run about another commit", func(t *testing.T) {
		s, e, _ := rowInLanding(t)
		for _, kind := range []string{KindGreen, KindGate, KindReview} {
			if err := s.AddEvidence(bg, "a", e, Evidence{Kind: kind, TreeSHA: "tree-c2", SubjectSHA: "c2", Result: ResultFail}); err != nil {
				t.Errorf("a failing %s run about another commit was refused: %v", kind, err)
			}
		}
	})
	t.Run("a failing run of a kind the landing does not read", func(t *testing.T) {
		s, e, _ := rowInLanding(t)
		for _, kind := range []string{KindRed, KindMerge} {
			if err := s.AddEvidence(bg, "a", e, Evidence{Kind: kind, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail}); err != nil {
				t.Errorf("a failing %s run about the candidate was refused although no landing reads it: %v", kind, err)
			}
		}
	})
	t.Run("a failing run of another card about the same commit", func(t *testing.T) {
		s, _, _ := rowInLanding(t)
		b := gated(t, s, "b", "c1")
		if err := s.AddEvidence(bg, "b", b.Epoch, Evidence{Kind: KindGate, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail}); err != nil {
			t.Errorf("a failing run of card b was refused because card a lands the same commit: %v", err)
		}
	})
}

// The refusal belongs to landing and to nothing before or after it: a row that waits or verifies is covered by the
// check that reads the ledger, a row that landed is covered by the post-landing rule.
func TestLandingWindow_OnlyARowInLandingRefuses(t *testing.T) {
	fail := Evidence{Kind: KindGate, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail}
	t.Run("waiting", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		if err := s.AddEvidence(bg, "a", c.Epoch, fail); err != nil {
			t.Errorf("refused while the row waits: %v", err)
		}
	})
	t.Run("verifying", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "trunk", ""); err != nil {
			t.Fatal(err)
		}
		if err := s.AddEvidence(bg, "a", c.Epoch, fail); err != nil {
			t.Errorf("refused while the row verifies (the gates run in this state): %v", err)
		}
	})
	t.Run("landed", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		landOne(t, s, "a", c.Epoch)
		if err := s.AddEvidence(bg, "a", c.Epoch, fail); err != nil {
			t.Fatalf("the retry after the landing is refused: %v", err)
		}
		// the retry is not a way round the rule after landing: it is the rule after landing that holds the card
		if err := finish(s, c.Epoch); err == nil {
			t.Error("a failure recorded after the landing let the card finish")
		}
		if got := cardOf(t, s, "a").Position.Stage; got == core.StageDone {
			t.Error("the card is DONE")
		}
	})
	t.Run("back to waiting", func(t *testing.T) {
		s, e, lock := rowInLanding(t)
		if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueWaiting, "trunk", ""); err != nil {
			t.Fatal(err)
		}
		if err := s.AddEvidence(bg, "a", e, fail); err != nil {
			t.Errorf("refused after the row left landing: %v", err)
		}
	})
}

// A row left in landing by a lander that died keeps refusing until the row is settled: the lock is gone but the row says
// landing, and only a reconcile (or the next land) settles it.
func TestLandingWindow_ARowLeftInLandingByADeadLanderKeepsRefusingUntilItIsSettled(t *testing.T) {
	s, e, lock := rowInLanding(t)
	if err := s.ReleaseLock(bg, LockMerge, lock); err != nil {
		t.Fatal(err)
	}
	fail := Evidence{Kind: KindGate, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail}
	for i := 0; i < 2; i++ {
		if err := s.AddEvidence(bg, "a", e, fail); err == nil || !strings.Contains(err.Error(), "being landed") {
			t.Fatalf("try %d: the failing run was recorded against a row nobody is landing: %v", i+1, err)
		}
	}
	// the settlement: a reconcile holds the landing lock and sends the row back to waiting
	lock2, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "reconcile", PID: 1, PIDStart: "t"}, LandingLockTTL)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock2, "c1"), QueueWaiting, "trunk", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.AddEvidence(bg, "a", e, fail); err != nil {
		t.Fatalf("still refused after the row was settled: %v", err)
	}
}

// Move records evidence too (its guard evidence), through the same insert: the refusal is the insert's, so that path
// cannot be a way round it.
func TestLandingWindow_AMoveCannotCarryTheRefusedRunEither(t *testing.T) {
	s, e, _ := rowInLanding(t)
	before := evidenceCount(t, s, "a")
	err := s.Move(bg, "a", e, core.Position{Stage: core.StageReviewing}, "lane", "taking it back",
		Evidence{Kind: KindGate, TreeSHA: "tree-c1", SubjectSHA: "c1", Result: ResultFail})
	if err == nil || !strings.Contains(err.Error(), "being landed") {
		t.Fatalf("Move with a failing run about the landing candidate: %v", err)
	}
	if got := evidenceCount(t, s, "a"); got != before {
		t.Errorf("the refused move changed the evidence: %d rows, was %d", got, before)
	}
}
