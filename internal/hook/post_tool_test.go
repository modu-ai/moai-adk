package hook

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestPostToolHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewPostToolHandler()

	if got := h.EventType(); got != EventPostToolUse {
		t.Errorf("EventType() = %q, want %q", got, EventPostToolUse)
	}
}

func TestPostToolHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        *HookInput
		wantDecision string
		checkData    bool
	}{
		{
			name: "normal tool output with metrics",
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PostToolUse",
				ToolName:      "Write",
				ToolInput:     json.RawMessage(`{"file_path": "main.go"}`),
				ToolOutput:    json.RawMessage(`{"success": true, "path": "main.go"}`),
			},
			wantDecision: DecisionAllow,
			checkData:    true,
		},
		{
			name: "empty tool output",
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PostToolUse",
				ToolName:      "Read",
			},
			wantDecision: DecisionAllow,
		},
		{
			name: "large tool output",
			input: &HookInput{
				SessionID:     "sess-1",
				CWD:           "/tmp",
				HookEventName: "PostToolUse",
				ToolName:      "Bash",
				ToolOutput:    json.RawMessage(`{"output": "` + strings.Repeat("x", 10000) + `"}`),
			},
			wantDecision: DecisionAllow,
			checkData:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPostToolHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}
			// PostToolUse handler is observation-only and uses hookSpecificOutput.hookEventName
			if got.HookSpecificOutput == nil {
				t.Fatal("HookSpecificOutput is nil")
			}
			if got.HookSpecificOutput.HookEventName != "PostToolUse" {
				t.Errorf("HookEventName = %q, want %q", got.HookSpecificOutput.HookEventName, "PostToolUse")
			}
			if tt.checkData && got.Data != nil {
				if !json.Valid(got.Data) {
					t.Errorf("Data is not valid JSON: %s", got.Data)
				}
			}
		})
	}
}
