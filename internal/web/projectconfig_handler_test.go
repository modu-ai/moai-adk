package web

import (
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// TestProjectOptionListsMatchCanonical guards the stable-order view-model option
// lists against drift from the canonical predicates (REQ-WC3-001/002 anti-pattern
// E.5.1: no parallel rule-set). The web lists fix the option ORDER for rendering,
// but their SET must equal the canonical source of truth.
func TestProjectOptionListsMatchCanonical(t *testing.T) {
	t.Parallel()

	// development_mode: web list must equal models.ValidDevelopmentModes().
	wantDev := make([]string, 0)
	for _, m := range models.ValidDevelopmentModes() {
		wantDev = append(wantDev, string(m))
	}
	// SPEC-WEB-CONSOLE-010: the web list is now schema-sourced via
	// developmentModeOptionList() (settings.DevelopmentModeOptionValues).
	gotDev := append([]string(nil), developmentModeOptionList()...)
	sort.Strings(wantDev)
	sort.Strings(gotDev)
	if strings.Join(gotDev, ",") != strings.Join(wantDev, ",") {
		t.Errorf("developmentModeOptionList set = %v, want %v (drift from models.ValidDevelopmentModes)", gotDev, wantDev)
	}

	// git_convention: web list must equal config.ValidConventions().
	// SPEC-WEB-CONSOLE-010: schema-sourced via conventionOptionList().
	wantConv := append([]string(nil), config.ValidConventions()...)
	gotConv := append([]string(nil), conventionOptionList()...)
	sort.Strings(wantConv)
	sort.Strings(gotConv)
	if strings.Join(gotConv, ",") != strings.Join(wantConv, ",") {
		t.Errorf("conventionOptionList set = %v, want %v (drift from config.ValidConventions)", gotConv, wantConv)
	}
}

// SPEC-DESIGN-MOAIWEBV2-001 M1: TestProjectFieldsetRendersSelects and
// TestProjectSelectsPreselectCurrentValues (AC-WC3-003/004 render tests) were removed
// with the orphan `project` render surface (fieldsetProject). development_mode /
// git_convention no longer render as web-console <select> widgets — they stay editable
// via yaml config / CLI. The server-side read/write/validate seams remain PRESERVED
// (REQ-MWV2-031) and are still covered by the handler round-trip tests below
// (TestSaveRejectsBogus*, TestSaveValidProjectConfigPersists, TestSaveEC2AtomicReject).
// The now-unused serveGetApp helper + the regexp import were removed with those tests.

// TestProjectReadSeamFailureRendersInlineError covers AC-WC3-004 (read failure):
// a read-seam error surfaces a readable inline error, never a blank page or panic.
func TestProjectReadSeamFailureRendersInlineError(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	a.readProjectConfig = func(string) (string, string, error) {
		return "", "", errors.New("disk read boom")
	}
	rec := serveGet(t, a.routes(), "/")
	if rec.Code < 400 {
		t.Errorf("read failure status = %d, want >= 400", rec.Code)
	}
	body := rec.Body.String()
	if strings.TrimSpace(body) == "" {
		t.Fatal("read failure produced a blank page")
	}
	if !strings.Contains(body, "disk read boom") {
		t.Errorf("read failure page should surface the error message, got: %s", body)
	}
}

// --- AC-WC3-001a/002a/005/EC: handler round-trip ---

// projectSaveForm builds a minimal valid POST form carrying the two project-config
// fields plus the baseline profile fields (so validatePrefs passes).
func projectSaveForm(devMode, convention string) url.Values {
	form := url.Values{}
	form.Set("__profile", "default")
	form.Set("development_mode", devMode)
	form.Set("git_convention", convention)
	return form
}

// TestSaveRejectsBogusDevelopmentMode covers AC-WC3-001a + EC: a non-canonical
// development_mode yields HTTP 400, a development_mode field error, and NO project
// config write (the write seam must not be invoked).
func TestSaveRejectsBogusDevelopmentMode(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	wrote := false
	a.writeProjectConfig = func(string, string, string) error { wrote = true; return nil }

	rec := servePost(t, a.routes(), "/save", projectSaveForm("xyz", "angular"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("bogus development_mode status = %d, want 400", rec.Code)
	}
	// SPEC-DESIGN-MOAIWEBV2-001 M1: field-error echo retired with the project render
	// surface; the server 400 + atomic no-write (below) are the preserved contract.
	if wrote {
		t.Error("write seam was invoked despite validation failure — must be atomic reject")
	}
}

// TestSaveRejectsBogusConvention covers AC-WC3-002a: a non-canonical git_convention
// yields 400 + field error + no write.
func TestSaveRejectsBogusConvention(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	wrote := false
	a.writeProjectConfig = func(string, string, string) error { wrote = true; return nil }

	rec := servePost(t, a.routes(), "/save", projectSaveForm("ddd", "gitflow"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("bogus git_convention status = %d, want 400", rec.Code)
	}
	// SPEC-DESIGN-MOAIWEBV2-001 M1: field-error echo retired with the project render
	// surface; the server 400 + atomic no-write (below) are the preserved contract.
	if wrote {
		t.Error("write seam invoked despite validation failure")
	}
}

// TestSaveValidProjectConfigPersists covers AC-WC3-001b/002b: valid values yield
// 200 and the write seam receives both values.
func TestSaveValidProjectConfigPersists(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	var gotDev, gotConv string
	called := false
	a.writeProjectConfig = func(_, devMode, convention string) error {
		called = true
		gotDev, gotConv = devMode, convention
		return nil
	}

	rec := servePost(t, a.routes(), "/save", projectSaveForm("ddd", "angular"))
	if rec.Code != http.StatusOK {
		t.Fatalf("valid save status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if !called {
		t.Fatal("write seam was not invoked on a valid save")
	}
	if gotDev != "ddd" || gotConv != "angular" {
		t.Errorf("write seam got (%q,%q), want (ddd,angular)", gotDev, gotConv)
	}
}

// TestSaveEmptyProjectConfigKeepsExisting covers EC-1: both fields empty → 200,
// write seam receives empty values (writeProjectConfig itself is the no-clobber
// guard), no error.
func TestSaveEmptyProjectConfigPasses(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	var gotDev, gotConv string
	a.writeProjectConfig = func(_, devMode, convention string) error {
		gotDev, gotConv = devMode, convention
		return nil
	}
	rec := servePost(t, a.routes(), "/save", projectSaveForm("", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("empty save status = %d, want 200", rec.Code)
	}
	if gotDev != "" || gotConv != "" {
		t.Errorf("write seam got (%q,%q), want empty (handler passes through, seam keeps existing)", gotDev, gotConv)
	}
}

// TestSaveEC2AtomicReject covers EC-2: one bogus + one valid → 400, FieldErrors
// has only development_mode, and NEITHER value is persisted (atomic reject).
func TestSaveEC2AtomicReject(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	wrote := false
	a.writeProjectConfig = func(string, string, string) error { wrote = true; return nil }

	rec := servePost(t, a.routes(), "/save", projectSaveForm("xyz", "angular"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("EC-2 status = %d, want 400", rec.Code)
	}
	// SPEC-DESIGN-MOAIWEBV2-001 M1: the development_mode field-error echo retired with
	// the project render surface; the server 400 + atomic no-write (below) are the
	// preserved contract (REQ-MWV2-031).
	if wrote {
		t.Error("EC-2 must be an atomic reject — no value persisted when any field is invalid")
	}
}

// TestSaveProjectWriteFailureSurfacesError covers REQ-WC3-005 error path: a write
// seam failure after a successful profile write surfaces a readable error.
func TestSaveProjectWriteFailureSurfacesError(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	a.writeProjectConfig = func(string, string, string) error {
		return errors.New("config save boom")
	}
	rec := servePost(t, a.routes(), "/save", projectSaveForm("ddd", "angular"))
	if rec.Code < 400 {
		t.Errorf("write failure status = %d, want >= 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "config save boom") {
		t.Errorf("write failure should surface the error, got: %s", rec.Body.String())
	}
}
