package hook

// Additional coverage tests for session_start.go functions below 85%.
// Targets: copyDirRecursive, ensureTeammateMode, Handle (windows guard path).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- copyDirRecursive ---

// TestCopyDirRecursive_BasicCopy verifies files are copied recursively.
func TestCopyDirRecursive_BasicCopy(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	// Create src structure: src/file.txt and src/subdir/inner.txt
	_ = os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0o644)
	subDir := filepath.Join(src, "subdir")
	_ = os.MkdirAll(subDir, 0o755)
	_ = os.WriteFile(filepath.Join(subDir, "inner.txt"), []byte("inner"), 0o644)

	dstTarget := filepath.Join(dst, "output")
	if err := copyDirRecursive(src, dstTarget); err != nil {
		t.Fatalf("copyDirRecursive error: %v", err)
	}

	// Verify file.txt
	data, err := os.ReadFile(filepath.Join(dstTarget, "file.txt"))
	if err != nil {
		t.Fatalf("file.txt not copied: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("file.txt content = %q, want 'hello'", data)
	}

	// Verify subdir/inner.txt
	data2, err := os.ReadFile(filepath.Join(dstTarget, "subdir", "inner.txt"))
	if err != nil {
		t.Fatalf("inner.txt not copied: %v", err)
	}
	if string(data2) != "inner" {
		t.Errorf("inner.txt content = %q, want 'inner'", data2)
	}
}

// TestCopyDirRecursive_EmptySrc verifies empty source directory copies successfully.
func TestCopyDirRecursive_EmptySrc(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "output")

	if err := copyDirRecursive(src, dst); err != nil {
		t.Fatalf("copyDirRecursive error for empty src: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("dst not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("dst should be a directory")
	}
}

// TestCopyDirRecursive_NonExistentSrc returns error.
func TestCopyDirRecursive_NonExistentSrc(t *testing.T) {
	t.Parallel()

	dst := filepath.Join(t.TempDir(), "output")
	err := copyDirRecursive("/nonexistent/source/dir", dst)
	if err == nil {
		t.Error("expected error for non-existent source directory")
	}
}

// --- ensureTeammateMode ---

// TestEnsureTeammateMode_OutsideTmux_SetsAuto writes "auto" when not in tmux.
// Not parallel: uses t.Setenv.
func TestEnsureTeammateMode_OutsideTmux_SetsAuto(t *testing.T) {
	t.Setenv("TMUX", "")

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Start with "tmux" so we force a write.
	initial := `{"teammateMode":"tmux"}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(initial), 0o644)

	result := ensureTeammateMode(dir)
	if result != "auto" {
		t.Errorf("expected 'auto' outside tmux, got %q", result)
	}

	// Verify settings.local.json was updated.
	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.local.json"))
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	var mode string
	_ = json.Unmarshal(raw["teammateMode"], &mode)
	if mode != "auto" {
		t.Errorf("settings.local.json teammateMode = %q, want 'auto'", mode)
	}
}

// TestEnsureTeammateMode_InsideTmux_SetsTmux writes "tmux" when inside tmux.
// Not parallel: uses t.Setenv.
func TestEnsureTeammateMode_InsideTmux_SetsTmux(t *testing.T) {
	t.Setenv("TMUX", "/private/tmp/tmux-1234/default,1234,0")

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Start with "auto" so we force a write.
	initial := `{"teammateMode":"auto"}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(initial), 0o644)

	result := ensureTeammateMode(dir)
	if result != "tmux" {
		t.Errorf("expected 'tmux' inside tmux, got %q", result)
	}

	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.local.json"))
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	var mode string
	_ = json.Unmarshal(raw["teammateMode"], &mode)
	if mode != "tmux" {
		t.Errorf("settings.local.json teammateMode = %q, want 'tmux'", mode)
	}
}

// TestEnsureTeammateMode_AlreadyCorrect_NoWrite skips write when value matches.
// Not parallel: uses t.Setenv.
func TestEnsureTeammateMode_AlreadyCorrect_NoWrite(t *testing.T) {
	t.Setenv("TMUX", "") // outside tmux → desired = "auto"

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Already "auto" — no write should happen.
	settings := filepath.Join(claudeDir, "settings.local.json")
	initial := `{"teammateMode":"auto"}`
	_ = os.WriteFile(settings, []byte(initial), 0o644)

	// Capture modification time before.
	info1, _ := os.Stat(settings)

	result := ensureTeammateMode(dir)
	if result != "auto" {
		t.Errorf("expected 'auto' when already correct, got %q", result)
	}

	// File modification time should be unchanged (no write).
	info2, _ := os.Stat(settings)
	if !info1.ModTime().Equal(info2.ModTime()) {
		t.Error("settings.local.json was rewritten unnecessarily (idempotency violation)")
	}
}

// TestEnsureTeammateMode_MissingFile_CreatesFile verifies file creation.
// Not parallel: uses t.Setenv.
func TestEnsureTeammateMode_MissingFile_CreatesFile(t *testing.T) {
	t.Setenv("TMUX", "")

	dir := t.TempDir()
	// No .claude dir or settings.local.json.

	result := ensureTeammateMode(dir)
	if result != "auto" {
		t.Errorf("expected 'auto', got %q", result)
	}

	// settings.local.json should be created.
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.local.json"))
	if err != nil {
		t.Fatalf("settings.local.json not created: %v", err)
	}
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	var mode string
	_ = json.Unmarshal(raw["teammateMode"], &mode)
	if mode != "auto" {
		t.Errorf("teammateMode = %q, want 'auto'", mode)
	}
}

// TestEnsureTeammateMode_RemovesLegacyEnvVar verifies CLAUDE_CODE_TEAMMATE_DISPLAY
// is cleaned up when present.
// Not parallel: uses t.Setenv.
func TestEnsureTeammateMode_RemovesLegacyEnvVar(t *testing.T) {
	t.Setenv("TMUX", "")

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Settings with legacy env var and mismatched mode.
	settings := `{"teammateMode":"tmux","env":{"CLAUDE_CODE_TEAMMATE_DISPLAY":"inline"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

	result := ensureTeammateMode(dir)
	if result != "auto" {
		t.Errorf("expected 'auto', got %q", result)
	}

	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.local.json"))
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)

	// Legacy env var should be removed.
	if envRaw, ok := raw["env"]; ok {
		var env map[string]string
		_ = json.Unmarshal(envRaw, &env)
		if _, exists := env["CLAUDE_CODE_TEAMMATE_DISPLAY"]; exists {
			t.Error("CLAUDE_CODE_TEAMMATE_DISPLAY should have been removed")
		}
	}
}

// TestEnsureTeammateMode_MalformedJSON_ReturnsEmpty verifies graceful handling.
// Not parallel: uses t.Setenv.
func TestEnsureTeammateMode_MalformedJSON_ReturnsEmpty(t *testing.T) {
	t.Setenv("TMUX", "")

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// Malformed JSON → JSON parse fails → returns "".
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte("{bad json"), 0o644)

	result := ensureTeammateMode(dir)
	if result != "" {
		t.Errorf("malformed JSON should return empty, got %q", result)
	}
}
