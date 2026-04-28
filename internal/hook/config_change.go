package hook

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// configChangeHandler processes ConfigChange events.
// It validates configuration file changes during a session.
type configChangeHandler struct{}

// NewConfigChangeHandler creates a new ConfigChange event handler.
func NewConfigChangeHandler() Handler {
	return &configChangeHandler{}
}

// EventType returns EventConfigChange.
func (h *configChangeHandler) EventType() EventType {
	return EventConfigChange
}

// Handle processes a ConfigChange event. It validates the changed config file
// as YAML and logs the result.
func (h *configChangeHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("config file changed",
		"session_id", input.SessionID,
		"config_file_path", input.ConfigFilePath,
		"configuration_source", input.ConfigurationSource,
	)

	// Validate the config file
	if err := h.validateConfig(input.ConfigFilePath); err != nil {
		return &HookOutput{
			Continue:      false,
			SystemMessage: fmt.Sprintf("Config reload failed: %v", err),
		}, nil
	}

	msg := fmt.Sprintf("%s reloaded successfully", input.ConfigFilePath)
	return &HookOutput{
		Continue:      true, // Explicitly allow continuation
		SystemMessage: msg,
	}, nil
}

// validateConfig checks that a config file is valid YAML.
func (h *configChangeHandler) validateConfig(filePath string) error {
	// Read the config file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Parse as YAML to validate syntax
	var config any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	return nil
}
