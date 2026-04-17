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
			// T-003: Opus 4.7 model label
			if txt.ModelOpus47 == "" {
				t.Errorf("ModelOpus47 is empty for lang %q", lang)
			}
			// T-003: Effort level UI strings
			if txt.EffortLevelTitle == "" {
				t.Errorf("EffortLevelTitle is empty for lang %q", lang)
			}
			if txt.EffortLevelDesc == "" {
				t.Errorf("EffortLevelDesc is empty for lang %q", lang)
			}
			if txt.EffortLevelDefault == "" {
				t.Errorf("EffortLevelDefault is empty for lang %q", lang)
			}
			if txt.EffortLevelXHigh == "" {
				t.Errorf("EffortLevelXHigh is empty for lang %q", lang)
			}
			if txt.EffortLevelMax == "" {
				t.Errorf("EffortLevelMax is empty for lang %q", lang)
			}
		})
	}
}

// TestGetProfileText_OpusModelValues verifies Opus 4.7 label contains the model ID.
func TestGetProfileText_OpusModelValues(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if !containsStr(txt.ModelOpus47, "claude-opus-4-7") {
			t.Errorf("lang=%q: ModelOpus47 %q does not contain model ID", lang, txt.ModelOpus47)
		}
	}
}

// TestGetProfileText_EffortLevelValues verifies effort level labels contain xhigh/max keywords.
func TestGetProfileText_EffortLevelValues(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		txt := getProfileText(lang)
		if !containsStr(txt.EffortLevelXHigh, "xhigh") {
			t.Errorf("lang=%q: EffortLevelXHigh %q does not contain 'xhigh'", lang, txt.EffortLevelXHigh)
		}
		if !containsStr(txt.EffortLevelMax, "max") {
			t.Errorf("lang=%q: EffortLevelMax %q does not contain 'max'", lang, txt.EffortLevelMax)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstr(s, sub))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
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
