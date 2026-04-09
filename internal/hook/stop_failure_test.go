package hook

import (
	"context"
	"testing"
)

func TestStopFailureHandler_EventType(t *testing.T) {
	h := NewStopFailureHandler()
	if h.EventType() != EventStopFailure {
		t.Errorf("EventType() = %q, want %q", h.EventType(), EventStopFailure)
	}
}

func TestStopFailureHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		errorType     string // input.ErrorType (new protocol)
		fallbackError string // input.Error (old protocol fallback)
		wantMsg       string // expected SystemMessage; empty means no message
	}{
		{
			name:      "rate_limit returns helpful message",
			errorType: "rate_limit",
			wantMsg:   "Rate limit reached. Wait a moment before continuing.",
		},
		{
			name:      "authentication_failed returns helpful message",
			errorType: "authentication_failed",
			wantMsg:   "Authentication failed. Check your API key or run 'moai glm setup'.",
		},
		{
			name:      "billing_error returns helpful message",
			errorType: "billing_error",
			wantMsg:   "Billing error detected. Check your account status.",
		},
		{
			name:      "max_output_tokens returns helpful message",
			errorType: "max_output_tokens",
			wantMsg:   "Output token limit reached. Try breaking the task into smaller steps.",
		},
		{
			name:      "server_error returns empty output",
			errorType: "server_error",
			wantMsg:   "",
		},
		{
			name:      "unknown error type returns empty output",
			errorType: "unknown",
			wantMsg:   "",
		},
		{
			name:    "empty error_type returns empty output",
			wantMsg: "",
		},
		{
			name:          "old protocol: rate_limit via Error field",
			errorType:     "",
			fallbackError: "rate_limit",
			wantMsg:       "Rate limit reached. Wait a moment before continuing.",
		},
		{
			name:          "old protocol: authentication_failed via Error field",
			errorType:     "",
			fallbackError: "authentication_failed",
			wantMsg:       "Authentication failed. Check your API key or run 'moai glm setup'.",
		},
		{
			name:          "ErrorType takes precedence over Error fallback",
			errorType:     "billing_error",
			fallbackError: "rate_limit",
			wantMsg:       "Billing error detected. Check your account status.",
		},
	}

	h := NewStopFailureHandler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := &HookInput{
				SessionID:     "test-session",
				HookEventName: "StopFailure",
				ErrorType:     tt.errorType,
				Error:         tt.fallbackError,
			}
			out, err := h.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			if out.SystemMessage != tt.wantMsg {
				t.Errorf("SystemMessage = %q, want %q", out.SystemMessage, tt.wantMsg)
			}
		})
	}
}
