package claim

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Presence is where the holder of an Issue runs and how long it said it would be
// there: the process id, the process start fingerprint (pids are reused, so the
// pair is the identity) and the end of its lease. It names a process on THIS host,
// so it is kept next to the cache and never goes to the control ref. It belongs to
// one holder and one epoch; a later holder starts with none.
type Presence struct {
	Holder     string    `json:"holder"`
	Epoch      int64     `json:"epoch"`
	PID        int       `json:"pid"`
	Start      string    `json:"start"`
	At         time.Time `json:"at"` // when it was recorded
	LeaseUntil time.Time `json:"lease_until"`
}

// OwnerState is what the recorded presence says about the holder now.
type OwnerState struct {
	// HasPresence is false when the current holder recorded none.
	HasPresence bool
	PID         int
	// Class is the process classification: dead only when the process is gone or
	// its pid belongs to another process now; live only when it is the same
	// process; anything else is indeterminate.
	Class homestate.OwnerClassification
	// Damaged is set when a presence file exists and cannot be read; File names it.
	Damaged bool
	File    string
	// LeaseExpired is information for the monitor. It never decides: a live
	// process whose lease ran out is still live.
	LeaseExpired bool
}

// Refusal is why a takeover that rests on the owner's recorded process must not go
// ahead: ErrHeld when the process is still running, ErrOwnerUnconfirmed when the
// presence file cannot be read or the process cannot be told running or gone. It is
// nil when no presence is recorded or the recorded process is gone. The runner's first
// look and the store's check under the Issue lock share it, so both refuse alike.
func (o OwnerState) Refusal() error {
	switch {
	case o.HasPresence && o.Class == homestate.OwnerLive:
		return fmt.Errorf("%w: the owner's process %d is still running", core.ErrHeld, o.PID)
	case o.HasPresence && o.Damaged:
		return fmt.Errorf("%w: the presence file %s cannot be read: once the owner's process is confirmed gone, remove that file (the owner's next Heartbeat writes a new one)", core.ErrOwnerUnconfirmed, o.File)
	case o.HasPresence && o.Class == homestate.OwnerIndeterminate:
		return fmt.Errorf("%w: the owner's process %d cannot be told running or gone", core.ErrOwnerUnconfirmed, o.PID)
	}
	return nil
}

func (s *Store) presencePath(i core.IssueRef) string {
	return filepath.Join(s.dir, i.Resource()+".presence")
}

// PresenceFile is where the presence of the Issue's holder is recorded.
func (s *Store) PresenceFile(i core.IssueRef) string { return s.presencePath(i) }

func (s *Store) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

// SetProcessProbe replaces how a recorded process is looked at; the default is the
// platform probe. Wiring and tests use it.
func (s *Store) SetProcessProbe(p func(int) (string, homestate.ProcessIdentityState)) { s.probe = p }

// Heartbeat records, for the current holder and epoch, the process it runs in and a
// lease of the given length from now. Calling it again renews the lease.
func (s *Store) Heartbeat(i core.IssueRef, holder string, epoch int64, pid int, start string, lease time.Duration) error {
	if pid < 1 || strings.TrimSpace(start) == "" || lease <= 0 {
		return fmt.Errorf("%w: a heartbeat needs the process id, its start fingerprint and a lease", core.ErrEvidenceIncomplete)
	}
	return s.withLock(i, func() error {
		if _, err := s.Verify(i, holder, epoch); err != nil {
			return err
		}
		b, err := json.Marshal(Presence{Holder: holder, Epoch: epoch, PID: pid, Start: start, At: s.clock(), LeaseUntil: s.clock().Add(lease)})
		if err != nil {
			return err
		}
		return s.writeAtomic(s.dir, i.Resource()+".presence.tmp-*", s.presencePath(i), b)
	})
}

// presenceState is what readPresence found.
type presenceState int

const (
	presenceAbsent  presenceState = iota // no file
	presenceDamaged                      // a file that cannot be read or is incomplete
	presenceOK
)

func (s *Store) readPresence(i core.IssueRef) (Presence, presenceState, error) {
	b, err := os.ReadFile(s.presencePath(i))
	if errors.Is(err, os.ErrNotExist) {
		return Presence{}, presenceAbsent, nil
	}
	if err != nil {
		return Presence{}, presenceAbsent, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	var p Presence
	if json.Unmarshal(b, &p) != nil || p.PID < 1 || strings.TrimSpace(p.Start) == "" {
		return p, presenceDamaged, nil
	}
	return p, presenceOK, nil
}

// recoverHint is the one line an operator needs when a presence file is damaged.
func (s *Store) recoverHint(i core.IssueRef) string {
	return fmt.Sprintf("the presence file %s cannot be read: once the owner's process is confirmed gone, remove that file (the owner's next Heartbeat writes a new one)", s.presencePath(i))
}

// OwnerLiveness reads the presence the current holder recorded and classifies its
// process. A presence of an earlier holder or epoch is not the current holder's.
func (s *Store) OwnerLiveness(i core.IssueRef) (OwnerState, error) {
	rec, ok, err := s.read(i)
	if err != nil {
		return OwnerState{}, err
	}
	if !ok {
		return OwnerState{}, core.ErrNotHeld
	}
	p, st, err := s.readPresence(i)
	if err != nil {
		return OwnerState{}, err
	}
	switch st {
	case presenceAbsent:
		return OwnerState{}, nil
	case presenceDamaged:
		// Something was recorded and cannot be read: that is not "nothing recorded",
		// which would let a takeover go ahead on the caller's word. It cannot be told
		// running from gone.
		return OwnerState{HasPresence: true, PID: p.PID, Class: homestate.OwnerIndeterminate, Damaged: true, File: s.presencePath(i)}, nil
	}
	if p.Holder != rec.Holder || p.Epoch != rec.Epoch {
		return OwnerState{}, nil // an earlier holder's record; the current holder has recorded none
	}
	return OwnerState{
		HasPresence:  true,
		PID:          p.PID,
		Class:        homestate.ClassifyOwnerWith(s.probe, p.PID, p.Start),
		LeaseExpired: !s.clock().Before(p.LeaseUntil),
	}, nil
}
