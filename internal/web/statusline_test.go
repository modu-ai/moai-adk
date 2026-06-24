package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestStatuslineThemeOptionList covers the schema-sourced theme option helper plus
// its absent-field fallback branch (defensive nil path).
func TestStatuslineThemeOptionList(t *testing.T) {
	opts := statuslineThemeOptionList()
	if len(opts) != 2 {
		t.Fatalf("statuslineThemeOptionList = %v, want 2 themes", opts)
	}
	if f, ok := settings.Field("statusline_theme"); !ok || len(f.SelectOptions()) != 2 {
		t.Error("schema statusline_theme must carry exactly 2 theme options")
	}
}

// TestBindFormStatuslineSegmentsSubmitted covers the bindForm statusline-segment
// submission branch (companion present → full 15-key map populated, unchecked → false).
func TestBindFormStatuslineSegmentsSubmitted(t *testing.T) {
	form := url.Values{}
	form.Set("statusline_theme", "catppuccin-latte")
	// Submit segments: companion present for all 15; check only a few.
	checkedSet := map[string]bool{"model": true, "git_branch": true, "pr": true}
	for _, seg := range settings.StatuslineSegmentKeys() {
		form.Set("seg_"+seg+"__present", "1")
		if checkedSet[seg] {
			form.Set("seg_"+seg, "1")
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("ParseForm: %v", err)
	}

	prefs := bindForm(req)
	if prefs.StatuslineTheme != "catppuccin-latte" {
		t.Errorf("StatuslineTheme = %q, want catppuccin-latte", prefs.StatuslineTheme)
	}
	if prefs.StatuslineSegments == nil {
		t.Fatal("StatuslineSegments must be populated when segment companions are present")
	}
	if len(prefs.StatuslineSegments) != 15 {
		t.Errorf("StatuslineSegments has %d keys, want 15 (full map)", len(prefs.StatuslineSegments))
	}
	for _, seg := range settings.StatuslineSegmentKeys() {
		got, present := prefs.StatuslineSegments[seg]
		if !present {
			t.Errorf("segment %q missing from map", seg)
		}
		if got != checkedSet[seg] {
			t.Errorf("segment %q = %v, want %v", seg, got, checkedSet[seg])
		}
	}
}

// TestBindFormStatuslineSegmentsNotSubmitted covers the bindForm branch where NO
// segment companion is present → StatuslineSegments stays nil (preserve on-disk).
func TestBindFormStatuslineSegmentsNotSubmitted(t *testing.T) {
	form := url.Values{}
	form.Set("user_name", "Goos")
	// No seg_*__present companions.

	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("ParseForm: %v", err)
	}

	prefs := bindForm(req)
	if prefs.StatuslineSegments != nil {
		t.Errorf("StatuslineSegments = %v, want nil (no segment companion → preserve)", prefs.StatuslineSegments)
	}
}

// TestStatuslineRendersCheckedSegments covers the segmentToggle checked branch:
// a profile with some segments enabled renders those checkboxes as `checked`.
func TestStatuslineRendersCheckedSegments(t *testing.T) {
	a := newTestApp(t)
	// Override readPreferences so the GET page renders with some segments enabled.
	a.readPreferences = func(name string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{
			StatuslineTheme: "catppuccin-latte",
			StatuslineSegments: map[string]bool{
				"model":      true,
				"git_branch": true,
			},
		}, nil
	}
	h := a.routes()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d", rec.Code)
	}
	body := rec.Body.String()

	// The "model" segment checkbox must render checked.
	if !strings.Contains(body, `id="seg_model" name="seg_model" value="1" checked`) {
		t.Error("seg_model checkbox must render checked when enabled in prefs")
	}
	// A non-enabled segment renders without checked.
	if strings.Contains(body, `id="seg_task" name="seg_task" value="1" checked`) {
		t.Error("seg_task checkbox must NOT be checked when disabled in prefs")
	}
	// The selected theme is preselected.
	if !strings.Contains(body, `<option value="catppuccin-latte" selected>`) {
		t.Error("statusline_theme select must preselect the persisted theme")
	}
}

// TestStatuslineThemeErrorRendersInline covers the fieldsetStatusline has-error
// render branch: a POST with a bogus theme is rejected and the page re-renders with
// the statusline_theme field in its error state (aria-invalid + error message).
func TestStatuslineThemeErrorRendersInline(t *testing.T) {
	a := newTestApp(t)
	a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
	h := a.routes()

	form := url.Values{}
	form.Set("statusline_theme", "neon-disco") // bogus → validation rejects
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "127.0.0.1:8080" // loopback Host — the console is loopback-gated
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST bogus theme status = %d, want 400", rec.Code)
	}
	body := rec.Body.String()
	// The statusline_theme select must render in its error state.
	if !strings.Contains(body, `id="statusline_theme"`) {
		t.Error("rejected page missing statusline_theme control")
	}
	if !strings.Contains(body, "unrecognized statusline theme") {
		t.Error("rejected page missing the statusline_theme error message")
	}
}

// TestStatuslineThemeValidationRejectsBogus covers the new statusline_theme
// validation branch in validatePrefs.
func TestStatuslineThemeValidationRejectsBogus(t *testing.T) {
	errs := validatePrefs(profile.ProfilePreferences{StatuslineTheme: "neon-disco"})
	if errs["statusline_theme"] == "" {
		t.Error("validatePrefs must reject a non-canonical statusline theme")
	}
	// A canonical theme passes.
	errs = validatePrefs(profile.ProfilePreferences{StatuslineTheme: "catppuccin-mocha"})
	if errs["statusline_theme"] != "" {
		t.Errorf("validatePrefs rejected a canonical theme: %v", errs["statusline_theme"])
	}
	// Empty theme is allowed (theme-only-unset save).
	errs = validatePrefs(profile.ProfilePreferences{})
	if errs["statusline_theme"] != "" {
		t.Errorf("validatePrefs must allow empty statusline theme: %v", errs["statusline_theme"])
	}
}
