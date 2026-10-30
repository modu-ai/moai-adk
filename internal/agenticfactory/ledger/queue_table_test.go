package ledger

import (
	"errors"
	"testing"
)

// Every (from, to) pair of queue states against the expected table, each with the epoch
// its target kind needs: the card epoch, or the landing lock's epoch for landing, landed
// and a reconcile back to waiting.
func TestQueueUpdate_TransitionTable_EveryPair(t *testing.T) {
	s := openTemp(t)
	if e, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "warm"}, LandingLockTTL); err != nil {
		t.Fatal(err)
	} else if err := s.ReleaseLock(bg, LockMerge, e); err != nil {
		t.Fatal(err)
	}
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	lock, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
	if err != nil || lock == c.Epoch {
		t.Fatalf("lock epoch %d must differ from card epoch %d: %v", lock, c.Epoch, err)
	}
	states := []QueueState{QueueWaiting, QueueVerifying, QueueLanding, QueueLanded, QueueFailed}
	allowed := map[QueueState][]QueueState{
		QueueWaiting:   {QueueVerifying, QueueFailed},
		QueueVerifying: {QueueVerifying, QueueLanding, QueueFailed, QueueWaiting},
		QueueLanding:   {QueueLanded, QueueFailed, QueueWaiting},
	}
	isAllowed := func(from, to QueueState) bool {
		for _, n := range allowed[from] {
			if n == to {
				return true
			}
		}
		return false
	}
	for _, from := range states {
		for _, to := range states {
			ep, wrong := c.Epoch, lock
			if to == QueueLanding || to == QueueLanded || (to == QueueWaiting && (from == QueueVerifying || from == QueueLanding)) {
				ep, wrong = lock, c.Epoch
			}
			reason := ""
			if to == QueueFailed {
				reason = "red"
			}
			reset := func() {
				if _, err := s.db.ExecContext(bg, `UPDATE merge_queue SET state=? WHERE card_id='a'`, string(from)); err != nil {
					t.Fatal(err)
				}
			}
			reset()
			// the epoch of the wrong kind is refused first, whatever the pair
			if err := s.QueueUpdate(bg, "a", wrong, to, "", reason); !errors.Is(err, ErrStaleEpoch) {
				t.Errorf("%s -> %s with the other kind's epoch: %v", from, to, err)
			}
			reset()
			err := s.QueueUpdate(bg, "a", ep, to, "", reason)
			switch {
			case isAllowed(from, to) && err != nil:
				t.Errorf("%s -> %s should be allowed: %v", from, to, err)
			case !isAllowed(from, to) && !errors.Is(err, ErrInvalid):
				t.Errorf("%s -> %s should be refused as invalid: %v", from, to, err)
			}
		}
	}
}

// A dead holder's lease goes at once; it does not wait for its expiry.
func TestReap_DeadHolderIsReclaimedBeforeItsLeaseExpires(t *testing.T) {
	s := openTemp(t)
	claimOne(t, s, "a", "internal/a/x.go")
	if countOf(t, s, `SELECT count(*) FROM leases WHERE expires_at > strftime('%Y-%m-%dT%H:%M:%fZ','now')`) != 1 {
		t.Fatal("the lease should still be running")
	}
	got, err := s.Reap(bg, alwaysDead)
	if err != nil || len(got) != 1 || got[0].CardID != "a" {
		t.Fatalf("%+v %v", got, err)
	}
}
