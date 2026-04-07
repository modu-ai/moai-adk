package hook

import (
	"context"
	"testing"
)

func TestTaskCreatedHandler_EventType(t *testing.T) {
	h := NewTaskCreatedHandler()
	if h.EventType() != EventTaskCreated {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventTaskCreated)
	}
}

func TestTaskCreatedHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "full fields",
			input: &HookInput{
				SessionID:   "sess-001",
				TaskID:      "task-123",
				TaskSubject: "Implement authentication",
			},
		},
		{
			name:  "empty input",
			input: &HookInput{},
		},
		{
			name: "session and task id only",
			input: &HookInput{
				SessionID: "sess-002",
				TaskID:    "task-456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskCreatedHandler()
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
