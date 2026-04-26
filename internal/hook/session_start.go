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
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	budgetruntime "github.com/modu-ai/moai-adk/internal/runtime"
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

	// GLM 팀 모드에서 팀원 tmux 팬이 ANTHROPIC_AUTH_TOKEN을 상속하도록
	// 현재 tmux 세션에 GLM 환경변수를 주입합니다.
	// ensureGLMCredentials가 settings.local.json에 토큰을 기록한 후에 실행해야
	// 최신 값을 읽을 수 있습니다.
	if input.ProjectDir != "" {
		if msg := ensureTmuxGLMEnv(input.ProjectDir); msg != "" {
			data["tmux_glm_env"] = msg
			slog.Info("tmux GLM 환경변수 주입", "message", msg)
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

	// Create symlinks in .claude/skills/ for any new evolved skills
	// stored under .moai/evolution/new-skills/ (R5: New Skill Symlink).
	if input.ProjectDir != "" {
		if n := ensureNewSkillSymlinks(input.ProjectDir); n > 0 {
			data["evolved_skills_linked"] = n
			slog.Info("evolved skill symlinks created", "count", n)
		}
	}

	// Present pending skill improvement proposals from the reflective learning
	// system.  This is non-blocking: errors are silently ignored.
	if input.ProjectDir != "" {
		if summary := PresentPendingProposals(input.ProjectDir); summary != "" {
			data["skill_proposals"] = summary
			slog.Info("reflective_write: pending proposals available for review",
				"session_id", input.SessionID,
			)
		}
	}

	// Initialize Token Circuit Breaker (SPEC-V3R3-ARCH-007).
	// Load runtime.yaml and create a per-session Tracker.
	// Errors are non-blocking: defaults are used when runtime.yaml is missing.
	if input.ProjectDir != "" {
		runtimeCfgPath := filepath.Join(input.ProjectDir, ".moai", "config", "sections", "runtime.yaml")
		runtimeCfg, err := budgetruntime.LoadRuntime(runtimeCfgPath)
		if err != nil {
			slog.Debug("runtime.yaml not found, using defaults",
				"path", runtimeCfgPath,
				"error", err,
			)
			runtimeCfg = budgetruntime.DefaultRuntimeConfig()
		}
		tracker := budgetruntime.NewTracker(runtimeCfg)
		tracker.SetProjectRoot(input.ProjectDir)
		data["runtime_tracker_initialized"] = true
		slog.Info("Token Circuit Breaker initialized",
			"spec", "SPEC-V3R3-ARCH-007",
			"pre_clear_threshold", runtimeCfg.PreClearThreshold,
			"hard_clear_threshold", runtimeCfg.HardClearThreshold,
			"stall_detection_seconds", runtimeCfg.StallDetectionSeconds,
		)
		// Store tracker in data for downstream use (debug info only).
		// The tracker is not yet wired into a shared session context — that is
		// deferred to the PreToolUse hook integration (future SPEC).
		_ = tracker
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal session data",
			"error", err.Error(),
		)
		return &HookOutput{}, nil
	}

	return &HookOutput{Data: jsonData}, nil
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
	if settings.Env["DISABLE_PROMPT_CACHING"] == "" {
		settings.Env["DISABLE_PROMPT_CACHING"] = "1"
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

	newData, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return ""
	}

	if err := os.WriteFile(settingsPath, newData, 0o644); err != nil {
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
//   - Inside tmux → "tmux" (teammates appear in separate panes)
//   - Outside tmux → removes override (project default "auto" applies)
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

	newData, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return ""
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return ""
	}

	if err := os.WriteFile(settingsPath, newData, 0o644); err != nil {
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
//   - Target: .claude/skills/<name> → ../../.moai/evolution/new-skills/<name>
//   - Existing valid symlinks are skipped.
//   - Broken symlinks are removed with a warning.
//   - On Windows, a directory copy is used as fallback.
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

		// 이름 검증: 경로 순회, null 바이트, 슬래시, 백슬래시, 숨김 파일 거부
		// TOCTOU 완화: ReadDir 결과 이름만 사용하고 직접 경로 조합하지 않음
		if name == "" || name == "." || name == ".." ||
			strings.ContainsAny(name, "/\\\x00") ||
			strings.HasPrefix(name, ".") {
			slog.Warn("ensureNewSkillSymlinks: 잘못된 스킬 이름 건너뜀",
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

	newData, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return ""
	}

	if err := os.MkdirAll(filepath.Join(projectRoot, ".claude"), 0o755); err != nil {
		return ""
	}

	if err := os.WriteFile(settingsPath, newData, 0o644); err != nil {
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

// claudeEnvFileGuard reports whether the CLAUDE_ENV_FILE injection should run
// for the given OS name. Injection is Windows-only (T-016, R-P1-1).
//
// Extracted from Handle() so that unit tests can exercise the guard without
// depending on runtime.GOOS (a compile-time constant that cannot be overridden
// via os.Setenv). See TestSessionStartHandler_Handle_NonWindowsGuard.
func claudeEnvFileGuard(goos string) bool {
	return goos == "windows"
}
