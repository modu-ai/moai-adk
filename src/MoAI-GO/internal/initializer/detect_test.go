package initializer

import (
	"os"
	"path/filepath"
	"testing"
)

// --- NewProjectDetector ---

func TestNewProjectDetector(t *testing.T) {
	pd := NewProjectDetector()
	if pd == nil {
		t.Fatal("NewProjectDetector returned nil")
	}
}

// --- Detect ---

func TestDetect_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	pd := NewProjectDetector()

	result, err := pd.Detect(tmpDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if result.HasClaudeDir {
		t.Error("HasClaudeDir should be false for empty dir")
	}
	if result.HasMoaiDir {
		t.Error("HasMoaiDir should be false for empty dir")
	}
	if result.IsMoaiADK {
		t.Error("IsMoaiADK should be false for empty dir")
	}
}

func TestDetect_ClaudeDirOnly(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	result, err := pd.Detect(tmpDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if !result.HasClaudeDir {
		t.Error("HasClaudeDir should be true")
	}
	if result.HasMoaiDir {
		t.Error("HasMoaiDir should be false")
	}
	if result.IsMoaiADK {
		t.Error("IsMoaiADK should be false (only .claude/ exists)")
	}
}

func TestDetect_MoaiDirOnly(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, ".moai"), 0755); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	result, err := pd.Detect(tmpDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if result.HasClaudeDir {
		t.Error("HasClaudeDir should be false")
	}
	if !result.HasMoaiDir {
		t.Error("HasMoaiDir should be true")
	}
	if result.IsMoaiADK {
		t.Error("IsMoaiADK should be false (only .moai/ exists)")
	}
}

func TestDetect_BothDirs(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, ".moai"), 0755); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	result, err := pd.Detect(tmpDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if !result.HasClaudeDir {
		t.Error("HasClaudeDir should be true")
	}
	if !result.HasMoaiDir {
		t.Error("HasMoaiDir should be true")
	}
	if !result.IsMoaiADK {
		t.Error("IsMoaiADK should be true (both dirs exist)")
	}
}

// --- CheckDirectoryEmpty ---

func TestCheckDirectoryEmpty_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	pd := NewProjectDetector()

	isEmpty, err := pd.CheckDirectoryEmpty(tmpDir)
	if err != nil {
		t.Fatalf("CheckDirectoryEmpty() error = %v", err)
	}
	if !isEmpty {
		t.Error("expected directory to be considered empty")
	}
}

func TestCheckDirectoryEmpty_OnlyDotfiles(t *testing.T) {
	tmpDir := t.TempDir()
	// Create dotfiles/directories
	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	isEmpty, err := pd.CheckDirectoryEmpty(tmpDir)
	if err != nil {
		t.Fatalf("CheckDirectoryEmpty() error = %v", err)
	}
	if !isEmpty {
		t.Error("expected directory with only dotfiles to be considered empty")
	}
}

func TestCheckDirectoryEmpty_NonDotFiles(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	isEmpty, err := pd.CheckDirectoryEmpty(tmpDir)
	if err != nil {
		t.Fatalf("CheckDirectoryEmpty() error = %v", err)
	}
	if isEmpty {
		t.Error("expected directory with non-dot files to not be considered empty")
	}
}

func TestCheckDirectoryEmpty_MixedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	isEmpty, err := pd.CheckDirectoryEmpty(tmpDir)
	if err != nil {
		t.Fatalf("CheckDirectoryEmpty() error = %v", err)
	}
	if isEmpty {
		t.Error("expected directory with mixed files to not be considered empty")
	}
}

func TestCheckDirectoryEmpty_InvalidDir(t *testing.T) {
	pd := NewProjectDetector()
	_, err := pd.CheckDirectoryEmpty("/nonexistent/path/12345")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// --- WarnExisting ---

func TestWarnExisting_NotMoaiADK(t *testing.T) {
	pd := NewProjectDetector()
	result := &DetectResult{IsMoaiADK: false}

	// Should not panic, and produce no output visible to test
	pd.WarnExisting(result)
}

func TestWarnExisting_IsMoaiADK(t *testing.T) {
	pd := NewProjectDetector()
	result := &DetectResult{IsMoaiADK: true}

	// Should not panic
	pd.WarnExisting(result)
}

// --- ShouldInit ---

func TestShouldInit_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	pd := NewProjectDetector()

	should, err := pd.ShouldInit(tmpDir, false)
	if err != nil {
		t.Fatalf("ShouldInit() error = %v", err)
	}
	if !should {
		t.Error("ShouldInit should return true for empty directory")
	}
}

func TestShouldInit_ExistingProject_NoForce(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, ".moai"), 0755); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	should, err := pd.ShouldInit(tmpDir, false)
	if err != nil {
		t.Fatalf("ShouldInit() error = %v", err)
	}
	if should {
		t.Error("ShouldInit should return false for existing project without force")
	}
}

func TestShouldInit_ExistingProject_WithForce(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, ".moai"), 0755); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	should, err := pd.ShouldInit(tmpDir, true)
	if err != nil {
		t.Fatalf("ShouldInit() error = %v", err)
	}
	if !should {
		t.Error("ShouldInit should return true with force flag")
	}
}

func TestShouldInit_NonEmptyDir_NoForce(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	should, err := pd.ShouldInit(tmpDir, false)
	if err != nil {
		t.Fatalf("ShouldInit() error = %v", err)
	}
	if should {
		t.Error("ShouldInit should return false for non-empty directory without force")
	}
}

func TestShouldInit_NonEmptyDir_WithForce(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	pd := NewProjectDetector()
	should, err := pd.ShouldInit(tmpDir, true)
	if err != nil {
		t.Fatalf("ShouldInit() error = %v", err)
	}
	if !should {
		t.Error("ShouldInit should return true with force flag")
	}
}
