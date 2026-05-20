// Package hook tests verify that settings.local.json — which may store
// sensitive GLM credentials (ANTHROPIC_AUTH_TOKEN) — is always persisted
// with mode 0o600. This locks AC-SEC-001, AC-SEC-002 of
// SPEC-V3R5-SECURITY-CRIT-001 and prevents regression to the v2.14.0→HEAD
// 0o644 vulnerability (CWE-732/552).
package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// expectedSettingsMode is the only mode tolerated for settings.local.json.
// On POSIX, 0o600 means owner-read+write, no group, no world.
const expectedSettingsMode os.FileMode = 0o600

// statSettings reads the mode bits from settings.local.json. Returns an
// error chain rather than panic so that t.Helper can be used.
func statSettings(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.Mode().Perm()
}

// TestEnsureGLMCredentialsFilePerm covers AC-SEC-001:
// ensureGLMCredentials writes settings.local.json with 0o600 when injecting
// ANTHROPIC_AUTH_TOKEN.
func TestEnsureGLMCredentialsFilePerm(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file-mode test skipped on Windows (ACL model differs)")
	}

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir claude: %v", err)
	}

	moaiDir := filepath.Join(dir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatalf("mkdir moai: %v", err)
	}

	// Pre-seed ~/.moai/.env.glm via HOME override so loadGLMKeyFromEnvFile
	// can pick it up without touching the real user home.
	t.Setenv("HOME", dir)

	homeMoai := filepath.Join(dir, ".moai")
	if err := os.MkdirAll(homeMoai, 0o700); err != nil {
		t.Fatalf("mkdir home moai: %v", err)
	}
	envGlm := filepath.Join(homeMoai, ".env.glm")
	if err := os.WriteFile(envGlm, []byte(`GLM_API_KEY="test-key-12345"`+"\n"), 0o600); err != nil {
		t.Fatalf("write .env.glm: %v", err)
	}

	// Seed settings.local.json with a GLM model marker but no AUTH_TOKEN so
	// ensureGLMCredentials enters the injection branch.
	settings := map[string]any{
		"env": map[string]string{
			"ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-4.7",
		},
	}
	data, err := json.MarshalIndent(settings, "", " ")
	if err != nil {
		t.Fatalf("marshal seed: %v", err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	// Seed with 0o644 to confirm the function CORRECTS the permission.
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	msg := ensureGLMCredentials(dir)
	if msg == "" {
		t.Fatalf("expected credential injection message, got empty (env file: %s, exists: %v)",
			envGlm, fileExists(envGlm))
	}

	got := statSettings(t, settingsPath)
	if got != expectedSettingsMode {
		t.Errorf("settings.local.json mode = %#o, want %#o (AC-SEC-001 regression)",
			got, expectedSettingsMode)
	}

	// Confirm AUTH_TOKEN was actually injected (sanity check).
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings after inject: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse settings after inject: %v", err)
	}
	envMap, _ := parsed["env"].(map[string]any)
	if envMap == nil || envMap["ANTHROPIC_AUTH_TOKEN"] != "test-key-12345" {
		t.Errorf("AUTH_TOKEN not injected; env=%v", envMap)
	}
}

// TestEnsureTeammateModeFilePerm covers AC-SEC-001 for the teammateMode
// write path in session_start.go (line 403 baseline).
func TestEnsureTeammateModeFilePerm(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file-mode test skipped on Windows")
	}

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir claude: %v", err)
	}

	// Simulate "inside tmux" so the write path activates.
	t.Setenv("TMUX", "/tmp/fake-tmux,1234,0")

	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	// Seed with 0o644 (insecure) — we expect the function to correct to 0o600.
	if err := os.WriteFile(settingsPath, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	mode := ensureTeammateMode(dir)
	if mode != "tmux" {
		t.Fatalf("expected mode=tmux, got %q", mode)
	}

	got := statSettings(t, settingsPath)
	if got != expectedSettingsMode {
		t.Errorf("settings.local.json mode = %#o, want %#o (AC-SEC-001 regression)",
			got, expectedSettingsMode)
	}
}

