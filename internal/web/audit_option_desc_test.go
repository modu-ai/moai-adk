package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// t206: the audit tab's option buttons carry English enum labels by deliberate
// design (G1-2, TestOptionLabelsStayEnglish) — that guard stays. What was
// missing is any explanation of what each option DOES, so an operator reading
// "Advisory" or "Multi (converge)" had nothing telling them how it acts.
//
// The explanation rides the purpose-built per-option surface: OptionDef.
// OptionDesc, which makes schemaRadioRow switch to the stacked layout and
// render <span class="seg__d" data-i18n={opt.OptionDesc}> under each label.
// Always visible — not a hover-only tooltip.

// auditOptionDescFields returns the audit fields that must carry per-option
// descriptions, paired with the key prefix their options resolve against.
func auditOptionDescFields() map[string]string {
	return map[string]string{
		"workflow.audit.model":        "f.workflow.audit.model.option.",
		"workflow.audit.gates.claude": "f.workflow.audit.gate.option.",
		"workflow.audit.gates.codex":  "f.workflow.audit.gate.option.",
		"workflow.audit.gates.glm":    "f.workflow.audit.gate.option.",
	}
}

func auditFieldsByName(t *testing.T) map[string]settings.FieldDef {
	t.Helper()
	byName := make(map[string]settings.FieldDef)
	for _, f := range settings.SectionFields(settings.SectionWorkflow) {
		byName[f.Name] = f
	}
	return byName
}

// TestAuditOptionsCarryOptionDesc asserts every audit enum option declares an
// OptionDesc key — the sole opt-in for the stacked per-option layout.
func TestAuditOptionsCarryOptionDesc(t *testing.T) {
	byName := auditFieldsByName(t)
	for name, prefix := range auditOptionDescFields() {
		f, ok := byName[name]
		if !ok {
			t.Fatalf("schema has no field %q", name)
		}
		if len(f.Options) == 0 {
			t.Fatalf("field %q declares no options", name)
		}
		for _, opt := range f.Options {
			if opt.OptionDesc == "" {
				t.Errorf("field %q option %q has no OptionDesc — the option renders with no explanation", name, opt.Value)
				continue
			}
			want := prefix + opt.Value
			if opt.OptionDesc != want {
				t.Errorf("field %q option %q OptionDesc = %q, want %q", name, opt.Value, opt.OptionDesc, want)
			}
		}
	}
}

// TestAuditOptionDescKeysAvoidOptGuard is the load-bearing constraint. applyI18n
// resolves any key containing ".opt." against the ENGLISH dictionary (G1-2), so
// a description key carrying that substring would be frozen to English in every
// locale — silently, since the text still renders. ".option." does not match
// ".opt." (the char after "opt" is "i"), which is why that shape is used.
func TestAuditOptionDescKeysAvoidOptGuard(t *testing.T) {
	byName := auditFieldsByName(t)
	for name := range auditOptionDescFields() {
		f, ok := byName[name]
		if !ok {
			t.Fatalf("schema has no field %q", name)
		}
		for _, opt := range f.Options {
			if strings.Contains(opt.OptionDesc, ".opt.") {
				t.Errorf("field %q option %q OptionDesc %q contains \".opt.\" — the G1-2 guard would freeze this description to English in every locale",
					name, opt.Value, opt.OptionDesc)
			}
		}
	}
}

// TestAuditOptionDescTranslatedInAllLocales asserts each description key
// resolves in all four locales. A key present only in en would render blank in
// ko/ja/zh, because applyI18n only writes textContent when the lookup is a
// non-empty string and <span class="seg__d"> carries no server-side baseline.
func TestAuditOptionDescTranslatedInAllLocales(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	byName := auditFieldsByName(t)

	seen := make(map[string]bool)
	for name, prefix := range auditOptionDescFields() {
		f := byName[name]
		for _, opt := range f.Options {
			key := prefix + opt.Value
			if seen[key] {
				continue
			}
			seen[key] = true
			// Four locale blocks ⇒ the quoted key must appear four times.
			if n := strings.Count(dict, `"`+key+`"`); n != 4 {
				t.Errorf("i18n.js declares key %q %d times, want 4 (en/ko/ja/zh)", key, n)
			}
		}
	}
}

// TestAuditRadioRendersStackedOptionDesc asserts the rendered page actually
// carries the per-option description spans. This is the render-path lock: the
// schema could declare OptionDesc while schemaRadioRow never emits it.
func TestAuditRadioRendersStackedOptionDesc(t *testing.T) {
	body := renderConsolePage(t)

	if !strings.Contains(body, "seg--stacked") {
		t.Error("console page renders no seg--stacked radio group — the per-option description layout never engaged")
	}
	byName := auditFieldsByName(t)
	for name, prefix := range auditOptionDescFields() {
		f := byName[name]
		for _, opt := range f.Options {
			want := `data-i18n="` + prefix + opt.Value + `"`
			if !strings.Contains(body, want) {
				t.Errorf("rendered console page has no %s (field %q option %q)", want, name, opt.Value)
			}
		}
	}
}

// TestOptionLabelsStayEnglishStillGuarded is a tripwire: this card must not
// weaken G1-2. If the guard is ever removed, the enum LABELS would start
// following the locale — the change this card explicitly declined to make.
func TestOptionLabelsStayEnglishStillGuarded(t *testing.T) {
	js := readEmbeddedAsset(t, "app.js")
	if !strings.Contains(js, `".opt."`) {
		t.Fatal(`app.js lost the ".opt." guard — enum option labels would follow the active locale, reversing the G1-2 decision this card upheld`)
	}
}
