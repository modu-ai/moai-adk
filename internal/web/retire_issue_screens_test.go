package web

// retire_issue_screens_test.go — P2 retirement of the Issue/mirror screens in
// internal/web (Issue #1948, local-only decision D1). The local board replaces
// them: /board is the landing screen, the old PR-tracking API and the monitor's
// Issue panels are gone, and the Config no longer takes Issue-shaped sources.
// The local settings and file screens stay.
//
// Defect answer: these tests catch an old Issue route that still answers, a
// monitor page that still shows PR tracking, intake requests or the context
// ledger, a Config that still accepts the Issue/mirror sources, a nav without
// the board or one that names an Issue, a root path that does not lead to the
// board, and a retirement that took the local settings with it.

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestRetire_OldIssueRouteIs404(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	if rec := serveGet(t, a.routes(), "/api/pr-tracking"); rec.Code != http.StatusNotFound {
		t.Errorf("GET /api/pr-tracking = %d, want 404 (the Issue API is retired)", rec.Code)
	}
	if rec := serveGet(t, a.routes(), "/api/browser"); rec.Code != http.StatusOK {
		t.Errorf("GET /api/browser = %d, want 200 (the saved browser choice is a local setting and stays)", rec.Code)
	}
}

func TestRetire_MonitorHasNoIssuePanels(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	rec := serveGet(t, a.routes(), "/monitor")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /monitor = %d", rec.Code)
	}
	body := rec.Body.String()
	for _, gone := range []string{"PR tracking", "Intake requests", "Context ledger"} {
		if strings.Contains(body, gone) {
			t.Errorf("the monitor still shows the retired %q panel", gone)
		}
	}
	for _, kept := range []string{"Sessions", "Goals", "Browser"} {
		if !strings.Contains(body, kept) {
			t.Errorf("the monitor lost its local %q panel", kept)
		}
	}
}

func TestRetire_ConfigTakesNoIssueSources(t *testing.T) {
	typ := reflect.TypeOf(Config{})
	for _, f := range []string{"MonitorRows", "IntakeRequests", "ContextModel"} {
		if _, ok := typ.FieldByName(f); ok {
			t.Errorf("Config still has the Issue/mirror source %s", f)
		}
	}
	if _, ok := typ.FieldByName("Ledger"); !ok {
		t.Errorf("Config must keep the ledger source")
	}
}

func TestRetire_NavLeadsWithTheBoardAndNamesNoIssue(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	body := serveGet(t, a.routes(), "/settings").Body.String()
	start := strings.Index(body, `class="nav"`)
	if start < 0 {
		t.Fatalf("no nav in the page")
	}
	nav := body[start:]
	if end := strings.Index(nav, "</ul>"); end > 0 {
		nav = nav[:end]
	}
	iBoard, iTodo := strings.Index(nav, `href="/board"`), strings.Index(nav, `href="/todo"`)
	if iBoard < 0 || iTodo < 0 || iBoard > iTodo {
		t.Errorf("the board must be the first nav row (board at %d, todo at %d)", iBoard, iTodo)
	}
	if strings.Contains(strings.ToLower(nav), "issue") {
		t.Errorf("the nav must not name an Issue: %q", nav)
	}
}

func TestRetire_RootLeadsToTheBoardAndOverviewMoves(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	rec := serveGet(t, a.routes(), "/")
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/board" {
		t.Errorf("GET / = %d Location %q, want 302 /board", rec.Code, rec.Header().Get("Location"))
	}
	if rec := serveGet(t, a.routes(), "/overview"); rec.Code != http.StatusOK {
		t.Errorf("GET /overview = %d, want 200 (the overview moves, it is not removed)", rec.Code)
	}
	if rec := serveGet(t, a.routes(), "/no-such-screen"); rec.Code != http.StatusNotFound {
		t.Errorf("an unregistered path = %d, want 404", rec.Code)
	}
	if rec := servePost(t, a.routes(), "/", nil); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST / = %d, want 405", rec.Code)
	}
}

func TestRetire_LocalSettingsScreenStays(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	for _, path := range []string{"/settings", "/todo", "/specs", "/factory", "/board"} {
		if rec := serveGet(t, a.routes(), path); rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, rec.Code)
		}
	}
}
