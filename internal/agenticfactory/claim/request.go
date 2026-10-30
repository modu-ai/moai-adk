package claim

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// Stage is a stored background-intake readiness stage (S51). They are not
// product states: ACCEPTED is not BUILDING, and BUILDING begins only after
// READY_ACKED. Whether the agent has begun is not a stage; ViewRequest derives
// it from the claim (Started).
type Stage string

const (
	StageAccepted         Stage = "ACCEPTED"          // intent and approved scope are fixed
	StageIssueConfirmed   Stage = "ISSUE_CONFIRMED"   // the real Issue is bound to the request by a lookup
	StageBranchRegistered Stage = "BRANCH_REGISTERED" // ref, base OID and any remote branch are checked
	StageAgentReserved    Stage = "AGENT_RESERVED"    // one agent and a slot are reserved; no development yet
	StageWTPrepared       Stage = "WT_PREPARED"       // the local worktree is verified and the preparer has stopped writing
	StageReadyAcked       Stage = "READY_ACKED"       // the reserved agent accepted the receipt; BUILDING may follow
)

// Park reasons for a request that is held outside the in-flight set. They say
// the request is waiting, not that anything failed.
const (
	ParkBackpressure = "backpressure"
	ParkNoSlot       = "no-slot"
)

// StageEvidence is the proof left at one stage.
type StageEvidence struct {
	Stage Stage  `json:"stage"`
	Ref   string `json:"ref"`
}

// Branch is the BRANCH_REGISTERED receipt: the ref, the base OID and, when a
// remote branch was needed, its remote OID. BaseOID is a full SHA.
type Branch struct {
	Ref       string `json:"ref"`
	BaseOID   string `json:"base_oid"`
	Remote    bool   `json:"remote"`
	RemoteOID string `json:"remote_oid,omitempty"`
}

// Ready is the preparer's WT_PREPARED receipt: where the worktree is, and that
// the preparer has stopped writing to it. Host and Path identify the local
// worktree; nothing assumes two hosts share a path.
type Ready struct {
	Host          string `json:"host"`
	Slot          string `json:"slot"`
	Path          string `json:"path"`
	Head          string `json:"head"`
	WriterStopped bool   `json:"writer_stopped"`
}

// Blocked is the BLOCKED or EFFECT_UNKNOWN row: the step that failed, why, and
// the condition to resume. Unknown means an effect may have happened (an Issue
// may exist) and must be looked up, never repeated.
type Blocked struct {
	Step      Stage  `json:"step"`
	Detail    string `json:"detail"`
	Retryable bool   `json:"retryable"`
	Unknown   bool   `json:"unknown,omitempty"`
	Resume    string `json:"resume,omitempty"`
	// Reason names a policy block (for example not-issue-author) in the closed
	// reason set; Candidate is the existing Issue that caused it, if any.
	Reason    core.BlockReason `json:"reason,omitempty"`
	Candidate int              `json:"candidate,omitempty"`
}

// Request is a durable background intake record. The payload itself is never
// stored, only its hash.
type Request struct {
	Key         string          `json:"key"`
	PayloadHash string          `json:"payload_hash"`
	Priority    int             `json:"priority"`
	Seq         int64           `json:"seq"`
	Stage       Stage           `json:"stage"`
	Parked      string          `json:"parked,omitempty"`
	Issue       int             `json:"issue,omitempty"`
	Branch      *Branch         `json:"branch,omitempty"`
	Agent       string          `json:"agent,omitempty"` // reserved agent's session id
	AgentName   string          `json:"agent_name,omitempty"`
	Slot        string          `json:"slot,omitempty"`
	Ready       *Ready          `json:"ready,omitempty"`
	AckedBy     string          `json:"acked_by,omitempty"`
	Blocked     *Blocked        `json:"blocked,omitempty"`
	Attempts    int             `json:"attempts,omitempty"`
	Evidence    []StageEvidence `json:"evidence"`
	Rev         int64           `json:"rev,omitempty"`

	// Started and StartedRef are filled only by ViewRequest(s), from the claim.
	// They are never stored.
	Started    bool   `json:"started,omitempty"`
	StartedRef string `json:"started_ref,omitempty"`

	// Fields of the first (S46) six-step layout, read once and migrated by
	// normalizeLegacy; they are never written.
	LegacyRetryable bool   `json:"retryable,omitempty"`
	LegacyLastError string `json:"last_error,omitempty"`
}

