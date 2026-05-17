package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfigChange_RT005ReloadIntegration verifies AC-04: when ConfigChange fires
// and the file is valid YAML, the handler emits AdditionalContext (or SystemMessage)
// with "<path> reloaded successfully" (SPEC-V3R2-RT-006 REQ-011).
func TestConfigChange_RT005ReloadIntegration(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	yamlPath := filepath.Join(tempDir, "quality.yaml")
	content := []byte("development_mode: tdd\ncoverage_target: 85\n")
	if err := os.WriteFile(yamlPath, content, 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	h := NewConfigChangeHandler()
	input := &HookInput{
		SessionID:      "sess-rt005",
		ConfigFilePath: yamlPath,
		HookEventName:  "ConfigChange",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
	// REQ-011: emit "reloaded successfully" message.
	if out.SystemMessage == "" {
		t.Error("expected SystemMessage with reload confirmation")
	}
	if !out.Continue {
		t.Errorf("expected Continue=true for valid config, got false (message: %s)", out.SystemMessage)
	}
}

// TestConfigChange_InvalidYAMLKeepsOldSettings verifies AC-05: when ConfigChange
// reload fails validation, Continue:false is set AND old settings are described
// in the error message (SPEC-V3R2-RT-006 REQ-062).
func TestConfigChange_InvalidYAMLKeepsOldSettings(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	yamlPath := filepath.Join(tempDir, "quality.yaml")
	// Deliberately invalid YAML.
	content := []byte(": bad yaml: :\n  key:\n")
	if err := os.WriteFile(yamlPath, content, 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	h := NewConfigChangeHandler()
	input := &HookInput{
		SessionID:      "sess-invalid",
		ConfigFilePath: yamlPath,
		HookEventName:  "ConfigChange",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}
	// REQ-062: Continue:false + message indicates old settings retained.
	if out.Continue {
		t.Error("expected Continue=false for invalid YAML, got true")
	}
	if out.SystemMessage == "" {
		t.Error("expected SystemMessage for invalid YAML rejection")
	}
	// Message must mention "retained" or "rejected" to indicate old settings kept.
	if !containsAny(out.SystemMessage, "retained", "rejected", "reload rejected") {
		t.Errorf("SystemMessage should mention old settings retained, got: %s", out.SystemMessage)
	}
}

// containsAny returns true if s contains any of the needles.
func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(strings.ToLower(s), strings.ToLower(n)) {
			return true
		}
	}
	return false
}

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
