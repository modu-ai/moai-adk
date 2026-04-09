package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCwdChangedHandler_EventType(t *testing.T) {
	h := NewCwdChangedHandler()
	if h.EventType() != EventCwdChanged {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventCwdChanged)
	}
}

func TestCwdChangedHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "directory changed with old/new cwd",
			input: &HookInput{
				SessionID: "sess-001",
				CWD:       "/Users/user/project/src",
				OldCwd:    "/Users/user/project",
				NewCwd:    "/Users/user/project/src",
			},
		},
		{
			name:  "empty input",
			input: &HookInput{},
		},
		{
			name: "root directory",
			input: &HookInput{
				SessionID: "sess-002",
				CWD:       "/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCwdChangedHandler()
			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Errorf("Handle() error = %v, want nil", err)
			}
			if out == nil {
				t.Error("Handle() returned nil output")
			}
		})
	}
}

func TestCwdChangedHandler_EnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "claude-env")

	// Create .moai/config directory to trigger MOAI_PROJECT_DIR export
	moaiDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("failed to create moai config dir: %v", err)
	}

	t.Setenv("CLAUDE_ENV_FILE", envFile)

	h := NewCwdChangedHandler()
	_, err := h.Handle(context.Background(), &HookInput{
		SessionID: "sess-env",
		CWD:       tmpDir,
		NewCwd:    tmpDir,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("failed to read env file: %v", err)
	}
	content := string(data)
	if content == "" {
		t.Error("env file is empty, expected MOAI_PROJECT_DIR export")
	}
}

func TestCwdChangedHandler_NoEnvFileWithoutMoaiDir(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "claude-env")

	t.Setenv("CLAUDE_ENV_FILE", envFile)

	h := NewCwdChangedHandler()
	_, err := h.Handle(context.Background(), &HookInput{
		SessionID: "sess-no-moai",
		CWD:       tmpDir,
		NewCwd:    tmpDir,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	// File should not be created when no .moai/config exists
	if _, err := os.Stat(envFile); err == nil {
		t.Error("env file should not exist when no .moai/config present")
	}
}
