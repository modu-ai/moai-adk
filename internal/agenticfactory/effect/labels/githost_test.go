package labels

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/effect"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/forge"
)

// miniGitHub is a fake label API behind forge.Transport. It records every request and talks to nobody.
type miniGitHub struct {
	labels map[string]map[string]string // lower name -> {name,color,description}
	reqs   []forge.Request
	status map[string]int // "METHOD path-prefix" -> forced status
	header map[string]string
}

func newMini() *miniGitHub {
	return &miniGitHub{labels: map[string]map[string]string{}, status: map[string]int{}}
}

func (m *miniGitHub) Do(_ context.Context, r forge.Request) (forge.Response, error) {
	m.reqs = append(m.reqs, r)
	for k, st := range m.status {
		parts := strings.SplitN(k, " ", 2)
		if r.Method == parts[0] && strings.HasPrefix(r.Path, parts[1]) {
			return forge.Response{Status: st, Header: m.header, Body: []byte(`{"message":"forced"}`)}, nil
		}
	}
	base := "/repos/o/r/labels"
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.Path, base):
		u, _ := url.Parse(r.Path)
		page := u.Query().Get("page")
		var all []map[string]string
		for _, l := range m.labels {
			all = append(all, l)
		}
		if page != "" && page != "1" { // everything fits on page 1 except when a test seeds more
			all = nil
		}
		raw, _ := json.Marshal(all)
		return forge.Response{Status: 200, Body: raw}, nil
	case r.Method == http.MethodPost && r.Path == base:
		var b map[string]string
		_ = json.Unmarshal(r.Body, &b)
		m.labels[strings.ToLower(b["name"])] = b
		return forge.Response{Status: 201, Body: []byte(`{}`)}, nil
	case r.Method == http.MethodPatch && strings.HasPrefix(r.Path, base+"/"):
		name, _ := url.PathUnescape(strings.TrimPrefix(r.Path, base+"/"))
		var b map[string]string
		_ = json.Unmarshal(r.Body, &b)
		cur, ok := m.labels[strings.ToLower(name)]
		if !ok {
			return forge.Response{Status: 404, Body: []byte(`{"message":"Not Found"}`)}, nil
		}
		if v, has := b["new_name"]; has {
			cur["name"] = v
		}
		cur["color"], cur["description"] = b["color"], b["description"]
		return forge.Response{Status: 200, Body: []byte(`{}`)}, nil
	}
	return forge.Response{Status: 404}, nil
}

func (m *miniGitHub) writes() (n int) {
	for _, r := range m.reqs {
		if r.Method != http.MethodGet {
			n++
		}
	}
	return
}

func ghHost(m *miniGitHub, allow bool, now time.Time) *GitHubHost {
	return &GitHubHost{
		Repo: "o/r", T: m, B: &forge.Budget{}, AllowWrite: allow, PerPage: 50,
		Backoff: forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3},
		Now:     func() time.Time { return now },
	}
}

var t0 = time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)

