package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// newTestConfigWithMemory returns a test config with memory settings configured.
func newTestConfigWithMemory(enabled bool, memDir string, maxTokens int, autoInject bool) *config.Config {
	c := newTestConfig()
	c.Memory = config.MemoryConfig{
		Enabled:    enabled,
		MemoryDir:  memDir,
		MaxTokens:  maxTokens,
		AutoInject: autoInject,
	}
	return c
}

func TestSessionStartHandler_EventType(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	h := NewSessionStartHandler(cfg)

	if got := h.EventType(); got != EventSessionStart {
		t.Errorf("EventType() = %q, want %q", got, EventSessionStart)
	}
}

func TestSessionStartHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		cfg          *config.Config
		input        *HookInput
		wantDecision string
		wantDataKeys []string
	}{
		{
			name: "normal session initialization with project config",
			cfg: func() *config.Config {
				c := newTestConfig()
				c.Project = models.ProjectConfig{
					Name:     "moai-adk-go",
					Type:     models.ProjectTypeCLI,
					Language: "go",
				}
				return c
			}(),
			input: &HookInput{
				SessionID:     "sess-abc-123",
				CWD:           t.TempDir(),
				HookEventName: "SessionStart",
				ProjectDir:    t.TempDir(),
			},
			wantDecision: DecisionAllow,
			wantDataKeys: []string{"project_name"},
		},
		{
			name: "session start with nil config returns allow",
			cfg:  nil,
			input: &HookInput{
				SessionID:     "sess-nil-cfg",
				CWD:           t.TempDir(),
				HookEventName: "SessionStart",
			},
			wantDecision: DecisionAllow,
		},
		{
			name: "session start with empty project config returns allow",
			cfg:  newTestConfig(),
			input: &HookInput{
				SessionID:     "sess-empty",
				CWD:           t.TempDir(),
				HookEventName: "SessionStart",
			},
			wantDecision: DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: tt.cfg}
			h := NewSessionStartHandler(cfg)

			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			// SessionStart does NOT use hookSpecificOutput per Claude Code protocol
			if got.HookSpecificOutput != nil {
				t.Errorf("HookSpecificOutput should be nil for SessionStart, got %+v", got.HookSpecificOutput)
			}

			if len(tt.wantDataKeys) > 0 && got.Data != nil {
				var data map[string]any
				if err := json.Unmarshal(got.Data, &data); err != nil {
					t.Fatalf("unmarshal data: %v", err)
				}
				for _, key := range tt.wantDataKeys {
					if _, ok := data[key]; !ok {
						t.Errorf("data missing key %q", key)
					}
				}
			}
		})
	}
}

func TestSessionStartHandler_LoadMemory(t *testing.T) {
	t.Parallel()

	// Create temp dir with .moai/memory/MEMORY.md
	projectDir := t.TempDir()
	memDir := filepath.Join(projectDir, ".moai", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	memContent := "# Project Memory\n\nThis is the project memory content."
	if err := os.WriteFile(filepath.Join(memDir, "MEMORY.md"), []byte(memContent), 0o644); err != nil {
		t.Fatalf("failed to write MEMORY.md: %v", err)
	}

	cfg := newTestConfigWithMemory(true, ".moai/memory", 5000, true)
	provider := &mockConfigProvider{cfg: cfg}
	h := NewSessionStartHandler(provider)

	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-memory-test",
		CWD:           projectDir,
		HookEventName: "SessionStart",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	}
	if got.SystemMessage == "" {
		t.Error("SystemMessage should not be empty when MEMORY.md exists")
	}
	if got.SystemMessage != strings.TrimSpace(memContent) {
		t.Errorf("SystemMessage = %q, want %q", got.SystemMessage, strings.TrimSpace(memContent))
	}
}

