package effect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"
)

// This file covers outward effects the factory performs while accepting and preparing work (S46,
// the "background intake" duty): creating an Issue, preparing a branch or worktree, observing
// checks. Rules, in order:
//
//  1. The intent is written durably BEFORE anything is sent, under a deterministic key.
//  2. A confirmed key is never sent again; a parked key is not sent until reconciled.
//  3. When a send's outcome is unknown (timeout, lost response, restart), the same key is looked
//     up in remote/local facts. Found -> confirmed. Authoritatively absent -> retryable. Anything
//     unsure -> PARKED, and nothing new is sent.
//  4. Every stage has a timeout; a failure is recorded as a retryable state.
//  5. Check observation is asynchronous: requesting it only records the intent.

// Kind is the type of outward effect.
type Kind string

const (
	IssueCreate     Kind = "issue-create"
	BranchPrepare   Kind = "branch-prepare"
	WorktreePrepare Kind = "worktree-prepare"
	CheckObserve    Kind = "check-observe"
	LedgerComment   Kind = "ledger-comment"
)

// Status is the life of one intent.
type Status string

const (
	StatusIntended   Status = "intended"   // recorded, nothing sent
	StatusSent       Status = "sent"       // send attempted; outcome not yet known
	StatusConfirmed  Status = "confirmed"  // the effect exists (provider id recorded)
	StatusFailed     Status = "failed"     // known not to have happened; safe to retry
	StatusParked     Status = "parked"     // outcome unknown; do not send until reconciled
	StatusSuperseded Status = "superseded" // replaced by a newer intent for the same target; never sent
)

// Intent is the durable record of one effect. It lives in the lane-18 claim store; this package
// only defines its shape and rules.
type Intent struct {
	Key        string `json:"key"`
	Kind       Kind   `json:"kind"`
	Issue      int    `json:"issue"`
	Target     string `json:"target"`
	Digest     string `json:"digest"`            // digest of the request payload
	Payload    string `json:"payload,omitempty"` // already-masked text to send, when the effect carries one
	Seq        int    `json:"seq,omitempty"`     // order among intents for one target
	Status     Status `json:"status"`
	Attempts   int    `json:"attempts"`
	ProviderID string `json:"provider_id,omitempty"`
	LastError  string `json:"last_error,omitempty"`
}

// Request describes an effect to perform.
type Request struct {
	Kind    Kind
	Issue   int
	Target  string
	Digest  string
	Payload string
}

// Key is the idempotency key: the same kind, target and payload always give the same key.
func Key(k Kind, target, digest string) string {
	h := sha256.Sum256([]byte(string(k) + "\x00" + target + "\x00" + digest))
	return hex.EncodeToString(h[:])[:32]
}

var markerRe = regexp.MustCompile(`<!-- agentic-factory:key=([0-9a-f]+) -->`)

// Marker is the text embedded in an Issue body so a later lookup can find it by key.
func Marker(key string) string { return "<!-- agentic-factory:key=" + key + " -->" }

// KeyFromMarker extracts the key from an Issue body, or "".
func KeyFromMarker(body string) string {
	if m := markerRe.FindStringSubmatch(body); m != nil {
		return m[1]
	}
	return ""
}

// ErrNotSent marks a send that provably never left (validation, no connection made). Only this
// error skips the lookup; every other error is treated as "may have been delivered".
var ErrNotSent = errors.New("effect: not sent")

// Sender performs the effect and returns the provider's id. Tests use fakes only.
type Sender interface {
	Send(ctx context.Context, in Intent) (providerID string, err error)
}

// LookupResult is what a fact query found. Authoritative means a "not found" can be trusted.
type LookupResult struct {
	Found         bool
	ProviderID    string
	Authoritative bool
}

// Lookuper queries remote and local facts for an intent's key (Issue body marker, branch and
// worktree path, check state).
type Lookuper interface {
	Lookup(ctx context.Context, in Intent) (LookupResult, error)
}

// IntentStore persists intents. It is a local cache behind this interface: the authority is the
// online control ref, and consumer ACK and cursor (the collector contract) are not implemented yet.
type IntentStore interface {
	Get(key string) (Intent, bool, error)
	Put(Intent) error
	List() ([]Intent, error)
}

// Timeouts bounds each stage.
type Timeouts struct {
	IssueCreate, BranchPrepare, WorktreePrepare, CheckObserve, LedgerComment time.Duration
}

// For returns the timeout for a kind, 0 when unset.
func (t Timeouts) For(k Kind) time.Duration {
	switch k {
	case IssueCreate:
		return t.IssueCreate
	case BranchPrepare:
		return t.BranchPrepare
	case WorktreePrepare:
		return t.WorktreePrepare
	case CheckObserve:
		return t.CheckObserve
	case LedgerComment:
		return t.LedgerComment
	}
	return 0
}

// Executor applies the rules above.
type Executor struct {
	Store    IntentStore
	Send     Sender
	Lookup   Lookuper
	Timeouts Timeouts
}

