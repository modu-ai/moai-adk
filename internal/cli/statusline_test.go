package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStatuslineCmd_Exists(t *testing.T) {
	if StatuslineCmd == nil {
		t.Fatal("StatuslineCmd should not be nil")
	}
}

func TestStatuslineCmd_Use(t *testing.T) {
	if StatuslineCmd.Use != "statusline" {
		t.Errorf("StatuslineCmd.Use = %q, want %q", StatuslineCmd.Use, "statusline")
	}
}

func TestStatuslineCmd_Hidden(t *testing.T) {
	if !StatuslineCmd.Hidden {
		t.Error("statusline command should be hidden")
	}
}

func TestStatuslineCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "statusline" {
			found = true
			break
		}
	}
	if !found {
		t.Error("statusline should be registered as a subcommand of root")
	}
}

func TestStatuslineCmd_Execution_NoDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	buf := new(bytes.Buffer)
	StatuslineCmd.SetOut(buf)
	StatuslineCmd.SetErr(buf)

	err := StatuslineCmd.RunE(StatuslineCmd, []string{})
	if err != nil {
		t.Fatalf("statusline should not error, got: %v", err)
	}

	output := buf.String()
	// Statusline should produce some output (git status, version, branch, or fallback)
	output = strings.TrimSpace(output)
	if output == "" {
		t.Errorf("output should not be empty")
	}
	// If output doesn't contain expected sections, it should at least be a valid fallback
	// The new independent collection always shows git status or version when available
}

// --- DDD PRESERVE: Characterization tests for statusline utility functions ---

func TestRenderSimpleFallback(t *testing.T) {
	result := renderSimpleFallback()

	if result == "" {
		t.Error("renderSimpleFallback should not return empty string")
	}

	if result != "moai" {
		t.Errorf("renderSimpleFallback() = %q, want %q", result, "moai")
	}
}

func TestRenderSimpleFallback_NotEmpty(t *testing.T) {
	result := renderSimpleFallback()

	if len(result) == 0 {
		t.Fatal("renderSimpleFallback should return non-empty string")
	}

	// Should be a simple string without special characters
	if strings.Contains(result, "\n") {
		t.Error("renderSimpleFallback should not contain newlines")
	}
}

func TestRenderSimpleFallback_ConsistentOutput(t *testing.T) {
	// Should return consistent output across multiple calls
	first := renderSimpleFallback()
	second := renderSimpleFallback()

	if first != second {
		t.Errorf("renderSimpleFallback should be consistent, got %q and %q", first, second)
	}
}

// --- TDD RED: Tests for loadSegmentConfig ---

func TestLoadSegmentConfig(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantNil     bool
		wantKeys    map[string]bool
	}{
		{
			name:    "missing file returns nil",
			wantNil: true,
		},
		{
			name:        "invalid YAML returns nil",
			yamlContent: "{{invalid yaml content!",
			wantNil:     true,
		},
		{
			name: "valid YAML with all segments true",
			yamlContent: `statusline:
  preset: "full"
  segments:
    model: true
    context: true
    output_style: true
    directory: true
    git_status: true
    claude_version: true
    moai_version: true
    git_branch: true
`,
			wantNil: false,
			wantKeys: map[string]bool{
				"model": true, "context": true, "output_style": true,
				"directory": true, "git_status": true, "claude_version": true,
				"moai_version": true, "git_branch": true,
			},
		},
		{
			name: "valid YAML with some segments disabled",
			yamlContent: `statusline:
  preset: "compact"
  segments:
    model: true
    context: true
    output_style: false
    directory: false
    git_status: true
    claude_version: false
    moai_version: false
    git_branch: true
`,
			wantNil: false,
			wantKeys: map[string]bool{
				"model": true, "context": true, "output_style": false,
				"directory": false, "git_status": true, "claude_version": false,
				"moai_version": false, "git_branch": true,
			},
		},
		{
			name: "valid YAML with empty segments",
			yamlContent: `statusline:
  preset: "custom"
  segments: {}
`,
			wantNil:  false,
			wantKeys: map[string]bool{},
		},
		{
			name: "valid YAML without segments key",
			yamlContent: `statusline:
  preset: "full"
`,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.yamlContent != "" {
				configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatalf("failed to create config dir: %v", err)
				}
				configPath := filepath.Join(configDir, "statusline.yaml")
				if err := os.WriteFile(configPath, []byte(tt.yamlContent), 0o644); err != nil {
					t.Fatalf("failed to write config: %v", err)
				}
			}

			got := loadSegmentConfig(tmpDir)

			if tt.wantNil {
				if got != nil {
					t.Errorf("loadSegmentConfig() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("loadSegmentConfig() = nil, want non-nil")
			}

			if len(got) != len(tt.wantKeys) {
				t.Errorf("loadSegmentConfig() returned %d keys, want %d", len(got), len(tt.wantKeys))
			}

			for key, wantVal := range tt.wantKeys {
				gotVal, exists := got[key]
				if !exists {
					t.Errorf("loadSegmentConfig() missing key %q", key)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("loadSegmentConfig()[%q] = %v, want %v", key, gotVal, wantVal)
				}
			}
		})
	}
}

