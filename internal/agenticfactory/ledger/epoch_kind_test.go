//go:build afred

package ledger

import (
	"errors"
	"testing"
)

// The card epoch and the landing lock's epoch count independently and can be the same
// number. A caller presents its epoch with its kind, and a write whose table row needs the
// other kind is refused, whatever the numbers say.
func TestQueueUpdate_CoincidingEpochs_KindDecidesWhichTableIsChecked(t *testing.T) {
	s := openTemp(t)
	c := gated(t, s, "a", "c1") // card epoch 1
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "other-lander"}, LandingLockTTL) // lock epoch 1
	if err != nil || lock != c.Epoch {
		t.Fatalf("the two epochs must coincide for this case: card %d lock %d %v", c.Epoch, lock, err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueVerifying, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("verifying with a lock epoch of the same number: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueLanding, "t0", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("landing with the card epoch, the lock's number being the same: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch), QueueWaiting, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("an old holder's waiting write with its card epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueFailed, "", "red"); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("failed with a lock epoch of the same number: %v", err)
	}
	items, _ := s.Queue(bg)
	if len(items) != 1 || items[0].State != QueueVerifying {
		t.Errorf("a refused write must leave the row verifying: %+v", items)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock), QueueLanding, "t0", ""); err != nil {
		t.Errorf("landing with the lock epoch: %v", err)
	}
}
