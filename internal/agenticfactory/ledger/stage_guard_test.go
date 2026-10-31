package ledger

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func backOut(t *testing.T, s *Store, epoch int64) {
	t.Helper()
	if err := s.Move(bg, "a", epoch, core.Position{Stage: core.StageReviewing}, "lane", "back out of the queue"); err != nil {
		t.Fatal(err)
	}
}

// A card that left MERGING at the row's own epoch must not have its row taken into
// verifying, landing or landed: the lander would merge work the card no longer offers.
func TestQueueUpdate_RowOfACardOutsideMergingCannotAdvance(t *testing.T) {
	t.Run("verifying", func(t *testing.T) {
		s := openTemp(t)
		c := gated(t, s, "a", "c1")
		if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
			t.Fatal(err)
		}
		backOut(t, s, c.Epoch)
		err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", "")
		if !errors.Is(err, ErrStaleEpoch) {
			t.Fatalf("verifying for a card at REVIEWING: %v", err)
		}
		if msg := err.Error(); !strings.Contains(msg, "verifying") || !strings.Contains(msg, "REVIEWING") || !strings.Contains(msg, "epoch 1") {
			t.Errorf("the message must name the step, the stage and the epoch: %q", msg)
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
		if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
			t.Fatal(err)
		}
		lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
		if err != nil {
			t.Fatal(err)
		}
		backOut(t, s, c.Epoch)
		if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanding, "t0", ""); !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("landing for a card at REVIEWING: %v", err)
		}
	})
	// landed is guarded when the row is not yet landing: a row that is only verifying cannot jump
	// there (not an edge). Once a row IS landing the trunk may have moved, so landed after the
	// card backed out is allowed and noted (TestQueueUpdate_LandedForAnAlreadyLanding...).
	t.Run("landed from verifying", func(t *testing.T) {
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
		backOut(t, s, c.Epoch)
		if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanded, "", ""); !errors.Is(err, ErrStaleEpoch) {
			t.Errorf("landed for a verifying row of a card at REVIEWING: %v", err)
		}
	})
}

// Failing a row, and sending a verifying or landing row back to waiting, stay possible for a
// card that left MERGING: they undo or end the attempt, they do not advance it.
func TestQueueUpdate_FailAndReconcileStillWorkOutsideMerging(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
	if err != nil {
		t.Fatal(err)
	}
	backOut(t, s, c.Epoch)
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueWaiting, "", ""); err != nil {
		t.Errorf("reconcile back to waiting: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueFailed, "", "the card left the queue"); err != nil {
		t.Errorf("failing the row: %v", err)
	}
}
