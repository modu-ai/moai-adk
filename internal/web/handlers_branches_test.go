package web

// handlers_branches_test.go — refusal and error paths of the settings handlers
// (handlers.go) that no test reached (Issue #1948, coverage of the hand-written
// code): the bind-address fallback, the settings route's path guard and profile
// name guard, a malformed save body, and a failing section-config read. Tests only.
//
// Defect answer: these tests catch an unknown path answered with the settings
// page, a profile name with a path separator read, a malformed save accepted, and
// a section-config read failure turned into a blank form.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResolveBindAddrFallbacks(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir(), Port: 9123})
	if got := a.resolveBindAddr(); got != loopbackHost+":9123" {
		t.Errorf("a configured port without a listener = %q, want %s:9123", got, loopbackHost)
	}
	b := newApp(Config{ProjectRoot: t.TempDir()})
	if got := b.resolveBindAddr(); got != loopbackHost {
		t.Errorf("no port and no listener = %q, want the bare loopback host", got)
	}
	c := newApp(Config{ProjectRoot: t.TempDir(), Port: 9123})
	c.bindAddr = func() string { return "127.0.0.1:55555" }
	if got := c.resolveBindAddr(); got != "127.0.0.1:55555" {
		t.Errorf("a bound listener wins: %q", got)
	}
}

func TestHandleIndexGuardsPathAndProfileName(t *testing.T) {
	a := newTestApp(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings/extra", nil)
	req.Host = "127.0.0.1:8080"
	a.handleIndex(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("a path other than / and /settings = %d, want 404", rec.Code)
	}
	bad := serveGet(t, a.routes(), "/settings?profile=..%2F..%2Fetc")
	if bad.Code != http.StatusBadRequest || !strings.Contains(bad.Body.String(), "invalid profile name") {
		t.Errorf("a profile name with a path separator = %d, want 400 invalid profile name", bad.Code)
	}
}

func TestHandleSaveRefusesAMalformedBody(t *testing.T) {
	a := newTestApp(t)
	var calls []saveStep
	recordingSeams(a, &calls, "")
	rec := postRawForm(t, a.routes(), "/save", "user_name=%zz")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "could not parse form") {
		t.Errorf("POST /save with a malformed body = %d, want 400 could not parse form", rec.Code)
	}
	if len(calls) != 0 {
		t.Errorf("a malformed body reached a persistence seam: %v", calls)
	}
}

func TestSettingsPageReportsAFailingSectionConfigRead(t *testing.T) {
	a := newTestApp(t)
	a.schemaCurrentValues = func(string) (map[string]string, error) {
		return nil, errors.New("section read failed for alice.private@example.com")
	}
	rec := serveGet(t, a.routes(), "/settings")
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "could not read section config") {
		t.Fatalf("GET /settings with a failing section read = %d, want 500 could not read section config", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "alice.private@example.com") {
		t.Errorf("the failure text reached the page unrefined")
	}
}
