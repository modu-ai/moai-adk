//go:build afred

package ledger

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// parkInPlace parks the card at its current stage the way an operator hold does.
func parkInPlace(t *testing.T, s *Store, id string, epoch int64) {
	t.Helper()
	cur := cardOf(t, s, id).Position
	to := cur
	to.Park = &core.Park{Reason: core.ReasonOperatorHold, Evidence: "paused by the operator", Resume: "the operator resumes"}
	if err := s.Move(bg, id, epoch, to, "operator", "hold", Evidence{Kind: KindGate, TreeSHA: "n/a", SubjectSHA: "n/a", Result: ResultPass, Guards: []string{core.GuardOperatorHold}}); err != nil {
		t.Fatalf("park: %v", err)
	}
}

// Final audit 5d90e66fe7 F1 (second half): a hold that lands mid-gate stops the row from being
// taken into verifying or landing; ending an attempt and settling a landing row stay possible
// because the trunk may already have moved.
func TestQueueUpdate_ParkedCardCannotAdvanceItsRow(t *testing.T) {
	t.Run("verifying", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		parkInPlace(t, s, "a", c.Epoch)
		err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", "")
		if !errors.Is(err, ErrStaleEpoch) {
			t.Fatalf("verifying for a parked card: %v", err)
		}
		if it := rowOf(t, s); it.State != QueueWaiting {
			t.Errorf("the row advanced: %+v", it)
		}
	})
	t.Run("landing", func(t *testing.T) {
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
		parkInPlace(t, s, "a", c.Epoch)
		if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanding, "t0", ""); !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("landing for a parked card: %v", err)
		}
	})
	t.Run("landed, failed and waiting stay possible", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
		if err != nil {
			t.Fatal(err)
		}
		for _, st := range []struct {
			ep    QueueEpoch
			state QueueState
		}{{CardEpoch(c.Epoch, "c1"), QueueVerifying}, {LockEpoch(lock, "c1"), QueueLanding}} {
			if err := s.QueueUpdate(bg, "a", st.ep, st.state, "t0", ""); err != nil {
				t.Fatal(err)
			}
		}
		parkInPlace(t, s, "a", c.Epoch)
		if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueWaiting, "", ""); err != nil {
			t.Errorf("reconcile back to waiting for a parked card: %v", err)
		}
		if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueFailed, "", "paused"); err != nil {
			t.Errorf("failing the row of a parked card: %v", err)
		}
	})
	t.Run("landed after landing", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
		if err != nil {
			t.Fatal(err)
		}
		for _, st := range []struct {
			ep    QueueEpoch
			state QueueState
		}{{CardEpoch(c.Epoch, "c1"), QueueVerifying}, {LockEpoch(lock, "c1"), QueueLanding}} {
			if err := s.QueueUpdate(bg, "a", st.ep, st.state, "t0", ""); err != nil {
				t.Fatal(err)
			}
		}
		parkInPlace(t, s, "a", c.Epoch)
		if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanded, "", ""); err != nil {
			t.Errorf("landed after landing for a parked card: %v", err)
		}
	})
}

// F6: a card withdrawn from MERGING leaves a waiting row at its own epoch; coming back at the
// same epoch must be able to queue it again.
func TestEnqueue_WithdrawnWaitingRowIsRestartedAtTheSameEpoch(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	backOut(t, s, c.Epoch)
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatalf("queueing a withdrawn card again at its own epoch: %v", err)
	}
	if it := rowOf(t, s); it.State != QueueWaiting || it.CandidateSHA != "c1" || it.CardEpoch != c.Epoch {
		t.Errorf("%+v", it)
	}
	if got := cardOf(t, s, "a").Position.Stage; got != core.StageMerging {
		t.Errorf("the card is at %s, not MERGING", got)
	}
}

// The restart is for a waiting row of a card that is not in MERGING. A waiting row of a card in
// MERGING, and a verifying row of a withdrawn card (a lander may be working it), stay refused.
func TestEnqueue_OnlyAWithdrawnWaitingRowIsRestartedInPlace(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); !errors.Is(err, ErrInvalid) {
		t.Errorf("a waiting row of a card still in MERGING: %v", err)
	}
	s2 := openTemp(t)
	c2 := gated(t, s2, "b", "c1")
	if err := s2.Enqueue(bg, "b", c2.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s2.QueueUpdate(bg, "b", CardEpoch(c2.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s2.Move(bg, "b", c2.Epoch, core.Position{Stage: core.StageReviewing}, "lane", "back out"); err != nil {
		t.Fatal(err)
	}
	if err := s2.Enqueue(bg, "b", c2.Epoch, "c1"); !errors.Is(err, ErrInvalid) {
		t.Errorf("a verifying row of a withdrawn card: %v", err)
	}
}
