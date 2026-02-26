package hook

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
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
// and clears tmux session env vars.
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

	// Always clear tmux session-level GLM env vars to restore Claude models.
	// This is safe to call unconditionally:
	//   - If not in tmux: early return (checks TMUX env var)
	//   - If env vars don't exist: tmux command is a no-op
	// This ensures the lead session returns to Claude after team completion.
	clearTmuxSessionEnv()

	// Clean up GLM env vars from settings.local.json.
	// This ensures that if the session ends unexpectedly (error, crash, agent not terminated),
	// the next session starts with Claude models instead of continuing to use GLM.
	// Note: We do NOT reset team_mode in llm.yaml - that is intentional config that
	// should persist across sessions for consistent team mode usage.
	// We also do NOT remove CLAUDE_CODE_TEAMMATE_DISPLAY - that is a display setting.
	if input.ProjectDir != "" {
		cleanupGLMSettings(input.ProjectDir)
	}

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

// clearTmuxSessionEnv removes GLM environment variables from tmux session.
// Called when team mode completes to restore Claude models for the lead session.
// This ensures that after --team mode, the leader returns to using Claude models
// instead of continuing to use GLM from the tmux session-level env vars.
func clearTmuxSessionEnv() {
	// Skip if not in tmux
	if os.Getenv("TMUX") == "" {
		return
	}

	// GLM environment variables to clear from tmux session.
	// ALL GLM vars including ANTHROPIC_AUTH_TOKEN are removed.
	// The GLM API key is stored persistently in ~/.moai/.env.glm
	// and re-injected by 'moai glm' when needed.
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

// cleanupGLMSettings removes GLM env vars from settings.local.json.
// This ensures the next session starts with Claude models.
func cleanupGLMSettings(projectDir string) {
	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("session_end: could not read settings.local.json", "error", err)
		}
		return
	}

	// Handle empty file gracefully (e.g. 0-byte file left by tests or crash)
	if len(data) == 0 {
		return
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		slog.Warn("session_end: could not parse settings.local.json", "error", err)
		return
	}

	env, ok := settings["env"].(map[string]any)
	if !ok {
		return // No env section
	}

	// Remove ALL GLM env vars from settings.local.json.
	// ANTHROPIC_AUTH_TOKEN is also removed: the GLM API key is stored
	// persistently in ~/.moai/.env.glm and re-injected by 'moai glm'.
	// Leaving AUTH_TOKEN in settings.local.json overrides Claude's OAuth
	// and causes /login errors.
	glmVars := []string{
		"ANTHROPIC_AUTH_TOKEN",
		"ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	}

	changed := false
	for _, key := range glmVars {
		if _, exists := env[key]; exists {
			delete(env, key)
			changed = true
		}
	}

	if !changed {
		return
	}

	// Clean up empty env map
	if len(env) == 0 {
		delete(settings, "env")
	}

	newData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		slog.Warn("session_end: could not marshal settings", "error", err)
		return
	}

	if err := os.WriteFile(settingsPath, newData, 0o644); err != nil {
		slog.Warn("session_end: could not write settings", "error", err)
	} else {
		slog.Info("session_end: cleaned up GLM env from settings.local.json")
	}
}
