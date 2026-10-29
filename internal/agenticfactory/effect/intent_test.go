package effect

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeSender struct {
	calls int
	id    string
	err   error
	block bool
}

func (f *fakeSender) Send(ctx context.Context, in Intent) (string, error) {
	f.calls++
	if f.block {
		<-ctx.Done()
		return "", ctx.Err()
	}
	return f.id, f.err
}

type fakeLookup struct {
	calls         int
	found         bool
	id            string
	authoritative bool
	err           error
}

func (f *fakeLookup) Lookup(context.Context, Intent) (LookupResult, error) {
	f.calls++
	return LookupResult{Found: f.found, ProviderID: f.id, Authoritative: f.authoritative}, f.err
}

var timeouts = Timeouts{IssueCreate: time.Second, BranchPrepare: time.Second, WorktreePrepare: time.Second, CheckObserve: time.Second}

func newExec(s Sender, l Lookuper, st IntentStore) *Executor {
	return &Executor{Store: st, Send: s, Lookup: l, Timeouts: timeouts}
}

func req() Request {
	return Request{Kind: IssueCreate, Issue: 7, Target: "repo", Digest: "d1"}
}

func TestKeyIsDeterministic(t *testing.T) {
	a := Key(IssueCreate, "repo", "d1")
	if a != Key(IssueCreate, "repo", "d1") || a == Key(IssueCreate, "repo", "d2") || len(a) < 16 {
		t.Errorf("key = %q", a)
	}
}

func TestIntentIsRecordedBeforeSend(t *testing.T) {
	st := NewMemIntentStore()
	var seenAtSend Status
	s := &fakeSender{id: "n1"}
	ex := newExec(sendFn(func(ctx context.Context, in Intent) (string, error) {
		got, _, _ := st.Get(in.Key)
		seenAtSend = got.Status
		return s.Send(ctx, in)
	}), &fakeLookup{}, st)
	if _, err := ex.Execute(t.Context(), req()); err != nil {
		t.Fatal(err)
	}
	if seenAtSend != StatusSent {
		t.Errorf("at send time the durable status was %q, want sent", seenAtSend)
	}
}

type sendFn func(context.Context, Intent) (string, error)

func (f sendFn) Send(ctx context.Context, in Intent) (string, error) { return f(ctx, in) }

func TestIntentWriteFailureMeansNoSend(t *testing.T) {
	st := NewMemIntentStore()
	st.FailPut = errors.New("disk full")
	s := &fakeSender{}
	if _, err := newExec(s, &fakeLookup{}, st).Execute(t.Context(), req()); err == nil {
		t.Fatal("expected error")
	}
	if s.calls != 0 {
		t.Errorf("sent %d times without a durable intent", s.calls)
	}
}

func TestConfirmedKeyIsNeverSentAgain(t *testing.T) {
	st := NewMemIntentStore()
	s := &fakeSender{id: "n1"}
	ex := newExec(s, &fakeLookup{}, st)
	in, _ := ex.Execute(t.Context(), req())
	if in.Status != StatusConfirmed || in.ProviderID != "n1" {
		t.Fatalf("first = %+v", in)
	}
	again, err := ex.Execute(t.Context(), req())
	if err != nil || again.Status != StatusConfirmed || s.calls != 1 {
		t.Errorf("second: %+v err=%v calls=%d", again, err, s.calls)
	}
}

func TestAmbiguousSendIsReconciledNeverResent(t *testing.T) {
	cases := []struct {
		name string
		lk   *fakeLookup
		want Status
	}{
		{"found remotely", &fakeLookup{found: true, id: "n9", authoritative: true}, StatusConfirmed},
		{"authoritatively absent", &fakeLookup{authoritative: true}, StatusFailed},
		{"lookup unsure", &fakeLookup{authoritative: false}, StatusParked},
		{"lookup error", &fakeLookup{err: errors.New("gh 502")}, StatusParked},
	}
	for _, c := range cases {
		s := &fakeSender{err: errors.New("connection reset")}
		in, _ := newExec(s, c.lk, NewMemIntentStore()).Execute(t.Context(), req())
		if in.Status != c.want {
			t.Errorf("%s: status = %s, want %s", c.name, in.Status, c.want)
		}
		if s.calls != 1 || c.lk.calls != 1 {
			t.Errorf("%s: sends=%d lookups=%d, want 1/1", c.name, s.calls, c.lk.calls)
		}
	}
}

