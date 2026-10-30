package web

// context_model_test.go — the ctxmodel ledger on the monitor (Issue #1948).
//
// Defect answer: these tests catch a missing, old-shape, or foreign-Issue model
// rendered as data, a pending/stale/restricted/unverified item shown as
// available, a JSON key renamed away from ctxmodel's, and the model view
// writing or building a link the model does not hold.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel"
)

const (
	cmSHA  = "0123456789abcdef0123456789abcdef01234567"
	cmHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	cmRows = `{"rows":[{"Project":"p","Issue":7,"State":{"value":"verifying","confirmed":true}}]}`
)

type okVerifier struct{}

func (okVerifier) Lookup(string, string) (string, bool, error) { return cmHash, true, nil }

func cmApp(t *testing.T, model func(context.Context, int) ([]byte, error)) *app {
	t.Helper()
	a := prApp(t, cmRows, nil)
	a.contextModel = model
	return a
}

func cmRow(t *testing.T, a *app) PRTrackRowVM {
	t.Helper()
	vm := a.buildPRTracking(context.Background())
	if len(vm.Rows) != 1 {
		t.Fatalf("rows = %+v", vm.Rows)
	}
	return vm.Rows[0]
}

func realLedgerJSON(t *testing.T) []byte {
	t.Helper()
	l := &ctxmodel.Ledger{Issue: 7}
	add := func(it ctxmodel.Item) {
		if err := l.Add(it); err != nil {
			t.Fatal(err)
		}
	}
	base := ctxmodel.Item{Section: ctxmodel.SectionSpec, Title: "spec", Path: ".moai/specs/SPEC-X/spec.md", Commit: cmSHA, TargetHead: cmSHA, SHA256: cmHash, Owner: "lane-18", Revision: "r2"}
	add(base)
	for _, to := range []ctxmodel.LinkState{ctxmodel.StateMasked, ctxmodel.StateCommitted, ctxmodel.StatePushed} {
		if err := l.Advance(0, to); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.VerifyRemote(0, okVerifier{}); err != nil {
		t.Fatal(err)
	}
	add(ctxmodel.Item{Section: ctxmodel.SectionPlan, Title: "plan", Path: ".moai/specs/SPEC-X/plan.md", Commit: cmSHA, SHA256: cmHash, Owner: "lane-18"}) // PREPARED
	add(ctxmodel.Item{Section: ctxmodel.SectionAudit, Title: "audit", Owner: "lane-20", Pending: true})
	stale := base
	stale.Section, stale.Title = ctxmodel.SectionDecision, "decision"
	add(stale)
	l.Items[3].State, l.Items[3].Stale = ctxmodel.StateRemoteVerified, true
	b, err := l.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestContextModel_States(t *testing.T) {
	real := realLedgerJSON(t)
	old := []byte(`{"issue":7,"items":[{"section":"spec","title":"old"}]}`)
	cases := []struct {
		name     string
		model    func(context.Context, int) ([]byte, error)
		known    bool
		note     string
		nItems   int
		statuses []string
	}{
		{name: "no model wired", model: nil, note: "not provided"},
		{name: "read error", model: func(context.Context, int) ([]byte, error) { return nil, errors.New("down") }, note: "model unreadable"},
		{name: "malformed", model: func(context.Context, int) ([]byte, error) { return []byte("{"), nil }, note: "model malformed"},
		{name: "another issue", model: func(context.Context, int) ([]byte, error) { return []byte(`{"issue":9,"items":[]}`), nil }, note: "another Issue"},
		{name: "old model", model: func(context.Context, int) ([]byte, error) { return old, nil }, known: true, nItems: 1, statuses: []string{"not-linkable"}},
		{name: "real model", model: func(context.Context, int) ([]byte, error) { return real, nil }, known: true, nItems: 4,
			statuses: []string{"available", "not-linkable", "pending", "stale"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cmRow(t, cmApp(t, tc.model)).Ledger
			if got.Known != tc.known || !strings.Contains(got.Note, tc.note) || len(got.Items) != tc.nItems {
				t.Fatalf("ledger = %+v", got)
			}
			for i, want := range tc.statuses {
				if got.Items[i].Status != want {
					t.Errorf("item %d status = %q, want %q", i, got.Items[i].Status, want)
				}
			}
		})
	}
}

func TestContextModel_FieldsAndGaps(t *testing.T) {
	real := realLedgerJSON(t)
	items := cmRow(t, cmApp(t, func(context.Context, int) ([]byte, error) { return real, nil })).Ledger.Items
	spec := items[0]
	if spec.Commit != cmSHA[:12] || spec.Hash != cmHash[:12] || spec.State != "REMOTE_VERIFIED" || spec.Owner != "lane-18" || spec.Revision != "r2" || spec.Path != ".moai/specs/SPEC-X/spec.md" {
		t.Errorf("spec item = %+v", spec)
	}
	if p := items[2]; p.Path != prGap || p.Commit != prGap || p.State != prGap {
		t.Errorf("a pending item has no link fields; they must be gaps: %+v", p)
	}
	old := cmRow(t, cmApp(t, func(context.Context, int) ([]byte, error) {
		return []byte(`{"issue":7,"items":[{"section":"spec","title":"old"}]}`), nil
	})).Ledger.Items[0]
	if old.Owner != prGap || old.Hash != prGap || old.Revision != prGap || old.VerifiedAt != prGap {
		t.Errorf("an old model must show gaps, not values: %+v", old)
	}
}

func TestContextModel_RedactsAndRenders(t *testing.T) {
	secret := "ghp_" + strings.Repeat("a", 36)
	model := []byte(`{"issue":7,"items":[{"section":"spec","title":"` + secret + `","owner":"lane-18"}]}`)
	a := cmApp(t, func(context.Context, int) ([]byte, error) { return model, nil })
	m, _ := a.buildMonitor(time.Now())
	html := renderTempl(t, Monitor(tg1ShellVM(), m))
	if strings.Contains(html, secret) {
		t.Fatal("a secret in a ledger title reached the page")
	}
	if !strings.Contains(html, "model spec") {
		t.Errorf("model row missing from the context panel")
	}
	panel := html[strings.Index(html, `id="context-ledger"`):]
	if end := strings.Index(panel, "</section>"); end > 0 {
		panel = panel[:end]
	}
	if strings.Contains(panel, "<a ") && strings.Contains(panel, "model spec") {
		// links may come only from the monitor's verified items, never the model
		if strings.Contains(panel[strings.Index(panel, "model spec")-300:], "<a href") {
			t.Errorf("model view built a link")
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/api/pr-tracking", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != 200 || strings.Contains(rec.Body.String(), secret) {
		t.Fatalf("api = %d (secret leaked: %v)", rec.Code, strings.Contains(rec.Body.String(), secret))
	}
}

// TestContextModel_NilSeamIsAGapNotAnError pins the start-up contract: a
// console started without the context source (Config.ContextModel nil, the
// Agentic Factory not configured) renders the ledger view as the gap
// "not provided" — not as "unreadable", not as a failed page or API call. A
// source that exists but fails stays visibly different ("unreadable").
func TestContextModel_NilSeamIsAGapNotAnError(t *testing.T) {
	cases := []struct {
		name       string
		model      func(context.Context, int) ([]byte, error)
		wantText   string
		rejectText string
	}{
		{"nil seam", nil, "model not provided", "unreadable"},
		{"failing seam", func(context.Context, int) ([]byte, error) { return nil, errors.New("down") }, "model model unreadable: down", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := cmApp(t, tc.model)
			m, err := a.buildMonitor(time.Now())
			if err != nil {
				t.Fatalf("buildMonitor: %v", err)
			}
			html := renderTempl(t, Monitor(tg1ShellVM(), m))
			if !strings.Contains(html, tc.wantText) {
				t.Errorf("monitor missing %q", tc.wantText)
			}
			if tc.rejectText != "" && strings.Contains(html, tc.rejectText) {
				t.Errorf("monitor shows %q for a source that is simply not wired", tc.rejectText)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/pr-tracking", nil)
			req.Host = "127.0.0.1:8080"
			rec := httptest.NewRecorder()
			a.routes().ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("api = %d", rec.Code)
			}
		})
	}
	// Nothing wired at all (no monitor rows either): the page still builds.
	bare := newApp(Config{ProjectRoot: t.TempDir()})
	if _, err := bare.buildMonitor(time.Now()); err != nil {
		t.Fatalf("a console with no sources must still build: %v", err)
	}
	if vm := bare.buildContextModel(context.Background(), 7); vm.Known || vm.Note != "not provided" {
		t.Fatalf("unwired model = %+v", vm)
	}
}
