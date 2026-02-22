package lifecycle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- TDD RED Phase: Tests for SessionCleanup ---

func TestCleanup_CleanTempFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupFiles      []string
		setupDirs       []string
		wantDeleted     int
		wantDirsDeleted int
		wantErr         bool
	}{
		{
			name:            "clean empty temp directory",
			setupFiles:      nil,
			setupDirs:       nil,
			wantDeleted:     0,
			wantDirsDeleted: 0,
			wantErr:         false,
		},
		{
			name: "clean temp files",
			setupFiles: []string{
				"temp/file1.tmp",
				"temp/file2.tmp",
			},
			setupDirs:       nil,
			wantDeleted:     2,
			wantDirsDeleted: 0,
			wantErr:         false,
		},
		{
			name: "clean temp directory with subdirs",
			setupFiles: []string{
				"temp/file1.tmp",
				"temp/subdir/file2.tmp",
			},
			setupDirs: []string{
				"temp/subdir",
			},
			wantDeleted:     2,
			wantDirsDeleted: 1,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory structure
			baseDir := t.TempDir()
			moaiDir := filepath.Join(baseDir, ".moai")

			// Setup directories
			tempDir := filepath.Join(moaiDir, "temp")
			if err := os.MkdirAll(tempDir, 0755); err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}

			for _, dir := range tt.setupDirs {
				dirPath := filepath.Join(moaiDir, dir)
				if err := os.MkdirAll(dirPath, 0755); err != nil {
					t.Fatalf("failed to create dir %s: %v", dir, err)
				}
			}

			// Setup files
			for _, file := range tt.setupFiles {
				filePath := filepath.Join(moaiDir, file)
				dir := filepath.Dir(filePath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("failed to create parent dir for %s: %v", file, err)
				}
				if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
					t.Fatalf("failed to create file %s: %v", file, err)
				}
			}

			config := CleanupConfig{
				TempDir:        filepath.Join(moaiDir, "temp"),
				CacheDir:       filepath.Join(moaiDir, "cache", "temp"),
				LogDir:         filepath.Join(moaiDir, "logs"),
				PreserveMemory: true,
			}

			cleanup := NewSessionCleanup(config)
			result, err := cleanup.CleanTempFiles()

			if (err != nil) != tt.wantErr {
				t.Errorf("CleanTempFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Fatal("CleanTempFiles() returned nil result")
			}

			if result.FilesDeleted != tt.wantDeleted {
				t.Errorf("FilesDeleted = %d, want %d", result.FilesDeleted, tt.wantDeleted)
			}

			if result.DirsDeleted != tt.wantDirsDeleted {
				t.Errorf("DirsDeleted = %d, want %d", result.DirsDeleted, tt.wantDirsDeleted)
			}
		})
	}
}

func TestCleanup_ClearCaches(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	moaiDir := filepath.Join(baseDir, ".moai")
	cacheDir := filepath.Join(moaiDir, "cache", "temp")

	// Create cache directory with files
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	// Create some cache files
	for i := range 3 {
		filePath := filepath.Join(cacheDir, "cache"+string(rune('0'+i))+".json")
		if err := os.WriteFile(filePath, []byte("{}"), 0644); err != nil {
			t.Fatalf("failed to create cache file: %v", err)
		}
	}

	config := CleanupConfig{
		TempDir:        filepath.Join(moaiDir, "temp"),
		CacheDir:       cacheDir,
		LogDir:         filepath.Join(moaiDir, "logs"),
		PreserveMemory: true,
	}

	cleanup := NewSessionCleanup(config)
	err := cleanup.ClearCaches()

	if err != nil {
		t.Fatalf("ClearCaches() error = %v", err)
	}

	// Verify cache directory is empty or doesn't exist
	entries, err := os.ReadDir(cacheDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to read cache dir: %v", err)
	}

	if len(entries) > 0 {
		t.Errorf("cache directory should be empty, has %d entries", len(entries))
	}
}

