package cli

// TestOAuthTokenPreservation: Reproduction test for the ANTHROPIC_AUTH_TOKEN loss bug.
//
// Bug: Users who manually set ANTHROPIC_AUTH_TOKEN in settings.local.json
// (e.g. an Anthropic API key) lose it after a moai glm → moai cc round-trip.
//
// Note: Claude Code OAuth tokens are stored in ~/.claude/ (internal credential
// storage), NOT in settings.local.json. OAuth users are unaffected by this bug.
// This backup logic protects API-key users who configure ANTHROPIC_AUTH_TOKEN
// directly in settings.local.json.
//
// Root cause:
//  1. moai glm writes ANTHROPIC_AUTH_TOKEN=<GLM_KEY> to .claude/settings.local.json,
//     overwriting any pre-existing API key stored there.
//  2. moai cc calls removeGLMEnv() which deletes ANTHROPIC_AUTH_TOKEN entirely,
//     never restoring the original API key.
//  3. On next Claude Code startup, the user's API key is gone.
//
// Fix: When injecting GLM env vars, back up any existing ANTHROPIC_AUTH_TOKEN
// as MOAI_BACKUP_AUTH_TOKEN. When removing GLM env vars, restore the backup.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// glmConfigForTest returns a minimal GLMConfigFromYAML for use in tests.
func glmConfigForTest() *GLMConfigFromYAML {
	return &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		Models: struct {
			High   string
			Medium string
			Low    string
		}{
			High:   "glm-5",
			Medium: "glm-4.7",
			Low:    "glm-4.7-flash",
		},
		EnvVar: "GLM_API_KEY",
	}
}

// TestOAuthToken_GLMOverwritesOAuthToken demonstrates the first part of the bug:
// moai glm overwrites a pre-existing ANTHROPIC_AUTH_TOKEN (OAuth token) in
// settings.local.json with the GLM API key, destroying the OAuth credential.
func TestOAuthToken_GLMOverwritesOAuthToken(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Simulate a user who has an OAuth token stored in settings.local.json.
	// This can happen when Claude Code's /login stores the credential locally.
	oauthToken := "claude-oauth-token-abc123"
	initialSettings := SettingsLocal{
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN": oauthToken,
		},
	}
	data, err := json.MarshalIndent(initialSettings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	glmKey := "glm-api-key-xyz789"
	glmConfig := glmConfigForTest()

	// Simulate moai glm: inject GLM env vars into settings.local.json.
	if err := injectGLMEnvForTeam(settingsPath, glmConfig, glmKey); err != nil {
		t.Fatalf("injectGLMEnvForTeam() error: %v", err)
	}

	// Read back settings.local.json
	readData, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var afterGLM SettingsLocal
	if err := json.Unmarshal(readData, &afterGLM); err != nil {
		t.Fatal(err)
	}

	// BUG DEMONSTRATED: the OAuth token is now replaced by the GLM key.
	// After the fix, the original OAuth token should be preserved as
	// MOAI_BACKUP_AUTH_TOKEN so it can be restored when moai cc runs.
	authToken := afterGLM.Env["ANTHROPIC_AUTH_TOKEN"]
	if authToken != glmKey {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q, want GLM key %q (showing token was injected)", authToken, glmKey)
	}

	// After the fix: the original OAuth token should be backed up.
	backupToken := afterGLM.Env["MOAI_BACKUP_AUTH_TOKEN"]
	if backupToken != oauthToken {
		t.Errorf("MOAI_BACKUP_AUTH_TOKEN = %q, want original OAuth token %q\n"+
			"FIX NEEDED: injectGLMEnvForTeam() must back up existing ANTHROPIC_AUTH_TOKEN "+
			"as MOAI_BACKUP_AUTH_TOKEN before overwriting it with the GLM key",
			backupToken, oauthToken)
	}
}

