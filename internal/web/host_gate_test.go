package web

// host_gate_test.go — AC-WC-009 as amended (card t613). The loopback Host check
// applies to every method on every route, /static/ included; the Sec-Fetch-Site
// same-origin check stays scoped to state-changing methods.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// hostGateRefusal is the text the Host gate writes on a refusal. Asserting it
// separates a Host-gate 403 from any other 403 (the CSRF gate, a handler).
const hostGateRefusal = "non-loopback Host header"

// hostGateCases is the Host column of the AC-WC-009 control matrix. Each allowed
// Host appears in its port-less and its with-port form. The expected outcome is
// written out rather than derived from isLoopbackHost, so the test does not use
// the code under test as its own oracle.
var hostGateCases = []struct {
	label   string
	host    string
	allowed bool
}{
	{"localhost", "localhost", true},
	{"localhost-port", "localhost:3041", true},
	{"ipv4", "127.0.0.1", true},
	{"ipv4-port", "127.0.0.1:3041", true},
	{"ipv6", "[::1]", true},
	{"ipv6-port", "[::1]:3041", true},
	{"foreign", "attacker.example.com:3041", false},
	{"absent", "", false},
}

// assertHostRefused fails unless rec is a 403 written by the Host gate.
func assertHostRefused(t *testing.T, rec *httptest.ResponseRecorder, what, host string) {
	t.Helper()
	if rec.Code != http.StatusForbidden {
		t.Fatalf("%s with Host %q: status = %d, want 403", what, host, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), hostGateRefusal) {
		t.Errorf("%s with Host %q: the 403 did not come from the Host gate; body: %q", what, host, rec.Body.String())
	}
}

// TestHostGateControlMatrix — AC-WC-009 as amended, control matrix: every Host
// row crossed with a static asset, a dynamic GET, and a valid mutating POST.
func TestHostGateControlMatrix(t *testing.T) {
	const marker = "t613-host-gate-marker"
	const assetText = "MoAI Web Console"

	for _, hc := range hostGateCases {
		t.Run(hc.label+"/static", func(t *testing.T) {
			a := newTestApp(t)
			rec := serveWithHost(t, a.routes(), http.MethodGet, "/static/app.js", hc.host)
			if hc.allowed {
				if rec.Code != http.StatusOK {
					t.Fatalf("GET /static/app.js with Host %q: status = %d, want 200", hc.host, rec.Code)
				}
				if !strings.Contains(rec.Body.String(), assetText) {
					t.Errorf("GET /static/app.js with Host %q: body is not the asset", hc.host)
				}
				return
			}
			assertHostRefused(t, rec, "GET /static/app.js", hc.host)
			if strings.Contains(rec.Body.String(), assetText) {
				t.Errorf("GET /static/app.js with Host %q: the refusal carried the asset", hc.host)
			}
		})

		t.Run(hc.label+"/settings", func(t *testing.T) {
			a := newTestApp(t)
			a.readPreferences = func(string) (profile.ProfilePreferences, error) {
				return profile.ProfilePreferences{UserName: marker}, nil
			}
			rec := serveWithHost(t, a.routes(), http.MethodGet, "/settings", hc.host)
			if hc.allowed {
				if rec.Code != http.StatusOK {
					t.Fatalf("GET /settings with Host %q: status = %d, want 200", hc.host, rec.Code)
				}
				if !strings.Contains(rec.Body.String(), marker) {
					t.Errorf("GET /settings with Host %q: the page did not render the profile value", hc.host)
				}
				return
			}
			assertHostRefused(t, rec, "GET /settings", hc.host)
			if strings.Contains(rec.Body.String(), marker) {
				t.Errorf("GET /settings with Host %q: the refusal carried the profile value", hc.host)
			}
		})

		t.Run(hc.label+"/save", func(t *testing.T) {
			a := newTestApp(t)
			var wrote, synced bool
			var written profile.ProfilePreferences
			a.writePreferences = func(_ string, p profile.ProfilePreferences) error {
				wrote, written = true, p
				return nil
			}
			a.syncToProject = func(string, profile.ProfilePreferences) error {
				synced = true
				return nil
			}
			form := url.Values{
				"__profile":       {"default"},
				"user_name":       {marker},
				"permission_mode": {"acceptEdits"},
			}
			req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Sec-Fetch-Site", "same-origin") // satisfies the unchanged CSRF gate
			req.Host = hc.host
			rec := httptest.NewRecorder()
			a.routes().ServeHTTP(rec, req)

			if hc.allowed {
				if rec.Code < 200 || rec.Code >= 300 {
					t.Fatalf("POST /save with Host %q: status = %d, want 2xx; body:\n%s", hc.host, rec.Code, rec.Body.String())
				}
				if !wrote || written.UserName != marker {
					t.Errorf("POST /save with Host %q: the change was not persisted (wrote=%t, user_name=%q)", hc.host, wrote, written.UserName)
				}
				if !synced {
					t.Errorf("POST /save with Host %q: the project config was not synced", hc.host)
				}
				return
			}
			assertHostRefused(t, rec, "POST /save", hc.host)
			if wrote || synced {
				t.Errorf("POST /save with Host %q: state changed (wrote=%t, synced=%t)", hc.host, wrote, synced)
			}
		})
	}
}

