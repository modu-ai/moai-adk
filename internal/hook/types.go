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

// Decision constants for hook output.
const (
	DecisionAllow = "allow"
	DecisionBlock = "block"
)

// HookInput represents the JSON payload received from Claude Code via stdin.
type HookInput struct {
	SessionID     string          `json:"session_id"`
	CWD           string          `json:"cwd"`
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name,omitempty"`
	ToolInput     json.RawMessage `json:"tool_input,omitempty"`
	ToolOutput    json.RawMessage `json:"tool_output,omitempty"`
	ProjectDir    string          `json:"project_dir,omitempty"`
}

// HookOutput represents the JSON payload written to stdout for Claude Code.
type HookOutput struct {
	Decision string          `json:"decision,omitempty"`
	Reason   string          `json:"reason,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// NewAllowOutput creates a HookOutput with Decision "allow".
func NewAllowOutput() *HookOutput {
	return &HookOutput{Decision: DecisionAllow}
}

// NewAllowOutputWithData creates a HookOutput with Decision "allow" and data.
func NewAllowOutputWithData(data json.RawMessage) *HookOutput {
	return &HookOutput{Decision: DecisionAllow, Data: data}
}

// NewBlockOutput creates a HookOutput with Decision "block" and a reason.
func NewBlockOutput(reason string) *HookOutput {
	return &HookOutput{Decision: DecisionBlock, Reason: reason}
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
