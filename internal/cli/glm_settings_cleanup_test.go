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

// cleanupViewDirtyEnv seeds a settings env carrying every key of the canonical
// cleanup view plus one user-owned key that must survive. It is the cleanup-view
// sibling of glmLiveDirtyEnv above, which stays on the LIVE view because the
// tmux-parity guard reads that one.
//
// MOAI_BACKUP_AUTH_TOKEN is deliberately left out, for the same reason it is
// left out of glmLiveDirtyEnv: its presence tells the cleanup paths to RESTORE
// that value as ANTHROPIC_AUTH_TOKEN, which is documented OAuth-preservation
// behaviour rather than residue — and since ANTHROPIC_AUTH_TOKEN is itself a
// member of the view, seeding the backup key would make the whole-view "every
// key is absent" traversal below false against a CORRECT implementation. The
// restore path keeps its own case, TestGLMCleanupRestoresBackedUpAuthToken.
func cleanupViewDirtyEnv() map[string]string {
	env := map[string]string{"CUSTOM_VAR": "keep_me"}
	for _, key := range config.SettingsAxisCleanupKeys() {
		if key == config.EnvMoaiBackupAuthToken {
			continue
		}
		env[key] = "seeded"
	}
	// The GLM-active indicator and the model override need realistic values so
	// the cleanup paths take their GLM branch rather than an early return.
	env[config.EnvAnthropicBaseURL] = "https://api.z.ai/api/anthropic"
	env[config.EnvAnthropicDefaultOpusModel] = "glm-5.3"
	return env
}

// assertCleanupViewFixtureIsDirty is the emptiness guard for the cleanup-view
// fixture: without it a fixture that stopped seeding the keys would make every
// assertion below pass vacuously. Mirrors assertFixtureIsDirty, which guards the
// live-view fixture.
func assertCleanupViewFixtureIsDirty(t *testing.T, env map[string]string) {
	t.Helper()
	for _, key := range config.SettingsAxisCleanupKeys() {
		if key == config.EnvMoaiBackupAuthToken {
			continue // absent by design — see cleanupViewDirtyEnv
		}
		if _, ok := env[key]; !ok {
			t.Fatalf("fixture is not dirty: %s absent before cleanup", key)
		}
	}
}

// assertCleanupViewCleared traverses the canonical view itself rather than a
// list written here, so a key added to the declaration widens every caller
// without a second edit.
func assertCleanupViewCleared(t *testing.T, who string, env map[string]string) {
	t.Helper()
	for _, key := range config.SettingsAxisCleanupKeys() {
		if v, ok := env[key]; ok {
			t.Errorf("%s left cleanup-view key %s=%q", who, key, v)
		}
	}
}

// TestStripGLMCredsCleanupViewEquivalence is AC-002 for consumer B: the runtime
// proof that stripGLMCredsAndSetTeammateMode deletes exactly the canonical
// cleanup view. It is the primary instrument for the five-lists-to-one collapse,
// because a textual scan cannot separate a consumer that kept its whole
// hand-written list from one that routed onto the declaration — both spell their
// keys as Go identifiers.
func TestStripGLMCredsCleanupViewEquivalence(t *testing.T) {
	dirty := cleanupViewDirtyEnv()
	assertCleanupViewFixtureIsDirty(t, dirty)
	_, settingsPath := cgTestProject(t, dirty)

	if err := mutateSettingsLocal(settingsPath, stripGLMCredsAndSetTeammateMode); err != nil {
		t.Fatalf("mutateSettingsLocal: %v", err)
	}

	got := readSettingsForTest(t, settingsPath)
	assertCleanupViewCleared(t, "stripGLMCredsAndSetTeammateMode", got.Env)
	if got.Env["CUSTOM_VAR"] != "keep_me" {
		t.Errorf("user key must survive cleanup, got env: %v", got.Env)
	}
	// The teammate-mode side effect is not part of the key axis and must not be
	// lost when the delete list is rerouted.
	if got.TeammateMode != "tmux" {
		t.Errorf("CG leader teammateMode must stay tmux, got %q", got.TeammateMode)
	}
}
