package web

import (
	"regexp"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestM5bD1_AtomicSave_AllPanelsInDOM verifies the atomic Save contract:
// all tabpanels are present in the DOM (CSS show/hide only — never removed),
// so ALL form fields submit regardless of which tab is active.
func TestM5bD1_AtomicSave_AllPanelsInDOM(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{UserName: "test"})

	// 1. All 15 tabpanels present in DOM.
	panelCount := strings.Count(body, `class="tabpanel`)
	if panelCount < 6 {
		t.Errorf("expected >= 6 tabpanel divs in DOM (6-tab console, atomic Save contract — SPEC-WEBCONF-SIMPLIFY-001 M3), found %d", panelCount)
	}

	// 2. Tab nav present with data-tab attributes (6 surviving tabs).
	tabCount := strings.Count(body, `data-tab=`)
	if tabCount < 6 {
		t.Errorf("expected >= 6 data-tab buttons (6-tab console — M3), found %d", tabCount)
	}

	// 3. First tab + first panel are active.
	if !strings.Contains(body, `class="tab is-active"`) {
		t.Error("no active tab button found")
	}
	if !strings.Contains(body, `class="tabpanel is-active"`) {
		t.Error("no active tabpanel found")
	}

	// 4. CSS rule for display:none on inactive panels.
	css := readEmbeddedAsset(t, "console.css")
	if !strings.Contains(css, ".tabpanel:not(.is-active)") {
		t.Error("console.css missing .tabpanel:not(.is-active) display:none rule")
	}

	// 5. wireTabs function present in app.js.
	js := readEmbeddedAsset(t, "app.js")
	if !strings.Contains(js, "function wireTabs") {
		t.Error("app.js missing wireTabs function")
	}

	// 6. All canonical form fields present (atomic Save — they submit from
	// non-active panels too). SPEC-DESIGN-MOAIWEBV2-001 M1: development_mode /
	// git_convention were removed with the orphan `project` panel (editable via
	// yaml config / CLI); the remaining panels' fields still submit atomically.
	for _, name := range []string{
		"user_name", "conversation_lang", "permission_mode", "model",
		"effort_level",
	} {
		if !strings.Contains(body, `name="`+name+`"`) {
			t.Errorf("form field name=%q missing from DOM (atomic Save contract broken)", name)
		}
	}
}

// TestM5bD3D4_NoCodeChipDataI18n verifies field__key code chips do NOT carry
// data-i18n (TestDataI18nKeysNoCodeChip equivalent — code tokens must not be
// translated). Uses the same regex as TestDataI18nWiring.
func TestM5bD3D4_NoCodeChipDataI18n(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{UserName: "test"})
	chipRe := regexp.MustCompile(`<code class="field__key"[^>]*data-i18n`)
	if chipRe.MatchString(body) {
		t.Error("a field__key code chip carries data-i18n (code tokens must not be translated)")
	}
}
