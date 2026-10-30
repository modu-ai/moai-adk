package effect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// The right to send is the compare-and-swap of Intended/Failed -> Sent: of two executors that read the same
// stored state, exactly one wins it, and the other follows what is stored and never sends.

type countSender struct {
	n      atomic.Int32
	id     string
	onSend func(Intent)
}

func (s *countSender) Send(_ context.Context, in Intent) (string, error) {
	s.n.Add(1)
	if s.onSend != nil {
		s.onSend(in)
	}
	return s.id, nil
}

type foundLookup struct{ id string }

func (l foundLookup) Lookup(context.Context, Intent) (LookupResult, error) {
	return LookupResult{Found: true, ProviderID: l.id, Authoritative: true}, nil
}

// hookStore runs a function after every Get, so a test can hold an executor between its read and its write.
type hookStore struct {
	IntentStore
	afterGet func(found bool, in Intent)
}

func (h hookStore) Get(k string) (Intent, bool, error) {
	in, ok, err := h.IntentStore.Get(k)
	if h.afterGet != nil {
		h.afterGet(ok, in)
	}
	return in, ok, err
}

func casTimeouts() Timeouts {
	d := time.Second
	return Timeouts{IssueCreate: d, BranchPrepare: d, WorktreePrepare: d, CheckObserve: d, LedgerComment: d, LabelApply: d}
}

func casReq() Request {
	return Request{Kind: IssueCreate, Issue: 7, Target: "repo:o/r", Digest: "d1", Payload: `{"title":"t"}`}
}

// The audit's scenario: B read "nothing stored", is held, A runs the whole effect, then B writes. B must see
// that the intent exists, follow it, and send nothing.
func TestAnExecutorThatReadNothingDoesNotSendAfterAnotherFinished(t *testing.T) {
	mem := NewMemIntentStore()
	sa, sb := &countSender{id: "issue-9"}, &countSender{id: "issue-10"}
	reached, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	b := &Executor{Store: hookStore{IntentStore: mem, afterGet: func(found bool, _ Intent) {
		if !found {
			once.Do(func() { close(reached); <-release })
		}
	}}, Send: sb, Lookup: foundLookup{"issue-10"}, Timeouts: casTimeouts()}
	a := &Executor{Store: mem, Send: sa, Lookup: foundLookup{"issue-9"}, Timeouts: casTimeouts()}

	type result struct {
		in  Intent
		err error
	}
	done := make(chan result, 1)
	go func() { in, err := b.Execute(t.Context(), casReq()); done <- result{in, err} }()
	<-reached
	ain, err := a.Execute(t.Context(), casReq())
	if err != nil || ain.Status != StatusConfirmed || ain.ProviderID != "issue-9" {
		t.Fatalf("A: %+v %v", ain, err)
	}
	close(release)
	var r result
	select {
	case r = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("B did not return")
	}
	if sa.n.Load()+sb.n.Load() != 1 {
		t.Fatalf("sends: A=%d B=%d, want exactly one send of the same key", sa.n.Load(), sb.n.Load())
	}
	if r.err != nil || r.in.Status != StatusConfirmed || r.in.ProviderID != "issue-9" {
		t.Errorf("B must follow the stored state, not overwrite it: %+v %v", r.in, r.err)
	}
	if got, _, _ := mem.Get(ain.Key); got.ProviderID != "issue-9" || got.Attempts != 1 {
		t.Errorf("the stored intent changed: %+v", got)
	}
}

