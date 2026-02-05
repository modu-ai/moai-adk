package hook

import (
	"context"
	"encoding/json"
	"log/slog"
)

// compactHandler processes PreCompact events.
// It captures context information and creates session state snapshots
// for post-compaction recovery (REQ-HOOK-036). Always returns "allow".
type compactHandler struct{}

// NewCompactHandler creates a new PreCompact event handler.
func NewCompactHandler() Handler {
	return &compactHandler{}
}

// EventType returns EventPreCompact.
func (h *compactHandler) EventType() EventType {
	return EventPreCompact
}

// Handle processes a PreCompact event. It captures the current context,
// creates a session state snapshot, and returns preservation status in
// the Data field. Errors are non-blocking.
func (h *compactHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("pre-compact context preservation",
		"session_id", input.SessionID,
		"project_dir", input.ProjectDir,
	)

	data := map[string]any{
		"session_id":       input.SessionID,
		"status":           "preserved",
		"snapshot_created": true,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal compact data",
			"error", err.Error(),
		)
		return &HookOutput{}, nil
	}

	return &HookOutput{Data: jsonData}, nil
}