func TestParkedIsNotSentAgainUntilReconciled(t *testing.T) {
	st := NewMemIntentStore()
	s := &fakeSender{err: errors.New("lost")}
	lk := &fakeLookup{authoritative: false}
	ex := newExec(s, lk, st)
	if in, _ := ex.Execute(t.Context(), req()); in.Status != StatusParked {
		t.Fatalf("status = %s", in.Status)
	}
	if in, _ := ex.Execute(t.Context(), req()); in.Status != StatusParked || s.calls != 1 {
		t.Errorf("parked intent was sent again: calls=%d status=%s", s.calls, in.Status)
	}
	lk.found, lk.id, lk.authoritative = true, "n5", true
	in, err := ex.Reconcile(t.Context(), Key(IssueCreate, "repo", "d1"))
	if err != nil || in.Status != StatusConfirmed || s.calls != 1 {
		t.Errorf("after reconcile: %+v err=%v calls=%d", in, err, s.calls)
	}
}

func TestNotSentErrorIsRetryableWithoutLookup(t *testing.T) {
	s := &fakeSender{err: ErrNotSent}
	lk := &fakeLookup{}
	ex := newExec(s, lk, NewMemIntentStore())
	in, _ := ex.Execute(t.Context(), req())
	if in.Status != StatusFailed || in.Attempts != 1 || lk.calls != 0 {
		t.Fatalf("in=%+v lookups=%d", in, lk.calls)
	}
	s.err, s.id = nil, "n2"
	in, _ = ex.Execute(t.Context(), req())
	if in.Status != StatusConfirmed || in.Attempts != 2 {
		t.Errorf("retry: %+v", in)
	}
}

func TestStageTimeoutIsAmbiguousAndReconciled(t *testing.T) {
	ex := newExec(&fakeSender{block: true}, &fakeLookup{authoritative: true}, NewMemIntentStore())
	ex.Timeouts.IssueCreate = 20 * time.Millisecond
	start := time.Now()
	in, _ := ex.Execute(t.Context(), req())
	if in.Status != StatusFailed || time.Since(start) > time.Second {
		t.Errorf("status=%s elapsed=%v", in.Status, time.Since(start))
	}
	ex.Timeouts.IssueCreate = 0
	if _, err := ex.Execute(t.Context(), Request{Kind: IssueCreate, Issue: 7, Target: "x", Digest: "y"}); err == nil {
		t.Error("a stage without a timeout must be refused")
	}
}

func TestRestartReconcilesSentIntentFromStore(t *testing.T) {
	st := NewMemIntentStore()
	k := Key(IssueCreate, "repo", "d1")
	_ = st.Put(Intent{Key: k, Kind: IssueCreate, Issue: 7, Target: "repo", Digest: "d1", Status: StatusSent})
	s := &fakeSender{}
	ex := newExec(s, &fakeLookup{found: true, id: "n7", authoritative: true}, st) // fresh executor after restart
	in, err := ex.Execute(t.Context(), req())
	if err != nil || in.Status != StatusConfirmed || in.ProviderID != "n7" || s.calls != 0 {
		t.Errorf("in=%+v err=%v sends=%d", in, err, s.calls)
	}
}

func TestChecksObservationIsAsynchronous(t *testing.T) {
	st := NewMemIntentStore()
	lk := &fakeLookup{found: true, id: "success", authoritative: true}
	ex := newExec(&fakeSender{}, lk, st)
	in, err := ex.RequestObservation(t.Context(), Request{Kind: CheckObserve, Issue: 7, Target: "pr-55", Digest: "h1"})
	if err != nil || in.Status != StatusIntended || lk.calls != 0 {
		t.Fatalf("request must only record: %+v err=%v lookups=%d", in, err, lk.calls)
	}
	done, err := ex.Tick(t.Context())
	if err != nil || len(done) != 1 || done[0].Status != StatusConfirmed || lk.calls != 1 {
		t.Errorf("tick: %+v err=%v lookups=%d", done, err, lk.calls)
	}
	if got, _, _ := st.Get(in.Key); got.Status != StatusConfirmed {
		t.Errorf("stored = %+v", got)
	}
}

func TestKeyMarkerRoundTrips(t *testing.T) {
	m := Marker("abc123")
	if !strings.Contains(m, "abc123") || KeyFromMarker("body text\n"+m+"\nmore") != "abc123" || KeyFromMarker("none") != "" {
		t.Errorf("marker = %q", m)
	}
}
