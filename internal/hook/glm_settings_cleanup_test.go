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
//
// It DERIVES from config.SettingsAxisLiveKeys() — the live view of the canonical
// settings-axis declaration — rather than re-spelling the list. The live view is
// the right one here and not the cleanup view, because this list states what the
// live PRODUCER can write: it is the other half of the t802 contract that
// whatever ensureGLMCredentials adds on session start, the cleanup removes on
// session end. What the cleanup deletes beyond that — the legacy tail an older
// binary could leave — is the cleanup view's business and is asserted separately
// by TestCleanupGLMSettingsLocalCleanupViewEquivalence.
func liveHookWrittenKeys() []string {
	return config.SettingsAxisLiveKeys()
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

// TestCleanupGLMSettingsLocalAdmitsBackupOnlyFileAndRestoresToken is the net
// benefit of the REQ-5 indicator widening, stated as an observation rather than
// an argument. A settings file carrying MOAI_BACKUP_AUTH_TOKEN and a GLM key in
// ANTHROPIC_AUTH_TOKEN but NO ANTHROPIC_BASE_URL was declined by the old
// single-key gate, so the user's backed-up OAuth token stayed stranded and the
// GLM key survived as their credential. Under the two-key indicator the file is
// admitted and the token is put back.
func TestCleanupGLMSettingsLocalAdmitsBackupOnlyFileAndRestoresToken(t *testing.T) {
	scrubGatewayEnv(t)
	root, settingsPath := writeCleanupFixture(t, map[string]string{
		config.EnvMoaiBackupAuthToken: "user-oauth-token",
		config.EnvAnthropicAuthToken:  "glm-key",
	})

	cleanupGLMSettingsLocal(root)

	env := readCleanupEnv(t, settingsPath)
	if env[config.EnvAnthropicAuthToken] != "user-oauth-token" {
		t.Errorf("backup-only file must be admitted and its token restored, got %s=%q",
			config.EnvAnthropicAuthToken, env[config.EnvAnthropicAuthToken])
	}
	if v, ok := env[config.EnvMoaiBackupAuthToken]; ok {
		t.Errorf("backup key must be consumed, got %s=%q", config.EnvMoaiBackupAuthToken, v)
	}
}

// TestCleanupGLMSettingsLocalIndicatorSet pins the indicator set at exactly two
// keys, in both directions. The exclusion half is the half that protects a
// non-GLM user: MOAI_STATUSLINE_CONTEXT_SIZE and the context-window pair are
// documented user-settable overrides, and admitting one of them would send a
// file carrying the user's own ANTHROPIC_AUTH_TOKEN into the restore branch's
// `else`, deleting that key. A test asserting only the admitted direction would
// not catch that.
func TestCleanupGLMSettingsLocalIndicatorSet(t *testing.T) {
	scrubGatewayEnv(t)

	const userToken = "user-own-token"

	cases := []struct {
		name      string
		indicator string
		admitted  bool
	}{
		{name: "base URL admits", indicator: config.EnvAnthropicBaseURL, admitted: true},
		{name: "backup token admits", indicator: config.EnvMoaiBackupAuthToken, admitted: true},
		{name: "statusline context size excluded", indicator: config.EnvStatuslineContextSize},
		{name: "auto compact window excluded", indicator: config.EnvClaudeCodeAutoCompactWindow},
		{name: "max context tokens excluded", indicator: config.EnvClaudeCodeMaxContextTokens},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scrubGatewayEnv(t)
			// The candidate is the only axis key in the file besides the token,
			// so the verdict below is attributable to it and nothing else. The
			// user's own token rides along as what is at stake.
			root, settingsPath := writeCleanupFixture(t, map[string]string{
				tc.indicator:                 "seeded",
				config.EnvAnthropicAuthToken: userToken,
				"CUSTOM_VAR":                 "keep_me",
			})

			cleanupGLMSettingsLocal(root)

			env := readCleanupEnv(t, settingsPath)
			if tc.admitted {
				if v, ok := env[tc.indicator]; ok && tc.indicator != config.EnvMoaiBackupAuthToken {
					t.Errorf("admitted file must be cleaned, but %s=%q survived", tc.indicator, v)
				}
				if v, ok := env[config.EnvMoaiBackupAuthToken]; ok {
					t.Errorf("admitted file must consume the backup key, got %q", v)
				}
			} else {
				// Excluded: the file is left untouched, the user's own token
				// included. This assertion fails if the key is ever promoted
				// into the indicator set.
				if v := env[tc.indicator]; v != "seeded" {
					t.Errorf("excluded indicator must not admit the file, but %s is now %q", tc.indicator, v)
				}
				if v := env[config.EnvAnthropicAuthToken]; v != userToken {
					t.Errorf("excluded indicator must leave the user's own token alone, got %q", v)
				}
			}
			if env["CUSTOM_VAR"] != "keep_me" {
				t.Errorf("user key must survive either way, got env: %v", env)
			}
		})
	}
}

