package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Card t802 regression guard. The live settings.local.json writer is the
// SessionStart hook's ensureGLMCredentials; every key it can add must leave with
// the GLM credentials, or the residue is permanent rather than delayed —
// ANTHROPIC_BASE_URL is the GLM-active indicator the SessionEnd cleanup gates
// on, so once it is gone nothing downstream can see the file as GLM again.
//
// This runs in the default suite on purpose: a repair whose test never runs is
// not protected. The remaining legacy-key divergences are held separately in
// glm_context_residue_probe_test.go under the `residue_probe` build tag.

// liveSettingsAxisKeys is every env key ensureGLMCredentials can write into
// .claude/settings.local.json, plus the OAuth-backup key the cleanup paths own.
func liveSettingsAxisKeys() []string {
	return []string{
		config.EnvAnthropicAuthToken,
		"MOAI_BACKUP_AUTH_TOKEN",
		config.EnvAnthropicBaseURL,
		config.EnvAnthropicDefaultOpusModel,
		config.EnvAnthropicDefaultSonnetModel,
		config.EnvAnthropicDefaultHaikuModel,
		config.EnvClaudeCodeDisableExperimentalBetas,
		config.EnvClaudeCodeAutoCompactWindow,
		config.EnvClaudeCodeMaxContextTokens,
	}
}

// glmLiveDirtyEnv seeds a settings env carrying every live key plus one
// user-owned key that must survive.
//
// MOAI_BACKUP_AUTH_TOKEN is deliberately left out: its presence tells the
// cleanup paths to RESTORE that value as ANTHROPIC_AUTH_TOKEN, which is the
// documented OAuth-preservation behaviour, not residue. The restore path has its
// own case below.
func glmLiveDirtyEnv() map[string]string {
	env := map[string]string{"CUSTOM_VAR": "keep_me"}
	for _, key := range liveSettingsAxisKeys() {
		if key == "MOAI_BACKUP_AUTH_TOKEN" {
			continue
		}
		env[key] = "seeded"
	}
	// The GLM-active indicator and the model overrides need realistic values so
	// the cleanup paths take their GLM branch rather than an early return.
	env[config.EnvAnthropicBaseURL] = "https://api.z.ai/api/anthropic"
	env[config.EnvAnthropicDefaultOpusModel] = "glm-5.3"
	return env
}

// assertFixtureIsDirty is the emptiness guard: without it a fixture that stopped
// seeding the keys would make every assertion below pass vacuously.
func assertFixtureIsDirty(t *testing.T, env map[string]string) {
	t.Helper()
	for _, key := range liveSettingsAxisKeys() {
		if key == "MOAI_BACKUP_AUTH_TOKEN" {
			continue // absent by design — see glmLiveDirtyEnv
		}
		if _, ok := env[key]; !ok {
			t.Fatalf("fixture is not dirty: %s absent before cleanup", key)
		}
	}
}

func assertLiveKeysCleared(t *testing.T, who string, env map[string]string) {
	t.Helper()
	for _, key := range liveSettingsAxisKeys() {
		if v, ok := env[key]; ok {
			t.Errorf("%s left live key %s=%q in settings.local.json", who, key, v)
		}
	}
}

// TestRemoveGLMEnvClearsEveryLiveKey covers the `moai cc` cleanup path.
func TestRemoveGLMEnvClearsEveryLiveKey(t *testing.T) {
	dirty := glmLiveDirtyEnv()
	assertFixtureIsDirty(t, dirty)
	_, settingsPath := cgTestProject(t, dirty)

	if err := removeGLMEnv(settingsPath); err != nil {
		t.Fatalf("removeGLMEnv: %v", err)
	}

	got := readSettingsForTest(t, settingsPath)
	assertLiveKeysCleared(t, "removeGLMEnv", got.Env)
	if got.Env["CUSTOM_VAR"] != "keep_me" {
		t.Errorf("user key must survive cleanup, got env: %v", got.Env)
	}
}

