package forge

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

// AC-AF-014 / design 4.5: each row of the response table gives the classification and the retry count the
// table states. Rows are named after the table (G1..G8 for GitHub, L1, L2, L4 for GitLab; L3 is blocked
// on OQ-23 and L5/L6 are budget rows, see TestBudget_*). "Retry count" is measured as transport calls made
// by a Client that keeps asking.

// driver answers every request with r and lets the test move the clock.
type driver struct {
	f   *fakeT
	c   *Client
	now time.Time
}

func newDriver(r Response) *driver {
	d := &driver{f: &fakeT{resps: []Response{r}}, now: t0}
	d.c = &Client{T: d.f, B: &Budget{}, Backoff: cfg(), Now: func() time.Time { return d.now }}
	return d
}

func TestForgeClassify_G1_RetryAfterIsASecondaryLimit_NoCallBeforeTheSecondsPass(t *testing.T) {
	for _, status := range []int{403, 429} {
		d := newDriver(resp(status, "", "retry-after", "120"))
		v := Classify(resp(status, "", "retry-after", "120"), t0, cfg(), 0)
		if v.Kind != KindLimited || v.Signal != "header" || !v.RetryAt.Equal(t0.Add(120*time.Second)) {
			t.Fatalf("%d: verdict = %+v", status, v)
		}
		res, err := d.c.Get(t.Context(), "/x")
		if err != nil || !res.Deferred || !res.NotBefore.Equal(t0.Add(120*time.Second)) {
			t.Fatalf("%d: res=%+v err=%v", status, res, err)
		}
		d.now = t0.Add(60 * time.Second)
		_, _ = d.c.Get(t.Context(), "/x")
		if len(d.f.calls) != 1 {
			t.Errorf("%d: calls = %d inside the retry-after, want 0 more", status, len(d.f.calls))
		}
		d.now = t0.Add(121 * time.Second)
		_, _ = d.c.Get(t.Context(), "/x")
		if len(d.f.calls) != 2 {
			t.Errorf("%d: calls = %d after the seconds passed, want the next call", status, len(d.f.calls))
		}
	}
}

func TestForgeClassify_G2_ExhaustedRemainingWaitsForTheReset_AndNamesTheResource(t *testing.T) {
	reset := t0.Add(30 * time.Minute)
	r := resp(403, "", "x-ratelimit-remaining", "0", "x-ratelimit-reset", strconv.FormatInt(reset.Unix(), 10), "x-ratelimit-resource", "search")
	v := Classify(r, t0, cfg(), 0)
	if v.Kind != KindLimited || !v.RetryAt.Equal(reset) {
		t.Fatalf("verdict = %+v, want rate-limited until the reset time", v)
	}
	d := newDriver(r)
	_, _ = d.c.Get(t.Context(), "/search/issues")
	d.now = t0.Add(10 * time.Minute)
	_, _ = d.c.Get(t.Context(), "/search/issues")
	if len(d.f.calls) != 1 {
		t.Errorf("calls = %d before the reset, want 0 more", len(d.f.calls))
	}
	// The block belongs to the resource the response named: another resource stays open.
	if ok, _ := d.c.B.AllowFor("core", d.now); !ok {
		t.Error("the exhausted search resource blocked the core resource")
	}
	if ok, _ := d.c.B.AllowFor("search", d.now); ok {
		t.Error("the named resource is not blocked")
	}
}

func TestForgeClassify_G3_SecondaryWording_BacksOffThenGivesUp(t *testing.T) {
	r := resp(403, `{"message":"You have exceeded a secondary rate limit."}`)
	v := Classify(r, t0, cfg(), 0)
	if v.Kind != KindLimited || v.Signal != "wording" || !v.RetryAt.Equal(t0.Add(cfg().MinSecondaryWait)) {
		t.Fatalf("verdict = %+v, want rate-limited by wording, waiting at least the configured minimum", v)
	}
	assertBoundedRetries(t, r)
}

func TestForgeClassify_G4_A429WithNoHintIsALimitNeverAPermission(t *testing.T) {
	r := resp(429, "")
	v := Classify(r, t0, cfg(), 0)
	if v.Kind != KindLimited {
		t.Fatalf("verdict = %+v, want rate-limited", v)
	}
	assertBoundedRetries(t, r)
}

// assertBoundedRetries drives a Client through repeated limited answers: it makes exactly MaxAttempts calls,
// each only after the previous wait, and then gives up.
func assertBoundedRetries(t *testing.T, r Response) {
	t.Helper()
	d := newDriver(r)
	var err error
	for i := 0; i < 20 && !errors.Is(err, ErrGiveUp); i++ {
		_, err = d.c.Get(t.Context(), "/x")
		d.now = d.now.Add(24 * time.Hour) // always past the wait
	}
	if !errors.Is(err, ErrGiveUp) || len(d.f.calls) != cfg().MaxAttempts {
		t.Errorf("err=%v calls=%d, want ErrGiveUp after exactly %d calls", err, len(d.f.calls), cfg().MaxAttempts)
	}
}

