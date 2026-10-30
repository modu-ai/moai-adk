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

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
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
	LabelApply      Kind = "label-apply"
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
	Digest     string `json:"digest"`                // digest of the request payload
	Payload    string `json:"payload,omitempty"`     // already-masked text to send, when the effect carries one
	Seq        int    `json:"seq,omitempty"`         // order among intents for one target
	NotBefore  int64  `json:"not_before,omitempty"`  // unix seconds; no send before this (rate limit)
	ParkReason string `json:"park_reason,omitempty"` // why the intent is parked when the cause is a refusal
	Status     Status `json:"status"`
	Attempts   int    `json:"attempts"`
	// SentAt is when the executor stamped the intent Sent, in unix nanoseconds (0 for a record written
	// before the field). It is stamped just before the compare-and-swap that grants the right to send, so a
	// reconciler's window starts a little before the send does; the window's second timeout covers that.
	SentAt int64 `json:"sent_at,omitempty"`
	// Rev is the revision the store held when the intent was read (0: not stored yet). It is never written
	// into the stored intent; PutIf uses it as the revision the write expects.
	Rev        int64  `json:"-"`
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

// ErrDenied marks a permission or authentication refusal: the effect did not happen, retrying cannot
// help, so the intent is parked with no lookup and no resend.
var ErrDenied = errors.New("effect: permission denied")

// LimitedError marks a rate limit: the effect did not happen and may be retried at RetryAt, never before.
type LimitedError struct{ RetryAt time.Time }

func (e *LimitedError) Error() string {
	return "effect: rate limited until " + e.RetryAt.UTC().Format(time.RFC3339)
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

// IntentStore persists intents. effect/docstore keeps them in the claim store's "intent" documents, which
// is where the authority (the online control ref, when attached) comes first. The collector's consumer
// ACK is forge.Collector.Ack and its cursor is forge.CursorStore, kept by effect/docstore in the claim
// store's "collector" documents. Retry-After /
// X-RateLimit-* backoff, conditional GET with ETag and the shared call budget are in package forge; this
// executor honours them through LimitedError and ErrDenied.
type IntentStore interface {
	// Get returns the stored intent with Rev set to the revision it was read at.
	Get(key string) (Intent, bool, error)
	// PutIf writes the intent only if the stored revision is still in.Rev (0: the intent must not be stored
	// yet) and returns the stored intent with its new Rev. A store that moved on is
	// core.ErrRevisionConflict and nothing is written. The Executor uses it so that two executors never both
	// take the right to send.
	//
	// There is deliberately no unconditional Put in this interface: a write that ignores the revision is
	// the double-send window itself. The concrete stores keep a Put for seeding fixtures only.
	PutIf(Intent) (Intent, error)
	List() ([]Intent, error)
}

// Timeouts bounds each stage.
type Timeouts struct {
	IssueCreate, BranchPrepare, WorktreePrepare, CheckObserve, LedgerComment, LabelApply time.Duration
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
	case LabelApply:
		return t.LabelApply
	}
	return 0
}

// Executor applies the rules above.
type Executor struct {
	Store    IntentStore
	Send     Sender
	Lookup   Lookuper
	Timeouts Timeouts
	Now      func() time.Time // defaults to time.Now
}

func (e *Executor) now() time.Time {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now()
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
		if in.NotBefore > e.now().Unix() {
			return in, nil // rate limited: not yet
		}
	} else {
		in = Intent{Key: key, Kind: r.Kind, Issue: r.Issue, Target: r.Target, Digest: r.Digest, Payload: r.Payload, Status: StatusIntended}
		created, err := e.Store.PutIf(in) // create-only: Rev 0
		if errors.Is(err, core.ErrRevisionConflict) {
			return e.follow(ctx, key)
		}
		if err != nil {
			return Intent{}, err
		}
		in = created
	}
	in.Attempts++
	in.Status = StatusSent
	in.SentAt = e.now().UnixNano()
	in.LastError = ""
	in.NotBefore = 0
	// Winning this write is the right to send: of several executors that read the same state, one revision
	// matches and the rest follow what is stored. Nothing is sent before it.
	won, err := e.Store.PutIf(in)
	if errors.Is(err, core.ErrRevisionConflict) {
		return e.follow(ctx, key)
	}
	if err != nil {
		return Intent{}, err
	}
	in = won
	sctx, cancel := context.WithTimeout(ctx, to)
	id, serr := e.Send.Send(sctx, in)
	cancel()
	switch {
	case serr == nil:
		in.Status, in.ProviderID = StatusConfirmed, id
	case errors.Is(serr, ErrRefused):
		in.Status, in.LastError = StatusParked, serr.Error()
		var re *RefusedError
		if errors.As(serr, &re) {
			in.ParkReason = string(re.Reason)
		}
	case errors.Is(serr, ErrDenied):
		in.Status, in.LastError = StatusParked, serr.Error()
	case errors.As(serr, new(*LimitedError)):
		var le *LimitedError
		errors.As(serr, &le)
		in.Status, in.LastError, in.NotBefore = StatusFailed, serr.Error(), le.RetryAt.Unix()
	case errors.Is(serr, ErrNotSent):
		in.Status, in.LastError = StatusFailed, serr.Error()
	default:
		in.LastError = serr.Error()
		if _, err := e.settle(in); err != nil {
			return Intent{}, err
		}
		return e.Reconcile(ctx, key)
	}
	return e.settle(in)
}