// normalizeLegacy migrates a record written by the first layout (stages ACK,
// ISSUE_PUBLISHED, WT_READY). The rule: a legacy record is never placed above
// the highest S51 stage whose evidence it carries. ACK becomes ACCEPTED. An
// ISSUE_PUBLISHED or WT_READY record had its Issue recorded without the
// re-lookup S51 requires, and WT_READY carries no branch or agent evidence, so
// both return to ACCEPTED as EFFECT_UNKNOWN: the Issue number is kept and a
// lookup confirms it, the old slot and worktree receipt are kept as data, and
// nothing is created again. A legacy failure of the publish step is treated as
// unknown too, because the old code could not tell "not sent" from "lost".
func normalizeLegacy(r *Request) {
	was := r.Stage
	switch was {
	case "ACK":
		r.Stage = StageAccepted
	case "ISSUE_PUBLISHED", "WT_READY":
		r.Stage = StageAccepted
		detail := "legacy ISSUE_PUBLISHED: the issue was recorded without a confirming lookup"
		if was == "WT_READY" {
			detail = "legacy WT_READY: no branch or agent evidence exists for this request"
		}
		r.Blocked = &Blocked{Step: StageAccepted, Detail: detail, Unknown: true, Resume: "FindIssue confirms the recorded issue"}
	}
	if r.LegacyRetryable || r.LegacyLastError != "" {
		if r.Blocked == nil {
			r.Blocked = &Blocked{Step: r.Stage, Detail: r.LegacyLastError, Retryable: r.LegacyRetryable, Unknown: was == "ACK" && r.LegacyRetryable}
			if r.Blocked.Unknown {
				r.Blocked.Resume = "FindIssue decides whether the issue was created"
			}
		}
		r.LegacyRetryable, r.LegacyLastError = false, ""
	}
}

// Policy bounds intake: requests in flight, and requests parked behind them.
type Policy struct{ MaxInFlight, MaxParked int }

func (s *Store) requestsDir() string { return filepath.Join(s.dir, "requests") }

func (s *Store) requestPath(key string) string {
	return filepath.Join(s.requestsDir(), key+".json")
}

func (s *Store) slotsDir() string { return filepath.Join(s.dir, "slots") }

// withRequests serialises every request write: sequence numbers and the
// in-flight count are decided inside one critical section.
func (s *Store) withRequests(fn func() error) error {
	if err := os.MkdirAll(s.requestsDir(), 0o700); err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return s.withLockFile(filepath.Join(s.dir, "requests.lock"), "requests", fn)
}

func (s *Store) readRequest(key string) (Request, bool, error) {
	b, err := os.ReadFile(s.requestPath(key))
	if errors.Is(err, os.ErrNotExist) {
		return Request{}, false, nil
	}
	if err != nil {
		return Request{}, false, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	var r Request
	if err := json.Unmarshal(b, &r); err != nil {
		return Request{}, false, fmt.Errorf("%w: corrupt request: %v", core.ErrStoreUnavailable, err)
	}
	normalizeLegacy(&r)
	return r, true, nil
}

func (s *Store) writeRequestLocal(r Request) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.requestsDir(), "req.tmp-*")
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
	if err := os.Rename(tmp.Name(), s.requestPath(r.Key)); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return nil
}

func (s *Store) listRequests() ([]Request, error) {
	ents, err := os.ReadDir(s.requestsDir())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	var out []Request
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		r, ok, err := s.readRequest(strings.TrimSuffix(e.Name(), ".json"))
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Seq < out[b].Seq })
	return out, nil
}

// writeRequest persists r: authority first when attached, then the cache.
func (s *Store) writeRequest(r *Request) error {
	r.Started, r.StartedRef = false, ""
	if s.ref != nil {
		old, _, err := s.readRequest(r.Key)
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
		rev, err := s.ref.CompareAndSwap(refCtx(), "request/"+r.Key, old.Rev, data)
		if err != nil {
			return refErr(err)
		}
		r.Rev = rev
	}
	return s.writeRequestLocal(*r)
}

// started reports whether the request's Issue has begun an attempt, and the
// evidence reference for it. This is the only definition of STARTED.
func (s *Store) started(r Request) (string, bool) {
	if r.Stage != StageReadyAcked || r.Issue <= 0 {
		return "", false
	}
	ref, err := core.NewIssueRef(r.Issue)
	if err != nil {
		return "", false
	}
	rec, ok, err := s.read(ref)
	if err != nil || !ok || rec.AttemptSeq == 0 {
		return "", false
	}
	who := rec.AgentName
	if who == "" {
		who = rec.Holder
	}
	return fmt.Sprintf("attempt %d by %s", rec.AttemptSeq, who), true
}

func (s *Store) inFlight(all []Request) int {
	n := 0
	for _, r := range all {
		if r.Parked != "" {
			continue
		}
		if _, st := s.started(r); st {
			continue
		}
		n++
	}
	return n
}

