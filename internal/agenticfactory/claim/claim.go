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
	// PreflightHead is the HEAD whose PR text was checked for closing keywords
	// that would close another author's Issue; the PR is published only for it.
	PreflightHead core.HeadSHA `json:"preflight_head,omitempty"`
	// Merged is the preserved fact that the PR reached the remote; it is
	// never reverted. Pending names the obligation (local-sync, deploy,
	// issue-close) that still keeps a merged Issue out of done.
	Merged  bool   `json:"merged,omitempty"`
	Pending string `json:"pending_effect,omitempty"`
	// History is the append-only trail of claims, handoffs and reopenings
	// under this one Issue key.
	History []Event `json:"history,omitempty"`

	// Stage, Sub and Park are the S46 position, computed from State, Return,
	// Reason, the receipt and Merged on every write and every read (step 2 of
	// the model migration: State stays the runner's progress record). Gate is the
	// merge-gate record on a stage and Refusal the last refused operation; neither
	// moves the state. Park is an object under its own key; `parked` is the intake
	// queue's string key and is not used here.
	Stage   core.Stage    `json:"stage,omitempty"`
	Sub     core.SubStage `json:"sub_stage,omitempty"`
	Park    *core.Park    `json:"park,omitempty"`
	Gate    *core.Gate    `json:"gate,omitempty"`
	Refusal *core.Refusal `json:"refusal,omitempty"`
	// Wait is what the owner waits for and what happens next; Repair is the E06
	// evidence while the owner repairs a PR it handed off.
	Wait   *core.Wait `json:"wait,omitempty"`
	Repair *Repair    `json:"repair,omitempty"`
	// RepairNeeded is the signal (CI failed, review asked for changes) that the
	// owner must repair its open PR before it takes new work. Worktree is the
	// write worktree bound to this Issue's work.
	RepairNeeded string `json:"repair_needed,omitempty"`
	// Revision is the approved contract revision applied to the current attempt
	// (canon E18 adopt); a new receipt ends it.
	Revision string `json:"revision,omitempty"`
	// PositionUnknown is set on a read when the S46 position could not be derived
	// from this record; it holds the reason and is never stored. Stage, Sub and Park
	// are empty then, which means "not known", not "no park".
	PositionUnknown string `json:"-"`
	Worktree        string `json:"worktree,omitempty"`
}

// Repair is the canon E06 evidence (REVIEWING -> BUILDING/FIX): the repair
// decision, the exclusive owner, the scope reserved again and the previous
// writer ended. All four are required.
type Repair struct {
	Decision       string `json:"decision"`
	Owner          string `json:"exclusive_owner"`
	ScopeReserved  bool   `json:"scope_reserved"`
	OldWriterEnded bool   `json:"old_writer_ended"`
}

func (p Repair) validate(holder string) error {
	switch {
	case strings.TrimSpace(p.Decision) == "":
		return fmt.Errorf("%w: the repair decision is missing", core.ErrEvidenceIncomplete)
	case p.Owner == "" || (holder != "" && p.Owner != holder):
		return fmt.Errorf("%w: the repair needs the exclusive owner", core.ErrEvidenceIncomplete)
	case !p.ScopeReserved:
		return fmt.Errorf("%w: the scope was not reserved again", core.ErrEvidenceIncomplete)
	case !p.OldWriterEnded:
		return fmt.Errorf("%w: the previous writer has not ended", core.ErrEvidenceIncomplete)
	}
	return nil
}

// The fixed codes PositionUnknown can hold. A reader (web, monitor) shows one of
// these, never the text of an error.
const (
	PositionUnmappable     = "unmappable-legacy-state" // the old state/return/reason cannot be mapped to the S46 model
	PositionInvalidGate    = "invalid-gate"
	PositionInvalidRefusal = "invalid-refusal"
	PositionInvalidWait    = "invalid-wait"
	PositionWaitOwner      = "wait-owner-mismatch"
	PositionInvalidRepair  = "invalid-repair"
	PositionUnknownOther   = "unknown" // anything else
)

// positionError is the error deriveStage returns: the cause unchanged for errors.Is,
// plus the fixed code a reader gets.
type positionError struct {
	code string
	err  error
}

func (e *positionError) Error() string { return e.err.Error() }
func (e *positionError) Unwrap() error { return e.err }

