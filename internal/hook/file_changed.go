package hook

import (
	"context"
	"log/slog"
)

// fileChangedHandler processes FileChanged events.
// Fired when a file is changed externally during a session.
// Available since Claude Code v2.1.83+.
type fileChangedHandler struct{}

// NewFileChangedHandler creates a new FileChanged event handler.
func NewFileChangedHandler() Handler {
	return &fileChangedHandler{}
}

// EventType returns EventFileChanged.
func (h *fileChangedHandler) EventType() EventType {
	return EventFileChanged
}

// Handle processes a FileChanged event. It logs the changed file path
// for observability. Future enhancements may trigger re-validation or
// MX tag checks on the changed file.
func (h *fileChangedHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("file changed externally",
		"session_id", input.SessionID,
		"file_path", input.FilePath,
	)
	return &HookOutput{}, nil
}
