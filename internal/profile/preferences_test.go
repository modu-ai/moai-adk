package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPreferencesPath_Default(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	path := GetPreferencesPath("default")
	expected := filepath.Join(tmpDir, "preferences.yaml")
	if path != expected {
		t.Errorf("GetPreferencesPath(default) = %q, want %q", path, expected)
	}
}

func TestGetPreferencesPath_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	path := GetPreferencesPath("")
	expected := filepath.Join(tmpDir, "preferences.yaml")
	if path != expected {
		t.Errorf("GetPreferencesPath('') = %q, want %q", path, expected)
	}
}

func TestGetPreferencesPath_Named(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	path := GetPreferencesPath("work")
	expected := filepath.Join(tmpDir, "work", "preferences.yaml")
	if path != expected {
		t.Errorf("GetPreferencesPath(work) = %q, want %q", path, expected)
	}
}

func TestReadPreferences_NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs, err := ReadPreferences("nonexistent")
	if err != nil {
		t.Fatalf("ReadPreferences(nonexistent) unexpected error: %v", err)
	}
	if prefs.UserName != "" || prefs.ConversationLang != "" {
		t.Errorf("ReadPreferences(nonexistent) = %+v, want zero-value", prefs)
	}
}

func TestWriteAndReadPreferences(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{
		UserName:         "testuser",
		ConversationLang: "ko",
		GitCommitLang:    "en",
		CodeCommentLang:  "en",
		DocLang:          "ko",
		ModelPolicy:      "high",
		Model:            "claude-opus-4-6",
		StatuslinePreset: "compact",
	}

	if err := WritePreferences("myprofile", prefs); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	got, err := ReadPreferences("myprofile")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}

	if got.UserName != "testuser" {
		t.Errorf("UserName = %q, want %q", got.UserName, "testuser")
	}
	if got.ConversationLang != "ko" {
		t.Errorf("ConversationLang = %q, want %q", got.ConversationLang, "ko")
	}
	if got.GitCommitLang != "en" {
		t.Errorf("GitCommitLang = %q, want %q", got.GitCommitLang, "en")
	}
	if got.ModelPolicy != "high" {
		t.Errorf("ModelPolicy = %q, want %q", got.ModelPolicy, "high")
	}
	if got.Model != "claude-opus-4-6" {
		t.Errorf("Model = %q, want %q", got.Model, "claude-opus-4-6")
	}
	if got.StatuslinePreset != "compact" {
		t.Errorf("StatuslinePreset = %q, want %q", got.StatuslinePreset, "compact")
	}
}

func TestWritePreferences_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{UserName: "testuser"}

	if err := WritePreferences("newprofile", prefs); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	// Verify directory was created
	profileDir := filepath.Join(tmpDir, "newprofile")
	info, err := os.Stat(profileDir)
	if err != nil {
		t.Fatalf("profile directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
}

func TestReadPreferences_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write invalid YAML
	prefsPath := filepath.Join(tmpDir, "preferences.yaml")
	if err := os.WriteFile(prefsPath, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadPreferences("default")
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestWritePreferences_DefaultProfile(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{UserName: "defaultuser"}

	if err := WritePreferences("default", prefs); err != nil {
		t.Fatalf("WritePreferences(default): %v", err)
	}

	// Should write to base dir, not a "default" subdirectory
	expectedPath := filepath.Join(tmpDir, "preferences.yaml")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("preferences file not at expected path %q: %v", expectedPath, err)
	}
}

func TestPreferences_StatuslineSegments(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{
		StatuslinePreset: "custom",
		StatuslineSegments: map[string]bool{
			"model":   true,
			"context": true,
			"git":     false,
		},
	}

	if err := WritePreferences("default", prefs); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	got, err := ReadPreferences("default")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}

	if len(got.StatuslineSegments) != 3 {
		t.Errorf("StatuslineSegments length = %d, want 3", len(got.StatuslineSegments))
	}
	if !got.StatuslineSegments["model"] {
		t.Error("StatuslineSegments[model] should be true")
	}
	if got.StatuslineSegments["git"] {
		t.Error("StatuslineSegments[git] should be false")
	}
}

