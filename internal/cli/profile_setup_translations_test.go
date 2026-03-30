package cli

import "testing"

func TestGetProfileText_AllLanguages(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		t.Run(lang, func(t *testing.T) {
			txt := getProfileText(lang)
			if txt.ConfiguringProfile == "" {
				t.Errorf("ConfiguringProfile is empty for lang %q", lang)
			}
			if txt.LangSelectTitle == "" {
				t.Errorf("LangSelectTitle is empty for lang %q", lang)
			}
			if txt.UserNameTitle == "" {
				t.Errorf("UserNameTitle is empty for lang %q", lang)
			}
			if txt.SetupCancelled == "" {
				t.Errorf("SetupCancelled is empty for lang %q", lang)
			}
			if txt.SavedProfile == "" {
				t.Errorf("SavedProfile is empty for lang %q", lang)
			}
			if txt.ModelPolicyHigh == "" {
				t.Errorf("ModelPolicyHigh is empty for lang %q", lang)
			}
			if txt.ModelOpus1M == "" {
				t.Errorf("ModelOpus1M is empty for lang %q", lang)
			}
			if txt.ModelSonnet1M == "" {
				t.Errorf("ModelSonnet1M is empty for lang %q", lang)
			}
		})
	}
}

func TestGetProfileText_FallbackToEnglish(t *testing.T) {
	en := getProfileText("en")
	fallback := getProfileText("unknown")
	if fallback.ConfiguringProfile != en.ConfiguringProfile {
		t.Errorf("fallback = %q, want %q", fallback.ConfiguringProfile, en.ConfiguringProfile)
	}
}

func TestGetProfileText_EmptyString(t *testing.T) {
	en := getProfileText("en")
	empty := getProfileText("")
	if empty.LangSelectTitle != en.LangSelectTitle {
		t.Errorf("empty string fallback = %q, want %q", empty.LangSelectTitle, en.LangSelectTitle)
	}
}
