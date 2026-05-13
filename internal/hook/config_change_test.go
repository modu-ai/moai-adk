package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigChangeHandler_EventType(t *testing.T) {
	h := NewConfigChangeHandler()
	if h.EventType() != EventConfigChange {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventConfigChange)
	}
}

func TestConfigChangeHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *HookInput
		createFile    bool
		fileContent   string
		expectMessage bool
		expectBlock   bool
	}{
		{
			name: "valid YAML config",
			input: &HookInput{
				SessionID:      "sess-001",
				ConfigFilePath: "quality.yaml",
				HookEventName:  "ConfigChange",
			},
			createFile:    true,
			fileContent:   "development_mode: ddd\ncoverage_target: 85\n",
			expectMessage: true,
			expectBlock:   false,
		},
		{
			name: "invalid YAML config",
			input: &HookInput{
				SessionID:      "sess-002",
				ConfigFilePath: "invalid.yaml",
				HookEventName:  "ConfigChange",
			},
			createFile:    true,
			fileContent:   "invalid: yaml: content:\n  - broken\n",
			expectMessage: true,
			expectBlock:   true,
		},
		{
			name: "empty file",
			input: &HookInput{
				SessionID:      "sess-003",
				ConfigFilePath: "empty.yaml",
				HookEventName:  "ConfigChange",
			},
			createFile:    true,
			fileContent:   "",
			expectMessage: true,
			expectBlock:   false,
		},
		{
			name: "non-existent file",
			input: &HookInput{
				SessionID:      "sess-004",
				ConfigFilePath: "/does/not/exist.yaml",
				HookEventName:  "ConfigChange",
			},
			createFile:    false,
			expectMessage: true,
			expectBlock:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewConfigChangeHandler()

			// Create temp file if needed
			if tt.createFile {
				tempDir := t.TempDir()
				filePath := tt.input.ConfigFilePath
				if !filepath.IsAbs(filePath) {
					filePath = filepath.Join(tempDir, filepath.Base(filePath))
				}

				if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				tt.input.ConfigFilePath = filePath
			}

			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}

			if tt.expectMessage && out.SystemMessage == "" {
				t.Error("expected SystemMessage")
			}
			if tt.expectBlock && out.Continue {
				t.Error("expected Continue=false for invalid config")
			}
			if !tt.expectBlock && !out.Continue {
				t.Error("expected Continue=true for valid config")
			}
		})
	}
}

func TestConfigChangeHandler_ValidateConfig(t *testing.T) {
	t.Parallel()

	h := &configChangeHandler{}

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "valid YAML",
			content:     "key: value\nnested:\n  item: 1\n",
			expectError: false,
		},
		{
			name:        "invalid YAML",
			content:     "key: value\n: badcolon\n",
			expectError: true,
		},
		{
			name:        "empty file",
			content:     "",
			expectError: false,
		},
		{
			name:        "YAML list",
			content:     "- item1\n- item2\n",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file
			tempFile, err := os.CreateTemp("", "config-test-*.yaml")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer func() { _ = os.Remove(tempFile.Name()) }()

			// Write content
			if _, err := tempFile.Write([]byte(tt.content)); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			_ = tempFile.Close()

			// Validate config
			err = h.validateConfig(tempFile.Name())
			if tt.expectError && err == nil {
				t.Error("expected error for invalid YAML")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestConfigChange_RT005ReloadIntegration implements AC-V3R2-RT-006-04:
// Given .moai/config/sections/quality.yaml is edited, When ConfigChange fires,
// Then the handler re-reads the file, runs typed loader, calls diff-aware reload,
// and emits AdditionalContext: "<path> reloaded successfully".
func TestConfigChange_RT005ReloadIntegration(t *testing.T) {
	t.Skip("RED: RT-005 Manager.Reload call not yet wired in config_change.go")
	// Implementation will:
	// 1. Create valid quality.yaml with known content
	// 2. Mock config.ManagerFromContext(ctx) to return Manager with Reload() tracking
	// 3. Call handler.Handle() with ConfigChange event
	// 4. Verify mgr.Reload(input.ConfigFilePath) was called once
	// 5. Verify HookOutput.AdditionalContext contains "reloaded successfully"
	// 6. Verify HookOutput.Continue is true
}

// TestConfigChange_InvalidYAMLKeepsOldSettings implements AC-V3R2-RT-006-05 and
// REQ-V3R2-RT-006-062: Given a quality.yaml edit introduces an invalid field,
// When ConfigChange reload fails validation, Then SystemMessage reports the error
// and Continue: false is set, AND old settings remain in memory.
func TestConfigChange_InvalidYAMLKeepsOldSettings(t *testing.T) {
	t.Skip("RED: Old settings retention logic not yet implemented")
	// Implementation will:
	// 1. Start with valid quality.yaml in memory
	// 2. Simulate ConfigChange with malformed YAML (unknown field)
	// 3. Verify handler rejects the reload (validation error)
	// 4. Verify HookOutput.SystemMessage contains "Config reload rejected: <field>: <error>"
	// 5. Verify HookOutput.Continue is false
	// 6. Verify old config remains in memory (not replaced with invalid new config)
}
