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

// triggerSessionIndex runs session indexing as an asynchronous subprocess.
// Fire-and-forget pattern: cmd.Start() without cmd.Wait().
// The hook timeout (5 seconds) is not blocked by the indexer process.
func triggerSessionIndex(sessionID, projectDir, gitBranch string) {
	moaiBin, err := os.Executable()
	if err != nil {
		slog.Warn("session_end: could not locate moai binary for search indexing",
			"error", err,
		)
		return
	}
	cmd := exec.Command(moaiBin, "search", "--index-session", sessionID,
		"--project-path", projectDir, "--git-branch", gitBranch)
	if err := cmd.Start(); err != nil {
		slog.Warn("session_end: could not start search indexer subprocess",
			"error", err,
		)
	}
	// No cmd.Wait() — fire-and-forget subprocess.
}

// detectGitBranch returns the current git branch name.
// Returns an empty string if the git command fails or the directory is not a git repository.
func detectGitBranch(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

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
	garbageCollectOrphanedTasks(homeDir)
	cleanupOrphanedTmuxSessions(ctx)

	// Always clear tmux session-level GLM env vars to restore Claude models.
	// This is safe to call unconditionally:
	//   - If not in tmux: early return (checks TMUX env var)
	//   - If env vars don't exist: tmux command is a no-op
	// This ensures the lead session returns to Claude after team completion.
	clearTmuxSessionEnv(ctx)

	// Clean up GLM env vars from settings.local.json.
	// This handles the case where the user ran 'moai glm' but ended the session
	// without running 'moai cc'. Without this, the stale GLM key persists and
	// the next session fails to authenticate with Claude (no API key error).
	projectDir := input.CWD
	if projectDir == "" {
		projectDir = input.ProjectDir // Fallback for legacy
	}
	if projectDir != "" {
		cleanupGLMSettingsLocal(projectDir)
	}

	// Trigger session indexing asynchronously (supports moai search feature).
	if input.SessionID != "" {
		triggerSessionIndex(input.SessionID, projectDir, detectGitBranch(ctx))
	}

	slog.Info("session_end: cleanup complete",
		"session_id", input.SessionID,
	)

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
				// Also remove the corresponding task directory when the team directory is successfully deleted
				tasksDir := filepath.Join(homeDir, ".claude", "tasks", entry.Name())
				if err := os.RemoveAll(tasksDir); err != nil {
					slog.Warn("session_end: could not remove task directory for session",
						"path", tasksDir,
						"error", err,
					)
				} else {
					slog.Info("session_end: removed task directory for session",
						"task_dir", tasksDir,
						"session_id", sessionID,
					)
				}
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
				// Also remove the corresponding task directory when a stale team directory is successfully deleted
				taskDir := filepath.Join(homeDir, ".claude", "tasks", entry.Name())
				if err := os.RemoveAll(taskDir); err != nil {
					slog.Warn("session_end: could not remove stale task directory",
						"path", taskDir,
						"error", err,
					)
				} else {
					slog.Info("session_end: removed stale task directory",
						"path", taskDir,
					)
				}
			}
		}
	}
}