func TestSessionStartHandler_LoadMemory_Disabled(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	memDir := filepath.Join(projectDir, ".moai", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(memDir, "MEMORY.md"), []byte("some memory"), 0o644); err != nil {
		t.Fatalf("failed to write MEMORY.md: %v", err)
	}

	// Disable memory injection
	cfg := newTestConfigWithMemory(false, ".moai/memory", 5000, true)
	provider := &mockConfigProvider{cfg: cfg}
	h := NewSessionStartHandler(provider)

	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-disabled",
		CWD:           projectDir,
		HookEventName: "SessionStart",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SystemMessage != "" {
		t.Errorf("SystemMessage should be empty when memory disabled, got %q", got.SystemMessage)
	}
}

func TestSessionStartHandler_LoadMemory_AutoInjectFalse(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	memDir := filepath.Join(projectDir, ".moai", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(memDir, "MEMORY.md"), []byte("some memory"), 0o644); err != nil {
		t.Fatalf("failed to write MEMORY.md: %v", err)
	}

	// Enabled but auto_inject false
	cfg := newTestConfigWithMemory(true, ".moai/memory", 5000, false)
	provider := &mockConfigProvider{cfg: cfg}
	h := NewSessionStartHandler(provider)

	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-no-inject",
		CWD:           projectDir,
		HookEventName: "SessionStart",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SystemMessage != "" {
		t.Errorf("SystemMessage should be empty when auto_inject false, got %q", got.SystemMessage)
	}
}

func TestSessionStartHandler_LoadMemory_NoFile(t *testing.T) {
	t.Parallel()

	// No MEMORY.md file - should return empty without error
	projectDir := t.TempDir()

	cfg := newTestConfigWithMemory(true, ".moai/memory", 5000, true)
	provider := &mockConfigProvider{cfg: cfg}
	h := NewSessionStartHandler(provider)

	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-no-file",
		CWD:           projectDir,
		HookEventName: "SessionStart",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SystemMessage != "" {
		t.Errorf("SystemMessage should be empty when MEMORY.md not found, got %q", got.SystemMessage)
	}
}

func TestSessionStartHandler_LoadMemory_Truncation(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	memDir := filepath.Join(projectDir, ".moai", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}

	// Create content longer than maxTokens * 4 chars
	maxTokens := 10 // small for testing
	longContent := strings.Repeat("x", maxTokens*4+100)
	if err := os.WriteFile(filepath.Join(memDir, "MEMORY.md"), []byte(longContent), 0o644); err != nil {
		t.Fatalf("failed to write MEMORY.md: %v", err)
	}

	cfg := newTestConfigWithMemory(true, ".moai/memory", maxTokens, true)
	provider := &mockConfigProvider{cfg: cfg}
	h := NewSessionStartHandler(provider)

	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-truncate",
		CWD:           projectDir,
		HookEventName: "SessionStart",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	maxChars := maxTokens * 4
	expectedTruncated := longContent[:maxChars] + "\n[truncated]"
	if got.SystemMessage != expectedTruncated {
		t.Errorf("SystemMessage length = %d, want truncated to %d chars + suffix",
			len(got.SystemMessage), maxChars)
	}
}

func TestSessionStartHandler_LoadMemory_EmptyFile(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	memDir := filepath.Join(projectDir, ".moai", "memory")
	if err := os.MkdirAll(memDir, 0o755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	// Write whitespace-only file
	if err := os.WriteFile(filepath.Join(memDir, "MEMORY.md"), []byte("   \n  \n"), 0o644); err != nil {
		t.Fatalf("failed to write MEMORY.md: %v", err)
	}

	cfg := newTestConfigWithMemory(true, ".moai/memory", 5000, true)
	provider := &mockConfigProvider{cfg: cfg}
	h := NewSessionStartHandler(provider)

	ctx := context.Background()
	input := &HookInput{
		SessionID:     "sess-empty-file",
		CWD:           projectDir,
		HookEventName: "SessionStart",
	}

	got, err := h.Handle(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SystemMessage != "" {
		t.Errorf("SystemMessage should be empty for whitespace-only MEMORY.md, got %q", got.SystemMessage)
	}
}
