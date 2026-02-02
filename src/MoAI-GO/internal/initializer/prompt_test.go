package initializer

import (
	"bufio"
	"strings"
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

// helper to create a Prompter with simulated input
func newTestPrompter(input string) *Prompter {
	return &Prompter{
		reader: bufio.NewReader(strings.NewReader(input)),
	}
}

// --- PromptLanguage ---

func TestPromptLanguage_DefaultSelection(t *testing.T) {
	p := newTestPrompter("\n") // empty input = default (English)

	code, err := p.PromptLanguage()
	if err != nil {
		t.Fatalf("PromptLanguage() error = %v", err)
	}
	if code != LanguageEnglish {
		t.Errorf("code = %q, want %q (default)", string(code), string(LanguageEnglish))
	}
}

func TestPromptLanguage_SelectKorean(t *testing.T) {
	p := newTestPrompter("2\n") // Korean is option 2

	code, err := p.PromptLanguage()
	if err != nil {
		t.Fatalf("PromptLanguage() error = %v", err)
	}
	if code != LanguageKorean {
		t.Errorf("code = %q, want %q", string(code), string(LanguageKorean))
	}
}

func TestPromptLanguage_SelectJapanese(t *testing.T) {
	p := newTestPrompter("3\n")

	code, err := p.PromptLanguage()
	if err != nil {
		t.Fatalf("PromptLanguage() error = %v", err)
	}
	if code != LanguageJapanese {
		t.Errorf("code = %q, want %q", string(code), string(LanguageJapanese))
	}
}

func TestPromptLanguage_SelectChinese(t *testing.T) {
	p := newTestPrompter("4\n")

	code, err := p.PromptLanguage()
	if err != nil {
		t.Fatalf("PromptLanguage() error = %v", err)
	}
	if code != LanguageChinese {
		t.Errorf("code = %q, want %q", string(code), string(LanguageChinese))
	}
}

func TestPromptLanguage_InvalidNumber(t *testing.T) {
	p := newTestPrompter("9\n")

	_, err := p.PromptLanguage()
	if err == nil {
		t.Error("expected error for invalid selection number")
	}
}

func TestPromptLanguage_InvalidInput(t *testing.T) {
	p := newTestPrompter("abc\n")

	_, err := p.PromptLanguage()
	if err == nil {
		t.Error("expected error for non-numeric input")
	}
}

func TestPromptLanguage_ZeroSelection(t *testing.T) {
	p := newTestPrompter("0\n")

	_, err := p.PromptLanguage()
	if err == nil {
		t.Error("expected error for zero selection")
	}
}

// --- PromptUserName ---

func TestPromptUserName_WithName(t *testing.T) {
	p := newTestPrompter("John Doe\n")

	name, err := p.PromptUserName()
	if err != nil {
		t.Fatalf("PromptUserName() error = %v", err)
	}
	if name != "John Doe" {
		t.Errorf("name = %q, want %q", name, "John Doe")
	}
}

func TestPromptUserName_EmptyDefault(t *testing.T) {
	p := newTestPrompter("\n") // empty = default "Developer"

	name, err := p.PromptUserName()
	if err != nil {
		t.Fatalf("PromptUserName() error = %v", err)
	}
	if name != "Developer" {
		t.Errorf("name = %q, want %q (default)", name, "Developer")
	}
}

func TestPromptUserName_WhitespaceOnly(t *testing.T) {
	p := newTestPrompter("   \n") // whitespace only = trimmed to empty = default

	name, err := p.PromptUserName()
	if err != nil {
		t.Fatalf("PromptUserName() error = %v", err)
	}
	if name != "Developer" {
		t.Errorf("name = %q, want %q (default)", name, "Developer")
	}
}

// --- PromptConfirm ---

func TestPromptConfirm_YesInput(t *testing.T) {
	p := newTestPrompter("y\n")

	result, err := p.PromptConfirm("Continue?", false)
	if err != nil {
		t.Fatalf("PromptConfirm() error = %v", err)
	}
	if !result {
		t.Error("expected true for 'y' input")
	}
}

func TestPromptConfirm_YesFullWord(t *testing.T) {
	p := newTestPrompter("yes\n")

	result, err := p.PromptConfirm("Continue?", false)
	if err != nil {
		t.Fatalf("PromptConfirm() error = %v", err)
	}
	if !result {
		t.Error("expected true for 'yes' input")
	}
}

func TestPromptConfirm_NoInput(t *testing.T) {
	p := newTestPrompter("n\n")

	result, err := p.PromptConfirm("Continue?", true)
	if err != nil {
		t.Fatalf("PromptConfirm() error = %v", err)
	}
	if result {
		t.Error("expected false for 'n' input")
	}
}

func TestPromptConfirm_DefaultTrue(t *testing.T) {
	p := newTestPrompter("\n") // empty = use default

	result, err := p.PromptConfirm("Continue?", true)
	if err != nil {
		t.Fatalf("PromptConfirm() error = %v", err)
	}
	if !result {
		t.Error("expected true for empty input with default=true")
	}
}

func TestPromptConfirm_DefaultFalse(t *testing.T) {
	p := newTestPrompter("\n")

	result, err := p.PromptConfirm("Continue?", false)
	if err != nil {
		t.Fatalf("PromptConfirm() error = %v", err)
	}
	if result {
		t.Error("expected false for empty input with default=false")
	}
}

func TestPromptConfirm_CaseInsensitive(t *testing.T) {
	p := newTestPrompter("Y\n")

	result, err := p.PromptConfirm("Continue?", false)
	if err != nil {
		t.Fatalf("PromptConfirm() error = %v", err)
	}
	if !result {
		t.Error("expected true for 'Y' input (case insensitive)")
	}
}
