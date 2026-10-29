package e2e

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// Host-call efficiency and limit handling on lane-20's forge package. Every
// header value, wait and status below is a TEST INPUT chosen for the case; none
// of them is a statement about a real host's limits.

var testBackoff = forge.Backoff{MinSecondaryWait: 60 * time.Second, Base: 2 * time.Second, Max: 10 * time.Minute, MaxAttempts: 5}

// cappedBackoff keeps Max >= MinSecondaryWait but small enough to show the cap.
var cappedBackoff = forge.Backoff{MinSecondaryWait: time.Second, Base: 2 * time.Second, Max: 5 * time.Second, MaxAttempts: 5}

type scriptTransport struct {
	mu      sync.Mutex
	calls   []forge.Request
	replies []forge.Response
	err     error
}

func (s *scriptTransport) Do(_ context.Context, r forge.Request) (forge.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, r)
	if s.err != nil {
		return forge.Response{}, s.err
	}
	if len(s.replies) == 0 {
		return forge.Response{Status: 200}, nil
	}
	out := s.replies[0]
	if len(s.replies) > 1 {
		s.replies = s.replies[1:]
	}
	return out, nil
}

func (s *scriptTransport) n() int { s.mu.Lock(); defer s.mu.Unlock(); return len(s.calls) }

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

var t0 = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) // a fixed test instant

func newClient(tr forge.Transport, clk *clock, reserve int) *forge.Client {
	return &forge.Client{T: tr, B: &forge.Budget{Reserve: reserve}, Backoff: testBackoff, Now: clk.Now}
}

// Classify separates a limit (recoverable) from a permission refusal (never retried).
func TestForge_Classify_LimitVersusDenied(t *testing.T) {
	reset := t0.Add(90 * time.Second)
	cases := []struct {
		name    string
		resp    forge.Response
		attempt int
		want    forge.Kind
		retryAt time.Time // zero = not checked
	}{
		{"429 with Retry-After", forge.Response{Status: 429, Header: map[string]string{"Retry-After": "30"}}, 0, forge.KindLimited, t0.Add(30 * time.Second)},
		{"403 with Retry-After", forge.Response{Status: 403, Header: map[string]string{"retry-after": "45"}}, 0, forge.KindLimited, t0.Add(45 * time.Second)},
		{"403 exhausted with reset", forge.Response{Status: 403, Header: map[string]string{"x-ratelimit-remaining": "0", "x-ratelimit-reset": strconv.FormatInt(reset.Unix(), 10)}}, 0, forge.KindLimited, reset},
		{"403 with limit wording only", forge.Response{Status: 403, Body: []byte("API rate limit exceeded")}, 0, forge.KindLimited, t0.Add(testBackoff.MinSecondaryWait)},
		{"429 without any hint, attempt 2 doubles twice", forge.Response{Status: 429}, 2, forge.KindLimited, t0.Add(4 * testBackoff.MinSecondaryWait)},
		{"403 permission, no signal", forge.Response{Status: 403, Body: []byte("Resource not accessible by integration")}, 0, forge.KindDenied, time.Time{}},
		{"401", forge.Response{Status: 401}, 0, forge.KindDenied, time.Time{}},
		{"502", forge.Response{Status: 502}, 0, forge.KindTransient, t0.Add(testBackoff.Base)},
		{"304", forge.Response{Status: 304}, 0, forge.KindNotModified, time.Time{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := forge.Classify(tc.resp, t0, testBackoff, tc.attempt)
			if v.Kind != tc.want {
				t.Fatalf("kind = %s, want %s", v.Kind, tc.want)
			}
			if !tc.retryAt.IsZero() && !v.RetryAt.Equal(tc.retryAt) {
				t.Errorf("RetryAt = %v, want %v", v.RetryAt, tc.retryAt)
			}
		})
	}
}

// A refreshed path costs one conditional call and no body when nothing changed.
func TestForge_ConditionalGet_HitsReturnTheCachedBody(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{
		{Status: 200, Header: map[string]string{"etag": `"e1"`}, Body: []byte(`[{"id":"1"}]`)},
		{Status: 304},
	}}
	clk := &clock{now: t0}
	c := newClient(tr, clk, 0)
	first, err := c.Get(context.Background(), "/issues")
	if err != nil || !first.Changed {
		t.Fatalf("first get: %+v, %v", first, err)
	}
	clk.now = t0.Add(time.Minute)
	second, err := c.Get(context.Background(), "/issues")
	if err != nil || !second.FromCache || second.Changed || string(second.Body) != `[{"id":"1"}]` {
		t.Fatalf("second get: %+v, %v", second, err)
	}
	if got := tr.calls[1].Header["if-none-match"]; got != `"e1"` {
		t.Errorf("second request validator = %q, want the stored ETag", got)
	}
	t.Logf("MEASURED conditional get: calls=%d conditional-requests=1 hits(304)=1 hit-rate=100%% (2-call fixture)", tr.n())
}

