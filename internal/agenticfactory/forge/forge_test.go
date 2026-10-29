package forge

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

var t0 = time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)

func cfg() Backoff {
	return Backoff{MinSecondaryWait: time.Minute, Base: 2 * time.Second, Max: 10 * time.Minute, MaxAttempts: 4}
}

func resp(status int, body string, kv ...string) Response {
	h := map[string]string{}
	for i := 0; i+1 < len(kv); i += 2 {
		h[kv[i]] = kv[i+1]
	}
	return Response{Status: status, Header: h, Body: []byte(body)}
}

func TestBackoffMustBeConfiguredNotDefaulted(t *testing.T) {
	if err := (Backoff{}).Validate(); err == nil {
		t.Error("zero backoff accepted: limits must come from config or documentation, never silently")
	}
	if err := cfg().Validate(); err != nil {
		t.Error(err)
	}
}

func TestClassifyTable(t *testing.T) {
	reset := strconv.FormatInt(t0.Add(7*time.Minute).Unix(), 10)
	cases := []struct {
		name string
		r    Response
		want Kind
		at   time.Duration // expected RetryAt - now, 0 = not checked
	}{
		{"ok", resp(200, "{}"), KindOK, 0},
		{"not modified", resp(304, ""), KindNotModified, 0},
		{"429 retry-after", resp(429, "", "retry-after", "30"), KindLimited, 30 * time.Second},
		{"403 primary: remaining 0", resp(403, "API rate limit exceeded", "x-ratelimit-remaining", "0", "x-ratelimit-reset", reset), KindLimited, 7 * time.Minute},
		{"403 secondary: retry-after", resp(403, "You have exceeded a secondary rate limit", "retry-after", "45"), KindLimited, 45 * time.Second},
		{"403 secondary: text only waits the configured minimum", resp(403, "You have exceeded a secondary rate limit"), KindLimited, time.Minute},
		{"403 permission denied", resp(403, `{"message":"Resource not accessible by integration"}`, "x-ratelimit-remaining", "4000"), KindDenied, 0},
		{"401", resp(401, "Bad credentials"), KindDenied, 0},
		{"gitlab 429", resp(429, "", "ratelimit-reset", reset, "ratelimit-remaining", "0"), KindLimited, 7 * time.Minute},
		{"502", resp(502, ""), KindTransient, 2 * time.Second},
		{"404", resp(404, ""), KindOther, 0},
	}
	for _, c := range cases {
		v := Classify(c.r, t0, cfg(), 0)
		if v.Kind != c.want {
			t.Errorf("%s: kind = %s, want %s", c.name, v.Kind, c.want)
			continue
		}
		if c.at != 0 && v.RetryAt.Sub(t0) != c.at {
			t.Errorf("%s: retry in %v, want %v", c.name, v.RetryAt.Sub(t0), c.at)
		}
	}
}

func TestBackoffGrowsExponentiallyAndIsCapped(t *testing.T) {
	prev := time.Duration(0)
	for a := 0; a < 3; a++ {
		d := Classify(resp(502, ""), t0, cfg(), a).RetryAt.Sub(t0)
		if d <= prev {
			t.Errorf("attempt %d wait %v not above %v", a, d, prev)
		}
		prev = d
	}
	if d := Classify(resp(502, ""), t0, cfg(), 30).RetryAt.Sub(t0); d != cfg().Max {
		t.Errorf("uncapped wait %v", d)
	}
}

func TestRetryAfterAsHTTPDate(t *testing.T) {
	d := t0.Add(90 * time.Second).UTC().Format(http.TimeFormat)
	if got := Classify(resp(429, "", "retry-after", d), t0, cfg(), 0).RetryAt.Sub(t0); got != 90*time.Second {
		t.Errorf("wait = %v", got)
	}
}

func TestBudgetFollowsHeadersAndBlocks(t *testing.T) {
	b := &Budget{Reserve: 10}
	if ok, _ := b.Allow(t0); !ok {
		t.Error("unknown budget must allow the first call")
	}
	reset := strconv.FormatInt(t0.Add(time.Hour).Unix(), 10)
	b.Observe(resp(200, "", "x-ratelimit-remaining", "10", "x-ratelimit-reset", reset, "x-ratelimit-limit", "5000"), t0)
	if ok, until := b.Allow(t0); ok || until.Sub(t0) != time.Hour {
		t.Errorf("at the reserve: ok=%v until=%v", ok, until)
	}
	if ok, _ := b.Allow(t0.Add(time.Hour + time.Second)); !ok {
		t.Error("after the reset the budget must reopen")
	}
	b.Block(t0.Add(time.Minute))
	if ok, _ := b.Allow(t0.Add(30 * time.Second)); ok {
		t.Error("blocked budget allowed a call")
	}
}

