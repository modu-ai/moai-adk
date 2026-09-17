//go:build residue_probe

package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Reproducer preserved under the `residue_probe` build tag so it stays in
// history and runnable (`go test -tags residue_probe ./internal/hook/`) without
// failing the default suite. It asserts the DESIRED post-cleanup state, so a
// failure here IS the defect report.
//
// PROBE (card t802): mechanical evidence for the
// SessionEnd settings.local.json cleanup. Two properties are probed —
// (1) the context-window keys survive a cleanup that does fire, and
// (2) once ANTHROPIC_BASE_URL is gone the cleanup can never fire again, so the
// residue is permanent rather than merely delayed.
// All writes stay inside t.TempDir(); no real settings file is touched.

func writeProbeSettings(t *testing.T, env map[string]string) (string, string) {
	t.Helper()
	root := t.TempDir()
	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	data, err := json.MarshalIndent(map[string]any{"env": env}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return root, settingsPath
}

func readProbeEnv(t *testing.T, path string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var raw struct {
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse settings: %v", err)
	}
	return raw.Env
}

// TestProbeCleanupGLMSettingsLocalLeavesContextWindow — SessionEnd cleanup path.
func TestProbeCleanupGLMSettingsLocalLeavesContextWindow(t *testing.T) {
	root, settingsPath := writeProbeSettings(t, map[string]string{
		config.EnvAnthropicAuthToken:          "glm-key",
		config.EnvAnthropicBaseURL:            "https://api.z.ai/api/anthropic",
		config.EnvAnthropicDefaultOpusModel:   "glm-5.3",
		config.EnvClaudeCodeMaxContextTokens:  "1000000",
		config.EnvClaudeCodeAutoCompactWindow: "1000000",
		config.EnvStatuslineContextSize:       "1000000",
	})

	cleanupGLMSettingsLocal(root)

	env := readProbeEnv(t, settingsPath)
	for _, key := range []string{
		config.EnvClaudeCodeMaxContextTokens,
		config.EnvClaudeCodeAutoCompactWindow,
		config.EnvStatuslineContextSize,
	} {
		if v, ok := env[key]; ok {
			t.Errorf("cleanupGLMSettingsLocal left %s=%q", key, v)
		}
	}
}

// TestProbeResidueIsPermanentOnceIndicatorIsGone — the stranding property.
// ANTHROPIC_BASE_URL is the GLM-active indicator. `moai cc`'s removeGLMEnv
// deletes it while leaving the context-window keys, so this later pass sees a
// non-GLM file and returns without touching the residue: nothing downstream can
// ever clean it.
func TestProbeResidueIsPermanentOnceIndicatorIsGone(t *testing.T) {
	root, settingsPath := writeProbeSettings(t, map[string]string{
		config.EnvClaudeCodeMaxContextTokens:  "1000000",
		config.EnvClaudeCodeAutoCompactWindow: "1000000",
	})

	cleanupGLMSettingsLocal(root)

	env := readProbeEnv(t, settingsPath)
	if v, ok := env[config.EnvClaudeCodeMaxContextTokens]; ok {
		t.Errorf("residue is permanent: %s=%q survives with no indicator present",
			config.EnvClaudeCodeMaxContextTokens, v)
	}
}
