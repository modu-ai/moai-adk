package ledger

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func holder(n string) LockHolder { return LockHolder{SessionID: n, PID: 4000, PIDStart: "start-" + n} }

func TestALockHasOneLiveHolderAndItsEpochOnlyGoesUp(t *testing.T) {
	s := openTemp(t)
	e1, err := s.AcquireLock(bg, LockMerge, holder("a"), time.Minute)
	if err != nil || e1 != 1 {
		t.Fatalf("first acquire: %d %v", e1, err)
	}
	if _, err := s.AcquireLock(bg, LockMerge, holder("b"), time.Minute); !errors.Is(err, ErrLockHeld) {
		t.Errorf("another live holder: %v", err)
	}
	if _, err := s.AcquireLock(bg, LockMerge, holder("a"), time.Minute); !errors.Is(err, ErrLockHeld) {
		t.Errorf("the holder asking again while live: %v", err)
	}
	if err := s.ReleaseLock(bg, LockMerge, e1+1); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("release with another epoch: %v", err)
	}
	if err := s.ReleaseLock(bg, LockMerge, e1); err != nil {
		t.Fatal(err)
	}
	if err := s.ReleaseLock(bg, LockMerge, e1); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("released twice: %v", err)
	}
	e2, err := s.AcquireLock(bg, LockMerge, holder("b"), time.Minute)
	if err != nil || e2 != e1+1 {
		t.Errorf("after a release the epoch is one higher: %d %v", e2, err)
	}
}

func TestALockWhoseTTLRanOutIsTakenOverAtAHigherEpoch(t *testing.T) {
	s := openTemp(t)
	e1, _ := s.AcquireLock(bg, "x", holder("a"), time.Minute)
	if _, err := s.db.ExecContext(bg, `UPDATE locks SET expires_at='2000-01-01T00:00:00.000Z'`); err != nil {
		t.Fatal(err)
	}
	e2, err := s.AcquireLock(bg, "x", holder("b"), time.Minute)
	if err != nil || e2 != e1+1 {
		t.Fatalf("takeover: %d %v", e2, err)
	}
	if err := s.ReleaseLock(bg, "x", e1); !errors.Is(err, ErrStaleEpoch) {
		t.Errorf("the old holder is fenced: %v", err)
	}
}

func TestLockArgumentsAreChecked(t *testing.T) {
	s := openTemp(t)
	for name, f := range map[string]func() error{
		"no name":   func() error { _, err := s.AcquireLock(bg, "", holder("a"), time.Minute); return err },
		"no ttl":    func() error { _, err := s.AcquireLock(bg, "x", holder("a"), 0); return err },
		"no holder": func() error { _, err := s.AcquireLock(bg, "x", LockHolder{}, time.Minute); return err },
	} {
		if err := f(); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// Reap expires the locks of holders judged Dead (the row stays, so the epoch goes on
// counting) and leaves the others.
func TestReapExpiresTheLocksOfDeadHoldersOnly(t *testing.T) {
	s := openTemp(t)
	live := LockHolder{SessionID: "live", PID: 1, PIDStart: "l"}
	dead := LockHolder{SessionID: "dead", PID: 2, PIDStart: "d"}
	eLive, _ := s.AcquireLock(bg, "live-lock", live, time.Hour)
	eDead, _ := s.AcquireLock(bg, "dead-lock", dead, time.Hour)
	if _, err := s.Reap(bg, func(pid int, _ string) Liveness {
		if pid == 2 {
			return LivenessDead
		}
		return LivenessAlive
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AcquireLock(bg, "live-lock", holder("x"), time.Minute); !errors.Is(err, ErrLockHeld) {
		t.Errorf("a live holder keeps its lock: %v", err)
	}
	e, err := s.AcquireLock(bg, "dead-lock", holder("x"), time.Minute)
	if err != nil || e != eDead+1 {
		t.Errorf("a dead holder's lock is taken over at epoch %d: %d %v", eDead+1, e, err)
	}
	_ = eLive
}

func TestOnlyOneOfManyAcquiresWins(t *testing.T) {
	s := openTemp(t)
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := s.AcquireLock(bg, "race", holder(string(rune('a'+i))), time.Minute); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if wins != 1 {
		t.Errorf("%d holders", wins)
	}
}
