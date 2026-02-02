package initializer

import (
	"testing"
)

// --- LanguageCode constants ---

func TestLanguageCodeConstants(t *testing.T) {
	tests := []struct {
		code LanguageCode
		want string
	}{
		{LanguageEnglish, "en"},
		{LanguageKorean, "ko"},
		{LanguageJapanese, "ja"},
		{LanguageChinese, "zh"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.code) != tt.want {
				t.Errorf("LanguageCode = %q, want %q", string(tt.code), tt.want)
			}
		})
	}
}

// --- SupportedLanguages ---

func TestSupportedLanguages(t *testing.T) {
	if len(SupportedLanguages) != 4 {
		t.Errorf("SupportedLanguages length = %d, want 4", len(SupportedLanguages))
	}

	// Verify all languages have required fields
	for i, lang := range SupportedLanguages {
		if lang.Code == "" {
			t.Errorf("SupportedLanguages[%d].Code is empty", i)
		}
		if lang.Name == "" {
			t.Errorf("SupportedLanguages[%d].Name is empty", i)
		}
		if lang.NativeName == "" {
			t.Errorf("SupportedLanguages[%d].NativeName is empty", i)
		}
		if lang.DisplayName == "" {
			t.Errorf("SupportedLanguages[%d].DisplayName is empty", i)
		}
	}
}

func TestSupportedLanguages_Order(t *testing.T) {
	expectedCodes := []LanguageCode{LanguageEnglish, LanguageKorean, LanguageJapanese, LanguageChinese}

	for i, expected := range expectedCodes {
		if SupportedLanguages[i].Code != expected {
			t.Errorf("SupportedLanguages[%d].Code = %q, want %q", i, SupportedLanguages[i].Code, expected)
		}
	}
}

// --- LanguageInfo ---

func TestLanguageInfo_DisplayName(t *testing.T) {
	tests := []struct {
		code        LanguageCode
		displayName string
	}{
		{LanguageEnglish, "English (en)"},
		{LanguageKorean, "Korean (ko)"},
		{LanguageJapanese, "Japanese (ja)"},
		{LanguageChinese, "Chinese (zh)"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			for _, lang := range SupportedLanguages {
				if lang.Code == tt.code {
					if lang.DisplayName != tt.displayName {
						t.Errorf("DisplayName = %q, want %q", lang.DisplayName, tt.displayName)
					}
					return
				}
			}
			t.Errorf("language code %q not found in SupportedLanguages", tt.code)
		})
	}
}

// --- NewPrompter ---

func TestNewPrompter(t *testing.T) {
	p := NewPrompter()
	if p == nil {
		t.Fatal("NewPrompter returned nil")
	}
	if p.reader == nil {
		t.Error("reader is nil")
	}
}
