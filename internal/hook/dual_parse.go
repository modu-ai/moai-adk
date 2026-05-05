package hook

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrHookProtocolLegacyRejected is returned when strict mode is enabled and the hook
	// used exit codes instead of JSON output.
	ErrHookProtocolLegacyRejected = errors.New("hook protocol: legacy exit code rejected in strict mode")

	// ErrHookInvalidPermissionDecision is returned when an invalid permission decision is encountered.
	ErrHookInvalidPermissionDecision = errors.New("hook protocol: invalid permission decision")

	// ErrHookSpecificOutputMismatch is returned when the hook-specific output does not match the expected event type.
	ErrHookSpecificOutputMismatch = errors.New("hook protocol: hook-specific output mismatch")
)

// HookSpecificOutputMismatch provides details about a hook-specific output type mismatch.
type HookSpecificOutputMismatch struct {
	Expected string
	Actual   string
}

func (e *HookSpecificOutputMismatch) Error() string {
	return fmt.Sprintf("%s: expected %s, got %s", ErrHookSpecificOutputMismatch, e.Expected, e.Actual)
}

// ParseHookOutput implements the dual-parse protocol for hook outputs.
// It attempts JSON unmarshal first (canonical Claude Code HookJSONOutput format),
// then falls back to exit-code synthesis for legacy hooks.
//
// Parameters:
//   - stdout: Raw stdout bytes from the hook process
//   - exitCode: Process exit code
//   - stderr: Stderr content (used for error messages in exit-code mode)
//
// Returns a HookResponse that combines both JSON and exit-code interpretations.
// If stdout contains valid JSON, exitCode is ignored (JSON takes precedence).
// If stdout is empty or invalid JSON, the response is synthesized from exitCode.
//
// @MX:ANCHOR: [AUTO] ParseHookOutput is a shared Validate interface implementation
// @MX:REASON: fan_in >= 3 — called by all hook handlers (SessionStart, PostToolUse, SessionEnd)
func ParseHookOutput(stdout []byte, exitCode int, stderr string) (*HookResponse, error) {
	trimmed := strings.TrimSpace(string(stdout))

	// Try JSON unmarshal first (canonical path)
	if trimmed != "" {
		var resp HookResponse
		if err := json.Unmarshal(stdout, &resp); err == nil {
			// Valid JSON: ignore exit code, return parsed response
			return &resp, nil
		}
		// JSON parse failed: fall through to exit-code synthesis
	}

	// Exit-code synthesis path (legacy fallback)
	return synthesizeFromExitCode(exitCode, stderr)
}

// synthesizeFromExitCode creates a HookResponse based on exit code.
// This implements the legacy protocol for hooks that don't output JSON.
func synthesizeFromExitCode(exitCode int, stderr string) (*HookResponse, error) {
	switch exitCode {
	case 0:
		// Exit code 0: allow, continue (default allow)
		return &HookResponse{}, nil

	case 2:
		// Exit code 2: deny with reason (stderr contains the reason)
		// This is used by TeammateIdle (keep-working) and TaskCompleted (reject completion)
		return &HookResponse{
			PermissionDecision: PermissionDecisionDeny,
			SystemMessage:      stderr,
		}, nil

	default:
		// Non-zero exit code: error (stderr contains the error message)
		return &HookResponse{
			SystemMessage: stderr,
		}, nil
	}
}

// ValidateHookResponse validates the fields of a HookResponse.
// Returns an error if any field contains invalid values.
// Truncates AdditionalContext if it exceeds 64 KiB (65536 bytes).
func ValidateHookResponse(resp *HookResponse) error {
	if resp == nil {
		return nil
	}

	// Validate PermissionDecision
	if resp.PermissionDecision != "" &&
		resp.PermissionDecision != PermissionDecisionAllow &&
		resp.PermissionDecision != PermissionDecisionAsk &&
		resp.PermissionDecision != PermissionDecisionDeny &&
		resp.PermissionDecision != PermissionDecisionDefer {
		return fmt.Errorf("%w: %s", ErrHookInvalidPermissionDecision, resp.PermissionDecision)
	}

	// Truncate AdditionalContext if over 64 KiB
	const maxContextSize = 64 * 1024 // 64 KiB
	if len(resp.AdditionalContext) > maxContextSize {
		originalLen := len(resp.AdditionalContext)
		resp.AdditionalContext = resp.AdditionalContext[:maxContextSize]
		// Add a notice to SystemMessage
		resp.SystemMessage += fmt.Sprintf("\n[Notice: additionalContext truncated from %d to %d bytes]", originalLen, maxContextSize)
	}

	return nil
}

// ToHookOutput converts a HookResponse to the legacy HookOutput format.
// This bridges the new typed HookResponse with the existing HookOutput struct.
func ToHookOutput(resp *HookResponse) *HookOutput {
	if resp == nil {
		return &HookOutput{}
	}

	output := &HookOutput{
		SystemMessage: resp.SystemMessage,
		SuppressOutput: false, // Not mapped in HookResponse
	}

	// Map Continue field
	// nil = no opinion (default true), false = halt, true = continue
	if resp.Continue != nil {
		output.Continue = *resp.Continue
		if !*resp.Continue {
			output.StopReason = resp.SystemMessage
		}
	} else {
		// nil means "no opinion", default to true (continue)
		output.Continue = true
	}

	// Map PermissionDecision to HookSpecificOutput
	if resp.PermissionDecision != "" || resp.AdditionalContext != "" || len(resp.UpdatedInput) > 0 {
		output.HookSpecificOutput = &HookSpecificOutput{
			HookEventName:            "", // Will be set by caller based on event type
			PermissionDecision:       string(resp.PermissionDecision),
			AdditionalContext:        resp.AdditionalContext,
			UpdatedInput:             resp.UpdatedInput,
		}
	}

	// Map Retry field
	if resp.Retry != nil {
		output.Retry = resp.Retry.Attempts > 0
	}

	return output
}

// ToHookResponse converts a legacy HookOutput to the new HookResponse format.
// This bridges the existing HookOutput with the new typed HookResponse struct.
func ToHookResponse(output *HookOutput) *HookResponse {
	if output == nil {
		return &HookResponse{}
	}

	resp := &HookResponse{
		SystemMessage: output.SystemMessage,
	}

	// Map Continue field
	if !output.Continue {
		f := false
		resp.Continue = &f
	}

	// Map HookSpecificOutput fields
	if output.HookSpecificOutput != nil {
		resp.PermissionDecision = PermissionDecision(output.HookSpecificOutput.PermissionDecision)
		resp.AdditionalContext = output.HookSpecificOutput.AdditionalContext
		resp.UpdatedInput = output.HookSpecificOutput.UpdatedInput
	}

	// Map Retry field
	if output.Retry {
		resp.Retry = &RetryHint{Attempts: 1}
	}

	return resp
}
