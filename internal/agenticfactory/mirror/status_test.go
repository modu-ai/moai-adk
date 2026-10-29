package mirror

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

var t0 = time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)

func timed(f *fakeGH) *Mirror {
	m := writable(f)
	m.Now = func() time.Time { return t0 }
	m.Backoff = forge.Backoff{MinSecondaryWait: time.Minute, Base: 5 * time.Second, Max: 10 * time.Minute, MaxAttempts: 5}
	return m
}

func blocked(t *testing.T, err error) *BlockedError {
	t.Helper()
	var be *BlockedError
	if !errors.As(err, &be) {
		t.Fatalf("want a BlockedError, got %v", err)
	}
	return be
}

// The decision rests on the status code and headers, never on error wording.
func TestWordingAloneNeverDecidesABlock(t *testing.T) {
	// stderr shouts rate limit, but no HTTP status reached stdout.
	f := &fakeGH{results: []result{{errOut: "HTTP 429: API rate limit exceeded", exit: 1}}}
	err := timed(f).WriteLedger(context.Background(), ledger())
	var be *BlockedError
	if err == nil || errors.As(err, &be) {
		t.Fatalf("wording without a status must be a plain error, got %v", err)
	}
	// A 500 is transient: reported with its status, not a block reason.
	f = &fakeGH{results: []result{httpFail(500)}}
	err = timed(f).WriteLedger(context.Background(), ledger())
	if err == nil || errors.As(err, &be) || !strings.Contains(err.Error(), "500") {
		t.Errorf("a server error is a plain status error, got %v", err)
	}
}

func TestLimitVersusDenialFollowsStatusAndHeaders(t *testing.T) {
	reset := t0.Add(7 * time.Minute)
	cases := []struct {
		name   string
		res    result
		reason core.BlockReason
		retry  time.Time
	}{
		{"403 with retry-after", httpFail(403, "retry-after", "60"), core.ReasonRateLimited, t0.Add(60 * time.Second)},
		{"403 exhausted with reset", httpFail(403, "x-ratelimit-remaining", "0", "x-ratelimit-reset", fmt.Sprint(reset.Unix())), core.ReasonRateLimited, reset},
		{"403 secondary limit wording in the body", httpFail(403, "body", `{"message":"You have exceeded a secondary rate limit."}`), core.ReasonRateLimited, t0.Add(time.Minute)},
		{"429", httpFail(429, "retry-after", "30"), core.ReasonRateLimited, t0.Add(30 * time.Second)},
		{"403 with no signal is a denial", httpFail(403), core.ReasonPermissionDenied, time.Time{}},
		{"401", httpFail(401), core.ReasonAuthFailed, time.Time{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fakeGH{results: []result{c.res}}
			be := blocked(t, timed(f).WriteLedger(context.Background(), ledger()))
			if be.Reason != c.reason || !be.RetryAt.Equal(c.retry) {
				t.Errorf("reason=%q retryAt=%s, want %q %s", be.Reason, be.RetryAt, c.reason, c.retry)
			}
			if len(f.calls) != 1 {
				t.Errorf("a block must not be retried, calls=%v", calls(f))
			}
		})
	}
}

// Backoff numbers come from configuration. Without them a limit with no hint
// has no known retry time, and the code says so instead of inventing one.
func TestLimitWithoutConfiguredBackoffHasNoRetryTime(t *testing.T) {
	f := &fakeGH{results: []result{httpFail(403, "body", "secondary rate limit")}}
	m := writable(f)
	m.Now = func() time.Time { return t0 }
	be := blocked(t, m.WriteLedger(context.Background(), ledger()))
	if !be.RetryAt.IsZero() || be.Reason != core.ReasonRateLimited {
		t.Errorf("block: %+v", be)
	}
}

func TestGhCallsRequestHeadersSoTheStatusIsVisible(t *testing.T) {
	f := &fakeGH{results: []result{{out: ""}, {out: "{}"}}}
	_ = timed(f).WriteLedger(context.Background(), ledger())
	for _, c := range f.calls {
		if len(c) < 2 || c[0] != "api" || c[1] != "-i" {
			t.Errorf("gh api must be called with -i, got %v", c)
		}
	}
}

func TestSuccessResponsesWithHeadersAreUnwrapped(t *testing.T) {
	post := result{out: "HTTP/2.0 201 Created\r\ncontent-type: application/json\r\n\r\n{\"id\":4242}"}
	f := &fakeGH{results: []result{{out: "HTTP/2.0 200 OK\r\n\r\n[]"}, post, {out: "HTTP/2.0 200 OK\r\n\r\n{}"}}}
	m := timed(f)
	if err := m.WriteLedger(context.Background(), ledgerWith("a")); err != nil {
		t.Fatal(err)
	}
	if err := m.WriteLedger(context.Background(), ledgerWith("b")); err != nil {
		t.Fatal(err)
	}
	if last := calls(f)[2]; !strings.Contains(last, "comments/4242") {
		t.Errorf("the id behind the response headers must be remembered: %s", last)
	}
}

func TestListingPagesUntilTheLedgerCommentIsFound(t *testing.T) {
	var page1 []string
	for i := 0; i < 100; i++ {
		page1 = append(page1, fmt.Sprintf(`{"id":%d,"body":"human note"}`, i+1))
	}
	f := &fakeGH{results: []result{
		{out: "[" + strings.Join(page1, ",") + "]"},
		{out: `[{"id":9001,"body":"` + Marker + `"}]`},
		{out: "{}"},
	}}
	if err := timed(f).WriteLedger(context.Background(), ledger()); err != nil {
		t.Fatal(err)
	}
	cs := calls(f)
	if len(cs) != 3 || !strings.Contains(cs[0], "page=1") || !strings.Contains(cs[1], "page=2") || !strings.Contains(cs[2], "comments/9001") || !strings.Contains(cs[2], "-X PATCH") {
		t.Errorf("paging sequence: %v", cs)
	}
}

func TestMissingIssueOnListIsAPlainStatusError(t *testing.T) {
	f := &fakeGH{results: []result{httpFail(404)}}
	err := timed(f).WriteLedger(context.Background(), ledger())
	var be *BlockedError
	var se *forge.StatusError
	if err == nil || errors.As(err, &be) || !errors.As(err, &se) || se.Status != 404 {
		t.Errorf("a 404 on the listing is not 'comment gone': %v", err)
	}
}

func TestErrorTextIsMaskedBeforeItIsReturned(t *testing.T) {
	f := &fakeGH{results: []result{{errOut: "boom password=hunter2 for dev@example.com in /Users/goos/x", exit: 1}}}
	err := timed(f).WriteLedger(context.Background(), ledger())
	if err == nil {
		t.Fatal("want an error")
	}
	for _, leak := range []string{"hunter2", "dev@example.com", "/Users/goos"} {
		if strings.Contains(err.Error(), leak) {
			t.Errorf("%q leaked into the error: %v", leak, err)
		}
	}
}
