package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// schemaBackedSelects lists the wizard selects whose option lists are derived
// from the shared settings schema, with the withEmpty flag each call site passes.
// model_policy is intentionally absent: the web console dropped that field, so it
// has no schema entry and keeps an inline option list in profile_setup.go.
// git_convention is absent because its Select was removed from the wizard — the
// schema still declares the field (the web console renders it), but the TUI no
// longer builds an option list for it.
var schemaBackedSelects = []struct {
	field     string
	withEmpty bool
}{
	{"model", true},
	{"effort_level", true},
	{"permission_mode", false},
	{"development_mode", true},
}

// TestSchemaSelectOptions_DerivedFromSchema is the anti-drift guard for the
// wizard's option lists: every schema-backed select must offer exactly the schema's
// option VALUES, in the schema's order. Re-introducing an inline list (the drift
// that made the wizard write canonical model ids the web console rejects) fails
// here as soon as the two lists disagree.
func TestSchemaSelectOptions_DerivedFromSchema(t *testing.T) {
	txt := getProfileText("en")
	for _, sel := range schemaBackedSelects {
		t.Run(sel.field, func(t *testing.T) {
			defs := settings.FieldOptionDefs(sel.field)
			if len(defs) == 0 {
				t.Fatalf("schema field %q exposes no options", sel.field)
			}
			want := make([]string, 0, len(defs)+1)
			if sel.withEmpty {
				want = append(want, "")
			}
			for _, d := range defs {
				want = append(want, d.Value)
			}

			opts := schemaSelectOptions(txt, sel.field, sel.withEmpty)
			got := make([]string, 0, len(opts))
			for _, o := range opts {
				got = append(got, o.Value)
			}
			if strings.Join(got, "|") != strings.Join(want, "|") {
				t.Errorf("option values = %q, want %q", got, want)
			}
		})
	}
}

// TestSchemaSelectOptions_EmptyLabelFromSchema asserts the empty option carries the
// canonical label from settings.EmptyLabelFor (NOT a wizard-local duplicate), and
// that permission_mode still offers no empty option — the wizard defaults it to
// acceptEdits and normalizes that back to "" on save.
func TestSchemaSelectOptions_EmptyLabelFromSchema(t *testing.T) {
	txt := getProfileText("en")
	for _, sel := range schemaBackedSelects {
		opts := schemaSelectOptions(txt, sel.field, sel.withEmpty)
		if !sel.withEmpty {
			for _, o := range opts {
				if o.Value == "" {
					t.Errorf("field %q: unexpected empty option %q", sel.field, o.Label)
				}
			}
			continue
		}
		want := settings.EmptyLabelFor(sel.field)
		if want == "" {
			t.Fatalf("field %q declares no empty label but the wizard requests one", sel.field)
		}
		if opts[0].Value != "" || opts[0].Label != want {
			t.Errorf("field %q: first option = {%q, %q}, want {%q, \"\"}", sel.field, opts[0].Label, opts[0].Value, want)
		}
	}
}

// TestSchemaSelectOptions_ModelValuesSurviveNormalizer closes the loop between the
// picker and the migration normalizer: a value the wizard offers must round-trip
// through normalizeModel unchanged, otherwise a saved preference would be silently
// rewritten (or the Select would fail to highlight the stored value) on the next run.
func TestSchemaSelectOptions_ModelValuesSurviveNormalizer(t *testing.T) {
	for _, o := range schemaSelectOptions(getProfileText("en"), "model", false) {
		if got := normalizeModel(o.Value); got != o.Value {
			t.Errorf("normalizeModel(%q) = %q, want the value unchanged", o.Value, got)
		}
	}
}

// TestSchemaSelectOptions_Localized verifies the bridged option labels resolve to
// localized prose rather than the bare wire value, for every select the wizard
// still renders. The git_convention wire-value-fallback expectation was dropped
// with that Select — the TUI no longer renders it, so the wizard has no labels to
// assert; the field's web-side labels are covered in internal/web.
func TestSchemaSelectOptions_Localized(t *testing.T) {
	for _, lang := range fourLocales {
		txt := getProfileText(lang)
		for _, field := range []string{"model", "effort_level", "permission_mode", "development_mode"} {
			for _, o := range schemaSelectOptions(txt, field, false) {
				if strings.TrimSpace(o.Label) == "" {
					t.Errorf("lang=%q field=%q value=%q: empty label", lang, field, o.Value)
				}
				if o.Label == o.Value {
					t.Errorf("lang=%q field=%q value=%q: label fell back to the wire value (missing schemaOptionBridge entry)", lang, field, o.Value)
				}
			}
		}
	}
}

// TestModelPolicyLabels_AgreeWithProfileMatrix was removed with the per-agent
// profile matrix (SPEC-AGENT-MODEL-INHERIT-001 M5): its derivation source (the
// deleted template matrix) no longer exists. The "Agent model policy" tier
// lines are now hand-written prose with no matrix to derive from; their
// main-session rewording is design H24 (doctrine-text surface), outside M5.
