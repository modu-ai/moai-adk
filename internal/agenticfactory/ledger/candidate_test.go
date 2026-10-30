package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Landing-lock writes name the candidate the lander verified (subject_sha is the authority).
func TestQueueUpdate_LockKindWriteNamesTheVerifiedCandidate(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, ""), QueueLanding, "t0", ""); !errors.Is(err, ErrInvalid) {
		t.Errorf("no candidate named: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "other"), QueueLanding, "t0", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("another candidate: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanding, "t0", ""); err != nil {
		t.Fatalf("the verified candidate: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "other"), QueueLanded, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("landed with another candidate: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanded, "", ""); err != nil {
		t.Errorf("landed with the verified candidate: %v", err)
	}
}

// lane-19 25th review P2 2: the old lander verified c1; the row was restarted with c2 and a
// new lander took it to verifying. The old lander gets the lock and writes landing for c1.
func TestQueueUpdate_OldLanderCannotMoveARestartedRow(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil { // old lander A
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReviewing}, "lane", "back out"); err != nil {
		t.Fatal(err)
	}
	if err := s.Release(bg, "a", c.Epoch); err != nil {
		t.Fatal(err)
	}
	c2, err := s.Claim(bg, req("next", 77))
	if err != nil || c2.Epoch != c.Epoch+1 {
		t.Fatalf("%+v %v", c2, err)
	}
	if err := s.AddEvidence(bg, "a", c2.Epoch, Evidence{Kind: KindGate, TreeSHA: "t2", SubjectSHA: "c2", Result: ResultPass}); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c2.Epoch, "c2"); err != nil { // restart with c2
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c2.Epoch, "c1"), QueueVerifying, "", ""); err != nil { // new lander B
		t.Fatal(err)
	}
	lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "A"}, LandingLockTTL)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanding, "t0", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Fatalf("the old lander moves the restarted row: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c2"), QueueLanding, "t0", ""); err != nil {
		t.Errorf("the new lander's candidate: %v", err)
	}
}

// P2 1: settling an older landing row names the candidate, moves the row to the card's
// current epoch when it lands, and the card can then finish.
func TestQueueUpdate_SettledLandedRowLetsTheCardFinish(t *testing.T) {
	s, e2, lock := staleLanding(t)
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, ""), QueueLanded, "", ""); !errors.Is(err, ErrInvalid) {
		t.Errorf("a settlement without a candidate: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c2"), QueueLanded, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a settlement naming another candidate: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanded, "", ""); err != nil {
		t.Fatal(err)
	}
	if it := rowOf(t, s); it.State != QueueLanded || it.CardEpoch != e2 || it.CandidateSHA != "c1" {
		t.Fatalf("the landed row follows the card's epoch: %+v", it)
	}
	for _, k := range []string{KindGreen, KindGate} {
		if err := s.AddEvidence(bg, "a", e2, Evidence{Kind: k, TreeSHA: "t1", SubjectSHA: "c1", Result: ResultPass}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Move(bg, "a", e2, core.Position{Stage: core.StageDone}, "lane", "the candidate landed"); err != nil {
		t.Errorf("the card finishes: %v", err)
	}
}
