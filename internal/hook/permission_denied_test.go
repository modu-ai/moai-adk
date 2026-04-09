package hook

import (
	"context"
	"testing"
)

func TestPermissionDeniedHandler_EventType(t *testing.T) {
	h := NewPermissionDeniedHandler()
	if h.EventType() != EventPermissionDenied {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventPermissionDenied)
	}
}

func TestPermissionDeniedHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "tool denied",
			input: &HookInput{
				SessionID: "sess-001",
				ToolName:  "Bash",
			},
		},
		{
			name:  "empty input",
			input: &HookInput{},
		},
		{
			name: "write tool denied",
			input: &HookInput{
				SessionID: "sess-002",
				ToolName:  "Write",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPermissionDeniedHandler()
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