// positionCode is the fixed code of a deriveStage error.
func positionCode(err error) string {
	var pe *positionError
	if errors.As(err, &pe) {
		return pe.code
	}
	if errors.Is(err, core.ErrLegacyUnmappable) {
		return PositionUnmappable
	}
	return PositionUnknownOther
}

// deriveStage recomputes Stage, Sub and Park from the old fields. A legacy pause
// whose reason is a gate (G) or refusal (R) reason is read as a Gate or Refusal
// on the stage it returns to, because only overlay reasons may park. A record that
// cannot be mapped (a pause without a return stage, for instance) is an error.
func (r *Record) deriveStage() error {
	r.Stage, r.Sub, r.Park = "", "", nil
	// A position that is derived is known: the marker a read set does not outlive
	// a write that repaired the record.
	r.PositionUnknown = ""
	if r.Gate != nil {
		if err := r.Gate.Validate(); err != nil {
			return &positionError{PositionInvalidGate, err}
		}
	}
	if r.Refusal != nil {
		if err := r.Refusal.Validate(); err != nil {
			return &positionError{PositionInvalidRefusal, err}
		}
	}
	if r.Wait != nil {
		if err := r.Wait.Validate(); err != nil {
			return &positionError{PositionInvalidWait, err}
		}
		if r.Holder != "" && r.Wait.Owner != r.Holder {
			return &positionError{PositionWaitOwner, fmt.Errorf("%w: the wait is owned by %q, the Issue by %q", core.ErrOwnershipUnverified, r.Wait.Owner, r.Holder)}
		}
	}
	if r.Repair != nil {
		if err := r.Repair.validate(r.Holder); err != nil {
			return &positionError{PositionInvalidRepair, err}
		}
	}
	l := core.Legacy{
		State: r.State, Repair: r.Repair != nil, Revision: r.Revision != "", Return: r.Return, Reason: r.Reason, Resume: r.Resume,
		HasReceipt: r.Receipt != nil, LeaderAccepted: r.Receipt != nil && r.Receipt.LeaderAccepted, Merged: r.Merged,
	}
	if r.State.IsPause() && r.Reason.Valid() && !core.IsParkable(r.Reason) {
		resume := r.Resume
		if resume == "" {
			resume = "see the record"
		}
		detail := "migrated from a legacy " + string(r.State) + " record"
		if core.IsGateReason(r.Reason) && r.Gate == nil {
			r.Gate = &core.Gate{Reason: r.Reason, Detail: detail, Resume: resume, Head: r.Head}
		} else if !core.IsGateReason(r.Reason) && r.Refusal == nil {
			r.Refusal = &core.Refusal{Reason: r.Reason, Detail: detail}
		}
		l.State, l.Reason, l.Return = r.Return, "", ""
	}
	pos, err := core.FromLegacy(l)
	if err != nil {
		return &positionError{PositionUnmappable, err}
	}
	r.Stage, r.Sub, r.Park = pos.Stage, pos.Sub, pos.Park
	return nil
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
	HandoffID      string         `json:"handoff_id"`
	LeaderAccepted bool           `json:"leader_accepted"`
	PR             int            `json:"pr"`
	Head           core.HeadSHA   `json:"head"`
	Branch         string         `json:"branch"`
	Worktree       string         `json:"worktree"`
	Attempt        core.AttemptID `json:"attempt,omitempty"`
	AttemptSeq     int            `json:"attempt_seq,omitempty"`
	Agent          string         `json:"agent,omitempty"`   // display name of the agent that published the PR
	Session        string         `json:"session,omitempty"` // its session id; stays unambiguous after the number is reused
	// The spec fields of a receipt (spec 148). A receipt stored before they existed
	// reads with them empty; only a new handoff must supply them.
	PRURL       string `json:"pr_url,omitempty"`
	Base        string `json:"base,omitempty"`
	AdoptedSpec string `json:"adopted_spec,omitempty"`
	TaskResult  string `json:"task_result,omitempty"`
	CheckStatus string `json:"check_status,omitempty"`
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
	// sync makes a write durable before it is reported; the zero value syncs
	// for real, tests inject failures.
	sync syncer
	// afterScan is a test seam: it runs between the check of the other records
	// and the reservation in BindWorktree.
	afterScan func()
	// lockPol is how long a lock is waited for; the zero value is the built-in
	// polling (lockTries x lockWait on the real clock). Tests replace the clock.
	lockPol lockPolicy
}

