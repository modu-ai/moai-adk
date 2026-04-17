package hook

// R-P1-1 Handle-level regression tests for the Windows-only CLAUDE_ENV_FILE guard.
//
// Background (SPEC-OPUS47-COMPAT-001, T-016):
//   - injectCLAUDEEnvFile must ONLY be called when GOOS == "windows".
//   - macOS/Linux sessions must never execute that code path so GLM env
//     injection is never affected.
//   - runtime.GOOS is a compile-time constant and cannot be overridden via
//     os.Setenv. The claudeEnvFileGuard(goos string) helper extracted from
//     Handle() receives the OS name as a parameter, enabling string-based
//     unit tests without any build-tag splitting or interface mocking.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- Tests for claudeEnvFileGuard (R-P1-1 unit) ---

// TestClaudeEnvFileGuard_WindowsOnly verifies that the guard returns true
// only for "windows" and false for all other OS identifiers.
func TestClaudeEnvFileGuard_WindowsOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		goos string
		want bool
	}{
		{"windows", true},
		{"darwin", false},
		{"linux", false},
		{"freebsd", false},
		{"openbsd", false},
		{"plan9", false},
		{"", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.goos, func(t *testing.T) {
			t.Parallel()
			got := claudeEnvFileGuard(tt.goos)
			if got != tt.want {
				t.Errorf("claudeEnvFileGuard(%q) = %v, want %v", tt.goos, got, tt.want)
			}
		})
	}
}

// --- R-P1-1 Handle-level regression: non-Windows does NOT inject env file ---

// TestSessionStartHandler_Handle_NonWindowsGuard verifies that on macOS/Linux
// ("darwin") the Handle() call does NOT produce a "claude_env_file" data key,
// meaning injectCLAUDEEnvFile is blocked by claudeEnvFileGuard.
//
// We verify the behavior by:
//  1. Creating a project dir with a .env file (so injectCLAUDEEnvFile would
//     produce output if called).
//  2. Calling Handle() with that project dir.
//  3. Asserting that Data JSON does NOT contain "claude_env_file".
//
// On the actual running OS (darwin/linux in CI), the real runtime.GOOS guard
// is exercised. The claudeEnvFileGuard unit test above provides cross-platform
// coverage for all OS strings without re-running the entire Handle() chain.
func TestSessionStartHandler_Handle_NonWindowsGuard(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Create .env file — injectCLAUDEEnvFile would inject CLAUDE_ENV_FILE if reached.
	_ = os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET=value\n"), 0o644)

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:     "test-non-windows",
		CWD:           dir,
		HookEventName: "SessionStart",
		ProjectDir:    dir,
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}

	var data map[string]any
	if out.Data != nil {
		if err := json.Unmarshal(out.Data, &data); err != nil {
			t.Fatalf("unmarshal data: %v", err)
		}
	}

	// On non-Windows the key must be absent.
	// (On Windows CI this test would fail, but we intentionally skip verifying
	//  the Windows path here — that is covered by TestClaudeEnvFileGuard_WindowsOnly
	//  and TestInjectCLAUDEEnvFile_* tests below.)
	if _, ok := data["claude_env_file"]; ok {
		// Allow on Windows (where the guard should fire).
		if claudeEnvFileGuard("windows") && !claudeEnvFileGuard("darwin") {
			// We are on darwin/linux — fail.
			t.Error("claude_env_file key present on non-Windows: injectCLAUDEEnvFile was called but should have been blocked by claudeEnvFileGuard")
		}
	}
}

