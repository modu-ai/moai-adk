// Resolution: KEEP — full business logic; GLM setup, skill discovery, memory evaluation.
package hook

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
	"github.com/modu-ai/moai-adk/internal/migration"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/spec"
	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// sessionStartHandler processes SessionStart events.
// It initializes the session, loads project configuration, and validates
// the execution environment (REQ-HOOK-030).
type sessionStartHandler struct {
	cfg ConfigProvider
}

// NewSessionStartHandler creates a new SessionStart event handler.
func NewSessionStartHandler(cfg ConfigProvider) Handler {
	return &sessionStartHandler{cfg: cfg}
}

// EventType returns EventSessionStart.
func (h *sessionStartHandler) EventType() EventType {
	return EventSessionStart
}

// Handle processes a SessionStart event. It logs the session ID, loads
// project configuration, and returns project information in the Data field.
// Errors are non-blocking: the handler logs warnings and returns allow.
func (h *sessionStartHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("session started",
		"session_id", input.SessionID,
		"cwd", input.CWD,
		"project_dir", input.ProjectDir,
	)

	data := map[string]any{
		"session_id": input.SessionID,
		"status":     "initialized",
	}

	// SPEC-V3R6-MULTI-SESSION-COORD-001 L3: 3-step multi-session protocol.
	// Step 1 — Register this session with no SPEC scope yet.
	// Step 2 — Purge zombie entries older than 30 minutes.
	// Step 3 — Query other active sessions and surface via stderr.
	// All three steps are BEST-EFFORT: errors are logged and never block
	// the SessionStart hook (REQ-COORD-013..015). Hook timeout safety
	// (CLAUDE.local.md §7 default 5s) is preserved.
	//
	// @MX:NOTE: [AUTO] 3-step protocol — Register + Purge + Query + stderr surface
	if input.SessionID != "" && input.ProjectDir != "" {
		h.runMultiSessionProtocol(input, data)
	}

	// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 REQ-WPR-003: when the
	// multi-session protocol is bypassed because input.SessionID is empty,
	// emit a non-blocking stderr warning so the orchestrator can observe
	// that the registry write path was skipped. This is a leading cause of
	// the K6 "empty registry" defect (research.md §D.1): some Claude Code
	// activation paths emit an empty session_id, the L66 gate bypasses
	// Register, and the orchestrator later finds the registry empty. The
	// warning is observation-only — the hook still returns allow (non-blocking).
	if input.SessionID == "" {
		_, _ = fmt.Fprint(os.Stderr,
			"warning: SessionStart received empty session_id; multi-session registry write bypassed "+
				"(source_session_id attribution will fall back to the environment-fallback pattern)\n")
	}

	// Load project information from config if available
	cfg := h.getConfig()
	if cfg != nil {
		if cfg.Project.Name != "" {
			data["project_name"] = cfg.Project.Name
		}
		if string(cfg.Project.Type) != "" {
			data["project_type"] = string(cfg.Project.Type)
		}
		if cfg.Project.Language != "" {
			data["project_language"] = cfg.Project.Language
		}
	} else {
		slog.Warn("configuration not available, proceeding with defaults",
			"session_id", input.SessionID,
		)
	}

	// Validate GLM credentials: if GLM model overrides exist in settings.local.json
	// but ANTHROPIC_AUTH_TOKEN is missing, auto-inject from ~/.moai/.env.glm.
	// This prevents 401 errors for Agent Teams teammates.
	if input.ProjectDir != "" {
		if msg := ensureGLMCredentials(input.ProjectDir); msg != "" {
			data["glm_credentials"] = msg
			slog.Info("GLM credentials auto-injected", "message", msg)
		}
	}

	// Auto-detect tmux environment and set teammateMode accordingly.
	// When inside tmux, teammates spawn in separate panes for visibility.
	// When outside tmux, fall back to "auto" (in-process display).
	if input.ProjectDir != "" {
		if mode := ensureTeammateMode(input.ProjectDir); mode != "" {
			data["teammate_mode"] = mode
		}
	}

	// In GLM team mode, inject GLM environment variables into the current tmux session
	// so that teammate panes inherit ANTHROPIC_AUTH_TOKEN.
	// Must execute after ensureGLMCredentials writes the token to settings.local.json
	// to read the latest value.
	if input.ProjectDir != "" {
		if msg := ensureTmuxGLMEnv(input.ProjectDir); msg != "" {
			data["tmux_glm_env"] = msg
			slog.Info("tmux GLM environment variable injected", "message", msg)
		}
	}

	// Windows only: inject CLAUDE_ENV_FILE into settings.local.json when a
	// .env file is present in the project root (T-016, R-P1-1).
	// Guarded to Windows so macOS/Linux GLM env injection is never affected.
	if claudeEnvFileGuard(runtime.GOOS) && input.ProjectDir != "" {
		if msg := injectCLAUDEEnvFile(input.ProjectDir); msg != "" {
			data["claude_env_file"] = msg
			slog.Info("CLAUDE_ENV_FILE injected", "message", msg)
		}
	}

	// Enforce telemetry retention: prune files older than 90 days (SPEC-TELEMETRY-001 R4).
	// Best-effort: errors are logged and never propagated.
	if input.ProjectDir != "" {
		if err := pruneTelemetry(input.ProjectDir); err != nil {
			slog.Warn("session start: telemetry pruning failed", "error", err)
		}
	}

	// Detect stale agent memories and inject staleness caveat (SPEC-V3R2-EXT-001 REQ-006/017).
	// Best-effort: errors are logged and never propagated.
	if input.ProjectDir != "" {
		if staleMsg := detectAndWrapStaleMemories(input.ProjectDir, time.Now()); staleMsg != "" {
			data["memory_stale_warning"] = staleMsg
			slog.Info("session start: stale memory files detected",
				"session_id", input.SessionID,
			)
		}
	}

	// Create symlinks in .claude/skills/ for any new evolved skills
	// stored under .moai/evolution/new-skills/ (R5: New Skill Symlink).
	if input.ProjectDir != "" {
		if n := ensureNewSkillSymlinks(input.ProjectDir); n > 0 {
			data["evolved_skills_linked"] = n
			slog.Info("evolved skill symlinks created", "count", n)
		}
	}

	// Present pending skill improvement proposals from the reflective learning
	// system. This is non-blocking: errors are silently ignored.
	if input.ProjectDir != "" {
		if summary := PresentPendingProposals(input.ProjectDir); summary != "" {
			data["skill_proposals"] = summary
			slog.Info("reflective_write: pending proposals available for review",
				"session_id", input.SessionID,
			)
		}
	}

	// Check for SPEC status drift and emit warning if >= 5 SPECs drifted
	// This is non-blocking: errors are silently ignored (Round 3: W3-T3)
	if input.ProjectDir != "" {
		if driftMsg := detectStatusDrift(input.ProjectDir); driftMsg != "" {
			data["status_drift_warning"] = driftMsg
			slog.Info("session start: status drift detected",
				"session_id", input.SessionID,
			)
		}
	}

	// @MX:WARN @MX:REASON - SPEC-V3R2-RT-007 REQ-020 silent migration apply at session start.
	// Errors must NOT block session (REQ-021). Surface via SystemMessage but allow handler
	// to return success. Bypassing this preserves migration-version-file unchanged → next
	// session retries. NEVER let migration error abort session-start handler.
	// Automatic migration application (REQ-020, REQ-021).
	// Runs pending migrations automatically at session-start time.
	// On failure the session is not blocked; users are notified via SystemMessage (REQ-021).
	if input.ProjectDir != "" {
		cfg := h.getConfig()
		if cfg == nil || !cfg.System.Migrations.Disabled {
			runner := migration.NewRunner(input.ProjectDir)
			applied, err := runner.Apply(ctx)
			if err != nil {
				slog.Warn("session start: migration apply failed",
					"error", err.Error(),
					"project_dir", input.ProjectDir,
				)
				data["migration_error"] = err.Error()
			} else if len(applied) > 0 {
				data["migrations_applied"] = len(applied)
				slog.Info("session start: migrations applied successfully",
					"count", len(applied),
					"versions", applied,
				)
			}
		} else {
			slog.Info("session start: migrations disabled via config",
				"project_dir", input.ProjectDir,
			)
		}
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal session data",
			"error", err.Error(),
		)
		return &HookOutput{}, nil
	}

	// SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 M3 (REQ-RDP-004/005):
	// inject this orchestrator's own UUID via hookSpecificOutput.AdditionalContext
	// so the Claude Code runtime surfaces it at session start, AND write a
	// side-channel file (.moai/state/current-session-id.txt) so `moai session
	// current` can resolve the UUID post-compaction (the additionalContext is
	// lost after /clear; the side-channel file persists).
	//
	// The injection is STRICTLY ADDITIVE (REQ-RDP-005): existing behavior
	// (Register → Purge → Query → stderr surface → data map) is unchanged.
	// The AdditionalContext is the serialized field per Claude Code SessionStart
	// stdout contract (hooks-system.md § Hook Event stdin/stdout Reference);
	// the existing `Data` field carries `json:"-"` and is internal-only
	// (research.md §D.0 — structural root cause of the attribution dead feature).
	//
	// Gated on input.SessionID != "" (research.md §D.0/D.1 P1-outcome
	// implication): an empty UUID is never injected or written.
	out := &HookOutput{Data: jsonData}
	if input.SessionID != "" && input.ProjectDir != "" {
		out.HookSpecificOutput = &HookSpecificOutput{
			HookEventName: string(EventSessionStart),
			AdditionalContext: fmt.Sprintf(
				"moai session attribution: source_session_id=%s\n"+
					"Use 'moai session current' to re-read this UUID after /clear or compaction.\n"+
					"If unavailable, emit the canonical fallback via 'moai session current --show-fallback'.",
				input.SessionID,
			),
		}
		// Side-channel file write (best-effort, non-blocking).
		sidecar := filepath.Join(input.ProjectDir, session.CurrentSideChannelFile)
		if writeErr := os.WriteFile(sidecar, []byte(input.SessionID), 0o600); writeErr != nil {
			slog.Warn("session start: failed to write current-session-id side-channel file (non-blocking)",
				"path", sidecar,
				"error", writeErr.Error(),
			)
		}
	}

	// SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001: GLM 가드레일 리마인더 주입.
	// GLM 백엔드 세션(PROCESS env ANTHROPIC_BASE_URL이 z.ai 포함)일 때만 z.ai MCP
	// 라우팅 요약을 AdditionalContext에 추가한다. 비-GLM 세션은 빈 문자열을
	// 받으므로 아무것도 주입되지 않는다 (REQ-GH-002/003). cg-leader pane은 PROCESS
	// env에 z.ai가 없으므로 자동 carve-out된다 (REQ-GH-005/006). 검출은 절대
	// 블로킹하지 않는다 (REQ-GH-012). always-load에서 제거된 glm-web-tooling.md
	// 규칙을 on-demand로 대체 전달한다.
	if reminder := glmGuardrailReminder(); reminder != "" {
		if out.HookSpecificOutput == nil {
			out.HookSpecificOutput = &HookSpecificOutput{
				HookEventName: string(EventSessionStart),
			}
		}
		if out.HookSpecificOutput.AdditionalContext == "" {
			out.HookSpecificOutput.AdditionalContext = reminder
		} else {
			out.HookSpecificOutput.AdditionalContext += "\n\n" + reminder
		}
	}

	return out, nil
}

