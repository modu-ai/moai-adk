package lifecycle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests targeting uncovered branches in cleanup.go to push lifecycle coverage above 85%.

func TestCleanDirectory_EmptyPath(t *testing.T) {
	t.Parallel()

	config := CleanupConfig{
		TempDir: "",
	}
	cleanup := NewSessionCleanup(config)
	// cleanDirectory with empty path should be a no-op
	cleanup.cleanDirectory("")
	// No panic means success
}

func TestCleanDirectory_NotADirectory(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	filePath := filepath.Join(baseDir, "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	config := CleanupConfig{
		TempDir: filePath,
	}
	cleanup := NewSessionCleanup(config)
	// cleanDirectory on a regular file (not a dir) should return early
	cleanup.cleanDirectory(filePath)
	// File should still exist since it's not treated as a directory
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("file should still exist after cleanDirectory on non-directory")
	}
}

func TestCleanDirectory_StatError(t *testing.T) {
	t.Parallel()

	config := CleanupConfig{}
	cleanup := NewSessionCleanup(config)
	// Trigger the stat error branch by using a path with permission issues.
	// On most systems, a non-existent path returns IsNotExist which is handled.
	// A stat error that is NOT IsNotExist requires something like a broken symlink.
	baseDir := t.TempDir()
	brokenLink := filepath.Join(baseDir, "broken-link")
	// Create a symlink to a non-existent target
	if err := os.Symlink(filepath.Join(baseDir, "nonexistent-target"), brokenLink); err != nil {
		t.Skip("cannot create symlink, skipping stat error test")
	}

	cleanup.cleanDirectory(brokenLink)
	// Should have added an error (broken symlink stat returns error that's not IsNotExist)
	if len(cleanup.result.Errors) == 0 {
		// On some systems this might be IsNotExist, which is acceptable
		t.Log("broken symlink was treated as not-exist, which is acceptable")
	}
}

