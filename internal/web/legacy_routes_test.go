package web

// legacy_routes_test.go — SPEC-LAUNCHER-ENTRY-FLAGS-001 M9 (card t1399),
// AC-019 and AC-018's per-file bound: the retired screen path is kept as a
// redirect and nothing else. The assertions are about what the route does, not
// about the handler's shape, so a legacy handler that still renders a page
// fails here even though the file-level word allowlist would accept it.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestLegacyKanbanRouteRedirects — AC-019: GET /kanban answers a 3xx whose
// Location is /factory and whose body is not a page; a query string rides
// along so an old bookmark keeps its profile; other methods stay refused; the
// target itself serves the screen.
func TestLegacyKanbanRouteRedirects(t *testing.T) {
	a := newTestApp(t)
	h := a.routes()

	t.Run("GET redirects to /factory with no page body", func(t *testing.T) {
		rec := serveGet(t, h, "/kanban")
		if rec.Code != http.StatusMovedPermanently {
			t.Fatalf("GET /kanban status = %d, want 301 (a retired path is cached by the browser)\nbody:\n%s", rec.Code, rec.Body.String())
		}
		if got := rec.Header().Get("Location"); got != "/factory" {
			t.Errorf("GET /kanban Location = %q, want /factory", got)
		}
		body := rec.Body.String()
		for _, page := range []string{"<html", "<body", "data-live", "Chain session board"} {
			if strings.Contains(body, page) {
				t.Errorf("the redirect response carries page content %q:\n%s", page, body)
			}
		}
		if len(body) > 200 {
			t.Errorf("the redirect response body is %d bytes, want the short redirect note only", len(body))
		}
	})

	t.Run("the query string is carried to the target", func(t *testing.T) {
		rec := serveGet(t, h, "/kanban?profile=work")
		if got := rec.Header().Get("Location"); got != "/factory?profile=work" {
			t.Errorf("GET /kanban?profile=work Location = %q, want /factory?profile=work", got)
		}
	})

	t.Run("other methods are refused", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/kanban", nil)
		req.Header.Set("Sec-Fetch-Site", "same-origin") // past the cross-site check, so the method decides
		req.Host = "127.0.0.1:8080"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST /kanban status = %d, want 405", rec.Code)
		}
		if rec.Header().Get("Location") != "" {
			t.Errorf("POST /kanban carries a Location header %q", rec.Header().Get("Location"))
		}
	})

	t.Run("the target serves the screen", func(t *testing.T) {
		rec := serveGet(t, h, "/factory")
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /factory status = %d, want 200", rec.Code)
		}
	})
}
