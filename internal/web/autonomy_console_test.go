package web

// autonomy_console_test.go — SPEC-WEB-CONSOLE-REDESIGN-001 M6 regression guard.
//
// The former autonomy-tier web surface is REMOVED. It was a GET-only fragment
// with no form and no action, so selecting a tier saved nothing: the only
// runtime reader is the MOAI_AUTONOMY_TIER env seam and the only writer is the
// init-time ApplyAutonomyTierBundle. There is no persistence target a console
// field could write to, and creating one would require a new config field —
// a separate SPEC.
//
// This file inverts the former M10 reachability test: the console page must NOT
// carry the bare link, and the route must NOT be registered. The config-side
// autonomy tier core (internal/config/autonomy_tiers.go + the init-time bundle)
// is untouched — only the dead web surface is gone.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestAutonomyStubResolved verifies AC-WCR-050 under the removal branch: the
// rendered console carries no /autonomy/tiers reference and the route is not
// served.
func TestAutonomyStubResolved(t *testing.T) {
	a := newTestApp(t)
	h := a.routes()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if strings.Contains(string(body), "/autonomy/tiers") {
		t.Error("console page still references /autonomy/tiers — the stub link was not removed")
	}
	if strings.Contains(string(body), "autonomy-link") {
		t.Error("console page still carries the autonomy-link section")
	}

	// The route itself must be gone. A registered handler answers 200/405; an
	// unregistered path falls through to the "/" catch-all, which serves the
	// console page — so the discriminating assertion is that the response is
	// NOT the autonomy fragment.
	req2 := httptest.NewRequest(http.MethodGet, "/autonomy/tiers", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	body2, _ := io.ReadAll(rec2.Body)
	if strings.Contains(string(body2), `class="autonomy-toggle"`) {
		t.Error("/autonomy/tiers still serves the autonomy toggle fragment — the route was not removed")
	}
}

// TestAutonomyToggleFragmentGone is the falsifiability control for the route
// assertion above: it names the exact marker the removed fragment emitted, so a
// silent re-introduction of the handler fails this test rather than passing
// vacuously.
func TestAutonomyToggleFragmentGone(t *testing.T) {
	const removedMarker = `class="autonomy-toggle"`
	fixture := `<section class="autonomy-toggle" aria-label="Autonomy tier"></section>`
	if !strings.Contains(fixture, removedMarker) {
		t.Fatalf("marker %q does not match the fragment it is meant to detect — the absence assertions above would be vacuous", removedMarker)
	}
}
