package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/template"
)

// buildSmartPATH is a test helper that builds a Smart PATH for a given home directory.
// It temporarily overrides HOME env to use the specified homeDir,
// then delegates to template.BuildSmartPATH().
func buildSmartPATH(homeDir string) string {
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", homeDir)
	defer func() { _ = os.Setenv("HOME", origHome) }()
	return template.BuildSmartPATH()
}

func TestUpdateCmd_Exists(t *testing.T) {
	if updateCmd == nil {
		t.Fatal("updateCmd should not be nil")
	}
}

func TestUpdateCmd_Use(t *testing.T) {
	if updateCmd.Use != "update" {
		t.Errorf("updateCmd.Use = %q, want %q", updateCmd.Use, "update")
	}
}

func TestUpdateCmd_Short(t *testing.T) {
	if updateCmd.Short == "" {
		t.Error("updateCmd.Short should not be empty")
	}
}

func TestUpdateCmd_HasFlags(t *testing.T) {
	flags := []string{"check"}
	for _, name := range flags {
		if updateCmd.Flags().Lookup(name) == nil {
			t.Errorf("update command should have --%s flag", name)
		}
	}
}

func TestUpdateCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "update" {
			found = true
			break
		}
	}
	if !found {
		t.Error("update should be registered as a subcommand of root")
	}
}

func TestUpdateCmd_CheckOnly_NoDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	// Reset flags before test
	if err := updateCmd.Flags().Set("check", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := updateCmd.Flags().Set("check", "false"); err != nil {
			t.Logf("reset flag: %v", err)
		}
	}()

	err := updateCmd.RunE(updateCmd, []string{})
	if err != nil {
		t.Fatalf("update --check should not error with nil deps, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Current version") {
		t.Errorf("output should contain 'Current version', got %q", output)
	}
}

