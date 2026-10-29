package forge

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
)

// Every Kind must be in the package doc's consumer-contract table, so a Kind added without telling the
// consumers fails here.
func TestEveryKindHasAConsumerContract(t *testing.T) {
	fset := token.NewFileSet()
	af, err := parser.ParseFile(fset, "forge.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	doc := af.Doc.Text()
	kinds := 0
	ast.Inspect(af, func(n ast.Node) bool {
		vs, ok := n.(*ast.ValueSpec)
		if !ok || vs.Type == nil {
			return true
		}
		if id, ok := vs.Type.(*ast.Ident); !ok || id.Name != "Kind" {
			return true
		}
		for _, v := range vs.Values {
			lit, ok := v.(*ast.BasicLit)
			if !ok {
				continue
			}
			kinds++
			if !regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(strings.Trim(lit.Value, `"`)) + `\s{2,}\S`).MatchString(doc) {
				t.Errorf("Kind %s has no row of its own in the package doc's consumer contract table", lit.Value)
			}
		}
		return true
	})
	if kinds < 7 {
		t.Errorf("found %d Kind constants, expected at least 7", kinds)
	}
}

// The host's wording is only a fallback. Wording that is not recognised must fall to the safe side:
// stop, no retry.
func TestUnrecognisedWordingStopsInsteadOfRetrying(t *testing.T) {
	for _, body := range []string{
		"", `{"message":"Forbidden"}`, "Rate-limit exceeded", "Too many requests",
		"속도 제한을 초과했습니다", `{"message":"abuse detection"}`, "<html>403</html>",
	} {
		f := &fakeT{resps: []Response{resp(403, body, "x-ratelimit-remaining", "4000")}}
		c := client(f, t0)
		_, err := c.Get(t.Context(), "/x")
		if !errors.Is(err, ErrDenied) || len(f.calls) != 1 {
			t.Errorf("%q: err=%v calls=%d, want a stop after one call", body, err, len(f.calls))
		}
		if v := Classify(resp(403, body, "x-ratelimit-remaining", "4000"), t0, cfg(), 0); v.Kind == KindLimited {
			t.Errorf("%q was read as a rate limit", body)
		}
	}
}

// 'not accessible' only changes the reason shown. Behaviour is identical to the unclassified case.
func TestNotAccessibleWordingChangesTheReasonNeverTheBehaviour(t *testing.T) {
	denied := resp(403, `{"message":"Resource not accessible by integration"}`, "x-ratelimit-remaining", "4000")
	other := resp(403, `{"message":"Forbidden"}`, "x-ratelimit-remaining", "4000")
	vd, vo := Classify(denied, t0, cfg(), 0), Classify(other, t0, cfg(), 0)
	if vd.Kind != KindDenied || vo.Kind != KindUnclassified || vd.Reason == vo.Reason {
		t.Fatalf("kinds/reasons: %+v / %+v", vd, vo)
	}
	run := func(r Response) (error, int) {
		f := &fakeT{resps: []Response{r}}
		_, err := client(f, t0).Get(t.Context(), "/x")
		return err, len(f.calls)
	}
	ed, nd := run(denied)
	eo, no := run(other)
	if !errors.Is(ed, ErrDenied) || !errors.Is(eo, ErrDenied) || nd != no || nd != 1 {
		t.Errorf("behaviour differs: %v/%d vs %v/%d", ed, nd, eo, no)
	}
	if !errors.Is(EffectError(vd), effect.ErrDenied) || !errors.Is(EffectError(vo), effect.ErrDenied) {
		t.Error("the effect mapping differs between the two")
	}
	if vd.RetryAt != vo.RetryAt || !vd.RetryAt.IsZero() {
		t.Error("neither may carry a retry time")
	}
}

// The 'rate limit' wording DOES change the retry decision, and only when no header decided. These tests
// pin its scope and its safety margins.
func TestRateLimitWordingIsOnlyAFallbackAndKeepsTheMinimumWait(t *testing.T) {
	// headers decide first: the wording never overrides them
	v := Classify(resp(403, "rate limit", "retry-after", "7"), t0, cfg(), 0)
	if v.Kind != KindLimited || v.Signal != "header" || v.RetryAt.Sub(t0) != 7*time.Second {
		t.Errorf("retry-after must win over wording: %+v", v)
	}
	v = Classify(resp(403, "Resource not accessible by integration", "retry-after", "7"), t0, cfg(), 0)
	if v.Kind != KindLimited || v.Signal != "header" {
		t.Errorf("a header must win over the permission wording: %+v", v)
	}
	// fallback: wording only, remaining positive (documented secondary-limit case)
	v = Classify(resp(403, "You have exceeded a secondary rate limit", "x-ratelimit-remaining", "4000"), t0, cfg(), 0)
	if v.Kind != KindLimited || v.Signal != "wording" || v.RetryAt.Sub(t0) < cfg().MinSecondaryWait {
		t.Errorf("wording fallback must wait at least the configured minimum: %+v", v)
	}
	// wording alone never applies to other statuses
	if k := Classify(resp(404, "rate limit"), t0, cfg(), 0).Kind; k != KindOther {
		t.Errorf("404 with the wording: %s", k)
	}
	if k := Classify(resp(500, "rate limit"), t0, cfg(), 0).Kind; k != KindTransient {
		t.Errorf("500 with the wording: %s", k)
	}
}

// A wording-driven retry cannot run forever: after MaxAttempts it stops.
func TestRetriesAreBoundedByMaxAttempts(t *testing.T) {
	now := t0
	f := &fakeT{resps: []Response{resp(403, "You have exceeded a secondary rate limit", "x-ratelimit-remaining", "4000")}}
	c := &Client{T: f, B: &Budget{}, Backoff: cfg(), Now: func() time.Time { return now }}
	var last error
	deferred := 0
	for i := 0; i < 10 && last == nil; i++ {
		r, err := c.Get(t.Context(), "/x")
		last = err
		if err == nil && r.Deferred {
			deferred++
			now = r.NotBefore.Add(time.Second)
		}
	}
	if !errors.Is(last, ErrGiveUp) {
		t.Fatalf("err = %v, want ErrGiveUp", last)
	}
	if len(f.calls) != cfg().MaxAttempts || deferred != cfg().MaxAttempts-1 {
		t.Errorf("calls=%d deferred=%d, want %d calls", len(f.calls), deferred, cfg().MaxAttempts)
	}
}

func TestSuccessResetsTheAttemptCount(t *testing.T) {
	now := t0
	f := &fakeT{resps: []Response{resp(502, ""), resp(502, ""), resp(200, "[]"), resp(502, ""), resp(502, ""), resp(200, "[]")}}
	c := &Client{T: f, B: &Budget{}, Backoff: cfg(), Now: func() time.Time { return now }}
	for i := 0; i < 6; i++ {
		r, err := c.Get(t.Context(), "/x")
		if err != nil {
			t.Fatalf("call %d: %v (success must reset the count)", i, err)
		}
		now = r.NotBefore.Add(time.Second)
		if !r.Deferred {
			now = now.Add(time.Hour)
		}
	}
}
