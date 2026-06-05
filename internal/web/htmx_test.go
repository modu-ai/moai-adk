package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// SPEC-WEB-CONSOLE-006 M4 — HTMX foundation. These are ADDITIVE assertions
// (REQ-WC6-001 permits a new assertion for the HTMX foundation; they do NOT
// modify any Class A/B test) verifying htmx.min.js is self-hosted + served +
// linked, the form carries only progressive-enhancement hx-* (no partial-swap),
// and the offline zero-network invariant holds.

// TestHtmxEmbeddedAndServed verifies AC-WC6-006/015: htmx.min.js is embedded into
// the binary, served from /static/htmx.min.js (200, offline, not Host-gated on
// GET), and is the real htmx library (not an empty/placeholder file).
func TestHtmxEmbeddedAndServed(t *testing.T) {
	// Embedded under assets/ and reachable via the static FS.
	js := readEmbeddedAsset(t, "htmx.min.js")
	if !strings.Contains(js, "htmx") {
		t.Error("embedded htmx.min.js does not look like the htmx library")
	}
	if len(js) < 10*1024 {
		t.Errorf("embedded htmx.min.js is only %d bytes — looks truncated/placeholder", len(js))
	}

	// Served from /static/htmx.min.js (200, offline). GET on a static asset is not
	// Host-gated, so a foreign host still reaches it.
	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/static/htmx.min.js", nil)
	req.Host = "evil.example.com"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /static/htmx.min.js status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Error("/static/htmx.min.js served body is empty")
	}
}

// TestHtmxLinkedBeforeAppJS verifies AC-WC6-015: the page links
// /static/htmx.min.js, and it loads before /static/app.js (so app.js can rely on
// htmx being present).
func TestHtmxLinkedBeforeAppJS(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	htmxIdx := strings.Index(body, `src="/static/htmx.min.js"`)
	appIdx := strings.Index(body, `src="/static/app.js"`)
	if htmxIdx < 0 {
		t.Fatal("rendered page does not link /static/htmx.min.js")
	}
	if appIdx < 0 {
		t.Fatal("rendered page does not link /static/app.js")
	}
	if htmxIdx > appIdx {
		t.Error("/static/htmx.min.js is linked AFTER /static/app.js — it must load before app.js")
	}
}

// TestHtmxFoundationNoPartialSwap verifies AC-WC6-015 + E.3: the form carries only
// progressive-enhancement hx-* (hx-boost) and NO section-scoped partial-swap
// attributes (hx-target / hx-swap / hx-get / a section-scoped hx-post) — those are
// deferred to SPEC-WEB-CONSOLE-007. The full-page POST contract is preserved.
func TestHtmxFoundationNoPartialSwap(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{})

	// Progressive enhancement present: the form is hx-boost'd.
	if !strings.Contains(body, `hx-boost="true"`) {
		t.Error("form does not carry the progressive-enhancement hx-boost attribute")
	}

	// No section-scoped partial-swap attributes (deferred to 007).
	for _, forbidden := range []string{"hx-target", "hx-swap", "hx-get"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("rendered page carries the partial-swap attribute %q — section-scoped partial-swap is deferred to SPEC-WEB-CONSOLE-007 (E.3)", forbidden)
		}
	}
	// The form still posts to /save as a full-page form (method=POST action=/save).
	if !strings.Contains(body, `method="POST"`) || !strings.Contains(body, `action="/save`) {
		t.Error("the form is no longer a full-page POST to /save — the full-page contract must be preserved")
	}
}

// TestHtmxOfflineZeroNetwork verifies AC-WC6-006: the served frontend assets
// (htmx.min.js + the page-linked scripts/styles) contain zero external https
// font/style/script URL (no CDN — htmx is self-hosted, not fetched from unpkg /
// jsdelivr / cdnjs).
func TestHtmxOfflineZeroNetwork(t *testing.T) {
	htmx := readEmbeddedAsset(t, "htmx.min.js")
	for _, forbidden := range []string{"unpkg.com", "cdn.jsdelivr", "cdnjs.cloudflare", "fonts.googleapis.com"} {
		if strings.Contains(htmx, forbidden) {
			t.Errorf("htmx.min.js contains external CDN reference %q (offline invariant broken)", forbidden)
		}
	}

	// The rendered page must not link an external htmx CDN — htmx is /static/-served.
	body := renderIndexBody(t, profile.ProfilePreferences{})
	for _, forbidden := range []string{
		`src="https://unpkg.com/htmx`,
		`src="https://cdn.jsdelivr`,
		`src="https://cdnjs`,
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("rendered page links an external htmx CDN %q (must be self-hosted /static/htmx.min.js)", forbidden)
		}
	}
}
