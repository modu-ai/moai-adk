package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LaunchConfig holds per-profile claude launch preferences.
// Stored at ~/.moai/claude-profiles/<name>/.launch.yaml.
// These settings act as defaults and can be overridden by project-level
// DO_CLAUDE_* env vars in .claude/settings.local.json or command-line flags.
type LaunchConfig struct {
	// Model overrides the default model (e.g., "claude-opus-4-6").
	// Empty string means use the system default.
	Model string `yaml:"model,omitempty"`
	// Bypass enables --dangerously-skip-permissions when true.
	Bypass bool `yaml:"bypass,omitempty"`
	// Continue enables --continue (resume previous session) when true.
	Continue bool `yaml:"continue,omitempty"`
	// Chrome controls Chrome MCP. nil means use the system default.
	// true = enable Chrome MCP, false = explicitly disable (--no-chrome).
	Chrome *bool `yaml:"chrome,omitempty"`
}

const launchConfigFile = ".launch.yaml"

// GetLaunchConfigPath returns the path to .launch.yaml for a profile.
func GetLaunchConfigPath(profileName string) string {
	baseDir := GetBaseDir()
	if profileName == "" || profileName == "default" {
		return filepath.Join(baseDir, launchConfigFile)
	}
	return filepath.Join(baseDir, profileName, launchConfigFile)
}

// ReadLaunchConfig reads the launch config for a profile.
// Returns a zero-value LaunchConfig if the file does not exist.
func ReadLaunchConfig(profileName string) (LaunchConfig, error) {
	path := GetLaunchConfigPath(profileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return LaunchConfig{}, nil
		}
		return LaunchConfig{}, fmt.Errorf("read launch config: %w", err)
	}
	var cfg LaunchConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return LaunchConfig{}, fmt.Errorf("parse launch config: %w", err)
	}
	return cfg, nil
}

// WriteLaunchConfig saves the launch config for a profile.
// Creates the profile directory if it does not exist.
func WriteLaunchConfig(profileName string, cfg LaunchConfig) error {
	path := GetLaunchConfigPath(profileName)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal launch config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write launch config: %w", err)
	}
	return nil
}