type fakeT struct {
	calls []Request
	resps []Response
	err   error
}

func (f *fakeT) Do(_ context.Context, r Request) (Response, error) {
	f.calls = append(f.calls, r)
	if f.err != nil {
		return Response{}, f.err
	}
	i := len(f.calls) - 1
	if i >= len(f.resps) {
		i = len(f.resps) - 1
	}
	return f.resps[i], nil
}

func client(f *fakeT, now time.Time) *Client {
	return &Client{T: f, B: &Budget{}, Backoff: cfg(), Now: func() time.Time { return now }}
}

func TestConditionalGetSendsValidatorsAndServesCacheOn304(t *testing.T) {
	f := &fakeT{resps: []Response{
		resp(200, `[{"number":1}]`, "etag", `"abc"`, "last-modified", "Wed, 21 Oct 2026 07:28:00 GMT"),
		resp(304, ""),
	}}
	c := client(f, t0)
	r1, err := c.Get(t.Context(), "/repos/o/r/pulls")
	if err != nil || !r1.Changed || r1.FromCache {
		t.Fatalf("first = %+v err=%v", r1, err)
	}
	r2, err := c.Get(t.Context(), "/repos/o/r/pulls")
	if err != nil || r2.Changed || !r2.FromCache || string(r2.Body) != `[{"number":1}]` {
		t.Fatalf("second = %+v err=%v", r2, err)
	}
	h := f.calls[1].Header
	if h["if-none-match"] != `"abc"` || h["if-modified-since"] == "" {
		t.Errorf("validators not sent: %v", h)
	}
	if len(f.calls[0].Header) != 0 {
		t.Errorf("first request carried validators: %v", f.calls[0].Header)
	}
}

func TestOnlyReadsAreIssued(t *testing.T) {
	f := &fakeT{resps: []Response{resp(200, "[]")}}
	_, _ = client(f, t0).Get(t.Context(), "/x")
	for _, c := range f.calls {
		if c.Method != "GET" {
			t.Errorf("method %s: the forge client is read-only", c.Method)
		}
	}
}

func TestLimitedDefersWithoutRetryingAndBlocksTheSharedBudget(t *testing.T) {
	f := &fakeT{resps: []Response{resp(429, "", "retry-after", "30")}}
	c := client(f, t0)
	r, err := c.Get(t.Context(), "/x")
	if err != nil || !r.Deferred || r.NotBefore.Sub(t0) != 30*time.Second {
		t.Fatalf("r=%+v err=%v", r, err)
	}
	r, _ = c.Get(t.Context(), "/y") // another path on the same budget
	if !r.Deferred || len(f.calls) != 1 {
		t.Errorf("a blocked budget must stop further calls: calls=%d deferred=%v", len(f.calls), r.Deferred)
	}
}

func TestPermissionDenialIsNotRetriedAndIsReportedAsDenied(t *testing.T) {
	f := &fakeT{resps: []Response{resp(403, `{"message":"Resource not accessible by integration"}`)}}
	c := client(f, t0)
	_, err := c.Get(t.Context(), "/x")
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("err = %v", err)
	}
	if len(f.calls) != 1 {
		t.Errorf("calls = %d, want 1 (no retry)", len(f.calls))
	}
}

func TestTransportErrorDefersAndNeverLoopsInside(t *testing.T) {
	f := &fakeT{err: errors.New("dial tcp: timeout")}
	r, err := client(f, t0).Get(t.Context(), "/x")
	if err != nil || !r.Deferred || len(f.calls) != 1 {
		t.Errorf("r=%+v err=%v calls=%d", r, err, len(f.calls))
	}
}

func TestPollIntervalHeaderIsHonoured(t *testing.T) {
	f := &fakeT{resps: []Response{resp(200, "[]", "x-poll-interval", "60")}}
	r, _ := client(f, t0).Get(t.Context(), "/x")
	if r.PollInterval != 60*time.Second {
		t.Errorf("poll interval = %v", r.PollInterval)
	}
}

