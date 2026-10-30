package forge

import (
	"testing"
	"time"
)

// AC-AF-014 (design 4.5 "budget rules" and rows L5, L6). Limits, remaining counts and resume times come from
// response headers only; without a header the remaining count is UNKNOWN; the waits follow configuration;
// state and safety writes keep the reserve; and a limit never moves a call to another credential or endpoint.

func TestBudget_NoRequestIsSentBeforeTheResumeTime(t *testing.T) {
	d := newDriver(resp(429, "", "retry-after", "300"))
	_, _ = d.c.Get(t.Context(), "/repos/o/r/issues")
	for _, at := range []time.Duration{time.Second, 100 * time.Second, 299 * time.Second} {
		d.now = t0.Add(at)
		res, _ := d.c.Get(t.Context(), "/repos/o/r/issues")
		if !res.Deferred || !res.NotBefore.Equal(t0.Add(300*time.Second)) {
			t.Errorf("at +%v: res=%+v, want deferred until the resume time", at, res)
		}
	}
	if len(d.f.calls) != 1 {
		t.Errorf("transport calls = %d, want 1: nothing may be sent before the resume time", len(d.f.calls))
	}
}

func TestBudget_NoHeaderMeansUNKNOWN_NeverASuppliedNumber(t *testing.T) {
	b := &Budget{}
	if _, known := b.RemainingFor("core"); known {
		t.Error("an unseen resource reads as known")
	}
	b.Observe(resp(200, ""), t0) // a response with no rate-limit header (GitLab Projects/Groups/Users API, L6)
	if n, known := b.RemainingFor(""); known {
		t.Errorf("a response without a header produced a remaining count (%d)", n)
	}
	b.Observe(resp(200, "", "ratelimit-remaining", "41", "ratelimit-reset", reset(time.Hour)), t0) // L5
	if n, known := b.RemainingFor(""); !known || n != 41 {
		t.Errorf("remaining = %d known=%v, want the stated 41", n, known)
	}
	// The window the count described has rolled over: the count is forgotten, not kept.
	b.AllowFor("", t0.Add(2*time.Hour))
	if _, known := b.RemainingFor(""); known {
		t.Error("a stale count survived the window it described")
	}
}

func TestBudget_TheWaitFollowsTheConfiguredDefaults(t *testing.T) {
	r := resp(403, `{"message":"secondary rate limit"}`)
	short := Backoff{MinSecondaryWait: 30 * time.Second, Base: time.Second, Max: time.Hour, MaxAttempts: 3}
	long := Backoff{MinSecondaryWait: 7 * time.Minute, Base: time.Second, Max: time.Hour, MaxAttempts: 3}
	a, b := Classify(r, t0, short, 0), Classify(r, t0, long, 0)
	if !a.RetryAt.Equal(t0.Add(30*time.Second)) || !b.RetryAt.Equal(t0.Add(7*time.Minute)) {
		t.Errorf("waits = %v / %v, want the configured 30s and 7m", a.RetryAt.Sub(t0), b.RetryAt.Sub(t0))
	}
	if err := (Backoff{}).Validate(); err == nil {
		t.Error("a Backoff with no configured values was accepted: there is no hidden default")
	}
}

func TestBudget_TheReserveBelongsToStateAndSafetyWrites_AndALimitNeverSwitchesCredentialOrEndpoint(t *testing.T) {
	b := &Budget{ReserveCalls: 10}
	b.Observe(resp(200, "", "x-ratelimit-remaining", "10", "x-ratelimit-reset", reset(time.Hour)), t0)
	if ok, _ := b.Allow(t0); ok {
		t.Error("an ordinary call went out at the reserve")
	}
	if ok, _ := b.AllowReserved("", t0); !ok {
		t.Error("a state/safety write was held back from the reserve")
	}
	if !b.Reserve(0) {
		t.Error("something more urgent than a report was held")
	}

	// A limited call is retried on the SAME path with the same credential-free headers: nothing is switched.
	d := newDriver(resp(429, "", "retry-after", "5"))
	_, _ = d.c.Get(t.Context(), "/repos/o/r/issues")
	d.now = t0.Add(10 * time.Second)
	_, _ = d.c.Get(t.Context(), "/repos/o/r/issues")
	if len(d.f.calls) != 2 {
		t.Fatalf("calls = %d, want the retry after the wait", len(d.f.calls))
	}
	for i, c := range d.f.calls {
		if c.Path != "/repos/o/r/issues" || c.Method != "GET" {
			t.Errorf("call %d went to %s %s, want the same endpoint", i, c.Method, c.Path)
		}
		for k := range c.Header {
			if k == "authorization" || k == "private-token" {
				t.Errorf("call %d carries a credential header %q the client added", i, k)
			}
		}
	}
}
