package web

// SPEC-MCP-CONSOLE-001 M3 — codex authentication surface tests.
//
// These tests cover AC-C-006 (probe state displayed, not recomputed),
// AC-C-007 (unauthenticated → names the command, no login route),
// and AC-C-008 (codex absent → graceful not-installed state).
//
// AC-C-009 (opt-in toggles write the seam the gates read) is tested in
// internal/cli/mcp_codex_consoletest_test.go because the fail-closed readers
// (readCodexReviewGateEnabled / readCodexTaskAllowWrite) live in package cli.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// renderAppBody renders GET / for the given app and returns the HTML body.
func renderAppBody(t *testing.T, a *app) string {
	t.Helper()
	h := a.routes()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

// codexTestApp returns a newTestApp with the codex probe pre-wired to the given
// view, so each test controls exactly what the console sees without depending
// on the host's real codex install state.
func codexTestApp(t *testing.T, state CodexStateView) *app {
	t.Helper()
	a := newTestApp(t)
	a.codexStateProbe = func(_ context.Context) CodexStateView { return state }
	return a
}

// ─── AC-C-006: probe state displayed, not recomputed ─────────────────────

// TestAC_C_006_ProbeStateDisplayed asserts that when codex is installed and
// authenticated, the rendered page shows the probe-reported state (binary,
// version, auth_provider). The console must DISPLAY these — it must not
// recompute them.
func TestAC_C_006_ProbeStateDisplayed(t *testing.T) {
	state := CodexStateView{
		Installed:        true,
		Binary:           "/usr/local/bin/codex",
		Version:          "1.2.3",
		AuthProvider:     codexAuthChatGPT,
		EnableReviewGate: true,
		AllowWrite:       false,
	}
	body := renderAppBody(t, codexTestApp(t, state))

	for _, want := range []string{
		"/usr/local/bin/codex",
		"1.2.3",
		"ChatGPT", // codexAuthProviderLabel maps the chatgpt token to this display spelling
	} {
		if !strings.Contains(body, want) {
			t.Errorf("AC-C-006: rendered page missing probe state %q", want)
		}
	}
}

// TestAC_C_006_NoSecondClassifierInWeb is the grep guard from acceptance.md
// AC-C-006: internal/web must NOT contain a second auth-classification
// implementation. The classification lives in internal/cli.classifyCodexAuth;
// web consumes the probe result via injection.
func TestAC_C_006_NoSecondClassifierInWeb(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob *.go: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		// The codex_state.go constants (codexAuthChatGPT etc.) are display-layer
		// tokens, NOT classification logic. The forbidden patterns are the
		// function name and the raw `codex login status` command string — those
		// are the classification mechanism.
		if strings.Contains(string(data), "classifyCodexAuth") {
			t.Errorf("AC-C-006: internal/web/%s references %q — web must not fork the classifier", f, "classifyCodexAuth")
		}
		if strings.Contains(string(data), "codex login status") {
			t.Errorf("AC-C-006: internal/web/%s contains %q — web must not run the auth probe", f, "codex login status")
		}
	}
}

// ─── AC-C-007: unauthenticated → names the command, no login flow ────────

// TestAC_C_007_UnauthenticatedNamesCommand asserts that when codex is installed
// but unauthenticated (auth_provider = unknown), the page surfaces the
// remediation instruction naming `codex login`.
func TestAC_C_007_UnauthenticatedNamesCommand(t *testing.T) {
	state := CodexStateView{
		Installed:    true,
		Binary:       "/usr/local/bin/codex",
		AuthProvider: codexAuthUnknown,
	}
	body := renderAppBody(t, codexTestApp(t, state))
	if !strings.Contains(body, "codex login") {
		t.Error("AC-C-007: unauthenticated state must name the `codex login` command as remediation")
	}
}

// TestAC_C_007_NoAuthRouteInApp asserts that internal/web/app.go performs no
// login, no OAuth redirect, and no browser launch for authentication. The
// console reports state and names the command; login is the user's action.
func TestAC_C_007_NoAuthRouteInApp(t *testing.T) {
	data, err := os.ReadFile("app.go")
	if err != nil {
		t.Fatalf("read app.go: %v", err)
	}
	body := string(data)
	for _, forbidden := range []string{
		"oauth",
		"/auth/",
		"login/redirect",
		"codex login\n", // executing login (not naming it as instruction)
	} {
		if strings.Contains(strings.ToLower(body), forbidden) {
			t.Errorf("AC-C-007: app.go contains %q — the console must not perform or redirect to an auth flow", forbidden)
		}
	}
	// No new route performs an external auth redirect: verify no HandleFunc
	// registration mentions auth/login/oauth.
	for _, pattern := range []string{`"login"`, `"/auth"`, `"oauth"`} {
		if strings.Contains(body, pattern) {
			t.Errorf("AC-C-007: app.go registers a route matching %q — no auth route is permitted", pattern)
		}
	}
}

// ─── AC-C-008: codex absent → graceful not-installed state ───────────────

// TestAC_C_008_CodexAbsentGraceful asserts that when codex is not installed,
// the page renders the not-installed state without erroring (200 OK), matching
// the probe's fail-open behavior (installed: false, auth_provider: unknown).
func TestAC_C_008_CodexAbsentGraceful(t *testing.T) {
	state := CodexStateView{
		Installed:    false,
		AuthProvider: codexAuthUnknown,
	}
	a := codexTestApp(t, state)
	h := a.routes()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("AC-C-008: codex-absent GET / status = %d, want 200 (must not error)", rec.Code)
	}
	body := rec.Body.String()
	// The not-installed indicator must be present. We check for a data attribute
	// or i18n key that the template emits for the not-installed branch.
	if !strings.Contains(body, "codex-not-installed") && !strings.Contains(body, "Not installed") && !strings.Contains(body, "not-installed") {
		t.Error("AC-C-008: codex-absent page must show the not-installed state")
	}
}

// TestAC_C_008_DefaultProbeIsFailOpen asserts that the default codexStateProbe
// (used when no probe is injected) returns a not-installed/unknown state,
// matching the real probe's fail-open for the codex-absent case.
func TestAC_C_008_DefaultProbeIsFailOpen(t *testing.T) {
	state := defaultCodexStateProbe(context.Background())
	if state.Installed {
		t.Error("AC-C-008: default probe must report Installed=false (fail-open)")
	}
	if state.AuthProvider != codexAuthUnknown {
		t.Errorf("AC-C-008: default probe AuthProvider = %q, want %q (fail-open)", state.AuthProvider, codexAuthUnknown)
	}
}
