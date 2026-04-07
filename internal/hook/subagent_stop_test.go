package hook

import (
	"context"
	"testing"
)

func TestSubagentStopHandler_EventType(t *testing.T) {
	h := NewSubagentStopHandler()
	if h.EventType() != EventSubagentStop {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventSubagentStop)
	}
}

func TestSubagentStopHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "full fields",
			input: &HookInput{
				SessionID:           "sess-001",
				AgentID:             "agent-abc",
				AgentName:           "expert-backend",
				AgentTranscriptPath: "/tmp/transcript.json",
			},
		},
		{
			name:  "empty input",
			input: &HookInput{},
		},
		{
			name: "session id only",
			input: &HookInput{
				SessionID: "sess-002",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewSubagentStopHandler()
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
