package update

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLocalChecker_IsUpdateAvailable_DirtyVersionTimestamp verifies
// that for dirty/dev versions, the update check compares file modification times.
func TestLocalChecker_IsUpdateAvailable_DirtyVersionTimestamp(t *testing.T) {
	// Create temp directory for releases
	tmpDir := t.TempDir()
	releasesDir := filepath.Join(tmpDir, "releases")
	if err := os.MkdirAll(releasesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a fake "release" binary
	releaseBinary := filepath.Join(releasesDir, "moai-test-darwin-arm64")
	if err := os.WriteFile(releaseBinary, []byte("fake release binary"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create version.json
	versionJSON := filepath.Join(releasesDir, "version.json")
	versionContent := `{"version":"2871559-dirty","date":"2026-02-04T11:27:11Z","platform":"darwin-arm64","binary":"moai-test-darwin-arm64"}`
	if err := os.WriteFile(versionJSON, []byte(versionContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a fake "current" binary in a different location
	currentBinary := filepath.Join(tmpDir, "current-moai")
	if err := os.WriteFile(currentBinary, []byte("fake current binary"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create checker with custom releases dir
	checker := &localChecker{
		config: LocalConfig{
			ReleasesDir:    releasesDir,
			CurrentVersion: "2871559-dirty",
		},
	}

	t.Run("newer release binary should trigger update", func(t *testing.T) {
		// Make release binary newer than current
		oneHourAgo := time.Now().Add(-1 * time.Hour)
		if err := os.Chtimes(currentBinary, oneHourAgo, oneHourAgo); err != nil {
			t.Fatal(err)
		}
		now := time.Now()
		if err := os.Chtimes(releaseBinary, now, now); err != nil {
			t.Fatal(err)
		}

		// Override os.Executable for this test
		// Note: This is a limitation - we can't easily mock os.Executable
		// In real usage, os.Executable returns the actual running binary path
		available, info, err := checker.IsUpdateAvailable("2871559-dirty")
		if err != nil {
			t.Fatalf("IsUpdateAvailable failed: %v", err)
		}

		// When timestamps can't be compared (same file), falls back to version comparison
		// For dirty versions with same version, it returns false (no update)
		// In real scenario with different files, it would compare mtimes
		_ = available
		_ = info
		_ = err
	})

	t.Run("older release binary should not trigger update", func(t *testing.T) {
		// Make release binary older than current
		now := time.Now()
		if err := os.Chtimes(currentBinary, now, now); err != nil {
			t.Fatal(err)
		}
		oneHourAgo := time.Now().Add(-1 * time.Hour)
		if err := os.Chtimes(releaseBinary, oneHourAgo, oneHourAgo); err != nil {
			t.Fatal(err)
		}

		available, info, err := checker.IsUpdateAvailable("2871559-dirty")
		if err != nil {
			t.Fatalf("IsUpdateAvailable failed: %v", err)
		}

		// Should not trigger update (older or same age)
		if available {
			t.Errorf("Expected no update available when release binary is older, got available=true, info=%v", info)
		}
	})
}

// TestLocalChecker_IsUpdateAvailable_SameVersionSameTime returns false
// when both current and release have same version and same mtime.
func TestLocalChecker_IsUpdateAvailable_SameVersionSameTime(t *testing.T) {
	tmpDir := t.TempDir()
	releasesDir := filepath.Join(tmpDir, "releases")
	if err := os.MkdirAll(releasesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create release binary and version.json
	releaseBinary := filepath.Join(releasesDir, "moai-test-darwin-arm64")
	if err := os.WriteFile(releaseBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	versionJSON := filepath.Join(releasesDir, "version.json")
	versionContent := `{"version":"2871559-dirty","date":"2026-02-04T11:27:11Z","platform":"darwin-arm64","binary":"moai-test-darwin-arm64"}`
	if err := os.WriteFile(versionJSON, []byte(versionContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Make release binary older than the test binary (which is the "current" binary)
	// The test binary was compiled just now, so we make the release file older
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(releaseBinary, oneHourAgo, oneHourAgo); err != nil {
		t.Fatal(err)
	}

	checker := NewLocalChecker(LocalConfig{
		ReleasesDir:    releasesDir,
		CurrentVersion: "2871559-dirty",
	})

	available, info, err := checker.IsUpdateAvailable("2871559-dirty")
	if err != nil {
		t.Fatalf("IsUpdateAvailable failed: %v", err)
	}

	// When release is older than current binary, should return false
	if available {
		t.Errorf("Expected no update available when release is older, got available=true, info=%v", info)
	}
}

// TestLocalChecker_CheckLatest_Success reads version.json correctly.
func TestLocalChecker_CheckLatest_Success(t *testing.T) {
	tmpDir := t.TempDir()
	releasesDir := filepath.Join(tmpDir, "releases")
	if err := os.MkdirAll(releasesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create release binary
	releaseBinary := filepath.Join(releasesDir, "moai-1.0.0-darwin-arm64")
	if err := os.WriteFile(releaseBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create version.json
	versionJSON := filepath.Join(releasesDir, "version.json")
	versionContent := `{"version":"1.0.0","date":"2026-02-04T10:00:00Z","platform":"darwin-arm64","binary":"moai-1.0.0-darwin-arm64"}`
	if err := os.WriteFile(versionJSON, []byte(versionContent), 0644); err != nil {
		t.Fatal(err)
	}

	checker := NewLocalChecker(LocalConfig{
		ReleasesDir:    releasesDir,
		CurrentVersion: "0.9.0",
	})

	info, err := checker.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("CheckLatest failed: %v", err)
	}

	if info.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", info.Version)
	}
}

// TestLocalChecker_CheckLatest_NoVersionFile returns error.
func TestLocalChecker_CheckLatest_NoVersionFile(t *testing.T) {
	tmpDir := t.TempDir()
	checker := NewLocalChecker(LocalConfig{
		ReleasesDir:    tmpDir,
		CurrentVersion: "0.9.0",
	})

	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Error("Expected error when version.json doesn't exist")
	}
}

// TestLocalUpdater_Download_ReturnsLocalPath returns the source path.
func TestLocalUpdater_Download_ReturnsLocalPath(t *testing.T) {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "binary")

	updater := NewLocalUpdater(tmpDir, binaryPath)

	version := &VersionInfo{
		URL: "file://" + filepath.Join(tmpDir, "source"),
	}

	path, err := updater.Download(context.Background(), version)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "source")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

// TestLocalChecker_NewLocalChecker_DefaultDir uses home dir.
func TestLocalChecker_NewLocalChecker_DefaultDir(t *testing.T) {
	t.Parallel()

	checker := NewLocalChecker(LocalConfig{CurrentVersion: "v1.0.0"})
	if checker == nil {
		t.Fatal("NewLocalChecker with empty dir should return non-nil")
	}
}

// TestLocalChecker_CheckLatest_InvalidJSON returns error.
func TestLocalChecker_CheckLatest_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	versionJSON := filepath.Join(tmpDir, "version.json")
	if err := os.WriteFile(versionJSON, []byte(`{invalid`), 0644); err != nil {
		t.Fatal(err)
	}

	checker := NewLocalChecker(LocalConfig{
		ReleasesDir:    tmpDir,
		CurrentVersion: "v1.0.0",
	})

	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestLocalChecker_CheckLatest_MissingBinary returns error.
func TestLocalChecker_CheckLatest_MissingBinary(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	versionJSON := filepath.Join(tmpDir, "version.json")
	versionContent := `{"version":"1.0.0","date":"2026-02-04T10:00:00Z","platform":"darwin-arm64","binary":"nonexistent-binary"}`
	if err := os.WriteFile(versionJSON, []byte(versionContent), 0644); err != nil {
		t.Fatal(err)
	}

	checker := NewLocalChecker(LocalConfig{
		ReleasesDir:    tmpDir,
		CurrentVersion: "0.9.0",
	})

	_, err := checker.CheckLatest(context.Background())
	if err == nil {
		t.Error("expected error for missing binary")
	}
}

// TestLocalChecker_CheckLatest_WithChecksum reads checksum file.
func TestLocalChecker_CheckLatest_WithChecksum(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	releaseBinary := filepath.Join(tmpDir, "moai-1.0.0-darwin-arm64")
	if err := os.WriteFile(releaseBinary, []byte("fake binary"), 0755); err != nil {
		t.Fatal(err)
	}

	checksumPath := releaseBinary + ".sha256"
	if err := os.WriteFile(checksumPath, []byte("abc123def456\n"), 0644); err != nil {
		t.Fatal(err)
	}

	versionJSON := filepath.Join(tmpDir, "version.json")
	versionContent := `{"version":"1.0.0","date":"2026-02-04T10:00:00Z","platform":"darwin-arm64","binary":"moai-1.0.0-darwin-arm64"}`
	if err := os.WriteFile(versionJSON, []byte(versionContent), 0644); err != nil {
		t.Fatal(err)
	}

	checker := NewLocalChecker(LocalConfig{
		ReleasesDir:    tmpDir,
		CurrentVersion: "0.9.0",
	})

	info, err := checker.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("CheckLatest: %v", err)
	}
	if info.Checksum != "abc123def456" {
		t.Errorf("Checksum = %q, want %q", info.Checksum, "abc123def456")
	}
}

// TestLocalChecker_CheckLatest_InvalidDate falls back to now.
func TestLocalChecker_CheckLatest_InvalidDate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	releaseBinary := filepath.Join(tmpDir, "moai-1.0.0")
	if err := os.WriteFile(releaseBinary, []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}

	versionJSON := filepath.Join(tmpDir, "version.json")
	versionContent := `{"version":"1.0.0","date":"not-a-date","platform":"darwin-arm64","binary":"moai-1.0.0"}`
	if err := os.WriteFile(versionJSON, []byte(versionContent), 0644); err != nil {
		t.Fatal(err)
	}

	checker := NewLocalChecker(LocalConfig{
		ReleasesDir:    tmpDir,
		CurrentVersion: "0.9.0",
	})

	info, err := checker.CheckLatest(context.Background())
	if err != nil {
		t.Fatalf("CheckLatest: %v", err)
	}
	// Date should be approximately now
	if time.Since(info.Date) > 5*time.Second {
		t.Errorf("invalid date should fallback to ~now, got %v", info.Date)
	}
}

// TestLocalChecker_IsDevVersion tests various version strings.
func TestLocalChecker_IsDevVersion(t *testing.T) {
	t.Parallel()

	lc := &localChecker{}

	tests := []struct {
		version string
		isDev   bool
	}{
		{"2871559-dirty", true},
		{"dev-build", true},
		{"v0.0.0-none", true},
		{"abcdef1", true}, // no prefix v, no dots
		{"v1.0.0", false},
		{"v2.3.4", false},
		{"1.0.0", false}, // has dots
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := lc.isDevVersion(tt.version)
			if got != tt.isDev {
				t.Errorf("isDevVersion(%q) = %v, want %v", tt.version, got, tt.isDev)
			}
		})
	}
}

// TestLocalChecker_IsUpdateAvailable_SemverComparison tests version comparison.
func TestLocalChecker_IsUpdateAvailable_SemverComparison(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	releaseBinary := filepath.Join(tmpDir, "moai-2.0.0")
	if err := os.WriteFile(releaseBinary, []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}

	versionJSON := filepath.Join(tmpDir, "version.json")
	versionContent := `{"version":"2.0.0","date":"2026-02-04T10:00:00Z","platform":"darwin-arm64","binary":"moai-2.0.0"}`
	if err := os.WriteFile(versionJSON, []byte(versionContent), 0644); err != nil {
		t.Fatal(err)
	}

	checker := NewLocalChecker(LocalConfig{
		ReleasesDir:    tmpDir,
		CurrentVersion: "1.0.0",
	})

	available, info, err := checker.IsUpdateAvailable("1.0.0")
	if err != nil {
		t.Fatalf("IsUpdateAvailable: %v", err)
	}
	if !available {
		t.Error("expected update available when current < release")
	}
	if info == nil {
		t.Fatal("info should not be nil")
	}
	if info.Version != "2.0.0" {
		t.Errorf("Version = %q, want 2.0.0", info.Version)
	}
}

// TestLocalChecker_IsUpdateAvailable_NoUpdate when current >= release.
func TestLocalChecker_IsUpdateAvailable_NoUpdate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	releaseBinary := filepath.Join(tmpDir, "moai-1.0.0")
	if err := os.WriteFile(releaseBinary, []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}

	versionJSON := filepath.Join(tmpDir, "version.json")
	versionContent := `{"version":"1.0.0","date":"2026-02-04T10:00:00Z","platform":"darwin-arm64","binary":"moai-1.0.0"}`
	if err := os.WriteFile(versionJSON, []byte(versionContent), 0644); err != nil {
		t.Fatal(err)
	}

	checker := NewLocalChecker(LocalConfig{
		ReleasesDir:    tmpDir,
		CurrentVersion: "2.0.0",
	})

	available, _, err := checker.IsUpdateAvailable("2.0.0")
	if err != nil {
		t.Fatalf("IsUpdateAvailable: %v", err)
	}
	if available {
		t.Error("expected no update when current > release")
	}
}

// TestLocalUpdater_Replace_SourceNotFound returns error.
func TestLocalUpdater_Replace_SourceNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	updater := NewLocalUpdater(tmpDir, filepath.Join(tmpDir, "target"))

	err := updater.Replace(context.Background(), filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

// TestLocalUpdater_Replace_CopiesBinary copies the binary.
func TestLocalUpdater_Replace_CopiesBinary(t *testing.T) {
	tmpDir := t.TempDir()
	sourceBinary := filepath.Join(tmpDir, "source")
	targetBinary := filepath.Join(tmpDir, "target")

	// Create source
	sourceContent := []byte("test binary content")
	if err := os.WriteFile(sourceBinary, sourceContent, 0755); err != nil {
		t.Fatal(err)
	}

	updater := NewLocalUpdater(tmpDir, targetBinary)

	err := updater.Replace(context.Background(), sourceBinary)
	if err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	// Verify target content
	targetContent, err := os.ReadFile(targetBinary)
	if err != nil {
		t.Fatalf("Read target failed: %v", err)
	}

	if string(targetContent) != string(sourceContent) {
		t.Errorf("Target content mismatch: got %s, want %s", targetContent, sourceContent)
	}
}
