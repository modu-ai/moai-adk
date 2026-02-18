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
// and kills orphaned tmux sessions.
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

// cleanupOrphanedTmuxSessions kills tmux sessions that are not currently
// attached. The cleanup is capped at 4 seconds to stay within the SessionEnd
// hook timeout budget. If tmux is not installed or no sessions exist, the
// function returns silently.
func cleanupOrphanedTmuxSessions(ctx context.Context) {
	// Reserve 4 seconds for tmux cleanup, leaving 1 second buffer.
	cleanupCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// List all tmux sessions.
	listCmd := exec.CommandContext(cleanupCtx, "tmux", "list-sessions")
	out, err := listCmd.Output()
	if err != nil {
		// tmux not installed, no server running, or no sessions — all fine.
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Sessions currently attached contain "(attached)".
		if strings.Contains(line, "(attached)") {
			continue
		}

		// Session name is the part before the first colon.
		name, _, found := strings.Cut(line, ":")
		if !found || name == "" {
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