// TestOAuthToken_CCDoesNotRestoreOAuthToken demonstrates the second part of the bug:
// moai cc deletes ANTHROPIC_AUTH_TOKEN without restoring the original OAuth token,
// so the credential is permanently lost from settings.local.json.
func TestOAuthToken_CCDoesNotRestoreOAuthToken(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Simulate the state after moai glm has run:
	// - ANTHROPIC_AUTH_TOKEN = GLM key
	// - MOAI_BACKUP_AUTH_TOKEN = original OAuth token (set by the fix)
	oauthToken := "claude-oauth-token-abc123"
	glmKey := "glm-api-key-xyz789"
	settingsAfterGLM := SettingsLocal{
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN":   glmKey,
			"MOAI_BACKUP_AUTH_TOKEN": oauthToken,
			"ANTHROPIC_BASE_URL":     "https://api.z.ai/api/anthropic",
		},
	}
	data, err := json.MarshalIndent(settingsAfterGLM, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Simulate moai cc: remove GLM env vars.
	if err := removeGLMEnv(settingsPath); err != nil {
		t.Fatalf("removeGLMEnv() error: %v", err)
	}

	// Read back settings.local.json
	readData, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var afterCC SettingsLocal
	if err := json.Unmarshal(readData, &afterCC); err != nil {
		t.Fatal(err)
	}

	// BUG DEMONSTRATED: after moai cc, ANTHROPIC_AUTH_TOKEN is gone entirely.
	// After the fix: the original OAuth token should be restored.
	authToken := afterCC.Env["ANTHROPIC_AUTH_TOKEN"]
	if authToken != oauthToken {
		t.Errorf("after moai cc: ANTHROPIC_AUTH_TOKEN = %q, want restored OAuth token %q\n"+
			"FIX NEEDED: removeGLMEnv() must restore MOAI_BACKUP_AUTH_TOKEN as ANTHROPIC_AUTH_TOKEN "+
			"when removing GLM env vars",
			authToken, oauthToken)
	}

	// After the fix: MOAI_BACKUP_AUTH_TOKEN should be cleaned up.
	backupToken := afterCC.Env["MOAI_BACKUP_AUTH_TOKEN"]
	if backupToken != "" {
		t.Errorf("after moai cc: MOAI_BACKUP_AUTH_TOKEN = %q, want empty (should be cleaned up after restore)", backupToken)
	}
}

// TestOAuthToken_FullCycle verifies the complete moai glm -> moai cc round-trip
// correctly preserves and restores the OAuth token.
//
// This is the primary regression test for the /login-on-restart bug.
func TestOAuthToken_FullCycle(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Initial state: user has an OAuth token.
	oauthToken := "claude-oauth-token-persistent"
	initialSettings := SettingsLocal{
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN":         oauthToken,
			"CLAUDE_CODE_TEAMMATE_DISPLAY": "auto",
		},
	}
	data, err := json.MarshalIndent(initialSettings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	glmKey := "glm-api-key-roundtrip"
	glmConfig := glmConfigForTest()

	// Step 1: moai glm injects GLM env vars (should back up OAuth token).
	if err := injectGLMEnvForTeam(settingsPath, glmConfig, glmKey); err != nil {
		t.Fatalf("injectGLMEnvForTeam() error: %v", err)
	}

	readData, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var afterGLM SettingsLocal
	if err := json.Unmarshal(readData, &afterGLM); err != nil {
		t.Fatal(err)
	}

	// Verify GLM key is active.
	if afterGLM.Env["ANTHROPIC_AUTH_TOKEN"] != glmKey {
		t.Errorf("step1: ANTHROPIC_AUTH_TOKEN = %q, want GLM key %q", afterGLM.Env["ANTHROPIC_AUTH_TOKEN"], glmKey)
	}
	// Verify OAuth token is backed up.
	if afterGLM.Env["MOAI_BACKUP_AUTH_TOKEN"] != oauthToken {
		t.Errorf("step1: MOAI_BACKUP_AUTH_TOKEN = %q, want %q\n"+
			"FIX NEEDED: injectGLMEnvForTeam() must back up existing token",
			afterGLM.Env["MOAI_BACKUP_AUTH_TOKEN"], oauthToken)
	}

	// Step 2: moai cc removes GLM env vars (should restore OAuth token).
	if err := removeGLMEnv(settingsPath); err != nil {
		t.Fatalf("removeGLMEnv() error: %v", err)
	}

	readData, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var afterCC SettingsLocal
	if err := json.Unmarshal(readData, &afterCC); err != nil {
		t.Fatal(err)
	}

	// Verify OAuth token is restored.
	if afterCC.Env["ANTHROPIC_AUTH_TOKEN"] != oauthToken {
		t.Errorf("step2: ANTHROPIC_AUTH_TOKEN = %q, want restored OAuth token %q\n"+
			"FIX NEEDED: removeGLMEnv() must restore the backed-up token",
			afterCC.Env["ANTHROPIC_AUTH_TOKEN"], oauthToken)
	}
	// Verify backup key is cleaned up.
	if _, exists := afterCC.Env["MOAI_BACKUP_AUTH_TOKEN"]; exists {
		t.Error("step2: MOAI_BACKUP_AUTH_TOKEN should be deleted after restore")
	}
	// Verify other settings are preserved.
	if afterCC.Env["CLAUDE_CODE_TEAMMATE_DISPLAY"] != "auto" {
		t.Errorf("step2: CLAUDE_CODE_TEAMMATE_DISPLAY = %q, want %q",
			afterCC.Env["CLAUDE_CODE_TEAMMATE_DISPLAY"], "auto")
	}
}

