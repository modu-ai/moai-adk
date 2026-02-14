package hook

import (
	"context"
	"log/slog"
)

// subagentStartHandler processes SubagentStart events.
// It logs subagent startup for session tracking.
type subagentStartHandler struct{}

// NewSubagentStartHandler creates a new SubagentStart event handler.
func NewSubagentStartHandler() Handler {
	return &subagentStartHandler{}
}

// EventType returns EventSubagentStart.
func (h *subagentStartHandler) EventType() EventType {
	return EventSubagentStart
}

// Handle processes a SubagentStart event. It logs the subagent startup details.
func (h *subagentStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("subagent started",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_transcript_path", input.AgentTranscriptPath,
	)
	return &HookOutput{}, nil
}
