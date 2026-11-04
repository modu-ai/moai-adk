package claim

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// LockInfo is the owner identity written into a lock file: the process id and
// its start fingerprint (pids are reused, so the pair is the identity).
type LockInfo struct {
	PID   int
	Start string
}

// LockStatus is what InspectLock observed.
type LockStatus struct {
	Held bool
	Info LockInfo
	// Class is the owner classification; empty when the lock carries no
	// readable identity.
	Class homestate.OwnerClassification
	// ReissueCandidate is true only when Class is OwnerDead. Live,
	// indeterminate and unreadable owners are never candidates, however old
	// the lock file is.
	ReissueCandidate bool
}

// GlobalLock names one of the store-wide locks, which guard the shared
// directories rather than one Issue.
type GlobalLock string

const (
	LockDocuments GlobalLock = "docs"     // the keyed documents
	LockRequests  GlobalLock = "requests" // the request queue
	LockAgents    GlobalLock = "agents"   // the agent registry
)

func (s *Store) globalLockPath(g GlobalLock) (string, error) {
	switch g {
	case LockDocuments, LockRequests, LockAgents:
		return filepath.Join(s.dir, string(g)+".lock"), nil
	}
	return "", fmt.Errorf("%w: global lock %q", core.ErrIDInvalid, string(g))
}

// parseLock reads the "<pid> <start>" a lock file holds; anything else is an
// unreadable identity (the zero LockInfo).
func parseLock(b []byte) LockInfo {
	f := strings.Fields(string(b))
	if len(f) != 2 {
		return LockInfo{}
	}
	pid, err := strconv.Atoi(f[0])
	if err != nil {
		return LockInfo{}
	}
	return LockInfo{PID: pid, Start: f[1]}
}

func readLockFile(path string) (LockInfo, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return LockInfo{}, core.ErrNotHeld
	}
	if err != nil {
		return LockInfo{}, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return parseLock(b), nil
}

func (s *Store) readLock(i core.IssueRef) (LockInfo, error) { return readLockFile(s.lock(i)) }

// inspectPath reports who holds the lock file at path and whether the holder is
// provably dead. It changes nothing.
func (s *Store) inspectPath(path string) (LockStatus, error) {
	info, err := readLockFile(path)
	if errors.Is(err, core.ErrNotHeld) {
		return LockStatus{}, nil
	}
	if err != nil {
		return LockStatus{}, err
	}
	st := LockStatus{Held: true, Info: info}
	if info.PID == 0 {
		return st, nil
	}
	st.Class = homestate.ClassifyOwnerWith(s.probe, info.PID, info.Start)
	st.ReissueCandidate = st.Class == homestate.OwnerDead
	return st, nil
}

// reissuePath removes the lock a provably dead owner left behind at path. It
// re-reads and re-classifies, and refuses (ErrBusy) unless the lock still carries
// exactly the identity that was inspected and that owner is still dead.
//
// The removal is a rename to a private name followed by a second check of what
// was actually moved: a plain remove after the check would delete a lock that
// another process reissued and a live one took in between. If the moved lock is
// not the inspected dead one it is put back with a link, which fails rather than
// overwrite a lock created since; in that case the displaced holder has lost its
// lock and the error says so.
func (s *Store) reissuePath(path, name string, inspected LockInfo) error {
	cur, err := s.inspectPath(path)
	if err != nil {
		return err
	}
	if !cur.Held {
		return core.ErrNotHeld
	}
	if cur.Info != inspected || !cur.ReissueCandidate {
		return fmt.Errorf("%w: %s lock owner is not a confirmed-dead match", core.ErrBusy, name)
	}
	if s.beforeReissueRemove != nil {
		s.beforeReissueRemove()
	}
	tomb := fmt.Sprintf("%s.reissue-%d", path, os.Getpid())
	if err := os.Rename(path, tomb); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return core.ErrNotHeld
		}
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	moved, rerr := readLockFile(tomb)
	if rerr == nil && moved == inspected &&
		homestate.ClassifyOwnerWith(s.probe, moved.PID, moved.Start) == homestate.OwnerDead {
		if err := os.Remove(tomb); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
		}
		return nil
	}
	if err := os.Link(tomb, path); err != nil {
		_ = os.Remove(tomb)
		return fmt.Errorf("%w: %s lock changed during the reissue and could not be put back: %v", core.ErrBusy, name, err)
	}
	_ = os.Remove(tomb)
	return fmt.Errorf("%w: %s lock owner changed during the reissue", core.ErrBusy, name)
}

// InspectLock reports who holds the Issue's lock and whether the holder is
// provably dead. It changes nothing.
func (s *Store) InspectLock(i core.IssueRef) (LockStatus, error) { return s.inspectPath(s.lock(i)) }

// ReissueLock removes the lock a provably dead owner left behind. It never takes
// over a live, unknown or changed lock.
func (s *Store) ReissueLock(i core.IssueRef, inspected LockInfo) error {
	return s.reissuePath(s.lock(i), i.Resource(), inspected)
}

// InspectGlobalLock is InspectLock for a store-wide lock.
func (s *Store) InspectGlobalLock(g GlobalLock) (LockStatus, error) {
	p, err := s.globalLockPath(g)
	if err != nil {
		return LockStatus{}, err
	}
	return s.inspectPath(p)
}

// ReissueGlobalLock is ReissueLock for a store-wide lock: a crashed holder of the
// documents, requests or agent-registry lock would otherwise block every caller
// for good.
func (s *Store) ReissueGlobalLock(g GlobalLock, inspected LockInfo) error {
	p, err := s.globalLockPath(g)
	if err != nil {
		return err
	}
	return s.reissuePath(p, string(g), inspected)
}

// noteBusy counts a wait that ended in Busy for the lock file at path and returns
// the number of Busy answers in a row; noteTaken ends the run.
func (s *Store) noteBusy(path string) int {
	s.busyMu.Lock()
	defer s.busyMu.Unlock()
	if s.busy == nil {
		s.busy = map[string]int{}
	}
	k := filepath.Base(path)
	s.busy[k]++
	return s.busy[k]
}

func (s *Store) noteTaken(path string) {
	s.busyMu.Lock()
	defer s.busyMu.Unlock()
	delete(s.busy, filepath.Base(path))
}

func (s *Store) busyStreak(path string) int {
	s.busyMu.Lock()
	defer s.busyMu.Unlock()
	return s.busy[filepath.Base(path)]
}

// BusyStreak is how many waits for the store-wide lock g ended in Busy in a row.
// A caller that sees it grow can inspect the lock (InspectGlobalLock) and reissue
// it when its holder is provably dead.
func (s *Store) BusyStreak(g GlobalLock) int {
	p, err := s.globalLockPath(g)
	if err != nil {
		return 0
	}
	return s.busyStreak(p)
}

// BusyStreakIssue is BusyStreak for an Issue's lock.
func (s *Store) BusyStreakIssue(i core.IssueRef) int { return s.busyStreak(s.lock(i)) }
