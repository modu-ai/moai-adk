package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestPrimaryRailKeepsThreeSurfaceIA pins the agreed first-level information
// architecture. Kanban, Specs, and Monitor remain route-compatible screens;
// they are not shown as primary navigation or Overview destinations.
func TestPrimaryRailKeepsThreeSurfaceIA(t *testing.T) {
	body := renderConsolePage(t)

	for _, want := range []string{
		`class="nav__row" href="/"`,
		`class="nav__row" href="/todo"`,
		`class="nav__row" href="/settings"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("primary rail is missing %s", want)
		}
	}
	for _, retired := range []string{
		`class="nav__row" href="/kanban"`,
		`class="nav__row" href="/specs"`,
		`class="nav__row" href="/monitor"`,
	} {
		if strings.Contains(body, retired) {
			t.Errorf("retired operational surface remains in primary rail: %s", retired)
		}
	}
	for _, retired := range []string{`href="/kanban"`, `href="/specs"`, `href="/monitor"`} {
		if strings.Contains(body, retired) {
			t.Errorf("settings page exposes an unrequested Overview destination: %s", retired)
		}
	}
}

// TestSettingsPageHeaderKeepsAllFourteenTabs checks the runtime-rendered
// detail header for every current settings tab. The header must identify the
// active page, its effect timing, and the complete 14-page sibling count.
func TestSettingsPageHeaderKeepsAllFourteenTabs(t *testing.T) {
	a := newTestApp(t)
	for _, tab := range consoleTabs() {
		req := httptest.NewRequest(http.MethodGet, "/settings?tab="+tab.ID, nil)
		rec := httptest.NewRecorder()
		a.routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /settings?tab=%s status = %d, want 200", tab.ID, rec.Code)
		}
		body := rec.Body.String()
		effectKey, _ := settingsTabEffect(tab.ID)
		for _, want := range []string{
			`class="settings-page-head"`,
			`data-i18n="settings.detail.pages.suffix"`,
			`data-i18n="` + settingsTabDescKey(tab.ID) + `"`,
			`data-i18n="` + effectKey + `"`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("tab %q header is missing %s", tab.ID, want)
			}
		}
		if !strings.Contains(body, `>14<`) {
			t.Errorf("tab %q header does not disclose the 14-page settings contract", tab.ID)
		}
	}
}
