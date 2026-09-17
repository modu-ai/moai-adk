//go:build residue_probe

package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Reproducer preserved under the `residue_probe` build tag so it stays in
// history and runnable (`go test -tags residue_probe ./internal/cli/`) without
// failing the default suite. It asserts the DESIRED post-cleanup state, so a
// failure here IS the defect report.
//
// PROBE (card t802): mechanical evidence that the
// settings.local.json GLM cleanup paths leave the context-window declaration
// behind. All writes stay inside t.TempDir(); no real settings file and no GLM
// credential is touched.

// residueKeys are the GLM-owned env keys a settings.local.json cleanup is
// expected to remove. buildTmuxClearVars (the tmux-axis list) clears all of
// them; the three settings-axis cleaners do not.
var residueKeys = []string{
	config.EnvClaudeCodeMaxContextTokens,
	config.EnvClaudeCodeAutoCompactWindow,
	config.EnvAnthropicDefaultFableModel,
}

func glmDirtyEnv() map[string]string {
	return map[string]string{
		config.EnvAnthropicAuthToken:          "glm-key",
		config.EnvAnthropicBaseURL:            "https://api.z.ai/api/anthropic",
		config.EnvAnthropicDefaultOpusModel:   "glm-5.3",
		config.EnvAnthropicDefaultSonnetModel: "glm-5.3",
		config.EnvAnthropicDefaultHaikuModel:  "glm-4.7-air",
		config.EnvAnthropicDefaultFableModel:  "glm-5.3",
		config.EnvClaudeCodeAutoCompactWindow: "1000000",
		config.EnvClaudeCodeMaxContextTokens:  "1000000",
		config.EnvStatuslineContextSize:       "1000000",
		"CUSTOM_VAR":                          "keep_me",
	}
}

// TestProbeRemoveGLMEnvLeavesContextWindow — `moai cc` cleanup path.
func TestProbeRemoveGLMEnvLeavesContextWindow(t *testing.T) {
	_, settingsPath := cgTestProject(t, glmDirtyEnv())

	if err := removeGLMEnv(settingsPath); err != nil {
		t.Fatalf("removeGLMEnv: %v", err)
	}

	got := readSettingsForTest(t, settingsPath)
	for _, key := range residueKeys {
		if v, ok := got.Env[key]; ok {
			t.Errorf("removeGLMEnv left %s=%q in settings.local.json", key, v)
		}
	}
	if got.Env["CUSTOM_VAR"] != "keep_me" {
		t.Errorf("user key must survive, got env: %v", got.Env)
	}
}

// TestProbeStripGLMCredsLeavesContextWindow — CG leader cleanup path.
func TestProbeStripGLMCredsLeavesContextWindow(t *testing.T) {
	_, settingsPath := cgTestProject(t, glmDirtyEnv())

	if err := mutateSettingsLocal(settingsPath, stripGLMCredsAndSetTeammateMode); err != nil {
		t.Fatalf("mutateSettingsLocal: %v", err)
	}

	got := readSettingsForTest(t, settingsPath)
	for _, key := range residueKeys {
		if v, ok := got.Env[key]; ok {
			t.Errorf("stripGLMCredsAndSetTeammateMode left %s=%q in settings.local.json", key, v)
		}
	}
}

// TestProbeTmuxClearVarsIsTheCompleteList is the positive control: the tmux-axis
// list already clears every residue key, which is why the settings-axis gap is a
// divergence between two lists rather than an unknown key set.
func TestProbeTmuxClearVarsIsTheCompleteList(t *testing.T) {
	cleared := map[string]bool{}
	for _, k := range buildTmuxClearVars() {
		cleared[k] = true
	}
	for _, key := range residueKeys {
		if !cleared[key] {
			t.Errorf("positive control broken: buildTmuxClearVars does not clear %s", key)
		}
	}
}