func TestGitHubHostEndToEndThroughTheApplier(t *testing.T) {
	m := newMini()
	m.labels["p0"] = map[string]string{"name": "p0", "color": "000000", "description": "old"}
	a := &Applier{Host: ghHost(m, true, t0), AllowWrite: true, Repo: "o/r", Store: effect.NewMemIntentStore(), Timeouts: effect.Timeouts{LabelApply: time.Second}}
	p, err := a.Plan(t.Context(), want)
	if err != nil {
		t.Fatal(err)
	}
	res, err := a.Apply(t.Context(), want, Approval{PlanDigest: p.Digest(), By: "operator"})
	if err != nil || !res.Verified {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	for _, r := range m.reqs {
		if r.Method == http.MethodDelete || !strings.HasPrefix(r.Path, "/repos/o/r/labels") {
			t.Errorf("unexpected request %s %s", r.Method, r.Path)
		}
	}
	// create shape and path escaping of a name with a colon
	var sawCreate, sawPatch bool
	for _, r := range m.reqs {
		if r.Method == http.MethodPost {
			var b map[string]string
			_ = json.Unmarshal(r.Body, &b)
			sawCreate = sawCreate || (b["name"] != "" && b["color"] != "")
		}
		if r.Method == http.MethodPatch && r.Path == "/repos/o/r/labels/p0" {
			sawPatch = true
		}
	}
	if !sawCreate || !sawPatch {
		t.Errorf("create=%v patch=%v reqs=%d", sawCreate, sawPatch, len(m.reqs))
	}
}

func TestGitHubHostRefusesWritesUnlessEnabledAndNeverDeletes(t *testing.T) {
	m := newMini()
	h := ghHost(m, false, t0)
	if err := h.Create(t.Context(), Label{Name: "x", Color: "ffffff"}); !errors.Is(err, ErrWriteDisabled) {
		t.Errorf("create err = %v", err)
	}
	if err := h.Update(t.Context(), "x", Label{Name: "x", Color: "ffffff"}); !errors.Is(err, ErrWriteDisabled) {
		t.Errorf("update err = %v", err)
	}
	if m.writes() != 0 || len(m.reqs) != 0 {
		t.Errorf("requests sent with writes disabled: %d", len(m.reqs))
	}
	for _, bad := range []forge.Request{
		{Method: http.MethodDelete, Path: "/repos/o/r/labels/x"},
		{Method: http.MethodPost, Path: "/repos/o/r/issues"},
		{Method: http.MethodPatch, Path: "/repos/other/r/labels/x"},
		{Method: http.MethodGet, Path: "/repos/o/r/collaborators"},
	} {
		if h.allowed(bad) {
			t.Errorf("allowed %s %s", bad.Method, bad.Path)
		}
	}
}

func TestGitHubHostPaginatesWithAConfiguredPageSize(t *testing.T) {
	m := newMini()
	h := ghHost(m, false, t0)
	if _, err := h.List(t.Context()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(m.reqs[0].Path, "per_page=50") || !strings.Contains(m.reqs[0].Path, "page=1") {
		t.Errorf("first request = %s", m.reqs[0].Path)
	}
	bad := ghHost(m, false, t0)
	bad.PerPage = 0
	if _, err := bad.List(t.Context()); err == nil {
		t.Error("an unconfigured page size was accepted")
	}
}

func TestGitHubHostRateLimitBlocksTheSharedBudgetAndSendsNothingMore(t *testing.T) {
	m := newMini()
	m.status["POST /repos/o/r/labels"] = 429
	m.header = map[string]string{"retry-after": "30"}
	h := ghHost(m, true, t0)
	err := h.Create(t.Context(), Label{Name: "x", Color: "ffffff"})
	var se *forge.StatusError
	if !errors.As(err, &se) || se.Status != 429 {
		t.Fatalf("err = %v", err)
	}
	n := len(m.reqs)
	err = h.Create(t.Context(), Label{Name: "y", Color: "ffffff"})
	var le *effect.LimitedError
	if !errors.As(err, &le) || le.RetryAt.Sub(t0) != 30*time.Second || len(m.reqs) != n {
		t.Errorf("err=%v requests %d -> %d", err, n, len(m.reqs))
	}
}

func TestGitHubHostPermissionDenialIsTypedForParking(t *testing.T) {
	m := newMini()
	m.status["POST /repos/o/r/labels"] = 403
	m.header = map[string]string{"x-ratelimit-remaining": "4000"}
	err := ghHost(m, true, t0).Create(t.Context(), Label{Name: "x", Color: "ffffff"})
	if mapped := forge.ForEffect(err, t0, forge.Backoff{MinSecondaryWait: time.Minute, Base: time.Second, Max: time.Minute, MaxAttempts: 3}); !errors.Is(mapped, effect.ErrDenied) {
		t.Errorf("mapped = %v", mapped)
	}
}