func TestCleanup_PreserveMemoryFiles(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	moaiDir := filepath.Join(baseDir, ".moai")
	memoryDir := filepath.Join(moaiDir, "memory")
	tempDir := filepath.Join(moaiDir, "temp")

	// Create memory directory with state files
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		t.Fatalf("failed to create memory dir: %v", err)
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create memory files that should be preserved
	memoryFile := filepath.Join(memoryDir, "last-session-state.json")
	if err := os.WriteFile(memoryFile, []byte(`{"key": "value"}`), 0644); err != nil {
		t.Fatalf("failed to create memory file: %v", err)
	}

	// Create temp file that should be deleted
	tempFile := filepath.Join(tempDir, "should-delete.tmp")
	if err := os.WriteFile(tempFile, []byte("temp"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	config := CleanupConfig{
		TempDir:        tempDir,
		CacheDir:       filepath.Join(moaiDir, "cache", "temp"),
		LogDir:         filepath.Join(moaiDir, "logs"),
		PreserveMemory: true,
	}

	cleanup := NewSessionCleanup(config)
	_, err := cleanup.CleanTempFiles()

	if err != nil {
		t.Fatalf("CleanTempFiles() error = %v", err)
	}

	// Memory file should still exist
	if _, err := os.Stat(memoryFile); os.IsNotExist(err) {
		t.Error("memory file should be preserved")
	}

	// Temp file should be deleted
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("temp file should be deleted")
	}
}

func TestCleanup_GenerateCleanupReport(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	moaiDir := filepath.Join(baseDir, ".moai")
	tempDir := filepath.Join(moaiDir, "temp")

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create some files to clean
	for i := range 2 {
		filePath := filepath.Join(tempDir, "file"+string(rune('0'+i))+".tmp")
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
	}

	config := CleanupConfig{
		TempDir:        tempDir,
		CacheDir:       filepath.Join(moaiDir, "cache", "temp"),
		LogDir:         filepath.Join(moaiDir, "logs"),
		PreserveMemory: true,
	}

	cleanup := NewSessionCleanup(config)

	// Clean first
	_, _ = cleanup.CleanTempFiles()

	report := cleanup.GenerateCleanupReport()

	// Report should contain cleanup information
	if report == "" {
		t.Error("GenerateCleanupReport() returned empty string")
	}

	if !strings.Contains(report, "Cleanup") && !strings.Contains(report, "cleanup") &&
		!strings.Contains(report, "deleted") && !strings.Contains(report, "files") {
		t.Errorf("GenerateCleanupReport() should contain cleanup info, got: %s", report)
	}
}

func TestCleanup_ContinueOnError(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	moaiDir := filepath.Join(baseDir, ".moai")
	tempDir := filepath.Join(moaiDir, "temp")

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create a file
	filePath := filepath.Join(tempDir, "normal.tmp")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	config := CleanupConfig{
		TempDir:        tempDir,
		CacheDir:       filepath.Join(moaiDir, "cache", "temp"),
		LogDir:         filepath.Join(moaiDir, "logs"),
		PreserveMemory: true,
	}

	cleanup := NewSessionCleanup(config)
	result, err := cleanup.CleanTempFiles()

	// Should not return error even if some files fail (REQ-HOOK-362)
	if err != nil {
		t.Fatalf("CleanTempFiles() should not return error, got: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	// The file that could be deleted should be counted
	if result.FilesDeleted < 1 {
		t.Errorf("FilesDeleted should be at least 1, got %d", result.FilesDeleted)
	}
}

func TestCleanup_NonExistentDirectories(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()

	config := CleanupConfig{
		TempDir:        filepath.Join(baseDir, "nonexistent", "temp"),
		CacheDir:       filepath.Join(baseDir, "nonexistent", "cache"),
		LogDir:         filepath.Join(baseDir, "nonexistent", "logs"),
		PreserveMemory: true,
	}

	cleanup := NewSessionCleanup(config)

	// Should not error on non-existent directories
	result, err := cleanup.CleanTempFiles()
	if err != nil {
		t.Fatalf("CleanTempFiles() should not error on non-existent dir: %v", err)
	}

	if result.FilesDeleted != 0 {
		t.Errorf("FilesDeleted should be 0, got %d", result.FilesDeleted)
	}

	err = cleanup.ClearCaches()
	if err != nil {
		t.Fatalf("ClearCaches() should not error on non-existent dir: %v", err)
	}
}

func TestNewSessionCleanup(t *testing.T) {
	t.Parallel()

	config := DefaultCleanupConfig()
	cleanup := NewSessionCleanup(config)

	if cleanup == nil {
		t.Fatal("NewSessionCleanup() returned nil")
	}
}

func TestCleanup_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ SessionCleanup = (*sessionCleanupImpl)(nil)
}

func TestCleanup_SessionLogs(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	moaiDir := filepath.Join(baseDir, ".moai")
	logDir := filepath.Join(moaiDir, "logs")
	tempDir := filepath.Join(moaiDir, "temp")

	// Create directories
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create session log files
	for i := range 3 {
		logFile := filepath.Join(logDir, "session-"+string(rune('0'+i))+".log")
		if err := os.WriteFile(logFile, []byte("log content"), 0644); err != nil {
			t.Fatalf("failed to create log file: %v", err)
		}
	}

	// Create a non-session log file that should be preserved
	otherLog := filepath.Join(logDir, "other.log")
	if err := os.WriteFile(otherLog, []byte("other log"), 0644); err != nil {
		t.Fatalf("failed to create other log: %v", err)
	}

	config := CleanupConfig{
		TempDir:           tempDir,
		CacheDir:          filepath.Join(moaiDir, "cache", "temp"),
		LogDir:            logDir,
		SessionLogPattern: "session-*.log",
		PreserveMemory:    true,
	}

	cleanup := NewSessionCleanup(config)
	result, err := cleanup.CleanTempFiles()

	if err != nil {
		t.Fatalf("CleanTempFiles() error = %v", err)
	}

	// Session logs should be deleted
	if result.FilesDeleted != 3 {
		t.Errorf("FilesDeleted = %d, want 3", result.FilesDeleted)
	}

	// Other log should still exist
	if _, err := os.Stat(otherLog); os.IsNotExist(err) {
		t.Error("other.log should be preserved")
	}
}

func TestCleanup_ClearCachesWithSubdirs(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	cacheDir := filepath.Join(baseDir, "cache")
	subDir := filepath.Join(cacheDir, "subdir")

	// Create cache directory with subdirectory
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create cache subdir: %v", err)
	}

	// Create files in both cache and subdir
	if err := os.WriteFile(filepath.Join(cacheDir, "file.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create cache file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create nested cache file: %v", err)
	}

	config := CleanupConfig{
		TempDir:        filepath.Join(baseDir, "temp"),
		CacheDir:       cacheDir,
		LogDir:         filepath.Join(baseDir, "logs"),
		PreserveMemory: true,
	}

	cleanup := NewSessionCleanup(config)
	err := cleanup.ClearCaches()

	if err != nil {
		t.Fatalf("ClearCaches() error = %v", err)
	}

	// Both files and subdirectory should be removed
	entries, err := os.ReadDir(cacheDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to read cache dir: %v", err)
	}

	if len(entries) > 0 {
		t.Errorf("cache directory should be empty, has %d entries", len(entries))
	}
}

func TestCleanup_FormatBytesVariousSizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 bytes"},
		{500, "500 bytes"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{1073741824, "1.00 GB"},
		{1610612736, "1.50 GB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
		}
	}
}

