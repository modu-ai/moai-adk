package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInstructionsLoadedHandler_EventType(t *testing.T) {
	h := NewInstructionsLoadedHandler()
	if h.EventType() != EventInstructionsLoaded {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventInstructionsLoaded)
	}
}

func TestInstructionsLoadedHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *HookInput
		createFile    bool
		fileContent   string
		expectMessage bool
	}{
		{
			name: "no instruction file path",
			input: &HookInput{
				SessionID:     "test-session",
				HookEventName: "InstructionsLoaded",
			},
			createFile:    false,
			expectMessage: false,
		},
		{
			name: "small file within budget",
			input: &HookInput{
				SessionID:           "test-session",
				InstructionFilePath: "CLAUDE.md",
				CWD:                 "",
				HookEventName:       "InstructionsLoaded",
			},
			createFile:    true,
			fileContent:   "# Small file\n\nThis is well within budget.",
			expectMessage: false,
		},
		{
			name: "file exceeding budget",
			input: &HookInput{
				SessionID:           "test-session",
				InstructionFilePath: "CLAUDE.md",
				CWD:                 "",
				HookEventName:       "InstructionsLoaded",
			},
			createFile:    true,
			fileContent:   string(make([]byte, 45000)), // 45KB
			expectMessage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewInstructionsLoadedHandler()

			// Create temp file if needed
			if tt.createFile {
				tempDir := t.TempDir()
				filePath := tt.input.InstructionFilePath
				if !filepath.IsAbs(filePath) {
					filePath = filepath.Join(tempDir, filePath)
				}

				if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}

				tt.input.InstructionFilePath = filePath
				tt.input.CWD = tempDir
			}

			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}

			if tt.expectMessage && out.SystemMessage == "" {
				t.Error("expected SystemMessage for budget violation")
			}
			if !tt.expectMessage && out.SystemMessage != "" {
				t.Errorf("unexpected SystemMessage: %v", out.SystemMessage)
			}
		})
	}
}

func TestInstructionsLoadedHandler_CheckCharacterBudget(t *testing.T) {
	t.Parallel()

	h := &instructionsLoadedHandler{}

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "empty file",
			content:     "",
			expectError: false,
		},
		{
			name:        "small file",
			content:     "# Hello\n\nWorld",
			expectError: false,
		},
		{
			name:        "exactly at limit",
			content:     string(make([]byte, 40000)),
			expectError: false,
		},
		{
			name:        "exceeds limit by one",
			content:     string(make([]byte, 40001)),
			expectError: true,
		},
		{
			name:        "far exceeds limit",
			content:     string(make([]byte, 50000)),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp file
			tempFile, err := os.CreateTemp("", "budget-test-*.md")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer func() { _ = os.Remove(tempFile.Name()) }()

			// Write content
			if _, err := tempFile.Write([]byte(tt.content)); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			_ = tempFile.Close()

			// Check budget
			err = h.checkCharacterBudget(tempFile.Name())
			if tt.expectError && err == nil {
				t.Error("expected error for budget violation")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
