package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// --- backupMoaiConfig additional edge case tests ---

func TestBackupMoaiConfig_ConfigPathIsFile(t *testing.T) {
	// When .moai/config is a regular file instead of a directory,
	// backupMoaiConfig should return an error.
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, defs.MoAIDir)
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create .moai/config as a regular file (not a directory)
	configPath := filepath.Join(moaiDir, defs.ConfigSubdir)
	if err := os.WriteFile(configPath, []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := backupMoaiConfig(tmpDir)
	if err == nil {
		t.Fatal("expected error when config path is a file, got nil")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("error should mention 'not a directory', got: %v", err)
	}
}

func TestBackupMoaiConfig_MetadataContainsAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, defs.MoAIDir, defs.ConfigSubdir, "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "user.yaml"), []byte("user:\n  name: tester\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	backupDir, err := backupMoaiConfig(tmpDir)
	if err != nil {
		t.Fatalf("backupMoaiConfig failed: %v", err)
	}
	if backupDir == "" {
		t.Fatal("expected non-empty backup dir")
	}

	// Read and validate metadata
	metaPath := filepath.Join(backupDir, "backup_metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}

	var meta BackupMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}

	if meta.BackupType != "config" {
		t.Errorf("BackupType = %q, want %q", meta.BackupType, "config")
	}
	if meta.Description != "config_backup" {
		t.Errorf("Description = %q, want %q", meta.Description, "config_backup")
	}
	if meta.ProjectRoot != tmpDir {
		t.Errorf("ProjectRoot = %q, want %q", meta.ProjectRoot, tmpDir)
	}
	if meta.TemplateDefaultsDir != ".template-defaults" {
		t.Errorf("TemplateDefaultsDir = %q, want %q", meta.TemplateDefaultsDir, ".template-defaults")
	}
	if len(meta.BackedUpItems) == 0 {
		t.Error("BackedUpItems should not be empty")
	}
	// BackedUpItems should contain forward-slash paths
	for _, item := range meta.BackedUpItems {
		if strings.Contains(item, "\\") {
			t.Errorf("BackedUpItems should use forward slashes, got: %s", item)
		}
	}
}

func TestBackupMoaiConfig_NestedSubdirectories(t *testing.T) {
	// Verify that deeply nested files are backed up properly.
	tmpDir := t.TempDir()
	deepDir := filepath.Join(tmpDir, defs.MoAIDir, defs.ConfigSubdir, "sections", "nested", "deep")
	if err := os.MkdirAll(deepDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deepDir, "deep.yaml"), []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	backupDir, err := backupMoaiConfig(tmpDir)
	if err != nil {
		t.Fatalf("backupMoaiConfig failed: %v", err)
	}

	// Verify the deeply nested file was backed up
	backedUpPath := filepath.Join(backupDir, "sections", "nested", "deep", "deep.yaml")
	if _, err := os.Stat(backedUpPath); os.IsNotExist(err) {
		t.Error("deeply nested file should be backed up")
	}

	data, err := os.ReadFile(backedUpPath)
	if err != nil {
		t.Fatalf("read backed up file: %v", err)
	}
	if string(data) != "key: value\n" {
		t.Errorf("backed up content = %q, want %q", string(data), "key: value\n")
	}
}

// --- saveTemplateDefaults tests ---

func TestSaveTemplateDefaults_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "template-defaults")

	err := saveTemplateDefaults(destDir)
	if err != nil {
		t.Fatalf("saveTemplateDefaults failed: %v", err)
	}

	// The function should have created the sections/ subdirectory
	sectionsDir := filepath.Join(destDir, "sections")
	info, err := os.Stat(sectionsDir)
	if os.IsNotExist(err) {
		t.Fatal("sections directory should have been created")
	}
	if !info.IsDir() {
		t.Error("sections should be a directory")
	}
}

func TestSaveTemplateDefaults_WritesYAMLFiles(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "template-defaults")

	err := saveTemplateDefaults(destDir)
	if err != nil {
		t.Fatalf("saveTemplateDefaults failed: %v", err)
	}

	// Check that at least some known section files were written.
	// The embedded templates should contain these standard section files.
	sectionsDir := filepath.Join(destDir, "sections")
	entries, err := os.ReadDir(sectionsDir)
	if err != nil {
		t.Fatalf("read sections dir: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("saveTemplateDefaults should write at least one section file")
	}

	// Check that files are non-empty
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sectionsDir, entry.Name()))
		if err != nil {
			t.Errorf("read %s: %v", entry.Name(), err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("file %s should not be empty", entry.Name())
		}
	}
}

