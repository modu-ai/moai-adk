package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Card t802 regression guard for the SessionEnd half of the pair. This function
// and ensureGLMCredentials are two halves of one contract: whatever the hook
// writes on session start, the hook removes on session end. Runs in the default
// suite on purpose — see internal/cli/glm_settings_cleanup_test.go for why.

// liveHookWrittenKeys is every env key ensureGLMCredentials can write, plus the
// OAuth-backup key this cleanup owns.
func liveHookWrittenKeys() []string {
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

// scrubGatewayEnv neutralises MOAI_LAUNCH_PROVIDER for the duration of the test.
// cleanupGLMSettingsLocal returns immediately when isGatewaySession() is true,
// and that variable is set in every launcher-started session — including the
// factory-lane sessions this suite is usually run from. Without the scrub the
// assertions below would pass or fail on the ambient environment rather than on
// the code, which is how an env-dependent test reports a defect that is not
// there (and hides one that is).
func scrubGatewayEnv(t *testing.T) {
	t.Helper()
	t.Setenv(config.EnvMoaiLaunchProvider, "")
}

func writeCleanupFixture(t *testing.T, env map[string]string) (string, string) {
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

func readCleanupEnv(t *testing.T, path string) map[string]string {
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

// TestCleanupGLMSettingsLocalClearsEveryLiveKey is the guard proper.
func TestCleanupGLMSettingsLocalClearsEveryLiveKey(t *testing.T) {
	scrubGatewayEnv(t)
	// MOAI_BACKUP_AUTH_TOKEN is deliberately left out: its presence tells the
	// cleanup to RESTORE that value as ANTHROPIC_AUTH_TOKEN, which is the
	// documented OAuth-preservation behaviour, not residue. The restore path has
	// its own case below.
	dirty := map[string]string{"CUSTOM_VAR": "keep_me"}
	for _, key := range liveHookWrittenKeys() {
		if key == "MOAI_BACKUP_AUTH_TOKEN" {
			continue
		}
		dirty[key] = "seeded"
	}
	// Realistic values so the cleanup takes its GLM branch rather than an early
	// return on the ANTHROPIC_BASE_URL indicator.
	dirty[config.EnvAnthropicBaseURL] = "https://api.z.ai/api/anthropic"
	dirty[config.EnvAnthropicDefaultOpusModel] = "glm-5.3"

	// Emptiness guard: without this a fixture that stopped seeding the keys would
	// make the assertions below pass vacuously.
	for _, key := range liveHookWrittenKeys() {
		if key == "MOAI_BACKUP_AUTH_TOKEN" {
			continue
		}
		if _, ok := dirty[key]; !ok {
			t.Fatalf("fixture is not dirty: %s absent before cleanup", key)
		}
	}

	root, settingsPath := writeCleanupFixture(t, dirty)

	cleanupGLMSettingsLocal(root)

	env := readCleanupEnv(t, settingsPath)
	for _, key := range liveHookWrittenKeys() {
		if v, ok := env[key]; ok {
			t.Errorf("cleanupGLMSettingsLocal left live key %s=%q", key, v)
		}
	}
	if env["CUSTOM_VAR"] != "keep_me" {
		t.Errorf("user key must survive cleanup, got env: %v", env)
	}
}

// TestCleanupGLMSettingsLocalRestoresBackedUpAuthToken is the counterpart case:
// a present MOAI_BACKUP_AUTH_TOKEN must be restored as ANTHROPIC_AUTH_TOKEN and
// the backup key consumed. Without it the guard above could be satisfied by a
// cleanup that destroys the user's OAuth token.
func TestCleanupGLMSettingsLocalRestoresBackedUpAuthToken(t *testing.T) {
	scrubGatewayEnv(t)
	root, settingsPath := writeCleanupFixture(t, map[string]string{
		config.EnvAnthropicAuthToken:          "glm-key",
		"MOAI_BACKUP_AUTH_TOKEN":              "user-oauth-token",
		config.EnvAnthropicBaseURL:            "https://api.z.ai/api/anthropic",
		config.EnvAnthropicDefaultOpusModel:   "glm-5.3",
		config.EnvClaudeCodeMaxContextTokens:  "1000000",
		config.EnvClaudeCodeAutoCompactWindow: "1000000",
	})

	cleanupGLMSettingsLocal(root)

	env := readCleanupEnv(t, settingsPath)
	if env[config.EnvAnthropicAuthToken] != "user-oauth-token" {
		t.Errorf("backed-up token must be restored, got %q", env[config.EnvAnthropicAuthToken])
	}
	if v, ok := env["MOAI_BACKUP_AUTH_TOKEN"]; ok {
		t.Errorf("backup key must be consumed, got %q", v)
	}
	for _, key := range []string{
		config.EnvAnthropicBaseURL,
		config.EnvAnthropicDefaultOpusModel,
		config.EnvClaudeCodeMaxContextTokens,
		config.EnvClaudeCodeAutoCompactWindow,
	} {
		if v, ok := env[key]; ok {
			t.Errorf("cleanupGLMSettingsLocal left live key %s=%q", key, v)
		}
	}
}

// TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone is the negative control: the
// ANTHROPIC_BASE_URL gate must keep this cleanup from touching a file that was
// never in GLM mode. It is the property that makes the gate worth keeping, and
// the reason removing the gate is a separate design decision rather than part of
// this repair.
func TestCleanupGLMSettingsLocalLeavesNonGLMFileAlone(t *testing.T) {
	scrubGatewayEnv(t)
	root, settingsPath := writeCleanupFixture(t, map[string]string{
		config.EnvAnthropicAuthToken: "user-oauth-or-api-key",
		"CUSTOM_VAR":                 "keep_me",
	})

	cleanupGLMSettingsLocal(root)

	env := readCleanupEnv(t, settingsPath)
	if env[config.EnvAnthropicAuthToken] != "user-oauth-or-api-key" {
		t.Errorf("non-GLM file must be left untouched, got env: %v", env)
	}
	if env["CUSTOM_VAR"] != "keep_me" {
		t.Errorf("non-GLM file must be left untouched, got env: %v", env)
	}
}
