package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/modu-ai/moai-adk/internal/hook/memo"
)

// postCompactHandler processes PostCompact events.
// It logs compaction completion and restores session context via the session memo.
type postCompactHandler struct{}

// NewPostCompactHandler creates a new PostCompact event handler.
func NewPostCompactHandler() Handler {
	return &postCompactHandler{}
}

// EventType returns EventPostCompact.
func (h *postCompactHandler) EventType() EventType {
	return EventPostCompact
}

// Handle processes a PostCompact event. It reads the session memo written
// by the pre-compact handler and injects it as a SystemMessage so Claude
// can recover context. Errors are non-blocking.
func (h *postCompactHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("post-compact recovery",
		"session_id", input.SessionID,
	)

	projectDir := resolveProjectDir(input)

	data := map[string]any{
		"session_id": input.SessionID,
		"status":     "recovered",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Warn("post-compact: failed to marshal data", "error", err)
		return &HookOutput{}, nil
	}

	// Attempt to restore context from the session memo (non-blocking).
	if projectDir != "" {
		content, readErr := memo.Read(projectDir, 2200)
		if readErr != nil {
			slog.Warn("post-compact: failed to read session memo",
				"error", readErr,
				"project_dir", projectDir,
			)
			// Continue with empty system message.
		} else if content != "" {
			slog.Info("post-compact: session memo restored",
				"project_dir", projectDir,
			)
			return &HookOutput{
				SystemMessage: fmt.Sprintf("[Session Memo - Restored after context compaction]\n\n%s", content),
				Data:          jsonData,
			}, nil
		}
	}

	return &HookOutput{Data: jsonData}, nil
}
