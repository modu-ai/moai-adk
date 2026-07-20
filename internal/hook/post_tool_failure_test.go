package hook

import (
	"context"
	"encoding/json"
	"strings"
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

// contentFreeUnknownMessage is the pre-fix fixed UnknownFailure string that
// AC-HFC-005 forbids for payloads carrying observable raw error text.
const contentFreeUnknownMessage = "UnknownFailure: Tool execution failed. Review error logs for details."

// TestPostToolUseFailureHandler_NestedToolResponse covers REQ-HFC-001/003:
// realistic PostToolUseFailure payloads nest failure text under
// tool_response.error / tool_response.stderr with an empty top-level Error.
// One nested case per category (AC-HFC-003).
func TestPostToolUseFailureHandler_NestedToolResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		toolResponse     string
		expectedCategory ErrorCategory
	}{
		{
			// AC-HFC-001b
			name:             "nested timeout in stderr",
			toolResponse:     `{"stderr":"run: context deadline exceeded while waiting"}`,
			expectedCategory: TimeoutError,
		},
		{
			// AC-HFC-001a
			name:             "nested permission denied in error",
			toolResponse:     `{"error":"permission denied: open /f","stderr":""}`,
			expectedCategory: PermissionDenied,
		},
		{
			name:             "nested context canceled",
			toolResponse:     `{"error":"context canceled"}`,
			expectedCategory: ContextCancelled,
		},
		{
			name:             "nested sandbox violation",
			toolResponse:     `{"stderr":"seccomp filter blocked syscall"}`,
			expectedCategory: SandboxViolation,
		},
		{
			name:             "nested oom killed",
			toolResponse:     `{"error":"process killed: out of memory"}`,
			expectedCategory: OOMKilled,
		},
		{
			name:             "nested exit status",
			toolResponse:     `{"error":"exit status 1","stderr":"make: *** [build] Error 1"}`,
			expectedCategory: ExitError,
		},
		{
			name:             "nested genuinely unclassifiable",
			toolResponse:     `{"error":"something went wrong"}`,
			expectedCategory: UnknownFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPostToolUseFailureHandler()
			input := &HookInput{
				SessionID:     "sess-nested",
				ToolName:      "Bash",
				ToolUseID:     "tool-nested",
				HookEventName: "PostToolUseFailure",
				ToolResponse:  json.RawMessage(tt.toolResponse),
			}

			got, err := h.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil || got.SystemMessage == "" {
				t.Fatal("expected non-empty SystemMessage")
			}
			expectedPrefix := string(tt.expectedCategory) + ":"
			if !strings.HasPrefix(got.SystemMessage, expectedPrefix) {
				t.Errorf("SystemMessage = %q, want prefix %q", got.SystemMessage, expectedPrefix)
			}
		})
	}
}