func TestForgeClassify_G5_NotAccessible_IsPermissionDenied_NoRetry(t *testing.T) {
	r := resp(403, `{"message":"Resource not accessible by integration"}`, "x-ratelimit-remaining", "4000", "x-accepted-github-permissions", "issues=write")
	if v := Classify(r, t0, cfg(), 0); v.Kind != KindDenied {
		t.Fatalf("verdict = %+v, want permission-denied", v)
	}
	assertDeniedOnce(t, r)
}

func TestForgeClassify_G6_A403WithNoSignal_IsUnclassified_NoRetry(t *testing.T) {
	r := resp(403, `{"message":"Forbidden"}`, "x-ratelimit-remaining", "4000")
	if v := Classify(r, t0, cfg(), 0); v.Kind != KindUnclassified {
		t.Fatalf("verdict = %+v, want unclassified-403: neither a permission nor a limit may be asserted", v)
	}
	assertDeniedOnce(t, r)
}

func TestForgeClassify_G7_A401IsAnAuthenticationFailure_NoRetry(t *testing.T) {
	r := resp(401, "")
	// The table names the class auth-failed. The code reports it as Denied with the reason
	// "authentication failed" (there is no separate Kind), so the two kinds of denial are told apart by the
	// reason, and the mirror maps Kind -> block reason. A distinct Kind would be a change in lane-18's map.
	if v := Classify(r, t0, cfg(), 0); v.Kind != KindDenied || v.Reason != "authentication failed" {
		t.Fatalf("verdict = %+v", v)
	}
	assertDeniedOnce(t, r)
}

func assertDeniedOnce(t *testing.T, r Response) {
	t.Helper()
	d := newDriver(r)
	_, err := d.c.Get(t.Context(), "/x")
	d.now = d.now.Add(24 * time.Hour)
	_, err2 := d.c.Get(t.Context(), "/x")
	if !errors.Is(err, ErrDenied) || !errors.Is(err2, ErrDenied) {
		t.Errorf("errors = %v / %v, want ErrDenied both times", err, err2)
	}
	if len(d.f.calls) != 2 {
		t.Errorf("calls = %d, want one per Get and never a retry inside one Get", len(d.f.calls))
	}
}

func TestForgeClassify_G8_A404IsNeitherDeletionNorPermission(t *testing.T) {
	r := resp(404, `{"message":"Not Found"}`)
	v := Classify(r, t0, cfg(), 0)
	if v.Kind != KindOther {
		t.Fatalf("verdict = %+v, want other: a 404 proves neither deletion nor missing permission", v)
	}
	d := newDriver(r)
	res, err := d.c.Get(t.Context(), "/x")
	if errors.Is(err, ErrDenied) || len(d.f.calls) != 1 {
		t.Errorf("res=%+v err=%v calls=%d, want no denial and one call", res, err, len(d.f.calls))
	}
}

func TestForgeClassify_L1_GitLab429WithRetryAfter_WaitsThatLong(t *testing.T) {
	r := resp(429, "", "retry-after", "45", "ratelimit-resettime", t0.Add(time.Hour).Format(time.RFC1123), "ratelimit-name", "throttle_authenticated_api")
	v := Classify(r, t0, cfg(), 0)
	if v.Kind != KindLimited || !v.RetryAt.Equal(t0.Add(45*time.Second)) {
		t.Fatalf("verdict = %+v, want rate-limited until Retry-After", v)
	}
	d := newDriver(r)
	_, _ = d.c.Get(t.Context(), "/api/v4/projects")
	d.now = t0.Add(20 * time.Second)
	_, _ = d.c.Get(t.Context(), "/api/v4/projects")
	if len(d.f.calls) != 1 {
		t.Errorf("calls = %d inside Retry-After, want 0 more", len(d.f.calls))
	}
	// RateLimit-Name is the resource: it blocks that name only.
	if ok, _ := d.c.B.AllowFor("other", d.now); !ok {
		t.Error("a named GitLab limit blocked an unrelated resource")
	}
}

func TestForgeClassify_L2_GitLab429WithoutRetryAfter_UsesTheConfiguredBackoff(t *testing.T) {
	r := resp(429, "")
	v := Classify(r, t0, cfg(), 0)
	if v.Kind != KindLimited || !v.RetryAt.Equal(t0.Add(cfg().MinSecondaryWait)) {
		t.Fatalf("verdict = %+v, want rate-limited with the configured minimum wait", v)
	}
}

func TestForgeClassify_L4_GitLab404IsNotEvidenceOfDeletionOrPermission(t *testing.T) {
	if v := Classify(resp(404, ""), t0, cfg(), 0); v.Kind != KindOther {
		t.Errorf("verdict = %+v", v)
	}
}
