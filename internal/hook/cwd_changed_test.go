package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCwdChangedHandler_EnvFileAppendPreservesContent is the G8 regression
// test: CLAUDE_ENV_FILE may already carry user/session content, so the write
// MUST append (the old os.WriteFile truncated and clobbered it), and repeat
// invocations MUST be idempotent (no duplicate export lines).
func TestCwdChangedHandler_EnvFileAppendPreservesContent(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "claude-env")

	// Pre-existing content that must survive the hook write.
	preexisting := "export USER_SET_VAR=\"keep-me\"\n"
	if err := os.WriteFile(envFile, []byte(preexisting), 0o644); err != nil {
		t.Fatalf("seed env file: %v", err)
	}

	moaiDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("failed to create moai config dir: %v", err)
	}

	t.Setenv("CLAUDE_ENV_FILE", envFile)

	h := NewCwdChangedHandler()
	input := &HookInput{SessionID: "sess-append", CWD: tmpDir, NewCwd: tmpDir}
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "USER_SET_VAR=\"keep-me\"") {
		t.Errorf("pre-existing content was clobbered (append required): %q", content)
	}
	if !strings.Contains(content, "MOAI_PROJECT_DIR=") {
		t.Errorf("MOAI_PROJECT_DIR export missing: %q", content)
	}

	// Idempotency: a second invocation must not duplicate the export line.
	if _, err := h.Handle(context.Background(), input); err != nil {
		t.Fatalf("Handle() second call error = %v", err)
	}
	data2, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("re-read env file: %v", err)
	}
	if n := strings.Count(string(data2), "MOAI_PROJECT_DIR="); n != 1 {
		t.Errorf("MOAI_PROJECT_DIR export appears %d times, want 1 (idempotent append): %q", n, string(data2))
	}
}

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
