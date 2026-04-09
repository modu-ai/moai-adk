package hook

import (
	"context"
	"testing"
)

func TestCwdChangedHandler_EventType(t *testing.T) {
	h := NewCwdChangedHandler()
	if h.EventType() != EventCwdChanged {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventCwdChanged)
	}
}

func TestCwdChangedHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "directory changed",
			input: &HookInput{
				SessionID: "sess-001",
				CWD:       "/Users/user/project/src",
			},
		},
		{
			name:  "empty input",
			input: &HookInput{},
		},
		{
			name: "root directory",
			input: &HookInput{
				SessionID: "sess-002",
				CWD:       "/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCwdChangedHandler()
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
