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
// (The injector half, injectGLMEnvForTeam, was removed with its dead caller
// enableTeamMode in #1531; only the removeGLMEnv restore half is exercised
// here now.)

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

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
