package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// HookInput represents the input payload from Claude Code to a hook.
// This matches the Python HookPayload structure exactly for compatibility.
//
// JSON Schema:
//
//	{
//	  "session_id": "string",
//	  "event": "string",
//	  "tool_name": "string (optional)",
//	  "tool_input": {},
//	  "tool_output": {}
//	}
type HookInput struct {
	SessionID  string                 `json:"session_id"`
	Event      string                 `json:"event"`
	ToolName   string                 `json:"tool_name,omitempty"`
	ToolInput  map[string]interface{} `json:"tool_input,omitempty"`
	ToolOutput map[string]interface{} `json:"tool_output,omitempty"`
}

// HookEvent represents all possible hook event types
type HookEvent string

const (
	EventSessionStart HookEvent = "session_start"
	EventPreTool      HookEvent = "pre_tool"
	EventPostTool     HookEvent = "post_tool"
	EventSessionEnd   HookEvent = "session_end"
	EventPreCompact   HookEvent = "pre_compact"
	EventStop         HookEvent = "stop"
	EventNotification HookEvent = "notification"
	EventQualityGate  HookEvent = "quality_gate"
	EventCommit       HookEvent = "commit"
	EventPush         HookEvent = "push"
	EventCompact      HookEvent = "compact"
)

// IsValid checks if the event is a valid hook event
func (e HookEvent) IsValid() bool {
	switch e {
	case EventSessionStart, EventPreTool, EventPostTool, EventSessionEnd,
		EventPreCompact, EventStop, EventNotification, EventQualityGate,
		EventCommit, EventPush, EventCompact:
		return true
	default:
		return false
	}
}

// ReadFromStdin reads and parses HookInput from stdin.
// This is the main entry point for hook execution.
func ReadFromStdin() (*HookInput, error) {
	// Read all input from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	// Parse JSON
	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse JSON input: %w", err)
	}

	// Validate required fields
	if input.SessionID == "" {
		return nil, fmt.Errorf("missing required field: session_id")
	}

	if input.Event == "" {
		return nil, fmt.Errorf("missing required field: event")
	}

	// Validate event type
	event := HookEvent(input.Event)
	if !event.IsValid() {
		return nil, fmt.Errorf("invalid event type: %s", input.Event)
	}

	// Initialize maps if nil
	if input.ToolInput == nil {
		input.ToolInput = make(map[string]interface{})
	}
	if input.ToolOutput == nil {
		input.ToolOutput = make(map[string]interface{})
	}

	return &input, nil
}

// GetToolInputString safely gets a string value from tool_input
func (h *HookInput) GetToolInputString(key string) string {
	if val, ok := h.ToolInput[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetToolInputPath gets the file_path from tool_input (common for Write/Edit tools)
func (h *HookInput) GetToolInputPath() string {
	return h.GetToolInputString("file_path")
}

// GetToolInputCommand gets the command from tool_input (for Bash tool)
func (h *HookInput) GetToolInputCommand() string {
	return h.GetToolInputString("command")
}

// GetToolInputMatcher gets the matcher from tool_input
func (h *HookInput) GetToolInputMatcher() string {
	return h.GetToolInputString("matcher")
}
