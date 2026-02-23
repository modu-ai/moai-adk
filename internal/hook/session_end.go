package hook

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// teamConfig is the minimal structure read from ~/.claude/teams/*/config.json.
type teamConfig struct {
	LeadSessionID string `json:"leadSessionId"`
}

// sessionEndHandler processes SessionEnd events.
// It persists session metrics, cleans up temporary resources, and optionally
// submits ranking data (REQ-HOOK-034). Always returns "allow".
type sessionEndHandler struct{}

// NewSessionEndHandler creates a new SessionEnd event handler.
func NewSessionEndHandler() Handler {
	return &sessionEndHandler{}
}

// EventType returns EventSessionEnd.
func (h *sessionEndHandler) EventType() EventType {
	return EventSessionEnd
}

// Handle processes a SessionEnd event. It logs the session completion,
// performs best-effort team directory cleanup, garbage-collects stale teams,
// clears tmux session env vars, and kills orphaned tmux sessions.
// SessionEnd hooks should not use hookSpecificOutput per Claude Code protocol.
// All cleanup is best-effort: errors are logged with slog.Warn, never returned.
func (h *sessionEndHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("session ending",
		"session_id", input.SessionID,
		"project_dir", input.ProjectDir,
	)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("session_end: could not determine home directory",
			"error", err,
		)
		return &HookOutput{}, nil
	}

	cleanupCurrentSessionTeam(input.SessionID, homeDir)
	garbageCollectStaleTeams(homeDir)
	cleanupOrphanedTmuxSessions(ctx)

	// Always clear tmux session-level GLM env vars to restore Claude models.
	// This is safe to call unconditionally:
	//   - If not in tmux: early return (checks TMUX env var)
	//   - If env vars don't exist: tmux command is a no-op
	// This ensures the lead session returns to Claude after team completion.
	clearTmuxSessionEnv()

	// SessionEnd hooks return empty JSON {} per Claude Code protocol
	// Do NOT use hookSpecificOutput for SessionEnd events
	return &HookOutput{}, nil
}

// cleanupCurrentSessionTeam removes the team directory whose leadSessionId
// matches the given sessionID. Errors are logged and never returned.
func cleanupCurrentSessionTeam(sessionID, homeDir string) {
	teamsDir := filepath.Join(homeDir, ".claude", "teams")

	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("session_end: could not read teams directory",
				"path", teamsDir,
				"error", err,
			)
		}
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		teamDir := filepath.Join(teamsDir, entry.Name())
		configPath := filepath.Join(teamDir, "config.json")

		data, err := os.ReadFile(configPath)
		if err != nil {
			// Missing config.json is normal; skip silently.
			continue
		}

		var cfg teamConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			slog.Warn("session_end: could not parse team config",
				"path", configPath,
				"error", err,
			)
			continue
		}

		if cfg.LeadSessionID == sessionID {
			if err := os.RemoveAll(teamDir); err != nil {
				slog.Warn("session_end: could not remove team directory",
					"path", teamDir,
					"error", err,
				)
			} else {
				slog.Info("session_end: removed team directory for session",
					"team_dir", teamDir,
					"session_id", sessionID,
				)
			}
		}
	}
}

// garbageCollectStaleTeams removes team directories that have not been
// modified in more than 24 hours. This catches teams left behind by
// interrupted sessions. Errors are logged and never returned.
func garbageCollectStaleTeams(homeDir string) {
	const staleDuration = 24 * time.Hour

	teamsDir := filepath.Join(homeDir, ".claude", "teams")

	entries, err := os.ReadDir(teamsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("session_end: could not read teams directory for GC",
				"path", teamsDir,
				"error", err,
			)
		}
		return
	}

	cutoff := time.Now().Add(-staleDuration)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			slog.Warn("session_end: could not stat team directory",
				"name", entry.Name(),
				"error", err,
			)
			continue
		}

		if info.ModTime().Before(cutoff) {
			teamDir := filepath.Join(teamsDir, entry.Name())
			if err := os.RemoveAll(teamDir); err != nil {
				slog.Warn("session_end: could not remove stale team directory",
					"path", teamDir,
					"error", err,
				)
			} else {
				slog.Info("session_end: removed stale team directory",
					"path", teamDir,
					"age", time.Since(info.ModTime()).Round(time.Minute),
				)
			}
		}
	}
}

// getCurrentTmuxSession returns the name of the current tmux session.
// Returns empty string if not in tmux or if detection fails.
func getCurrentTmuxSession(ctx context.Context) string {
	// Check if we're in tmux
	if os.Getenv("TMUX") == "" {
		return ""
	}

	// Use tmux display-message to get current session name.
	cmd := exec.CommandContext(ctx, "tmux", "display-message", "-p", "#S")
	out, err := cmd.Output()
	if err != nil {
		slog.Warn("session_end: could not get current tmux session",
			"error", err,
		)
		return ""
	}

	return strings.TrimSpace(string(out))
}

// cleanupOrphanedTmuxSessions kills tmux sessions that are not currently
// attached. The cleanup is capped at 4 seconds to stay within the SessionEnd
// hook timeout budget. If tmux is not installed or no sessions exist, the
// function returns silently.
func cleanupOrphanedTmuxSessions(ctx context.Context) {
	// Reserve 4 seconds for tmux cleanup, leaving 1 second buffer.
	cleanupCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// Get current tmux session name to protect it from being killed.
	currentSession := getCurrentTmuxSession(cleanupCtx)

	// List all tmux sessions.
	listCmd := exec.CommandContext(cleanupCtx, "tmux", "list-sessions")
	out, err := listCmd.Output()
	if err != nil {
		// tmux not installed, no server running, or no sessions — all fine.
		return
	}

	lines := strings.SplitSeq(strings.TrimSpace(string(out)), "\n")
	for line := range lines {
		if line == "" {
			continue
		}
		// Skip the current tmux session - never kill the user's actual session.
		name, _, found := strings.Cut(line, ":")
		if !found || name == "" {
			continue
		}
		if name == currentSession {
			continue
		}

		// Sessions currently attached contain "(attached)".
		if strings.Contains(line, "(attached)") {
			continue
		}

		killCmd := exec.CommandContext(cleanupCtx, "tmux", "kill-session", "-t", name)
		if err := killCmd.Run(); err != nil {
			slog.Warn("session_end: could not kill orphaned tmux session",
				"session", name,
				"error", err,
			)
		} else {
			slog.Info("session_end: killed orphaned tmux session",
				"session", name,
			)
		}
	}
}

// clearTmuxSessionEnv removes GLM environment variables from tmux session.
// Called when team mode completes to restore Claude models for the lead session.
// This ensures that after --team mode, the leader returns to using Claude models
// instead of continuing to use GLM from the tmux session-level env vars.
func clearTmuxSessionEnv() {
	// Skip if not in tmux
	if os.Getenv("TMUX") == "" {
		return
	}

	// GLM environment variables to clear
	envVars := []string{
		"ANTHROPIC_AUTH_TOKEN",
		"ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	}

	for _, name := range envVars {
		cmd := exec.Command("tmux", "set-environment", "-u", name)
		if err := cmd.Run(); err != nil {
			// Log warning but don't fail - variable might not exist
			slog.Warn("session_end: failed to clear tmux env",
				"env", name,
				"error", err,
			)
		} else {
			slog.Info("session_end: cleared tmux env", "env", name)
		}
	}
}
