package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProfilePreferences holds per-profile user preferences.
// Stored at ~/.moai/claude-profiles/<name>/preferences.yaml.
// These settings apply across all projects when using this profile.
type ProfilePreferences struct {
	// Identity
	UserName string `yaml:"user_name,omitempty"`

	// Language settings
	ConversationLang string `yaml:"conversation_lang,omitempty"` // "en", "ko", "ja", "zh"
	GitCommitLang    string `yaml:"git_commit_lang,omitempty"`
	CodeCommentLang  string `yaml:"code_comment_lang,omitempty"`
	DocLang          string `yaml:"doc_lang,omitempty"`

	// LLM settings
	ModelPolicy string `yaml:"model_policy,omitempty"` // "high", "medium", "low"
	Model       string `yaml:"model,omitempty"`        // e.g. "claude-opus-4-6"

	// Launch settings
	// PermissionMode sets the Claude Code permission mode at launch.
	// Valid values: "", "default", "acceptEdits", "plan", "auto", "bypassPermissions", "dontAsk".
	// Empty string means use the project settings.json default (typically "acceptEdits").
	PermissionMode string `yaml:"permission_mode,omitempty"`

	// Bypass is deprecated. Use PermissionMode = "bypassPermissions" instead.
	// Kept for backward compatibility: ReadPreferences migrates Bypass=true
	// to PermissionMode="bypassPermissions" automatically.
	Bypass bool `yaml:"bypass,omitempty"`

	// Display settings
	StatuslineMode     string          `yaml:"statusline_mode,omitempty"`     // "default", "full"
	StatuslinePreset   string          `yaml:"statusline_preset,omitempty"`   // "full", "compact", "minimal", "custom"
	StatuslineSegments map[string]bool `yaml:"statusline_segments,omitempty"` // segment toggles for custom preset
	StatuslineTheme    string          `yaml:"statusline_theme,omitempty"`    // "default", "catppuccin-mocha", "catppuccin-latte"
}

const (
	preferencesFile       = "preferences.yaml"
	legacyPreferencesFile = ".preferences.yaml"
)

// ValidPermissionModes lists all Claude Code permission mode values.
// @MX:NOTE: [AUTO] Claude Code 권한 모드 전체 목록. CLI 입력 검증 및 프로필 위자드에서 참조.
var ValidPermissionModes = []string{
	"",                  // use project default
	"default",           // ask permissions for everything
	"acceptEdits",       // auto-accept file edits, ask for commands
	"plan",              // read-only exploration and planning
	"auto",              // background classifier checks actions
	"bypassPermissions", // skip all checks (isolated envs only)
	"dontAsk",           // only pre-approved tools
}

// IsValidPermissionMode checks whether mode is a recognized Claude Code permission mode.
// @MX:NOTE: [AUTO] 권한 모드 문자열 검증. CLI, GLM 런처 등에서 사용자 입력 유효성 검사에 사용.
func IsValidPermissionMode(mode string) bool {
	for _, m := range ValidPermissionModes {
		if m == mode {
			return true
		}
	}
	return false
}

// GetPreferencesPath returns the path to preferences.yaml for a profile.
func GetPreferencesPath(profileName string) string {
	baseDir := GetBaseDir()
	if profileName == "" || profileName == "default" {
		return filepath.Join(baseDir, preferencesFile)
	}
	return filepath.Join(baseDir, profileName, preferencesFile)
}

// IsSetup checks if a profile has been configured by looking for
// preferences.yaml (or legacy .preferences.yaml).
func IsSetup(profileName string) bool {
	path := GetPreferencesPath(profileName)
	if _, err := os.Stat(path); err == nil {
		return true
	}
	// Check legacy dot-prefixed file
	dir := filepath.Dir(path)
	legacyPath := filepath.Join(dir, legacyPreferencesFile)
	if _, err := os.Stat(legacyPath); err == nil {
		return true
	}
	return false
}

// ReadPreferences reads the preferences for a profile.
// Returns a zero-value ProfilePreferences if the file does not exist.
// Automatically migrates legacy .preferences.yaml to preferences.yaml.
func ReadPreferences(profileName string) (ProfilePreferences, error) {
	path := GetPreferencesPath(profileName)
	migrateOldFile(filepath.Dir(path), legacyPreferencesFile, preferencesFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ProfilePreferences{}, nil
		}
		return ProfilePreferences{}, fmt.Errorf("read preferences: %w", err)
	}
	var prefs ProfilePreferences
	if err := yaml.Unmarshal(data, &prefs); err != nil {
		return ProfilePreferences{}, fmt.Errorf("parse preferences: %w", err)
	}

	// Migrate legacy Bypass field to PermissionMode.
	// If Bypass is true and PermissionMode is not yet set, promote it.
	if prefs.Bypass && prefs.PermissionMode == "" {
		prefs.PermissionMode = "bypassPermissions"
	}

	return prefs, nil
}

// migrateOldFile renames a legacy dot-prefixed file to its new name
// if the old file exists and the new file does not.
func migrateOldFile(dir, oldName, newName string) {
	oldPath := filepath.Join(dir, oldName)
	newPath := filepath.Join(dir, newName)

	// Only migrate if old file exists and new file does not
	if _, err := os.Stat(oldPath); err != nil {
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		return // new file already exists, skip
	}
	_ = os.Rename(oldPath, newPath)
}

// WritePreferences saves the preferences for a profile.
// Creates the profile directory if it does not exist.
func WritePreferences(profileName string, prefs ProfilePreferences) error {
	path := GetPreferencesPath(profileName)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	data, err := yaml.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("marshal preferences: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write preferences: %w", err)
	}
	return nil
}
