package resident

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// (a) The six bounds on external effects: all set, with the values the leader decided for OD-20.
func TestEveryExternalEffectHasItsDecidedTimeout(t *testing.T) {
	r, _ := build(t)
	got := r.Timeouts()
	for name, tc := range map[string]struct{ got, want time.Duration }{
		"IssueCreate":     {got.IssueCreate, 30 * time.Second},
		"LedgerComment":   {got.LedgerComment, 30 * time.Second},
		"LabelApply":      {got.LabelApply, 30 * time.Second},
		"CheckObserve":    {got.CheckObserve, 10 * time.Second},
		"BranchPrepare":   {got.BranchPrepare, 60 * time.Second},
		"WorktreePrepare": {got.WorktreePrepare, 120 * time.Second},
	} {
		if tc.got <= 0 {
			t.Errorf("%s has no timeout", name)
		}
		if tc.got != tc.want {
			t.Errorf("%s = %v, want %v", name, tc.got, tc.want)
		}
	}
	for _, k := range []effect.Kind{effect.IssueCreate, effect.BranchPrepare, effect.WorktreePrepare, effect.CheckObserve, effect.LedgerComment, effect.LabelApply} {
		if got.For(k) <= 0 {
			t.Errorf("kind %s has no timeout", k)
		}
	}
	if l := r.Ledger(nil, nil); l.Timeouts != got {
		t.Errorf("the ledger carries %+v, not the wired timeouts", l.Timeouts)
	}
}

type blockingSender struct {
	mu    sync.Mutex
	sends int
}

// Send blocks until its context ends - the case the timeout exists for.
func (b *blockingSender) Send(ctx context.Context, _ effect.Intent) (string, error) {
	b.mu.Lock()
	b.sends++
	b.mu.Unlock()
	<-ctx.Done()
	return "", ctx.Err()
}

func (b *blockingSender) count() int { b.mu.Lock(); defer b.mu.Unlock(); return b.sends }

type fixedLookup struct {
	res   effect.LookupResult
	err   error
	calls int
}

func (f *fixedLookup) Lookup(context.Context, effect.Intent) (effect.LookupResult, error) {
	f.calls++
	return f.res, f.err
}

func issueRequest() effect.Request {
	return effect.Request{Kind: effect.IssueCreate, Issue: 7, Target: "repo:o/a", Digest: "d1", Payload: `{"title":"t"}`}
}

// quick makes the bounds short enough to wait for; the contract under test does not depend on the length.
func quick(r *Resident) {
	r.timeouts = effect.Timeouts{IssueCreate: 30 * time.Millisecond, BranchPrepare: 30 * time.Millisecond, WorktreePrepare: 30 * time.Millisecond,
		CheckObserve: 30 * time.Millisecond, LedgerComment: 30 * time.Millisecond, LabelApply: 30 * time.Millisecond}
}

// (b) A send that never answers expires: the intent was sent and may have been delivered, so it is
// reconciled by looking, never sent again - and what the look finds is what it becomes.
func TestAnExpiredSendIsReconciledByLookupAndNeverSentTwice(t *testing.T) {
	for name, tc := range map[string]struct {
		res    effect.LookupResult
		err    error
		status effect.Status
	}{
		"found remotely":           {effect.LookupResult{Found: true, ProviderID: "n9", Authoritative: true}, nil, effect.StatusConfirmed},
		"unsure":                   {effect.LookupResult{Authoritative: false}, nil, effect.StatusParked},
		"the lookup itself failed": {effect.LookupResult{}, errors.New("gh 502"), effect.StatusParked},
	} {
		r, _ := build(t)
		quick(r)
		send, look := &blockingSender{}, &fixedLookup{res: tc.res, err: tc.err}
		ex := r.Executor(send, look)
		in, err := ex.Execute(t.Context(), issueRequest())
		if err != nil || in.Status != tc.status {
			t.Fatalf("%s: status=%s err=%v, want %s", name, in.Status, err, tc.status)
		}
		if in.LastError == "" && tc.status == effect.StatusParked && tc.err != nil {
			t.Errorf("%s: the failure was not recorded", name)
		}
		if send.count() != 1 || look.calls != 1 {
			t.Errorf("%s: sends=%d lookups=%d", name, send.count(), look.calls)
		}
		// asking again must not send again
		again, err := ex.Execute(t.Context(), issueRequest())
		if err != nil || again.Status != tc.status || send.count() != 1 {
			t.Errorf("%s: second execute status=%s err=%v sends=%d", name, again.Status, err, send.count())
		}
	}
}

type recordingStore struct {
	effect.IntentStore
	statuses []effect.Status
}

func (s *recordingStore) Put(in effect.Intent) error {
	s.statuses = append(s.statuses, in.Status)
	return s.IntentStore.Put(in)
}

// The expiry leaves the intent recorded as Sent (with the error) before the lookup resolves it, so a
// crash between the two is also reconciled, not resent.
func TestTheExpiryIsRecordedAsSentBeforeTheLookupResolvesIt(t *testing.T) {
	r, _ := build(t)
	quick(r)
	ex := r.Executor(&blockingSender{}, &fixedLookup{res: effect.LookupResult{Found: true, ProviderID: "n9", Authoritative: true}})
	rec := &recordingStore{IntentStore: ex.Store}
	ex.Store = rec
	if _, err := ex.Execute(t.Context(), issueRequest()); err != nil {
		t.Fatal(err)
	}
	want := []effect.Status{effect.StatusIntended, effect.StatusSent, effect.StatusSent, effect.StatusConfirmed}
	if len(rec.statuses) != len(want) {
		t.Fatalf("statuses = %v, want %v", rec.statuses, want)
	}
	for i := range want {
		if rec.statuses[i] != want[i] {
			t.Fatalf("statuses = %v, want %v", rec.statuses, want)
		}
	}
}

// The contract that a missing bound is refused is kept: nothing is sent.
func TestAZeroTimeoutIsRefusedAndNothingIsSent(t *testing.T) {
	r, _ := build(t)
	r.timeouts = effect.Timeouts{}
	send := &blockingSender{}
	if _, err := r.Executor(send, &fixedLookup{}).Execute(t.Context(), issueRequest()); err == nil || send.count() != 0 {
		t.Errorf("err=%v sends=%d", err, send.count())
	}
}
