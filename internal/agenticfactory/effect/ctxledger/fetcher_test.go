package ctxledger

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

type stepT struct {
	calls []forge.Request
	resp  []forge.Response
}

func (s *stepT) Do(_ context.Context, r forge.Request) (forge.Response, error) {
	s.calls = append(s.calls, r)
	i := len(s.calls) - 1
	if i >= len(s.resp) {
		i = len(s.resp) - 1
	}
	return s.resp[i], nil
}

func fetcher(tr forge.Transport, now time.Time) ForgeFetcher {
	return ForgeFetcher{Repo: "o/r", Client: &forge.Client{
		T: tr, B: &forge.Budget{}, Now: func() time.Time { return now },
		Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3},
	}}
}

func TestForgeFetcherUsesTheEtagAndReportsNotModified(t *testing.T) {
	tr := &stepT{resp: []forge.Response{
		{Status: 200, Header: map[string]string{"etag": `"v1"`}, Body: []byte(`{"id":501,"body":"hello"}`)},
		{Status: 304},
	}}
	f := fetcher(tr, time.Now())
	c, _, nm, err := f.Fetch(t.Context(), 7, 501, "")
	if err != nil || nm || c.Body != "hello" || c.ID != 501 {
		t.Fatalf("first: %+v nm=%v err=%v", c, nm, err)
	}
	_, _, nm, err = f.Fetch(t.Context(), 7, 501, "")
	if err != nil || !nm {
		t.Fatalf("second: nm=%v err=%v", nm, err)
	}
	if tr.calls[1].Header["if-none-match"] != `"v1"` || tr.calls[1].Method != "GET" || tr.calls[1].Path != "/repos/o/r/issues/comments/501" {
		t.Errorf("second request = %+v", tr.calls[1])
	}
}

func TestForgeFetcherDeferredIsALimitNotAnAnswer(t *testing.T) {
	now := time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)
	tr := &stepT{resp: []forge.Response{{Status: 429, Header: map[string]string{"retry-after": "30"}}}}
	_, _, _, err := fetcher(tr, now).Fetch(t.Context(), 7, 501, "")
	var le *effect.LimitedError
	if !errors.As(err, &le) || le.RetryAt.Sub(now) != 30*time.Second {
		t.Errorf("err = %v", err)
	}
}

func TestForgeFetcherPermissionDenialIsTyped(t *testing.T) {
	tr := &stepT{resp: []forge.Response{{Status: 403, Body: []byte(`{"message":"Resource not accessible"}`)}}}
	_, _, _, err := fetcher(tr, time.Now()).Fetch(t.Context(), 7, 501, "")
	if !errors.Is(err, forge.ErrDenied) {
		t.Errorf("err = %v", err)
	}
}