func TestCleanup_GenerateCleanupReportWithErrors(t *testing.T) {
	t.Parallel()

	config := CleanupConfig{
		TempDir:        "/nonexistent/path",
		CacheDir:       "/nonexistent/cache",
		LogDir:         "/nonexistent/logs",
		PreserveMemory: true,
	}

	cleanup := NewSessionCleanup(config)

	// Generate report without cleanup
	report := cleanup.GenerateCleanupReport()

	if !strings.Contains(report, "Cleanup") {
		t.Errorf("report should contain 'Cleanup', got: %s", report)
	}
}

func TestCleanup_GenerateCleanupReportNilResult(t *testing.T) {
	t.Parallel()

	config := DefaultCleanupConfig()
	cleanup := &sessionCleanupImpl{
		config: config,
		result: nil,
	}

	report := cleanup.GenerateCleanupReport()

	if !strings.Contains(report, "No cleanup performed") {
		t.Errorf("report should indicate no cleanup, got: %s", report)
	}
}

func TestCleanup_EmptyCacheDir(t *testing.T) {
	t.Parallel()

	config := CleanupConfig{
		TempDir:        "",
		CacheDir:       "",
		LogDir:         "",
		PreserveMemory: true,
	}

	cleanup := NewSessionCleanup(config)
	err := cleanup.ClearCaches()

	// Should not error with empty config
	if err != nil {
		t.Fatalf("ClearCaches() with empty config error = %v", err)
	}
}
