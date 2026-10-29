package claim

import (
	"errors"
	"fmt"
	"os"
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

func (s *Store) readLock(i core.IssueRef) (LockInfo, error) {
	b, err := os.ReadFile(s.lock(i))
	if errors.Is(err, os.ErrNotExist) {
		return LockInfo{}, core.ErrNotHeld
	}
	if err != nil {
		return LockInfo{}, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	f := strings.Fields(string(b))
	if len(f) != 2 {
		return LockInfo{}, nil // present but identity unreadable
	}
	pid, err := strconv.Atoi(f[0])
	if err != nil {
		return LockInfo{}, nil
	}
	return LockInfo{PID: pid, Start: f[1]}, nil
}

// InspectLock reports who holds the Issue's lock and whether the holder is
// provably dead. It changes nothing.
func (s *Store) InspectLock(i core.IssueRef) (LockStatus, error) {
	info, err := s.readLock(i)
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

// ReissueLock removes the lock a provably dead owner left behind. It re-reads
// and re-classifies at the moment of removal and refuses (ErrBusy) unless the
// lock still carries exactly the identity that was inspected and that owner is
// still dead. It never takes over a live, unknown or changed lock.
func (s *Store) ReissueLock(i core.IssueRef, inspected LockInfo) error {
	cur, err := s.InspectLock(i)
	if err != nil {
		return err
	}
	if !cur.Held {
		return core.ErrNotHeld
	}
	if cur.Info != inspected || !cur.ReissueCandidate {
		return fmt.Errorf("%w: %s lock owner is not a confirmed-dead match", core.ErrBusy, i.Resource())
	}
	if err := os.Remove(s.lock(i)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return nil
}