// Both read the same Failed intent (a retry is allowed) and both try to take the right to send.
func TestOfTwoExecutorsRetryingTheSameIntentExactlyOneSends(t *testing.T) {
	mem := NewMemIntentStore()
	key := Key(IssueCreate, "repo:o/r", "d1")
	if err := mem.Put(Intent{Key: key, Kind: IssueCreate, Issue: 7, Target: "repo:o/r", Digest: "d1", Status: StatusFailed, Attempts: 1}); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	// Each executor waits once, on its first read; neither proceeds until both have read.
	barrier := func() func(bool, Intent) {
		var once sync.Once
		return func(bool, Intent) { once.Do(func() { wg.Done(); wg.Wait() }) }
	}
	sa, sb := &countSender{id: "issue-9"}, &countSender{id: "issue-9"}
	ea := &Executor{Store: hookStore{mem, barrier()}, Send: sa, Lookup: foundLookup{"issue-9"}, Timeouts: casTimeouts()}
	eb := &Executor{Store: hookStore{mem, barrier()}, Send: sb, Lookup: foundLookup{"issue-9"}, Timeouts: casTimeouts()}
	var ra, rb Intent
	var ea1, eb1 error
	var g sync.WaitGroup
	g.Add(2)
	go func() { defer g.Done(); ra, ea1 = ea.Execute(t.Context(), casReq()) }()
	go func() { defer g.Done(); rb, eb1 = eb.Execute(t.Context(), casReq()) }()
	g.Wait()
	if sa.n.Load()+sb.n.Load() != 1 {
		t.Fatalf("sends: %d + %d, want exactly one", sa.n.Load(), sb.n.Load())
	}
	for name, r := range map[string]struct {
		in  Intent
		err error
	}{"A": {ra, ea1}, "B": {rb, eb1}} {
		if r.err != nil || r.in.Status != StatusConfirmed {
			t.Errorf("%s: %+v %v, want the effect confirmed whichever executor sent it", name, r.in, r.err)
		}
	}
}

// Another party resolves the intent while our send is in flight (a reconciler's lookup, say). Our send
// succeeded and we hold the provider's id, so the write that ends our attempt stands.
func TestAConflictOnTheFinalWriteKeepsTheSendersFact(t *testing.T) {
	mem := NewMemIntentStore()
	s := &countSender{id: "issue-9"}
	s.onSend = func(in Intent) {
		cur, _, _ := mem.Get(in.Key)
		cur.Status, cur.LastError = StatusParked, "someone else parked it"
		if _, err := mem.PutIf(cur); err != nil {
			t.Errorf("the interfering write failed: %v", err)
		}
	}
	e := &Executor{Store: mem, Send: s, Lookup: foundLookup{"issue-9"}, Timeouts: casTimeouts()}
	in, err := e.Execute(t.Context(), casReq())
	if err != nil || in.Status != StatusConfirmed || in.ProviderID != "issue-9" || s.n.Load() != 1 {
		t.Errorf("%+v %v sends=%d", in, err, s.n.Load())
	}
	if got, _, _ := mem.Get(in.Key); got.Status != StatusConfirmed || got.ProviderID != "issue-9" {
		t.Errorf("stored: %+v", got)
	}
}

// MemIntentStore follows the PutIf contract.
func TestMemIntentStoreRevisionsAndCompareAndSwap(t *testing.T) {
	m := NewMemIntentStore()
	a := Intent{Key: "k", Kind: IssueCreate, Status: StatusIntended}
	created, err := m.PutIf(a) // Rev 0: must not exist
	if err != nil || created.Rev <= 0 {
		t.Fatalf("create: %+v %v", created, err)
	}
	if _, err := m.PutIf(a); !errors.Is(err, core.ErrRevisionConflict) { // still Rev 0: it exists now
		t.Errorf("a second create-only write: %v, want a conflict", err)
	}
	got, ok, _ := m.Get("k")
	if !ok || got.Rev != created.Rev {
		t.Fatalf("Get carries the revision: %+v want %d", got, created.Rev)
	}
	got.Status = StatusSent
	next, err := m.PutIf(got)
	if err != nil || next.Rev <= got.Rev {
		t.Fatalf("update: %+v %v", next, err)
	}
	if _, err := m.PutIf(got); !errors.Is(err, core.ErrRevisionConflict) { // got.Rev is stale now
		t.Errorf("a write on a stale revision: %v, want a conflict", err)
	}
	if latest, _, _ := m.Get("k"); latest.Status != StatusSent || latest.Rev != next.Rev {
		t.Errorf("a conflicting write changed the stored intent: %+v", latest)
	}
	if err := m.Put(Intent{Key: "k", Status: StatusConfirmed}); err != nil { // the unconditional write moves the revision too
		t.Fatal(err)
	}
	if _, err := m.PutIf(next); !errors.Is(err, core.ErrRevisionConflict) {
		t.Errorf("after an unconditional Put the old revision is stale: %v", err)
	}
	raw, _ := json.Marshal(Intent{Key: "k", Rev: 5})
	if strings.Contains(string(raw), "rev") {
		t.Errorf("the revision is stored in the intent: %s", raw)
	}
}