func TestReadPreferences_MigratesLegacyFile(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write a legacy dot-prefixed file
	legacyPath := filepath.Join(tmpDir, ".preferences.yaml")
	content := []byte("user_name: legacyuser\nconversation_lang: ko\n")
	if err := os.WriteFile(legacyPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	// ReadPreferences should migrate and return data
	prefs, err := ReadPreferences("default")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}
	if prefs.UserName != "legacyuser" {
		t.Errorf("UserName = %q, want %q", prefs.UserName, "legacyuser")
	}
	if prefs.ConversationLang != "ko" {
		t.Errorf("ConversationLang = %q, want %q", prefs.ConversationLang, "ko")
	}

	// Verify legacy file was renamed
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Error("legacy file should have been renamed")
	}
	newPath := filepath.Join(tmpDir, "preferences.yaml")
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("new file should exist: %v", err)
	}
}

func TestReadPreferences_MigrationSkippedWhenNewExists(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write both old and new files
	legacyPath := filepath.Join(tmpDir, ".preferences.yaml")
	newPath := filepath.Join(tmpDir, "preferences.yaml")
	if err := os.WriteFile(legacyPath, []byte("user_name: old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("user_name: new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should read the new file, not migrate
	prefs, err := ReadPreferences("default")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}
	if prefs.UserName != "new" {
		t.Errorf("UserName = %q, want %q (should read new file)", prefs.UserName, "new")
	}

	// Legacy file should still exist (not removed when new file already exists)
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		t.Error("legacy file should still exist when new file already exists")
	}
}

func TestIsSetup_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write preferences
	if err := WritePreferences("default", ProfilePreferences{UserName: "test"}); err != nil {
		t.Fatal(err)
	}

	if !IsSetup("default") {
		t.Error("IsSetup(default) = false, want true")
	}
}

func TestIsSetup_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	if IsSetup("nonexistent") {
		t.Error("IsSetup(nonexistent) = true, want false")
	}
}

func TestIsSetup_LegacyFileOnly(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write only the legacy file
	legacyPath := filepath.Join(tmpDir, ".preferences.yaml")
	if err := os.WriteFile(legacyPath, []byte("user_name: legacy\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !IsSetup("default") {
		t.Error("IsSetup(default) = false, want true (legacy file exists)")
	}
}

func TestIsSetup_NamedProfile(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	// Write preferences for named profile
	if err := WritePreferences("work", ProfilePreferences{UserName: "worker"}); err != nil {
		t.Fatal(err)
	}

	if !IsSetup("work") {
		t.Error("IsSetup(work) = false, want true")
	}
	if IsSetup("personal") {
		t.Error("IsSetup(personal) = true, want false")
	}
}

func TestPreferences_StatuslineTheme(t *testing.T) {
	tests := []struct {
		name  string
		theme string
	}{
		{"default theme", "default"},
		{"catppuccin-mocha", "catppuccin-mocha"},
		{"catppuccin-latte", "catppuccin-latte"},
		{"empty theme", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			orig := BaseDirOverride
			defer func() { BaseDirOverride = orig }()
			BaseDirOverride = tmpDir

			prefs := ProfilePreferences{
				StatuslineTheme: tt.theme,
			}

			if err := WritePreferences("default", prefs); err != nil {
				t.Fatalf("WritePreferences: %v", err)
			}

			got, err := ReadPreferences("default")
			if err != nil {
				t.Fatalf("ReadPreferences: %v", err)
			}

			if got.StatuslineTheme != tt.theme {
				t.Errorf("StatuslineTheme = %q, want %q", got.StatuslineTheme, tt.theme)
			}
		})
	}
}

func TestPreferences_StatuslineThemePersistsWithOtherFields(t *testing.T) {
	tmpDir := t.TempDir()
	orig := BaseDirOverride
	defer func() { BaseDirOverride = orig }()
	BaseDirOverride = tmpDir

	prefs := ProfilePreferences{
		UserName:         "testuser",
		StatuslinePreset: "compact",
		StatuslineTheme:  "catppuccin-mocha",
	}

	if err := WritePreferences("default", prefs); err != nil {
		t.Fatalf("WritePreferences: %v", err)
	}

	got, err := ReadPreferences("default")
	if err != nil {
		t.Fatalf("ReadPreferences: %v", err)
	}

	if got.UserName != "testuser" {
		t.Errorf("UserName = %q, want %q", got.UserName, "testuser")
	}
	if got.StatuslinePreset != "compact" {
		t.Errorf("StatuslinePreset = %q, want %q", got.StatuslinePreset, "compact")
	}
	if got.StatuslineTheme != "catppuccin-mocha" {
		t.Errorf("StatuslineTheme = %q, want %q", got.StatuslineTheme, "catppuccin-mocha")
	}
}
