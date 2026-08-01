package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/template"
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
					t.Errorf("field %q: unexpected empty option %q", sel.field, o.Key)
				}
			}
			continue
		}
		want := settings.EmptyLabelFor(sel.field)
		if want == "" {
			t.Fatalf("field %q declares no empty label but the wizard requests one", sel.field)
		}
		if opts[0].Value != "" || opts[0].Key != want {
			t.Errorf("field %q: first option = {%q, %q}, want {%q, \"\"}", sel.field, opts[0].Key, opts[0].Value, want)
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
				if strings.TrimSpace(o.Key) == "" {
					t.Errorf("lang=%q field=%q value=%q: empty label", lang, field, o.Value)
				}
				if o.Key == o.Value {
					t.Errorf("lang=%q field=%q value=%q: label fell back to the wire value (missing schemaOptionBridge entry)", lang, field, o.Value)
				}
			}
		}
	}
}

// effortRank orders the effort vocabulary so the model-policy prose can be checked
// against the matrix's actual span.
var effortRank = map[string]int{
	template.EffortLevelLow:    0,
	template.EffortLevelMedium: 1,
	template.EffortLevelHigh:   2,
	template.EffortLevelXHigh:  3,
	template.EffortLevelMax:    4,
}

// TestModelPolicyLabels_AgreeWithProfileMatrix keeps the three hand-written
// "Agent model policy" tier lines honest against template.defaultProfileMatrix,
// which is the authoritative per-agent {model, effort} SSOT.
//
// The prose is kept hand-written rather than generated: a generated line would
// have to spell out the sonnet agent names (4 of them in the low column), which is
// longer than the huh option line usefully holds, and the editorial grouping
// ("single-shot rows", "docs/e2e") is a human summary, not a matrix fact. This test
// is the rot guard instead — it derives the opus effort SPAN, the sonnet effort,
// and the sonnet row membership from the matrix and fails when the prose disagrees.
func TestModelPolicyLabels_AgreeWithProfileMatrix(t *testing.T) {
	matrix := template.DefaultProfileMatrix()
	opusDisplay := "Opus " + strings.TrimPrefix(template.ModelAliasCanonicalID("opus"), "claude-opus-")

	for _, tc := range []struct {
		profile string
		label   func(profileSetupText) string
	}{
		{template.PerformanceTierHigh, func(p profileSetupText) string { return p.ModelPolicyHigh }},
		{template.PerformanceTierMedium, func(p profileSetupText) string { return p.ModelPolicyMedium }},
		{template.PerformanceTierLow, func(p profileSetupText) string { return p.ModelPolicyLow }},
	} {
		t.Run(tc.profile, func(t *testing.T) {
			cells, ok := matrix[tc.profile]
			if !ok {
				t.Fatalf("profile %q absent from the matrix", tc.profile)
			}

			hi, lo := "", ""
			sonnetEffort := ""
			sonnetRows := map[string]bool{}
			for agent, cell := range cells {
				switch cell.Model {
				case "opus":
					if hi == "" || effortRank[cell.Effort] > effortRank[hi] {
						hi = cell.Effort
					}
					if lo == "" || effortRank[cell.Effort] < effortRank[lo] {
						lo = cell.Effort
					}
				case "sonnet":
					sonnetRows[agent] = true
					sonnetEffort = cell.Effort
				default:
					t.Fatalf("unexpected model %q for agent %q — the tier prose only describes opus/sonnet", cell.Model, agent)
				}
			}

			wantSpan := "(" + hi + "~" + lo + ")"
			wantSonnet := "Sonnet (" + sonnetEffort

			for _, lang := range fourLocales {
				label := tc.label(getProfileText(lang))
				if !strings.Contains(label, opusDisplay) {
					t.Errorf("lang=%q label %q should name %q", lang, label, opusDisplay)
				}
				if !strings.Contains(label, wantSpan) {
					t.Errorf("lang=%q label %q should state the matrix opus effort span %s", lang, label, wantSpan)
				}
				if !strings.Contains(label, wantSonnet) {
					t.Errorf("lang=%q label %q should state %q...", lang, label, wantSonnet)
				}
				// The editorial row grouping must track which agents are actually on
				// sonnet in this column.
				if got, want := strings.Contains(label, "docs"), sonnetRows["manager-docs"]; got != want {
					t.Errorf("lang=%q label %q mentions docs=%v but manager-docs on sonnet=%v", lang, label, got, want)
				}
				if got, want := strings.Contains(label, "e2e"), sonnetRows["e2e-tester"]; got != want {
					t.Errorf("lang=%q label %q mentions e2e=%v but e2e-tester on sonnet=%v", lang, label, got, want)
				}
				for _, retired := range []string{"haiku", "Haiku", "fable", "Fable"} {
					if strings.Contains(label, retired) {
						t.Errorf("lang=%q label %q names %q, which has no cell in the matrix", lang, label, retired)
					}
				}
			}
		})
	}
}
