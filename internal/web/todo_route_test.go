package web

// todo_route_test.go — SPEC-WEB-TODO-QUEUE-001 M2: the /todo route surface.
//
// Resolved decision G-4 chose a top-level route over a panel on /kanban, so the
// cost is in scope and pinned here: one route, a sixth navigation row marked as
// the current location, an iconAt case (a missing one renders a blank glyph
// rather than an error), and nav.todo in all four locale maps.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
)

// getTodoPage requests GET /todo through the full route tree and returns the
// recorder.
func getTodoPage(t *testing.T, a *app) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/todo", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	return rec
}

// navRowHrefRe captures the href of every rendered rail navigation row.
var navRowHrefRe = regexp.MustCompile(`<a class="nav__row" href="([^"]*)"`)

// TestTodoRouteServesPage — AC-WTQ-002 first half: GET /todo answers 200.
func TestTodoRouteServesPage(t *testing.T) {
	a := newTestApp(t)

	rec := getTodoPage(t, a)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /todo status = %d, want 200\nbody:\n%s", rec.Code, rec.Body.String())
	}
}

// TestTodoRouteRejectsNonGET keeps the route inside the read-only posture the
// other four screens hold (REQ-WTQ-001).
func TestTodoRouteRejectsNonGET(t *testing.T) {
	a := newTestApp(t)

	// Two layers refuse a non-GET request, and this names which does what rather
	// than accepting either code for any method. An earlier form checked POST
	// alone and accepted "405 or 403", which proves neither layer — 403 is also
	// what a request that never reached the method gate returns.

	// POST/PUT/PATCH are stopped FIRST by the CSRF guard (hostCheckMiddleware,
	// app.go), which refuses before routing and is the stronger protection.
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
		req := httptest.NewRequest(method, "/todo", nil)
		rec := httptest.NewRecorder()
		a.routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("%s /todo with no same-origin header: status = %d, want %d from the CSRF guard",
				method, rec.Code, http.StatusForbidden)
		}
	}

	// With that guard satisfied, the route's own read-only gate must still refuse
	// them, with the header naming what is allowed. This is the half the earlier
	// form could not reach. DELETE is outside the guard's switch, so it arrives
	// here either way.
	for _, method := range []string{
		http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
	} {
		req := httptest.NewRequest(method, "/todo", nil)
		req.Host = "127.0.0.1:8080"
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		rec := httptest.NewRecorder()
		a.routes().ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /todo status = %d, want %d", method, rec.Code, http.StatusMethodNotAllowed)
		}
		if got := rec.Header().Get("Allow"); got != http.MethodGet {
			t.Errorf("%s /todo Allow header = %q, want %q", method, got, http.MethodGet)
		}
	}
}

// TestTodoNavRowIsSixthAndCurrent — AC-WTQ-002 second half: the rail carries
// six rows, the sixth links to /todo, and it is marked as the current location
// while /todo is being served (REQ-WTQ-002).
func TestTodoNavRowIsSixthAndCurrent(t *testing.T) {
	a := newTestApp(t)

	body := getTodoPage(t, a).Body.String()

	hrefs := navRowHrefRe.FindAllStringSubmatch(body, -1)
	if len(hrefs) != 6 {
		got := make([]string, 0, len(hrefs))
		for _, m := range hrefs {
			got = append(got, m[1])
		}
		t.Fatalf("rail carries %d navigation rows (%v), want 6", len(hrefs), got)
	}
	if hrefs[5][1] != "/todo" {
		t.Errorf("sixth navigation row links to %q, want \"/todo\"", hrefs[5][1])
	}
	// The six rows keep their existing order with todo appended sixth.
	want := []string{"/", "/kanban", "/specs", "/monitor", "/settings", "/todo"}
	for i, w := range want {
		if hrefs[i][1] != w {
			t.Errorf("navigation row %d links to %q, want %q", i, hrefs[i][1], w)
		}
	}
	if !strings.Contains(body, `href="/todo" aria-current="page"`) {
		t.Errorf("the /todo row is not marked aria-current=\"page\" while /todo is served")
	}
}

// TestTodoNavRowNotCurrentElsewhere — the aria-current marking is conditional
// on the route, which is what REQ-WTQ-002's "while" clause states.
func TestTodoNavRowNotCurrentElsewhere(t *testing.T) {
	a := newTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/kanban", nil)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)

	if strings.Contains(rec.Body.String(), `href="/todo" aria-current="page"`) {
		t.Errorf("the /todo row is marked current while /kanban is served")
	}
}

// TestTodoIconCaseExists — AC-WTQ-002 third half: navRow calls @iconAt(id, 16)
// with the nav id, so a missing case renders a blank glyph rather than an
// error. Measured baseline on the pre-change tree: 0.
func TestTodoIconCaseExists(t *testing.T) {
	src, err := os.ReadFile("icons.templ")
	if err != nil {
		t.Fatalf("read icons.templ: %v", err)
	}
	body := string(src)
	start := strings.Index(body, "templ iconAt")
	if start < 0 {
		t.Fatal("icons.templ has no iconAt template")
	}
	if n := strings.Count(body[start:], `case "todo":`); n != 1 {
		t.Fatalf("iconAt carries %d `case \"todo\"` arms, want 1", n)
	}
}

// TestTodoNavI18nInFourLocales — AC-WTQ-011: nav.todo is present once per
// locale map. Measured baseline on the pre-change tree: 0.
func TestTodoNavI18nInFourLocales(t *testing.T) {
	src, err := os.ReadFile("assets/i18n.js")
	if err != nil {
		t.Fatalf("read i18n.js: %v", err)
	}
	if n := strings.Count(string(src), `"nav.todo"`); n != 4 {
		t.Fatalf("i18n.js carries %d \"nav.todo\" entries, want 4 (en, ko, ja, zh)", n)
	}
}
