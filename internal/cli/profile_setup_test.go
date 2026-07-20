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

// TestGetProfileText_ModeFields + TestNormalizeStatuslinePreset removed
// (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001): the mode + preset i18n fields and
// the normalizeStatuslinePreset helper were deleted. Theme fields are covered
// by TestGetProfileText_ThemeFields above.

// TestStatuslineAllSegments_CardinalityAndOrder verifies the 16-segment
// canonical list matches .moai/config/sections/statusline.yaml keys
// (cache_hit added by SPEC-WEB-CONSOLE-011 M6).
// SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001 REQ-SPW-002 / REQ-SPW-003.
func TestStatuslineAllSegments_CardinalityAndOrder(t *testing.T) {
	if got, want := len(statuslineAllSegments), 16; got != want {
		t.Fatalf("statuslineAllSegments length = %d, want %d", got, want)
	}
	expected := []string{
		"cache_hit", "claude_version", "context", "directory",
		"effort_thinking", "git_branch", "git_status", "moai_version",
		"model", "output_style", "pr", "session_time",
		"task", "usage_5h", "usage_7d", "worktree",
	}
	for i, key := range expected {
		if statuslineAllSegments[i] != key {
			t.Errorf("statuslineAllSegments[%d] = %q, want %q", i, statuslineAllSegments[i], key)
		}
	}
}