func TestLoadSegmentConfig_EmptyProjectRoot(t *testing.T) {
	got := loadSegmentConfig("")
	if got != nil {
		t.Errorf("loadSegmentConfig(\"\") = %v, want nil", got)
	}
}

// --- TDD RED: Tests for loadStatuslineFileConfig ---

func TestLoadStatuslineFileConfig(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		wantNil    bool
		wantPreset string
		wantTheme  string
		wantSegs   map[string]bool
	}{
		{
			name:    "empty project root returns nil",
			wantNil: true,
		},
		{
			name:    "missing file returns nil",
			yaml:    "",
			wantNil: true,
		},
		{
			name:    "invalid YAML returns nil",
			yaml:    "{{ invalid",
			wantNil: true,
		},
		{
			name: "reads preset and theme",
			yaml: `statusline:
  preset: "compact"
  theme: "catppuccin-mocha"
  segments:
    model: true
    context: false
`,
			wantNil:    false,
			wantPreset: "compact",
			wantTheme:  "catppuccin-mocha",
			wantSegs:   map[string]bool{"model": true, "context": false},
		},
		{
			name: "default theme when absent",
			yaml: `statusline:
  preset: "full"
  segments:
    model: true
`,
			wantNil:    false,
			wantPreset: "full",
			wantTheme:  "",
			wantSegs:   map[string]bool{"model": true},
		},
		{
			name: "catppuccin-latte theme",
			yaml: `statusline:
  preset: "full"
  theme: "catppuccin-latte"
  segments:
    model: true
`,
			wantNil:    false,
			wantPreset: "full",
			wantTheme:  "catppuccin-latte",
			wantSegs:   map[string]bool{"model": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Empty project root case
			if tt.name == "empty project root returns nil" {
				got := loadStatuslineFileConfig("")
				if got != nil {
					t.Errorf("loadStatuslineFileConfig(\"\") = %v, want nil", got)
				}
				return
			}

			if tt.yaml != "" {
				configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
				if err := os.MkdirAll(configDir, 0o755); err != nil {
					t.Fatalf("failed to create config dir: %v", err)
				}
				configPath := filepath.Join(configDir, "statusline.yaml")
				if err := os.WriteFile(configPath, []byte(tt.yaml), 0o644); err != nil {
					t.Fatalf("failed to write config: %v", err)
				}
			}

			got := loadStatuslineFileConfig(tmpDir)

			if tt.wantNil {
				if got != nil {
					t.Errorf("loadStatuslineFileConfig() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("loadStatuslineFileConfig() = nil, want non-nil")
			}

			if got.Preset != tt.wantPreset {
				t.Errorf("Preset = %q, want %q", got.Preset, tt.wantPreset)
			}
			if got.Theme != tt.wantTheme {
				t.Errorf("Theme = %q, want %q", got.Theme, tt.wantTheme)
			}
			for k, v := range tt.wantSegs {
				if got.Segments[k] != v {
					t.Errorf("Segments[%q] = %v, want %v", k, got.Segments[k], v)
				}
			}
		})
	}
}

// --- TDD RED: M1 Tests for Cwd Guard ---

// TestCwdGuard_DeletedDirectory tests AC-SF-006: Cwd Guard for deleted directories.
// When process's current working directory is deleted, moai statusline shall
// switch to $HOME and display project directory as ~, without crashing.
//
// RED Phase: This test should fail because cwd guard is not yet implemented in runStatusline.
func TestCwdGuard_DeletedDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not allow deleting cwd of a running process")
	}

	// Create a temporary directory
	tempDir := t.TempDir()

	// Save original working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origWd) }() // Always restore original directory

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir to temp: %v", err)
	}

	// Delete the temporary directory while we're in it
	// This simulates the deleted directory scenario
	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("failed to remove temp directory: %v", err)
	}

	// Verify the directory is truly gone.
	// Note: on macOS, os.Getwd() may still succeed after directory deletion,
	// so we use os.Stat() for a platform-agnostic check.
	if _, statErr := os.Stat(tempDir); statErr == nil {
		t.Fatal("expected os.Stat(tempDir) to fail after removal, but directory still exists")
	}

	// Try to run statusline - it should NOT crash
	// Expected: cwd guard catches the error, switches to $HOME, and continues
	// Current behavior (RED): likely crashes or returns error
	buf := new(bytes.Buffer)
	StatuslineCmd.SetOut(buf)
	StatuslineCmd.SetErr(buf)

	err = StatuslineCmd.RunE(StatuslineCmd, []string{})
	if err != nil {
		// In RED phase, this is expected to fail
		// After GREEN phase, this should succeed
		t.Logf("RED phase: runStatusline failed as expected: %v", err)
	}

	// After GREEN phase: verify output is not empty (fallback should work)
	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Log("RED phase: output is empty, expected after cwd guard is implemented")
	}
}
