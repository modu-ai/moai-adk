package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/tmux"
)

// glmTmuxKeys is the list of GLM-related environment variables to inject into the tmux session.
// New tmux panes (teammates) must have these variables set at the session level to use the GLM API.
//
// Issue #742: MOAI_STATUSLINE_CONTEXT_SIZE is included so the Claude Code
// statusline reflects the real GLM model context window (128K/200K/etc.)
// instead of the Claude slot's nominal size (1M for the Opus slot).
var glmTmuxKeys = []string{
	"ANTHROPIC_AUTH_TOKEN",
	"ANTHROPIC_BASE_URL",
	"ANTHROPIC_DEFAULT_OPUS_MODEL",
	"ANTHROPIC_DEFAULT_SONNET_MODEL",
	"ANTHROPIC_DEFAULT_HAIKU_MODEL",
	"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS",
	"DISABLE_PROMPT_CACHING",
	"API_TIMEOUT_MS",
	"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC",
	"MOAI_STATUSLINE_CONTEXT_SIZE",
}

// buildGLMTmuxEnvVars extracts only GLM-related variables from the env map in settings.local.json.
//
// Rules:
//   - Returns nil if ANTHROPIC_AUTH_TOKEN is empty (injection not needed)
//   - Only keys listed in glmTmuxKeys are included in the result
//   - Keys absent from settings are not included in the result
func buildGLMTmuxEnvVars(env map[string]string) map[string]string {
	if env["ANTHROPIC_AUTH_TOKEN"] == "" {
		return nil
	}

	result := make(map[string]string)
	for _, key := range glmTmuxKeys {
		if val, ok := env[key]; ok && val != "" {
			result[key] = val
		}
	}

	// Always include ANTHROPIC_AUTH_TOKEN if present
	if token := env["ANTHROPIC_AUTH_TOKEN"]; token != "" {
		result["ANTHROPIC_AUTH_TOKEN"] = token
	}

	return result
}

// formatTmuxGLMEnvSummary returns a summary string for the tmux session environment variable injection result.
func formatTmuxGLMEnvSummary(n int) string {
	return fmt.Sprintf("tmux session environment variables injected for GLM teammates (%d variables)", n)
}

// ensureTmuxGLMEnv is called from the SessionStart hook to inject environment variables
// into the current tmux session so that teammate tmux panes can inherit ANTHROPIC_AUTH_TOKEN in GLM mode.
//
// Behavior:
//  1. No-op if TMUX env var is absent (not inside a tmux session)
//  2. No-op if teammateMode in settings.local.json is not "tmux"
//  3. No-op if ANTHROPIC_AUTH_TOKEN is absent
//  4. Inject GLM-related environment variables into the current session via tmux set-environment
//
// Error handling: all errors are only logged as warnings; an empty string is returned.
// This function must never cause the SessionStart hook to fail.
//
// @MX:ANCHOR: [AUTO] Entry point for GLM+team mode tmux environment variable injection in the SessionStart hook
// @MX:REASON: Called directly from Handle; 3+ test cases in session_start_glm_tmux_test.go verify this
func ensureTmuxGLMEnv(projectDir string) string {
	// 1. Check if inside a tmux session
	if os.Getenv("TMUX") == "" {
		return ""
	}

	// 2. Read settings.local.json
	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil || len(data) == 0 {
		return ""
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}

	// 3. Verify teammateMode == "tmux"
	var teammateMode string
	if v, ok := raw["teammateMode"]; ok {
		_ = json.Unmarshal(v, &teammateMode)
	}
	if teammateMode != "tmux" {
		return ""
	}

	// 4. Extract env map
	envRaw, ok := raw["env"]
	if !ok {
		return ""
	}
	var env map[string]string
	if err := json.Unmarshal(envRaw, &env); err != nil {
		return ""
	}

	// 5. Build GLM env var map (returns nil if AUTH_TOKEN is absent)
	vars := buildGLMTmuxEnvVars(env)
	if len(vars) == 0 {
		return ""
	}

	// 6. Run tmux set-environment
	mgr := tmux.NewSessionManager()
	if err := mgr.InjectEnv(context.Background(), vars); err != nil {
		slog.Warn("ensureTmuxGLMEnv: tmux env var injection failed",
			"error", err.Error(),
			"hint", "tmux binary may be missing or running outside a session",
		)
		return ""
	}

	summary := formatTmuxGLMEnvSummary(len(vars))
	slog.Info("ensureTmuxGLMEnv: GLM env vars injected into tmux session",
		"vars", len(vars),
	)
	return summary
}
