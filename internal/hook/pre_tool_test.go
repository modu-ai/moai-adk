package hook

import (
	"context"
	"encoding/json"
	"testing"
)

func TestPreToolHandler_EventType(t *testing.T) {
	t.Parallel()

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	h := NewPreToolHandler(cfg, DefaultSecurityPolicy())

	if got := h.EventType(); got != EventPreToolUse {
		t.Errorf("EventType() = %q, want %q", got, EventPreToolUse)
	}
}

func TestPreToolHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		policy       *SecurityPolicy
		input        *HookInput
		wantDecision string
		wantReason   bool
	}{
		{
			name:   "allowed tool passes security check",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Read",
				ToolInput:     json.RawMessage(`{"file_path": "/tmp/test.go"}`),
			},
			wantDecision: DecisionAllow,
		},
		{
			name: "blocked tool is rejected",
			policy: &SecurityPolicy{
				BlockedTools: []string{"DangerousTool"},
			},
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "DangerousTool",
			},
			wantDecision: DecisionDeny,
			wantReason:   true,
		},
		{
			name:   "Bash tool with dangerous rm -rf / command is blocked",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(`{"command": "rm -rf /"}`),
			},
			wantDecision: DecisionDeny,
			wantReason:   true,
		},
		{
			name:   "Bash tool with safe command is allowed",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(`{"command": "go test ./..."}`),
			},
			wantDecision: DecisionAllow,
		},
		{
			name:   "empty tool name is allowed (no policy match)",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "",
			},
			wantDecision: DecisionAllow,
		},
		{
			name:   "nil policy allows everything",
			policy: nil,
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(`{"command": "rm -rf /"}`),
			},
			wantDecision: DecisionAllow,
		},
		{
			name:   "Bash tool with dangerous fork bomb is blocked",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Bash",
				ToolInput:     json.RawMessage(`{"command": ":(){ :|:& };:"}`),
			},
			wantDecision: DecisionBlock,
			wantReason:   true,
		},
		{
			name:   "tool with nil tool_input is allowed",
			policy: DefaultSecurityPolicy(),
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PreToolUse",
				ToolName:      "Read",
			},
			wantDecision: DecisionAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfigProvider{cfg: newTestConfig()}
			h := NewPreToolHandler(cfg, tt.policy)

			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.Decision != tt.wantDecision {
				t.Errorf("Decision = %q, want %q", got.Decision, tt.wantDecision)
			}
			if tt.wantReason && got.Reason == "" {
				t.Error("expected non-empty Reason for block decision")
			}
		})
	}
}

func TestDefaultSecurityPolicy(t *testing.T) {
	t.Parallel()

	policy := DefaultSecurityPolicy()

	if policy == nil {
		t.Fatal("DefaultSecurityPolicy() returned nil")
	}
	if len(policy.DangerousBashPatterns) == 0 {
		t.Error("DangerousBashPatterns should not be empty")
	}
	if len(policy.DenyPatterns) == 0 {
		t.Error("DenyPatterns should not be empty")
	}
	if len(policy.AskPatterns) == 0 {
		t.Error("AskPatterns should not be empty")
	}
	if len(policy.SensitiveContentPatterns) == 0 {
		t.Error("SensitiveContentPatterns should not be empty")
	}
}
