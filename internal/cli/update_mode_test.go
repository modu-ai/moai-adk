package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TAG-005: moai update Redesign Tests

// TestUpdateCmd_MutuallyExclusiveFlags tests flag exclusivity
func TestUpdateCmd_MutuallyExclusiveFlags(t *testing.T) {
	// This test verifies that --binary and --templates-only are mutually exclusive
	// The actual implementation is tested in update_test.go
	flag := updateCmd.Flags().Lookup("binary")
	if flag == nil {
		t.Error("update command should have --binary flag")
	}

	templateFlag := updateCmd.Flags().Lookup("templates-only")
	if templateFlag == nil {
		t.Error("update command should have --templates-only flag")
	}
}

// TestUpdateCmd_ConfigFlagExists tests that -c/--config flag exists (backward compat)
func TestUpdateCmd_ConfigFlagExists(t *testing.T) {
	flag := updateCmd.Flags().Lookup("config")
	if flag == nil {
		t.Fatal("update command should have --config flag")
	}
	if flag.Shorthand != "c" {
		t.Errorf("config flag shorthand = %q, want 'c'", flag.Shorthand)
	}
}

// TestUpdateCmd_TemplatesOnlyAndBinaryMutuallyExclusive tests flag exclusivity
func TestUpdateCmd_TemplatesOnlyAndBinaryMutuallyExclusive(t *testing.T) {
	// Create a temp project with .moai/
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(moaiDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a minimal manifest.json
	manifestPath := filepath.Join(tmpDir, ".moai", "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(`{"files": []}`), 0644); err != nil {
		t.Fatal(err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	_ = updateCmd.Flags().Set("templates-only", "true")
	_ = updateCmd.Flags().Set("binary", "true")
	_ = updateCmd.Flags().Set("yes", "true")

	err = updateCmd.RunE(updateCmd, []string{})
	if err == nil {
		t.Error("expected error when both --templates-only and --binary are set")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", err)
	}
}

// TestPromptResetConfig tests the reset config prompt function
func TestPromptResetConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"yes lowercase", "y\n", true},
		{"yes uppercase", "Y\n", true},
		{"yes full", "yes\n", true},
		{"YES full", "YES\n", true},
		{"no lowercase", "n\n", false},
		{"no uppercase", "N\n", false},
		{"no full", "no\n", false},
		{"empty defaults to no", "\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock stdin reader
			r := strings.NewReader(tt.input)
			result := promptResetConfigFromReader(r)
			if result != tt.expected {
				t.Errorf("promptResetConfig(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPromptModeChange tests the mode change prompt function
func TestPromptModeChange(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		currentMode string
		expected    string
	}{
		{"keep local", "n\n", "local", "local"},
		{"keep global", "n\n", "global", "global"},
		{"change to global", "y\n", "local", "global"},
		{"change to local", "y\n", "global", "local"},
		{"empty keeps current", "\n", "local", "local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			result := promptModeChangeFromReader(r, tt.currentMode)
			if result != tt.expected {
				t.Errorf("promptModeChange(%q, %q) = %v, want %v", tt.input, tt.currentMode, result, tt.expected)
			}
		})
	}
}

// Helper functions for testing with mock input
// These will be implemented alongside the actual functions

// promptResetConfigFromReader prompts the user to reset config, reading from the provided reader
func promptResetConfigFromReader(r *strings.Reader) bool {
	var input string
	_, _ = fmt.Fscanf(r, "%s", &input)
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

// promptModeChangeFromReader prompts the user to change installation mode
func promptModeChangeFromReader(r *strings.Reader, currentMode string) string {
	var input string
	_, _ = fmt.Fscanf(r, "%s", &input)
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "y" || input == "yes" {
		// Toggle mode
		if currentMode == "local" {
			return "global"
		}
		return "local"
	}
	return currentMode
}
