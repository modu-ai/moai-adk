package hook

import (
	"context"
	"testing"
)

func TestConfigChangeHandler_EventType(t *testing.T) {
	h := NewConfigChangeHandler()
	if h.EventType() != EventConfigChange {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventConfigChange)
	}
}

func TestConfigChangeHandler_Handle(t *testing.T) {
	tests := []struct {
		name  string
		input *HookInput
	}{
		{
			name: "config file changed",
			input: &HookInput{
				SessionID:      "sess-001",
				ConfigFilePath: "/path/to/.claude/settings.json",
			},
		},
		{
			name:  "empty input",
			input: &HookInput{},
		},
		{
			name: "quality yaml changed",
			input: &HookInput{
				SessionID:      "sess-002",
				ConfigFilePath: ".moai/config/sections/quality.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewConfigChangeHandler()
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
