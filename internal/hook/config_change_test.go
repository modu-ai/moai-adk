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
