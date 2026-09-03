package web

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// gate_panel_test.go — the Gate console panel (SPEC-PRECOMMIT-GATE-SCOPE-001
// M2 / REQ-009 / AC-010): tab + panel render wiring, the bool companion
// naming contract, the 4-locale i18n keys, and the seam save path landing in
// exactly gate.yaml.

// TestGatePanelRendered checks the tab nav, the panel, and the form controls
// render on the console page.
func TestGatePanelRendered(t *testing.T) {
	body := renderConsolePage(t)
	for _, probe := range []string{
		`data-tab="gate"`,
		`data-panel="gate"`,
		`name="gate.pre_commit.enabled"`,
		`name="gate.pre_commit.enabled__present"`,
	} {
		if !strings.Contains(body, probe) {
			t.Errorf("rendered console missing %q", probe)
		}
	}
}

// TestGateI18nKeysInAllLocales checks the panel and field labels resolve in
// all four locales.
func TestGateI18nKeysInAllLocales(t *testing.T) {
	keys := []string{
		"sec.gate.title", "sec.gate.desc",
		"f.gate.pre_commit.enabled.title", "f.gate.pre_commit.enabled.desc",
	}
	for _, k := range keys {
		if !i18nKeyInAllLocales(t, k) {
			t.Errorf("i18n.js missing gate key %q in all 4 locales", k)
		}
	}
}

// TestGateFieldsWired guards the settings-side registration from the web side.
func TestGateFieldsWired(t *testing.T) {
	if n := len(settings.SectionFields(settings.SectionGate)); n != 1 {
		t.Errorf("SectionFields(SectionGate) = %d fields, want 1", n)
	}
}

// TestGateSavePath is AC-010(2): toggling the field and saving persists
// gate.pre_commit.enabled to exactly .moai/config/sections/gate.yaml — no
// other section file is touched, comments and non-modeled keys survive.
func TestGateSavePath(t *testing.T) {
	a, root := newSchemaTestApp(t)

	// Seed a gate.yaml carrying a comment and a non-modeled key (the fixture
	// family seeds the other sections; gate is written here so the assertion
	// scope stays local to this file).
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	gateBefore := "# gate header comment\ngate:\n  pre_commit:\n    enabled: false\n  user_only_key: keep-me\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "gate.yaml"), []byte(gateBefore), 0o644); err != nil {
		t.Fatal(err)
	}
	workflowBefore := readSectionFile(t, root, "workflow")

	// Unchecked (companion present, value absent) → explicit false write.
	rec := postSave(t, a, url.Values{
		"gate.pre_commit.enabled":          {""},
		"gate.pre_commit.enabled__present": {"1"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save status = %d, want 200 (body: %.300s)", rec.Code, rec.Body.String())
	}
	after := readSectionFile(t, root, "gate")
	if !strings.Contains(after, "enabled: false") {
		t.Errorf("explicit false toggle not persisted:\n%s", after)
	}
	if !strings.Contains(after, "user_only_key: keep-me") || !strings.Contains(after, "# gate header comment") {
		t.Errorf("comment or non-modeled key lost:\n%s", after)
	}
	if readSectionFile(t, root, "workflow") != workflowBefore {
		t.Error("forged gate POST mutated an unrelated section file")
	}

	// Checked (companion present, value present) → true write.
	rec = postSave(t, a, url.Values{
		"gate.pre_commit.enabled":          {"on"},
		"gate.pre_commit.enabled__present": {"1"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save (on) status = %d, want 200 (body: %.300s)", rec.Code, rec.Body.String())
	}
	after = readSectionFile(t, root, "gate")
	values, err := settings.SchemaCurrentValues(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := values["gate.pre_commit.enabled"]; got != "true" {
		t.Errorf("persisted gate.pre_commit.enabled = %q, want true (file:\n%s)", got, after)
	}
}
