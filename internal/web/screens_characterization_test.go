package web

// screens_characterization_test.go — characterization tests for the existing
// screen handlers in screens.go (Issue #1948, B12). They assert what the code
// does today and change no production code. A behavior that looks like a defect
// is not pinned here; it is reported (see the notes at the end of the file).
//
// Defect answer: these tests catch a screen that answers a non-GET, accepts a
// profile name with a path separator, drops the project's current profile from
// the shell, names the wrong rename targets, answers an unknown path with a
// page, turns a failing SPEC scan into a half-drawn 200, or hides a render
// failure behind a 200.

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/modu-ai/moai-adk/internal/profile"
)

func failingSpecApp(t *testing.T) *app {
	t.Helper()
	a := newApp(Config{ProjectRoot: t.TempDir()})
	a.specs = newSpecCache(func(string) ([]SpecRowVM, map[string][]FindingVM, error) {
		return nil, nil, errors.New("scan failed")
	}, specCacheMaxAge)
	return a
}

func TestShellVMProfilesAndRenameTargets(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	a.listProfiles = func() []profile.ProfileEntry {
		return []profile.ProfileEntry{{Name: "default"}, {Name: "work", Current: true}, {Name: "other"}, {Name: "spare"}}
	}
	req := httptest.NewRequest(http.MethodGet, "/monitor", nil)
	vm := a.shellVM(req, "monitor", "Monitor", "crumb")
	if vm.Profile != "work" || len(vm.Profiles) != 4 {
		t.Fatalf("profile %q, %d profiles", vm.Profile, len(vm.Profiles))
	}
	// Rename targets: not default, not the current profile.
	if got := strings.Join(vm.RenameTargets, ","); got != "other,spare" {
		t.Errorf("rename targets = %q, want other,spare", got)
	}
	// A valid ?profile= overrides the shown profile; the targets were already chosen.
	req = httptest.NewRequest(http.MethodGet, "/monitor?profile=other", nil)
	if vm := a.shellVM(req, "monitor", "Monitor", "crumb"); vm.Profile != "other" {
		t.Errorf("?profile=other shown as %q", vm.Profile)
	}
	// An invalid one is ignored by the shell (readOnly is what refuses it).
	req = httptest.NewRequest(http.MethodGet, "/monitor?profile=..%2Fx", nil)
	if vm := a.shellVM(req, "monitor", "Monitor", "crumb"); vm.Profile != "work" {
		t.Errorf("an invalid ?profile= must not change the shown profile, got %q", vm.Profile)
	}
	// No profile list and no query → "default".
	b := newApp(Config{ProjectRoot: t.TempDir()})
	b.listProfiles = func() []profile.ProfileEntry { return nil }
	if vm := b.shellVM(httptest.NewRequest(http.MethodGet, "/", nil), "overview", "Overview", ""); vm.Profile != "default" {
		t.Errorf("fallback profile = %q", vm.Profile)
	}
}

func TestReadOnlyGateRefusesWritesAndBadProfileNames(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	for _, path := range []string{"/", "/factory", "/specs", "/monitor", "/todo"} {
		rec := servePost(t, a.routes(), path, url.Values{})
		if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != http.MethodGet {
			t.Errorf("POST %s = %d Allow %q, want 405 GET", path, rec.Code, rec.Header().Get("Allow"))
		}
	}
	rec := serveGet(t, a.routes(), "/factory?profile=..%2Fetc")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid profile name") {
		t.Errorf("a profile name with a path separator = %d %q", rec.Code, rec.Body.String())
	}
}

func TestOverviewAnswersOnlyItsOwnPath(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	if rec := serveGet(t, a.routes(), "/overview"); rec.Code != http.StatusOK {
		t.Errorf("GET /overview = %d", rec.Code)
	}
	if rec := serveGet(t, a.routes(), "/no-such-screen"); rec.Code != http.StatusNotFound {
		t.Errorf("an unregistered path = %d, want 404", rec.Code)
	}
}

func TestScreensTurnAFailedSpecScanIntoA500(t *testing.T) {
	for path, prefix := range map[string]string{
		"/overview": "overview unavailable",
		"/factory":  "factory unavailable",
		"/specs":    "spec list unavailable",
	} {
		rec := serveGet(t, failingSpecApp(t).routes(), path)
		if rec.Code != http.StatusInternalServerError || !strings.HasPrefix(rec.Body.String(), prefix) {
			t.Errorf("GET %s with a failing scan = %d %q, want 500 starting %q", path, rec.Code, rec.Body.String(), prefix)
		}
	}
}

func TestSpecsPassesItsFiltersThrough(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	a.specs = newSpecCache(func(string) ([]SpecRowVM, map[string][]FindingVM, error) {
		return []SpecRowVM{{ID: "SPEC-A-001", Status: "draft"}, {ID: "SPEC-B-001", Status: "completed"}}, nil, nil
	}, specCacheMaxAge)
	rec := serveGet(t, a.routes(), "/specs?status=draft&q=SPEC-A")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "SPEC-A-001") {
		t.Fatalf("GET /specs filtered = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "SPEC-B-001") {
		t.Errorf("a completed SPEC leaked through the draft filter")
	}
}

func TestRenderPageReportsARenderFailureAs500(t *testing.T) {
	a := newApp(Config{ProjectRoot: t.TempDir()})
	rec := httptest.NewRecorder()
	a.renderPage(rec, templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, _ = io.WriteString(w, "half a page")
		return errors.New("boom")
	}))
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "render failed") {
		t.Errorf("a failing render = %d %q, want 500 with the failure named", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "half a page") {
		t.Errorf("a half-drawn page must not be sent")
	}
}

// Findings recorded, not pinned (reported to the leader instead):
//   - handleOverview, handleFactory and handleSpecs write err.Error() of a failed
//     SPEC scan into the 500 body unrefined, while handleMonitor passes it through
//     refineText. Pinning the raw echo would harden a possible leak of a path or
//     address in the error text, so these tests check the status and the prefix only.
