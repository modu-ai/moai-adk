package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// stubConfigProvider is a minimal ConfigProvider for testing.
type stubConfigProvider struct {
	cfg *config.Config
}

func (s *stubConfigProvider) Get() *config.Config {
	return s.cfg
}

func TestSubagentStartHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewSubagentStartHandler()

	if got := h.EventType(); got != EventSubagentStart {
		t.Errorf("EventType() = %q, want %q", got, EventSubagentStart)
	}
}

func TestSubagentStartHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "subagent with transcript path",
			input: &HookInput{
				SessionID:           "sess-sa-1",
				AgentID:             "agent-1",
				AgentTranscriptPath: "/tmp/transcript.jsonl",
				HookEventName:       "SubagentStart",
			},
		},
		{
			name: "subagent without transcript",
			input: &HookInput{
				SessionID:     "sess-sa-2",
				AgentID:       "agent-2",
				HookEventName: "SubagentStart",
			},
		},
		{
			name: "minimal input",
			input: &HookInput{
				SessionID: "sess-sa-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewSubagentStartHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			} else if got.HookSpecificOutput != nil {
				t.Error("SubagentStart hook without config should not set hookSpecificOutput")
			}
		})
	}
}

func TestSubagentStartHandlerWithConfig_NilConfig(t *testing.T) {
	t.Parallel()

	h := NewSubagentStartHandlerWithConfig(nil)
	ctx := context.Background()
	got, err := h.Handle(ctx, &HookInput{SessionID: "sess-nil-cfg"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
		return // staticcheck SA5011
	}
	// nil config must not produce additionalContext
	if got.HookSpecificOutput != nil {
		t.Errorf("nil config should not produce hookSpecificOutput, got %+v", got.HookSpecificOutput)
	}
}

func TestSubagentStartHandlerWithConfig_ReturnsProjectMetadata(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Project: models.ProjectConfig{
			Name:     "moai-adk",
			Type:     "backend",
			Language: "go",
		},
	}
	stub := &stubConfigProvider{cfg: cfg}

	h := NewSubagentStartHandlerWithConfig(stub)
	ctx := context.Background()
	got, err := h.Handle(ctx, &HookInput{SessionID: "sess-meta"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	}
	if got.HookSpecificOutput == nil {
		t.Fatal("expected hookSpecificOutput, got nil")
	}
	if got.HookSpecificOutput.AdditionalContext == "" {
		t.Error("expected non-empty additionalContext")
	}
	// Verify project metadata is present in output
	ctx2 := got.HookSpecificOutput.AdditionalContext
	for _, want := range []string{"moai-adk", "backend", "go"} {
		if !contains(ctx2, want) {
			t.Errorf("additionalContext %q missing %q", ctx2, want)
		}
	}
	// Verify length constraint
	if len(ctx2) > 199 {
		t.Errorf("additionalContext length %d exceeds 199 chars", len(ctx2))
	}
}

func TestSubagentStartHandlerWithConfig_HookEventName(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Project: models.ProjectConfig{
			Name: "test-project",
		},
	}
	stub := &stubConfigProvider{cfg: cfg}

	h := NewSubagentStartHandlerWithConfig(stub)
	ctx := context.Background()
	got, err := h.Handle(ctx, &HookInput{SessionID: "sess-event"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.HookSpecificOutput == nil {
		t.Fatal("expected hookSpecificOutput")
	}
	if got.HookSpecificOutput.HookEventName != "SubagentStart" {
		t.Errorf("HookEventName = %q, want %q", got.HookSpecificOutput.HookEventName, "SubagentStart")
	}
}

func TestSubagentStartHandlerWithConfig_IncludesActiveSpec(t *testing.T) {
	t.Parallel()

	// Set up a temporary project directory with a SPEC file
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}
	specFile := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specFile, []byte("# SPEC-TEST-001: Test Feature\n"), 0o644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	cfg := &config.Config{
		Project: models.ProjectConfig{
			Name: "test-project",
		},
	}
	stub := &stubConfigProvider{cfg: cfg}

	h := NewSubagentStartHandlerWithConfig(stub)
	ctx := context.Background()
	got, err := h.Handle(ctx, &HookInput{
		SessionID: "sess-spec",
		CWD:       tmpDir,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.HookSpecificOutput == nil {
		t.Fatal("expected hookSpecificOutput")
	}
	if !contains(got.HookSpecificOutput.AdditionalContext, "SPEC-TEST-001") {
		t.Errorf("additionalContext %q missing SPEC-TEST-001", got.HookSpecificOutput.AdditionalContext)
	}
}

func TestSubagentStartHandlerWithConfig_EmptyConfig(t *testing.T) {
	t.Parallel()

	// Config with no project fields should produce no output
	stub := &stubConfigProvider{cfg: &config.Config{}}

	h := NewSubagentStartHandlerWithConfig(stub)
	ctx := context.Background()
	got, err := h.Handle(ctx, &HookInput{SessionID: "sess-empty"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("got nil output")
	}
	// No fields → no hookSpecificOutput
	if got.HookSpecificOutput != nil {
		t.Errorf("empty config should not produce hookSpecificOutput, got %+v", got.HookSpecificOutput)
	}
}

// contains is a helper to check substring presence.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
