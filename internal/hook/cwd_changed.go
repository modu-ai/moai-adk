// Resolution: KEEP — CLAUDE_ENV_FILE write on working directory change.
package hook

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
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

	// t74: keep the registry entry's CWD in step with the session, so anchor
	// detection sees sessions that entered a worktree mid-session.
	relocateSessionCwd(input, newCwd)

	// Write project-specific environment to CLAUDE_ENV_FILE if available.
	// This persists env vars into subsequent Bash tool calls.
	if envFile := os.Getenv(config.EnvClaudeEnvFile); envFile != "" && newCwd != "" {
		h.writeEnvFile(envFile, newCwd)
	}

	return &HookOutput{}, nil
}

// writeEnvFile appends project-specific environment variables to CLAUDE_ENV_FILE.
//
// t236 / issue #1640 (2026-09-02): the body lives in the package-level
// stampProjectDirEnv so the PostToolUse EnterWorktree/ExitWorktree branch can
// reuse the same stamp — Claude Code emits no CwdChanged for those tool-based
// tree moves, so this handler is no longer the only legitimate caller.
func (h *cwdChangedHandler) writeEnvFile(envFile, cwd string) {
	stampProjectDirEnv(envFile, cwd)
}

// @MX:NOTE: [AUTO] t236/#1640 — sole producer of the MOAI_PROJECT_DIR export;
// now called from BOTH CwdChanged and the PostToolUse worktree-move branch,
// because Claude Code emits no CwdChanged for EnterWorktree/ExitWorktree.
// No Go code consumes MOAI_PROJECT_DIR yet (verified 2026-09-02) — the stamp
// serves user scripts and future consumers only.
// stampProjectDirEnv appends `export MOAI_PROJECT_DIR="<cwd>"` to envFile when
// cwd holds a .moai/config directory. Append-only (CLAUDE_ENV_FILE may already
// hold user/session content — never a truncating write), idempotent (an export
// line already present is not appended again), non-blocking (errors are logged
// but never propagated).
func stampProjectDirEnv(envFile, cwd string) {
	// Check if the new directory has a .envrc or .env file
	var exports []string

	// If .moai/config exists, export MOAI_PROJECT_DIR
	if _, err := os.Stat(filepath.Join(cwd, ".moai", "config")); err == nil {
		exports = append(exports, "export MOAI_PROJECT_DIR=\""+cwd+"\"")
	}

	if len(exports) == 0 {
		return
	}

	// Idempotency: skip export lines that already exist verbatim in the file.
	existing, readErr := os.ReadFile(envFile)
	if readErr != nil && !os.IsNotExist(readErr) {
		slog.Warn("cwd_changed: failed to read CLAUDE_ENV_FILE",
			"error", readErr,
			"env_file", envFile,
		)
		return
	}
	existingLines := make(map[string]bool)
	for _, line := range strings.Split(string(existing), "\n") {
		existingLines[strings.TrimSpace(line)] = true
	}

	content := ""
	pending := 0
	for _, e := range exports {
		if existingLines[e] {
			continue
		}
		content += e + "\n"
		pending++
	}
	if pending == 0 {
		return
	}

	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		slog.Warn("cwd_changed: failed to open CLAUDE_ENV_FILE for append",
			"error", err,
			"env_file", envFile,
		)
		return
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(content); err != nil {
		slog.Warn("cwd_changed: failed to append to CLAUDE_ENV_FILE",
			"error", err,
			"env_file", envFile,
		)
	} else {
		slog.Debug("cwd_changed: appended env file",
			"env_file", envFile,
			"exports", pending,
		)
	}
}
