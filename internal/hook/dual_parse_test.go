package hook

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestParseHookOutput_ValidJSON(t *testing.T) {
	stdout := `{"permissionDecision":"allow","additionalContext":"test context"}`
	exitCode := 1 // Should be ignored when JSON is valid
	stderr := "some error"

	resp, err := ParseHookOutput([]byte(stdout), exitCode, stderr)
	if err != nil {
		t.Fatalf("ParseHookOutput() error = %v", err)
	}

	if resp.PermissionDecision != PermissionDecisionAllow {
		t.Errorf("PermissionDecision = %v, want %v", resp.PermissionDecision, PermissionDecisionAllow)
	}
	if resp.AdditionalContext != "test context" {
		t.Errorf("AdditionalContext = %v, want %v", resp.AdditionalContext, "test context")
	}
}

func TestParseHookOutput_EmptyStdout_ExitCode0(t *testing.T) {
	resp, err := ParseHookOutput([]byte(""), 0, "")
	if err != nil {
		t.Fatalf("ParseHookOutput() error = %v", err)
	}

	// Exit code 0: allow, continue (default)
	if resp.PermissionDecision != "" {
		t.Errorf("PermissionDecision should be empty (no opinion), got %v", resp.PermissionDecision)
	}
	if resp.SystemMessage != "" {
		t.Errorf("SystemMessage should be empty, got %v", resp.SystemMessage)
	}
}

func TestParseHookOutput_EmptyStdout_ExitCode2(t *testing.T) {
	stderr := "Access denied"
	resp, err := ParseHookOutput([]byte(""), 2, stderr)
	if err != nil {
		t.Fatalf("ParseHookOutput() error = %v", err)
	}

	// Exit code 2: deny with reason
	if resp.PermissionDecision != PermissionDecisionDeny {
		t.Errorf("PermissionDecision = %v, want %v", resp.PermissionDecision, PermissionDecisionDeny)
	}
	if resp.SystemMessage != stderr {
		t.Errorf("SystemMessage = %v, want %v", resp.SystemMessage, stderr)
	}
}

func TestParseHookOutput_EmptyStdout_NonZeroExitCode(t *testing.T) {
	stderr := "Hook execution failed"
	resp, err := ParseHookOutput([]byte(""), 1, stderr)
	if err != nil {
		t.Fatalf("ParseHookOutput() error = %v", err)
	}

	// Non-zero exit code: error message
	if resp.SystemMessage != stderr {
		t.Errorf("SystemMessage = %v, want %v", resp.SystemMessage, stderr)
	}
}

func TestParseHookOutput_MalformedJSON(t *testing.T) {
	stdout := `{invalid json}`
	exitCode := 1
	stderr := "Parse error"

	resp, err := ParseHookOutput([]byte(stdout), exitCode, stderr)
	if err != nil {
		t.Fatalf("ParseHookOutput() error = %v", err)
	}

	// Should fall back to exit code synthesis
	if resp.SystemMessage != stderr {
		t.Errorf("SystemMessage = %v, want %v", resp.SystemMessage, stderr)
	}
}

func TestParseHookOutput_WhitespaceOnly(t *testing.T) {
	stdout := "   \n\t  "
	resp, err := ParseHookOutput([]byte(stdout), 0, "")
	if err != nil {
		t.Fatalf("ParseHookOutput() error = %v", err)
	}

	// Should be treated as empty stdout
	if resp.PermissionDecision != "" {
		t.Errorf("PermissionDecision should be empty (no opinion), got %v", resp.PermissionDecision)
	}
}

func TestValidateHookResponse_ValidDecisions(t *testing.T) {
	validDecisions := []PermissionDecision{
		PermissionDecisionAllow,
		PermissionDecisionAsk,
		PermissionDecisionDeny,
		PermissionDecisionDefer,
		"", // Empty means "no opinion"
	}

	for _, decision := range validDecisions {
		t.Run(string(decision), func(t *testing.T) {
			resp := &HookResponse{PermissionDecision: decision}
			err := ValidateHookResponse(resp)
			if err != nil {
				t.Errorf("ValidateHookResponse() unexpectedly returned error: %v", err)
			}
		})
	}
}

func TestValidateHookResponse_InvalidDecision(t *testing.T) {
	resp := &HookResponse{PermissionDecision: "invalid"}
	err := ValidateHookResponse(resp)

	if err == nil {
		t.Error("ValidateHookResponse() should have returned error for invalid decision")
	}
	if !errors.Is(err, ErrHookInvalidPermissionDecision) {
		t.Errorf("Error should be ErrHookInvalidPermissionDecision, got %v", err)
	}
}

func TestValidateHookResponse_ContextTruncation(t *testing.T) {
	const maxSize = 64 * 1024

	tests := []struct {
		name          string
		contextSize   int
		expectTrunc   bool
		expectNotice  bool
	}{
		{
			name:        "under limit",
			contextSize: maxSize - 100,
			expectTrunc: false,
		},
		{
			name:        "at limit",
			contextSize: maxSize,
			expectTrunc: false,
		},
		{
			name:        "over limit",
			contextSize: maxSize + 1000,
			expectTrunc: true,
			expectNotice: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &HookResponse{
				AdditionalContext: strings.Repeat("x", tt.contextSize),
			}

			err := ValidateHookResponse(resp)
			if err != nil {
				t.Fatalf("ValidateHookResponse() error = %v", err)
			}

			if tt.expectTrunc && len(resp.AdditionalContext) != maxSize {
				t.Errorf("AdditionalContext not truncated to %d bytes, got %d", maxSize, len(resp.AdditionalContext))
			}
			if !tt.expectTrunc && len(resp.AdditionalContext) != tt.contextSize {
				t.Errorf("AdditionalContext was unexpectedly truncated")
			}
			if tt.expectNotice && resp.SystemMessage == "" {
				t.Error("SystemMessage should contain truncation notice")
			}
			if !tt.expectNotice && resp.SystemMessage != "" {
				t.Errorf("SystemMessage should not contain notice, got %v", resp.SystemMessage)
			}
		})
	}
}

