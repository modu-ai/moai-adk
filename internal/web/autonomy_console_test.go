package web

// SPEC-AUTONOMY-TIERS-001 M10 — web toggle reachable from the main console page.
//
// Gap-2 closure (console 완성): the /autonomy/tiers endpoint + toggle fragment
// (M8) exist and are tested, but the toggle is NOT reachable from the main
// console page — a user navigating `moai web` never sees it. M10 makes the
// toggle discoverable: the main console page (GET /) render MUST include a
// reference that reaches the autonomy toggle, so a user on the main page can
// see/reach it.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandleIndex_ReachableFromMainConsole asserts the main console page (GET /)
// render includes a reference to the autonomy toggle — i.e. the page reaches
// /autonomy/tiers so a user on `moai web` can discover and navigate to the
// toggle. Without M10 the GET / body carries no autonomy reference at all.
func TestHandleIndex_ReachableFromMainConsole(t *testing.T) {
	a := newTestApp(t)
	h := a.routes()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "/autonomy/tiers") {
		t.Errorf("main console page MUST reference /autonomy/tiers so the toggle is reachable; body did not contain it.\nfirst 400 bytes: %s",
			truncate(string(body), 400))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
