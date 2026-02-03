package models

// LangNameMap maps ISO 639-1 language codes to full names with native script.
// This is the canonical source for language names across the application.
// Supported languages: Korean, English, Japanese, Chinese
var LangNameMap = map[string]string{
	"ko": "Korean (한국어)",
	"en": "English",
	"ja": "Japanese (日本語)",
	"zh": "Chinese (中文)",
}

// SupportedLanguages returns all supported language codes.
func SupportedLanguages() []string {
	return []string{"ko", "en", "ja", "zh"}
}

// GetLanguageName returns the full language name for a code.
// Returns "English" if the code is not found.
func GetLanguageName(code string) string {
	if name, ok := LangNameMap[code]; ok {
		return name
	}
	return "English"
}

// IsValidLanguageCode checks if the given code is a supported language.
func IsValidLanguageCode(code string) bool {
	_, ok := LangNameMap[code]
	return ok
}
