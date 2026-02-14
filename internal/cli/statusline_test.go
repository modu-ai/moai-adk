package cli

import (
	"bytes"
	"os"
	"path/filepath"
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
