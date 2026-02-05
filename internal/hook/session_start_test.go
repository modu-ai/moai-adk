package hook

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

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