// While a limit is in force no call is spent; the call resumes after the stated time.
func TestForge_Limited_DefersAndSpendsNoCalls(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{
		{Status: 429, Header: map[string]string{"retry-after": "30"}},
		{Status: 200, Body: []byte(`[]`)},
	}}
	clk := &clock{now: t0}
	c := newClient(tr, clk, 0)
	r1, _ := c.Get(context.Background(), "/p")
	if !r1.Deferred || !r1.NotBefore.Equal(t0.Add(30*time.Second)) {
		t.Fatalf("limited get: %+v", r1)
	}
	clk.now = t0.Add(10 * time.Second)
	for range 5 {
		if r, _ := c.Get(context.Background(), "/p"); !r.Deferred {
			t.Fatal("a call was allowed inside the limit window")
		}
	}
	if tr.n() != 1 {
		t.Errorf("transport calls during the window = %d, want only the first", tr.n())
	}
	clk.now = t0.Add(31 * time.Second)
	if r, err := c.Get(context.Background(), "/p"); err != nil || r.Deferred {
		t.Errorf("after the window: %+v, %v", r, err)
	}
	t.Logf("MEASURED limit window: calls=%d (1 limited + 1 after recovery), calls spent inside the window=0", tr.n())
}

// A permission refusal is reported once and spends no further calls by itself.
func TestForge_Denied_ReportedOnceNeverRetriedByTheClient(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{{Status: 403, Body: []byte("Resource not accessible")}}}
	c := newClient(tr, &clock{now: t0}, 0)
	_, err := c.Get(context.Background(), "/p")
	if !errors.Is(err, forge.ErrDenied) {
		t.Fatalf("err = %v, want ErrDenied", err)
	}
	if tr.n() != 1 {
		t.Errorf("transport calls = %d, want 1", tr.n())
	}
}

// The shared budget stops calls at the reserve until the host's reset time.
func TestForge_BudgetReserve_StopsBeforeReset(t *testing.T) {
	reset := t0.Add(100 * time.Second)
	tr := &scriptTransport{replies: []forge.Response{
		{Status: 200, Header: map[string]string{"x-ratelimit-remaining": "2", "x-ratelimit-reset": strconv.FormatInt(reset.Unix(), 10)}, Body: []byte(`[]`)},
	}}
	clk := &clock{now: t0}
	c := newClient(tr, clk, 2)
	if _, err := c.Get(context.Background(), "/p"); err != nil {
		t.Fatal(err)
	}
	clk.now = t0.Add(10 * time.Second)
	r, _ := c.Get(context.Background(), "/p")
	if !r.Deferred || !r.NotBefore.Equal(reset) {
		t.Fatalf("at the reserve: %+v, want deferred until %v", r, reset)
	}
	clk.now = reset.Add(time.Second)
	if r, _ := c.Get(context.Background(), "/p"); r.Deferred {
		t.Error("still deferred after the reset time")
	}
	if tr.n() != 2 {
		t.Errorf("transport calls = %d, want 2 (one before, one after the reset)", tr.n())
	}
}

// Transport failures back off with doubling waits capped at Max; the client never sleeps.
func TestForge_TransportErrors_BackoffGrowsAndIsCapped(t *testing.T) {
	tr := &scriptTransport{err: errors.New("connection reset")}
	clk := &clock{now: t0}
	c := &forge.Client{T: tr, B: &forge.Budget{}, Backoff: cappedBackoff, Now: clk.Now}
	want := []time.Duration{2 * time.Second, 4 * time.Second, 5 * time.Second, 5 * time.Second}
	for i, w := range want {
		r, err := c.Get(context.Background(), "/p")
		if err != nil || !r.Deferred {
			t.Fatalf("attempt %d: %+v, %v", i, r, err)
		}
		if got := r.NotBefore.Sub(t0); got != w {
			t.Errorf("attempt %d: wait = %v, want %v", i, got, w)
		}
	}
}

// The forge client holds no credential: requests carry only validators, so a
// refusal cannot lead to a different key or account being tried.
func TestForge_NoCredentialOrAccountSwitching(t *testing.T) {
	tr := &scriptTransport{replies: []forge.Response{
		{Status: 200, Header: map[string]string{"etag": `"e1"`}, Body: []byte(`[]`)},
		{Status: 401},
	}}
	clk := &clock{now: t0}
	c := newClient(tr, clk, 0)
	_, _ = c.Get(context.Background(), "/p")
	clk.now = t0.Add(time.Minute)
	if _, err := c.Get(context.Background(), "/p"); !errors.Is(err, forge.ErrDenied) {
		t.Fatalf("401: %v, want ErrDenied", err)
	}
	allowed := map[string]bool{"if-none-match": true, "if-modified-since": true}
	for _, call := range tr.calls {
		for k := range call.Header {
			if !allowed[k] {
				t.Errorf("request carries header %q beyond the validators", k)
			}
		}
	}
	if tr.n() != 2 {
		t.Errorf("transport calls = %d, want 2 (no retry after the 401)", tr.n())
	}
}
