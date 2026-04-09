package hook

import (
	"context"
	"log/slog"
)

// configChangeHandler processes ConfigChange events.
// Fired when configuration files change during a session.
// Available since Claude Code v2.1.49+.
type configChangeHandler struct{}

// NewConfigChangeHandler creates a new ConfigChange event handler.
func NewConfigChangeHandler() Handler {
	return &configChangeHandler{}
}

// EventType returns EventConfigChange.
func (h *configChangeHandler) EventType() EventType {
	return EventConfigChange
}

// Handle processes a ConfigChange event. It logs the changed configuration file path.
func (h *configChangeHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("config file changed",
		"session_id", input.SessionID,
		"config_file_path", input.ConfigFilePath,
	)
	return &HookOutput{}, nil
}