// garbageCollectOrphanedTasks cleans up orphaned task directories under ~/.claude/tasks/
// that have no corresponding team directory. Collects task directories left behind by
// interrupted sessions or incomplete cleanup. Errors are logged and never returned.
func garbageCollectOrphanedTasks(homeDir string) {
	tasksDir := filepath.Join(homeDir, ".claude", "tasks")
	teamsDir := filepath.Join(homeDir, ".claude", "teams")

	taskEntries, err := os.ReadDir(tasksDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("session_end: could not read tasks directory for orphan GC",
				"path", tasksDir,
				"error", err,
			)
		}
		return
	}

	for _, entry := range taskEntries {
		if !entry.IsDir() {
			continue
		}

		// Check whether the corresponding team directory exists
		teamDir := filepath.Join(teamsDir, entry.Name())
		if _, err := os.Stat(teamDir); err == nil {
			// Team directory exists, so this is not an orphan — keep it
			continue
		}

		// No team directory, so remove the orphaned task directory
		taskDir := filepath.Join(tasksDir, entry.Name())
		if err := os.RemoveAll(taskDir); err != nil {
			slog.Warn("session_end: could not remove orphaned task directory",
				"path", taskDir,
				"error", err,
			)
		} else {
			slog.Info("session_end: removed orphaned task directory",
				"path", taskDir,
			)
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

// moaiTmuxSessionPrefix is the naming convention for tmux sessions created by
// MoAI Agent Teams. Only sessions matching this prefix are eligible for cleanup.
const moaiTmuxSessionPrefix = "moai-"

// cleanupOrphanedTmuxSessions kills detached tmux sessions created by MoAI
// Agent Teams (prefix "moai-"). User-created sessions are never touched.
// The cleanup is capped at 4 seconds to stay within the SessionEnd hook
// timeout budget. If tmux is not installed or no sessions exist, the function
// returns silently.
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
		if cleanupCtx.Err() != nil {
			slog.Warn("session_end: tmux cleanup timed out",
				"timeout", 4*time.Second,
			)
		}
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

		// Only kill sessions created by MoAI (prefixed with "moai-").
		// Never kill user-created tmux sessions.
		if !strings.HasPrefix(name, moaiTmuxSessionPrefix) {
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

// glmEnvVarsToClean is the list of GLM-specific environment variables removed
// from the tmux session on session end.
// ANTHROPIC_AUTH_TOKEN is included: moai glm/cg sets it to the GLM API key in
// the tmux session. Removing it unconditionally restores the pre-v2.6 behavior
// that had no login issues. The user's real Claude credential is stored in
// ~/.claude/ (system credential storage), not in the tmux environment, so
// unsetting the tmux var is always safe — it either removes a GLM key (correct)
// or is a no-op when it was never set.
var glmEnvVarsToClean = []string{
	"ANTHROPIC_AUTH_TOKEN",
	"ANTHROPIC_BASE_URL",
	"ANTHROPIC_DEFAULT_OPUS_MODEL",
	"ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL",
}

// clearTmuxSessionEnv removes GLM environment variables from tmux session.
// Called when team mode completes to restore Claude models for the lead session.
// This ensures that after --team mode, the leader returns to using Claude models
// instead of continuing to use GLM from the tmux session-level env vars.
func clearTmuxSessionEnv(ctx context.Context) {
	// Skip if not in tmux
	if os.Getenv("TMUX") == "" {
		return
	}

	for _, name := range glmEnvVarsToClean {
		cmd := exec.CommandContext(ctx, "tmux", "set-environment", "-u", name)
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

// cleanupGLMSettingsLocal removes GLM env vars from .claude/settings.local.json
// in the given project directory. This handles sessions ended without running
// 'moai cc': the stale GLM key in settings.local.json would otherwise cause the
// next Claude Code session to fail auth ("no API key available" / /login loop).
//
// Cleanup logic mirrors removeGLMEnv() in internal/cli/cc.go:
//   - If MOAI_BACKUP_AUTH_TOKEN exists, restore it as ANTHROPIC_AUTH_TOKEN.
//   - Otherwise, delete ANTHROPIC_AUTH_TOKEN (it was a GLM key, not OAuth).
//   - Always delete: MOAI_BACKUP_AUTH_TOKEN, ANTHROPIC_BASE_URL, and the three
//     ANTHROPIC_DEFAULT_*_MODEL vars.
//
// ANTHROPIC_BASE_URL is used as the GLM-active indicator: Claude Code's OAuth
// flow never sets this variable, so its presence reliably signals GLM mode.
//
// All operations are best-effort. Errors are logged with slog.Warn and never
// returned, following the SessionEnd convention of non-fatal cleanup.
func cleanupGLMSettingsLocal(projectDir string) {
	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("session_end: could not read settings.local.json",
				"path", settingsPath,
				"error", err,
			)
		}
		return
	}

	// Treat empty file as no-op (same as removeGLMEnv in cc.go).
	if len(data) == 0 {
		return
	}

	// Round-trip the file as a raw JSON map to preserve all unknown fields.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		slog.Warn("session_end: could not parse settings.local.json",
			"path", settingsPath,
			"error", err,
		)
		return
	}

	// Extract the "env" object. If absent or not an object, nothing to clean.
	envRaw, hasEnv := raw["env"]
	if !hasEnv {
		return
	}

	var env map[string]string
	if err := json.Unmarshal(envRaw, &env); err != nil {
		slog.Warn("session_end: could not parse env in settings.local.json",
			"path", settingsPath,
			"error", err,
		)
		return
	}

	// ANTHROPIC_BASE_URL is the GLM-active indicator.
	// Claude Code's own OAuth flow never sets this variable.
	if _, glmActive := env["ANTHROPIC_BASE_URL"]; !glmActive {
		// Not in GLM mode — nothing to clean.
		return
	}

	// Restore backed-up OAuth token if present; otherwise remove the GLM key.
	if backup, ok := env["MOAI_BACKUP_AUTH_TOKEN"]; ok && backup != "" {
		env["ANTHROPIC_AUTH_TOKEN"] = backup
	} else {
		delete(env, "ANTHROPIC_AUTH_TOKEN")
	}

	delete(env, "MOAI_BACKUP_AUTH_TOKEN")
	delete(env, "ANTHROPIC_BASE_URL")
	delete(env, "ANTHROPIC_DEFAULT_HAIKU_MODEL")
	delete(env, "ANTHROPIC_DEFAULT_SONNET_MODEL")
	delete(env, "ANTHROPIC_DEFAULT_OPUS_MODEL")

	// Re-encode the cleaned env map back into the raw JSON document.
	if len(env) == 0 {
		delete(raw, "env")
	} else {
		envData, err := json.Marshal(env)
		if err != nil {
			slog.Warn("session_end: could not marshal cleaned env",
				"path", settingsPath,
				"error", err,
			)
			return
		}
		raw["env"] = json.RawMessage(envData)
	}

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		slog.Warn("session_end: could not marshal settings.local.json",
			"path", settingsPath,
			"error", err,
		)
		return
	}

	if err := os.WriteFile(settingsPath, out, 0o644); err != nil {
		slog.Warn("session_end: could not write settings.local.json",
			"path", settingsPath,
			"error", err,
		)
		return
	}

	slog.Info("session_end: removed GLM env vars from settings.local.json",
		"path", settingsPath,
	)
}
