package web

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestProjectContinuationI18nKeysInAllLocales is the AC-PCK-011 enforcer, and
// it exists because NO existing test can fire on the omission it guards.
//
// Measured on this tree, with the field declared as a bare closedSeam (no
// withOptionDesc wrapper) and only its five label keys in i18n.js, the entire
// internal/web and internal/settings suites pass. The three all-sections
// sweeps in option_desc_test.go iterate allOptionDescFields, which skips any
// field carrying no OptionDesc (`if opt.OptionDesc != ""`, option_desc_test.go
// :27-28), so they pass VACUOUSLY here; and TestI18nKeyCoverageForward /
// Reverse compare the four locale maps to each other rather than the schema to
// the dictionary, so five declared keys landing in all four maps is a pass.
// A REQ-PCK-011 under-delivery would therefore ship silently green.
//
// (TestI18nUntranslatedValues does fire on this field, but on a different
// axis — identical `.opt.` label VALUES across locales — and satisfying it
// leaves the missing per-option descriptions untouched. It is not a substitute
// enforcer.)
//
// Conjunct (a) below is the one that fires on the missing wrapper.
func TestProjectContinuationI18nKeysInAllLocales(t *testing.T) {
	const fieldName = "workflow.project.continuation"

	// (a) The FieldDef carries a non-empty OptionDesc on all three options.
	//     This is what goes red when the withOptionDesc wrapper is dropped.
	var field *settings.FieldDef
	for _, section := range settings.AllSections() {
		for _, f := range settings.SectionFields(section) {
			if f.Name == fieldName {
				fCopy := f
				field = &fCopy
			}
		}
	}
	if field == nil {
		t.Fatalf("schema declares no field %q", fieldName)
	}
	if len(field.Options) != 3 {
		t.Fatalf("field %q has %d options, want 3", fieldName, len(field.Options))
	}
	for _, opt := range field.Options {
		if opt.OptionDesc == "" {
			t.Errorf("field %q option %q carries no OptionDesc — the withOptionDesc wrapper is missing, and no existing sweep can catch that (they skip fields with no OptionDesc)",
				fieldName, opt.Value)
		}
	}

	// (b) All eight keys resolve in all four locale maps: title, desc, three
	//     `.opt.` labels, three `.option.` descriptions. 8 x 4 = 32 entries.
	keys := []string{
		"f.workflow.project.continuation.title",
		"f.workflow.project.continuation.desc",
		"f.workflow.project.continuation.opt.none",
		"f.workflow.project.continuation.opt.card",
		"f.workflow.project.continuation.opt.pipeline",
		"f.workflow.project.continuation.option.none",
		"f.workflow.project.continuation.option.card",
		"f.workflow.project.continuation.option.pipeline",
	}
	for _, k := range keys {
		if !i18nKeyInAllLocales(t, k) {
			t.Errorf("i18n.js missing key %q in all 4 locales", k)
		}
	}
}
