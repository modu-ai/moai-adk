package cli

import (
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/settings"
	"github.com/modu-ai/moai-adk/internal/template"
)

// languageOptionLabels carries the native-name label and description of each
// conversation-language value, the same text the init wizard's
// conversation_language question shows. Native names are never translated.
var languageOptionLabels = map[string]wizard.Option{
	"en": {Label: "English", Desc: "English"},
	"ko": {Label: "Korean (한국어)", Desc: "한국어"},
	"ja": {Label: "Japanese (日本語)", Desc: "日本語"},
	"zh": {Label: "Chinese (中文)", Desc: "中文"},
}

// profileLanguageOptions lists the schema's conversation-language values with their
// native-name labels. A value the label table does not know is shown as-is.
func profileLanguageOptions() []wizard.Option {
	defs := settings.FieldOptionDefs("conversation_lang")
	opts := make([]wizard.Option, 0, len(defs))
	for _, d := range defs {
		o, ok := languageOptionLabels[d.Value]
		if !ok {
			o = wizard.Option{Label: d.Value}
		}
		o.Value = d.Value
		opts = append(opts, o)
	}
	return opts
}

// profileModelPolicyOptions lists the CLI-only model_policy field: the schema's empty
// label plus template.ValidModelPolicies(), labelled from t.
func profileModelPolicyOptions(t profileSetupText) []wizard.Option {
	labels := map[string]string{
		string(template.ModelPolicyHigh):   t.ModelPolicyHigh,
		string(template.ModelPolicyMedium): t.ModelPolicyMedium,
		string(template.ModelPolicyLow):    t.ModelPolicyLow,
	}
	opts := []wizard.Option{{Label: settings.EmptyLabelFor("model_policy"), Value: ""}}
	for _, v := range template.ValidModelPolicies() {
		label := labels[v]
		if label == "" {
			label = v
		}
		opts = append(opts, wizard.Option{Label: label, Value: v})
	}
	return opts
}

// buildProfileOptions builds the option lists of the profile wizard select
// fields in t's locale (design.md §3): values from the settings schema (and
// template.ValidModelPolicies for model_policy), labels through the cli label
// bridge. The wizard package receives the lists as arguments, so it never
// resolves schema labels itself.
//
// @MX:NOTE: [AUTO] Wired by the M5 absorption — runProfileSetup builds these lists in the run's initial locale and hands them to the profileWizardRunner seam.
func buildProfileOptions(t profileSetupText) wizard.ProfileOptions {
	return wizard.ProfileOptions{
		Language:        profileLanguageOptions(),
		Model:           schemaSelectOptions(t, "model", true),
		ModelPolicy:     profileModelPolicyOptions(t),
		EffortLevel:     schemaSelectOptions(t, "effort_level", true),
		PermissionMode:  schemaSelectOptions(t, "permission_mode", false),
		DevelopmentMode: schemaSelectOptions(t, "development_mode", true),
	}
}
