package hook

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/modu-ai/moai-adk-go/internal/config"
)

// DefaultHookTimeout is the default timeout for hook execution (30 seconds).
const DefaultHookTimeout = 30 * time.Second

// EventType represents a Claude Code hook event type.
type EventType string

const (
	// EventSessionStart is triggered when a new Claude Code session begins.
	EventSessionStart EventType = "SessionStart"

	// EventPreToolUse is triggered before a tool is executed.
	EventPreToolUse EventType = "PreToolUse"

	// EventPostToolUse is triggered after a tool has been executed.
	EventPostToolUse EventType = "PostToolUse"

	// EventSessionEnd is triggered when a Claude Code session ends.
	EventSessionEnd EventType = "SessionEnd"

	// EventStop is triggered when Claude Code requests a stop.
	EventStop EventType = "Stop"

	// EventPreCompact is triggered before context compaction.
	EventPreCompact EventType = "PreCompact"
)

// ValidEventTypes returns all valid event types.
func ValidEventTypes() []EventType {
	return []EventType{
		EventSessionStart,
		EventPreToolUse,
		EventPostToolUse,
		EventSessionEnd,
		EventStop,
		EventPreCompact,
	}
}

// IsValidEventType checks if the given event type is valid.
func IsValidEventType(et EventType) bool {
	for _, v := range ValidEventTypes() {
		if v == et {
			return true
		}
	}
	return false
}

// Permission decision constants for PreToolUse hooks (Claude Code protocol).
const (
	DecisionAllow = "allow"
	DecisionDeny  = "deny"
	DecisionAsk   = "ask"
	DecisionBlock = "deny" // Alias for backward compatibility
)

// HookInput represents the JSON payload received from Claude Code via stdin.
type HookInput struct {
	SessionID     string          `json:"session_id,omitempty"`
	CWD           string          `json:"cwd,omitempty"`
	HookEventName string          `json:"hook_event_name,omitempty"`
	ToolName      string          `json:"tool_name,omitempty"`
	ToolInput     json.RawMessage `json:"tool_input,omitempty"`
	ToolOutput    json.RawMessage `json:"tool_output,omitempty"`
	ToolResponse  json.RawMessage `json:"tool_response,omitempty"`
	ProjectDir    string          `json:"project_dir,omitempty"`
}

// HookSpecificOutput represents the hookSpecificOutput field for PreToolUse/PostToolUse.
type HookSpecificOutput struct {
	HookEventName            string `json:"hookEventName,omitempty"`
	PermissionDecision       string `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
	AdditionalContext        string `json:"additionalContext,omitempty"`
}

// HookOutput represents the JSON payload written to stdout for Claude Code.
// The format varies by event type per Claude Code protocol.
type HookOutput struct {
	// For SessionStart/SessionEnd: continue flag and system message
	Continue      bool   `json:"continue,omitempty"`
	SystemMessage string `json:"systemMessage,omitempty"`

	// For PreToolUse/PostToolUse: hook-specific output
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`

	// For silent operations
	SuppressOutput bool `json:"suppressOutput,omitempty"`

	// Legacy fields for backward compatibility (internal use)
	Decision string          `json:"-"`
	Reason   string          `json:"-"`
	Data     json.RawMessage `json:"-"`
}

// NewAllowOutput creates a HookOutput with permissionDecision "allow".
func NewAllowOutput() *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecision: DecisionAllow,
		},
		Decision: DecisionAllow,
	}
}

// NewAllowOutputWithData creates a HookOutput with permissionDecision "allow" and data.
func NewAllowOutputWithData(data json.RawMessage) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecision: DecisionAllow,
		},
		Decision: DecisionAllow,
		Data:     data,
	}
}

// NewDenyOutput creates a HookOutput with permissionDecision "deny" and a reason.
func NewDenyOutput(reason string) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecision:       DecisionDeny,
			PermissionDecisionReason: reason,
		},
		Decision: DecisionDeny,
		Reason:   reason,
	}
}

// NewAskOutput creates a HookOutput with permissionDecision "ask" and a reason.
func NewAskOutput(reason string) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			PermissionDecision:       DecisionAsk,
			PermissionDecisionReason: reason,
		},
		Decision: DecisionAsk,
		Reason:   reason,
	}
}

// NewBlockOutput creates a HookOutput with permissionDecision "deny" (alias for NewDenyOutput).
func NewBlockOutput(reason string) *HookOutput {
	return NewDenyOutput(reason)
}

// NewSuppressOutput creates a HookOutput that suppresses output.
func NewSuppressOutput() *HookOutput {
	return &HookOutput{SuppressOutput: true}
}

// NewSessionOutput creates a HookOutput for SessionStart/SessionEnd events.
func NewSessionOutput(continueSession bool, message string) *HookOutput {
	return &HookOutput{
		Continue:      continueSession,
		SystemMessage: message,
	}
}

// NewPostToolOutput creates a HookOutput with additionalContext for PostToolUse.
func NewPostToolOutput(context string) *HookOutput {
	return &HookOutput{
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: context,
		},
	}
}

// Handler processes a specific hook event type.
type Handler interface {
	// Handle processes the hook input and returns output.
	// ctx carries cancellation and timeout signals.
	Handle(ctx context.Context, input *HookInput) (*HookOutput, error)

	// EventType returns the event type this handler processes.
	EventType() EventType
}

// Registry manages handler registration and event dispatching.
type Registry interface {
	// Register adds a handler to the registry for its declared event type.
	Register(handler Handler)

	// Dispatch sends an event to all registered handlers for the given event type.
	// Handlers are executed sequentially. If any handler returns Decision "block",
	// remaining handlers are skipped and the block result is returned immediately.
	Dispatch(ctx context.Context, event EventType, input *HookInput) (*HookOutput, error)

	// Handlers returns all handlers registered for the given event type.
	Handlers(event EventType) []Handler
}

// Protocol handles JSON communication with Claude Code via stdin/stdout.
type Protocol interface {
	// ReadInput reads and parses JSON from the given reader.
	ReadInput(r io.Reader) (*HookInput, error)

	// WriteOutput serializes the output as JSON to the given writer.
	WriteOutput(w io.Writer, output *HookOutput) error
}

// Contract defines the hook execution contract per ADR-012.
type Contract interface {
	// Validate checks that the execution environment meets contract requirements.
	Validate(ctx context.Context) error

	// Guarantees returns the list of guaranteed execution conditions.
	Guarantees() []string

	// NonGuarantees returns the list of non-guaranteed execution conditions.
	NonGuarantees() []string
}

// ConfigProvider provides read access to application configuration.
// It is satisfied by *config.ConfigManager.
type ConfigProvider interface {
	Get() *config.Config
}
