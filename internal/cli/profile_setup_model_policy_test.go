package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// TestProfileText_ModelPolicyLabels verifies AC-WC2-006: the model_policy select
// labels are non-empty for all four locales (en/ko/ja/zh). The strings predate
// SPEC-WEB-CONSOLE-002 in the profileSetupText struct; this guards that the
// wizard's new model_policy select has localized labels to render.
func TestProfileText_ModelPolicyLabels(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		t.Run(lang, func(t *testing.T) {
			txt := getProfileText(lang)
			for name, val := range map[string]string{
				"ModelPolicyTitle":  txt.ModelPolicyTitle,
				"ModelPolicyDesc":   txt.ModelPolicyDesc,
				"ModelPolicyHigh":   txt.ModelPolicyHigh,
				"ModelPolicyMedium": txt.ModelPolicyMedium,
				"ModelPolicyLow":    txt.ModelPolicyLow,
			} {
				if val == "" {
					t.Errorf("lang %q: %s is empty", lang, name)
				}
			}
		})
	}
}

// TestProfileSetup_ModelPolicySelectPresent is the AC-ITI-010 S1 guard,
// re-aimed per design.md §10 from a source grep to a behavior test over the
// ABSORBED question set: the profile wizard asks model_policy with the three
// canonical policy values plus the schema's empty option (labels non-empty),
// and the chosen policy reaches preferences.yaml through the wizard save path
// — the persistence half is shared with
// TestProfileSetupAbsorbed_SavePersistsAcrossSurfaces, which asserts the
// saved ModelPolicy value end to end.
func TestProfileSetup_ModelPolicySelectPresent(t *testing.T) {
	opts := buildProfileOptions(getProfileText("en"))
	qs := wizard.ProfileQuestions(opts, wizard.ProfileResult{})
	q := wizard.QuestionByID(qs, "model_policy")
	if q == nil {
		t.Fatal("the absorbed profile question set has no model_policy question")
	}

	offered := map[string]string{}
	for _, o := range q.Options {
		offered[o.Value] = o.Label
	}
	// The empty option must exist (AC-ITI-005 (4)); its label is the schema's
	// EmptyLabelFor accessor, which returns "" while model_policy has no
	// schema field — a documented pre-existing blank, not asserted non-empty.
	for _, v := range []string{"", "high", "medium", "low"} {
		label, ok := offered[v]
		if !ok {
			t.Errorf("model_policy select does not offer the canonical value %q", v)
			continue
		}
		if v != "" && strings.TrimSpace(label) == "" {
			t.Errorf("model_policy option %q renders an empty label", v)
		}
	}
}