// TestPostToolUseFailureHandler_UnknownFailureExcerpt covers REQ-HFC-005
// (AC-HFC-005): a genuinely unclassifiable failure must carry a bounded
// excerpt of the observed raw error text, never the content-free string.
func TestPostToolUseFailureHandler_UnknownFailureExcerpt(t *testing.T) {
	t.Parallel()

	h := NewPostToolUseFailureHandler()
	input := &HookInput{
		SessionID:     "sess-unk",
		ToolName:      "Bash",
		HookEventName: "PostToolUseFailure",
		ToolResponse:  json.RawMessage(`{"error":"something went wrong"}`),
	}

	got, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.SystemMessage == contentFreeUnknownMessage {
		t.Fatalf("SystemMessage is the content-free string: %q", got.SystemMessage)
	}
	if !strings.HasPrefix(got.SystemMessage, "UnknownFailure:") {
		t.Errorf("SystemMessage = %q, want UnknownFailure: prefix", got.SystemMessage)
	}
	if !strings.Contains(got.SystemMessage, "something went wrong") {
		t.Errorf("SystemMessage = %q, want raw-error excerpt %q", got.SystemMessage, "something went wrong")
	}

	// Boundedness: a very long raw error is truncated to maxErrorExcerptLen.
	long := strings.Repeat("z", 500)
	got2, err := h.Handle(context.Background(), &HookInput{
		SessionID:     "sess-unk2",
		ToolName:      "Bash",
		HookEventName: "PostToolUseFailure",
		Error:         long,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got2.SystemMessage, "...") {
		t.Errorf("long excerpt not truncated: %q", got2.SystemMessage[:50])
	}
	if len(got2.SystemMessage) > len("UnknownFailure: Tool execution failed: ")+maxErrorExcerptLen+3 {
		t.Errorf("excerpt exceeds bound: len=%d", len(got2.SystemMessage))
	}
}

// TestPostToolUseFailureHandler_TopLevelErrorPrecedence covers REQ-HFC-002:
// when both top-level Error and nested tool_response text are present, the
// top-level Error wins for the displayed excerpt while classification sees both.
func TestPostToolUseFailureHandler_TopLevelErrorPrecedence(t *testing.T) {
	t.Parallel()

	h := NewPostToolUseFailureHandler()
	input := &HookInput{
		SessionID:     "sess-prec",
		ToolName:      "Bash",
		HookEventName: "PostToolUseFailure",
		Error:         "top-level mystery text",
		ToolResponse:  json.RawMessage(`{"error":"nested other text"}`),
	}

	got, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both sources unclassifiable → UnknownFailure with the TOP-LEVEL excerpt.
	if !strings.HasPrefix(got.SystemMessage, "UnknownFailure:") {
		t.Fatalf("SystemMessage = %q, want UnknownFailure: prefix", got.SystemMessage)
	}
	if !strings.Contains(got.SystemMessage, "top-level mystery text") {
		t.Errorf("SystemMessage = %q, want top-level excerpt", got.SystemMessage)
	}
	// Classification still considers nested text when top-level misses a category.
	input2 := &HookInput{
		SessionID:     "sess-prec2",
		ToolName:      "Bash",
		HookEventName: "PostToolUseFailure",
		Error:         "opaque wrapper failure",
		ToolResponse:  json.RawMessage(`{"stderr":"permission denied: open /etc/x"}`),
	}
	got2, err := h.Handle(context.Background(), input2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got2.SystemMessage, "PermissionDenied:") {
		t.Errorf("SystemMessage = %q, want PermissionDenied: (classification union)", got2.SystemMessage)
	}
}

// TestPostToolUseFailureHandler_ToolResponseResilience covers REQ-HFC-006
// (AC-HFC-006): malformed, bare-string, and absent tool_response degrade
// gracefully — no panic, no handler error, non-empty SystemMessage.
func TestPostToolUseFailureHandler_ToolResponseResilience(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		toolResponse json.RawMessage
	}{
		{name: "malformed json", toolResponse: json.RawMessage(`[}`)},
		{name: "bare json string", toolResponse: json.RawMessage(`"boom went the tool"`)},
		{name: "json array", toolResponse: json.RawMessage(`["a","b"]`)},
		{name: "absent tool_response", toolResponse: nil},
		{name: "empty object", toolResponse: json.RawMessage(`{}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewPostToolUseFailureHandler()
			input := &HookInput{
				SessionID:     "sess-res",
				ToolName:      "Bash",
				HookEventName: "PostToolUseFailure",
				ToolResponse:  tt.toolResponse,
			}
			got, err := h.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil || got.SystemMessage == "" {
				t.Fatal("expected non-empty SystemMessage (fail-open)")
			}
		})
	}
}

// TestResolveTraceSessionID covers REQ-HFC-004 (AC-HFC-004): when session_id
// is absent (substituted to "unknown"), the trace session identifier is
// derived from the transcript_path session UUID; a present session_id wins;
// an unresolvable payload keeps the documented "unknown" last-resort fallback.
func TestResolveTraceSessionID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *HookInput
		want  string
	}{
		{
			name:  "explicit session_id wins",
			input: &HookInput{SessionID: "sess-explicit", TranscriptPath: "/x/.claude/projects/p/0d661925-1111-2222-3333-444455556666.jsonl"},
			want:  "sess-explicit",
		},
		{
			name:  "unknown session_id derives UUID from transcript_path",
			input: &HookInput{SessionID: "unknown", TranscriptPath: "/x/.claude/projects/p/0d661925-1111-2222-3333-444455556666.jsonl"},
			want:  "0d661925-1111-2222-3333-444455556666",
		},
		{
			name:  "empty session_id derives UUID from transcript_path",
			input: &HookInput{SessionID: "", TranscriptPath: "/x/.claude/projects/p/0d661925-1111-2222-3333-444455556666.jsonl"},
			want:  "0d661925-1111-2222-3333-444455556666",
		},
		{
			name:  "non-UUID transcript base falls back to unknown",
			input: &HookInput{SessionID: "unknown", TranscriptPath: "/x/notes.jsonl"},
			want:  "unknown",
		},
		{
			name:  "no sources falls back to unknown",
			input: &HookInput{SessionID: "unknown"},
			want:  "unknown",
		},
		{
			name:  "nil input falls back to empty",
			input: nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveTraceSessionID(tt.input); got != tt.want {
				t.Errorf("resolveTraceSessionID() = %q, want %q", got, tt.want)
			}
		})
	}
}
