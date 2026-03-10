package cli

import (
	"testing"
)

// TestGetProfileText_ThemeFields verifies that all supported languages
// include translations for the new statusline theme selector fields.
func TestGetProfileText_ThemeFields(t *testing.T) {
	langs := []string{"en", "ko", "ja", "zh"}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			text := getProfileText(lang)
			if text.StatuslineThemeTitle == "" {
				t.Errorf("lang %q: StatuslineThemeTitle is empty", lang)
			}
			if text.StatuslineThemeDesc == "" {
				t.Errorf("lang %q: StatuslineThemeDesc is empty", lang)
			}
			if text.ThemeMoaiDark == "" {
				t.Errorf("lang %q: ThemeMoaiDark is empty", lang)
			}
			if text.ThemeMoaiLight == "" {
				t.Errorf("lang %q: ThemeMoaiLight is empty", lang)
			}
		})
	}
}

// TestGetProfileText_ModeFields verifies that all supported languages
// include translations for the statusline mode selector fields.
// REQ-V3-MODE-003: Profile wizard must display compact/default/full mode names.
func TestGetProfileText_ModeFields(t *testing.T) {
	langs := []string{"en", "ko", "ja", "zh"}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			text := getProfileText(lang)
			if text.StatuslineModeTitle == "" {
				t.Errorf("lang %q: StatuslineModeTitle is empty", lang)
			}
			if text.StatuslineModeDesc == "" {
				t.Errorf("lang %q: StatuslineModeDesc is empty", lang)
			}
			// Validate v3 mode labels
			if text.ModeDefault == "" {
				t.Errorf("lang %q: ModeDefault is empty", lang)
			}
			if text.ModeCompact == "" {
				t.Errorf("lang %q: ModeCompact is empty", lang)
			}
			if text.ModeFull == "" {
				t.Errorf("lang %q: ModeFull is empty", lang)
			}
			// Also validate deprecated fields for backward compatibility
			if text.ModeVerbose == "" {
				t.Errorf("lang %q: ModeVerbose is empty", lang)
			}
			if text.ModeMinimal == "" {
				t.Errorf("lang %q: ModeMinimal is empty", lang)
			}
		})
	}
}

