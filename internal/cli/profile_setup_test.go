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

// TestNormalizeStatuslinePreset verifies the canonical-list normalizer
// behavior covering EC-SPW-001 (preserve valid), EC-SPW-002 (reset invalid),
// EC-SPW-003 (preserve empty).
// SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001 REQ-SPW-004.
func TestNormalizeStatuslinePreset(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "EC-SPW-003 empty preserved", input: "", want: ""},
		{name: "EC-SPW-001 valid full", input: "full", want: "full"},
		{name: "EC-SPW-001 valid compact", input: "compact", want: "compact"},
		{name: "EC-SPW-001 valid minimal", input: "minimal", want: "minimal"},
		{name: "EC-SPW-001 valid custom", input: "custom", want: "custom"},
		{name: "EC-SPW-002 invalid legacy fullbar", input: "fullbar", want: ""},
		{name: "EC-SPW-002 invalid typo cusotm", input: "cusotm", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeStatuslinePreset(tt.input)
			if got != tt.want {
				t.Errorf("normalizeStatuslinePreset(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestStatuslineAllSegments_CardinalityAndOrder verifies the 15-segment
// canonical list matches .moai/config/sections/statusline.yaml keys.
// SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001 REQ-SPW-002 / REQ-SPW-003.
func TestStatuslineAllSegments_CardinalityAndOrder(t *testing.T) {
	if got, want := len(statuslineAllSegments), 15; got != want {
		t.Fatalf("statuslineAllSegments length = %d, want %d", got, want)
	}
	expected := []string{
		"claude_version", "context", "directory", "effort_thinking",
		"git_branch", "git_status", "moai_version", "model",
		"output_style", "pr", "session_time", "task",
		"usage_5h", "usage_7d", "worktree",
	}
	for i, key := range expected {
		if statuslineAllSegments[i] != key {
			t.Errorf("statuslineAllSegments[%d] = %q, want %q", i, statuslineAllSegments[i], key)
		}
	}
}