// PutRequest records a request durably. The same key with the same payload
// hash returns the existing record (created=false); the same key with another
// hash is ErrConflict. Admission follows Policy: in flight, else parked, else
// ErrIntakeFull with nothing recorded.
func (s *Store) PutRequest(key, payloadHash string, priority int, p Policy) (Request, bool, error) {
	if _, err := core.ParseRunID(key); err != nil {
		return Request{}, false, err
	}
	var out Request
	var created bool
	err := s.withRequests(func() error {
		cur, ok, err := s.readRequest(key)
		if err != nil {
			return err
		}
		if ok {
			if cur.PayloadHash != payloadHash {
				return fmt.Errorf("%w: %s", core.ErrConflict, key)
			}
			out = cur
			return nil
		}
		all, err := s.listRequests()
		if err != nil {
			return err
		}
		var seq int64
		parked := 0
		for _, r := range all {
			if r.Seq > seq {
				seq = r.Seq
			}
			if r.Parked != "" {
				parked++
			}
		}
		out = Request{
			Key: key, PayloadHash: payloadHash, Priority: priority, Seq: seq + 1, Stage: StageAccepted,
			Evidence: []StageEvidence{{Stage: StageAccepted, Ref: "accepted"}},
		}
		switch {
		case s.inFlight(all) < p.MaxInFlight:
		case parked < p.MaxParked:
			out.Parked = ParkBackpressure
		default:
			out = Request{}
			return fmt.Errorf("%w: %d in flight, %d parked", core.ErrIntakeFull, s.inFlight(all), parked)
		}
		created = true
		return s.writeRequest(&out)
	})
	if err != nil {
		return Request{}, false, err
	}
	return out, created, nil
}

// GetRequest returns the stored record; ErrNotHeld when there is none.
func (s *Store) GetRequest(key string) (Request, error) {
	r, ok, err := s.readRequest(key)
	if err != nil {
		return Request{}, err
	}
	if !ok {
		return Request{}, core.ErrNotHeld
	}
	return r, nil
}

// ListRequests returns every stored request in arrival order.
func (s *Store) ListRequests() ([]Request, error) { return s.listRequests() }

func (s *Store) view(r Request) Request {
	if ref, ok := s.started(r); ok {
		r.Started, r.StartedRef = true, ref
	}
	return r
}

// ViewRequest returns the request with Started derived from the claim state.
// Nothing is written: a reader (the monitor) sees the same value the owner does.
func (s *Store) ViewRequest(key string) (Request, error) {
	r, err := s.GetRequest(key)
	if err != nil {
		return Request{}, err
	}
	return s.view(r), nil
}

// ViewRequests is ViewRequest for every request, in arrival order.
func (s *Store) ViewRequests() ([]Request, error) {
	all, err := s.listRequests()
	if err != nil {
		return nil, err
	}
	for i := range all {
		all[i] = s.view(all[i])
	}
	return all, nil
}

var stageRank = map[Stage]int{
	StageAccepted: 0, StageIssueConfirmed: 1, StageBranchRegistered: 2,
	StageAgentReserved: 3, StageWTPrepared: 4, StageReadyAcked: 5,
}

// UpdateRequest applies fn to a copy. Key, payload hash, priority and arrival
// order are fixed; the stage stays or advances by one step.
func (s *Store) UpdateRequest(key string, fn func(*Request) error) (Request, error) {
	var out Request
	err := s.withRequests(func() error {
		cur, ok, err := s.readRequest(key)
		if err != nil {
			return err
		}
		if !ok {
			return core.ErrNotHeld
		}
		next := cur
		if err := fn(&next); err != nil {
			return err
		}
		next.Key, next.PayloadHash, next.Priority, next.Seq = cur.Key, cur.PayloadHash, cur.Priority, cur.Seq
		from, fok := stageRank[cur.Stage]
		to, tok := stageRank[next.Stage]
		if !fok || !tok || (to != from && to != from+1) {
			return fmt.Errorf("%w: request stage %s -> %s", core.ErrTransitionInvalid, cur.Stage, next.Stage)
		}
		out = next
		return s.writeRequest(&out)
	})
	if err != nil {
		return Request{}, err
	}
	return out, nil
}

