package hook

import (
	"context"
	"testing"
)

func TestPostToolUseFailureHandler_EventType(t *testing.T) {
	t.Parallel()

	h := NewPostToolUseFailureHandler()

	if got := h.EventType(); got != EventPostToolUseFailure {
		t.Errorf("EventType() = %q, want %q", got, EventPostToolUseFailure)
	}
}

func TestPostToolUseFailureHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            *HookInput
		expectedCategory ErrorCategory
		wantMessage      bool
	}{
		{
			name: "timeout error",
			input: &HookInput{
				SessionID:     "sess-001",
				ToolName:      "Bash",
				ToolUseID:     "tool-123",
				Error:         "context deadline exceeded",
				IsInterrupt:   false,
				HookEventName: "PostToolUseFailure",
			},
			expectedCategory: TimeoutError,
			wantMessage:      true,
		},
		{
			name: "permission denied",
			input: &HookInput{
				SessionID:     "sess-002",
				ToolName:      "Write",
				ToolUseID:     "tool-456",
				Error:         "permission denied: open /file.txt",
				IsInterrupt:   false,
				HookEventName: "PostToolUseFailure",
			},
			expectedCategory: PermissionDenied,
			wantMessage:      true,
		},
		{
			name: "context cancelled",
			input: &HookInput{
				SessionID:     "sess-003",
				ToolName:      "Read",
				ToolUseID:     "tool-789",
				Error:         "operation cancelled",
				IsInterrupt:   true,
				HookEventName: "PostToolUseFailure",
			},
			expectedCategory: ContextCancelled,
			wantMessage:      true,
		},
		{
			name: "sandbox violation",
			input: &HookInput{
				SessionID:     "sess-004",
				ToolName:      "Bash",
				ToolUseID:     "tool-abc",
				Error:         "seccomp filter violation",
				IsInterrupt:   false,
				HookEventName: "PostToolUseFailure",
			},
			expectedCategory: SandboxViolation,
			wantMessage:      true,
		},
		{
			name: "oom killed",
			input: &HookInput{
				SessionID:     "sess-005",
				ToolName:      "Bash",
				ToolUseID:     "tool-def",
				Error:         "signal: killed (exit status 137)",
				IsInterrupt:   false,
				HookEventName: "PostToolUseFailure",
			},
			expectedCategory: OOMKilled,
			wantMessage:      true,
		},
		{
			name: "exit error",
			input: &HookInput{
				SessionID:     "sess-006",
				ToolName:      "Bash",
				ToolUseID:     "tool-ghi",
				Error:         "exit status 1",
				IsInterrupt:   false,
				HookEventName: "PostToolUseFailure",
			},
			expectedCategory: ExitError,
			wantMessage:      true,
		},
		{
			name: "unknown failure",
			input: &HookInput{
				SessionID:     "sess-007",
				ToolName:      "Read",
				ToolUseID:     "tool-jkl",
				Error:         "something went wrong",
				IsInterrupt:   false,
				HookEventName: "PostToolUseFailure",
			},
			expectedCategory: UnknownFailure,
			wantMessage:      true,
		},
		{
			name: "empty error",
			input: &HookInput{
				SessionID:     "sess-008",
				ToolName:      "Bash",
				ToolUseID:     "tool-mno",
				HookEventName: "PostToolUseFailure",
			},
			expectedCategory: UnknownFailure,
			wantMessage:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPostToolUseFailureHandler()
			ctx := context.Background()
			got, err := h.Handle(ctx, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("got nil output")
			}

			// Check that system message is present
			if tt.wantMessage && got.SystemMessage == "" {
				t.Error("Handle() expected SystemMessage to be set")
			}

			// Check that message starts with category name
			if tt.wantMessage && got.SystemMessage != "" {
				expectedPrefix := string(tt.expectedCategory) + ":"
				if len(got.SystemMessage) < len(expectedPrefix) ||
					got.SystemMessage[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("Handle() SystemMessage = %v, want prefix %v", got.SystemMessage, expectedPrefix)
				}
			}
		})
	}
}

// TestPostToolFailure_TimeoutClassification implements AC-V3R2-RT-006-08:
// Given a Bash tool fails with exit code 124 (timeout),
// When PostToolUseFailure fires, Then SystemMessage classifies the error as "TimeoutError"
// with actionable hint.
func TestPostToolFailure_TimeoutClassification(t *testing.T) {
	t.Parallel()

	h := NewPostToolUseFailureHandler()

	// Simulate Bash tool failure with timeout (exit code 124)
	input := &HookInput{
		SessionID:     "test-session",
		ToolName:      "Bash",
		ToolUseID:     "tool-timeout-124",
		ExitCode:      124,
		Error:         "exit status 124",
		Stderr:        "command timed out after 30s",
		IsInterrupt:   false,
		HookEventName: "PostToolUseFailure",
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AC-08: SystemMessage MUST classify as TimeoutError
	if out.SystemMessage == "" {
		t.Error("expected SystemMessage for timeout error (AC-08)")
	}

	// Verify classification
	expectedPrefix := "TimeoutError:"
	if len(out.SystemMessage) < len(expectedPrefix) || out.SystemMessage[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("AC-08: expected 'TimeoutError:' prefix, got: %v", out.SystemMessage)
	}

	// Verify actionable hint (AdditionalContext should contain diagnostic hints)
	if out.AdditionalContext == "" {
		t.Error("AC-08: expected AdditionalContext with actionable hint")
	}

	t.Logf("AC-08 verified: SystemMessage = %v", out.SystemMessage)
}
