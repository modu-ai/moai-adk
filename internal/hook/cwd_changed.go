package hook

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

// cwdChangedHandler processes CwdChanged events.
// Fired when the working directory changes during a session.
// Supports CLAUDE_ENV_FILE for persisting environment variables.
// Available since Claude Code v2.1.83+.
type cwdChangedHandler struct{}

// NewCwdChangedHandler creates a new CwdChanged event handler.
func NewCwdChangedHandler() Handler {
	return &cwdChangedHandler{}
}

// EventType returns EventCwdChanged.
func (h *cwdChangedHandler) EventType() EventType {
	return EventCwdChanged
}

// Handle processes a CwdChanged event. It logs the directory change
// and writes relevant environment variables to CLAUDE_ENV_FILE if available.
func (h *cwdChangedHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	newCwd := input.NewCwd
	if newCwd == "" {
		newCwd = input.CWD
	}

	slog.Info("working directory changed",
		"session_id", input.SessionID,
		"old_cwd", input.OldCwd,
		"new_cwd", newCwd,
	)

	// Write project-specific environment to CLAUDE_ENV_FILE if available.
	// This persists env vars into subsequent Bash tool calls.
	if envFile := os.Getenv("CLAUDE_ENV_FILE"); envFile != "" && newCwd != "" {
		h.writeEnvFile(envFile, newCwd)
	}

	return &HookOutput{}, nil
}

// writeEnvFile appends project-specific environment variables to CLAUDE_ENV_FILE.
// Non-blocking: errors are logged but never propagated.
func (h *cwdChangedHandler) writeEnvFile(envFile, cwd string) {
	// Check if the new directory has a .envrc or .env file
	var exports []string

	// If .moai/config exists, export MOAI_PROJECT_DIR
	if _, err := os.Stat(filepath.Join(cwd, ".moai", "config")); err == nil {
		exports = append(exports, "export MOAI_PROJECT_DIR=\""+cwd+"\"")
	}

	if len(exports) == 0 {
		return
	}

	content := ""
	for _, e := range exports {
		content += e + "\n"
	}

	if err := os.WriteFile(envFile, []byte(content), 0o644); err != nil {
		slog.Warn("cwd_changed: failed to write CLAUDE_ENV_FILE",
			"error", err,
			"env_file", envFile,
		)
	} else {
		slog.Debug("cwd_changed: wrote env file",
			"env_file", envFile,
			"exports", len(exports),
		)
	}
}
