package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetLaunchConfigPath_Default(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	path := GetLaunchConfigPath("default")
	expected := filepath.Join(tmpDir, ".launch.yaml")
	if path != expected {
		t.Errorf("GetLaunchConfigPath(default) = %q, want %q", path, expected)
	}
}

func TestGetLaunchConfigPath_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	path := GetLaunchConfigPath("")
	expected := filepath.Join(tmpDir, ".launch.yaml")
	if path != expected {
		t.Errorf("GetLaunchConfigPath('') = %q, want %q", path, expected)
	}
}

func TestGetLaunchConfigPath_Named(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	path := GetLaunchConfigPath("work")
	expected := filepath.Join(tmpDir, "work", ".launch.yaml")
	if path != expected {
		t.Errorf("GetLaunchConfigPath(work) = %q, want %q", path, expected)
	}
}

func TestReadLaunchConfig_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	cfg, err := ReadLaunchConfig("work")
	if err != nil {
		t.Fatalf("ReadLaunchConfig(nonexistent) unexpected error: %v", err)
	}
	if cfg.Model != "" || cfg.Bypass || cfg.Continue || cfg.Chrome != nil {
		t.Errorf("ReadLaunchConfig(nonexistent) = %+v, want zero-value", cfg)
	}
}

func TestWriteAndReadLaunchConfig(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	chrome := true
	want := LaunchConfig{
		Model:    "claude-opus-4-6",
		Bypass:   true,
		Continue: false,
		Chrome:   &chrome,
	}
	if err := WriteLaunchConfig("work", want); err != nil {
		t.Fatalf("WriteLaunchConfig failed: %v", err)
	}

	got, err := ReadLaunchConfig("work")
	if err != nil {
		t.Fatalf("ReadLaunchConfig failed: %v", err)
	}
	if got.Model != want.Model {
		t.Errorf("Model = %q, want %q", got.Model, want.Model)
	}
	if got.Bypass != want.Bypass {
		t.Errorf("Bypass = %v, want %v", got.Bypass, want.Bypass)
	}
	if got.Continue != want.Continue {
		t.Errorf("Continue = %v, want %v", got.Continue, want.Continue)
	}
	if got.Chrome == nil || *got.Chrome != *want.Chrome {
		t.Errorf("Chrome = %v, want %v", got.Chrome, want.Chrome)
	}
}

func TestWriteLaunchConfig_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Profile dir does not exist yet
	cfg := LaunchConfig{Model: "claude-sonnet-4-6"}
	if err := WriteLaunchConfig("newprofile", cfg); err != nil {
		t.Fatalf("WriteLaunchConfig failed: %v", err)
	}

	path := filepath.Join(tmpDir, "newprofile", ".launch.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("launch config file should be created")
	}
}

func TestWriteLaunchConfig_Default(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	cfg := LaunchConfig{Bypass: true}
	if err := WriteLaunchConfig("default", cfg); err != nil {
		t.Fatalf("WriteLaunchConfig(default) failed: %v", err)
	}

	path := filepath.Join(tmpDir, ".launch.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("default launch config file should be at base dir root")
	}
}

func TestReadLaunchConfig_ChromeDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	disabled := false
	want := LaunchConfig{Chrome: &disabled}
	if err := WriteLaunchConfig("test", want); err != nil {
		t.Fatalf("WriteLaunchConfig failed: %v", err)
	}

	got, err := ReadLaunchConfig("test")
	if err != nil {
		t.Fatalf("ReadLaunchConfig failed: %v", err)
	}
	if got.Chrome == nil || *got.Chrome != false {
		t.Errorf("Chrome = %v, want false", got.Chrome)
	}
}
