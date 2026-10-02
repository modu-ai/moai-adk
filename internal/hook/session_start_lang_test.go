package hook

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// configWithLang returns a ConfigProvider whose conversation_language is lang.
//
// Moved verbatim from session_start_kanban_i18n_test.go (SPEC-LAUNCHER-ENTRY-FLAGS-001
// M5b, deleted with the kanban notice): retained tests call it.
func configWithLang(lang string) ConfigProvider {
	return &auditConfigProvider{cfg: &config.Config{
		Language: models.LanguageConfig{ConversationLanguage: lang},
	}}
}

// TestOperatorLangFailsOpen asserts locale resolution degrades to English
// rather than failing the session start, matching the surrounding hook code.
// Moved with configWithLang: operatorLang is retained (session_start_lang.go).
func TestOperatorLangFailsOpen(t *testing.T) {
	cases := map[string]ConfigProvider{
		"nil provider": nil,
		"nil config":   &auditConfigProvider{cfg: nil},
		"empty lang":   configWithLang(""),
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			if got := operatorLang(cfg); got != langEnglish {
				t.Errorf("operatorLang = %q, want %q", got, langEnglish)
			}
		})
	}
	if got := operatorLang(configWithLang("ja")); got != "ja" {
		t.Errorf("operatorLang = %q, want %q", got, "ja")
	}
}