// follow returns what is stored after losing a compare-and-swap. A sent intent is reconciled from facts;
// anything else is returned as stored. It never sends: the executor that won the right to send owns it.
func (e *Executor) follow(ctx context.Context, key string) (Intent, error) {
	cur, ok, err := e.Store.Get(key)
	if err != nil {
		return Intent{}, err
	}
	if !ok {
		return Intent{}, fmt.Errorf("effect: intent %s vanished after a write conflict", key)
	}
	if cur.Status == StatusSent {
		return e.Reconcile(ctx, key)
	}
	return cur, nil
}

// settle writes the outcome of an attempt this executor owns. When the stored intent moved meanwhile (a
// reconciler resolved it, say) a confirmed outcome already stored stands; otherwise our outcome is written
// over the fresh revision once — the provider id we hold is a fact the stored copy may lack.
func (e *Executor) settle(in Intent) (Intent, error) {
	out, err := e.Store.PutIf(in)
	if !errors.Is(err, core.ErrRevisionConflict) {
		return out, err
	}
	cur, ok, gerr := e.Store.Get(in.Key)
	if gerr != nil {
		return Intent{}, gerr
	}
	if ok && cur.Status == StatusConfirmed {
		return cur, nil
	}
	if ok {
		in.Rev = cur.Rev
	}
	return e.Store.PutIf(in)
}

// Reconcile resolves a sent or parked intent from facts. It never sends.
//
// An authoritative "nothing there" is not read as failure while a send may still be in flight — another
// executor's, or this one's whose outcome came back unknown (a timeout, a reset connection: the request
// may still land). The intent stays Sent, and a retry therefore cannot duplicate it, until SentAt plus two
// of the kind's configured timeouts have passed: one for the send's own limit, one for the write that
// stamped SentAt and for delivery lag. No new constant: the window is derived from the same timeout the
// sender runs under.
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
		if in.Status == StatusSent && in.SentAt > 0 && e.now().UnixNano() < in.SentAt+2*int64(to) {
			return in, nil // the sender may still be mid-flight: look again later, write nothing
		}
		in.Status = StatusFailed
	default:
		in.Status = StatusParked
	}
	out, err := e.Store.PutIf(in)
	if errors.Is(err, core.ErrRevisionConflict) {
		// Another party resolved it while we looked; what is stored stands.
		cur, ok, gerr := e.Store.Get(key)
		if gerr != nil || !ok {
			return Intent{}, fmt.Errorf("effect: intent %s unreadable after a write conflict: %v", key, gerr)
		}
		return cur, nil
	}
	return out, err
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
	out, err := e.Store.PutIf(in)
	if errors.Is(err, core.ErrRevisionConflict) { // recorded by someone else meanwhile: idempotent
		cur, _, gerr := e.Store.Get(key)
		return cur, gerr
	}
	return out, err
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
			if _, err := e.Store.PutIf(in); err != nil && !errors.Is(err, core.ErrRevisionConflict) {
				return done, err
			}
			continue
		}
		if res.Found {
			in.Status, in.ProviderID = StatusConfirmed, res.ProviderID
			out, err := e.Store.PutIf(in)
			if errors.Is(err, core.ErrRevisionConflict) {
				continue // resolved by someone else; the next tick sees the stored state
			}
			if err != nil {
				return done, err
			}
			done = append(done, out)
		}
	}
	return done, nil
}

// MemIntentStore is an in-memory IntentStore for tests.
type MemIntentStore struct {
	mu      sync.Mutex
	m       map[string]Intent
	rev     map[string]int64
	next    int64
	FailPut error
}

func NewMemIntentStore() *MemIntentStore {
	return &MemIntentStore{m: map[string]Intent{}, rev: map[string]int64{}}
}

func (s *MemIntentStore) Get(k string) (Intent, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	in, ok := s.m[k]
	in.Rev = s.rev[k]
	return in, ok, nil
}

func (s *MemIntentStore) Put(in Intent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.FailPut != nil {
		return s.FailPut
	}
	s.store(in)
	return nil
}

func (s *MemIntentStore) PutIf(in Intent) (Intent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.FailPut != nil {
		return Intent{}, s.FailPut
	}
	if s.rev[in.Key] != in.Rev {
		return Intent{}, fmt.Errorf("%w: intent %s is at revision %d, not %d", core.ErrRevisionConflict, in.Key, s.rev[in.Key], in.Rev)
	}
	return s.store(in), nil
}

// store writes the intent under a fresh revision; the caller holds the lock.
func (s *MemIntentStore) store(in Intent) Intent {
	s.next++
	in.Rev = s.next
	s.rev[in.Key] = in.Rev
	s.m[in.Key] = in
	return in
}

func (s *MemIntentStore) List() ([]Intent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Intent, 0, len(s.m))
	for k, in := range s.m {
		in.Rev = s.rev[k]
		out = append(out, in)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}
