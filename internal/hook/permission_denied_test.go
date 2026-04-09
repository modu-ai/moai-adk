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
		name      string
		input     *HookInput
		wantRetry bool
	}{
		// Read-only tools must signal retry.
		{
			name:      "Read tool retries",
			input:     &HookInput{SessionID: "sess-001", ToolName: "Read"},
			wantRetry: true,
		},
		{
			name:      "Grep tool retries",
			input:     &HookInput{SessionID: "sess-001", ToolName: "Grep"},
			wantRetry: true,
		},
		{
			name:      "Glob tool retries",
			input:     &HookInput{SessionID: "sess-001", ToolName: "Glob"},
			wantRetry: true,
		},
		{
			name:      "WebFetch tool retries",
			input:     &HookInput{SessionID: "sess-001", ToolName: "WebFetch"},
			wantRetry: true,
		},
		{
			name:      "WebSearch tool retries",
			input:     &HookInput{SessionID: "sess-001", ToolName: "WebSearch"},
			wantRetry: true,
		},
		{
			name:      "Skill tool retries",
			input:     &HookInput{SessionID: "sess-001", ToolName: "Skill"},
			wantRetry: true,
		},
		{
			name:      "ListMcpResourcesTool retries",
			input:     &HookInput{SessionID: "sess-001", ToolName: "ListMcpResourcesTool"},
			wantRetry: true,
		},
		{
			name:      "ReadMcpResourceTool retries",
			input:     &HookInput{SessionID: "sess-001", ToolName: "ReadMcpResourceTool"},
			wantRetry: true,
		},
		// Write and execute tools must not retry.
		{
			name:      "Write tool does not retry",
			input:     &HookInput{SessionID: "sess-002", ToolName: "Write"},
			wantRetry: false,
		},
		{
			name:      "Edit tool does not retry",
			input:     &HookInput{SessionID: "sess-002", ToolName: "Edit"},
			wantRetry: false,
		},
		{
			name:      "Bash tool does not retry",
			input:     &HookInput{SessionID: "sess-002", ToolName: "Bash"},
			wantRetry: false,
		},
		// Unknown tools must not retry.
		{
			name:      "unknown tool does not retry",
			input:     &HookInput{SessionID: "sess-003", ToolName: "UnknownTool"},
			wantRetry: false,
		},
		{
			name:      "empty tool name does not retry",
			input:     &HookInput{},
			wantRetry: false,
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
				t.Fatal("Handle() returned nil output")
			}
			if out.Retry != tt.wantRetry {
				t.Errorf("Handle() Retry = %v, want %v (tool=%q)", out.Retry, tt.wantRetry, tt.input.ToolName)
			}
		})
	}
}
