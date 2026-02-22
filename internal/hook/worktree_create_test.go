package hook

import (
	"context"
	"testing"
)

func TestWorktreeCreateHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewWorktreeCreateHandler()

	if got := h.EventType(); got != EventWorktreeCreate {
		t.Errorf("EventType() = %q, want %q", got, EventWorktreeCreate)
	}
}

func TestWorktreeCreateHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "full worktree create payload",
			input: &HookInput{
				SessionID:      "sess-wt-1",
				AgentID:        "agent-backend-dev",
				AgentName:      "team-backend-dev",
				WorktreePath:   "/repo/.claude/worktrees/backend-impl",
				WorktreeBranch: "agent/backend-impl",
				HookEventName:  "WorktreeCreate",
			},
		},
		{
			name: "minimal worktree create",
			input: &HookInput{
				SessionID:    "sess-wt-2",
				WorktreePath: "/repo/.claude/worktrees/agent-xyz",
			},
		},
		{
			name: "empty input",
			input: &HookInput{
				SessionID: "sess-wt-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewWorktreeCreateHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput != nil {
				t.Error("WorktreeCreate hook should not set hookSpecificOutput")
			}
			if got.Decision != "" {
				t.Errorf("WorktreeCreate hook should not set decision, got %q", got.Decision)
			}
		})
	}
}
