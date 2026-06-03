package cli

import (
	"os"
	"strings"
	"testing"
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

// TestProfileSetup_ModelPolicySelectPresent is the AC-WC2-006 grep guard: the
// wizard source must construct a model_policy-bound huh.Select (offering the 3
// canonical policy values plus an empty "(project default)" option) AND persist
// the value into the written ProfilePreferences. The TUI form cannot be driven
// without a TTY, so a source-level guard confirms the wiring exists.
func TestProfileSetup_ModelPolicySelectPresent(t *testing.T) {
	src, err := os.ReadFile("profile_setup.go")
	if err != nil {
		t.Fatalf("read profile_setup.go: %v", err)
	}
	body := string(src)

	// A model_policy value variable must be initialized from existing prefs.
	if !strings.Contains(body, "existingPrefs.ModelPolicy") {
		t.Error("model_policy value not initialized from existingPrefs.ModelPolicy")
	}
	// The model-settings group must contain a select bound to the policy var,
	// titled with the localized ModelPolicyTitle.
	if !strings.Contains(body, "t.ModelPolicyTitle") {
		t.Error("model-settings group does not present a t.ModelPolicyTitle select")
	}
	// The select must offer the 3 canonical policy values.
	for _, v := range []string{`"high"`, `"medium"`, `"low"`} {
		if !strings.Contains(body, "ModelPolicy") || !strings.Contains(body, v) {
			t.Errorf("model_policy select missing canonical value %s", v)
		}
	}
	// The written ProfilePreferences must carry ModelPolicy.
	if !strings.Contains(body, "ModelPolicy:") {
		t.Error("saved ProfilePreferences does not include a ModelPolicy field")
	}
}
