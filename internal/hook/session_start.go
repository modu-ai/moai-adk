package hook

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
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
