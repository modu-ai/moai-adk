package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// fourLocales는 i18n parity 검증 대상 4개 로케일이다.
var fourLocales = []string{"en", "ko", "ja", "zh"}

// TestI18nKeySetParity covers AC-WC10-016: every schema i18n key resolves through
// the TUI bridge resolver into a NON-EMPTY value for all 4 locales. (The web side
// of this AC — matching f./seg. dotted keys in i18n.js — is asserted in
// internal/web; this is the TUI half: resolution THROUGH the named bridge, not by
// expecting an f.* literal in the struct file, per design §F.2.)
func TestI18nKeySetParity(t *testing.T) {
	for _, f := range settings.AllFields() {
		for _, locale := range fourLocales {
			label, ok := schemaKeyToTUIField(f.I18nKey, locale)
			if !ok {
				t.Errorf("schema field %q (key %q): no TUI bridge entry", f.Name, f.I18nKey)
				continue
			}
			if strings.TrimSpace(label.Title) == "" {
				t.Errorf("schema field %q (key %q) locale %q: bridge resolved to EMPTY title", f.Name, f.I18nKey, locale)
			}
			// Non-segment fields (select/text/int/float/bool with f. prefix) must also
			// have a non-empty description. Statusline segments (seg. prefix) have title only.
			if !strings.HasPrefix(f.I18nKey, "seg.") && strings.TrimSpace(label.Desc) == "" {
				t.Errorf("schema field %q (key %q) locale %q: bridge resolved to EMPTY desc", f.Name, f.I18nKey, locale)
			}
		}
	}
}

// TestI18nSegmentParity covers the segment half of AC-WC10-016: every one of the 15
// canonical statusline segment keys resolves through the bridge for all 4 locales.
func TestI18nSegmentParity(t *testing.T) {
	for _, seg := range settings.StatuslineSegmentKeys() {
		key := "seg." + seg
		for _, locale := range fourLocales {
			label, ok := schemaKeyToTUIField(key, locale)
			if !ok {
				t.Errorf("segment key %q: no TUI bridge entry", key)
				continue
			}
			if strings.TrimSpace(label.Title) == "" {
				t.Errorf("segment key %q locale %q: bridge resolved to EMPTY title", key, locale)
			}
		}
	}
}

// TestBridgeFieldDefResolver covers the fieldDefTUILabel convenience wrapper: every
// schema FieldDef resolves through it.
func TestBridgeFieldDefResolver(t *testing.T) {
	for _, f := range settings.AllFields() {
		if _, ok := fieldDefTUILabel(f, "en"); !ok {
			t.Errorf("fieldDefTUILabel could not resolve field %q (key %q)", f.Name, f.I18nKey)
		}
	}
	if _, ok := schemaKeyToTUIField("__nonexistent__", "en"); ok {
		t.Error("schemaKeyToTUIField reported a bogus key as resolvable")
	}
}

// TestTUIRendersSchemaFieldSet covers AC-WC10-010 (TUI side): the TUI renders a
// widget for every schema field. Since the huh form is constructed inline (not
// programmatically enumerable), this asserts (a) the bridge has an entry for every
// schema field name (label path exists) AND (b) the wizard source binds value
// variables for the 7 new nested fields. Together these prove the TUI covers the
// full 34-field schema set.
func TestTUIRendersSchemaFieldSet(t *testing.T) {
	// (a) Every schema field name resolves through the bridge.
	for _, f := range settings.AllFields() {
		if _, ok := schemaKeyToTUIField(f.I18nKey, "en"); !ok {
			t.Errorf("schema field %q has no TUI bridge label path (AC-WC10-010)", f.Name)
		}
	}

	// (b) The wizard source binds value variables for the 7 nested fields + the
	// statusline segments + the launch/quality selects.
	data, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	src := string(data)
	bindings := []string{
		// Identity / Language / Launch
		"&userName", "&gitCommitLang", "&codeCommentLang", "&docLang",
		"&model", "&modelPolicy", "&effortLevel", "&permissionMode",
		// Statusline
		"&statuslineTheme", "&statuslineSegmentsSelection",
		// Quality scalar + 3 nested
		"&developmentMode", "&nestedCoverageTarget", "&nestedEnforceQuality", "&nestedMinCoverage",
		// Git convention scalar + 4 nested
		"&gitConvention", "&nestedAutoDetection", "&nestedConfidence", "&nestedSampleSize", "&nestedEnforceOnPush",
	}
	for _, b := range bindings {
		if !strings.Contains(src, b) {
			t.Errorf("profile_setup.go must bind a widget to %s (AC-WC10-010 TUI field coverage)", b)
		}
	}
}
