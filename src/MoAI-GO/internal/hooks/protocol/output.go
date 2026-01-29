package protocol

import (
	"encoding/json"
	"os"
)

// HookResponse represents the output from a hook back to Claude Code.
// This matches the Python HookResult structure exactly for compatibility.
//
// JSON Schema:
//
//	{
//	  "system_message": "string (optional)",
//	  "continue_execution": true,
//	  "context_files": ["string"],
//	  "hook_specific_output": {},
//	  "block_execution": false,
//	  "suppressOutput": true,
//	  "error": "string (optional)"
//	}
type HookResponse struct {
	SystemMessage      string         `json:"systemMessage,omitempty"`
	ContinueExecution  bool           `json:"continue"`
	ContextFiles       []string       `json:"context_files,omitempty"`
	HookSpecificOutput map[string]any `json:"hookSpecificOutput,omitempty"`
	BlockExecution     bool           `json:"block_execution,omitempty"`
	SuppressOutput     bool           `json:"suppressOutput,omitempty"`
	Error              string         `json:"error,omitempty"`
}

// SecurityDecision represents the security decision for pre-tool hooks
type SecurityDecision string

const (
	DecisionAllow SecurityDecision = "allow"
	DecisionBlock SecurityDecision = "block"
	DecisionWarn  SecurityDecision = "warn"
)

// NewSecurityResponse creates a response for security guard hooks
func NewSecurityResponse(decision SecurityDecision, reason string) *HookResponse {
	resp := &HookResponse{
		ContinueExecution: decision != DecisionBlock,
		SuppressOutput:    decision == DecisionAllow,
	}

	if decision != DecisionAllow {
		resp.HookSpecificOutput = map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       string(decision),
			"permissionDecisionReason": reason,
		}
	}

	if decision == DecisionBlock {
		resp.BlockExecution = true
	}

	return resp
}

// NewMessageResponse creates a response with a system message
func NewMessageResponse(message string, suppress bool) *HookResponse {
	return &HookResponse{
		SystemMessage:     message,
		ContinueExecution: true,
		SuppressOutput:    suppress,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err string) *HookResponse {
	return &HookResponse{
		Error:             err,
		ContinueExecution: true, // Continue on error for graceful degradation
	}
}

// WriteToStdout writes the HookResponse to stdout as JSON
func (h *HookResponse) WriteToStdout() error {
	// Remove empty values for cleaner output
	output := h.clean()

	// Marshal to JSON
	data, err := json.Marshal(output)
	if err != nil {
		return err
	}

	// Write to stdout
	_, err = os.Stdout.Write(data)
	return err
}

// clean removes empty/nil values for cleaner JSON output
// Matches Python behavior: "Remove empty/None values to keep output clean"
func (h *HookResponse) clean() *HookResponse {
	cleaned := &HookResponse{
		ContinueExecution: h.ContinueExecution,
	}

	// Only include non-empty fields
	if h.SystemMessage != "" {
		cleaned.SystemMessage = h.SystemMessage
	}

	if len(h.ContextFiles) > 0 {
		cleaned.ContextFiles = h.ContextFiles
	}

	if len(h.HookSpecificOutput) > 0 {
		cleaned.HookSpecificOutput = h.HookSpecificOutput
	}

	if h.BlockExecution {
		cleaned.BlockExecution = true
	}

	if h.SuppressOutput {
		cleaned.SuppressOutput = true
	}

	if h.Error != "" {
		cleaned.Error = h.Error
	}

	return cleaned
}

// WithContext adds context files to the response
func (h *HookResponse) WithContext(files []string) *HookResponse {
	h.ContextFiles = files
	return h
}

// WithHookOutput adds hook-specific output
func (h *HookResponse) WithHookOutput(key string, value any) *HookResponse {
	if h.HookSpecificOutput == nil {
		h.HookSpecificOutput = make(map[string]any)
	}
	h.HookSpecificOutput[key] = value
	return h
}
