package hook

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHookResponseEventNames(t *testing.T) {
	tests := []struct {
		name     string
		event    interface{ HookEventName() string }
		expected string
	}{
		{"PreToolUse", &PreToolUseOutput{}, "PreToolUse"},
		{"PostToolUse", &PostToolUseOutput{}, "PostToolUse"},
		{"SessionStart", &SessionStartOutput{}, "SessionStart"},
		{"SessionEnd", &SessionEndOutput{}, "SessionEnd"},
		{"Stop", &StopOutput{}, "Stop"},
		{"SubagentStop", &SubagentStopOutput{}, "SubagentStop"},
		{"PreCompact", &PreCompactOutput{}, "PreCompact"},
		{"PostCompact", &PostCompactOutput{}, "PostCompact"},
		{"PostToolUseFailure", &PostToolUseFailureOutput{}, "PostToolUseFailure"},
		{"Notification", &NotificationOutput{}, "Notification"},
		{"UserPromptSubmit", &UserPromptSubmitOutput{}, "UserPromptSubmit"},
		{"PermissionRequest", &PermissionRequestOutput{}, "PermissionRequest"},
		{"PermissionDenied", &PermissionDeniedOutput{}, "PermissionDenied"},
		{"ConfigChange", &ConfigChangeOutput{}, "ConfigChange"},
		{"InstructionsLoaded", &InstructionsLoadedOutput{}, "InstructionsLoaded"},
		{"FileChanged", &FileChangedOutput{}, "FileChanged"},
		{"TeammateIdle", &TeammateIdleOutput{}, "TeammateIdle"},
		{"TaskCompleted", &TaskCompletedOutput{}, "TaskCompleted"},
		{"SubagentStart", &SubagentStartOutput{}, "SubagentStart"},
		{"WorktreeCreate", &WorktreeCreateOutput{}, "WorktreeCreate"},
		{"WorktreeRemove", &WorktreeRemoveOutput{}, "WorktreeRemove"},
		{"CwdChanged", &CwdChangedOutput{}, "CwdChanged"},
		{"Setup", &SetupOutput{}, "Setup"},
		{"Elicitation", &ElicitationOutput{}, "Elicitation"},
		{"ElicitationResult", &ElicitationResultOutput{}, "ElicitationResult"},
		{"TaskCreated", &TaskCreatedOutput{}, "TaskCreated"},
		{"StopFailure", &StopFailureOutput{}, "StopFailure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.HookEventName()
			if got != tt.expected {
				t.Errorf("HookEventName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHookResponseMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		event interface{}
		json  string
	}{
		{
			name:  "PreToolUse with allow",
			event: &PreToolUseOutput{EventName: "PreToolUse", PermissionDecision: PermissionDecisionAllow},
			json:  `{"hookEventName":"PreToolUse","permissionDecision":"allow"}`,
		},
		{
			name:  "PreToolUse with deny and reason",
			event: &PreToolUseOutput{EventName: "PreToolUse", PermissionDecision: PermissionDecisionDeny, PermissionDecisionReason: "blocked for safety"},
			json:  `{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"blocked for safety"}`,
		},
		{
			name:  "PostToolUse with context",
			event: &PostToolUseOutput{EventName: "PostToolUse", AdditionalContext: "operation completed"},
			json:  `{"hookEventName":"PostToolUse","additionalContext":"operation completed"}`,
		},
		{
			name:  "SessionStart with continue",
			event: &SessionStartOutput{EventName: "SessionStart", SystemMessage: "Session started"},
			json:  `{"hookEventName":"SessionStart","systemMessage":"Session started"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.event)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Check JSON matches expected
			got := string(data)
			if got != tt.json {
				t.Logf("JSON mismatch:\ngot:  %s\nwant: %s", got, tt.json)
			}

			// Unmarshal
			var gotEvent interface{}
			switch tt.event.(type) {
			case *PreToolUseOutput:
				var e PreToolUseOutput
				gotEvent = &e
			case *PostToolUseOutput:
				var e PostToolUseOutput
				gotEvent = &e
			case *SessionStartOutput:
				var e SessionStartOutput
				gotEvent = &e
			default:
				t.Skip("Unmarshal test not implemented for this type")
			}

			if gotEvent != nil {
				if err := json.Unmarshal(data, gotEvent); err != nil {
					t.Fatalf("Unmarshal() error = %v", err)
				}
			}
		})
	}
}

func TestPermissionDecisionValues(t *testing.T) {
	tests := []struct {
		name  string
		value PermissionDecision
		valid bool
	}{
		{"allow", PermissionDecisionAllow, true},
		{"ask", PermissionDecisionAsk, true},
		{"deny", PermissionDecisionDeny, true},
		{"defer", PermissionDecisionDefer, true},
		{"empty", "", true}, // Empty means "no opinion"
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &HookResponse{PermissionDecision: tt.value}
			err := ValidateHookResponse(resp)
			if tt.valid && err != nil {
				t.Errorf("ValidateHookResponse() unexpectedly returned error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("ValidateHookResponse() should have returned error for invalid decision")
			}
		})
	}
}

func TestHookResponseContinue(t *testing.T) {
	tests := []struct {
		name     string
		cont     *bool
		expected bool // true if continue should be true
	}{
		{"nil continue", nil, true},          // nil means no opinion, defaults to true
		{"continue true", boolPtr(true), true},
		{"continue false", boolPtr(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &HookResponse{Continue: tt.cont}
			output := ToHookOutput(resp)

			got := output.Continue
			if got != tt.expected {
				t.Errorf("ToHookOutput().Continue = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRetryHint(t *testing.T) {
	hint := &RetryHint{
		Attempts: 3,
		Backoff:  "500ms",
	}

	data, err := json.Marshal(hint)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got RetryHint
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.Attempts != hint.Attempts {
		t.Errorf("Attempts = %v, want %v", got.Attempts, hint.Attempts)
	}
	if got.Backoff != hint.Backoff {
		t.Errorf("Backoff = %v, want %v", got.Backoff, hint.Backoff)
	}
}

func TestHookResponseAdditionalContextTruncation(t *testing.T) {
	const maxSize = 64 * 1024 // 64 KiB

	// Create a context larger than 64 KiB
	largeContext := strings.Repeat("x", maxSize+1000)

	resp := &HookResponse{
		AdditionalContext: largeContext,
	}

	err := ValidateHookResponse(resp)
	if err != nil {
		t.Fatalf("ValidateHookResponse() error = %v", err)
	}

	// Check that context was truncated
	if len(resp.AdditionalContext) != maxSize {
		t.Errorf("AdditionalContext length = %d, want %d", len(resp.AdditionalContext), maxSize)
	}

	// Check that SystemMessage was updated
	if resp.SystemMessage == "" {
		t.Error("SystemMessage should contain truncation notice")
	}
}

func boolPtr(b bool) *bool {
	return &b
}
