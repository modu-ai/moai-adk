// Package trace provides structured JSON trace logging for hook executions.
// Traces are written asynchronously to JSONL files and summarized at session end.
package trace

import "time"

// TraceEntry represents a single hook execution trace record.
// Fields follow the SPEC-OBSERVE-001 JSON schema.
type TraceEntry struct {
	// Timestamp when the handler was invoked.
	Timestamp time.Time `json:"ts"`
	// Event is the Claude Code hook event type (e.g. "PreToolUse").
	Event string `json:"event"`
	// Handler is the Go type name of the handler that processed the event.
	Handler string `json:"handler"`
	// Tool is the tool name for tool-related events (PreToolUse, PostToolUse).
	Tool string `json:"tool,omitempty"`
	// DurationMs is the handler execution time in milliseconds.
	DurationMs int64 `json:"duration_ms"`
	// Decision is the outcome of the handler (e.g. "allow", "deny", "block").
	Decision string `json:"decision,omitempty"`
	// Reason is the explanation for deny/block decisions.
	Reason string `json:"reason,omitempty"`
	// Error is the error message if the handler returned an error.
	Error string `json:"error,omitempty"`
	// SessionID is the Claude Code session identifier.
	SessionID string `json:"session_id"`
}