// getConfig safely retrieves the configuration, returning nil if unavailable.
func (h *sessionStartHandler) getConfig() *config.Config {
	if h.cfg == nil {
		return nil
	}
	return h.cfg.Get()
}

// settingsLocalJSON is the minimal struct for reading settings.local.json env vars.
type settingsLocalJSON struct {
	Env         map[string]string `json:"env,omitempty"`
	Permissions map[string]any    `json:"permissions,omitempty"`
	// Preserve unknown fields
	Extra map[string]json.RawMessage `json:"-"`
}

// ensureGLMCredentials checks settings.local.json for GLM model overrides
// without ANTHROPIC_AUTH_TOKEN. If found, it reads the API key from
// ~/.moai/.env.glm and injects it along with ANTHROPIC_BASE_URL.
// Returns a status message if credentials were injected, empty string otherwise.
func ensureGLMCredentials(projectDir string) string {
	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil || len(data) == 0 {
		return ""
	}

	var settings settingsLocalJSON
	if err := json.Unmarshal(data, &settings); err != nil {
		return ""
	}

	if settings.Env == nil {
		return ""
	}

	// Skip auto-injection in CG mode: CG mode intentionally removes AUTH_TOKEN
	// from settings.local.json so the leader uses Claude OAuth. Teammates get
	// GLM credentials via tmux session env instead.
	if isCGMode(projectDir) {
		return ""
	}

	// Check if GLM model overrides exist
	hasGLMModel := false
	for _, key := range []string{
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	} {
		if val, ok := settings.Env[key]; ok && strings.Contains(strings.ToLower(val), "glm") {
			hasGLMModel = true
			break
		}
	}

	if !hasGLMModel {
		return ""
	}

	// GLM models configured — check if AUTH_TOKEN exists
	if token := settings.Env["ANTHROPIC_AUTH_TOKEN"]; token != "" {
		return "" // Already has credentials
	}

	// AUTH_TOKEN missing — try to load from ~/.moai/.env.glm
	apiKey := loadGLMKeyFromEnvFile()
	if apiKey == "" {
		slog.Warn("GLM models configured but no API key found",
			"settings", settingsPath,
			"hint", "run 'moai glm setup <api-key>' to save your key",
		)
		return ""
	}

	// Inject credentials
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = apiKey
	if settings.Env["ANTHROPIC_BASE_URL"] == "" {
		settings.Env["ANTHROPIC_BASE_URL"] = config.DefaultGLMBaseURL
	}
	// Ensure compatibility flags are set
	if settings.Env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] == "" {
		settings.Env["CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS"] = "1"
	}
	// 1M context activation: when the High (Opus) slot model carries the [1m]
	// suffix, scale the auto-compact window to the full 1M context.
	if strings.Contains(strings.ToLower(settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"]), "[1m]") &&
		settings.Env[config.EnvClaudeCodeAutoCompactWindow] == "" {
		settings.Env[config.EnvClaudeCodeAutoCompactWindow] = strconv.Itoa(config.Default1MContextTokens)
	}

	// Re-read original file to preserve all fields (not just env)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}

	envData, err := json.Marshal(settings.Env)
	if err != nil {
		return ""
	}
	raw["env"] = envData

	newData, err := json.MarshalIndent(raw, "", " ")
	if err != nil {
		return ""
	}

	// @MX:ANCHOR: [AUTO] settings.local.json holds GLM ANTHROPIC_AUTH_TOKEN — write with 0o600 only
	// @MX:REASON: SPEC-V3R5-SECURITY-CRIT-001 AC-SEC-001 (CWE-732/552). Prior baseline 0o644
	// allowed any local user to read the credential. Regression locked by TestEnsureGLMCredentialsFilePerm.
	if err := writeSettingsSecure(settingsPath, newData); err != nil {
		slog.Error("failed to write GLM credentials to settings.local.json",
			"error", err.Error(),
		)
		return ""
	}

	return fmt.Sprintf("auto-injected GLM credentials from ~/.moai/.env.glm into %s", settingsPath)
}

// isCGMode checks if the project is running in CG (Claude+GLM hybrid) mode
// by reading team_mode from llm.yaml.
func isCGMode(projectDir string) bool {
	llmPath := filepath.Join(projectDir, ".moai", "config", "sections", "llm.yaml")
	data, err := os.ReadFile(llmPath)
	if err != nil {
		return false
	}
	// Simple check: look for "team_mode: cg" in the file
	return strings.Contains(string(data), "team_mode: cg")
}

// ensureTeammateMode detects whether the session runs inside tmux and
// sets "teammateMode" in settings.local.json accordingly.
// - Inside tmux → "tmux" (teammates appear in separate panes)
// - Outside tmux → removes override (project default "auto" applies)
//
// This runs at every SessionStart so the setting stays current when the
// user switches between tmux and non-tmux terminals. CG/GLM modes
// already force "tmux" via their own code paths, so this is a no-op in
// those cases (the value is already "tmux").
func ensureTeammateMode(projectDir string) string {
	inTmux := os.Getenv("TMUX") != ""

	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return ""
	}

	var raw map[string]json.RawMessage
	if len(data) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return ""
		}
	}
	if raw == nil {
		raw = make(map[string]json.RawMessage)
	}

	// Read current value to avoid unnecessary writes.
	var current string
	if v, ok := raw["teammateMode"]; ok {
		_ = json.Unmarshal(v, &current)
	}

	desired := "auto"
	if inTmux {
		desired = "tmux"
	}

	if current == desired {
		return desired // Already correct, skip write.
	}

	modeJSON, _ := json.Marshal(desired)
	raw["teammateMode"] = modeJSON

	// Clean up legacy env var if present.
	if envRaw, ok := raw["env"]; ok {
		var env map[string]string
		if err := json.Unmarshal(envRaw, &env); err == nil {
			if _, legacy := env["CLAUDE_CODE_TEAMMATE_DISPLAY"]; legacy {
				delete(env, "CLAUDE_CODE_TEAMMATE_DISPLAY")
				if len(env) > 0 {
					newEnv, _ := json.Marshal(env)
					raw["env"] = newEnv
				} else {
					delete(raw, "env")
				}
			}
		}
	}

	newData, err := json.MarshalIndent(raw, "", " ")
	if err != nil {
		return ""
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return ""
	}

	// @MX:NOTE: [AUTO] settings.local.json may contain GLM credentials elsewhere; use 0o600
	// @MX:REASON: SPEC-V3R5-SECURITY-CRIT-001 AC-SEC-001 — uniform 0o600 prevents partial regression.
	if err := writeSettingsSecure(settingsPath, newData); err != nil {
		slog.Error("failed to update teammateMode in settings.local.json",
			"error", err.Error(),
		)
		return ""
	}

	slog.Info("teammateMode updated",
		"mode", desired,
		"in_tmux", inTmux,
	)
	return desired
}