type notFoundLookup struct{}

func (notFoundLookup) Lookup(context.Context, Intent) (LookupResult, error) {
	return LookupResult{Found: false, Authoritative: true}, nil
}

// F1b (b): the executor that lost the right to send reconciles from facts, but the winner's send may not
// have reached the provider yet. An authoritative "nothing there" inside the send's own time limit is not
// evidence of failure: the loser leaves the intent Sent, so a third executor never sees Failed and resends.
func TestALoserDoesNotWriteFailedWhileTheWinnersSendIsInFlight(t *testing.T) {
	mem := NewMemIntentStore()
	now := time.Unix(1_000_000, 0)
	entered, release := make(chan struct{}), make(chan struct{})
	sa := &countSender{id: "issue-9", onSend: func(Intent) { close(entered); <-release }}
	a := &Executor{Store: mem, Send: sa, Lookup: foundLookup{"issue-9"}, Timeouts: casTimeouts(), Now: func() time.Time { return now }}
	done := make(chan struct{})
	go func() { defer close(done); _, _ = a.Execute(t.Context(), casReq()) }()
	<-entered

	sb, sc := &countSender{id: "issue-10"}, &countSender{id: "issue-11"}
	b := &Executor{Store: mem, Send: sb, Lookup: notFoundLookup{}, Timeouts: casTimeouts(), Now: func() time.Time { return now.Add(time.Millisecond) }}
	c := &Executor{Store: mem, Send: sc, Lookup: notFoundLookup{}, Timeouts: casTimeouts(), Now: func() time.Time { return now.Add(2 * time.Millisecond) }}
	got, err := b.Execute(t.Context(), casReq())
	if err != nil || got.Status != StatusSent {
		t.Fatalf("B: %+v %v, want the intent left Sent while A's send is in flight", got, err)
	}
	if got, err := c.Execute(t.Context(), casReq()); err != nil || got.Status != StatusSent {
		t.Fatalf("C: %+v %v, want Sent", got, err)
	}
	close(release)
	<-done
	if n := sa.n.Load() + sb.n.Load() + sc.n.Load(); n != 1 {
		t.Fatalf("sends = %d, want exactly one", n)
	}
	if st, _, _ := mem.Get(Key(IssueCreate, "repo:o/r", "d1")); st.Status != StatusConfirmed {
		t.Errorf("stored = %+v, want Confirmed", st)
	}
}

// Past the send's time limit the same reading is evidence: the intent is Failed and may be retried.
func TestAnAuthoritativeNothingAfterTheSendLimitIsFailed(t *testing.T) {
	mem := NewMemIntentStore()
	now := time.Unix(1_000_000, 0)
	key := Key(IssueCreate, "repo:o/r", "d1")
	if err := mem.Put(Intent{Key: key, Kind: IssueCreate, Issue: 7, Target: "repo:o/r", Digest: "d1", Status: StatusSent, Attempts: 1, SentAt: now.UnixNano()}); err != nil {
		t.Fatal(err)
	}
	e := &Executor{Store: mem, Send: &countSender{}, Lookup: notFoundLookup{}, Timeouts: casTimeouts(), Now: func() time.Time { return now.Add(3 * time.Second) }}
	got, err := e.Reconcile(t.Context(), key)
	if err != nil || got.Status != StatusFailed {
		t.Errorf("%+v %v, want Failed once the send limit has passed", got, err)
	}
}

type failingSender struct{ n atomic.Int32 }

func (f *failingSender) Send(context.Context, Intent) (string, error) {
	f.n.Add(1)
	return "", errors.New("connection reset") // the outcome is unknown: the request may still land
}

