//go:build residue_probe

package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Reproducer for the settings-axis divergences card t802 did NOT repair, kept
// under the `residue_probe` build tag so it stays in history and runnable
// (`go test -tags residue_probe ./internal/cli/`) without failing the default
// suite. It asserts the DESIRED post-cleanup state, so a failure here is the
// remaining defect, not a broken test.
//
// The live-key set IS repaired and guarded in the default suite — see
// glm_settings_cleanup_test.go. What is left here is the legacy-key tail: keys
// only the deleted injectGLMEnv ever wrote, which a settings.local.json written
// by an older binary can still carry. Handed to card t888 (R3: fold the
// settings-axis key lists into one SSOT).
//
// All writes stay inside t.TempDir(); no real settings file and no GLM
// credential is touched.

// legacyOnlyKeys were written only by injectGLMEnv (deleted in t802) and so can
// appear only in files left by an older binary.
var legacyOnlyKeys = []string{
	config.EnvAnthropicDefaultFableModel,
	"API_TIMEOUT_MS",
}

func glmLegacyDirtyEnv() map[string]string {
	env := map[string]string{
		config.EnvAnthropicAuthToken:          "glm-key",
		config.EnvAnthropicBaseURL:            "https://api.z.ai/api/anthropic",
		config.EnvAnthropicDefaultOpusModel:   "glm-5.3",
		config.EnvAnthropicDefaultSonnetModel: "glm-5.3",
		config.EnvAnthropicDefaultHaikuModel:  "glm-4.7-air",
	}
	for _, key := range legacyOnlyKeys {
		env[key] = "legacy"
	}
	return env
}

// TestProbeStripGLMCredsLeavesLegacyKeys — the CG leader cleanup clears every
// live key but not the legacy tail, while removeGLMEnv clears FABLE. The two
// settings-axis lists are still hand-enumerated and still diverge from each
// other; that divergence is what t888 removes.
func TestProbeStripGLMCredsLeavesLegacyKeys(t *testing.T) {
	_, settingsPath := cgTestProject(t, glmLegacyDirtyEnv())

	if err := mutateSettingsLocal(settingsPath, stripGLMCredsAndSetTeammateMode); err != nil {
		t.Fatalf("mutateSettingsLocal: %v", err)
	}

	got := readSettingsForTest(t, settingsPath)
	for _, key := range legacyOnlyKeys {
		if v, ok := got.Env[key]; ok {
			t.Errorf("stripGLMCredsAndSetTeammateMode left legacy key %s=%q", key, v)
		}
	}
}

// TestProbeSettingsAxisListsAgree states the invariant t888 should establish:
// the two settings-axis cleanup paths should clear the same key set. Today they
// do not, and the assertion below names which keys differ.
func TestProbeSettingsAxisListsAgree(t *testing.T) {
	_, removePath := cgTestProject(t, glmLegacyDirtyEnv())
	if err := removeGLMEnv(removePath); err != nil {
		t.Fatalf("removeGLMEnv: %v", err)
	}
	_, stripPath := cgTestProject(t, glmLegacyDirtyEnv())
	if err := mutateSettingsLocal(stripPath, stripGLMCredsAndSetTeammateMode); err != nil {
		t.Fatalf("mutateSettingsLocal: %v", err)
	}

	afterRemove := readSettingsForTest(t, removePath).Env
	afterStrip := readSettingsForTest(t, stripPath).Env

	for key := range afterStrip {
		if _, ok := afterRemove[key]; !ok {
			t.Errorf("stripGLMCredsAndSetTeammateMode keeps %q that removeGLMEnv clears", key)
		}
	}
	for key := range afterRemove {
		if _, ok := afterStrip[key]; !ok {
			t.Errorf("removeGLMEnv keeps %q that stripGLMCredsAndSetTeammateMode clears", key)
		}
	}
}
