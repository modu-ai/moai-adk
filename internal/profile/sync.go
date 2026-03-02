package profile

import (
	"fmt"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// SyncToProjectConfig synchronizes profile preferences to
// the project's .moai/config/sections/ YAML files.
// Only non-empty preference values overwrite existing config values.
func SyncToProjectConfig(projectRoot string, prefs ProfilePreferences) error {
	mgr := config.NewConfigManager()
	cfg, err := mgr.LoadRaw(projectRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}

	changed := false

	// Sync user section
	if prefs.UserName != "" && cfg.User.Name != prefs.UserName {
		cfg.User = models.UserConfig{Name: prefs.UserName}
		if err := mgr.SetSection("user", cfg.User); err != nil {
			return fmt.Errorf("set user section: %w", err)
		}
		changed = true
	}

	// Sync language section
	lang := cfg.Language
	langChanged := false

	if prefs.ConversationLang != "" && lang.ConversationLanguage != prefs.ConversationLang {
		lang.ConversationLanguage = prefs.ConversationLang
		lang.ConversationLanguageName = prefs.ConversationLang
		langChanged = true
	}
	if prefs.GitCommitLang != "" && lang.GitCommitMessages != prefs.GitCommitLang {
		lang.GitCommitMessages = prefs.GitCommitLang
		langChanged = true
	}
	if prefs.CodeCommentLang != "" && lang.CodeComments != prefs.CodeCommentLang {
		lang.CodeComments = prefs.CodeCommentLang
		langChanged = true
	}
	if prefs.DocLang != "" && lang.Documentation != prefs.DocLang {
		lang.Documentation = prefs.DocLang
		langChanged = true
	}

	if langChanged {
		if err := mgr.SetSection("language", lang); err != nil {
			return fmt.Errorf("set language section: %w", err)
		}
		changed = true
	}

	if changed {
		if err := mgr.Save(); err != nil {
			return fmt.Errorf("save project config: %w", err)
		}
	}

	return nil
}
