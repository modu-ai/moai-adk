package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
	"gopkg.in/yaml.v3"
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

	// Sync statusline section (written directly to avoid config manager dependency)
	if prefs.StatuslinePreset != "" || prefs.StatuslineTheme != "" || prefs.StatuslineSegments != nil || prefs.StatuslineMode != "" {
		if err := syncStatusline(projectRoot, prefs); err != nil {
			return fmt.Errorf("sync statusline: %w", err)
		}
	}

	return nil
}

// statuslineData is the internal YAML structure for statusline.yaml.
type statuslineData struct {
	Mode     string          `yaml:"mode,omitempty"`
	Preset   string          `yaml:"preset,omitempty"`
	Segments map[string]bool `yaml:"segments,omitempty"`
	Theme    string          `yaml:"theme,omitempty"`
}

// statuslineFileWrap wraps statuslineData under the "statusline" top-level key.
type statuslineFileWrap struct {
	Statusline statuslineData `yaml:"statusline"`
}

// syncStatusline writes StatuslinePreset, StatuslineSegments, and StatuslineTheme
// to .moai/config/sections/statusline.yaml. When the file is absent, all segments
// default to enabled and preset defaults to "full" (REQ-SLE-022).
func syncStatusline(projectRoot string, prefs ProfilePreferences) error {
	sectionsDir := filepath.Join(projectRoot, ".moai", "config", "sections")
	statuslineFile := filepath.Join(sectionsDir, "statusline.yaml")

	// Read current statusline.yaml if it exists
	var current statuslineFileWrap
	data, err := os.ReadFile(statuslineFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read statusline.yaml: %w", err)
	}
	if err == nil {
		if err := yaml.Unmarshal(data, &current); err != nil {
			return fmt.Errorf("parse statusline.yaml: %w", err)
		}
	}

	// Apply defaults when statusline.yaml was absent (REQ-SLE-022)
	if current.Statusline.Preset == "" {
		current.Statusline.Preset = "full"
	}
	if current.Statusline.Segments == nil {
		current.Statusline.Segments = defaultStatuslineSegments()
	}

	// Merge preferences (non-empty values override existing config)
	if prefs.StatuslineMode != "" {
		current.Statusline.Mode = prefs.StatuslineMode
	}
	if prefs.StatuslinePreset != "" {
		current.Statusline.Preset = prefs.StatuslinePreset
	}
	if prefs.StatuslineTheme != "" {
		current.Statusline.Theme = prefs.StatuslineTheme
	}
	if prefs.StatuslineSegments != nil {
		current.Statusline.Segments = prefs.StatuslineSegments
	}

	// Write statusline.yaml
	yamlData, err := yaml.Marshal(current)
	if err != nil {
		return fmt.Errorf("marshal statusline.yaml: %w", err)
	}
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(statuslineFile, yamlData, 0o644); err != nil {
		return fmt.Errorf("write statusline.yaml: %w", err)
	}
	return nil
}

// defaultStatuslineSegments returns the default segment map with all segments enabled.
func defaultStatuslineSegments() map[string]bool {
	return map[string]bool{
		"model":          true,
		"context":        true,
		"output_style":   true,
		"directory":      true,
		"git_status":     true,
		"claude_version": true,
		"moai_version":   true,
		"git_branch":     true,
	}
}