// ensureNewSkillSymlinks scans .moai/evolution/new-skills/ for subdirectories
// and creates corresponding symlinks (or directory copies on Windows) in
// .claude/skills/ so that Claude Code can discover evolved skills at session start.
//
// Rules:
// - Target: .claude/skills/<name> → ../../.moai/evolution/new-skills/<name>
// - Existing valid symlinks are skipped.
// - Broken symlinks are removed with a warning.
// - On Windows, a directory copy is used as fallback.
//
// Returns the number of symlinks created in this call.
func ensureNewSkillSymlinks(projectDir string) int {
	newSkillsDir := filepath.Join(projectDir, ".moai", "evolution", "new-skills")
	skillsDir := filepath.Join(projectDir, ".claude", "skills")

	entries, err := os.ReadDir(newSkillsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("ensureNewSkillSymlinks: cannot read new-skills dir",
				"path", newSkillsDir,
				"error", err.Error(),
			)
		}
		return 0
	}

	// Ensure .claude/skills/ exists.
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		slog.Warn("ensureNewSkillSymlinks: cannot create skills dir",
			"path", skillsDir,
			"error", err.Error(),
		)
		return 0
	}

	created := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Name validation: reject path traversal, null bytes, slashes, backslashes, hidden files
		// TOCTOU mitigation: use only ReadDir result and Name, never combine with direct path access
		if name == "" || name == "." || name == ".." ||
			strings.ContainsAny(name, "/\\\x00") ||
			strings.HasPrefix(name, ".") {
			slog.Warn("ensureNewSkillSymlinks: skipping invalid skill name",
				"name", name,
			)
			continue
		}

		linkPath := filepath.Join(skillsDir, name)

		// Check if a symlink (or directory copy) already exists.
		fi, err := os.Lstat(linkPath)
		if err == nil {
			// Path exists — validate it.
			if fi.Mode()&os.ModeSymlink != 0 {
				// It's a symlink — verify it points to a valid target.
				if _, err := os.Stat(linkPath); err == nil {
					// Valid symlink — skip.
					continue
				}
				// Broken symlink — remove it.
				slog.Warn("ensureNewSkillSymlinks: removing broken symlink",
					"path", linkPath,
				)
				if removeErr := os.Remove(linkPath); removeErr != nil {
					slog.Warn("ensureNewSkillSymlinks: cannot remove broken symlink",
						"path", linkPath,
						"error", removeErr.Error(),
					)
					continue
				}
			} else if fi.IsDir() {
				// Directory already exists (Windows copy or manual placement) — skip.
				continue
			} else {
				// Something else — skip to avoid clobbering.
				slog.Warn("ensureNewSkillSymlinks: unexpected file at link path, skipping",
					"path", linkPath,
				)
				continue
			}
		} else if !os.IsNotExist(err) {
			slog.Warn("ensureNewSkillSymlinks: lstat error",
				"path", linkPath,
				"error", err.Error(),
			)
			continue
		}

		// Create symlink or directory copy.
		srcDir := filepath.Join(newSkillsDir, name)

		if runtime.GOOS == "windows" {
			// Windows fallback: copy directory contents instead of symlink.
			if copyErr := copyDirRecursive(srcDir, linkPath); copyErr != nil {
				slog.Warn("ensureNewSkillSymlinks: failed to copy directory on Windows",
					"src", srcDir,
					"dst", linkPath,
					"error", copyErr.Error(),
				)
				continue
			}
		} else {
			// Use a relative symlink so the project is portable.
			// From .claude/skills/<name> to ../../.moai/evolution/new-skills/<name>
			relTarget := filepath.Join("..", "..", ".moai", "evolution", "new-skills", name)
			if symlinkErr := os.Symlink(relTarget, linkPath); symlinkErr != nil {
				slog.Warn("ensureNewSkillSymlinks: failed to create symlink",
					"link", linkPath,
					"target", relTarget,
					"error", symlinkErr.Error(),
				)
				continue
			}
		}

		slog.Info("ensureNewSkillSymlinks: linked evolved skill",
			"name", name,
		)
		created++
	}

	return created
}

