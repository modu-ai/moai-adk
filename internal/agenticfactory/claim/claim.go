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
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Record is the authoritative local claim for one Issue.
type Record struct {
	Issue  int    `json:"issue"`
	Holder string `json:"holder"` // empty once released
	// AgentName is the display name (agent-N) of the holder; Holder carries
	// the immutable session id. Neither is changed by Update.
	AgentName string           `json:"agent_name,omitempty"`
	Dispatch  string           `json:"dispatch"`
	Run       core.RunID       `json:"run"`
	Attempt   core.AttemptID   `json:"attempt,omitempty"`
	Epoch     int64            `json:"epoch"`
	Seq       int64            `json:"seq"`
	State     core.State       `json:"state"`
	Return    core.State       `json:"return_stage,omitempty"`
	Reason    core.BlockReason `json:"block_reason,omitempty"`
	Resume    string           `json:"resume_condition,omitempty"`
	Head      core.HeadSHA     `json:"head,omitempty"`
	Delivery  core.Delivery    `json:"delivery,omitempty"`
	Deploy    string           `json:"deploy,omitempty"`
	Mirror    string           `json:"mirror_state,omitempty"` // current | stale | blocked(<reason>)
	// AttemptSeq counts attempts on this Issue (initial plus repairs).
	AttemptSeq int `json:"attempt_seq,omitempty"`
	// Receipt is set when the agent published a PR and handed the Issue off.
	Receipt *Receipt `json:"receipt,omitempty"`
	// Closed is set with done: closing the Issue is part of finishing it.
	Closed bool `json:"closed,omitempty"`
	// Rev is the revision of the authoritative copy this cache entry mirrors.
	Rev int64 `json:"rev,omitempty"`
	// Merged is the preserved fact that the PR reached the remote; it is
	// never reverted. Pending names the obligation (local-sync, deploy,
	// issue-close) that still keeps a merged Issue out of done.
	Merged  bool   `json:"merged,omitempty"`
	Pending string `json:"pending_effect,omitempty"`
	// History is the append-only trail of claims, handoffs and reopenings
	// under this one Issue key.
	History []Event `json:"history,omitempty"`
}

// Event is one entry of an Issue's history.
type Event struct {
	Kind    string `json:"kind"` // claim | handoff | reopen | repair-claim
	Agent   string `json:"agent,omitempty"`
	Session string `json:"session,omitempty"`
	Attempt string `json:"attempt,omitempty"`
	Head    string `json:"head,omitempty"`
	PR      int    `json:"pr,omitempty"`
	Note    string `json:"note,omitempty"`
}

// AddEvent appends to the history; entries are never removed or rewritten.
func (r *Record) AddEvent(e Event) { r.History = append(r.History, e) }

// Receipt is the PR handoff record. It outlives the agent's claim: the
// branch and worktree it names are preserved, and only a repair claim on that
// branch may write to the Issue again.
type Receipt struct {
	PR         int            `json:"pr"`
	Head       core.HeadSHA   `json:"head"`
	Branch     string         `json:"branch"`
	Worktree   string         `json:"worktree"`
	Attempt    core.AttemptID `json:"attempt,omitempty"`
	AttemptSeq int            `json:"attempt_seq,omitempty"`
	Agent      string         `json:"agent,omitempty"`   // display name of the agent that published the PR
	Session    string         `json:"session,omitempty"` // its session id; stays unambiguous after the number is reused
}

// Store keeps records under one directory: issue-<n>.json, guarded by
// issue-<n>.lock created with O_EXCL.
type Store struct {
	dir   string
	probe func(int) (string, homestate.ProcessIdentityState)
	// maxOpenPRs holds new claims while this many handed-off Issues are not
	// done. It has no default: below 1 (unset) new claims are refused, and a
	// value below 1 is rejected, never clamped. Repair claims are not held by it.
	maxOpenPRs int
	// ref is the authoritative store when attached; the files under dir are
	// then a cache that Rebuild can recreate.
	ref ControlRef
	// budget is told about every authority write; see core.Budget.
	budget core.Budget
}

