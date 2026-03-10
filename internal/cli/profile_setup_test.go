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
			if text.ThemeDefault == "" {
				t.Errorf("lang %q: ThemeDefault is empty", lang)
			}
			if text.ThemeCatppuccinMocha == "" {
				t.Errorf("lang %q: ThemeCatppuccinMocha is empty", lang)
			}
			if text.ThemeCatppuccinLatte == "" {
				t.Errorf("lang %q: ThemeCatppuccinLatte is empty", lang)
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

// TestGetSegmentDefault verifies the helper handles nil and missing keys correctly.
func TestGetSegmentDefault(t *testing.T) {
	tests := []struct {
		name     string
		segments map[string]bool
		key      string
		def      bool
		want     bool
	}{
		{"nil map returns default true", nil, "model", true, true},
		{"nil map returns default false", nil, "model", false, false},
		{"present key returns its value", map[string]bool{"model": false}, "model", true, false},
		{"missing key returns default", map[string]bool{"context": true}, "model", true, true},
		{"missing key returns false default", map[string]bool{"context": true}, "model", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSegmentDefault(tt.segments, tt.key, tt.def)
			if got != tt.want {
				t.Errorf("getSegmentDefault(%v, %q, %v) = %v, want %v",
					tt.segments, tt.key, tt.def, got, tt.want)
			}
		})
	}
}
