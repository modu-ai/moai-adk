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
//
// SPEC-WEB-CONSOLE-011 M2b 범위 조정: 10섹션 확장 필드(PersistSeam /
// PersistTypedSection)는 기술 식별자 key chip으로 렌더하며 data-i18n을 방출하지
// 않는다 — i18n.js 헤더 계약("Field identifiers stay in English as code chips
// and are NOT translated")과 동일 원칙. 따라서 per-field title/desc 사전 항목
// 의무는 data-i18n을 실제 방출하는 기존 34-필드 위젯에만 적용된다. 렌더된
// data-i18n 키의 사전 존재는 TestDataI18nKeysSubsetOfDictionary가 전수 강제한다.
//
// M3 redesign: statusline schema fields (statusline_theme + statusline_segment.*)
// remain in the schema but their web i18n keys (f.statusline_theme.* / seg.*) are
// intentionally removed — the web console no longer renders the section. They are
// skipped here; the TUI half (internal/cli TestI18nKeySetParity) still resolves
// them through the bridge.
func TestI18nKeySetParity(t *testing.T) {
	// Statusline fields stay in the schema (TUI path) but their web i18n keys are
	// removed (M3 redesign) — skip them in the web dictionary parity check. The
	// git_strategy section was removed from the web console (tab + panel + its
	// i18n keys); its fields stay in the schema (config in git-strategy.yaml) but
	// carry no web i18n keys, so they are skipped here too.
	webRemovedFields := map[string]bool{}
	for _, f := range settings.SectionFields(settings.SectionStatusline) {
		webRemovedFields[f.Name] = true
	}
	for _, f := range settings.SectionFields(settings.SectionGitStrategy) {
		webRemovedFields[f.Name] = true
	}

	for _, f := range settings.AllFields() {
		// M5-b D3: PersistSeam / PersistTypedSection 필드도 이제 field__title /
		// field__desc data-i18n 을 방출한다 (과거 key-chip 전용에서 승격). 따라서
		// 이 필드들의 .title / .desc 도 사전에 존재해야 한다.
		if webRemovedFields[f.Name] {
			continue // statusline / git_strategy: schema-preserved, web-i18n-removed
		}
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
		// M5-b D4: select 필드의 각 option 도 data-i18n 을 방출한다. OptionDef.I18nKey
		// 가 비어있지 않은 모든 option 의 i18n 키가 4-locale 사전에 존재해야 한다.
		for _, opt := range f.Options {
			if opt.I18nKey == "" {
				continue
			}
			if !i18nKeyInAllLocales(t, opt.I18nKey) {
				t.Errorf("i18n.js missing option key %q in all 4 locales (field %q)", opt.I18nKey, f.Name)
			}
		}
		// EmptyLabelKey 도 비어있지 않으면 사전에 존재해야 한다.
		if f.EmptyLabelKey != "" {
			if !i18nKeyInAllLocales(t, f.EmptyLabelKey) {
				t.Errorf("i18n.js missing empty-label key %q in all 4 locales (field %q)", f.EmptyLabelKey, f.Name)
			}
		}
	}
}

// TestI18nSegmentKeysRemovedFromWebDictionary covers the M3 redesign: the seg.*
// segment label keys are removed from the web i18n dictionary across all 4
// locales (the Statusline section is no longer rendered in the web console). The
// canonical segment keys themselves remain in the schema (StatuslineSegmentKeys)
// for the TUI / profile_setup CLI path.
func TestI18nSegmentKeysRemovedFromWebDictionary(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	for _, seg := range settings.StatuslineSegmentKeys() {
		key := "seg." + seg
		if count := strings.Count(dict, `"`+key+`":`); count != 0 {
			t.Errorf("i18n.js must NOT carry removed segment key %q (found %d occurrence(s))", key, count)
		}
	}
}
