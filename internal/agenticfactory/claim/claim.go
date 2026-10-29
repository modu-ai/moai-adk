// Package claim stores one ownership record per Issue and refuses every write
// whose holder and epoch do not match it (Issue #1948, design §2.2, §4).
//
// Differences from internal/factory slot leases, on purpose: a claim is keyed
// by Issue, never taken over because it looks expired (a live owner would
// become a second writer), and every refusal names why the write was refused.
package claim

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Record is the authoritative local claim for one Issue.
type Record struct {
	Issue    int              `json:"issue"`
	Holder   string           `json:"holder"` // empty once released
	Dispatch string           `json:"dispatch"`
	Run      core.RunID       `json:"run"`
	Attempt  core.AttemptID   `json:"attempt,omitempty"`
	Epoch    int64            `json:"epoch"`
	Seq      int64            `json:"seq"`
	State    core.State       `json:"state"`
	Return   core.State       `json:"return_stage,omitempty"`
	Reason   core.BlockReason `json:"block_reason,omitempty"`
	Resume   string           `json:"resume_condition,omitempty"`
	Head     core.HeadSHA     `json:"head,omitempty"`
	Delivery core.Delivery    `json:"delivery,omitempty"`
}

// Store keeps records under one directory: issue-<n>.json, guarded by
// issue-<n>.lock created with O_EXCL.
type Store struct{ dir string }

const (
	lockTries = 5
	lockWait  = 10 * time.Millisecond
)

// Open prepares dir (created if absent).
func Open(dir string) (*Store, error) {
	if dir == "" {
		return nil, errors.New("agenticfactory/claim: empty store directory")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return &Store{dir: dir}, nil
}

func (s *Store) path(i core.IssueRef) string { return filepath.Join(s.dir, i.Resource()+".json") }
func (s *Store) lock(i core.IssueRef) string { return filepath.Join(s.dir, i.Resource()+".lock") }

// withLock runs fn while holding the Issue's lock. A held lock is Busy
// (retry), never Held (another owner). A lock left by a crashed process is not
// stolen; the operator removes it.
func (s *Store) withLock(i core.IssueRef, fn func() error) error {
	var f *os.File
	var err error
	for n := 0; n < lockTries; n++ {
		f, err = os.OpenFile(s.lock(i), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
		}
		time.Sleep(lockWait)
	}
	if err != nil {
		return fmt.Errorf("%w: %s", core.ErrBusy, i.Resource())
	}
	_ = f.Close()
	defer func() { _ = os.Remove(s.lock(i)) }()
	return fn()
}

// read returns (zero, false, nil) when no record file exists.
func (s *Store) read(i core.IssueRef) (Record, bool, error) {
	b, err := os.ReadFile(s.path(i))
	if errors.Is(err, os.ErrNotExist) {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	var r Record
	if err := json.Unmarshal(b, &r); err != nil {
		return Record{}, false, fmt.Errorf("%w: corrupt record: %v", core.ErrStoreUnavailable, err)
	}
	return r, true, nil
}

func (s *Store) write(i core.IssueRef, r Record) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.dir, i.Resource()+".tmp-*")
	if err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if err := os.Rename(tmp.Name(), s.path(i)); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return nil
}

// Load returns the live claim, or ErrNotHeld when none (or released).
func (s *Store) Load(i core.IssueRef) (Record, error) {
	r, ok, err := s.read(i)
	if err != nil {
		return Record{}, err
	}
	if !ok || r.Holder == "" {
		return Record{}, core.ErrNotHeld
	}
	return r, nil
}

// Acquire admits exactly one holder per Issue inside one critical section.
// The same dispatch id arriving again is a duplicate; a different holder or a
// new dispatch on an owned Issue is Held.
func (s *Store) Acquire(i core.IssueRef, holder, dispatch string, run core.RunID) (Record, error) {
	if holder == "" || dispatch == "" {
		return Record{}, fmt.Errorf("%w: holder and dispatch are required", core.ErrIDInvalid)
	}
	var out Record
	err := s.withLock(i, func() error {
		cur, ok, err := s.read(i)
		if err != nil {
			return err
		}
		if ok && cur.Holder != "" {
			if cur.Dispatch == dispatch {
				return fmt.Errorf("%w: %s dispatch %q", core.ErrDuplicateDispatch, i.Resource(), dispatch)
			}
			return fmt.Errorf("%w: %s held by %q", core.ErrHeld, i.Resource(), cur.Holder)
		}
		out = Record{
			Issue: i.Number, Holder: holder, Dispatch: dispatch, Run: run,
			Epoch: cur.Epoch + 1, Seq: cur.Seq + 1, State: core.StateClaimed,
		}
		return s.write(i, out)
	})
	if err != nil {
		return Record{}, err
	}
	return out, nil
}

// Verify confirms holder and epoch own the live claim. Every failure wraps
// ErrOwnershipUnverified together with the specific cause.
func (s *Store) Verify(i core.IssueRef, holder string, epoch int64) (Record, error) {
	r, err := s.Load(i)
	if err != nil {
		return Record{}, fmt.Errorf("%w: %w", core.ErrOwnershipUnverified, err)
	}
	if r.Holder != holder {
		return Record{}, fmt.Errorf("%w: %w: holder %q", core.ErrOwnershipUnverified, core.ErrForeign, holder)
	}
	if r.Epoch != epoch {
		return Record{}, fmt.Errorf("%w: epoch %d, record at %d", core.ErrOwnershipUnverified, epoch, r.Epoch)
	}
	return r, nil
}

// Update verifies ownership, applies fn to a copy, and persists it with Seq+1.
// A refused or failed update leaves the stored record untouched.
func (s *Store) Update(i core.IssueRef, holder string, epoch int64, fn func(*Record) error) (Record, error) {
	var out Record
	err := s.withLock(i, func() error {
		cur, err := s.Verify(i, holder, epoch)
		if err != nil {
			return err
		}
		next := cur
		if err := fn(&next); err != nil {
			return err
		}
		next.Issue, next.Holder, next.Epoch, next.Seq = cur.Issue, cur.Holder, cur.Epoch, cur.Seq+1
		out = next
		return s.write(i, out)
	})
	if err != nil {
		return Record{}, err
	}
	return out, nil
}

// Release hands the claim back. The record stays as a tombstone so the epoch
// keeps increasing across holders.
func (s *Store) Release(i core.IssueRef, holder string) error {
	return s.withLock(i, func() error {
		cur, ok, err := s.read(i)
		if err != nil {
			return err
		}
		if !ok || cur.Holder == "" {
			return core.ErrNotHeld
		}
		if cur.Holder != holder {
			return fmt.Errorf("%w: %q is not the holder", core.ErrForeign, holder)
		}
		cur.Holder, cur.Seq = "", cur.Seq+1
		return s.write(i, cur)
	})
}
