package ledger

import (
	"errors"
	"testing"
)

// A lander that could not take the landing lock leaves its row verifying; the next Land
// enters verifying again at the card epoch. Only the row's own card epoch may do that.
func TestQueueUpdate_VerifyingReentry_AtCardEpoch(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
		t.Errorf("verifying -> verifying at the card epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch+1, "c1"), QueueVerifying, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("verifying -> verifying with another epoch: %v", err)
	}
	items, _ := s.Queue(bg)
	if len(items) != 1 || items[0].State != QueueVerifying {
		t.Fatalf("the row stays verifying: %+v", items)
	}
	// the re-entry does not widen the table: landing and landed cannot go back to verifying
	lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanding, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); !errors.Is(err, ErrInvalid) {
		t.Errorf("landing -> verifying: %v", err)
	}
}
