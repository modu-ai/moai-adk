//go:build afred

package ledger

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// landingRow takes a card to a row that is landing under the merge lock, then lets the card's
// holder back out of MERGING: the trunk may already have moved while the row says landing.
func landingRow(t *testing.T) (s *Store, lock, epoch int64) {
	t.Helper()
	s = openTemp(t)
	c := gated(t, s, "a", "c1")
	if err := s.Enqueue(bg, "a", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	var err error
	if lock, err = s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanding, "t0", ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "a", c.Epoch, core.Position{Stage: core.StageReviewing}, "lane", "back out"); err != nil {
		t.Fatal(err)
	}
	return s, lock, c.Epoch
}

// Landing is already written, so the trunk may have moved: the lock holder can still say
// landed after the card left MERGING, and the mismatch is recorded.
func TestQueueUpdate_LandedForAnAlreadyLandingRowSurvivesTheCardLeavingMerging(t *testing.T) {
	s, lock, _ := landingRow(t)
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "c1"), QueueLanded, "", ""); err != nil {
		t.Fatalf("landed after the card left MERGING: %v", err)
	}
	if it := rowOf(t, s); it.State != QueueLanded {
		t.Errorf("%+v", it)
	}
	notes, _ := s.NotesSince(bg, 0)
	found := false
	for _, n := range notes {
		found = found || (n.Kind == NoteDecision && n.CardID == "a" && strings.Contains(n.Text, "landed") && strings.Contains(n.Text, "REVIEWING"))
	}
	if !found {
		t.Errorf("no note records that the card was at REVIEWING when its row landed: %+v", notes)
	}
}

// The entry guard stays: verifying and landing are still refused for a card outside MERGING,
// and landed still needs the lock epoch and the candidate.
func TestQueueUpdate_LandedExceptionDoesNotWidenTheGuard(t *testing.T) {
	s, lock, e := landingRow(t)
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock+1, "c1"), QueueLanded, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a wrong lock epoch: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", LockEpoch(lock, "other"), QueueLanded, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("a wrong candidate: %v", err)
	}
	if err := s.QueueUpdate(bg, "a", CardEpoch(e, "c1"), QueueLanded, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("the card epoch for landed: %v", err)
	}
	// a waiting row of a card outside MERGING still cannot enter verifying
	s2 := openTemp(t)
	c := gated(t, s2, "b", "c1")
	if err := s2.Enqueue(bg, "b", c.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	if err := s2.Move(bg, "b", c.Epoch, core.Position{Stage: core.StageReviewing}, "lane", "back out"); err != nil {
		t.Fatal(err)
	}
	if err := s2.QueueUpdate(bg, "b", CardEpoch(c.Epoch, "c1"), QueueVerifying, "", ""); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("verifying for a card outside MERGING: %v", err)
	}
	// and a row that is only verifying cannot jump to landed for such a card
	s3 := openTemp(t)
	c3 := gated(t, s3, "c", "c1")
	if err := s3.Enqueue(bg, "c", c3.Epoch, "c1"); err != nil {
		t.Fatal(err)
	}
	lk, _ := s3.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq"}, LandingLockTTL)
	if err := s3.QueueUpdate(bg, "c", CardEpoch(c3.Epoch, "c1"), QueueVerifying, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := s3.Move(bg, "c", c3.Epoch, core.Position{Stage: core.StageReviewing}, "lane", "back out"); err != nil {
		t.Fatal(err)
	}
	if err := s3.QueueUpdate(bg, "c", LockEpoch(lk, "c1"), QueueLanded, "", ""); err == nil {
		t.Error("verifying -> landed is not an edge, whatever the card's stage")
	}
}

// ReapLocks frees a dead holder's merge lock and nothing else.
func TestReapLocks_ExpiresADeadHoldersLockAndTouchesNothingElse(t *testing.T) {
	s := openTemp(t)
	claimOne(t, s, "a", "internal/a/x.go")
	if _, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq", PID: 4242, PIDStart: "t"}, LandingLockTTL); err != nil {
		t.Fatal(err)
	}
	if err := s.db.QueryRowContext(bg, `SELECT 1`).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(bg, `UPDATE leases SET expires_at=?`, longAgo); err != nil {
		t.Fatal(err)
	}
	notesBefore := rowCount(t, s, "notes")
	if err := s.ReapLocks(bg, alwaysDead); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "next"}, time.Minute); err != nil {
		t.Errorf("the dead holder's lock must be free: %v", err)
	}
	if rowCount(t, s, "leases") != 1 || cardOf(t, s, "a").ReclaimCount != 0 {
		t.Error("a card lease was reclaimed by a lock-only reap")
	}
	if rowCount(t, s, "notes") != notesBefore {
		t.Error("a lock-only reap wrote a note")
	}
	if err := s.ReapLocks(bg, nil); !errors.Is(err, ErrInvalid) {
		t.Errorf("a nil judge: %v", err)
	}
}

// A live or unknown holder keeps its lock.
func TestReapLocks_LeavesALiveOrUnknownHoldersLockAlone(t *testing.T) {
	for name, v := range map[string]Liveness{"alive": LivenessAlive, "unknown": LivenessUnknown} {
		s := openTemp(t)
		if _, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "mq", PID: 4242, PIDStart: "t"}, LandingLockTTL); err != nil {
			t.Fatal(err)
		}
		if err := s.ReapLocks(bg, func(int, string) Liveness { return v }); err != nil {
			t.Fatal(err)
		}
		if _, err := s.AcquireLock(bg, LockMerge, LockHolder{SessionID: "next"}, time.Minute); !errors.Is(err, ErrLockHeld) {
			t.Errorf("%s holder lost its lock: %v", name, err)
		}
	}
}
