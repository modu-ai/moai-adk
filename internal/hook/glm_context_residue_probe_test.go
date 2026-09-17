//go:build residue_probe

package hook

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Reproducer for what card t802 did NOT repair on the SessionEnd cleanup, kept
// under the `residue_probe` build tag so it stays in history and runnable
// (`go test -tags residue_probe ./internal/hook/`) without failing the default
// suite. It asserts the DESIRED post-cleanup state, so a failure here is the
// remaining defect, not a broken test.
//
// The live-key set IS repaired and guarded in the default suite — see
// glm_settings_cleanup_test.go. Two things are left, both handed to card t888:
//   - the legacy-key tail only the deleted injectGLMEnv ever wrote;
//   - the stranding property: this cleanup gates on ANTHROPIC_BASE_URL, so a
//     file that lost the indicator while keeping GLM keys can never be cleaned.
//     t802 removed the LIVE route into that state (removeGLMEnv now clears the
//     context-window pair together with the indicator), so the state is now
//     reachable only from a file an older binary wrote. Closing the gate itself
//     is a design change — see the negative control in the default-suite guard
//     for why the gate is not simply removable.
//
// All writes stay inside t.TempDir(); no real settings file is touched.

// TestProbeCleanupLeavesLegacyAndOtherAxisKeys — keys the SessionEnd cleanup
// still does not clear. MOAI_STATUSLINE_CONTEXT_SIZE belongs to the tmux axis
// but reaches settings.local.json via an older binary, so it lands here too.
func TestProbeCleanupLeavesLegacyAndOtherAxisKeys(t *testing.T) {
	scrubGatewayEnv(t)
	root, settingsPath := writeCleanupFixture(t, map[string]string{
		config.EnvAnthropicAuthToken:         "glm-key",
		config.EnvAnthropicBaseURL:           "https://api.z.ai/api/anthropic",
		config.EnvAnthropicDefaultOpusModel:  "glm-5.3",
		config.EnvAnthropicDefaultFableModel: "glm-5.3",
		config.EnvStatuslineContextSize:      "1000000",
		"API_TIMEOUT_MS":                     "3000000",
	})

	cleanupGLMSettingsLocal(root)

	env := readCleanupEnv(t, settingsPath)
	for _, key := range []string{
		config.EnvAnthropicDefaultFableModel,
		config.EnvStatuslineContextSize,
		"API_TIMEOUT_MS",
	} {
		if v, ok := env[key]; ok {
			t.Errorf("cleanupGLMSettingsLocal left %s=%q", key, v)
		}
	}
}

// TestProbeStrandedResidueSurvivesIndicatorLoss — the stranding property. The
// live route into this state is closed as of t802; what remains is a file an
// older binary left behind, which no pass can ever clean while the indicator
// gate stands.
func TestProbeStrandedResidueSurvivesIndicatorLoss(t *testing.T) {
	scrubGatewayEnv(t)
	root, settingsPath := writeCleanupFixture(t, map[string]string{
		config.EnvClaudeCodeMaxContextTokens:  "1000000",
		config.EnvClaudeCodeAutoCompactWindow: "1000000",
	})

	cleanupGLMSettingsLocal(root)

	env := readCleanupEnv(t, settingsPath)
	if v, ok := env[config.EnvClaudeCodeMaxContextTokens]; ok {
		t.Errorf("stranded residue: %s=%q survives with no indicator present",
			config.EnvClaudeCodeMaxContextTokens, v)
	}
}
