package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-AUTONOMY-TIERS-001 M6 — template neutrality + opt-in guard (AC-006).
//
// AC-006: the distributed template MUST NOT ship fully-autonomous as a default,
// recommended, or pre-selected tier. The template's settings.json.tmpl carries
// no defaultMode: bypassPermissions and no fully-autonomous pre-selection. A
// user opts in via the --autonomy-tier flag, the web toggle (with proof), or a
// manual settings.local.json entry.
//
// This guard is the run-phase regression for AC-006; the CI guard
// (template-neutrality-check.yaml) is the safety net for SPEC IDs / REQ tokens.

// templateSettingsTmpl returns the template settings.json.tmpl body, skipping
// the test cleanly if the template tree is absent (consumer-project checkout).
func templateSettingsTmpl(t *testing.T) []byte {
	t.Helper()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repoRoot, "internal", "template", "templates", ".claude", "settings.json.tmpl")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("template settings.json.tmpl not found at %s (consumer checkout?)", path)
	}
	return body
}

func TestTemplate_NoBypassPermissionsDefault(t *testing.T) {
	// AC-006: the distributed template MUST NOT ship bypassPermissions.
	body := templateSettingsTmpl(t)
	// A defaultMode bypassPermissions would be a forbidden pre-pick of the
	// dangerous tier. The local dev .claude/settings.json carries it
	// intentionally (CLAUDE.local.md §22) but the TEMPLATE must not.
	if strings.Contains(string(body), `"defaultMode": "bypassPermissions"`) {
		t.Errorf("AC-006 violation: template settings.json.tmpl ships defaultMode: bypassPermissions (forbidden — fully-autonomous is opt-in only)")
	}
}

func TestTemplate_NoFullyAutonomousPreSelection(t *testing.T) {
	// AC-006: no fully-autonomous pre-selection or recommendation in the
	// template. The selector OFFERS the tier (via the --autonomy-tier flag
	// help at runtime); the template does not PRE-PICK it.
	body := templateSettingsTmpl(t)
	forbidden := []string{
		// A default autonomy-tier key in the template config would pre-pick it.
		"autonomy_tier: fully-autonomous",
		"autonomy-tier: fully-autonomous",
		`"MOAI_AUTONOMY_TIER": "fully-autonomous"`,
	}
	for _, f := range forbidden {
		if strings.Contains(string(body), f) {
			t.Errorf("AC-006 violation: template contains %q (fully-autonomous pre-selection is forbidden)", f)
		}
	}
}
