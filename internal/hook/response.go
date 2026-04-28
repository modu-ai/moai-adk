package hook

import (
	"encoding/json"
)

// HookResponse represents the canonical Claude Code HookJSONOutput schema.
// This is the typed equivalent of the JSON output that hooks can return.
// The dual-protocol parser (ParseHookOutput) attempts to unmarshal stdout
// into this struct before falling back to exit-code synthesis.
type HookResponse struct {
	// AdditionalContext provides extra context for the current operation.
	// Used by: PreToolUse, PostToolUse, UserPromptSubmit
	AdditionalContext string `json:"additionalContext,omitempty"`

	// PermissionDecision controls tool execution permission.
	// Used by: PreToolUse, PermissionRequest.
	// Values: "", "allow", "ask", "deny". Empty string means "no opinion" (defer to other hooks).
	PermissionDecision PermissionDecision `json:"permissionDecision,omitempty"`

	// UpdatedInput contains modified input for the current operation.
	// Used by: PreToolUse (modifies tool input), UserPromptSubmit (modifies user prompt).
	UpdatedInput json.RawMessage `json:"updatedInput,omitempty"`

	// SystemMessage is a warning message shown to the user.
	SystemMessage string `json:"systemMessage,omitempty"`

	// Continue controls whether Claude continues processing.
	// nil = no opinion (default true), false = halt processing.
	// Used by: SessionStart, SessionEnd.
	Continue *bool `json:"continue,omitempty"`

	// WatchPaths declares file paths to watch for changes.
	// Used by: FileChanged, InstructionsLoaded.
	WatchPaths []string `json:"watchPaths,omitempty"`

	// Retry signals that the model should retry a denied operation.
	// Used by: PermissionDenied (v2.1.89+).
	Retry *RetryHint `json:"retry,omitempty"`

	// HookSpecificOutput contains event-specific output data.
	// This field preserves the full event-specific JSON structure.
	HookSpecificOutput json.RawMessage `json:"hookSpecificOutput,omitempty"`
}

// PermissionDecision is a string type for permission decision values.
type PermissionDecision string

const (
	// PermissionDecisionAllow allows the operation to proceed.
	PermissionDecisionAllow PermissionDecision = "allow"

	// PermissionDecisionAsk prompts the user for confirmation.
	PermissionDecisionAsk PermissionDecision = "ask"

	// PermissionDecisionDeny blocks the operation.
	PermissionDecisionDeny PermissionDecision = "deny"

	// PermissionDecisionDefer pauses headless sessions; resume with --resume (v2.1.89+).
	PermissionDecisionDefer PermissionDecision = "defer"
)

// RetryHint provides retry configuration for denied operations.
type RetryHint struct {
	// Attempts is the number of retry attempts allowed.
	Attempts int `json:"attempts,omitempty"`

	// Backoff is the duration to wait before retrying (e.g., "500ms", "1s").
	Backoff string `json:"backoff,omitempty"`
}

// Event-specific variant types for HookResponse.HookSpecificOutput.
// Each type implements HookEventName() string to identify itself.
// These types are named with "Output" suffix to avoid conflict with EventType constants.

// PreToolUseOutput represents the PreToolUse event-specific output.
type PreToolUseOutput struct {
	EventName               string          `json:"hookEventName,omitempty"`
	PermissionDecision      PermissionDecision `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string          `json:"permissionDecisionReason,omitempty"`
	AdditionalContext       string          `json:"additionalContext,omitempty"`
	UpdatedInput            json.RawMessage `json:"updatedInput,omitempty"`
}

func (e *PreToolUseOutput) HookEventName() string { return "PreToolUse" }

// PostToolUseOutput represents the PostToolUse event-specific output.
type PostToolUseOutput struct {
	EventName          string `json:"hookEventName,omitempty"`
	AdditionalContext  string `json:"additionalContext,omitempty"`
	UpdatedMCPToolOutput string `json:"updatedMCPToolOutput,omitempty"`
}

func (e *PostToolUseOutput) HookEventName() string { return "PostToolUse" }

// SessionStartOutput represents the SessionStart event-specific output.
type SessionStartOutput struct {
	EventName    string `json:"hookEventName,omitempty"`
	SystemMessage string `json:"systemMessage,omitempty"`
	Continue     *bool  `json:"continue,omitempty"`
}

func (e *SessionStartOutput) HookEventName() string { return "SessionStart" }

// SessionEndOutput represents the SessionEnd event-specific output.
type SessionEndOutput struct {
	EventName    string `json:"hookEventName,omitempty"`
	SystemMessage string `json:"systemMessage,omitempty"`
	Continue     *bool  `json:"continue,omitempty"`
}

func (e *SessionEndOutput) HookEventName() string { return "SessionEnd" }

// StopOutput represents the Stop event-specific output.
type StopOutput struct {
	EventName string `json:"hookEventName,omitempty"`
	Decision  string `json:"decision,omitempty"` // "block" to prevent stopping
	Reason    string `json:"reason,omitempty"`
}

func (e *StopOutput) HookEventName() string { return "Stop" }

// SubagentStopOutput represents the SubagentStop event-specific output.
type SubagentStopOutput struct {
	EventName string `json:"hookEventName,omitempty"`
	Decision  string `json:"decision,omitempty"` // "block" to prevent stopping
	Reason    string `json:"reason,omitempty"`
}

func (e *SubagentStopOutput) HookEventName() string { return "SubagentStop" }

// PreCompactOutput represents the PreCompact event-specific output.
type PreCompactOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *PreCompactOutput) HookEventName() string { return "PreCompact" }

// PostCompactOutput represents the PostCompact event-specific output.
type PostCompactOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *PostCompactOutput) HookEventName() string { return "PostCompact" }

// PostToolUseFailureOutput represents the PostToolUseFailure event-specific output.
type PostToolUseFailureOutput struct {
	EventName string `json:"hookEventName,omitempty"`
	Decision  string `json:"decision,omitempty"` // "block" to halt after failure
	Reason    string `json:"reason,omitempty"`
}

func (e *PostToolUseFailureOutput) HookEventName() string { return "PostToolUseFailure" }

// NotificationOutput represents the Notification event-specific output.
type NotificationOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *NotificationOutput) HookEventName() string { return "Notification" }

// UserPromptSubmitOutput represents the UserPromptSubmit event-specific output.
type UserPromptSubmitOutput struct {
	EventName         string          `json:"hookEventName,omitempty"`
	AdditionalContext string          `json:"additionalContext,omitempty"`
	UpdatedInput      json.RawMessage `json:"updatedInput,omitempty"`
	SessionTitle      string          `json:"sessionTitle,omitempty"`
}

func (e *UserPromptSubmitOutput) HookEventName() string { return "UserPromptSubmit" }

// PermissionRequestOutput represents the PermissionRequest event-specific output.
type PermissionRequestOutput struct {
	EventName               string          `json:"hookEventName,omitempty"`
	PermissionDecision      PermissionDecision `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string          `json:"permissionDecisionReason,omitempty"`
}