// F1b P2-1: an unknown send outcome is not failure inside the window either. The same executor's own
// reconcile leaves the intent Sent, so a retry cannot create a duplicate; past the window it is Failed.
func TestAnUnknownSendOutcomeStaysSentInsideTheWindow(t *testing.T) {
	mem := NewMemIntentStore()
	now := time.Unix(1_000_000, 0)
	clock := now
	fs := &failingSender{}
	e := &Executor{Store: mem, Send: fs, Lookup: notFoundLookup{}, Timeouts: casTimeouts(), Now: func() time.Time { return clock }}
	got, err := e.Execute(t.Context(), casReq())
	if err != nil || got.Status != StatusSent {
		t.Fatalf("%+v %v, want Sent: the outcome is unknown", got, err)
	}
	if again, err := e.Execute(t.Context(), casReq()); err != nil || again.Status != StatusSent || fs.n.Load() != 1 {
		t.Fatalf("retry inside the window: %+v %v sends=%d, want Sent and no second send", again, err, fs.n.Load())
	}
	clock = now.Add(10 * time.Second)
	if late, err := e.Reconcile(t.Context(), got.Key); err != nil || late.Status != StatusFailed {
		t.Errorf("past the window: %+v %v, want Failed", late, err)
	}
}

// F1b P2-2: the window is kept in nanoseconds, so a timeout under a second is honoured.
func TestASubSecondTimeoutHasAWindow(t *testing.T) {
	mem := NewMemIntentStore()
	now := time.Unix(1_000_000, 0)
	key := Key(IssueCreate, "repo:o/r", "d1")
	if err := mem.Put(Intent{Key: key, Kind: IssueCreate, Issue: 7, Target: "repo:o/r", Digest: "d1", Status: StatusSent, Attempts: 1, SentAt: now.UnixNano()}); err != nil {
		t.Fatal(err)
	}
	short := Timeouts{IssueCreate: 100 * time.Millisecond}
	clock := now.Add(50 * time.Millisecond)
	e := &Executor{Store: mem, Send: &countSender{}, Lookup: notFoundLookup{}, Timeouts: short, Now: func() time.Time { return clock }}
	if got, _ := e.Reconcile(t.Context(), key); got.Status != StatusSent {
		t.Errorf("inside the window: %+v, want Sent", got)
	}
	clock = now.Add(time.Second)
	if got, _ := e.Reconcile(t.Context(), key); got.Status != StatusFailed {
		t.Errorf("past the window: %+v, want Failed", got)
	}
}

type errSender struct{ err error }

func (e errSender) Send(context.Context, Intent) (string, error) { return "", e.err }

type errLookup struct{ err error }

func (e errLookup) Lookup(context.Context, Intent) (LookupResult, error) {
	return LookupResult{}, e.err
}

// F8: an error text can carry a credential (a failing gh call echoing its URL or header). It is stored
// through the one sanitizer, whichever path recorded it.
func TestAStoredLastErrorNeverHoldsASecretOrPersonalData(t *testing.T) {
	secret := "ghp_abcdefghijklmnopqrstuvwxyz0123456789"
	cases := map[string]*Executor{
		"not sent":     {Send: errSender{fmt.Errorf("%w: POST https://x:%s@github.com failed", ErrNotSent, secret)}, Lookup: notFoundLookup{}},
		"denied":       {Send: errSender{fmt.Errorf("%w: token %s someone@example.com", ErrDenied, secret)}, Lookup: notFoundLookup{}},
		"unknown":      {Send: errSender{fmt.Errorf("reset while using %s at /Users/someone/work", secret)}, Lookup: notFoundLookup{}},
		"lookup error": {Send: errSender{errors.New("reset")}, Lookup: errLookup{fmt.Errorf("gh said Authorization: Bearer %s", secret)}},
	}
	for name, e := range cases {
		e.Store, e.Timeouts = NewMemIntentStore(), casTimeouts()
		in, _ := e.Execute(t.Context(), casReq())
		stored, _, _ := e.Store.Get(in.Key)
		for _, bad := range []string{secret, "someone@example.com", "/Users/someone"} {
			if strings.Contains(stored.LastError, bad) || strings.Contains(in.LastError, bad) {
				t.Errorf("%s: LastError holds %q: %q", name, bad, stored.LastError)
			}
		}
		if stored.LastError == "" {
			t.Errorf("%s: LastError is empty, the failure left no trace", name)
		}
	}
}