// TestInjectCLAUDEEnvFilePerm covers AC-SEC-001 for the CLAUDE_ENV_FILE
// write path in session_start.go (line 642 baseline).
func TestInjectCLAUDEEnvFilePerm(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file-mode test skipped on Windows")
	}

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir claude: %v", err)
	}

	// .env file must exist for injectCLAUDEEnvFile to act.
	envFile := filepath.Join(dir, ".env")
	if err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatalf("seed .env: %v", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if err := os.WriteFile(settingsPath, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	msg := injectCLAUDEEnvFile(dir)
	if msg == "" {
		t.Fatalf("expected CLAUDE_ENV_FILE injection message, got empty")
	}

	got := statSettings(t, settingsPath)
	if got != expectedSettingsMode {
		t.Errorf("settings.local.json mode = %#o, want %#o (AC-SEC-001 regression)",
			got, expectedSettingsMode)
	}
}

// TestSessionEndSettingsPerm covers AC-SEC-002:
// the cleanup write-back at session_end.go:667 also persists 0o600.
func TestSessionEndSettingsPerm(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file-mode test skipped on Windows")
	}

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("mkdir claude: %v", err)
	}

	// Seed settings with GLM env vars so removeGLMEnvFromSettings has
	// something to clean up. Use 0o644 to confirm cleanup corrects the mode.
	settings := map[string]any{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN":         "leaked-token-abc",
			"ANTHROPIC_BASE_URL":           "https://example.com",
			"ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-4.7",
		},
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatalf("marshal seed: %v", err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	cleanupGLMSettingsLocal(dir)

	got := statSettings(t, settingsPath)
	if got != expectedSettingsMode {
		t.Errorf("settings.local.json mode after cleanup = %#o, want %#o (AC-SEC-002 regression)",
			got, expectedSettingsMode)
	}

	// Verify token was actually removed (sanity).
	raw, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read after cleanup: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse after cleanup: %v", err)
	}
	if env, ok := parsed["env"].(map[string]any); ok {
		if _, leaked := env["ANTHROPIC_AUTH_TOKEN"]; leaked {
			t.Errorf("ANTHROPIC_AUTH_TOKEN was not removed by session_end cleanup")
		}
	}
}

// TestNoSettingsLocalJSONWith0o644 covers AC-SEC-003: a static (grep-style)
// guard. Reads source files and asserts no os.WriteFile(settingsPath, …, 0o644)
// occurrence remains in scope. Pre-existing 0o644 writes for non-sensitive
// paths (reportPath, dstPath under copyDirRecursive) are explicitly allowed.
func TestNoSettingsLocalJSONWith0o644(t *testing.T) {
	t.Parallel()

	targets := []struct {
		path           string
		mustNotContain []string // substrings forbidden in the file
	}{
		{
			path: "session_start.go",
			mustNotContain: []string{
				// settingsPath writes — any of these patterns indicate regression.
				`os.WriteFile(settingsPath, newData, 0o644)`,
				`os.WriteFile(settingsPath, out, 0o644)`,
			},
		},
		{
			path: "session_end.go",
			mustNotContain: []string{
				`os.WriteFile(settingsPath, out, 0o644)`,
				`os.WriteFile(settingsPath, newData, 0o644)`,
			},
		},
	}

	for _, tt := range targets {
		t.Run(tt.path, func(t *testing.T) {
			b, err := os.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("read %s: %v", tt.path, err)
			}
			body := string(b)
			for _, banned := range tt.mustNotContain {
				if strings.Contains(body, banned) {
					t.Errorf("forbidden 0o644 settings.local.json write detected in %s:\n  pattern: %q",
						tt.path, banned)
				}
			}
		})
	}
}

// fileExists is a tiny helper for diagnostic messages.
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