// Execute performs the request at most once per key.
func (e *Executor) Execute(ctx context.Context, r Request) (Intent, error) {
	to := e.Timeouts.For(r.Kind)
	if to <= 0 {
		return Intent{}, fmt.Errorf("effect: no timeout configured for %s", r.Kind)
	}
	key := Key(r.Kind, r.Target, r.Digest)
	in, ok, err := e.Store.Get(key)
	if err != nil {
		return Intent{}, err
	}
	if ok {
		switch in.Status {
		case StatusConfirmed, StatusParked:
			return in, nil
		case StatusSent:
			return e.Reconcile(ctx, key)
		}
	} else {
		in = Intent{Key: key, Kind: r.Kind, Issue: r.Issue, Target: r.Target, Digest: r.Digest, Payload: r.Payload, Status: StatusIntended}
		if err := e.Store.Put(in); err != nil {
			return Intent{}, err
		}
	}
	in.Attempts++
	in.Status = StatusSent
	in.LastError = ""
	if err := e.Store.Put(in); err != nil {
		return Intent{}, err
	}
	sctx, cancel := context.WithTimeout(ctx, to)
	id, serr := e.Send.Send(sctx, in)
	cancel()
	switch {
	case serr == nil:
		in.Status, in.ProviderID = StatusConfirmed, id
	case errors.Is(serr, ErrNotSent):
		in.Status, in.LastError = StatusFailed, serr.Error()
	default:
		in.LastError = serr.Error()
		if err := e.Store.Put(in); err != nil {
			return Intent{}, err
		}
		return e.Reconcile(ctx, key)
	}
	return in, e.Store.Put(in)
}

// Reconcile resolves a sent or parked intent from facts. It never sends.
func (e *Executor) Reconcile(ctx context.Context, key string) (Intent, error) {
	in, ok, err := e.Store.Get(key)
	if err != nil || !ok {
		return Intent{}, fmt.Errorf("effect: no intent %s: %v", key, err)
	}
	if in.Status != StatusSent && in.Status != StatusParked {
		return in, nil
	}
	to := e.Timeouts.For(in.Kind)
	if to <= 0 {
		return Intent{}, fmt.Errorf("effect: no timeout configured for %s", in.Kind)
	}
	lctx, cancel := context.WithTimeout(ctx, to)
	res, lerr := e.Lookup.Lookup(lctx, in)
	cancel()
	switch {
	case lerr != nil:
		in.Status, in.LastError = StatusParked, lerr.Error()
	case res.Found:
		in.Status, in.ProviderID = StatusConfirmed, res.ProviderID
	case res.Authoritative:
		in.Status = StatusFailed
	default:
		in.Status = StatusParked
	}
	return in, e.Store.Put(in)
}

// RequestObservation records the wish to observe checks and returns at once; the result arrives
// through Tick. The caller never waits on it.
func (e *Executor) RequestObservation(_ context.Context, r Request) (Intent, error) {
	if r.Kind != CheckObserve {
		return Intent{}, fmt.Errorf("effect: %s is not an observation", r.Kind)
	}
	return e.Record(r)
}

// Record stores the intent (idempotent per key) without sending anything.
func (e *Executor) Record(r Request) (Intent, error) {
	key := Key(r.Kind, r.Target, r.Digest)
	if in, ok, err := e.Store.Get(key); err != nil || ok {
		return in, err
	}
	in := Intent{Key: key, Kind: r.Kind, Issue: r.Issue, Target: r.Target, Digest: r.Digest, Status: StatusIntended}
	return in, e.Store.Put(in)
}

// Tick looks at every pending observation once and returns those that became confirmed. A lookup
// error or "not yet" leaves the observation pending for the next tick.
func (e *Executor) Tick(ctx context.Context) ([]Intent, error) {
	all, err := e.Store.List()
	if err != nil {
		return nil, err
	}
	var done []Intent
	for _, in := range all {
		if in.Kind != CheckObserve || in.Status != StatusIntended {
			continue
		}
		to := e.Timeouts.For(CheckObserve)
		if to <= 0 {
			return done, errors.New("effect: no timeout configured for check-observe")
		}
		lctx, cancel := context.WithTimeout(ctx, to)
		res, lerr := e.Lookup.Lookup(lctx, in)
		cancel()
		if lerr != nil {
			in.LastError = lerr.Error()
			if err := e.Store.Put(in); err != nil {
				return done, err
			}
			continue
		}
		if res.Found {
			in.Status, in.ProviderID = StatusConfirmed, res.ProviderID
			if err := e.Store.Put(in); err != nil {
				return done, err
			}
			done = append(done, in)
		}
	}
	return done, nil
}

// MemIntentStore is an in-memory IntentStore for tests.
type MemIntentStore struct {
	mu      sync.Mutex
	m       map[string]Intent
	FailPut error
}

func NewMemIntentStore() *MemIntentStore { return &MemIntentStore{m: map[string]Intent{}} }

func (s *MemIntentStore) Get(k string) (Intent, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	in, ok := s.m[k]
	return in, ok, nil
}

func (s *MemIntentStore) Put(in Intent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.FailPut != nil {
		return s.FailPut
	}
	s.m[in.Key] = in
	return nil
}

func (s *MemIntentStore) List() ([]Intent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Intent, 0, len(s.m))
	for _, in := range s.m {
		out = append(out, in)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}