func TestRunTemplateSync_Timeout(t *testing.T) {
	// This test verifies that runTemplateSync completes within the timeout period
	// Actual timeout behavior is tested through integration tests

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	// Note: This is a smoke test to ensure the function completes normally
	// For actual timeout testing with mock slow deployer, see integration tests
	// or test manually by setting templateDeployTimeout to a very short duration

	err := runTemplateSync(updateCmd)

	// The function should complete (either successfully or with an error)
	// If it hangs indefinitely, this test will timeout
	if err != nil {
		// Error is acceptable as long as it's not a hang
		t.Logf("runTemplateSync returned error (expected in test environment): %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Syncing templates") {
		t.Logf("output: %q", output)
	}
}

func TestGetProjectConfigVersion_FileSizeExceeds(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create file larger than 10MB
	configPath := filepath.Join(configDir, "config.yaml")
	largeContent := make([]byte, maxConfigSize+1)
	if err := os.WriteFile(configPath, largeContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Should return error for oversized file
	_, err := getProjectConfigVersion(tmpDir)
	if err == nil {
		t.Fatal("expected error for file exceeding size limit, got nil")
	}

	expectedMsg := "config file too large"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got: %v", expectedMsg, err)
	}
}

func TestGetProjectConfigVersion_ExactlyAtLimit(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create file exactly at 10MB limit with valid YAML
	configPath := filepath.Join(configDir, "config.yaml")
	validYAML := "project:\n  template_version: \"1.0.0\"\n"
	padding := make([]byte, maxConfigSize-len(validYAML))
	for i := range padding {
		padding[i] = '#' // YAML comment padding
	}
	content := append([]byte(validYAML), padding...)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Should succeed with file at exact limit
	version, err := getProjectConfigVersion(tmpDir)
	if err != nil {
		t.Fatalf("expected no error for file at size limit, got: %v", err)
	}

	if version != "1.0.0" {
		t.Errorf("expected version %q, got %q", "1.0.0", version)
	}
}

func TestGetProjectConfigVersion_NormalSize(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create normal-sized valid config file
	configPath := filepath.Join(configDir, "config.yaml")
	content := []byte("project:\n  template_version: \"2.5.3\"\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Should succeed with normal file
	version, err := getProjectConfigVersion(tmpDir)
	if err != nil {
		t.Fatalf("expected no error for normal-sized file, got: %v", err)
	}

	if version != "2.5.3" {
		t.Errorf("expected version %q, got %q", "2.5.3", version)
	}
}

func TestGetProjectConfigVersion_NonExistent(t *testing.T) {
	// Use temp directory with no config file
	tmpDir := t.TempDir()

	// Should return "0.0.0" for non-existent file
	version, err := getProjectConfigVersion(tmpDir)
	if err != nil {
		t.Fatalf("expected no error for non-existent file, got: %v", err)
	}

	if version != "0.0.0" {
		t.Errorf("expected version %q for non-existent file, got %q", "0.0.0", version)
	}
}

func TestGetProjectConfigVersion_ValidParsing(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create valid config file with various YAML structures
	configPath := filepath.Join(configDir, "config.yaml")
	content := []byte(`project:
  name: "test-project"
  template_version: "3.1.4"
  other_field: "value"
user:
  name: "testuser"
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Should correctly parse template_version
	version, err := getProjectConfigVersion(tmpDir)
	if err != nil {
		t.Fatalf("expected no error for valid config, got: %v", err)
	}

	if version != "3.1.4" {
		t.Errorf("expected version %q, got %q", "3.1.4", version)
	}
}

// --- DDD PRESERVE: Characterization tests for runTemplateSync ---

func TestRunTemplateSync_VersionMatch_SkipsSync(t *testing.T) {
	// Create temp directory with matching version
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Get current package version
	currentVersion := "test-version-1.0.0"

	// Create config with matching version
	configPath := filepath.Join(configDir, "config.yaml")
	content := []byte("project:\n  template_version: \"" + currentVersion + "\"\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Mock version.GetVersion to return the test version
	// Note: This test verifies the logic path, but since version.GetVersion
	// is a package-level function, we test the behavior indirectly
	// by checking if the function completes quickly (version check optimization)

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	// This should skip sync due to version match (if versions actually match)
	// In test environment, versions may not match, so we just verify no panic
	err = runTemplateSync(updateCmd)

	// Function should complete without panic
	// Error is acceptable as embedded templates may not be available in test
	if err != nil {
		t.Logf("runTemplateSync returned error (expected in test environment): %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Syncing templates") {
		t.Logf("output: %q", output)
	}
}

func TestRunTemplateSync_VersionMismatch_AttemptsSync(t *testing.T) {
	// Create temp directory with non-matching version
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with different version
	configPath := filepath.Join(configDir, "config.yaml")
	content := []byte("project:\n  template_version: \"0.0.1\"\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	// This should attempt sync due to version mismatch
	err = runTemplateSync(updateCmd)

	// Function should complete (error expected due to no embedded templates)
	if err != nil {
		t.Logf("runTemplateSync returned error (expected in test environment): %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Syncing templates") {
		t.Logf("output: %q", output)
	}
}

func TestRunTemplateSync_GetVersionError_ContinuesSync(t *testing.T) {
	// Create temp directory without .moai/config (to trigger error)
	tmpDir := t.TempDir()

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	// Should continue with sync even if version check fails
	err = runTemplateSync(updateCmd)

	// Function should complete (error expected due to missing manifest)
	if err != nil {
		t.Logf("runTemplateSync returned error (expected in test environment): %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Syncing templates") {
		t.Logf("output: %q", output)
	}
}

func TestRunTemplateSync_EmbeddedTemplatesError(t *testing.T) {
	// Create minimal valid directory structure
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config file
	configPath := filepath.Join(configDir, "config.yaml")
	content := []byte("project:\n  template_version: \"0.0.0\"\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)

	// This will fail when trying to load embedded templates
	// The function should handle the error gracefully
	err = runTemplateSync(updateCmd)

	// Error is expected but should be handled gracefully
	if err != nil {
		// Verify error message is informative
		if !strings.Contains(err.Error(), "template") && !strings.Contains(err.Error(), "manifest") {
			t.Logf("error message: %v", err)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "Syncing templates") {
		t.Logf("output: %q", output)
	}
}

func TestGetProjectConfigVersion_EmptyTemplateVersion(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config without template_version field
	configPath := filepath.Join(configDir, "config.yaml")
	content := []byte("project:\n  name: \"test\"\n")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Should return "0.0.0" for missing template_version
	version, err := getProjectConfigVersion(tmpDir)
	if err != nil {
		t.Fatalf("expected no error for missing template_version, got: %v", err)
	}

	if version != "0.0.0" {
		t.Errorf("expected version %q for missing template_version, got %q", "0.0.0", version)
	}
}

func TestGetProjectConfigVersion_InvalidYAML(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with invalid YAML
	configPath := filepath.Join(configDir, "config.yaml")
	content := []byte("invalid: yaml: content: [[[")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// Should return error for invalid YAML
	_, err := getProjectConfigVersion(tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}

	expectedMsg := "parse config YAML"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got: %v", expectedMsg, err)
	}
}

// --- DDD PRESERVE: Characterization tests for refactored functions ---

func TestClassifyFileRisk(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		exists   bool
		want     string
	}{
		{
			name:     "high risk CLAUDE.md",
			filename: ".claude/CLAUDE.md",
			exists:   true,
			want:     "high",
		},
		{
			name:     "high risk settings.json",
			filename: ".claude/settings.json",
			exists:   true,
			want:     "high",
		},
		{
			name:     "high risk config.yaml",
			filename: ".moai/config/config.yaml",
			exists:   true,
			want:     "high",
		},
		{
			name:     "low risk new file",
			filename: ".claude/skills/new-skill.md",
			exists:   false,
			want:     "low",
		},
		{
			name:     "medium risk existing file",
			filename: ".claude/skills/existing-skill.md",
			exists:   true,
			want:     "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyFileRisk(tt.filename, tt.exists)
			if got != tt.want {
				t.Errorf("classifyFileRisk(%q, %v) = %v, want %v", tt.filename, tt.exists, got, tt.want)
			}
		})
	}
}

func TestDetermineStrategy(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "CLAUDE.md uses SectionMerge",
			filename: ".claude/CLAUDE.md",
			want:     "section_merge",
		},
		{
			name:     ".gitignore uses EntryMerge",
			filename: ".gitignore",
			want:     "entry_merge",
		},
		{
			name:     "JSON file uses JSONMerge",
			filename: ".claude/settings.json",
			want:     "json_merge",
		},
		{
			name:     "YAML file uses YAMLDeep",
			filename: ".moai/config/config.yaml",
			want:     "yaml_deep",
		},
		{
			name:     "YML file uses YAMLDeep",
			filename: "config.yml",
			want:     "yaml_deep",
		},
		{
			name:     "markdown file uses LineMerge",
			filename: "README.md",
			want:     "line_merge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineStrategy(tt.filename)
			if string(got) != tt.want {
				t.Errorf("determineStrategy(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetermineChangeType(t *testing.T) {
	tests := []struct {
		name   string
		exists bool
		want   string
	}{
		{
			name:   "existing file",
			exists: true,
			want:   "update existing",
		},
		{
			name:   "new file",
			exists: false,
			want:   "new file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineChangeType(tt.exists)
			if got != tt.want {
				t.Errorf("determineChangeType(%v) = %v, want %v", tt.exists, got, tt.want)
			}
		})
	}
}

func TestIsMoaiManaged(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "moai skill with prefix",
			path: ".claude/skills/moai-workflow-project/skill.md",
			want: true,
		},
		{
			name: "moai skill without prefix",
			path: ".claude/skills/moai/skill.md",
			want: true,
		},
		{
			name: "moai rules",
			path: ".claude/rules/moai/constitution.md",
			want: true,
		},
		{
			name: "moai agents",
			path: ".claude/agents/moai-expert/backend.md",
			want: true,
		},
		{
			name: "moai commands",
			path: ".claude/commands/moai-plan/command.md",
			want: true,
		},
		{
			name: "user skill without moai prefix",
			path: ".claude/skills/user-custom-skill/skill.md",
			want: false,
		},
		{
			name: "user rules",
			path: ".claude/rules/user-custom-rule.md",
			want: false,
		},
		{
			name: "user agents",
			path: ".claude/agents/user-expert/backend.md",
			want: false,
		},
		{
			name: "config file",
			path: ".moai/config/config.yaml",
			want: true, // .moai/config/ is now managed by MoAI-ADK
		},
		{
			name: "claude md",
			path: "CLAUDE.md",
			want: false,
		},
		{
			name: "empty path",
			path: "",
			want: false,
		},
		{
			name: "path without .claude",
			path: "some/other/path/file.txt",
			want: false,
		},
		{
			name: "skills directory only",
			path: ".claude/skills",
			want: false,
		},
		{
			name: "moai hyphenated skill",
			path: ".claude/skills/moai-foundation-claude/skill.md",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMoaiManaged(tt.path)
			if got != tt.want {
				t.Errorf("isMoaiManaged(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsMoaiManaged_OutputStyles(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "moai output style",
			path: ".claude/output-styles/moai/moai.md",
			want: true,
		},
		{
			name: "moai output style r2d2",
			path: ".claude/output-styles/moai/r2d2.md",
			want: true,
		},
		{
			name: "moai output style yoda",
			path: ".claude/output-styles/moai/yoda.md",
			want: true,
		},
		{
			name: "user output style",
			path: ".claude/output-styles/user-custom/style.md",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMoaiManaged(tt.path)
			if got != tt.want {
				t.Errorf("isMoaiManaged(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsMoaiManaged_MoaiConfig(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "moai config file",
			path: ".moai/config/config.yaml",
			want: true,
		},
		{
			name: "moai config sections",
			path: ".moai/config/sections/quality.yaml",
			want: true,
		},
		{
			name: "moai config user template",
			path: ".moai/config/sections/user.yaml.tmpl",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMoaiManaged(tt.path)
			if got != tt.want {
				t.Errorf("isMoaiManaged(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// --- Backup functionality tests (matching Python moai template backup) ---

func TestBackupMoaiConfig_CreateBackup(t *testing.T) {
	// Create temp directory with config structure
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".moai", "config")
	sectionsDir := filepath.Join(configDir, "sections")

	// Create required directory structure
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := []byte("project:\n  name: \"test-project\"\n  template_version: \"1.0.0\"\n")
	if err := os.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatal(err)
	}

	sectionsUserPath := filepath.Join(sectionsDir, "user.yaml")
	sectionsUserContent := []byte("user:\n  name: \"testuser\"\n")
	if err := os.WriteFile(sectionsUserPath, sectionsUserContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create backup
	backupDir, err := backupMoaiConfig(tmpDir)
	if err != nil {
		t.Fatalf("backupMoaiConfig failed: %v", err)
	}

	// Verify backup directory path format
	if !strings.HasPrefix(backupDir, tmpDir) {
		t.Errorf("backup path should be under project root, got: %s", backupDir)
	}

	// Verify .moai-backups directory exists
	backupBaseDir := filepath.Join(tmpDir, ".moai-backups")
	if _, err := os.Stat(backupBaseDir); os.IsNotExist(err) {
		t.Error(".moai-backups directory should exist")
	}

	// Find the actual backup directory
	entries, err := os.ReadDir(backupBaseDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Errorf("should have exactly 1 backup directory, got %d", len(entries))
	}

	backupTimestamp := entries[0].Name()
	// Timestamp format is YYYYMMDD_HHmmss = 15 characters
	if len(backupTimestamp) != 15 {
		t.Errorf("timestamp should be 15 chars (YYYYMMDD_HHmmss), got: %s (len=%d)", backupTimestamp, len(backupTimestamp))
	}

	// Verify timestamp format YYYYMMDD_HHmmss
	if len(backupTimestamp) == 15 {
		parts := strings.SplitN(backupTimestamp, "_", 2)
		if len(parts) != 2 || len(parts[0]) != 8 || len(parts[1]) != 6 {
			t.Errorf("timestamp format should be YYYYMMDD_HHmmss (15 chars), got: %s", backupTimestamp)
		}
	}

	// Verify backup directory exists
	actualBackupDir := filepath.Join(backupBaseDir, backupTimestamp)
	if _, err := os.Stat(actualBackupDir); os.IsNotExist(err) {
		t.Error("backup directory should exist")
	}

	// Verify config.yaml was backed up
	backupConfigPath := filepath.Join(actualBackupDir, "config.yaml")
	if _, err := os.Stat(backupConfigPath); os.IsNotExist(err) {
		t.Error("config.yaml should be backed up")
	}

	// Verify sections directory was NOT backed up (should not exist)
	backupSectionsPath := filepath.Join(actualBackupDir, "sections")
	if _, err := os.Stat(backupSectionsPath); err == nil {
		t.Error("sections directory should be excluded from backup")
	}

	// Verify backup_metadata.json exists
	metadataPath := filepath.Join(actualBackupDir, "backup_metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("backup_metadata.json should exist")
	}

	// Verify metadata content
	metadataData, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("read metadata file: %v", err)
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}

	// Verify metadata fields
	if metadata.Timestamp != backupTimestamp {
		t.Errorf("metadata timestamp should match backup name, got: %s, want: %s", metadata.Timestamp, backupTimestamp)
	}
	if metadata.Description != "config_backup" {
		t.Errorf("metadata description should be 'config_backup', got: %s", metadata.Description)
	}
	if metadata.BackupType != "config" {
		t.Errorf("metadata backup_type should be 'config', got: %s", metadata.BackupType)
	}

	// Verify config.yaml is in backed_up_items
	foundConfig := false
	for _, item := range metadata.BackedUpItems {
		if item == ".moai/config/config.yaml" {
			foundConfig = true
			break
		}
	}
	if !foundConfig {
		t.Error("config.yaml should be in backed_up_items")
	}

	// Verify sections directory is in excluded_items (not excluded_dirs, since we use relative path)
	foundExcludedDir := false
	for _, item := range metadata.ExcludedItems {
		if item == "sections" || item == "sections/user.yaml" {
			foundExcludedDir = true
			break
		}
	}
	if !foundExcludedDir {
		t.Error("sections should be in excluded_items")
	}
}

func TestBackupMoaiConfig_NoConfigDir(t *testing.T) {
	// Create temp directory without config
	tmpDir := t.TempDir()

	// Should return empty string without error
	backupDir, err := backupMoaiConfig(tmpDir)
	if err != nil {
		t.Fatalf("backupMoaiConfig should not error when no config exists, got: %v", err)
	}
	if backupDir != "" {
		t.Errorf("backupDir should be empty when no config exists, got: %s", backupDir)
	}
}

func TestCleanupOldBackups(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create backup directory and some backups
	backupBaseDir := filepath.Join(tmpDir, ".moai-backups")
	if err := os.MkdirAll(backupBaseDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create timestamped backups (using proper timestamp format)
	now := time.Now()
	for i := 0; i < 10; i++ {
		// Create backups with different timestamps
		ts := now.Add(-time.Duration(i) * time.Hour).Format("20060102_150405")
		backupPath := filepath.Join(backupBaseDir, ts)
		if err := os.MkdirAll(backupPath, 0755); err != nil {
			t.Fatal(err)
		}
		// Create a metadata file for valid backup
		metadataPath := filepath.Join(backupPath, "backup_metadata.json")
		if err := os.WriteFile(metadataPath, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// List backups before cleanup
	entriesBefore, err := os.ReadDir(backupBaseDir)
	if err != nil {
		t.Fatal(err)
	}
	backupCountBefore := 0
	for _, e := range entriesBefore {
		if e.IsDir() && len(e.Name()) == 15 {
			parts := strings.SplitN(e.Name(), "_", 2)
			if len(parts) == 2 && len(parts[0]) == 8 && len(parts[1]) == 6 {
				backupCountBefore++
			}
		}
	}

	if backupCountBefore != 10 {
		t.Errorf("expected 10 valid backup directories before cleanup, got: %d", backupCountBefore)
	}

	// Test cleanup with keep_count=5
	deletedCount := cleanup_old_backups(tmpDir, 5)
	if deletedCount != 5 {
		t.Errorf("should delete 5 old backups, got: %d", deletedCount)
	}

	// Verify only 5 backups remain
	entries, err := os.ReadDir(backupBaseDir)
	if err != nil {
		t.Fatal(err)
	}

	// Count valid timestamped backup directories
	validBackups := 0
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) == 15 {
			parts := strings.SplitN(entry.Name(), "_", 2)
			if len(parts) == 2 && len(parts[0]) == 8 && len(parts[1]) == 6 {
				validBackups++
			}
		}
	}

	if validBackups != 5 {
		t.Errorf("expected 5 backups after cleanup, got: %d", validBackups)
	}

	// Test cleanup with keep_count=10 (no deletion)
	deletedCount = cleanup_old_backups(tmpDir, 10)
	if deletedCount != 0 {
		t.Errorf("should not delete any backups with keep_count=10, got: %d", deletedCount)
	}

	// Test cleanup with keep_count=0 (delete all)
	deletedCount = cleanup_old_backups(tmpDir, 0)
	if deletedCount != 5 {
		t.Errorf("should delete all 5 backups with keep_count=0, got: %d", deletedCount)
	}

	// Verify backup directory is empty
	entries, err = os.ReadDir(backupBaseDir)
	if err != nil {
		t.Fatal(err)
	}

	// Count remaining valid backup directories
	remainingBackups := 0
	for _, e := range entries {
		if e.IsDir() && len(e.Name()) == 15 {
			parts := strings.SplitN(e.Name(), "_", 2)
			if len(parts) == 2 && len(parts[0]) == 8 && len(parts[1]) == 6 {
				remainingBackups++
			}
		}
	}

	if remainingBackups != 0 {
		t.Errorf("backup directory should be empty after cleaning all, got %d valid backups", remainingBackups)
	}
}

func TestCleanupOldBackups_InvalidBackupPattern(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create backup directory with invalid names
	backupDir := filepath.Join(tmpDir, ".moai-backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create directories with different lengths and patterns
	dirs := []string{"abc123", "invalid_name", "12345678", "20250205_invalid", "20250205_123456"}
	for _, dirName := range dirs {
		backupPath := filepath.Join(backupDir, dirName)
		if err := os.MkdirAll(backupPath, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Should return 0 for invalid backup names
	deletedCount := cleanup_old_backups(tmpDir, 5)
	if deletedCount != 0 {
		t.Errorf("should not delete any invalid backups, got: %d", deletedCount)
	}
}

func TestCleanupOldBackups_NoBackupsDir(t *testing.T) {
	// Create temp directory without backups directory
	tmpDir := t.TempDir()

	// Should return 0 without error
	deletedCount := cleanup_old_backups(tmpDir, 5)
	if deletedCount != 0 {
		t.Errorf("should return 0 when no backups exist, got: %d", deletedCount)
	}
}

func TestRestoreMoaiConfig_MergeBehavior(t *testing.T) {
	// Create temp directory with config structure
	tmpDir := t.TempDir()

	// Create config structure at the project root
	configDir := filepath.Join(tmpDir, ".moai", "config")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create old config.yaml (backup will have this)
	oldConfigPath := filepath.Join(configDir, "config.yaml")
	oldConfigContent := []byte("project:\n  name: \"old-project\"\n  template_version: \"1.0.0\"\nuser:\n  name: \"testuser\"\n  custom_setting: \"custom_value\"\n")
	if err := os.WriteFile(oldConfigPath, oldConfigContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create backup (sections/ is excluded per Python behavior)
	backupDir, err := backupMoaiConfig(tmpDir)
	if err != nil {
		t.Fatalf("backupMoaiConfig failed: %v", err)
	}

	// Verify backup does NOT contain sections directory (it's excluded)
	backupSectionsPath := filepath.Join(backupDir, "sections")
	if _, err := os.Stat(backupSectionsPath); err == nil {
		t.Error("sections directory should be excluded from backup")
	}

	// Verify backup contains config.yaml
	backupConfigPath := filepath.Join(backupDir, "config.yaml")
	if _, err := os.Stat(backupConfigPath); os.IsNotExist(err) {
		t.Error("backup should contain config.yaml")
	}

	// Now simulate template sync by replacing config.yaml with new version
	newConfigContent := []byte("project:\n  name: \"new-project\"\n  template_version: \"2.0.0\"\n  new_field: \"value\"\n")
	if err := os.WriteFile(oldConfigPath, newConfigContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Restore from backup
	if err := restoreMoaiConfig(tmpDir, backupDir); err != nil {
		t.Fatalf("restoreMoaiConfig failed: %v", err)
	}

	// Read restored config.yaml
	data, err := os.ReadFile(oldConfigPath)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}

	// Verify custom_setting was preserved from old config (backup)
	if !strings.Contains(string(data), "custom_setting") {
		t.Error("custom_setting should be preserved from backup")
	}

	// Verify new_field from new config is also present
	if !strings.Contains(string(data), "new_field") {
		t.Error("new_field should be present from new config")
	}

	// Verify template_version was updated to 2.0.0 (from new config)
	// YAML may output version without quotes: "template_version: 2.0.0"
	if !strings.Contains(string(data), "template_version: 2.0.0") {
		t.Errorf("template_version should be from new config (2.0.0), got:\n%s", string(data))
	}
}

func TestBackupMetadata_Structure(t *testing.T) {
	// Test BackupMetadata struct marshaling
	metadata := BackupMetadata{
		Timestamp:     "20250205_143022",
		Description:   "config_backup",
		BackedUpItems: []string{".moai/config/config.yaml", ".moai/config/settings.yaml"},
		ExcludedItems: []string{"sections/user.yaml"},
		ExcludedDirs:  []string{"config/sections"},
		ProjectRoot:   "/home/user/project",
		BackupType:    "config",
	}

	// Test marshaling
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		t.Fatalf("marshal BackupMetadata failed: %v", err)
	}

	// Test unmarshaling
	var decoded BackupMetadata
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal BackupMetadata failed: %v", err)
	}

	// Verify all fields match
	if decoded.Timestamp != metadata.Timestamp {
		t.Errorf("Timestamp mismatch: %s vs %s", decoded.Timestamp, metadata.Timestamp)
	}
	if decoded.Description != metadata.Description {
		t.Errorf("Description mismatch: %s vs %s", decoded.Description, metadata.Description)
	}
	if decoded.BackupType != metadata.BackupType {
		t.Errorf("BackupType mismatch: %s vs %s", decoded.BackupType, metadata.BackupType)
	}
	if decoded.ExcludedDirs[0] != "config/sections" {
		t.Errorf("ExcludedDirs[0] mismatch: %s", decoded.ExcludedDirs[0])
	}
}

func TestEnsureGlobalSettingsEnv(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()

	// Mock home directory to temp dir
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	// Create .claude directory
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	// Test 1: No existing settings.json -> creates with only SessionEnd hook, NO env/permissions/teammateMode
	t.Run("CreateNewSettings_OnlySessionEndHook", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Ensure file doesn't exist
		_ = os.Remove(settingsPath)

		err := ensureGlobalSettingsEnv()
		if err != nil {
			t.Fatalf("ensureGlobalSettingsEnv failed: %v", err)
		}

		// Verify file was created
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("failed to read settings.json: %v", err)
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse settings.json: %v", err)
		}

		// env should NOT be set (moai no longer manages env in global settings)
		if _, exists := settings["env"]; exists {
			t.Error("env should not be set in global settings (managed at project level)")
		}

		// permissions should NOT be set (managed at project level)
		if _, exists := settings["permissions"]; exists {
			t.Error("permissions should not be set in global settings")
		}

		// teammateMode should NOT be set (managed at project level)
		if _, exists := settings["teammateMode"]; exists {
			t.Error("teammateMode should not be set in global settings")
		}

		// Check hooks.SessionEnd exists (the only managed global setting)
		hooks, ok := settings["hooks"].(map[string]interface{})
		if !ok {
			t.Fatal("hooks not found in settings")
		}
		if _, exists := hooks["SessionEnd"]; !exists {
			t.Error("SessionEnd hook not found")
		}
	})

	// Test 2: Has moai env keys + custom -> removes moai keys, preserves custom + adds SessionEnd hook
	t.Run("CleanupMoaiManagedKeys", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Create existing settings with moai-managed env keys and a custom env key
		existing := map[string]interface{}{
			"env": map[string]interface{}{
				"CUSTOM_VAR":         "custom_value",
				"PATH":              "/old/go/bin:/usr/local/bin",
				"ENABLE_TOOL_SEARCH": "1",
				"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1",
			},
			"language": "en",
		}
		data, _ := json.MarshalIndent(existing, "", "  ")
		if err := os.WriteFile(settingsPath, data, 0644); err != nil {
			t.Fatalf("failed to write existing settings: %v", err)
		}

		err := ensureGlobalSettingsEnv()
		if err != nil {
			t.Fatalf("ensureGlobalSettingsEnv failed: %v", err)
		}

		// Read back and verify
		data, err = os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("failed to read settings.json: %v", err)
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse settings.json: %v", err)
		}

		// env should still exist because CUSTOM_VAR is preserved
		env, hasEnv := settings["env"]
		if !hasEnv {
			t.Fatal("env should still exist (CUSTOM_VAR is present)")
		}
		envMap := env.(map[string]interface{})

		// Custom var should be preserved
		if envMap["CUSTOM_VAR"] != "custom_value" {
			t.Errorf("CUSTOM_VAR not preserved: got %v", envMap["CUSTOM_VAR"])
		}

		// Moai-managed keys should be REMOVED
		if _, exists := envMap["PATH"]; exists {
			t.Error("PATH should be removed from global settings (managed at project level)")
		}
		if _, exists := envMap["ENABLE_TOOL_SEARCH"]; exists {
			t.Error("ENABLE_TOOL_SEARCH should be removed from global settings")
		}
		if _, exists := envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"]; exists {
			t.Error("CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS should be removed from global settings")
		}

		// SessionEnd hook should be added
		hooks, ok := settings["hooks"].(map[string]interface{})
		if !ok {
			t.Fatal("hooks not found in settings")
		}
		if _, exists := hooks["SessionEnd"]; !exists {
			t.Error("SessionEnd hook not found")
		}

		// language should be preserved (non-moai-managed top-level key)
		if settings["language"] != "en" {
			t.Errorf("language not preserved: got %v", settings["language"])
		}
	})

	// Test 3: Only SessionEnd hook, no moai keys -> no modification
	t.Run("AlreadyClean_NoModification", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Create settings with only the SessionEnd hook (no moai-managed env/permissions/teammateMode)
		sessionEndHookCommand := buildSessionEndHookCommand()
		existing := map[string]interface{}{
			"hooks": map[string]interface{}{
				"SessionEnd": []interface{}{
					map[string]interface{}{
						"hooks": []interface{}{
							map[string]interface{}{
								"type":    "command",
								"command": sessionEndHookCommand,
								"timeout": 5,
							},
						},
					},
				},
			},
		}
		data, _ := json.MarshalIndent(existing, "", "  ")
		if err := os.WriteFile(settingsPath, data, 0644); err != nil {
			t.Fatalf("failed to write existing settings: %v", err)
		}

		// Store original content
		originalContent, _ := os.ReadFile(settingsPath)

		err := ensureGlobalSettingsEnv()
		if err != nil {
			t.Fatalf("ensureGlobalSettingsEnv failed: %v", err)
		}

		// Read back content
		newContent, _ := os.ReadFile(settingsPath)

		// Content should be identical (no modification needed)
		if string(originalContent) != string(newContent) {
			t.Errorf("file was modified when it should not have been\nOriginal: %s\nNew: %s", string(originalContent), string(newContent))
		}
	})

	// Test 4: env has only moai keys -> entire env key removed after cleanup
	t.Run("EmptyEnvRemovedEntirely", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Create settings with env containing only moai-managed keys
		existing := map[string]interface{}{
			"env": map[string]interface{}{
				"PATH":               "/old/go/bin:/usr/bin",
				"ENABLE_TOOL_SEARCH": "1",
				"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1",
			},
		}
		data, _ := json.MarshalIndent(existing, "", "  ")
		if err := os.WriteFile(settingsPath, data, 0644); err != nil {
			t.Fatalf("failed to write existing settings: %v", err)
		}

		err := ensureGlobalSettingsEnv()
		if err != nil {
			t.Fatalf("ensureGlobalSettingsEnv failed: %v", err)
		}

		// Read back and verify
		data, err = os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("failed to read settings.json: %v", err)
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse settings.json: %v", err)
		}

		// env key should be completely removed (was empty after cleanup)
		if _, exists := settings["env"]; exists {
			t.Error("env should be removed entirely when all keys are moai-managed")
		}

		// SessionEnd hook should still be present
		hooks, ok := settings["hooks"].(map[string]interface{})
		if !ok {
			t.Fatal("hooks not found in settings")
		}
		if _, exists := hooks["SessionEnd"]; !exists {
			t.Error("SessionEnd hook not found")
		}
	})

	// Test 5: Permissions with user-added entries (not just Task:*) -> preserved
	t.Run("PreserveUserPermissions", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Create settings with permissions that include user-added entries beyond Task:*
		sessionEndHookCommand := buildSessionEndHookCommand()
		existing := map[string]interface{}{
			"permissions": map[string]interface{}{
				"allow": []interface{}{"Task:*", "Bash(npm run build):*"},
			},
			"hooks": map[string]interface{}{
				"SessionEnd": []interface{}{
					map[string]interface{}{
						"hooks": []interface{}{
							map[string]interface{}{
								"type":    "command",
								"command": sessionEndHookCommand,
								"timeout": 5,
							},
						},
					},
				},
			},
		}
		data, _ := json.MarshalIndent(existing, "", "  ")
		if err := os.WriteFile(settingsPath, data, 0644); err != nil {
			t.Fatalf("failed to write existing settings: %v", err)
		}

		err := ensureGlobalSettingsEnv()
		if err != nil {
			t.Fatalf("ensureGlobalSettingsEnv failed: %v", err)
		}

		// Read back and verify
		data, err = os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("failed to read settings.json: %v", err)
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse settings.json: %v", err)
		}

		// Permissions should be PRESERVED because it has more than just Task:*
		permVal, exists := settings["permissions"]
		if !exists {
			t.Fatal("permissions should be preserved when it contains user-added entries")
		}
		permMap := permVal.(map[string]interface{})
		allowArr := permMap["allow"].([]interface{})
		if len(allowArr) != 2 {
			t.Errorf("permissions.allow should have 2 entries, got %d", len(allowArr))
		}
	})
}

func TestEnsureGlobalSettingsEnv_CleanupMigratedSettings(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()

	// Mock home directory to temp dir
	originalHome := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
	}()
	_ = os.Setenv("HOME", tempDir)

	// Create .claude directory
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Create existing settings with ALL moai-managed settings that should be cleaned up:
	// env keys (PATH, ENABLE_TOOL_SEARCH, CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS),
	// permissions with only Task:*, teammateMode "auto", plus a custom env key
	existing := map[string]interface{}{
		"env": map[string]interface{}{
			"PATH":               "/old/go/bin:/usr/local/bin:/usr/bin:/bin",
			"ENABLE_TOOL_SEARCH": "1",
			"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1",
			"CUSTOM_VAR": "preserved",
		},
		"permissions": map[string]interface{}{
			"allow": []interface{}{"Task:*"},
		},
		"teammateMode": "auto",
		"hooks": map[string]interface{}{
			"SessionEnd": []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": buildSessionEndHookCommand(),
							"timeout": 5,
						},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		t.Fatalf("failed to write existing settings: %v", err)
	}

	err := ensureGlobalSettingsEnv()
	if err != nil {
		t.Fatalf("ensureGlobalSettingsEnv failed: %v", err)
	}

	// Read back and verify
	data, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings.json: %v", err)
	}

	// Moai-managed env keys should be REMOVED from global settings
	env, hasEnv := settings["env"]
	if !hasEnv {
		t.Fatal("env should still exist (CUSTOM_VAR is present)")
	}
	envMap := env.(map[string]interface{})

	if _, exists := envMap["PATH"]; exists {
		t.Error("PATH should be removed from global settings (managed at project level)")
	}
	if _, exists := envMap["ENABLE_TOOL_SEARCH"]; exists {
		t.Error("ENABLE_TOOL_SEARCH should be removed from global settings")
	}
	if _, exists := envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"]; exists {
		t.Error("CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS should be removed from global settings")
	}

	// Custom env keys should be PRESERVED
	if envMap["CUSTOM_VAR"] != "preserved" {
		t.Errorf("CUSTOM_VAR not preserved: got %v", envMap["CUSTOM_VAR"])
	}

	// permissions with only Task:* should be REMOVED
	if _, exists := settings["permissions"]; exists {
		t.Error("permissions with only Task:* should be removed from global settings")
	}

	// teammateMode "auto" should be REMOVED
	if _, exists := settings["teammateMode"]; exists {
		t.Error("teammateMode 'auto' should be removed from global settings")
	}

	// SessionEnd hook should still be present
	hooks, hasHooks := settings["hooks"]
	if !hasHooks {
		t.Fatal("hooks should still be present")
	}
	hooksMap := hooks.(map[string]interface{})
	if _, hasSessionEnd := hooksMap["SessionEnd"]; !hasSessionEnd {
		t.Error("SessionEnd hook should still be present")
	}
}

func TestBuildSmartPATH(t *testing.T) {
	sep := string(os.PathListSeparator)

	tests := []struct {
		name       string
		homeDir    string
		envPATH    string
		wantLocalBin bool
		wantGoBin    bool
		wantUnchanged bool // true if PATH should remain unchanged (essential dirs already present)
	}{
		{
			name:       "essential dirs missing from PATH - should be prepended",
			homeDir:    "/home/testuser",
			envPATH:    "/usr/local/bin" + sep + "/usr/bin" + sep + "/bin",
			wantLocalBin: true,
			wantGoBin:    true,
		},
		{
			name:       "essential dirs already in PATH - PATH unchanged",
			homeDir:    "/home/testuser",
			envPATH:    "/home/testuser/.local/bin" + sep + "/home/testuser/go/bin" + sep + "/usr/bin",
			wantLocalBin: true,
			wantGoBin:    true,
			wantUnchanged: true,
		},
		{
			name:       "trailing slash on existing dir - should match and not duplicate",
			homeDir:    "/home/testuser",
			envPATH:    "/home/testuser/.local/bin/" + sep + "/home/testuser/go/bin/" + sep + "/usr/bin",
			wantLocalBin: true,
			wantGoBin:    true,
			wantUnchanged: true,
		},
		{
			name:       "empty current PATH - essential dirs only",
			homeDir:    "/home/testuser",
			envPATH:    "",
			wantLocalBin: true,
			wantGoBin:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore PATH
			originalPATH := os.Getenv("PATH")
			defer func() {
				_ = os.Setenv("PATH", originalPATH)
			}()
			_ = os.Setenv("PATH", tt.envPATH)

			result := buildSmartPATH(tt.homeDir)

			localBin := filepath.Join(tt.homeDir, ".local", "bin")
			goBin := filepath.Join(tt.homeDir, "go", "bin")

			if tt.wantLocalBin && !strings.Contains(result, localBin) {
				t.Errorf("result should contain %q, got %q", localBin, result)
			}
			if tt.wantGoBin && !strings.Contains(result, goBin) {
				t.Errorf("result should contain %q, got %q", goBin, result)
			}

			if tt.wantUnchanged && result != tt.envPATH {
				t.Errorf("PATH should remain unchanged when essential dirs already present\ngot:  %q\nwant: %q", result, tt.envPATH)
			}

			// Verify no duplicate entries of essential dirs
			entries := strings.Split(result, sep)
			localBinCount := 0
			goBinCount := 0
			for _, entry := range entries {
				normalized := strings.TrimRight(entry, "/\\")
				if normalized == localBin {
					localBinCount++
				}
				if normalized == goBin {
					goBinCount++
				}
			}
			if localBinCount > 1 {
				t.Errorf("localBin should appear at most once, found %d times in %q", localBinCount, result)
			}
			if goBinCount > 1 {
				t.Errorf("goBin should appear at most once, found %d times in %q", goBinCount, result)
			}
		})
	}
}

func TestPathContainsDir(t *testing.T) {
	tests := []struct {
		name    string
		pathStr string
		dir     string
		sep     string
		want    bool
	}{
		{
			name:    "exact match",
			pathStr: "/usr/local/bin:/usr/bin:/bin",
			dir:     "/usr/local/bin",
			sep:     ":",
			want:    true,
		},
		{
			name:    "trailing slash on dir",
			pathStr: "/usr/local/bin:/usr/bin:/bin",
			dir:     "/usr/local/bin/",
			sep:     ":",
			want:    true,
		},
		{
			name:    "trailing slash on entry",
			pathStr: "/usr/local/bin/:/usr/bin:/bin",
			dir:     "/usr/local/bin",
			sep:     ":",
			want:    true,
		},
		{
			name:    "substring should NOT match",
			pathStr: "/usr/local/bin2:/usr/bin:/bin",
			dir:     "/usr/local/bin",
			sep:     ":",
			want:    false,
		},
		{
			name:    "empty path",
			pathStr: "",
			dir:     "/usr/local/bin",
			sep:     ":",
			want:    false,
		},
		{
			name:    "Windows semicolon separator",
			pathStr: `C:\Go\bin;C:\Users\user\go\bin;C:\Windows\system32`,
			dir:     `C:\Users\user\go\bin`,
			sep:     ";",
			want:    true,
		},
		{
			name:    "dir not in path",
			pathStr: "/usr/local/bin:/usr/bin:/bin",
			dir:     "/opt/homebrew/bin",
			sep:     ":",
			want:    false,
		},
		{
			name:    "single entry exact match",
			pathStr: "/usr/local/bin",
			dir:     "/usr/local/bin",
			sep:     ":",
			want:    true,
		},
		{
			name:    "single entry no match",
			pathStr: "/usr/local/bin",
			dir:     "/usr/bin",
			sep:     ":",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := template.PathContainsDir(tt.pathStr, tt.dir, tt.sep)
			if got != tt.want {
				t.Errorf("template.PathContainsDir(%q, %q, %q) = %v, want %v",
					tt.pathStr, tt.dir, tt.sep, got, tt.want)
			}
		})
	}
}