const (
	lockTries = 40
	lockWait  = 10 * time.Millisecond
)

// OpenWithLimit is Open with the open-PR backlog limit configured. The limit
// must be at least 1; there is no default and no upper bound is invented.
func OpenWithLimit(dir string, maxOpenPRs int) (*Store, error) {
	if maxOpenPRs < 1 {
		return nil, fmt.Errorf("%w: %d", core.ErrLimitInvalid, maxOpenPRs)
	}
	s, err := Open(dir)
	if err != nil {
		return nil, err
	}
	s.maxOpenPRs = maxOpenPRs
	return s, nil
}

// SetMaxOpenPRs sets the backlog limit. A value below 1 is rejected and the
// previous value stays.
func (s *Store) SetMaxOpenPRs(n int) error {
	if n < 1 {
		return fmt.Errorf("%w: %d", core.ErrLimitInvalid, n)
	}
	s.maxOpenPRs = n
	return nil
}

// MaxOpenPRs returns the configured limit, 0 when none was set.
func (s *Store) MaxOpenPRs() int { return s.maxOpenPRs }

// Open prepares dir (created if absent). The store it returns has no backlog
// limit, so it reads and repairs but refuses new claims until one is set; a
// monitor opens it this way.
func Open(dir string) (*Store, error) {
	if dir == "" {
		return nil, errors.New("agenticfactory/claim: empty store directory")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return &Store{dir: dir, probe: homestate.ProbeProcessIdentity}, nil
}

func (s *Store) path(i core.IssueRef) string { return filepath.Join(s.dir, i.Resource()+".json") }
func (s *Store) lock(i core.IssueRef) string { return filepath.Join(s.dir, i.Resource()+".lock") }

// withLock runs fn while holding the Issue's lock. A held lock is Busy
// (retry), never Held (another owner). A lock left by a crashed process is not
// stolen; the operator removes it.
func (s *Store) withLock(i core.IssueRef, fn func() error) error {
	return s.withLockFile(s.lock(i), i.Resource(), fn)
}

// withLockFile is withLock for any lock path; name appears in a Busy error.
func (s *Store) withLockFile(path, name string, fn func() error) error {
	var f *os.File
	var err error
	for n := 0; n < lockTries; n++ {
		f, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
		}
		// Jitter keeps many contenders from polling in lockstep.
		time.Sleep(lockWait + time.Duration(rand.Int64N(int64(lockWait))))
	}
	if err != nil {
		return fmt.Errorf("%w: %s", core.ErrBusy, name)
	}
	// Record who holds the lock so a crashed owner can later be told apart
	// from a live one. An unprovable identity is written empty and is never
	// recoverable (it classifies indeterminate).
	start, state := s.probe(os.Getpid())
	if state != homestate.ProcessIdentityLive {
		start = ""
	}
	_, _ = fmt.Fprintf(f, "%d %s\n", os.Getpid(), start)
	_ = f.Close()
	defer func() { _ = os.Remove(path) }()
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

func (s *Store) writeLocal(i core.IssueRef, r Record) error {
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

// write persists r: to the authoritative copy first when one is attached
// (compare-and-swap against the revision this cache holds), then to the
// cache. A refusal leaves the cache untouched; r.Rev carries the new revision.
func (s *Store) write(i core.IssueRef, r *Record) error {
	if s.ref != nil {
		old, _, err := s.read(i)
		if err != nil {
			return err
		}
		pushed := *r
		pushed.Rev = 0
		data, err := json.Marshal(pushed)
		if err != nil {
			return err
		}
		s.announce()
		rev, err := s.ref.CompareAndSwap(refCtx(), "issue/"+strconv.Itoa(i.Number), old.Rev, data)
		if err != nil {
			return refErr(err)
		}
		r.Rev = rev
	}
	return s.writeLocal(i, *r)
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
// new dispatch on an owned Issue is Held. An Issue with a published PR takes
// only a repair claim, and a full PR backlog holds the claim.
func (s *Store) Acquire(i core.IssueRef, holder, dispatch string, run core.RunID) (Record, error) {
	return s.acquire(i, holder, "", dispatch, run)
}

// AcquireFor is Acquire for a registered agent: the record stores the
// session id as Holder and the display name separately.
func (s *Store) AcquireFor(i core.IssueRef, a core.Agent, dispatch string, run core.RunID) (Record, error) {
	var out Record
	err := s.withRegistryLock(func() error {
		if err := s.verifyAgent(a); err != nil {
			return err
		}
		var e error
		out, e = s.acquire(i, a.SessionID, a.DisplayName(), dispatch, run)
		return e
	})
	return out, err
}

func (s *Store) acquire(i core.IssueRef, holder, name, dispatch string, run core.RunID) (Record, error) {
	if holder == "" || dispatch == "" {
		return Record{}, fmt.Errorf("%w: holder and dispatch are required", core.ErrIDInvalid)
	}
	if s.maxOpenPRs < 1 {
		return Record{}, fmt.Errorf("%w", core.ErrBacklogUnset)
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
		if ok && cur.State == core.StateDone {
			return fmt.Errorf("%w: %s is done", core.ErrTransitionInvalid, i.Resource())
		}
		if ok && cur.Receipt != nil {
			return fmt.Errorf("%w: %s PR #%d", core.ErrRepairOnly, i.Resource(), cur.Receipt.PR)
		}
		n, err := s.openPRs()
		if err != nil {
			return err
		}
		if n >= s.maxOpenPRs {
			return fmt.Errorf("%w: %d open of %d", core.ErrBacklogFull, n, s.maxOpenPRs)
		}
		out = Record{
			Issue: i.Number, Holder: holder, AgentName: name, Dispatch: dispatch, Run: run,
			Epoch: cur.Epoch + 1, Seq: cur.Seq + 1, State: core.StateClaimed,
		}
		out.History = append(append([]Event(nil), cur.History...), Event{Kind: "claim", Agent: name, Session: holder})
		return s.write(i, &out)
	})
	if err != nil {
		return Record{}, err
	}
	return out, nil
}

// AcquireRepair admits one holder to repair a handed-off Issue on the branch
// its receipt names. Contention is Held, the same dispatch again is a
// duplicate, and a different branch is refused: a repair never forks.
func (s *Store) AcquireRepair(i core.IssueRef, holder, dispatch string, run core.RunID, branch string) (Record, error) {
	return s.acquireRepair(i, holder, "", dispatch, run, branch)
}

// AcquireRepairFor is AcquireRepair for a registered agent.
func (s *Store) AcquireRepairFor(i core.IssueRef, a core.Agent, dispatch string, run core.RunID, branch string) (Record, error) {
	var out Record
	err := s.withRegistryLock(func() error {
		if err := s.verifyAgent(a); err != nil {
			return err
		}
		var e error
		out, e = s.acquireRepair(i, a.SessionID, a.DisplayName(), dispatch, run, branch)
		return e
	})
	return out, err
}

func (s *Store) acquireRepair(i core.IssueRef, holder, name, dispatch string, run core.RunID, branch string) (Record, error) {
	if holder == "" || dispatch == "" {
		return Record{}, fmt.Errorf("%w: holder and dispatch are required", core.ErrIDInvalid)
	}
	var out Record
	err := s.withLock(i, func() error {
		cur, ok, err := s.read(i)
		if err != nil {
			return err
		}
		if !ok {
			return core.ErrNotHeld
		}
		if cur.Holder != "" {
			if cur.Dispatch == dispatch {
				return fmt.Errorf("%w: %s dispatch %q", core.ErrDuplicateDispatch, i.Resource(), dispatch)
			}
			return fmt.Errorf("%w: %s held by %q", core.ErrHeld, i.Resource(), cur.Holder)
		}
		if cur.Receipt == nil || cur.State != core.StateVerifying || cur.Receipt.Branch != branch {
			return fmt.Errorf("%w: no reopened PR on branch %q for %s", core.ErrTransitionInvalid, branch, i.Resource())
		}
		out = cur
		out.Holder, out.AgentName, out.Dispatch, out.Run = holder, name, dispatch, run
		out.Epoch, out.Seq, out.State = cur.Epoch+1, cur.Seq+1, core.StateClaimed
		out.Return, out.Reason, out.Resume = "", "", ""
		out.History = append(append([]Event(nil), cur.History...), Event{Kind: "repair-claim", Agent: name, Session: holder})
		return s.write(i, &out)
	})
	if err != nil {
		return Record{}, err
	}
	return out, nil
}

// openPRs counts handed-off Issues that are not done.
func (s *Store) openPRs() (int, error) {
	recs, err := s.List()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range recs {
		if r.Receipt != nil && r.State != core.StateDone {
			n++
		}
	}
	return n, nil
}

// Read returns the stored record including a released one (holder empty), or
// ErrNotHeld when the Issue has never been claimed.
func (s *Store) Read(i core.IssueRef) (Record, error) {
	r, ok, err := s.read(i)
	if err != nil {
		return Record{}, err
	}
	if !ok {
		return Record{}, core.ErrNotHeld
	}
	return r, nil
}

// List returns every stored record (held, released, handed off) ordered by
// Issue number. A leader that restarts rebuilds its view from it.
func (s *Store) List() ([]Record, error) {
	ents, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	var out []Record
	for _, e := range ents {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "issue-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, "issue-"), ".json"))
		if err != nil || n <= 0 {
			continue
		}
		ref, _ := core.NewIssueRef(n)
		r, ok, err := s.read(ref)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Issue < out[b].Issue })
	return out, nil
}

// UpdateRelease verifies ownership, applies fn, and releases the claim in the
// same write. A refused or failed update leaves the holder in place.
func (s *Store) UpdateRelease(i core.IssueRef, holder string, epoch int64, fn func(*Record) error) (Record, error) {
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
		next.Issue, next.AgentName, next.Epoch, next.Seq = cur.Issue, cur.AgentName, cur.Epoch, cur.Seq+1
		next.Holder = ""
		out = next
		return s.write(i, &out)
	})
	if err != nil {
		return Record{}, err
	}
	return out, nil
}

// UpdateReleased applies fn to a record nobody holds (a handed-off or done
// Issue), for the leader or monitor that tracks the PR. It refuses a held
// record, so it can never race an agent's own writes.
func (s *Store) UpdateReleased(i core.IssueRef, fn func(*Record) error) (Record, error) {
	var out Record
	err := s.withLock(i, func() error {
		cur, ok, err := s.read(i)
		if err != nil {
			return err
		}
		if !ok {
			return core.ErrNotHeld
		}
		if cur.Holder != "" {
			return fmt.Errorf("%w: %s is held by %q", core.ErrTransitionInvalid, i.Resource(), cur.Holder)
		}
		next := cur
		if err := fn(&next); err != nil {
			return err
		}
		next.Issue, next.Holder, next.AgentName, next.Epoch, next.Seq = cur.Issue, "", cur.AgentName, cur.Epoch, cur.Seq+1
		out = next
		return s.write(i, &out)
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
		next.Issue, next.Holder, next.AgentName, next.Epoch, next.Seq = cur.Issue, cur.Holder, cur.AgentName, cur.Epoch, cur.Seq+1
		out = next
		return s.write(i, &out)
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
		return s.write(i, &cur)
	})
}
