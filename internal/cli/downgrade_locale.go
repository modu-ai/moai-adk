package cli

import (
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/profile"
)

// resolveDowngradeLocale picks the language of the `moai update --version`
// downgrade confirmation (design.md §6, lead decision Q3): the project's
// language.yaml conversation_language when cwd is a MoAI project, else the
// active profile's conversation language, else English. A language.yaml that
// is missing, unreadable, or lacks the key counts as no project value.
func resolveDowngradeLocale(cwd string) string {
	if info, err := os.Stat(filepath.Join(cwd, ".moai")); err == nil && info.IsDir() {
		if v := wizard.ReadLocaleFromProject(cwd); v != "" {
			return v
		}
	}
	if prefs, err := profile.ReadPreferences(profile.GetCurrentName()); err == nil && prefs.ConversationLang != "" {
		return prefs.ConversationLang
	}
	return "en"
}