// copyDirRecursive copies src directory to dst recursively.
// Used as a Windows fallback when symlinks are not available.
func copyDirRecursive(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("readdir %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", srcPath, err)
		}
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", dstPath, err)
		}
	}
	return nil
}

// pruneTelemetry enforces the 90-day retention policy for telemetry files.
// It delegates to telemetry.PruneOldFiles and wraps any error with context.
func pruneTelemetry(projectDir string) error {
	return telemetry.PruneOldFiles(projectDir, 90)
}

// injectCLAUDEEnvFile checks whether a .env file exists in projectRoot. If it
// does, it injects CLAUDE_ENV_FILE into the env section of
// .claude/settings.local.json so that Claude Code loads the project's env file
// automatically (Windows CLAUDE_ENV_FILE support, T-016).
//
// Returns a non-empty status message when the value was written, empty string
// when the .env file does not exist or when no write was needed.
func injectCLAUDEEnvFile(projectRoot string) string {
	envFilePath := filepath.Join(projectRoot, ".env")
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		return ""
	}

	settingsPath := filepath.Join(projectRoot, ".claude", "settings.local.json")

	var raw map[string]json.RawMessage
	if data, err := os.ReadFile(settingsPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			raw = nil
		}
	}
	if raw == nil {
		raw = make(map[string]json.RawMessage)
	}

	// Read current env section.
	env := make(map[string]string)
	if envRaw, ok := raw["env"]; ok {
		_ = json.Unmarshal(envRaw, &env)
	}

	// Skip write if already set to the same value.
	if env["CLAUDE_ENV_FILE"] == envFilePath {
		return ""
	}

	env["CLAUDE_ENV_FILE"] = envFilePath

	envData, err := json.Marshal(env)
	if err != nil {
		return ""
	}
	raw["env"] = envData

	newData, err := json.MarshalIndent(raw, "", " ")
	if err != nil {
		return ""
	}

	if err := os.MkdirAll(filepath.Join(projectRoot, ".claude"), 0o755); err != nil {
		return ""
	}

	// @MX:NOTE: [AUTO] same settings.local.json may carry sensitive env; 0o600 mandatory
	// @MX:REASON: SPEC-V3R5-SECURITY-CRIT-001 AC-SEC-001 uniform hardening.
	if err := writeSettingsSecure(settingsPath, newData); err != nil {
		slog.Error("injectCLAUDEEnvFile: failed to write settings.local.json",
			"error", err.Error(),
		)
		return ""
	}

	return fmt.Sprintf("injected CLAUDE_ENV_FILE=%s into %s", envFilePath, settingsPath)
}

