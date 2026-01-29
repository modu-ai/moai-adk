package initializer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/moai-adk-go/internal/output"
)

// SettingsJSON represents the Claude Code settings.json structure
type SettingsJSON struct {
	Hooks map[string]HookConfig `json:"hooks"`
}

// HookConfig represents a single hook configuration
type HookConfig struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// SettingsGenerator generates settings.json with direct binary path
type SettingsGenerator struct {
	binaryPath string
}

// NewSettingsGenerator creates a new settings generator
func NewSettingsGenerator() (*SettingsGenerator, error) {
	// Get absolute path to current binary
	binaryPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("error getting binary path: %w", err)
	}

	return &SettingsGenerator{
		binaryPath: binaryPath,
	}, nil
}

// Generate generates the settings.json content
func (sg *SettingsGenerator) Generate() (*SettingsJSON, error) {
	hooks := make(map[string]HookConfig)

	// Define all hooks with direct binary path (NO shell wrapper)
	hookCommands := map[string]string{
		"SessionStart": "session-start",
		"PreToolUse":   "pre-tool-use",
		"PostToolUse":  "post-tool-use",
	}

	for hookName, hookCmd := range hookCommands {
		hooks[hookName] = HookConfig{
			Type:    "command",
			Command: fmt.Sprintf("%s hook %s --project-dir \"$CLAUDE_PROJECT_DIR\"", sg.binaryPath, hookCmd),
		}
	}

	return &SettingsJSON{
		Hooks: hooks,
	}, nil
}

// WriteToFile writes settings.json to the target directory
func (sg *SettingsGenerator) WriteToFile(targetDir string) error {
	settings, err := sg.Generate()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	settingsPath := filepath.Join(targetDir, ".claude")
	if err := os.MkdirAll(settingsPath, 0755); err != nil {
		return fmt.Errorf("error creating .claude directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling settings.json: %w", err)
	}

	// Write to file
	settingsFile := filepath.Join(settingsPath, "settings.json")
	if err := os.WriteFile(settingsFile, data, 0644); err != nil {
		return fmt.Errorf("error writing settings.json: %w", err)
	}

	fmt.Println(output.SuccessStyle.Render(fmt.Sprintf("✓ Generated: %s", settingsFile)))
	return nil
}

// GetBinaryPath returns the detected binary path
func (sg *SettingsGenerator) GetBinaryPath() string {
	return sg.binaryPath
}
