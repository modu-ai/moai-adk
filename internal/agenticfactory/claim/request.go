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

// Stage is a background-intake stage. Only the first three are stored;
// StageStarted is derived from the claim (see ViewRequest) and can never be
// written.
type Stage string

const (
	StageACK            Stage = "ACK"
	StageIssuePublished Stage = "ISSUE_PUBLISHED"
	StageWTReady        Stage = "WT_READY"
	StageStarted        Stage = "STARTED" // derived only
)

// Park reasons for a request that is held outside the in-flight set.
const (
	ParkBackpressure = "backpressure"
	ParkNoSlot       = "no-slot"
)

// StageEvidence is the proof left at one stage.
type StageEvidence struct {
	Stage Stage  `json:"stage"`
	Ref   string `json:"ref"`
}

// Ready is the preparer's READY receipt: the worktree exists and the preparer
// has stopped writing to it.
type Ready struct {
	Slot          string `json:"slot"`
	Path          string `json:"path"`
	Head          string `json:"head"`
	WriterStopped bool   `json:"writer_stopped"`
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
	Slot        string          `json:"slot,omitempty"`
	Attempts    int             `json:"attempts,omitempty"`
	Retryable   bool            `json:"retryable,omitempty"`
	LastError   string          `json:"last_error,omitempty"`
	Evidence    []StageEvidence `json:"evidence"`
	Ready       *Ready          `json:"ready,omitempty"`
	Rev         int64           `json:"rev,omitempty"`
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
	if r.Stage != StageWTReady || r.Issue <= 0 {
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
			Key: key, PayloadHash: payloadHash, Priority: priority, Seq: seq + 1, Stage: StageACK,
			Evidence: []StageEvidence{{Stage: StageACK, Ref: "accepted"}},
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
		r.Stage = StageStarted
		r.Evidence = append(append([]StageEvidence(nil), r.Evidence...), StageEvidence{Stage: StageStarted, Ref: ref})
	}
	return r
}

// ViewRequest returns the request with STARTED derived from the claim state.
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

var stageRank = map[Stage]int{StageACK: 0, StageIssuePublished: 1, StageWTReady: 2}

// UpdateRequest applies fn to a copy. Key, payload hash, priority and arrival
// order are fixed; the stage stays or advances by one stored step; STARTED can
// never be written.
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

// ReadyGate refuses to let an agent begin writing before the preparer's READY
// receipt exists. An Issue that no intake request names is not gated.
func (s *Store) ReadyGate(issue int) error {
	all, err := s.listRequests()
	if err != nil {
		return err
	}
	for _, r := range all {
		if r.Issue != issue {
			continue
		}
		if r.Parked != "" || r.Stage != StageWTReady || r.Ready == nil || !r.Ready.WriterStopped {
			return fmt.Errorf("%w: worktree for issue-%d is not ready (%s)", core.ErrTransitionInvalid, issue, r.Stage)
		}
		return nil
	}
	return nil
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
