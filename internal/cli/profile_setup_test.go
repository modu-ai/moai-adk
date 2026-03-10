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
// REQ-V3-MODE-003: 프로필 위저드는 compact/default/full 이름을 표시한다.
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
			// v3 모드 레이블 검증
			if text.ModeDefault == "" {
				t.Errorf("lang %q: ModeDefault is empty", lang)
			}
			if text.ModeCompact == "" {
				t.Errorf("lang %q: ModeCompact is empty", lang)
			}
			if text.ModeFull == "" {
				t.Errorf("lang %q: ModeFull is empty", lang)
			}
			// 하위 호환성을 위한 deprecated 필드도 검증
			if text.ModeVerbose == "" {
				t.Errorf("lang %q: ModeVerbose is empty", lang)
			}
			if text.ModeMinimal == "" {
				t.Errorf("lang %q: ModeMinimal is empty", lang)
			}
		})
	}
}