func TestSaveTemplateDefaults_StripsTmplExtension(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "template-defaults")

	err := saveTemplateDefaults(destDir)
	if err != nil {
		t.Fatalf("saveTemplateDefaults failed: %v", err)
	}

	// No files should have .tmpl extension in the output
	sectionsDir := filepath.Join(destDir, "sections")
	entries, err := os.ReadDir(sectionsDir)
	if err != nil {
		t.Fatalf("read sections dir: %v", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmpl") {
			t.Errorf("output file should not have .tmpl extension: %s", entry.Name())
		}
	}
}

func TestSaveTemplateDefaults_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "template-defaults")
	sectionsDir := filepath.Join(destDir, "sections")

	// Create the directory and a placeholder file
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "placeholder.yaml"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Call saveTemplateDefaults - should not fail even if directory already exists
	err := saveTemplateDefaults(destDir)
	if err != nil {
		t.Fatalf("saveTemplateDefaults failed on existing directory: %v", err)
	}
}

// --- restoreMoaiConfigLegacy tests ---

func TestRestoreMoaiConfigLegacy_RestoresFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the config directory (simulating freshly deployed templates)
	configDir := filepath.Join(tmpDir, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a backup directory with a legacy file
	backupDir := filepath.Join(tmpDir, "backup-legacy")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a file in the backup (not in sections/ subdirectory, legacy format)
	backupFile := filepath.Join(backupDir, "custom.yaml")
	backupContent := []byte("custom:\n  key: legacy-value\n")
	if err := os.WriteFile(backupFile, backupContent, 0o644); err != nil {
		t.Fatal(err)
	}

	err := restoreMoaiConfigLegacy(tmpDir, backupDir, configDir)
	if err != nil {
		t.Fatalf("restoreMoaiConfigLegacy failed: %v", err)
	}

	// Verify the file was restored to the config directory
	restoredPath := filepath.Join(configDir, "custom.yaml")
	data, err := os.ReadFile(restoredPath)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if !strings.Contains(string(data), "legacy-value") {
		t.Errorf("restored file should contain backup content, got:\n%s", string(data))
	}
}

