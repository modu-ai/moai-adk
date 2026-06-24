package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestSchemaEmptyLabelParity covers AC-WC10-014 (web side): the web console renders
// the schema's canonical empty-option label for each field that has one, and the
// schema returns exactly one empty-option label per field. This is the web half of
// the cross-surface parity check — the TUI half lives in internal/cli
// (TestTUIEmptyLabelsSchemaSourced). Both surfaces sourcing from
// settings.EmptyLabelFor guarantees identical strings.
func TestSchemaEmptyLabelParity(t *testing.T) {
	body := renderConsolePage(t)

	// The 4 documented label drifts are resolved by single-sourcing. The web
	// renders these via the optSelect Empty arg / langOptionTags. Assert the schema
	// canonical label string appears in the rendered page for the select fields that
	// carry an empty option.
	cases := map[string]string{
		"model":            settings.EmptyLabelFor("model"),            // "(project default)"
		"effort_level":     settings.EmptyLabelFor("effort_level"),     // "(runtime default)"
		"development_mode": settings.EmptyLabelFor("development_mode"), // "(project default)"
	}
	for field, label := range cases {
		if label == "" {
			t.Errorf("schema returned empty EmptyLabel for %q", field)
			continue
		}
		if !strings.Contains(body, label) {
			t.Errorf("rendered page missing canonical empty label %q for field %q", label, field)
		}
	}

	// Schema returns exactly ONE empty label per field (not multiple drifting values).
	// Verify the 4 fields with documented drift each return a single non-empty label.
	for _, field := range []string{"model", "effort_level", "git_convention", "conversation_lang"} {
		if settings.EmptyLabelFor(field) == "" {
			t.Errorf("field %q must have a single canonical empty label (drift resolution)", field)
		}
	}
}

// webI18nKey reports whether the embedded i18n.js carries the given dotted key in
// all 4 locale blocks.
func i18nKeyInAllLocales(t *testing.T, key string) bool {
	t.Helper()
	dict := readEmbeddedAsset(t, "i18n.js")
	// Each locale block carries the key as `"<key>":`. There are 4 locales, so a
	// key present in all 4 appears at least 4 times.
	return strings.Count(dict, `"`+key+`":`) >= 4
}

// TestI18nKeySetParity covers AC-WC10-016 (web side): every schema i18n key (the
// f.* field keys + seg.* segment keys) has a matching flat dotted-key entry in
// window.MOAI_I18N for all 4 locales. The TUI half (bridge resolution) lives in
// internal/cli (TestI18nKeySetParity there).
func TestI18nKeySetParity(t *testing.T) {
	for _, f := range settings.AllFields() {
		if strings.HasPrefix(f.I18nKey, "seg.") {
			// Segment keys: the seg.<segment> label must exist in all 4 locales.
			if !i18nKeyInAllLocales(t, f.I18nKey) {
				t.Errorf("i18n.js missing segment key %q in all 4 locales", f.I18nKey)
			}
			continue
		}
		// Field keys: the web renders data-i18n="<f.key>.title" / ".desc". Both must
		// exist in all 4 locales.
		for _, suffix := range []string{".title", ".desc"} {
			key := f.I18nKey + suffix
			if !i18nKeyInAllLocales(t, key) {
				t.Errorf("i18n.js missing key %q in all 4 locales (schema field %q)", key, f.Name)
			}
		}
	}
}

// TestI18nSegmentKeysComplete covers AC-WC10-016 segment half (web side): all 15
// canonical segment keys (seg.<key>) exist in all 4 locales in i18n.js.
func TestI18nSegmentKeysComplete(t *testing.T) {
	for _, seg := range settings.StatuslineSegmentKeys() {
		key := "seg." + seg
		if !i18nKeyInAllLocales(t, key) {
			t.Errorf("i18n.js missing segment key %q in all 4 locales", key)
		}
	}
}
