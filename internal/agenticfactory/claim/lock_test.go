package claim

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// writeLock plants a lock file as a crashed owner would have left it.
func writeLock(t *testing.T, s *Store, i core.IssueRef, pid int, start string) {
	t.Helper()
	if err := os.WriteFile(s.lock(i), []byte(fmt.Sprintf("%d %s\n", pid, start)), 0o600); err != nil {
		t.Fatal(err)
	}
}

func probeOf(state homestate.ProcessIdentityState, fp string) func(int) (string, homestate.ProcessIdentityState) {
	return func(int) (string, homestate.ProcessIdentityState) { return fp, state }
}

func TestLockRecordsOwnerIdentity(t *testing.T) {
	s := open(t)
	s.probe = probeOf(homestate.ProcessIdentityLive, "fp-self")
	i := ref(t, 1)
	var seen LockInfo
	err := s.withLock(i, func() error {
		var e error
		seen, e = s.readLock(i)
		return e
	})
	if err != nil || seen.PID != os.Getpid() || seen.Start != "fp-self" {
		t.Errorf("lock identity: %+v %v", seen, err)
	}
}

func TestDeadOwnerLockIsCandidateButNeverAutoTaken(t *testing.T) {
	s := open(t)
	s.probe = probeOf(homestate.ProcessIdentityDead, "")
	i := ref(t, 1)
	writeLock(t, s, i, 4242, "fp-old")

	st, err := s.InspectLock(i)
	if err != nil || !st.Held || !st.ReissueCandidate || st.Class != homestate.OwnerDead {
		t.Fatalf("inspect: %+v %v", st, err)
	}
	// Inspection and Acquire do not recover by themselves.
	if _, err := s.Acquire(i, "a", "d1", "run-1"); !errors.Is(err, core.ErrBusy) {
		t.Errorf("acquire must stay Busy until reissued: %v", err)
	}
	if err := s.ReissueLock(i, st.Info); err != nil {
		t.Fatalf("reissue: %v", err)
	}
	if _, err := s.Acquire(i, "a", "d1", "run-1"); err != nil {
		t.Errorf("acquire after reissue: %v", err)
	}
}

func TestLiveOwnerExpiredLockStaysBusy(t *testing.T) {
	s := open(t)
	s.probe = probeOf(homestate.ProcessIdentityLive, "fp-old")
	i := ref(t, 1)
	writeLock(t, s, i, 4242, "fp-old")
	old := time.Now().Add(-72 * time.Hour)
	if err := os.Chtimes(s.lock(i), old, old); err != nil {
		t.Fatal(err)
	}
	st, err := s.InspectLock(i)
	if err != nil || st.ReissueCandidate || st.Class != homestate.OwnerLive {
		t.Fatalf("live owner must not be a candidate: %+v %v", st, err)
	}
	if err := s.ReissueLock(i, st.Info); !errors.Is(err, core.ErrBusy) {
		t.Errorf("reissue of live owner: %v", err)
	}
	if _, err := os.Stat(s.lock(i)); err != nil {
		t.Errorf("live owner's lock was removed: %v", err)
	}
	if _, err := s.Acquire(i, "a", "d1", "run-1"); !errors.Is(err, core.ErrBusy) {
		t.Errorf("acquire: %v", err)
	}
}

func TestReusedPidWithDifferentFingerprintIsDead(t *testing.T) {
	s := open(t)
	s.probe = probeOf(homestate.ProcessIdentityLive, "fp-new-process")
	i := ref(t, 1)
	writeLock(t, s, i, 4242, "fp-old")
	if st, _ := s.InspectLock(i); !st.ReissueCandidate {
		t.Errorf("reused pid must be a candidate: %+v", st)
	}
}

func TestUnreadableOrIndeterminateOwnerStaysBusy(t *testing.T) {
	cases := map[string]func(*Store, core.IssueRef){
		"empty lock": func(s *Store, i core.IssueRef) { _ = os.WriteFile(s.lock(i), nil, 0o600) },
		"garbage":    func(s *Store, i core.IssueRef) { _ = os.WriteFile(s.lock(i), []byte("xyz"), 0o600) },
		"probe cannot tell": func(s *Store, i core.IssueRef) {
			s.probe = probeOf(homestate.ProcessIdentityIndeterminate, "")
			writeLock(t, s, i, 4242, "fp-old")
		},
	}
	for name, plant := range cases {
		t.Run(name, func(t *testing.T) {
			s := open(t)
			i := ref(t, 1)
			plant(s, i)
			st, err := s.InspectLock(i)
			if err != nil || st.ReissueCandidate {
				t.Fatalf("must not be a candidate: %+v %v", st, err)
			}
			if err := s.ReissueLock(i, st.Info); !errors.Is(err, core.ErrBusy) {
				t.Errorf("reissue: %v", err)
			}
		})
	}
}

func TestReissueRefusesWhenLockChangedSinceInspection(t *testing.T) {
	s := open(t)
	s.probe = probeOf(homestate.ProcessIdentityDead, "")
	i := ref(t, 1)
	writeLock(t, s, i, 4242, "fp-old")
	st, _ := s.InspectLock(i)
	// A new owner took the lock between inspection and reissue.
	writeLock(t, s, i, 5555, "fp-new")
	if err := s.ReissueLock(i, st.Info); !errors.Is(err, core.ErrBusy) {
		t.Errorf("changed lock must not be reissued: %v", err)
	}
	if _, err := os.Stat(s.lock(i)); err != nil {
		t.Errorf("lock removed: %v", err)
	}
}

func TestInspectWithoutLock(t *testing.T) {
	s := open(t)
	st, err := s.InspectLock(ref(t, 1))
	if err != nil || st.Held {
		t.Errorf("no lock: %+v %v", st, err)
	}
	if err := s.ReissueLock(ref(t, 1), LockInfo{PID: 1, Start: "x"}); !errors.Is(err, core.ErrNotHeld) {
		t.Errorf("reissue without lock: %v", err)
	}
}