// TestStripGLMCredsClearsEveryLiveKey covers the CG leader cleanup path.
func TestStripGLMCredsClearsEveryLiveKey(t *testing.T) {
	dirty := glmLiveDirtyEnv()
	assertFixtureIsDirty(t, dirty)
	_, settingsPath := cgTestProject(t, dirty)

	if err := mutateSettingsLocal(settingsPath, stripGLMCredsAndSetTeammateMode); err != nil {
		t.Fatalf("mutateSettingsLocal: %v", err)
	}

	got := readSettingsForTest(t, settingsPath)
	assertLiveKeysCleared(t, "stripGLMCredsAndSetTeammateMode", got.Env)
	if got.Env["CUSTOM_VAR"] != "keep_me" {
		t.Errorf("user key must survive cleanup, got env: %v", got.Env)
	}
	if got.TeammateMode != "tmux" {
		t.Errorf("CG leader teammateMode must stay tmux, got %q", got.TeammateMode)
	}
}

// TestGLMCleanupRestoresBackedUpAuthToken is the counterpart case: when a
// MOAI_BACKUP_AUTH_TOKEN is present the cleanup paths must restore it as
// ANTHROPIC_AUTH_TOKEN rather than delete it, and the backup key itself must go.
// Without this case the "every live key cleared" guard above could be satisfied
// by a cleanup that destroys the user's OAuth token.
func TestGLMCleanupRestoresBackedUpAuthToken(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cleanup func(t *testing.T, settingsPath string)
	}{
		{"removeGLMEnv", func(t *testing.T, settingsPath string) {
			if err := removeGLMEnv(settingsPath); err != nil {
				t.Fatalf("removeGLMEnv: %v", err)
			}
		}},
		{"stripGLMCredsAndSetTeammateMode", func(t *testing.T, settingsPath string) {
			if err := mutateSettingsLocal(settingsPath, stripGLMCredsAndSetTeammateMode); err != nil {
				t.Fatalf("mutateSettingsLocal: %v", err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dirty := glmLiveDirtyEnv()
			dirty["MOAI_BACKUP_AUTH_TOKEN"] = "user-oauth-token"
			dirty[config.EnvAnthropicAuthToken] = "glm-key"
			_, settingsPath := cgTestProject(t, dirty)

			tc.cleanup(t, settingsPath)

			got := readSettingsForTest(t, settingsPath)
			if got.Env[config.EnvAnthropicAuthToken] != "user-oauth-token" {
				t.Errorf("backed-up token must be restored, got %q", got.Env[config.EnvAnthropicAuthToken])
			}
			if v, ok := got.Env["MOAI_BACKUP_AUTH_TOKEN"]; ok {
				t.Errorf("backup key must be consumed, got %q", v)
			}
			// Everything else on the live set still leaves.
			for _, key := range liveSettingsAxisKeys() {
				switch key {
				case config.EnvAnthropicAuthToken, "MOAI_BACKUP_AUTH_TOKEN":
					continue
				}
				if v, ok := got.Env[key]; ok {
					t.Errorf("%s left live key %s=%q", tc.name, key, v)
				}
			}
		})
	}
}

// TestTmuxClearVarsCoversEveryLiveKeyItOwns keeps the two axes in step. The
// tmux-axis list was already complete when the settings axis was not, which is
// what made the gap a divergence between two hand-enumerated lists rather than
// an unknown key set. ANTHROPIC_AUTH_TOKEN is excluded there by documented
// intent (it may be an OAuth token that must survive mode switches).
func TestTmuxClearVarsCoversEveryLiveKeyItOwns(t *testing.T) {
	cleared := map[string]bool{}
	for _, key := range buildTmuxClearVars() {
		cleared[key] = true
	}
	for _, key := range liveSettingsAxisKeys() {
		switch key {
		case config.EnvAnthropicAuthToken, "MOAI_BACKUP_AUTH_TOKEN":
			continue // settings-axis only; excluded from the tmux list by design
		}
		if !cleared[key] {
			t.Errorf("buildTmuxClearVars does not clear %s", key)
		}
	}
}
