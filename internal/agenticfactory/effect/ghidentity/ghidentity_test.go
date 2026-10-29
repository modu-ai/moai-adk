package ghidentity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

var _ effect.Identity = (*GitHub)(nil)

type fakeT struct {
	calls []forge.Request
	by    map[string]forge.Response
}

func (f *fakeT) Do(_ context.Context, r forge.Request) (forge.Response, error) {
	f.calls = append(f.calls, r)
	if resp, ok := f.by[r.Path]; ok {
		return resp, nil
	}
	return forge.Response{Status: 404}, nil
}

func ok(body string) forge.Response { return forge.Response{Status: 200, Body: []byte(body)} }

var t0 = time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)

func newGH(f *fakeT, maxAge time.Duration, now *time.Time) *GitHub {
	return &GitHub{
		MaxAge: maxAge, Now: func() time.Time { return *now },
		Client: &forge.Client{
			T: f, B: &forge.Budget{}, Now: func() time.Time { return *now },
			Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3},
		},
	}
}

func TestIDsComeFromNumericFieldsOfReadOnlyGETs(t *testing.T) {
	now := t0
	f := &fakeT{by: map[string]forge.Response{
		"/user":                        ok(`{"login":"bot","id":1001}`),
		"/repos/o/r/issues/7":          ok(`{"number":7,"user":{"login":"someone","id":2002}}`),
		"/repos/o/r/issues/comments/9": ok(`{"id":9,"user":{"login":"bot","id":1001}}`),
	}}
	g := newGH(f, time.Minute, &now)
	if id, err := g.CurrentUserID(t.Context()); err != nil || id != 1001 {
		t.Errorf("user = %d, %v", id, err)
	}
	if id, err := g.IssueAuthorID(t.Context(), "o/r", 7); err != nil || id != 2002 {
		t.Errorf("issue author = %d, %v", id, err)
	}
	if id, err := g.CommentAuthorID(t.Context(), "o/r", 9); err != nil || id != 1001 {
		t.Errorf("comment author = %d, %v", id, err)
	}
	for _, c := range f.calls {
		if c.Method != "GET" || len(c.Body) != 0 {
			t.Errorf("non-read request %s %s", c.Method, c.Path)
		}
	}
}

func TestALoginWithoutANumericIDIsUnknownNeverCompared(t *testing.T) {
	now := t0
	f := &fakeT{by: map[string]forge.Response{
		"/user":                ok(`{"login":"bot"}`),
		"/repos/o/r/issues/7":  ok(`{"user":{"login":"someone"}}`),
		"/repos/o/r/issues/8":  ok(`{"user":null}`),
		"/repos/o/r/issues/9":  ok(`{"user":{"id":0}}`),
		"/repos/o/r/issues/10": ok(`not json`),
	}}
	g := newGH(f, time.Minute, &now)
	if _, err := g.CurrentUserID(t.Context()); err == nil {
		t.Error("a user response with only a login was accepted")
	}
	for _, n := range []int{7, 8, 9, 10, 11} {
		if id, err := g.IssueAuthorID(t.Context(), "o/r", n); err == nil || id != 0 {
			t.Errorf("issue %d: id=%d err=%v, want unknown", n, id, err)
		}
	}
}

func TestIDsAreReusedWithinMaxAgeAndFetchedAgainWhenStale(t *testing.T) {
	now := t0
	f := &fakeT{by: map[string]forge.Response{"/user": ok(`{"id":1001}`)}}
	g := newGH(f, time.Minute, &now)
	for i := 0; i < 3; i++ {
		if _, err := g.CurrentUserID(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	if len(f.calls) != 1 {
		t.Fatalf("calls = %d, want 1 within MaxAge", len(f.calls))
	}
	now = now.Add(2 * time.Minute)
	f.by["/user"] = ok(`{"id":3003}`) // the token now belongs to someone else
	id, err := g.CurrentUserID(t.Context())
	if err != nil || id != 3003 || len(f.calls) != 2 {
		t.Errorf("stale cache: id=%d err=%v calls=%d, want a fresh lookup", id, err, len(f.calls))
	}
	zero := newGH(f, 0, &now)
	_, _ = zero.CurrentUserID(t.Context())
	_, _ = zero.CurrentUserID(t.Context())
	if got := len(f.calls); got != 4 {
		t.Errorf("MaxAge 0 must not cache: calls=%d", got)
	}
}

func TestLimitsDenialsAndErrorsAreUnknown(t *testing.T) {
	now := t0
	for name, resp := range map[string]forge.Response{
		"429": {Status: 429, Header: map[string]string{"retry-after": "30"}},
		"403": {Status: 403, Body: []byte(`{"message":"Resource not accessible by integration"}`)},
		"502": {Status: 502},
		"404": {Status: 404},
	} {
		f := &fakeT{by: map[string]forge.Response{"/user": resp}}
		g := newGH(f, time.Minute, &now)
		if id, err := g.CurrentUserID(t.Context()); err == nil || id != 0 {
			t.Errorf("%s: id=%d err=%v, want unknown", name, id, err)
		}
	}
	if _, err := newGH(&fakeT{by: map[string]forge.Response{}}, time.Minute, &now).CurrentUserID(t.Context()); errors.Is(err, nil) {
		t.Error("an unreachable lookup must be an error")
	}
}
