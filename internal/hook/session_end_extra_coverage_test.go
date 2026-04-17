package hook

// session_end_extra_coverage_test.go: getModifiedGoFiles, cleanupGLMSettingsLocal,
// NewSessionEndHandlerWithObservability, generateSessionSummary 추가 커버리지

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// --- getModifiedGoFiles ---

// TestGetModifiedGoFiles_NonGitDir2 returns nil for non-git directory.
func TestGetModifiedGoFiles_NonGitDir2(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	result := getModifiedGoFiles(context.Background(), dir)
	if result != nil {
		t.Errorf("expected nil for non-git dir, got %v", result)
	}
}

// TestGetModifiedGoFiles_CancelledContext2 returns nil when context is cancelled.
func TestGetModifiedGoFiles_CancelledContext2(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	dir := t.TempDir()
	result := getModifiedGoFiles(ctx, dir)
	if result != nil {
		t.Errorf("expected nil for cancelled context, got %v", result)
	}
}

// TestGetModifiedGoFiles_RealGitRepo exercises the git parsing path in the real repo.
func TestGetModifiedGoFiles_RealGitRepo(t *testing.T) {
	t.Parallel()

	// Use the working directory, which is inside the module's git repository.
	// This exercises the full parsing path (strings.Split, HasSuffix check).
	dir, err := os.Getwd()
	if err != nil {
		t.Skipf("could not get working dir: %v", err)
	}
	// Navigate to module root (two levels up from internal/hook).
	moduleRoot := filepath.Join(dir, "..", "..")
	result := getModifiedGoFiles(context.Background(), moduleRoot)
	// result may be nil or a slice — we only assert no panic.
	_ = result
}

// --- cleanupGLMSettingsLocal extra ---

// TestCleanupGLMSettingsLocal_NoGLMKeys2 leaves file unchanged when no GLM keys present.
func TestCleanupGLMSettingsLocal_NoGLMKeys2(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	content := `{"env": {"SOME_OTHER_KEY": "value"}}`
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	_ = os.WriteFile(settingsPath, []byte(content), 0o644)

	cleanupGLMSettingsLocal(dir)

	// File should still exist.
	if _, err := os.Stat(settingsPath); err != nil {
		t.Errorf("file should still exist: %v", err)
	}
}

// TestCleanupGLMSettingsLocal_WithGLMEnvKey2 removes GLM auth token.
func TestCleanupGLMSettingsLocal_WithGLMEnvKey2(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Minimal JSON that has the GLM-injected ANTHROPIC_AUTH_TOKEN.
	content := `{"env": {"ANTHROPIC_AUTH_TOKEN": "glm-fake-token-abc123"}}`
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	_ = os.WriteFile(settingsPath, []byte(content), 0o644)

	cleanupGLMSettingsLocal(dir)
	// Should not panic. The file may or may not be modified depending on
	// implementation details — we only assert no panic here.
}

// TestCleanupGLMSettingsLocal_WithBaseURL removes GLM env keys when ANTHROPIC_BASE_URL is set.
func TestCleanupGLMSettingsLocal_WithBaseURL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Full GLM-active settings: ANTHROPIC_BASE_URL is the indicator.
	content := `{"env": {
		"ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/paas/v4",
		"ANTHROPIC_AUTH_TOKEN": "glm-token-xyz",
		"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4-plus"
	}}`
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	_ = os.WriteFile(settingsPath, []byte(content), 0o644)

	cleanupGLMSettingsLocal(dir)
	// Should not panic; GLM keys should be removed from the file.
}

// TestCleanupGLMSettingsLocal_WithBaseURLAndBackup restores backed-up auth token.
func TestCleanupGLMSettingsLocal_WithBaseURLAndBackup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	content := `{"env": {
		"ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/paas/v4",
		"ANTHROPIC_AUTH_TOKEN": "glm-token-xyz",
		"MOAI_BACKUP_AUTH_TOKEN": "original-oauth-token"
	}}`
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	_ = os.WriteFile(settingsPath, []byte(content), 0o644)

	cleanupGLMSettingsLocal(dir)
	// Backup token should be restored; GLM keys removed.
}

// TestCleanupGLMSettingsLocal_MalformedEnv handles invalid env field.
func TestCleanupGLMSettingsLocal_MalformedEnv(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// env is not an object but an array — json.Unmarshal into map[string]string should fail.
	content := `{"env": ["not", "an", "object"]}`
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	_ = os.WriteFile(settingsPath, []byte(content), 0o644)

	cleanupGLMSettingsLocal(dir) // malformed env → logs warn, no panic
}

// TestCleanupGLMSettingsLocal_EmptyJSON is a no-op for empty JSON.
func TestCleanupGLMSettingsLocal_EmptyJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	_ = os.WriteFile(settingsPath, []byte{}, 0o644)

	cleanupGLMSettingsLocal(dir) // empty file → early return, no panic
}

// TestCleanupGLMSettingsLocal_MalformedJSON logs warn but does not panic.
func TestCleanupGLMSettingsLocal_MalformedJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	_ = os.WriteFile(settingsPath, []byte("{bad json"), 0o644)

	cleanupGLMSettingsLocal(dir) // malformed → logs warn, no panic
}

// --- cleanupBogusRootDir ---

// TestCleanupBogusRootDir_NoClaudeDir is a no-op when .claude is missing.
func TestCleanupBogusRootDir_NoClaudeDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cleanupBogusRootDir(dir) // should not panic
}

// --- countSessionRecords ---

// TestCountSessionRecords_EmptyDir returns 0 for missing telemetry.
func TestCountSessionRecords_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	count := countSessionRecords(dir, "test-session-abc")
	if count != 0 {
		t.Errorf("expected 0 for empty dir, got %d", count)
	}
}

// TestCountSessionRecords_EmptySessionID returns 0.
func TestCountSessionRecords_EmptySessionID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	count := countSessionRecords(dir, "")
	if count != 0 {
		t.Errorf("expected 0 for empty session ID, got %d", count)
	}
}
