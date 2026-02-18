package models

import (
	"sort"
	"testing"
)

func TestSupportedLanguages(t *testing.T) {
	t.Parallel()

	langs := SupportedLanguages()

	// Must return exactly the expected set of language codes.
	expected := []string{"ko", "en", "ja", "zh", "es", "fr", "de"}
	if len(langs) != len(expected) {
		t.Fatalf("SupportedLanguages() returned %d codes, want %d", len(langs), len(expected))
	}

	// Sort both slices so order does not matter.
	got := make([]string, len(langs))
	copy(got, langs)
	sort.Strings(got)
	want := make([]string, len(expected))
	copy(want, expected)
	sort.Strings(want)

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("SupportedLanguages() sorted[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSupportedLanguages_ConsistentWithLangNameMap(t *testing.T) {
	t.Parallel()

	// Every code returned by SupportedLanguages must have a matching LangNameMap entry.
	for _, code := range SupportedLanguages() {
		if _, ok := LangNameMap[code]; !ok {
			t.Errorf("SupportedLanguages() includes %q but LangNameMap does not", code)
		}
	}

	// Every LangNameMap key must appear in SupportedLanguages.
	set := make(map[string]struct{})
	for _, code := range SupportedLanguages() {
		set[code] = struct{}{}
	}
	for code := range LangNameMap {
		if _, ok := set[code]; !ok {
			t.Errorf("LangNameMap has key %q not in SupportedLanguages()", code)
		}
	}
}

func TestGetLanguageName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code string
		want string
	}{
		{"ko", "Korean (\ud55c\uad6d\uc5b4)"},
		{"en", "English"},
		{"ja", "Japanese (\u65e5\u672c\u8a9e)"},
		{"zh", "Chinese (\u4e2d\u6587)"},
		{"es", "Spanish (Espa\u00f1ol)"},
		{"fr", "French (Fran\u00e7ais)"},
		{"de", "German (Deutsch)"},
		// Invalid codes should fall back to "English".
		{"xx", "English"},
		{"", "English"},
		{"KO", "English"}, // case-sensitive
		{"eng", "English"},
	}

	for _, tt := range tests {
		t.Run("code_"+tt.code, func(t *testing.T) {
			t.Parallel()
			got := GetLanguageName(tt.code)
			if got != tt.want {
				t.Errorf("GetLanguageName(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsValidLanguageCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code string
		want bool
	}{
		{"ko", true},
		{"en", true},
		{"ja", true},
		{"zh", true},
		{"es", true},
		{"fr", true},
		{"de", true},
		// Invalid codes.
		{"xx", false},
		{"", false},
		{"KO", false},
		{"eng", false},
		{"korean", false},
		{"  ", false},
	}

	for _, tt := range tests {
		t.Run("code_"+tt.code, func(t *testing.T) {
			t.Parallel()
			got := IsValidLanguageCode(tt.code)
			if got != tt.want {
				t.Errorf("IsValidLanguageCode(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
