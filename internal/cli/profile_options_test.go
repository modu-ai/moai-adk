package cli

import (
	"slices"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/template"
)

var profileTestLocales = []string{"en", "ko", "ja", "zh"}

// schemaValues is the wire-value list of a settings schema select field.
func schemaValues(field string) []string {
	var vs []string
	for _, d := range settings.FieldOptionDefs(field) {
		vs = append(vs, d.Value)
	}
	return vs
}

// profileQuestionOptions indexes the profile question set built from
// buildProfileOptions by question id.
func profileQuestionOptions(locale string) map[string][]wizard.Option {
	qs := wizard.ProfileQuestions(buildProfileOptions(getProfileText(locale)), wizard.ProfileResult{})
	byID := make(map[string][]wizard.Option, len(qs))
	for _, q := range qs {
		if q.Type == wizard.QuestionTypeSelect {
			byID[q.ID] = q.Options
		}
	}
	return byID
}

// TestProfileOptions_ValueSetsEqualSchema — AC-ITI-005 (4): the option value
// set of every profile select field equals the settings schema (plus the
// empty value where the wizard offers one), in every locale.
func TestProfileOptions_ValueSetsEqualSchema(t *testing.T) {
	want := map[string][]string{
		"conversation_language": schemaValues("conversation_lang"),
		"git_commit_lang":       schemaValues("conversation_lang"),
		"code_comment_lang":     schemaValues("conversation_lang"),
		"doc_lang":              schemaValues("conversation_lang"),
		"model":                 append([]string{""}, schemaValues("model")...),
		"effort_level":          append([]string{""}, schemaValues("effort_level")...),
		"development_mode":      append([]string{""}, schemaValues("development_mode")...),
		"permission_mode":       schemaValues("permission_mode"),
		"model_policy":          append([]string{""}, template.ValidModelPolicies()...),
	}
	// Positive existence: the schema supplies values for every field.
	for id, vs := range want {
		if len(vs) < 2 {
			t.Fatalf("expected value list for %q has %d entries; the schema lost the field", id, len(vs))
		}
	}
	for _, loc := range profileTestLocales {
		got := profileQuestionOptions(loc)
		if len(got) != len(want) {
			t.Errorf("locale %s: %d select questions, want %d (%v)", loc, len(got), len(want), got)
		}
		for id, wantVals := range want {
			opts, ok := got[id]
			if !ok || len(opts) == 0 {
				t.Errorf("locale %s: select %q has no options", loc, id)
				continue
			}
			var gotVals []string
			for _, o := range opts {
				gotVals = append(gotVals, o.Value)
			}
			if len(gotVals) != len(slices.Compact(slices.Sorted(slices.Values(gotVals)))) {
				t.Errorf("locale %s: select %q repeats a value: %q", loc, id, gotVals)
			}
			if !slices.Equal(slices.Sorted(slices.Values(gotVals)), slices.Sorted(slices.Values(wantVals))) {
				t.Errorf("locale %s: select %q values %q, want the set %q", loc, id, gotVals, wantVals)
			}
		}
	}
}

// TestProfileOptions_Labels pins where the labels come from: language labels
// are the init wizard's native names (never translated), schema options go
// through the label bridge, and model_policy labels are the localized
// profileSetupText strings.
func TestProfileOptions_Labels(t *testing.T) {
	wantLang := []string{"English", "Korean (한국어)", "Japanese (日本語)", "Chinese (中文)"}
	for _, loc := range profileTestLocales {
		txt := getProfileText(loc)
		got := profileQuestionOptions(loc)

		for _, id := range []string{"conversation_language", "git_commit_lang", "code_comment_lang", "doc_lang"} {
			var labels []string
			for _, o := range got[id] {
				labels = append(labels, o.Label)
			}
			if !slices.Equal(labels, wantLang) {
				t.Errorf("locale %s: %s labels %q, want %q", loc, id, labels, wantLang)
			}
		}

		policy := map[string]string{}
		for _, o := range got["model_policy"] {
			policy[o.Value] = o.Label
		}
		for v, l := range map[string]string{"high": txt.ModelPolicyHigh, "medium": txt.ModelPolicyMedium, "low": txt.ModelPolicyLow, "": settings.EmptyLabelFor("model_policy")} {
			if policy[v] != l {
				t.Errorf("locale %s: model_policy %q label %q, want %q", loc, v, policy[v], l)
			}
		}

		for _, id := range []string{"model", "effort_level", "permission_mode", "development_mode"} {
			for _, o := range got[id] {
				if o.Value == "" {
					continue
				}
				d := settings.OptionDef{Value: o.Value}
				for _, def := range settings.FieldOptionDefs(id) {
					if def.Value == o.Value {
						d = def
					}
				}
				if want := optionLabelFor(txt, d); o.Label != want {
					t.Errorf("locale %s: %s option %q label %q, want the bridge label %q", loc, id, o.Value, o.Label, want)
				}
			}
		}
	}
}

// TestHuhV1Options_PreservesLabelAndValue — the v1 adapter keeps each
// option's label and value in order.
func TestHuhV1Options_PreservesLabelAndValue(t *testing.T) {
	in := schemaSelectOptions(getProfileText("ko"), "model", true)
	if len(in) < 2 {
		t.Fatalf("schemaSelectOptions(model) returned %d options", len(in))
	}
	out := huhV1Options(in)
	if len(out) != len(in) {
		t.Fatalf("adapter returned %d options, want %d", len(out), len(in))
	}
	for i := range in {
		if out[i].Key != in[i].Label || out[i].Value != in[i].Value {
			t.Errorf("option %d = {%q, %q}, want {%q, %q}", i, out[i].Key, out[i].Value, in[i].Label, in[i].Value)
		}
	}
}