// cleanupViewDirtyEnv seeds a settings env carrying every key of the canonical
// cleanup view plus one user-owned key that must survive.
//
// MOAI_BACKUP_AUTH_TOKEN is deliberately left out: its presence tells the
// cleanup paths to RESTORE that value as ANTHROPIC_AUTH_TOKEN, which is the
// documented OAuth-preservation behaviour, not residue — and since
// ANTHROPIC_AUTH_TOKEN is itself a member of the view, seeding the backup key
// would make the whole-view "every key is absent" traversal below false against
// a CORRECT implementation. The restore path has its own named cases
// (TestCleanupGLMSettingsLocalRestoresBackedUpAuthToken and
// TestCleanupGLMSettingsLocalAdmitsBackupOnlyFileAndRestoresToken).
func cleanupViewDirtyEnv() map[string]string {
	env := map[string]string{"CUSTOM_VAR": "keep_me"}
	for _, key := range config.SettingsAxisCleanupKeys() {
		if key == config.EnvMoaiBackupAuthToken {
			continue
		}
		env[key] = "seeded"
	}
	// The GLM-active indicator and the model override need realistic values so
	// the cleanup takes its GLM branch rather than an early return. With the
	// backup key absent by design, ANTHROPIC_BASE_URL is what admits the file.
	env[config.EnvAnthropicBaseURL] = "https://api.z.ai/api/anthropic"
	env[config.EnvAnthropicDefaultOpusModel] = "glm-5.3"
	return env
}

// assertFixtureIsDirty is the emptiness guard: without it a fixture that stopped
// seeding the keys would make every assertion below pass vacuously.
func assertFixtureIsDirty(t *testing.T, env map[string]string) {
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

// TestCleanupGLMSettingsLocalCleanupViewEquivalence is AC-002 for consumer C:
// the runtime proof that cleanupGLMSettingsLocal deletes exactly the canonical
// cleanup view. It is the primary instrument for the five-lists-to-one collapse,
// because a textual scan cannot distinguish a consumer that kept its entire
// hand-written list from one that routed onto the declaration — both spell their
// keys as Go identifiers.
//
// The assertion iterates config.SettingsAxisCleanupKeys() rather than a list
// written here, so a key added to the canonical declaration widens this test
// without a second edit.
func TestCleanupGLMSettingsLocalCleanupViewEquivalence(t *testing.T) {
	scrubGatewayEnv(t)

	dirty := cleanupViewDirtyEnv()
	assertFixtureIsDirty(t, dirty)

	root, settingsPath := writeCleanupFixture(t, dirty)

	cleanupGLMSettingsLocal(root)

	env := readCleanupEnv(t, settingsPath)
	for _, key := range config.SettingsAxisCleanupKeys() {
		if v, ok := env[key]; ok {
			t.Errorf("cleanupGLMSettingsLocal left cleanup-view key %s=%q", key, v)
		}
	}
	if env["CUSTOM_VAR"] != "keep_me" {
		t.Errorf("user key must survive cleanup, got env: %v", env)
	}
}
