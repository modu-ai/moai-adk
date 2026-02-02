package initializer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- NewSettingsGenerator ---

func TestNewSettingsGenerator(t *testing.T) {
	sg, err := NewSettingsGenerator()
	if err != nil {
		t.Fatalf("NewSettingsGenerator() error = %v", err)
	}
	if sg == nil {
		t.Fatal("NewSettingsGenerator returned nil")
	}
	if sg.binaryPath == "" {
		t.Error("binaryPath is empty")
	}
}

// --- GetBinaryPath ---

func TestGetBinaryPath(t *testing.T) {
	sg, err := NewSettingsGenerator()
	if err != nil {
		t.Fatalf("NewSettingsGenerator() error = %v", err)
	}

	path := sg.GetBinaryPath()
	if path == "" {
		t.Error("GetBinaryPath() returned empty string")
	}

	// Binary path should be absolute
	if !filepath.IsAbs(path) {
		t.Errorf("GetBinaryPath() = %q, expected absolute path", path)
	}
}

// --- Generate ---

func TestGenerate(t *testing.T) {
	sg, err := NewSettingsGenerator()
	if err != nil {
		t.Fatalf("NewSettingsGenerator() error = %v", err)
	}

	settings, err := sg.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if settings == nil {
		t.Fatal("Generate() returned nil")
	}

	// Check hooks are present
	if settings.Hooks == nil {
		t.Fatal("Hooks is nil")
	}

	// Verify all required hooks are present
	requiredHooks := []string{"SessionStart", "PreToolUse", "PostToolUse"}
	for _, hookName := range requiredHooks {
		hook, exists := settings.Hooks[hookName]
		if !exists {
			t.Errorf("missing hook %q", hookName)
			continue
		}
		if hook.Type != "command" {
			t.Errorf("hook %q type = %q, want %q", hookName, hook.Type, "command")
		}
		if hook.Command == "" {
			t.Errorf("hook %q command is empty", hookName)
		}
		// Verify command contains binary path
		if !strings.Contains(hook.Command, sg.binaryPath) {
			t.Errorf("hook %q command does not contain binary path", hookName)
		}
		// Verify command contains $CLAUDE_PROJECT_DIR
		if !strings.Contains(hook.Command, "$CLAUDE_PROJECT_DIR") {
			t.Errorf("hook %q command does not contain $CLAUDE_PROJECT_DIR", hookName)
		}
	}
}

func TestGenerate_HookCommandFormat(t *testing.T) {
	sg, err := NewSettingsGenerator()
	if err != nil {
		t.Fatalf("NewSettingsGenerator() error = %v", err)
	}

	settings, err := sg.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check that hook commands use kebab-case event names
	expectedSubCommands := map[string]string{
		"SessionStart": "session-start",
		"PreToolUse":   "pre-tool-use",
		"PostToolUse":  "post-tool-use",
	}

	for hookName, expectedCmd := range expectedSubCommands {
		hook := settings.Hooks[hookName]
		if !strings.Contains(hook.Command, expectedCmd) {
			t.Errorf("hook %q command %q does not contain subcommand %q", hookName, hook.Command, expectedCmd)
		}
	}
}

// --- WriteToFile ---

func TestWriteToFile(t *testing.T) {
	sg, err := NewSettingsGenerator()
	if err != nil {
		t.Fatalf("NewSettingsGenerator() error = %v", err)
	}

	tmpDir := t.TempDir()
	if err := sg.WriteToFile(tmpDir); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	// Verify file was created
	settingsFile := filepath.Join(tmpDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v", err)
	}

	// Verify hooks key exists
	if _, exists := parsed["hooks"]; !exists {
		t.Error("settings.json missing 'hooks' key")
	}
}

func TestWriteToFile_CreatesClaudeDirectory(t *testing.T) {
	sg, err := NewSettingsGenerator()
	if err != nil {
		t.Fatalf("NewSettingsGenerator() error = %v", err)
	}

	tmpDir := t.TempDir()
	if err := sg.WriteToFile(tmpDir); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	// Verify .claude directory was created
	claudeDir := filepath.Join(tmpDir, ".claude")
	info, err := os.Stat(claudeDir)
	if err != nil {
		t.Fatalf(".claude directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error(".claude path is not a directory")
	}
}

func TestWriteToFile_JSONIsIndented(t *testing.T) {
	sg, err := NewSettingsGenerator()
	if err != nil {
		t.Fatalf("NewSettingsGenerator() error = %v", err)
	}

	tmpDir := t.TempDir()
	if err := sg.WriteToFile(tmpDir); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	settingsFile := filepath.Join(tmpDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	// Indented JSON should contain newlines and spaces
	content := string(data)
	if !strings.Contains(content, "\n") {
		t.Error("settings.json is not indented (no newlines)")
	}
	if !strings.Contains(content, "  ") {
		t.Error("settings.json is not indented (no spaces)")
	}
}

// --- SettingsJSON struct ---

func TestSettingsJSONStruct(t *testing.T) {
	s := SettingsJSON{
		Hooks: map[string]HookConfig{
			"SessionStart": {Type: "command", Command: "echo hello"},
		},
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var parsed SettingsJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	hook, exists := parsed.Hooks["SessionStart"]
	if !exists {
		t.Fatal("SessionStart hook not found after roundtrip")
	}
	if hook.Type != "command" {
		t.Errorf("hook Type = %q, want %q", hook.Type, "command")
	}
	if hook.Command != "echo hello" {
		t.Errorf("hook Command = %q, want %q", hook.Command, "echo hello")
	}
}

func TestHookConfigStruct(t *testing.T) {
	hc := HookConfig{Type: "command", Command: "/usr/bin/moai hook session-start"}
	if hc.Type != "command" {
		t.Errorf("Type = %q", hc.Type)
	}
	if hc.Command == "" {
		t.Error("Command is empty")
	}
}

func TestWriteToFile_Overwrite(t *testing.T) {
	sg, err := NewSettingsGenerator()
	if err != nil {
		t.Fatalf("NewSettingsGenerator() error = %v", err)
	}

	tmpDir := t.TempDir()

	// Write once
	if err := sg.WriteToFile(tmpDir); err != nil {
		t.Fatalf("first WriteToFile() error = %v", err)
	}

	// Write again (should overwrite without error)
	if err := sg.WriteToFile(tmpDir); err != nil {
		t.Fatalf("second WriteToFile() error = %v", err)
	}

	// Verify file still valid
	settingsFile := filepath.Join(tmpDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v", err)
	}
}

func TestGenerate_NoShellWrapper(t *testing.T) {
	sg, err := NewSettingsGenerator()
	if err != nil {
		t.Fatalf("NewSettingsGenerator() error = %v", err)
	}

	settings, err := sg.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify NO shell wrapper (direct binary execution)
	for hookName, hook := range settings.Hooks {
		if strings.Contains(hook.Command, "bash -l -c") {
			t.Errorf("hook %q should not contain shell wrapper, got %q", hookName, hook.Command)
		}
	}
}