// TestSessionStartHandler_Handle_PreserveGLMEnvPath verifies that the GLM env
// injection path (ensureGLMCredentials + ensureTmuxGLMEnv) runs independently
// of the Windows guard and produces expected data keys when GLM model overrides
// are NOT present (no credentials injected, no error).
func TestSessionStartHandler_Handle_PreserveGLMEnvPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// settings.local.json with no GLM overrides — ensureGLMCredentials returns "".
	settings := `{"env":{"SOME_VAR":"hello"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

	h := NewSessionStartHandler(nil)
	input := &HookInput{
		SessionID:     "test-glm-path",
		CWD:           dir,
		HookEventName: "SessionStart",
		ProjectDir:    dir,
	}

	out, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if out == nil {
		t.Fatal("Handle() returned nil")
	}

	// Must return allow (empty decision = allow).
	var data map[string]any
	if out.Data != nil {
		if err := json.Unmarshal(out.Data, &data); err != nil {
			t.Fatalf("unmarshal data: %v", err)
		}
	}

	// session_id and status are always present.
	if _, ok := data["session_id"]; !ok {
		t.Error("data missing 'session_id' key")
	}
	if _, ok := data["status"]; !ok {
		t.Error("data missing 'status' key")
	}

	// glm_credentials must be absent — no injection happened.
	if _, ok := data["glm_credentials"]; ok {
		t.Error("glm_credentials present unexpectedly — GLM path was triggered without GLM models")
	}
}

// TestSessionStartHandler_Handle_WindowsOnlyInjection verifies the
// claudeEnvFileGuard contract: guard returns true for "windows" and the
// injectCLAUDEEnvFile function can inject CLAUDE_ENV_FILE when called directly.
// This test does NOT call Handle() with a real runtime.GOOS override; instead
// it exercises the two halves independently.
func TestSessionStartHandler_Handle_WindowsOnlyInjection(t *testing.T) {
	t.Parallel()

	// Part A: guard contract
	if !claudeEnvFileGuard("windows") {
		t.Error("claudeEnvFileGuard('windows') must return true")
	}
	if claudeEnvFileGuard("linux") {
		t.Error("claudeEnvFileGuard('linux') must return false")
	}
	if claudeEnvFileGuard("darwin") {
		t.Error("claudeEnvFileGuard('darwin') must return false")
	}

	// Part B: injectCLAUDEEnvFile works correctly when called directly
	// (simulating what Handle() would do on Windows).
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Create .env file.
	envFile := filepath.Join(dir, ".env")
	_ = os.WriteFile(envFile, []byte("KEY=val\n"), 0o644)

	msg := injectCLAUDEEnvFile(dir)
	if msg == "" {
		t.Error("injectCLAUDEEnvFile should return non-empty message when .env exists")
	}

	// Verify settings.local.json contains CLAUDE_ENV_FILE.
	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.local.json"))
	if err != nil {
		t.Fatalf("read settings.local.json: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse settings.local.json: %v", err)
	}
	var env map[string]string
	if err := json.Unmarshal(raw["env"], &env); err != nil {
		t.Fatalf("parse env: %v", err)
	}
	if env["CLAUDE_ENV_FILE"] != envFile {
		t.Errorf("CLAUDE_ENV_FILE = %q, want %q", env["CLAUDE_ENV_FILE"], envFile)
	}
}

// --- injectCLAUDEEnvFile error path tests ---

// TestInjectCLAUDEEnvFile_MissingEnvFile2 verifies that missing .env → returns "".
// (Variant to avoid duplicate with session_start_test.go:TestInjectCLAUDEEnvFile_NoEnvFile.)
func TestInjectCLAUDEEnvFile_MissingEnvFile2(t *testing.T) {
	t.Parallel()

	// Different project dir with no .env file; verifies no side effects.
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	msg := injectCLAUDEEnvFile(dir)
	if msg != "" {
		t.Errorf("no .env file: expected empty msg, got %q", msg)
	}
}

// TestInjectCLAUDEEnvFile_AlreadyInjected verifies idempotency:
// a second call with the same .env path returns "" (no write needed).
func TestInjectCLAUDEEnvFile_AlreadyInjected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	envFile := filepath.Join(dir, ".env")
	_ = os.WriteFile(envFile, []byte("K=V\n"), 0o644)

	// First call should inject.
	msg1 := injectCLAUDEEnvFile(dir)
	if msg1 == "" {
		t.Fatal("first injectCLAUDEEnvFile should succeed")
	}

	// Second call — already set to same value.
	msg2 := injectCLAUDEEnvFile(dir)
	if msg2 != "" {
		t.Errorf("second injectCLAUDEEnvFile should return empty (idempotent), got %q", msg2)
	}
}

// TestInjectCLAUDEEnvFile_ExistingEnvSection verifies that an existing "env"
// section is preserved and only CLAUDE_ENV_FILE is added.
func TestInjectCLAUDEEnvFile_ExistingEnvSection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Existing settings with an env var.
	existing := `{"env":{"EXISTING_KEY":"existing_value"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(existing), 0o644)

	envFile := filepath.Join(dir, ".env")
	_ = os.WriteFile(envFile, []byte("K=V\n"), 0o644)

	msg := injectCLAUDEEnvFile(dir)
	if msg == "" {
		t.Fatal("injectCLAUDEEnvFile should return non-empty")
	}

	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.local.json"))
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var env map[string]string
	if err := json.Unmarshal(raw["env"], &env); err != nil {
		t.Fatalf("parse env: %v", err)
	}

	if env["EXISTING_KEY"] != "existing_value" {
		t.Errorf("EXISTING_KEY lost: got %q", env["EXISTING_KEY"])
	}
	if env["CLAUDE_ENV_FILE"] != envFile {
		t.Errorf("CLAUDE_ENV_FILE = %q, want %q", env["CLAUDE_ENV_FILE"], envFile)
	}
}

// TestInjectCLAUDEEnvFile_MalformedJSON verifies graceful handling when
// settings.local.json contains invalid JSON.
func TestInjectCLAUDEEnvFile_MalformedJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte("{invalid json"), 0o644)

	envFile := filepath.Join(dir, ".env")
	_ = os.WriteFile(envFile, []byte("K=V\n"), 0o644)

	// Should still succeed — malformed JSON is treated as empty (raw = nil → new map).
	msg := injectCLAUDEEnvFile(dir)
	if msg == "" {
		t.Error("injectCLAUDEEnvFile should succeed even when existing JSON is malformed")
	}
}

// TestInjectCLAUDEEnvFile_EmptySettingsFile verifies behavior with an empty file.
func TestInjectCLAUDEEnvFile_EmptySettingsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Write an empty settings file.
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(""), 0o644)

	envFile := filepath.Join(dir, ".env")
	_ = os.WriteFile(envFile, []byte("K=V\n"), 0o644)

	msg := injectCLAUDEEnvFile(dir)
	if msg == "" {
		t.Error("injectCLAUDEEnvFile should succeed with empty settings file")
	}
}