// PromoteParked unparks requests by priority (highest first), then arrival.
// A backpressure park needs a free in-flight place; a no-slot park needs a
// free worktree slot (freeSlots).
func (s *Store) PromoteParked(p Policy, freeSlots int) ([]Request, error) {
	var promoted []Request
	err := s.withRequests(func() error {
		all, err := s.listRequests()
		if err != nil {
			return err
		}
		inflight := s.inFlight(all)
		var parked []Request
		for _, r := range all {
			if r.Parked != "" {
				parked = append(parked, r)
			}
		}
		sort.SliceStable(parked, func(a, b int) bool {
			if parked[a].Priority != parked[b].Priority {
				return parked[a].Priority > parked[b].Priority
			}
			return parked[a].Seq < parked[b].Seq
		})
		for _, r := range parked {
			switch r.Parked {
			case ParkBackpressure:
				if inflight >= p.MaxInFlight {
					continue
				}
				inflight++
			case ParkNoSlot:
				if freeSlots <= 0 {
					continue
				}
				freeSlots--
			default:
				continue
			}
			r.Parked = ""
			if err := s.writeRequest(&r); err != nil {
				return err
			}
			promoted = append(promoted, r)
		}
		return nil
	})
	return promoted, err
}

// ReadyGate refuses to let an agent begin writing before the whole readiness
// chain is done: the request must be READY_ACKED, not parked, not blocked, the
// preparer's receipt must say it stopped writing, and the agent that acked must
// be the one now starting. An Issue that no intake request names is not gated. When
// several requests name the Issue the newest one is read (audit F21).
func (s *Store) ReadyGate(issue int, holder string) error {
	all, err := s.listRequests()
	if err != nil {
		return err
	}
	// Several requests can name one Issue. The newest (highest Seq; listRequests sorts
	// by it) is the current attempt and the one the gate reads; an older one is history
	// and neither blocks a newer, acknowledged request nor passes a newer one that is
	// not ready.
	var r *Request
	for k := range all {
		if all[k].Issue == issue {
			r = &all[k]
		}
	}
	if r == nil {
		return nil
	}
	switch {
	case r.Parked != "" || r.Blocked != nil || r.Stage != StageReadyAcked:
		return fmt.Errorf("%w: worktree for issue-%d is not ready (%s)", core.ErrTransitionInvalid, issue, r.Stage)
	case r.Ready == nil || !r.Ready.WriterStopped:
		return fmt.Errorf("%w: the preparer has not stopped writing for issue-%d", core.ErrTransitionInvalid, issue)
	case r.AckedBy == "" || r.AckedBy != holder:
		return fmt.Errorf("%w: issue-%d was acknowledged by another agent", core.ErrTransitionInvalid, issue)
	}
	return nil
}

// VerifyAgent reports whether a is still the registered holder of its number.
func (s *Store) VerifyAgent(a core.Agent) error { return s.verifyAgent(a) }

// ReserveRequest reserves one agent and one worktree slot for a request that
// has its branch registered (AGENT_RESERVED). It does not allow development:
// that waits for the preparer and the agent's own ack. An agent that is already
// reserved for, or working on, another request is refused (ErrHeld), and with
// no free slot the request is parked (ErrNoSlot) instead of being half
// reserved. A slot the request already holds (a migrated record) is reused.
func (s *Store) ReserveRequest(key string, a core.Agent, slots []string) (Request, error) {
	var out Request
	err := s.withRequests(func() error {
		cur, ok, err := s.readRequest(key)
		if err != nil {
			return err
		}
		if !ok {
			return core.ErrNotHeld
		}
		if cur.Stage != StageBranchRegistered || cur.Parked != "" || cur.Blocked != nil {
			return fmt.Errorf("%w: %s is %s, an agent is reserved once the branch is registered and nothing is blocked", core.ErrTransitionInvalid, key, cur.Stage)
		}
		if err := s.verifyAgent(a); err != nil {
			return err
		}
		all, err := s.listRequests()
		if err != nil {
			return err
		}
		for _, o := range all {
			if o.Key != key && o.Agent == a.SessionID && s.holdsAgent(o) {
				return fmt.Errorf("%w: %s is already reserved for %s", core.ErrHeld, a.DisplayName(), o.Key)
			}
		}
		slot := cur.Slot
		if slot != "" {
			if err := s.AcquireSlot(slot, key); err != nil {
				return err
			}
		} else {
			for _, name := range slots {
				if aerr := s.AcquireSlot(name, key); aerr == nil {
					slot = name
					break
				} else if !errors.Is(aerr, core.ErrHeld) {
					return aerr
				}
			}
			if slot == "" {
				cur.Parked = ParkNoSlot
				if werr := s.writeRequest(&cur); werr != nil {
					return werr
				}
				return fmt.Errorf("%w: %s", core.ErrNoSlot, key)
			}
		}
		cur.Stage, cur.Agent, cur.AgentName, cur.Slot = StageAgentReserved, a.SessionID, a.DisplayName(), slot
		cur.Evidence = append(cur.Evidence, StageEvidence{Stage: StageAgentReserved, Ref: a.DisplayName() + "@" + slot})
		out = cur
		return s.writeRequest(&out)
	})
	if err != nil {
		return Request{}, err
	}
	return out, nil
}

