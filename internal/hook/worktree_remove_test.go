package hook

import (
	"context"
	"testing"
)

func TestWorktreeRemoveHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewWorktreeRemoveHandler()

	if got := h.EventType(); got != EventWorktreeRemove {
		t.Errorf("EventType() = %q, want %q", got, EventWorktreeRemove)
	}
}

func TestWorktreeRemoveHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "full worktree remove payload",
			input: &HookInput{
				SessionID:      "sess-wtr-1",
				AgentID:        "agent-frontend-dev",
				AgentName:      "team-frontend-dev",
				WorktreePath:   "/repo/.claude/worktrees/frontend-impl",
				WorktreeBranch: "agent/frontend-impl",
				HookEventName:  "WorktreeRemove",
			},
		},
		{
			name: "minimal worktree remove",
			input: &HookInput{
				SessionID:    "sess-wtr-2",
				WorktreePath: "/repo/.claude/worktrees/agent-abc",
			},
		},
		{
			name: "empty input",
			input: &HookInput{
				SessionID: "sess-wtr-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewWorktreeRemoveHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			if got.HookSpecificOutput != nil {
				t.Error("WorktreeRemove hook should not set hookSpecificOutput")
			}
			if got.Decision != "" {
				t.Errorf("WorktreeRemove hook should not set decision, got %q", got.Decision)
			}
		})
	}
}
