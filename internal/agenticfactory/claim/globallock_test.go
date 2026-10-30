package claim

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// quickWait makes a held lock answer Busy at once instead of after the real budget.
func quickWait(s *Store) {
	s.lockPol = lockPolicy{tries: 2, wait: time.Millisecond, sleep: func(time.Duration) {}}
}

// plantGlobal writes a store-wide lock file as a crashed owner would have left it.
func plantGlobal(t *testing.T, s *Store, g GlobalLock, pid int, start string) {
	t.Helper()
	p, err := s.globalLockPath(g)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(fmt.Sprintf("%d %s\n", pid, start)), 0o600); err != nil {
		t.Fatal(err)
	}
}

// useGlobal runs the operation that takes the named store-wide lock.
func useGlobal(s *Store, g GlobalLock) error {
	switch g {
	case LockDocuments:
		return s.withDocs(func() error { return nil })
	case LockRequests:
		return s.withRequests(func() error { return nil })
	default:
		return s.withRegistryLock(func() error { return nil })
	}
}

var allGlobals = []GlobalLock{LockDocuments, LockRequests, LockAgents}

func TestADeadOwnersGlobalLockIsInspectedAndReissued(t *testing.T) {
	for _, g := range allGlobals {
		t.Run(string(g), func(t *testing.T) {
			s := open(t)
			quickWait(s)
			s.probe = probeOf(homestate.ProcessIdentityDead, "")
			plantGlobal(t, s, g, 4242, "fp-old")

			if err := useGlobal(s, g); !errors.Is(err, core.ErrBusy) {
				t.Fatalf("a stale lock must answer Busy until reissued: %v", err)
			}
			st, err := s.InspectGlobalLock(g)
			if err != nil || !st.Held || !st.ReissueCandidate || st.Info.PID != 4242 || st.Class != homestate.OwnerDead {
				t.Fatalf("inspect: %+v %v", st, err)
			}
			if err := s.ReissueGlobalLock(g, st.Info); err != nil {
				t.Fatalf("reissue: %v", err)
			}
			if err := useGlobal(s, g); err != nil {
				t.Errorf("the lock must be usable after the reissue: %v", err)
			}
		})
	}
}

func TestALiveOwnersGlobalLockIsNeverReissued(t *testing.T) {
	for _, g := range allGlobals {
		t.Run(string(g), func(t *testing.T) {
			s := open(t)
			quickWait(s)
			s.probe = probeOf(homestate.ProcessIdentityLive, "fp-old")
			plantGlobal(t, s, g, 4242, "fp-old")
			old := time.Now().Add(-72 * time.Hour)
			p, _ := s.globalLockPath(g)
			if err := os.Chtimes(p, old, old); err != nil {
				t.Fatal(err)
			}
			st, err := s.InspectGlobalLock(g)
			if err != nil || st.ReissueCandidate || st.Class != homestate.OwnerLive {
				t.Fatalf("a live owner must not be a candidate however old the file: %+v %v", st, err)
			}
			touched := false
			s.beforeReissueRemove = func() { touched = true }
			if err := s.ReissueGlobalLock(g, st.Info); !errors.Is(err, core.ErrBusy) {
				t.Errorf("reissue of a live owner: %v", err)
			}
			if touched {
				t.Error("a live owner's lock must be refused before anything is moved")
			}
			if _, err := os.Stat(p); err != nil {
				t.Errorf("the live owner's lock was removed: %v", err)
			}
		})
	}
}

func TestAnUnknownGlobalLockIsRefused(t *testing.T) {
	s := open(t)
	if _, err := s.InspectGlobalLock("../escape"); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("inspect: %v", err)
	}
	if err := s.ReissueGlobalLock("nope", LockInfo{PID: 1, Start: "x"}); !errors.Is(err, core.ErrIDInvalid) {
		t.Errorf("reissue: %v", err)
	}
}