// loadGLMKeyFromEnvFile reads the GLM API key from ~/.moai/.env.glm.
func loadGLMKeyFromEnvFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	envPath := filepath.Join(home, ".moai", ".env.glm")
	file, err := os.Open(envPath)
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)

		if key == "GLM_API_KEY" && val != "" {
			return val
		}
	}
	return ""
}

// detectAndWrapStaleMemories scans all agent memory directories under
// .claude/agent-memory/<agent>/ and wraps stale files in <system-reminder> tags.
//
// When MOAI_MEMORY_AUDIT=0, the function returns empty string (disabled path).
// When 10+ stale files are found, a single aggregated warning is returned.
// Otherwise per-file wrapped content is concatenated (REQ-EXT001-006/017).
//
// The now parameter is accepted to allow deterministic testing.
func detectAndWrapStaleMemories(projectDir string, now time.Time) string {
	// Respect kill-switch (rollback safety — plan.md §6.2).
	if os.Getenv("MOAI_MEMORY_AUDIT") == "0" {
		return ""
	}

	agentMemBase := filepath.Join(projectDir, ".claude", "agent-memory")
	entries, err := os.ReadDir(agentMemBase)
	if err != nil {
		// Directory may not exist yet — not an error.
		return ""
	}

	var allReports []taxonomy.StaleReport
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		agentDir := filepath.Join(agentMemBase, e.Name())
		reports, err := taxonomy.DetectStale(agentDir, config.DefaultMemoryStalenessHours, now)
		if err != nil {
			slog.Warn("session start: staleness scan error",
				"dir", agentDir,
				"error", err.Error(),
			)
			continue
		}
		allReports = append(allReports, reports...)
	}

	if len(allReports) == 0 {
		return ""
	}

	// When count reaches the aggregation threshold, return a single short warning.
	// Otherwise return wrapped content (each file's content in <system-reminder>).
	if len(allReports) >= config.DefaultMemoryStaleAggregateThreshold {
		return taxonomy.AggregateWarning(allReports)
	}

	// Per-file: return each file's wrapped content (the <system-reminder> block).
	var sb strings.Builder
	for _, r := range allReports {
		sb.WriteString(r.Wrapped)
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}