// TestOAuthToken_NoExistingToken verifies that when there is no pre-existing
// ANTHROPIC_AUTH_TOKEN, the GLM/CC round-trip works correctly without creating
// a spurious backup key.
func TestOAuthToken_NoExistingToken(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Initial state: no ANTHROPIC_AUTH_TOKEN (fresh install or OAuth-only user).
	initialSettings := SettingsLocal{
		Env: map[string]string{
			"CLAUDE_CODE_TEAMMATE_DISPLAY": "auto",
		},
	}
	data, err := json.MarshalIndent(initialSettings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	glmKey := "glm-api-key-noexisting"
	glmConfig := glmConfigForTest()

	// Step 1: moai glm injects GLM key (no prior token to back up).
	if err := injectGLMEnvForTeam(settingsPath, glmConfig, glmKey); err != nil {
		t.Fatalf("injectGLMEnvForTeam() error: %v", err)
	}

	readData, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var afterGLM SettingsLocal
	if err := json.Unmarshal(readData, &afterGLM); err != nil {
		t.Fatal(err)
	}

	// No backup should be created when there was no pre-existing token.
	if _, exists := afterGLM.Env["MOAI_BACKUP_AUTH_TOKEN"]; exists {
		t.Errorf("MOAI_BACKUP_AUTH_TOKEN should NOT be set when there was no prior ANTHROPIC_AUTH_TOKEN, got: %q",
			afterGLM.Env["MOAI_BACKUP_AUTH_TOKEN"])
	}

	// GLM key should be present.
	if afterGLM.Env["ANTHROPIC_AUTH_TOKEN"] != glmKey {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q, want %q", afterGLM.Env["ANTHROPIC_AUTH_TOKEN"], glmKey)
	}

	// Step 2: moai cc removes GLM key (no backup to restore).
	if err := removeGLMEnv(settingsPath); err != nil {
		t.Fatalf("removeGLMEnv() error: %v", err)
	}

	readData, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var afterCC SettingsLocal
	if err := json.Unmarshal(readData, &afterCC); err != nil {
		t.Fatal(err)
	}

	// ANTHROPIC_AUTH_TOKEN should be gone (nothing to restore).
	if _, exists := afterCC.Env["ANTHROPIC_AUTH_TOKEN"]; exists {
		t.Errorf("ANTHROPIC_AUTH_TOKEN should be removed (no backup), got: %q",
			afterCC.Env["ANTHROPIC_AUTH_TOKEN"])
	}
}