func TestConsecutiveBusyAnswersAreCountedAndReported(t *testing.T) {
	s := open(t)
	quickWait(s)
	s.probe = probeOf(homestate.ProcessIdentityLive, "fp")
	plantGlobal(t, s, LockDocuments, 4242, "fp")
	i := ref(t, 1)
	writeLock(t, s, i, 4242, "fp")

	if n := s.BusyStreak(LockDocuments); n != 0 {
		t.Fatalf("no wait has failed yet: %d", n)
	}
	err := useGlobal(s, LockDocuments)
	if !errors.Is(err, core.ErrBusy) || strings.Contains(err.Error(), "consecutive") {
		t.Errorf("the first Busy is plain: %v", err)
	}
	err = useGlobal(s, LockDocuments)
	if !errors.Is(err, core.ErrBusy) || !strings.Contains(err.Error(), "2 consecutive") {
		t.Errorf("the second Busy must say it is the second in a row: %v", err)
	}
	if n := s.BusyStreak(LockDocuments); n != 2 {
		t.Errorf("streak = %d, want 2", n)
	}
	if n := s.BusyStreakIssue(i); n != 0 {
		t.Errorf("the streak is per lock, the Issue lock has none: %d", n)
	}
	if err := s.withLock(i, func() error { return nil }); !errors.Is(err, core.ErrBusy) {
		t.Fatal(err)
	}
	if n := s.BusyStreakIssue(i); n != 1 {
		t.Errorf("Issue streak = %d, want 1", n)
	}

	p, _ := s.globalLockPath(LockDocuments)
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := useGlobal(s, LockDocuments); err != nil {
		t.Fatal(err)
	}
	if n := s.BusyStreak(LockDocuments); n != 0 {
		t.Errorf("a successful take must reset the streak: %d", n)
	}
}

// splitProbe tells a dead pid from a live one that has a different fingerprint.
func splitProbe(dead int) func(int) (string, homestate.ProcessIdentityState) {
	return func(pid int) (string, homestate.ProcessIdentityState) {
		if pid == dead {
			return "", homestate.ProcessIdentityDead
		}
		return "fp-live", homestate.ProcessIdentityLive
	}
}

// F13: between the inspection and the removal another process can reissue the
// lock and a live one take it. The reissue must not remove that lock.
func TestAReissueNeverRemovesALockTakenAfterTheInspection(t *testing.T) {
	cases := map[string]struct {
		path    func(*Store) string
		inspect func(*Store) (LockStatus, error)
		reissue func(*Store, LockInfo) error
	}{
		"issue": {
			path:    func(s *Store) string { return s.lock(ref(t, 1)) },
			inspect: func(s *Store) (LockStatus, error) { return s.InspectLock(ref(t, 1)) },
			reissue: func(s *Store, in LockInfo) error { return s.ReissueLock(ref(t, 1), in) },
		},
		"documents": {
			path:    func(s *Store) string { p, _ := s.globalLockPath(LockDocuments); return p },
			inspect: func(s *Store) (LockStatus, error) { return s.InspectGlobalLock(LockDocuments) },
			reissue: func(s *Store, in LockInfo) error { return s.ReissueGlobalLock(LockDocuments, in) },
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s := open(t)
			s.probe = splitProbe(4242)
			p := c.path(s)
			if err := os.WriteFile(p, []byte("4242 fp-old\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			st, err := c.inspect(s)
			if err != nil || !st.ReissueCandidate {
				t.Fatalf("inspect: %+v %v", st, err)
			}
			// the stale lock is reissued by someone else and a live process takes it
			s.beforeReissueRemove = func() {
				if err := os.WriteFile(p, []byte("777 fp-live\n"), 0o600); err != nil {
					t.Error(err)
				}
			}
			if err := c.reissue(s, st.Info); !errors.Is(err, core.ErrBusy) {
				t.Errorf("reissue after the owner changed: %v", err)
			}
			b, err := os.ReadFile(p)
			if err != nil || string(b) != "777 fp-live\n" {
				t.Errorf("the live owner's lock must be intact: %q %v", b, err)
			}
			left, _ := filepath.Glob(p + ".reissue-*")
			if len(left) != 0 {
				t.Errorf("a reissue left a temporary file behind: %v", left)
			}
		})
	}
}
