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

// M3 of the surgical web-console redesign removed the Statusline section's web
// rendering: the fieldset, theme picker, segment toggles, bindForm segment block,
// and statusline_theme validation are all gone. The statusline schema fields
// (settings.SectionStatusline) + accessors (StatuslineSegmentKeys) + schema_bridge
// + statusline.yaml are PRESERVED — they remain consumed by the TUI /
// profile_setup CLI path. The settings-package tests cover the schema half; this
// file guards the web half: statusline is absent from render, bind, and validate.

// TestStatuslineSectionAbsentFromConsole asserts the rendered GET / page carries
// NO statusline fieldset: no sec.statusline chrome, no statusline_theme select,
// no seg_<key> checkbox toggles, and no seg_<key>__present companions.
func TestStatuslineSectionAbsentFromConsole(t *testing.T) {
	a := newTestApp(t)
	a.readPreferences = func(name string) (profile.ProfilePreferences, error) {
		// Prefs still carry statusline fields (consumed by the TUI path); the web
		// console must NOT render them even when populated.
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

	// No statusline section chrome / title / theme select.
	for _, want := range []string{
		`data-i18n="sec.statusline.title"`,
		`data-i18n="sec.statusline.desc"`,
		`data-i18n="f.statusline_theme.title"`,
		`id="statusline_theme"`,
		`name="statusline_theme"`,
	} {
		if strings.Contains(body, want) {
			t.Errorf("rendered page must NOT contain statusline marker %q (section removed)", want)
		}
	}
	// No segment toggle controls or __present companions for any canonical key.
	for _, seg := range settings.StatuslineSegmentKeys() {
		if strings.Contains(body, `name="seg_`+seg+`"`) {
			t.Errorf("rendered page must NOT contain segment toggle seg_%s (section removed)", seg)
		}
		if strings.Contains(body, `name="seg_`+seg+`__present"`) {
			t.Errorf("rendered page must NOT contain __present companion for seg_%s (section removed)", seg)
		}
	}
}

// TestBindFormIgnoresStatusline asserts bindForm no longer binds statusline_theme
// or any segment toggle: even when a client posts them, the bound prefs carry no
// statusline values (the web handler leaves them zero so syncStatusline preserves
// on-disk statusline config).
func TestBindFormIgnoresStatusline(t *testing.T) {
	form := url.Values{}
	form.Set("statusline_theme", "catppuccin-latte")
	for _, seg := range settings.StatuslineSegmentKeys() {
		form.Set("seg_"+seg+"__present", "1")
		form.Set("seg_"+seg, "1")
	}

	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("ParseForm: %v", err)
	}

	prefs := bindForm(req)
	if prefs.StatuslineTheme != "" {
		t.Errorf("StatuslineTheme = %q, want empty (web handler no longer binds it)", prefs.StatuslineTheme)
	}
	if prefs.StatuslineSegments != nil {
		t.Errorf("StatuslineSegments = %v, want nil (web handler no longer binds segments)", prefs.StatuslineSegments)
	}
}

// TestValidatePrefsNoStatuslineTheme asserts validatePrefs has no statusline_theme
// branch: a bogus theme is NOT rejected (the validation was removed alongside the
// section). The settings schema remains the authority for canonical themes — the
// web layer simply does not surface the field.
func TestValidatePrefsNoStatuslineTheme(t *testing.T) {
	// A bogus theme must NOT produce a statusline_theme error (branch removed).
	errs := validatePrefs(profile.ProfilePreferences{StatuslineTheme: "neon-disco"})
	if _, ok := errs["statusline_theme"]; ok {
		t.Errorf("validatePrefs must NOT reject statusline_theme (web validation removed): %v", errs["statusline_theme"])
	}
}