// How a held lock is waited for: lockTries attempts, lockWait apart. The budget
// (about two seconds, longer when the machine oversleeps) has to carry a writer
// through the queue of the resident writers on the documents lock (the main loop,
// notify, webhook, the collector) when one holder is slow, which the old 40 tries
// (about 400ms) did not: under the deterministic harness in lockwait_test.go,
// four writers holding for 100ms were told Busy 17.8% of the time, for 200ms
// 42.8%, and four writers with a 450ms stall every 50th hold 4.5%. Two seconds
// gives no Busy in any of those cases, and in the 1500ms-stall case; a lock whose
// holder stays away longer still ends in Busy, the contract's retry signal, never
// in Held. The polling is still unfair (a waiter's turn depends on when its poll
// lands), which the tail wait shows: the budget is what covers it.
//
// @MX:DEBT: the wait is a fixed poll with no order among waiters, and its ceiling is paid by callers
// @MX:CEILING: a webhook answering 503 waits up to about two seconds (lockTries x lockWait) for a held lock;
// the tail wait is up to about twice the fair bound (4 writers x 200ms hold: p95 1s, max 1.2s against 600ms)
// @MX:UPGRADE: replace the poll with a fair queue (ticket order, with recovery of a dead waiter's ticket) when
// resident writers on the documents lock grow past four, or when a measured Busy appears under this budget
const (
	lockTries = 200
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
	tries, wait, sleep := s.lockPol.resolve()
	for n := 0; n < tries; n++ {
		f, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
		}
		sleep(wait)
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
	// An old record has no S46 position. One that cannot be mapped is still
	// returned, so a reader is never locked out and a listing does not fail for one
	// record, but it is marked: its empty position means "not known", and the
	// write path refuses the same record for the same reason (see write).
	r.PositionUnknown = ""
	if err := r.deriveStage(); err != nil {
		r.PositionUnknown = positionCode(err)
	}
	return r, true, nil
}

func (s *Store) writeLocal(i core.IssueRef, r Record) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return s.writeAtomic(s.dir, i.Resource()+".tmp-*", s.path(i), b)
}

// write persists r: to the authoritative copy first when one is attached
// (compare-and-swap against the revision this cache holds), then to the
// cache. A refusal leaves the cache untouched; r.Rev carries the new revision.
func (s *Store) write(i core.IssueRef, r *Record) error {
	if err := r.deriveStage(); err != nil {
		return err
	}
	// A write that puts the Issue into PARKED carries its wait record (design
	// 2.1c), and while the park stays no write takes that record away. A record
	// parked before this rule, with no wait record, is rewritten as it is.
	if r.Park != nil && r.Wait == nil {
		if prev, had, err := s.read(i); err != nil {
			return err
		} else if !had || prev.Park == nil {
			return fmt.Errorf("%w: a park needs a wait record (owner, next action, recheck time, diagnosis, retry limit, escalation target)", core.ErrEvidenceIncomplete)
		} else if prev.Wait != nil {
			// the park stays, so its wait record stays: only a parked record that
			// never had one (stored before the rule) may be rewritten without
			return fmt.Errorf("%w: a parked Issue keeps its wait record", core.ErrEvidenceIncomplete)
		}
	}
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
		if err := s.repairFirst(holder, i); err != nil {
			return err
		}
		if err := s.heldIssuesRecorded(holder, i); err != nil {
			return err
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
		if rc := cur.Receipt; rc != nil && rc.Session != "" && rc.Session != holder {
			// The PR belongs to the agent that published it. A receipt from before
			// the session was recorded names nobody and stays open to any holder.
			return fmt.Errorf("%w: issue-%d belongs to session %s", core.ErrNotResponsibleAgent, i.Number, rc.Session)
		}
		if cur.Receipt == nil || cur.State != core.StateVerifying || cur.Receipt.Branch != branch {
			return fmt.Errorf("%w: no reopened PR on branch %q for %s", core.ErrTransitionInvalid, branch, i.Resource())
		}
		out = cur
		out.Holder, out.AgentName, out.Dispatch, out.Run = holder, name, dispatch, run
		out.Epoch, out.Seq, out.State = cur.Epoch+1, cur.Seq+1, core.StateClaimed
		out.Return, out.Reason, out.Resume = "", "", ""
		out.Gate, out.Refusal = nil, nil
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