func TestEffectErrorMapsLimitAndDenial(t *testing.T) {
	lim := EffectError(Classify(resp(429, "", "retry-after", "30"), t0, cfg(), 0))
	var le *effect.LimitedError
	if !errors.As(lim, &le) || le.RetryAt.Sub(t0) != 30*time.Second {
		t.Errorf("limit mapped to %v", lim)
	}
	if den := EffectError(Classify(resp(403, `{"message":"forbidden"}`), t0, cfg(), 0)); !errors.Is(den, effect.ErrDenied) {
		t.Errorf("denial mapped to %v", den)
	}
	if EffectError(Verdict{Kind: KindOK}) != nil {
		t.Error("ok must map to nil")
	}
}

// lane-17 table G5/G6: a 403 is a permission denial only when the host says so; a 403 with no signal at
// all is unclassified. Both stop and neither retries.
func TestUnsignalled403IsUnclassifiedNotPermissionNotRateLimit(t *testing.T) {
	v := Classify(resp(403, `{"message":"Forbidden"}`, "x-ratelimit-remaining", "4000"), t0, cfg(), 0)
	if v.Kind != KindUnclassified {
		t.Fatalf("kind = %s", v.Kind)
	}
	f := &fakeT{resps: []Response{resp(403, `{"message":"Forbidden"}`, "x-ratelimit-remaining", "4000")}}
	c := client(f, t0)
	if _, err := c.Get(t.Context(), "/x"); !errors.Is(err, ErrDenied) || len(f.calls) != 1 {
		t.Errorf("err=%v calls=%d: an unclassified 403 must stop without a retry", err, len(f.calls))
	}
	if !errors.Is(EffectError(v), effect.ErrDenied) {
		t.Error("an unclassified 403 must park the intent")
	}
}

func TestGitLabResetTimeHeaderIsReadAndMissingHeadersStayUnknown(t *testing.T) {
	at := t0.Add(5 * time.Minute).UTC().Format(http.TimeFormat)
	b := &Budget{}
	b.Observe(resp(200, "", "ratelimit-remaining", "0", "ratelimit-resettime", at), t0)
	if ok, until := b.Allow(t0); ok || until.Sub(t0) != 5*time.Minute {
		t.Errorf("ok=%v until=%v", ok, until.Sub(t0))
	}
	unknown := &Budget{}
	unknown.Observe(resp(200, ""), t0) // Projects/Groups/Users API answers carry no rate-limit headers
	if ok, _ := unknown.Allow(t0); !ok {
		t.Error("absent headers must read as unknown, not as an exhausted budget")
	}
	v := Classify(resp(429, "", "ratelimit-resettime", at, "ratelimit-remaining", "0"), t0, cfg(), 0)
	if v.Kind != KindLimited || v.RetryAt.Sub(t0) != 5*time.Minute {
		t.Errorf("429 with a reset date: %+v", v)
	}
}

// Limits are read from response headers and configuration, never carried as constants. The scan flags
// the documented limit values and the hour/minute units anywhere anywhere in the package (which carries the
// two documentation-stated values together with their source URL).
func TestNoHardcodedLimitConstants(t *testing.T) {
	banned := map[string]bool{"60": true, "100": true, "900": true, "1000": true, "3600": true, "5000": true, "15000": true}
	fset := token.NewFileSet()
	files, _ := filepath.Glob("*.go")
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range af.Decls {
			ast.Inspect(d, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.BasicLit:
					if x.Kind == token.INT && banned[x.Value] {
						t.Errorf("%s: numeric literal %s looks like a hardcoded limit", fset.Position(x.Pos()), x.Value)
					}
				case *ast.SelectorExpr:
					if id, ok := x.X.(*ast.Ident); ok && id.Name == "time" && (x.Sel.Name == "Minute" || x.Sel.Name == "Hour") {
						t.Errorf("%s: time.%s anywhere in the package", fset.Position(x.Pos()), x.Sel.Name)
					}
				}
				return true
			})
		}
	}
}

func TestBackoffMaxBelowTheMinimumIsRejectedAndNeverShortensIt(t *testing.T) {
	bad := cfg()
	bad.Max = bad.MinSecondaryWait / 2
	if err := bad.Validate(); err == nil {
		t.Error("Max below MinSecondaryWait accepted")
	}
	bad = cfg()
	bad.Max = bad.Base / 2
	if err := bad.Validate(); err == nil {
		t.Error("Max below Base accepted")
	}
	// Even an unvalidated config must not shorten the documented minimum.
	if d := Classify(resp(403, "secondary rate limit"), t0, bad, 0).RetryAt.Sub(t0); d < bad.MinSecondaryWait {
		t.Errorf("wait %v is shorter than the minimum %v", d, bad.MinSecondaryWait)
	}
}
