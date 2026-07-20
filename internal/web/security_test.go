package web

// SPEC-INTERNAL-SECURITY-001 M1 — web security regression tests (REQ-SEC-001,
// REQ-SEC-002). These are RED before the GREEN implementation:
//   - AC-SEC-001a: GET read path must reject a traversal ?profile= payload 4xx
//     (currently no guard on handleIndex — unauthenticated arbitrary-path read).
//   - AC-SEC-002a: cross-site POST (Sec-Fetch-Site: cross-site) must be 4xx.
//   - AC-SEC-002 (conservative default): absent Sec-Fetch-Site on a
//     state-changing POST must be 4xx.
//   - AC-SEC-002b (NFR-SEC-003 behavior preservation): same-origin POST must
//     still succeed (no false-positive deny).

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestGETReadPathRejectsTraversalProfile verifies AC-SEC-001a: the GET read
// path (handleIndex) rejects a traversal ?profile= payload with 4xx before any
// file read occurs. Without the guard, GET /?profile=../../etc is an
// unauthenticated arbitrary-path read primitive.
func TestGETReadPathRejectsTraversalProfile(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	h := a.routes()

	badNames := []string{
		"../../etc/passwd",
		"../x",
		"a/b",
		"a\\b",
		"..",
		".hidden",
	}
	for _, bad := range badNames {
		rec := serveGet(t, h, "/?profile="+url.QueryEscape(bad))
		if rec.Code < 400 || rec.Code >= 500 {
			t.Errorf("GET /?profile=%q status = %d, want 4xx (traversal must be rejected on read path)", bad, rec.Code)
		}
	}
}

// TestGETReadPathAllowsValidProfileName verifies NFR-SEC-003: a normal profile
// name on the GET read path still renders 200 (no false-positive deny from the
// new guard).
func TestGETReadPathAllowsValidProfileName(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	h := a.routes()

	for _, valid := range []string{"", "default", "work", "my-profile"} {
		rec := serveGet(t, h, "/?profile="+url.QueryEscape(valid))
		if rec.Code != http.StatusOK {
			t.Errorf("GET /?profile=%q status = %d, want 200 (valid name must not be denied)", valid, rec.Code)
		}
	}
}

// TestCrossSitePostRejected verifies AC-SEC-002a: a POST with
// Sec-Fetch-Site: cross-site to a state-changing route is rejected 4xx and no
// persistence occurs. A drive-by CSRF auto-submit from a malicious page sends
// this header value (it cannot be stripped by cross-origin JS per the Fetch
// Metadata spec).
func TestCrossSitePostRejected(t *testing.T) {
	a := newTestApp(t)
	var wrote bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}}
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-site POST status = %d, want 403", rec.Code)
	}
	if wrote {
		t.Error("persistence occurred on a cross-site POST (state must be unchanged)")
	}
}

// TestAbsentFetchSiteHeaderRejected verifies AC-SEC-002 conservative default:
// a state-changing POST with NO Sec-Fetch-Site header is rejected 4xx. A modern
// browser always sends the header for fetch-initiated requests, so an absent
// header indicates a legacy client or non-browser agent — rejected by default.
func TestAbsentFetchSiteHeaderRejected(t *testing.T) {
	a := newTestApp(t)
	var wrote bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}}
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Deliberately NO Sec-Fetch-Site header.
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("absent-Sec-Fetch-Site POST status = %d, want 403", rec.Code)
	}
	if wrote {
		t.Error("persistence occurred on an absent-Sec-Fetch-Site POST (state must be unchanged)")
	}
}

// TestSameOriginPostAllowed verifies AC-SEC-002b / NFR-SEC-003: a same-origin
// POST (Sec-Fetch-Site: same-origin) to a state-changing route is processed
// normally — no false-positive deny from the new enforcement.
func TestSameOriginPostAllowed(t *testing.T) {
	a := newTestApp(t)
	var wrote bool
	a.writePreferences = func(string, profile.ProfilePreferences) error { wrote = true; return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{"__profile": {"default"}, "permission_mode": {"acceptEdits"}}
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("same-origin POST rejected with 403, want allowed; body:\n%s", rec.Body.String())
	}
	if !wrote {
		t.Error("same-origin POST did not persist (WritePreferences not called)")
	}
}

// TestCrossSiteShutdownRejected verifies AC-SEC-002a on the /__shutdown__ route:
// a cross-site POST must not trigger a shutdown (DoS prevention).
func TestCrossSiteShutdownRejected(t *testing.T) {
	a := newTestApp(t)
	var triggered bool
	a.triggerShutdown = func() { triggered = true }
	h := a.routes()

	req := httptest.NewRequest(http.MethodPost, "/__shutdown__", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("cross-site POST /__shutdown__ status = %d, want 403", rec.Code)
	}
	if triggered {
		t.Error("triggerShutdown fired on a cross-site POST (must be rejected before the handler)")
	}
}