// TestHostGateCoversEveryRouteAndMethod — AC-WC-009 as amended, second clause:
// every registered read route refuses a foreign and an absent Host, for GET and
// for methods the former gate never looked at (HEAD, OPTIONS, DELETE). The
// loopback control proves each refusal is the Host gate: the same method on the
// same route from a loopback Host is not refused.
func TestHostGateCoversEveryRouteAndMethod(t *testing.T) {
	routes := []string{"/", "/kanban", "/monitor", "/todo", "/specs", "/settings", "/events", "/static/app.js"}
	methods := []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete}
	h := newTestApp(t).routes()

	for _, path := range routes {
		for _, method := range methods {
			for _, host := range []string{"attacker.example.com:3041", ""} {
				rec := serveGateProbe(h, method, path, host)
				if rec.Code != http.StatusForbidden {
					t.Errorf("%s %s with Host %q: status = %d, want 403", method, path, host, rec.Code)
					continue
				}
				if method != http.MethodHead && !strings.Contains(rec.Body.String(), hostGateRefusal) {
					t.Errorf("%s %s with Host %q: the 403 did not come from the Host gate", method, path, host)
				}
			}
			if rec := serveGateProbe(h, method, path, "127.0.0.1:3041"); rec.Code == http.StatusForbidden {
				t.Errorf("%s %s from a loopback Host: status = 403, want the route's own answer", method, path)
			}
		}
		if rec := serveGateProbe(h, http.MethodGet, path, "127.0.0.1:3041"); rec.Code != http.StatusOK {
			t.Errorf("GET %s from a loopback Host: status = %d, want 200", path, rec.Code)
		}
	}
}

// TestHostGateDoesNotWidenFetchSiteCheck — AC-WC-009 as amended: the
// Sec-Fetch-Site same-origin gate is not widened. A loopback GET carrying a
// cross-site Sec-Fetch-Site value is still served, while a loopback POST with
// the same value is still refused by that gate.
func TestHostGateDoesNotWidenFetchSiteCheck(t *testing.T) {
	a := newTestApp(t)
	var wrote bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	get := httptest.NewRequest(http.MethodGet, "/settings", nil)
	get.Header.Set("Sec-Fetch-Site", "cross-site")
	get.Host = "127.0.0.1:3041"
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, get)
	if getRec.Code != http.StatusOK {
		t.Errorf("loopback GET with Sec-Fetch-Site cross-site: status = %d, want 200", getRec.Code)
	}

	form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}}
	post := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	post.Header.Set("Sec-Fetch-Site", "cross-site")
	post.Host = "127.0.0.1:3041"
	postRec := httptest.NewRecorder()
	h.ServeHTTP(postRec, post)
	if postRec.Code != http.StatusForbidden {
		t.Errorf("loopback POST with Sec-Fetch-Site cross-site: status = %d, want 403", postRec.Code)
	}
	if strings.Contains(postRec.Body.String(), hostGateRefusal) {
		t.Error("the cross-site POST was refused by the Host gate, not the Sec-Fetch-Site gate")
	}
	if wrote {
		t.Error("persistence occurred on a cross-site POST")
	}
}

// serveGateProbe sends one request with the given Host. /events is a stream that
// ends only when its request context does, so its request carries an
// already-cancelled context: the handler writes its ready event and returns
// instead of holding the test open.
func serveGateProbe(h http.Handler, method, path, host string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Host = host
	if path == "/events" {
		ctx, cancel := context.WithCancel(req.Context())
		cancel()
		req = req.WithContext(ctx)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
