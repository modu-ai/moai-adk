package hook

import (
	"context"
	"log/slog"

	"github.com/modu-ai/moai-adk/internal/journal"
)

// subagentStartHandler processes SubagentStart events.
// It logs subagent startup for session tracking and replay logging.
type subagentStartHandler struct {
	replay *journal.ReplayWriter
}

// NewSubagentStartHandler creates a new SubagentStart event handler.
// If replay is nil, replay logging is disabled.
func NewSubagentStartHandler(replay *journal.ReplayWriter) Handler {
	return &subagentStartHandler{replay: replay}
}

// EventType returns EventSubagentStart.
func (h *subagentStartHandler) EventType() EventType {
	return EventSubagentStart
}

// Handle processes a SubagentStart event. It logs the subagent startup details
// and records a replay action entry.
func (h *subagentStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("subagent started",
		"session_id", input.SessionID,
		"agent_id", input.AgentID,
		"agent_transcript_path", input.AgentTranscriptPath,
	)

	if h.replay != nil {
		if err := h.replay.LogAction(journal.ActionEntry{
			SessionID: input.SessionID,
			AgentName: input.AgentName,
			AgentID:   input.AgentID,
			Action:    "start",
			Details:   input.AgentTranscriptPath,
		}); err != nil {
			slog.Warn("failed to log replay action", "error", err)
		}
	}

	return &HookOutput{}, nil
}