func TestRestoreMoaiConfigLegacy_SkipsMetadata(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	backupDir := filepath.Join(tmpDir, "backup-legacy")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write backup_metadata.json (should be skipped)
	metadataFile := filepath.Join(backupDir, "backup_metadata.json")
	if err := os.WriteFile(metadataFile, []byte(`{"timestamp":"test"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write .template-defaults/ directory (should be skipped)
	templateDefaultsDir := filepath.Join(backupDir, ".template-defaults")
	if err := os.MkdirAll(templateDefaultsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDefaultsDir, "test.yaml"), []byte("skip: me"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := restoreMoaiConfigLegacy(tmpDir, backupDir, configDir)
	if err != nil {
		t.Fatalf("restoreMoaiConfigLegacy failed: %v", err)
	}

	// backup_metadata.json should NOT be restored to config dir
	restoredMeta := filepath.Join(configDir, "backup_metadata.json")
	if _, err := os.Stat(restoredMeta); !os.IsNotExist(err) {
		t.Error("backup_metadata.json should not be restored to config directory")
	}

	// .template-defaults/ files should NOT be restored
	restoredDefaults := filepath.Join(configDir, ".template-defaults")
	if _, err := os.Stat(restoredDefaults); !os.IsNotExist(err) {
		t.Error(".template-defaults should not be restored to config directory")
	}
}

func TestRestoreMoaiConfigLegacy_MergesWithExistingTarget(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create an existing target file (simulating new template)
	targetPath := filepath.Join(configDir, "user.yaml")
	targetContent := []byte("user:\n  name: new-template\n  new_field: added\n")
	if err := os.WriteFile(targetPath, targetContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a backup with user's old config
	backupDir := filepath.Join(tmpDir, "backup-legacy")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	backupContent := []byte("user:\n  name: my-custom-name\n")
	if err := os.WriteFile(filepath.Join(backupDir, "user.yaml"), backupContent, 0o644); err != nil {
		t.Fatal(err)
	}

	err := restoreMoaiConfigLegacy(tmpDir, backupDir, configDir)
	if err != nil {
		t.Fatalf("restoreMoaiConfigLegacy failed: %v", err)
	}

	// Read the merged result
	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read merged file: %v", err)
	}

	// User's name should be merged in
	if !strings.Contains(string(data), "my-custom-name") {
		t.Errorf("merged file should contain user's custom name, got:\n%s", string(data))
	}
}

func TestRestoreMoaiConfigLegacy_EmptyBackup(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Empty backup directory
	backupDir := filepath.Join(tmpDir, "backup-empty")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Should complete without error
	err := restoreMoaiConfigLegacy(tmpDir, backupDir, configDir)
	if err != nil {
		t.Fatalf("restoreMoaiConfigLegacy should succeed with empty backup, got: %v", err)
	}
}

// --- restoreMoaiConfig (3-way merge path) tests ---

func TestRestoreMoaiConfig_FallsBackToLegacyWhenNoSections(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory
	configDir := filepath.Join(tmpDir, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create backup without sections/ directory (legacy format)
	backupDir := filepath.Join(tmpDir, "backup")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Put a YAML file at root level of backup
	if err := os.WriteFile(filepath.Join(backupDir, "user.yaml"), []byte("user:\n  name: legacy\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := restoreMoaiConfig(tmpDir, backupDir)
	if err != nil {
		t.Fatalf("restoreMoaiConfig should fall back to legacy, got: %v", err)
	}

	// The file should be restored via the legacy path
	restoredPath := filepath.Join(configDir, "user.yaml")
	data, err := os.ReadFile(restoredPath)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if !strings.Contains(string(data), "legacy") {
		t.Errorf("expected legacy content to be restored, got:\n%s", string(data))
	}
}

func TestRestoreMoaiConfig_3WayMergeWithTemplateDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory with new template data
	configDir := filepath.Join(tmpDir, defs.MoAIDir, defs.ConfigSubdir)
	sectionsDir := filepath.Join(configDir, "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// New template version
	newContent := []byte("user:\n  name: template-default\n  new_feature: enabled\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, "user.yaml"), newContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Create backup with sections/ subdirectory and .template-defaults/
	backupDir := filepath.Join(tmpDir, "backup")
	backupSectionsDir := filepath.Join(backupDir, "sections")
	templateDefaultsDir := filepath.Join(backupDir, ".template-defaults", "sections")
	if err := os.MkdirAll(backupSectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(templateDefaultsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// User's old config (user modified the name)
	oldContent := []byte("user:\n  name: my-custom-name\n")
	if err := os.WriteFile(filepath.Join(backupSectionsDir, "user.yaml"), oldContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Old template defaults (original template value before user modified it)
	baseContent := []byte("user:\n  name: template-default\n")
	if err := os.WriteFile(filepath.Join(templateDefaultsDir, "user.yaml"), baseContent, 0o644); err != nil {
		t.Fatal(err)
	}

	err := restoreMoaiConfig(tmpDir, backupDir)
	if err != nil {
		t.Fatalf("restoreMoaiConfig failed: %v", err)
	}

	// Read the merged result
	data, err := os.ReadFile(filepath.Join(sectionsDir, "user.yaml"))
	if err != nil {
		t.Fatalf("read merged file: %v", err)
	}

	resultStr := string(data)
	// User's custom name should be preserved (old != base)
	if !strings.Contains(resultStr, "my-custom-name") {
		t.Errorf("3-way merge should preserve user's custom name, got:\n%s", resultStr)
	}
	// New feature from template should be present
	if !strings.Contains(resultStr, "new_feature") {
		t.Errorf("3-way merge should include new template field, got:\n%s", resultStr)
	}
}

func TestRestoreMoaiConfig_SkipsNonYAMLFiles(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, defs.MoAIDir, defs.ConfigSubdir)
	sectionsDir := filepath.Join(configDir, "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create backup with sections/ containing a non-YAML file
	backupDir := filepath.Join(tmpDir, "backup")
	backupSectionsDir := filepath.Join(backupDir, "sections")
	if err := os.MkdirAll(backupSectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a .json file in sections (should be skipped)
	if err := os.WriteFile(filepath.Join(backupSectionsDir, "notes.json"), []byte(`{"skip":"this"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write a .yaml file (should be processed)
	yamlContent := []byte("custom:\n  key: value\n")
	if err := os.WriteFile(filepath.Join(backupSectionsDir, "custom.yaml"), yamlContent, 0o644); err != nil {
		t.Fatal(err)
	}

	err := restoreMoaiConfig(tmpDir, backupDir)
	if err != nil {
		t.Fatalf("restoreMoaiConfig failed: %v", err)
	}

	// JSON file should NOT be restored
	jsonPath := filepath.Join(sectionsDir, "notes.json")
	if _, err := os.Stat(jsonPath); !os.IsNotExist(err) {
		t.Error("non-YAML file should not be restored from sections backup")
	}

	// YAML file should be restored (target does not exist, so restore as-is)
	yamlPath := filepath.Join(sectionsDir, "custom.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Error("YAML file should be restored from sections backup")
	}
}

// --- cleanMoaiManagedPaths additional tests ---

func TestCleanMoaiManagedPaths_OnlyUserFilesPreserved(t *testing.T) {
	root := t.TempDir()

	// Create user files in .claude/ that should NOT be deleted
	userAgentDir := filepath.Join(root, ".claude", "agents", "custom")
	if err := os.MkdirAll(userAgentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	userAgentFile := filepath.Join(userAgentDir, "my-agent.md")
	if err := os.WriteFile(userAgentFile, []byte("custom agent"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create moai-managed files
	moaiAgentDir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(moaiAgentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moaiAgentDir, "expert.md"), []byte("managed"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := cleanMoaiManagedPaths(root, &buf)
	if err != nil {
		t.Fatalf("cleanMoaiManagedPaths failed: %v", err)
	}

	// User's custom agent directory should still exist
	if _, err := os.Stat(userAgentFile); os.IsNotExist(err) {
		t.Error("user's custom agent file should be preserved")
	}

	// Moai-managed agent directory should be removed
	if _, err := os.Stat(moaiAgentDir); !os.IsNotExist(err) {
		t.Error("moai-managed agents directory should be removed")
	}
}

// --- detectGoBinPathForUpdate tests ---

func TestDetectGoBinPathForUpdate_ReturnsNonEmpty(t *testing.T) {
	homeDir := t.TempDir()
	result := detectGoBinPathForUpdate(homeDir)
	if result == "" {
		t.Fatal("detectGoBinPathForUpdate should return a non-empty string")
	}
}

func TestDetectGoBinPathForUpdate_FallbackWithHomeDir(t *testing.T) {
	// When GOBIN and GOPATH are empty (simulated by the fact that
	// the real go env commands return something), the function should
	// still return a valid path. We test the fallback by checking
	// the result is a valid path string.
	homeDir := t.TempDir()
	result := detectGoBinPathForUpdate(homeDir)

	// The result should be an absolute path (or at least non-empty)
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got: %s", result)
	}
}

func TestDetectGoBinPathForUpdate_EmptyHomeDir(t *testing.T) {
	// When homeDir is empty, should still return a valid path
	// (either from go env or the last resort fallback)
	result := detectGoBinPathForUpdate("")
	if result == "" {
		t.Fatal("detectGoBinPathForUpdate should return a non-empty string even with empty homeDir")
	}
}

func TestDetectGoBinPathForUpdate_ResultContainsBin(t *testing.T) {
	homeDir := t.TempDir()
	result := detectGoBinPathForUpdate(homeDir)

	// All possible results should contain "bin" in the path
	if !strings.Contains(result, "bin") {
		t.Errorf("expected path to contain 'bin', got: %s", result)
	}
}

// --- runTemplateSyncWithProgress tests ---
// This function requires a cobra.Command context, so we test minimal error paths.

func TestRunTemplateSyncWithProgress_VersionUpToDate(t *testing.T) {
	// Create a minimal cobra command with required flags
	tmpDir := t.TempDir()

	// Create the config structure with a version matching the current version
	sectionsDir := filepath.Join(tmpDir, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write system.yaml with the current version to trigger "up-to-date" path
	currentVersion := "0.0.0-test"
	systemContent := []byte("moai:\n  template_version: \"" + currentVersion + "\"\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.SystemYAML), systemContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// We cannot easily test the full function since it depends on cobra command
	// and many subsystems. Instead, verify the sub-component getProjectConfigVersion
	// works correctly.
	ver, err := getProjectConfigVersion(tmpDir)
	if err != nil {
		t.Fatalf("getProjectConfigVersion failed: %v", err)
	}
	if ver != currentVersion {
		t.Errorf("getProjectConfigVersion = %q, want %q", ver, currentVersion)
	}
}

func TestGetProjectConfigVersion_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// No system.yaml exists, should return "0.0.0" to force update
	ver, err := getProjectConfigVersion(tmpDir)
	if err != nil {
		t.Fatalf("getProjectConfigVersion should not error for missing file, got: %v", err)
	}
	if ver != "0.0.0" {
		t.Errorf("getProjectConfigVersion = %q, want %q for missing config", ver, "0.0.0")
	}
}

func TestGetProjectConfigVersion_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := []byte("moai:\n  template_version: \"2.5.0\"\n")
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.SystemYAML), content, 0o644); err != nil {
		t.Fatal(err)
	}

	ver, err := getProjectConfigVersion(tmpDir)
	if err != nil {
		t.Fatalf("getProjectConfigVersion failed: %v", err)
	}
	if ver != "2.5.0" {
		t.Errorf("getProjectConfigVersion = %q, want %q", ver, "2.5.0")
	}
}