// holdsAgent reports whether request o still keeps its reserved agent busy: it
// is reserved, prepared or acked and not yet started, or it started and the
// agent's claim on the Issue is still held. Once the agent handed the PR off
// (the claim is released) the agent is free for a new request.
func (s *Store) holdsAgent(o Request) bool {
	if o.Stage != StageAgentReserved && o.Stage != StageWTPrepared && o.Stage != StageReadyAcked {
		return false
	}
	if _, started := s.started(o); !started {
		return true
	}
	ref, err := core.NewIssueRef(o.Issue)
	if err != nil {
		return true
	}
	rec, found, err := s.read(ref)
	return err != nil || !found || rec.Holder != ""
}

// AckRequestReady records the reserved agent's explicit acceptance of the
// preparation receipt (READY_ACKED). Only the agent the request is reserved for
// may ack, only after the preparer stopped writing.
func (s *Store) AckRequestReady(key string, a core.Agent) (Request, error) {
	var out Request
	err := s.withRequests(func() error {
		cur, ok, err := s.readRequest(key)
		if err != nil {
			return err
		}
		if !ok {
			return core.ErrNotHeld
		}
		if cur.Stage != StageWTPrepared || cur.Parked != "" || cur.Blocked != nil {
			return fmt.Errorf("%w: %s is %s, the receipt is acknowledged once the worktree is prepared", core.ErrTransitionInvalid, key, cur.Stage)
		}
		if err := s.verifyAgent(a); err != nil {
			return err
		}
		if cur.Agent != a.SessionID {
			return fmt.Errorf("%w: %s is reserved for %s, not %s", core.ErrForeign, key, cur.AgentName, a.DisplayName())
		}
		if cur.Ready == nil || !cur.Ready.WriterStopped {
			return fmt.Errorf("%w: the preparer has not stopped writing", core.ErrEvidenceIncomplete)
		}
		cur.Stage, cur.AckedBy = StageReadyAcked, a.SessionID
		cur.Evidence = append(cur.Evidence, StageEvidence{Stage: StageReadyAcked, Ref: "acked by " + a.DisplayName()})
		out = cur
		return s.writeRequest(&out)
	})
	if err != nil {
		return Request{}, err
	}
	return out, nil
}

type slotFile struct {
	Owner string `json:"owner"`
}

func (s *Store) slotPath(name string) string { return filepath.Join(s.slotsDir(), name+".json") }

// AcquireSlot takes a worktree slot exclusively. The owner asking again is
// idempotent; anyone else is Held. The slot file appears fully written or not
// at all.
func (s *Store) AcquireSlot(name, owner string) error {
	if _, err := core.ParseRunID(name); err != nil {
		return err
	}
	if owner == "" {
		return fmt.Errorf("%w: slot owner is required", core.ErrIDInvalid)
	}
	if err := os.MkdirAll(s.slotsDir(), 0o700); err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	b, _ := json.Marshal(slotFile{Owner: owner})
	tmp, err := os.CreateTemp(s.slotsDir(), "slot.tmp-*")
	if err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	if err := os.Link(tmp.Name(), s.slotPath(name)); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
		}
		cur, oerr := s.SlotOwner(name)
		if oerr != nil {
			return oerr
		}
		if cur == owner {
			return nil
		}
		return fmt.Errorf("%w: slot %s is held by %q", core.ErrHeld, name, cur)
	}
	return nil
}

// SlotOwner returns the slot's owner, or "" when it is free.
func (s *Store) SlotOwner(name string) (string, error) {
	b, err := os.ReadFile(s.slotPath(name))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	var f slotFile
	if err := json.Unmarshal(b, &f); err != nil {
		return "", fmt.Errorf("%w: corrupt slot: %v", core.ErrStoreUnavailable, err)
	}
	return f.Owner, nil
}

// ReleaseSlot frees a slot; only its owner may.
func (s *Store) ReleaseSlot(name, owner string) error {
	cur, err := s.SlotOwner(name)
	if err != nil {
		return err
	}
	if cur == "" {
		return core.ErrNotHeld
	}
	if cur != owner {
		return fmt.Errorf("%w: slot %s belongs to %q", core.ErrForeign, name, cur)
	}
	if err := os.Remove(s.slotPath(name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %v", core.ErrStoreUnavailable, err)
	}
	return nil
}
