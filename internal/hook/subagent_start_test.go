package hook

import (
	"context"
	"testing"
)

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
			}
			if got.HookSpecificOutput != nil {
				t.Error("SubagentStart hook should not set hookSpecificOutput")
			}
		})
	}
}