func TestToHookOutput_AllowDecision(t *testing.T) {
	resp := &HookResponse{
		PermissionDecision: PermissionDecisionAllow,
		AdditionalContext:  "test context",
		SystemMessage:      "test message",
	}

	output := ToHookOutput(resp)

	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != string(PermissionDecisionAllow) {
		t.Errorf("PermissionDecision = %v, want %v", output.HookSpecificOutput.PermissionDecision, PermissionDecisionAllow)
	}
	if output.HookSpecificOutput.AdditionalContext != "test context" {
		t.Errorf("AdditionalContext = %v, want %v", output.HookSpecificOutput.AdditionalContext, "test context")
	}
	if output.SystemMessage != "test message" {
		t.Errorf("SystemMessage = %v, want %v", output.SystemMessage, "test message")
	}
}

func TestToHookOutput_ContinueFalse(t *testing.T) {
	f := false
	resp := &HookResponse{
		Continue:       &f,
		SystemMessage:  "Stopping execution",
	}

	output := ToHookOutput(resp)

	if output.Continue != false {
		t.Errorf("Continue = %v, want false", output.Continue)
	}
	if output.StopReason != "Stopping execution" {
		t.Errorf("StopReason = %v, want %v", output.StopReason, "Stopping execution")
	}
}

func TestToHookOutput_ContinueNil(t *testing.T) {
	resp := &HookResponse{
		Continue: nil, // No opinion
	}

	output := ToHookOutput(resp)

	if output.Continue != true { // nil means "no opinion", defaults to true (continue)
		t.Errorf("Continue should default to true when nil, got %v", output.Continue)
	}
}

func TestToHookResponse_RoundTrip(t *testing.T) {
	original := &HookOutput{
		SystemMessage: "test message",
		Continue:      true,
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecision:       string(PermissionDecisionAllow),
			AdditionalContext:        "context",
			PermissionDecisionReason: "reason",
		},
		Retry: true,
	}

	resp := ToHookResponse(original)
	output := ToHookOutput(resp)

	if output.SystemMessage != original.SystemMessage {
		t.Errorf("SystemMessage = %v, want %v", output.SystemMessage, original.SystemMessage)
	}
	if output.Continue != original.Continue {
		t.Errorf("Continue = %v, want %v", output.Continue, original.Continue)
	}
	if output.Retry != original.Retry {
		t.Errorf("Retry = %v, want %v", output.Retry, original.Retry)
	}
	if output.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput should not be nil")
	}
	if output.HookSpecificOutput.PermissionDecision != original.HookSpecificOutput.PermissionDecision {
		t.Errorf("PermissionDecision = %v, want %v", output.HookSpecificOutput.PermissionDecision, original.HookSpecificOutput.PermissionDecision)
	}
}

func TestSynthesizeFromExitCode_AllCases(t *testing.T) {
	tests := []struct {
		name          string
		exitCode      int
		stderr        string
		expectDecision PermissionDecision
		expectMessage string
	}{
		{
			name:          "exit 0 - allow",
			exitCode:      0,
			stderr:        "",
			expectDecision: "",
			expectMessage: "",
		},
		{
			name:          "exit 2 - deny",
			exitCode:      2,
			stderr:        "Access denied",
			expectDecision: PermissionDecisionDeny,
			expectMessage: "Access denied",
		},
		{
			name:          "exit 1 - error",
			exitCode:      1,
			stderr:        "Hook failed",
			expectDecision: "",
			expectMessage: "Hook failed",
		},
		{
			name:          "exit 3 - error",
			exitCode:      3,
			stderr:        "Unknown error",
			expectDecision: "",
			expectMessage: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := synthesizeFromExitCode(tt.exitCode, tt.stderr)
			if err != nil {
				t.Fatalf("synthesizeFromExitCode() error = %v", err)
			}

			if resp.PermissionDecision != tt.expectDecision {
				t.Errorf("PermissionDecision = %v, want %v", resp.PermissionDecision, tt.expectDecision)
			}
			if resp.SystemMessage != tt.expectMessage {
				t.Errorf("SystemMessage = %v, want %v", resp.SystemMessage, tt.expectMessage)
			}
		})
	}
}

func TestHookResponse_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		response *HookResponse
		expect   string
	}{
		{
			name:     "empty response",
			response: &HookResponse{},
			expect:   `{}`,
		},
		{
			name: "with permission decision",
			response: &HookResponse{
				PermissionDecision: PermissionDecisionAllow,
			},
			expect: `{"permissionDecision":"allow"}`,
		},
		{
			name: "with all fields",
			response: &HookResponse{
				PermissionDecision: PermissionDecisionDeny,
				AdditionalContext:  "test context",
				SystemMessage:      "test message",
				Retry: &RetryHint{
					Attempts: 3,
					Backoff:  "500ms",
				},
			},
			expect: `{"permissionDecision":"deny","additionalContext":"test context","systemMessage":"test message","retry":{"attempts":3,"backoff":"500ms"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			got := strings.TrimSpace(string(data))
			if got != tt.expect {
				t.Logf("JSON:\ngot:  %s\nwant: %s", got, tt.expect)
			}
		})
	}
}
