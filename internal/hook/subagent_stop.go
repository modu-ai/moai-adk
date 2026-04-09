package hook

import (
	"context"
	"log/slog"
)

// subagentStopHandler processes SubagentStop events.
// It logs subagent termination for session tracking.
type subagentStopHandler struct{}

// NewSubagentStopHandler creates a new SubagentStop event handler.
func NewSubagentStopHandler() Handler {
	return &subagentStopHandler{}
}

// EventType returns EventSubagentStop.
func (h *subagentStopHandler) EventType() EventType {
	return EventSubagentStop
}

// Handle processes a SubagentStop event. It logs subagent termination details.
func (h *subagentStopHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("subagent stopped",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_name", input.AgentName,
		"agent_transcript_path", input.AgentTranscriptPath,
	)
	return &HookOutput{}, nil
}
