package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetBaseDir_Default(t *testing.T) {
	BaseDirOverride = ""
	dir := GetBaseDir()
	if dir == "" || dir == "." {
		// Only fail if HOME is set (CI environments may not have it)
		if os.Getenv("HOME") != "" {
			t.Error("GetBaseDir should return a valid path when HOME is set")
		}
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".moai", "claude-profiles")
	if dir != expected {
		t.Errorf("GetBaseDir() = %q, want %q", dir, expected)
	}
}

func TestGetBaseDir_Override(t *testing.T) {
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()

	BaseDirOverride = "/tmp/test-profiles"
	dir := GetBaseDir()
	if dir != "/tmp/test-profiles" {
		t.Errorf("GetBaseDir() = %q, want /tmp/test-profiles", dir)
	}
}

func TestGetCurrentName_Default(t *testing.T) {
	t.Setenv("MOAI_PROFILE", "")
	name := GetCurrentName()
	if name != "default" {
		t.Errorf("GetCurrentName() = %q, want %q", name, "default")
	}
}

func TestGetCurrentName_WithProfile(t *testing.T) {
	t.Setenv("MOAI_PROFILE", "work")
	name := GetCurrentName()
	if name != "work" {
		t.Errorf("GetCurrentName() = %q, want %q", name, "work")
	}
}

func TestList_DefaultOnly(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	t.Setenv("MOAI_PROFILE", "")

	entries := List()
	if len(entries) != 1 {
		t.Fatalf("List() returned %d entries, want 1", len(entries))
	}
	if entries[0].Name != "default" {
		t.Errorf("entries[0].Name = %q, want %q", entries[0].Name, "default")
	}
	if !entries[0].Current {
		t.Error("default should be current when MOAI_PROFILE is empty")
	}
}

func TestList_WithProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Create profile directories
	if err := os.MkdirAll(filepath.Join(tmpDir, "work"), 0755); err != nil {
		t.Fatalf("MkdirAll(work): %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "personal"), 0755); err != nil {
		t.Fatalf("MkdirAll(personal): %v", err)
	}
	// Create a file (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "notes.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("WriteFile(notes.txt): %v", err)
	}

	t.Setenv("MOAI_PROFILE", "work")

	entries := List()
	if len(entries) != 3 {
		t.Fatalf("List() returned %d entries, want 3", len(entries))
	}

	// Check that work is marked as current
	found := false
	for _, e := range entries {
		if e.Name == "work" && e.Current {
			found = true
		}
		if e.Name == "default" && e.Current {
			t.Error("default should not be current when work is active")
		}
	}
	if !found {
		t.Error("work profile should be marked as current")
	}
}

func TestEnsureDir_Default(t *testing.T) {
	err := EnsureDir("default")
	if err != nil {
		t.Errorf("EnsureDir(default) should be no-op, got: %v", err)
	}
}

func TestEnsureDir_Empty(t *testing.T) {
	err := EnsureDir("")
	if err != nil {
		t.Errorf("EnsureDir('') should be no-op, got: %v", err)
	}
}

func TestEnsureDir_CreatesAndSetsEnv(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	t.Setenv("MOAI_PROFILE", "")
	t.Setenv("CLAUDE_CONFIG_DIR", "") // Clear any inherited value

	err := EnsureDir("myprofile")
	if err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}

	// Check directory was created
	profileDir := filepath.Join(tmpDir, "myprofile")
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		t.Error("profile directory should be created")
	}

	// Check MOAI_PROFILE was set (not CLAUDE_CONFIG_DIR)
	if os.Getenv("MOAI_PROFILE") != "myprofile" {
		t.Errorf("MOAI_PROFILE = %q, want %q", os.Getenv("MOAI_PROFILE"), "myprofile")
	}

	// Verify CLAUDE_CONFIG_DIR is NOT set by EnsureDir
	if os.Getenv("CLAUDE_CONFIG_DIR") != "" {
		t.Errorf("CLAUDE_CONFIG_DIR should not be set, got %q", os.Getenv("CLAUDE_CONFIG_DIR"))
	}
}

func TestDelete_DefaultProfile(t *testing.T) {
	err := Delete("default")
	if err == nil {
		t.Error("Delete(default) should return error")
	}
}

func TestDelete_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	err := Delete("nonexistent")
	if err == nil {
		t.Error("Delete(nonexistent) should return error")
	}
}

func TestDelete_Success(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	t.Setenv("MOAI_PROFILE", "")

	// Create the profile
	profileDir := filepath.Join(tmpDir, "testprofile")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("MkdirAll(testprofile): %v", err)
	}

	err := Delete("testprofile")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := os.Stat(profileDir); !os.IsNotExist(err) {
		t.Error("profile directory should be deleted")
	}
}