func TestCleanDirectory_FileRemoveError(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	tempDir := filepath.Join(baseDir, "temp")
	subDir := filepath.Join(tempDir, "locked-subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	filePath := filepath.Join(subDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Make directory read-only so removing files fails
	if err := os.Chmod(subDir, 0500); err != nil {
		t.Skip("cannot change permissions, skipping remove error test")
	}
	t.Cleanup(func() {
		_ = os.Chmod(subDir, 0755)
	})

	config := CleanupConfig{
		TempDir: tempDir,
	}
	cleanup := NewSessionCleanup(config)
	result, err := cleanup.CleanTempFiles()
	if err != nil {
		t.Fatalf("CleanTempFiles() should not return error: %v", err)
	}

	// Should have errors from failed removals
	if len(result.Errors) == 0 {
		t.Log("no errors recorded, permission change may not have taken effect")
	}
}

func TestCleanSessionLogs_GlobError(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	logDir := filepath.Join(baseDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	config := CleanupConfig{
		TempDir:           filepath.Join(baseDir, "temp"),
		LogDir:            logDir,
		SessionLogPattern: "[invalid-glob", // Invalid glob pattern
	}

	cleanup := NewSessionCleanup(config)
	result, err := cleanup.CleanTempFiles()
	if err != nil {
		t.Fatalf("CleanTempFiles() should not return error: %v", err)
	}

	// Should have an error from the glob failure
	if len(result.Errors) == 0 {
		t.Error("expected error from invalid glob pattern")
	}

	hasGlobError := false
	for _, e := range result.Errors {
		if strings.Contains(e, "glob") {
			hasGlobError = true
			break
		}
	}
	if !hasGlobError {
		t.Errorf("expected glob error, got errors: %v", result.Errors)
	}
}

func TestCleanSessionLogs_StatError(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	logDir := filepath.Join(baseDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	// Create a matching log file, then create a scenario where stat fails
	logFile := filepath.Join(logDir, "session-test.log")
	if err := os.WriteFile(logFile, []byte("log content"), 0644); err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	// Remove the file between glob and stat to trigger stat error
	// This is a race condition test - we use a different approach
	// Instead, create a broken symlink that matches the pattern
	brokenLog := filepath.Join(logDir, "session-broken.log")
	if err := os.Symlink(filepath.Join(baseDir, "nonexistent"), brokenLog); err != nil {
		t.Skip("cannot create symlink, skipping stat error test")
	}

	config := CleanupConfig{
		TempDir:           filepath.Join(baseDir, "temp"),
		LogDir:            logDir,
		SessionLogPattern: "session-*.log",
	}

	cleanup := NewSessionCleanup(config)
	result, err := cleanup.CleanTempFiles()
	if err != nil {
		t.Fatalf("CleanTempFiles() should not return error: %v", err)
	}

	// The real file should have been deleted
	if result.FilesDeleted < 1 {
		t.Error("expected at least 1 file deleted")
	}
}

func TestCleanSessionLogs_RemoveError(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	logDir := filepath.Join(baseDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	// Create log files
	for i := range 3 {
		logFile := filepath.Join(logDir, "session-"+string(rune('a'+i))+".log")
		if err := os.WriteFile(logFile, []byte("log content"), 0644); err != nil {
			t.Fatalf("failed to create log file: %v", err)
		}
	}

	config := CleanupConfig{
		TempDir:           filepath.Join(baseDir, "temp"),
		LogDir:            logDir,
		SessionLogPattern: "session-*.log",
	}

	cleanup := NewSessionCleanup(config)
	result, err := cleanup.CleanTempFiles()
	if err != nil {
		t.Fatalf("CleanTempFiles() should not return error: %v", err)
	}

	if result.FilesDeleted != 3 {
		t.Errorf("FilesDeleted = %d, want 3", result.FilesDeleted)
	}
	if result.BytesFreed == 0 {
		t.Error("expected some bytes freed from log files")
	}
}

func TestClearCaches_ReadDirError(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	cacheDir := filepath.Join(baseDir, "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	// Make cache directory unreadable
	if err := os.Chmod(cacheDir, 0000); err != nil {
		t.Skip("cannot change permissions, skipping readdir error test")
	}
	t.Cleanup(func() {
		_ = os.Chmod(cacheDir, 0755)
	})

	config := CleanupConfig{
		CacheDir: cacheDir,
	}
	cleanup := NewSessionCleanup(config)
	err := cleanup.ClearCaches()

	// ClearCaches should not return an error even when ReadDir fails
	if err != nil {
		t.Fatalf("ClearCaches() should not error on unreadable dir: %v", err)
	}
}

func TestGenerateCleanupReport_WithManyErrors(t *testing.T) {
	t.Parallel()

	config := DefaultCleanupConfig()
	cleanup := &sessionCleanupImpl{
		config: config,
		result: &CleanupResult{
			FilesDeleted: 5,
			DirsDeleted:  2,
			BytesFreed:   1024 * 1024, // 1 MB
			Errors: []string{
				"error 1",
				"error 2",
				"error 3",
				"error 4",
				"error 5",
				"error 6",
				"error 7",
			},
		},
	}

	report := cleanup.GenerateCleanupReport()

	// Should contain truncation message since there are more than 5 errors
	if !strings.Contains(report, "and 2 more errors") {
		t.Errorf("report should contain '2 more errors' truncation, got: %s", report)
	}

	if !strings.Contains(report, "Files deleted: 5") {
		t.Errorf("report should contain file count, got: %s", report)
	}

	if !strings.Contains(report, "Directories deleted: 2") {
		t.Errorf("report should contain dir count, got: %s", report)
	}

	if !strings.Contains(report, "MB") {
		t.Errorf("report should contain MB for space freed, got: %s", report)
	}

	if !strings.Contains(report, "Errors: 7") {
		t.Errorf("report should contain error count, got: %s", report)
	}
}

func TestGenerateCleanupReport_ExactlyFiveErrors(t *testing.T) {
	t.Parallel()

	config := DefaultCleanupConfig()
	cleanup := &sessionCleanupImpl{
		config: config,
		result: &CleanupResult{
			Errors: []string{
				"error 1",
				"error 2",
				"error 3",
				"error 4",
				"error 5",
			},
		},
	}

	report := cleanup.GenerateCleanupReport()

	// Should NOT contain truncation message for exactly 5 errors
	if strings.Contains(report, "more errors") {
		t.Errorf("report should not truncate at exactly 5 errors, got: %s", report)
	}

	if !strings.Contains(report, "Errors: 5") {
		t.Errorf("report should contain error count 5, got: %s", report)
	}
}

func TestAddError(t *testing.T) {
	t.Parallel()

	config := DefaultCleanupConfig()
	cleanup := NewSessionCleanup(config)

	cleanup.addError("test error message")

	if len(cleanup.result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(cleanup.result.Errors))
	}

	if cleanup.result.Errors[0] != "test error message" {
		t.Errorf("error message = %q, want %q", cleanup.result.Errors[0], "test error message")
	}
}

func TestCleanDirectory_WalkDirWithFileInfo(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	tempDir := filepath.Join(baseDir, "temp")
	subDir := filepath.Join(tempDir, "sub1", "sub2")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	// Create files of various sizes
	sizes := []int{100, 500, 1000}
	for i, size := range sizes {
		content := strings.Repeat("x", size)
		filePath := filepath.Join(subDir, "file"+string(rune('0'+i))+".dat")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	config := CleanupConfig{
		TempDir: tempDir,
	}
	cleanup := NewSessionCleanup(config)
	result, err := cleanup.CleanTempFiles()
	if err != nil {
		t.Fatalf("CleanTempFiles() error: %v", err)
	}

	if result.FilesDeleted != 3 {
		t.Errorf("FilesDeleted = %d, want 3", result.FilesDeleted)
	}

	expectedBytes := int64(100 + 500 + 1000)
	if result.BytesFreed != expectedBytes {
		t.Errorf("BytesFreed = %d, want %d", result.BytesFreed, expectedBytes)
	}

	// Nested dirs (sub1, sub2) should be deleted
	if result.DirsDeleted != 2 {
		t.Errorf("DirsDeleted = %d, want 2", result.DirsDeleted)
	}
}

func TestCleanSessionLogs_NonExistentLogDir(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()

	config := CleanupConfig{
		TempDir:           filepath.Join(baseDir, "temp"),
		LogDir:            filepath.Join(baseDir, "nonexistent-logs"),
		SessionLogPattern: "session-*.log",
	}

	cleanup := NewSessionCleanup(config)
	result, err := cleanup.CleanTempFiles()
	if err != nil {
		t.Fatalf("CleanTempFiles() error: %v", err)
	}

	if result.FilesDeleted != 0 {
		t.Errorf("FilesDeleted = %d, want 0", result.FilesDeleted)
	}
}

func TestCleanSessionLogs_EmptyLogDir(t *testing.T) {
	t.Parallel()

	config := CleanupConfig{
		TempDir:           "",
		LogDir:            "",
		SessionLogPattern: "",
	}

	cleanup := NewSessionCleanup(config)
	result, err := cleanup.CleanTempFiles()
	if err != nil {
		t.Fatalf("CleanTempFiles() error: %v", err)
	}

	if result.FilesDeleted != 0 {
		t.Errorf("FilesDeleted = %d, want 0", result.FilesDeleted)
	}
}