func (e *PermissionRequestOutput) HookEventName() string { return "PermissionRequest" }

// PermissionDeniedOutput represents the PermissionDenied event-specific output.
type PermissionDeniedOutput struct {
	EventName string     `json:"hookEventName,omitempty"`
	Retry     *RetryHint `json:"retry,omitempty"`
}

func (e *PermissionDeniedOutput) HookEventName() string { return "PermissionDenied" }

// ConfigChangeOutput represents the ConfigChange event-specific output.
type ConfigChangeOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *ConfigChangeOutput) HookEventName() string { return "ConfigChange" }

// InstructionsLoadedOutput represents the InstructionsLoaded event-specific output.
type InstructionsLoadedOutput struct {
	EventName string   `json:"hookEventName,omitempty"`
	WatchPaths []string `json:"watchPaths,omitempty"`
}

func (e *InstructionsLoadedOutput) HookEventName() string { return "InstructionsLoaded" }

// FileChangedOutput represents the FileChanged event-specific output.
type FileChangedOutput struct {
	EventName string   `json:"hookEventName,omitempty"`
	WatchPaths []string `json:"watchPaths,omitempty"`
}

func (e *FileChangedOutput) HookEventName() string { return "FileChanged" }

// TeammateIdleOutput represents the TeammateIdle event-specific output.
type TeammateIdleOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *TeammateIdleOutput) HookEventName() string { return "TeammateIdle" }

// TaskCompletedOutput represents the TaskCompleted event-specific output.
type TaskCompletedOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *TaskCompletedOutput) HookEventName() string { return "TaskCompleted" }

// SubagentStartOutput represents the SubagentStart event-specific output.
type SubagentStartOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *SubagentStartOutput) HookEventName() string { return "SubagentStart" }

// WorktreeCreateOutput represents the WorktreeCreate event-specific output.
type WorktreeCreateOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *WorktreeCreateOutput) HookEventName() string { return "WorktreeCreate" }

// WorktreeRemoveOutput represents the WorktreeRemove event-specific output.
type WorktreeRemoveOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *WorktreeRemoveOutput) HookEventName() string { return "WorktreeRemove" }

// CwdChangedOutput represents the CwdChanged event-specific output.
type CwdChangedOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *CwdChangedOutput) HookEventName() string { return "CwdChanged" }

// SetupOutput represents the Setup event-specific output.
type SetupOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *SetupOutput) HookEventName() string { return "Setup" }

// ElicitationOutput represents the Elicitation event-specific output.
type ElicitationOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *ElicitationOutput) HookEventName() string { return "Elicitation" }

// ElicitationResultOutput represents the ElicitationResult event-specific output.
type ElicitationResultOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *ElicitationResultOutput) HookEventName() string { return "ElicitationResult" }

// TaskCreatedOutput represents the TaskCreated event-specific output.
type TaskCreatedOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *TaskCreatedOutput) HookEventName() string { return "TaskCreated" }

// StopFailureOutput represents the StopFailure event-specific output.
type StopFailureOutput struct {
	EventName string `json:"hookEventName,omitempty"`
}

func (e *StopFailureOutput) HookEventName() string { return "StopFailure" }
