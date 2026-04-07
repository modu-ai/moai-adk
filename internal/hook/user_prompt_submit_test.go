package hook

import (
	"context"
	"testing"
)

func TestUserPromptSubmitHandler_EventType(t *testing.T) {
	h := NewUserPromptSubmitHandler()
	if h.EventType() != EventUserPromptSubmit {
		t.Errorf("EventType() = %v, want %v", h.EventType(), EventUserPromptSubmit)
	}
}

func TestDetectWorkflowContext(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		wantEmpty   bool
		wantKeyword string
	}{
		{
			name:        "contains loop keyword",
			prompt:      "/moai loop fix errors",
			wantEmpty:   false,
			wantKeyword: "loop",
		},
		{
			name:        "contains run keyword",
			prompt:      "/moai run SPEC-001",
			wantEmpty:   false,
			wantKeyword: "run",
		},
		{
			name:        "contains plan keyword",
			prompt:      "/moai plan add authentication",
			wantEmpty:   false,
			wantKeyword: "plan",
		},
		{
			name:      "no workflow keyword",
			prompt:    "what is the weather today",
			wantEmpty: true,
		},
		{
			name:      "empty prompt",
			prompt:    "",
			wantEmpty: true,
		},
		{
			name:        "case insensitive LOOP",
			prompt:      "LOOP until fixed",
			wantEmpty:   false,
			wantKeyword: "loop",
		},
		{
			name:        "keyword embedded in word",
			prompt:      "please plan the work",
			wantEmpty:   false,
			wantKeyword: "plan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectWorkflowContext(tt.prompt)
			if tt.wantEmpty && got != "" {
				t.Errorf("detectWorkflowContext(%q) = %q, want empty", tt.prompt, got)
			}
			if !tt.wantEmpty {
				if got == "" {
					t.Errorf("detectWorkflowContext(%q) = empty, want non-empty (keyword: %s)", tt.prompt, tt.wantKeyword)
				}
			}
		})
	}
}

func TestUserPromptSubmitHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		input          *HookInput
		wantAdditional bool
	}{
		{
			name: "prompt with loop keyword",
			input: &HookInput{
				SessionID: "sess-001",
				Prompt:    "/moai loop",
			},
			wantAdditional: true,
		},
		{
			name: "prompt with run keyword",
			input: &HookInput{
				SessionID: "sess-002",
				Prompt:    "/moai run SPEC-001",
			},
			wantAdditional: true,
		},
		{
			name: "prompt with plan keyword",
			input: &HookInput{
				SessionID: "sess-003",
				Prompt:    "let's plan the feature",
			},
			wantAdditional: true,
		},
		{
			name: "normal prompt no keywords",
			input: &HookInput{
				SessionID: "sess-004",
				Prompt:    "explain this function",
			},
			wantAdditional: false,
		},
		{
			name: "long prompt gets truncated in log",
			input: &HookInput{
				SessionID: "sess-005",
				Prompt:    "this is a very long prompt that exceeds one hundred characters and should be truncated when logged but still fully checked for keywords including run",
			},
			wantAdditional: true,
		},
		{
			name:           "empty input",
			input:          &HookInput{},
			wantAdditional: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserPromptSubmitHandler()
			out, err := h.Handle(context.Background(), tt.input)
			if err != nil {
				t.Errorf("Handle() error = %v, want nil", err)
			}
			if out == nil {
				t.Fatal("Handle() returned nil output")
			}

			hasAdditional := out.HookSpecificOutput != nil && out.HookSpecificOutput.AdditionalContext != ""
			if hasAdditional != tt.wantAdditional {
				t.Errorf("HookSpecificOutput.AdditionalContext non-empty = %v, want %v", hasAdditional, tt.wantAdditional)
			}
		})
	}
}
