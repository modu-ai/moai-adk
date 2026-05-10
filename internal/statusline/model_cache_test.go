package statusline

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadModelCache_Success tests reading model name from cache file.
// AC-SF-004: Cache file exists → returns cached model name.
//
// RED Phase: This test should FAIL because model_cache.go doesn't exist yet.
func TestReadModelCache_Success(t *testing.T) {
	tempDir := t.TempDir()

	// Create .moai/state directory structure
	stateDir := filepath.Join(tempDir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("failed to create state dir: %v", err)
	}

	// Create cache file with model name
	cachePath := filepath.Join(stateDir, "last-model.txt")
	content := []byte("Sonnet 4\n")
	if err := os.WriteFile(cachePath, content, 0o644); err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}

	// Read cache (tempDir is treated as homeDir)
	got, err := ReadModelCache(tempDir)
	if err != nil {
		t.Fatalf("ReadModelCache() error: %v", err)
	}

	expected := "Sonnet 4"
	if got != expected {
		t.Errorf("ReadModelCache() = %q, want %q", got, expected)
	}
}

// TestReadModelCache_FileNotExist tests reading when cache file doesn't exist.
// Expected: returns empty string, no error.
func TestReadModelCache_FileNotExist(t *testing.T) {
	tempDir := t.TempDir()
	// No cache file created

	got, err := ReadModelCache(tempDir)
	if err != nil {
		t.Fatalf("ReadModelCache() should not error when file missing, got: %v", err)
	}

	if got != "" {
		t.Errorf("ReadModelCache() missing file = %q, want empty string", got)
	}
}

// TestReadModelCache_Corrupted tests reading corrupted cache file.
// AC-SF-005: Empty or binary data → returns empty string.
func TestReadModelCache_Corrupted(t *testing.T) {
	tempDir := t.TempDir()

	// Create empty cache file
	cachePath := filepath.Join(tempDir, "last-model.txt")
	if err := os.WriteFile(cachePath, []byte{}, 0o644); err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}

	got, err := ReadModelCache(tempDir)
	if err != nil {
		t.Fatalf("ReadModelCache() should not error on corrupted file, got: %v", err)
	}

	if got != "" {
		t.Errorf("ReadModelCache() corrupted file = %q, want empty string", got)
	}
}

// TestWriteModelCache_CreatesDirectory tests AC-SF-005 part 1: Auto-create directory.
// EC-SF-002: Cache directory doesn't exist → create automatically.
func TestWriteModelCache_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Remove state directory (if exists)
	stateDir := filepath.Join(tempDir, ".moai", "state")
	// stateDir doesn't exist - WriteModelCache should create it

	modelName := "Opus"
	err := WriteModelCache(tempDir, modelName)
	if err != nil {
		t.Fatalf("WriteModelCache() error: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(stateDir)
	if err != nil {
		t.Fatalf("state directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("state path should be a directory")
	}

	// Verify cache file was created
	cachePath := filepath.Join(stateDir, "last-model.txt")
	content, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("cache file should exist: %v", err)
	}

	if string(content) != modelName {
		t.Errorf("cache content = %q, want %q", string(content), modelName)
	}
}

// TestWriteModelCache_AtomicReplace tests AC-SF-005 part 2: Atomic replacement.
// EC-SF-002: Existing file → atomic write to temp + rename.
func TestWriteModelCache_AtomicReplace(t *testing.T) {
	tempDir := t.TempDir()
	stateDir := filepath.Join(tempDir, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("failed to create state dir: %v", err)
	}

	// Write initial model
	initialModel := "Sonnet 4"
	if err := WriteModelCache(tempDir, initialModel); err != nil {
		t.Fatalf("WriteModelCache() initial error: %v", err)
	}

	// Write new model (should atomically replace)
	newModel := "Opus"
	if err := WriteModelCache(tempDir, newModel); err != nil {
		t.Fatalf("WriteModelCache() update error: %v", err)
	}

	// Verify content was replaced
	cachePath := filepath.Join(stateDir, "last-model.txt")
	content, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}

	if string(content) != newModel {
		t.Errorf("cache content after update = %q, want %q", string(content), newModel)
	}
}

// TestWriteModelCache_ReadOnlyDir tests EC-SF-003: Write failure silent ignore.
// When cache directory cannot be written, error should be silent (no panic).
func TestWriteModelCache_ReadOnlyDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file at the cache location (blocking directory creation)
	stateDir := filepath.Join(tempDir, ".moai", "state")
	if err := os.MkdirAll(filepath.Dir(stateDir), 0o755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}

	// Create a FILE named "state" (blocking directory creation)
	if err := os.WriteFile(stateDir, []byte("blocked"), 0o644); err != nil {
		t.Fatalf("failed to create blocking file: %v", err)
	}

	// Should not panic on write error (MkdirAll will fail because "state" is a file)
	err := WriteModelCache(tempDir, "Opus")
	// Error is silently ignored - function should return nil
	if err != nil {
		t.Errorf("WriteModelCache() should silently ignore errors, got: %v", err)
	}
}