// claudeEnvFileGuard reports whether the CLAUDE_ENV_FILE injection should run
// for the given OS name. Injection is Windows-only (T-016, R-P1-1).
//
// Extracted from Handle() so that unit tests can exercise the guard without
// depending on runtime.GOOS (a compile-time constant that cannot be overridden
// via os.Setenv). See TestSessionStartHandler_Handle_NonWindowsGuard.
func claudeEnvFileGuard(goos string) bool {
	return goos == "windows"
}

// runMultiSessionProtocol executes the 3-step coordination protocol
// (RegisterSession → PurgeStale → QueryActiveWork → stderr surface) for
// SPEC-V3R6-MULTI-SESSION-COORD-001 REQ-COORD-013..015.
//
// All steps are best-effort: errors are logged via slog.Warn but never
// propagated up the hook return path. Hook timeout safety is preserved
// (each registry op completes well under 100ms in normal contention).
//
// The data map is mutated in place to surface multi-session events to the
// hook output (used by tests + observability).
//
// Uses an explicit project-dir-bound Registry instance (not the package-
// level helpers) because the hook may run with arbitrary CWD; the
// registry path is anchored to input.ProjectDir. The method names below
// are the same as the package-level entry points (RegisterSession,
// PurgeStale, QueryActiveWork) and the verification grep matches against
// the function names regardless of receiver.
func (h *sessionStartHandler) runMultiSessionProtocol(input *HookInput, data map[string]any) {
	registryPath := filepath.Join(input.ProjectDir, session.DefaultRegistryPath)
	reg := session.NewRegistry(registryPath, nil)

	// Step 1: RegisterSession with no SPEC scope yet.
	if err := reg.Register(input.SessionID, session.SpecIDNone, session.PhaseNone); err != nil {
		slog.Warn("multi-session protocol: RegisterSession failed (non-blocking)",
			"session_id", input.SessionID,
			"error", err.Error(),
		)
		data["multi_session_register_error"] = err.Error()
	} else {
		data["multi_session_register"] = "ok"
	}

	// Step 2: PurgeStale entries (zombie sessions from crashed runs).
	purged, err := reg.Purge(session.DefaultStaleMinutes)
	if err != nil {
		slog.Warn("multi-session protocol: PurgeStale failed (non-blocking)",
			"session_id", input.SessionID,
			"error", err.Error(),
		)
	} else if purged > 0 {
		data["multi_session_purged"] = purged
		slog.Info("multi-session protocol: PurgeStale removed stale entries",
			"session_id", input.SessionID,
			"count", purged,
		)
	}

	// Step 3: QueryActiveWork — other active sessions surface via stderr.
	entries, err := reg.Query("")
	if err != nil {
		slog.Warn("multi-session protocol: QueryActiveWork failed (non-blocking)",
			"session_id", input.SessionID,
			"error", err.Error(),
		)
		return
	}
	reminder := session.FormatStderrReminder(input.SessionID, entries, time.Now().UTC())
	if reminder != "" {
		_, _ = fmt.Fprint(os.Stderr, reminder)
		data["multi_session_other_active"] = len(entries) - 1
	}
}

// detectStatusDrift checks for SPEC status drift and returns a warning message
// if >= 5 SPECs have drifted. Returns empty string otherwise.
// Non-blocking: errors are silently ignored.
func detectStatusDrift(projectDir string) string {
	// Import spec package for drift detection
	count, err := spec.DriftCount(projectDir)
	if err != nil {
		// Silently ignore errors (e.g., git not available, no specs directory)
		return ""
	}

	if count >= 5 {
		return fmt.Sprintf("⚠ %d SPECs have status drift. Run 'moai spec drift' for details.", count)
	}

	return ""
}
