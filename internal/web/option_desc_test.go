package web

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// t212: the audit fields guarded in audit_option_desc_test.go are not the only
// ones carrying per-option descriptions — report.format does too, and its keys
// were written by copying the enum-LABEL key shape (".opt."), which is exactly
// what that guard forbids: applyI18n resolves such a key against the ENGLISH
// dictionary in every locale, so the ko/ja/zh text sits in i18n.js and never
// renders. The tests here sweep EVERY section instead of naming one field set,
// so a future description-bearing field inherits the protection without anyone
// remembering to extend a list.

// allOptionDescFields returns every schema field declaring at least one
// OptionDesc, keyed by field name, across all sections.
func allOptionDescFields(t *testing.T) map[string]settings.FieldDef {
	t.Helper()
	found := make(map[string]settings.FieldDef)
	for _, section := range settings.AllSections() {
		for _, f := range settings.SectionFields(section) {
			for _, opt := range f.Options {
				if opt.OptionDesc != "" {
					found[f.Name] = f
					break
				}
			}
		}
	}
	// Floor: a sweep that finds nothing passes vacuously. These fields are known
	// to carry descriptions, so an absence means the sweep measures nothing.
	for _, name := range []string{
		"workflow.audit.model", "workflow.audit.gates.claude",
		"workflow.audit.gates.codex", "workflow.audit.gates.glm",
		"report.format",
	} {
		if _, ok := found[name]; !ok {
			t.Fatalf("sweep found no OptionDesc on field %q — the sweep is measuring nothing", name)
		}
	}
	return found
}

// TestEveryOptionDescKeyAvoidsOptGuard extends the audit-scoped
// TestAuditOptionDescKeysAvoidOptGuard to every description-bearing field.
func TestEveryOptionDescKeyAvoidsOptGuard(t *testing.T) {
	for name, f := range allOptionDescFields(t) {
		for _, opt := range f.Options {
			if strings.Contains(opt.OptionDesc, ".opt.") {
				t.Errorf("field %q option %q OptionDesc %q contains \".opt.\" — the G1-2 guard resolves it against the English dictionary, so its ko/ja/zh translation never renders",
					name, opt.Value, opt.OptionDesc)
			}
		}
	}
}

// TestEveryOptionDescKeyTranslatedInAllLocales asserts every description key
// resolves in all four locale blocks, for every description-bearing field.
func TestEveryOptionDescKeyTranslatedInAllLocales(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	seen := make(map[string]bool)
	for name, f := range allOptionDescFields(t) {
		for _, opt := range f.Options {
			if opt.OptionDesc == "" || seen[opt.OptionDesc] {
				continue
			}
			seen[opt.OptionDesc] = true
			if n := strings.Count(dict, `"`+opt.OptionDesc+`"`); n != 4 {
				t.Errorf("i18n.js declares key %q (field %q) %d times, want 4 (en/ko/ja/zh)", opt.OptionDesc, name, n)
			}
		}
	}
}

// TestEveryOptionDescRendersOnConsolePage is the render-path lock: the schema
// could declare an OptionDesc that schemaRadioRow never emits.
func TestEveryOptionDescRendersOnConsolePage(t *testing.T) {
	body := renderConsolePage(t)
	for name, f := range allOptionDescFields(t) {
		for _, opt := range f.Options {
			want := `data-i18n="` + opt.OptionDesc + `"`
			if !strings.Contains(body, want) {
				t.Errorf("rendered console page has no %s (field %q option %q)", want, name, opt.Value)
			}
		}
	}
}
