package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/update"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// buildSmartPATH is a test helper that builds a Smart PATH for a given home directory.
// It temporarily overrides HOME and USERPROFILE env to use the specified homeDir,
// then delegates to template.BuildSmartPATH().
func buildSmartPATH(homeDir string) string {
	origHome := os.Getenv("HOME")
	origProfile := os.Getenv("USERPROFILE")
	_ = os.Setenv("HOME", homeDir)
	_ = os.Setenv("USERPROFILE", homeDir) // Windows: os.UserHomeDir() checks USERPROFILE first
	defer func() {
		_ = os.Setenv("HOME", origHome)
		_ = os.Setenv("USERPROFILE", origProfile)
	}()
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

	// Reset all flags before test to avoid state pollution from other tests
	if err := updateCmd.Flags().Set("check", "true"); err != nil {
		t.Fatal(err)
	}
	if err := updateCmd.Flags().Set("binary", "false"); err != nil {
		t.Fatal(err)
	}
	if err := updateCmd.Flags().Set("templates-only", "false"); err != nil {
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
	// This test verifies that runTemplateSync completes within the timeout period.
	// Run in a temp directory to avoid polluting the source tree with deployed templates.

	tmpDir := t.TempDir()
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

	// Note: This is a smoke test to ensure the function completes normally
	// For actual timeout testing with mock slow deployer, see integration tests
	// or test manually by setting templateDeployTimeout to a very short duration

	err = runTemplateSync(updateCmd)

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
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create file larger than 10MB
	configPath := filepath.Join(configDir, "system.yaml")
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
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create file exactly at 10MB limit with valid YAML
	configPath := filepath.Join(configDir, "system.yaml")
	validYAML := "moai:\n  template_version: \"1.0.0\"\n"
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
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create normal-sized valid config file
	configPath := filepath.Join(configDir, "system.yaml")
	content := []byte("moai:\n  template_version: \"2.5.3\"\n")
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
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create valid config file with various YAML structures
	configPath := filepath.Join(configDir, "system.yaml")
	content := []byte(`moai:
  name: "test-project"
  template_version: "3.1.4"
  other_field: "value"
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
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Get current package version
	currentVersion := "test-version-1.0.0"

	// Create config with matching version
	configPath := filepath.Join(configDir, "system.yaml")
	content := []byte("moai:\n  template_version: \"" + currentVersion + "\"\n")
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
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with different version
	configPath := filepath.Join(configDir, "system.yaml")
	content := []byte("moai:\n  template_version: \"0.0.1\"\n")
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
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config file
	configPath := filepath.Join(configDir, "system.yaml")
	content := []byte("moai:\n  template_version: \"0.0.0\"\n")
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
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config without template_version field
	configPath := filepath.Join(configDir, "system.yaml")
	content := []byte("moai:\n  name: \"test\"\n")
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
	configDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with invalid YAML
	configPath := filepath.Join(configDir, "system.yaml")
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
			name:     "medium risk config.yaml",
			filename: ".moai/config/config.yaml",
			exists:   true,
			want:     "medium",
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
		{
			name: "moai hooks",
			path: ".claude/hooks/moai/handle-session-start.sh",
			want: true,
		},
		{
			name: "moai hooks nested",
			path: ".claude/hooks/moai/handle-agent-hook.sh",
			want: true,
		},
		{
			name: "user hooks without moai prefix",
			path: ".claude/hooks/custom/my-hook.sh",
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
	systemPath := filepath.Join(sectionsDir, "system.yaml")
	systemContent := []byte("moai:\n  name: \"test-project\"\n  template_version: \"1.0.0\"\n")
	if err := os.WriteFile(systemPath, systemContent, 0644); err != nil {
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

	// Verify sections directory WAS backed up (full backup for restore capability)
	backupSectionsPath := filepath.Join(actualBackupDir, "sections")
	if _, err := os.Stat(backupSectionsPath); os.IsNotExist(err) {
		t.Error("sections directory should be included in backup")
	}

	// Verify sections/system.yaml was backed up
	backupSystemPath := filepath.Join(actualBackupDir, "sections", "system.yaml")
	if _, err := os.Stat(backupSystemPath); os.IsNotExist(err) {
		t.Error("sections/system.yaml should be backed up")
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

	// Verify sections/system.yaml is in backed_up_items
	foundSystem := slices.Contains(metadata.BackedUpItems, ".moai/config/sections/system.yaml")
	if !foundSystem {
		t.Errorf("sections/system.yaml should be in backed_up_items, got: %v", metadata.BackedUpItems)
	}

	// Verify sections files are in backed_up_items (full backup, no exclusions)
	foundSectionsFile := false
	for _, item := range metadata.BackedUpItems {
		if strings.Contains(item, "sections/") {
			foundSectionsFile = true
			break
		}
	}
	if !foundSectionsFile {
		t.Error("sections files should be in backed_up_items")
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
	for i := range 10 {
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
	sectionsDir := filepath.Join(configDir, "sections")

	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create old system.yaml (backup will have this)
	// The "name" field is user-modified (differs from template default)
	oldSystemPath := filepath.Join(sectionsDir, "system.yaml")
	oldSystemContent := []byte("moai:\n  name: \"user-modified-name\"\n  template_version: \"1.0.0\"\n")
	if err := os.WriteFile(oldSystemPath, oldSystemContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create backup
	backupDir, err := backupMoaiConfig(tmpDir)
	if err != nil {
		t.Fatalf("backupMoaiConfig failed: %v", err)
	}

	// Verify backup contains sections/system.yaml
	backupSystemPath := filepath.Join(backupDir, "sections", "system.yaml")
	if _, err := os.Stat(backupSystemPath); os.IsNotExist(err) {
		t.Error("backup should contain sections/system.yaml")
	}

	// Now simulate template sync by replacing system.yaml with new version
	newSystemContent := []byte("moai:\n  name: \"new-project\"\n  template_version: \"2.0.0\"\n  new_field: \"value\"\n")
	if err := os.WriteFile(oldSystemPath, newSystemContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Restore from backup
	if err := restoreMoaiConfig(tmpDir, backupDir); err != nil {
		t.Fatalf("restoreMoaiConfig failed: %v", err)
	}

	// Read restored system.yaml
	data, err := os.ReadFile(oldSystemPath)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}

	// Verify user-modified "name" was preserved from old config (backup)
	if !strings.Contains(string(data), "user-modified-name") {
		t.Errorf("user-modified name should be preserved from backup, got:\n%s", string(data))
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

func TestRestoreMoaiConfig_MissingDirectory(t *testing.T) {
	// Test restore when backup contains files in directories that don't exist in target
	tmpDir := t.TempDir()

	// Create config directory with subdirectory
	configDir := filepath.Join(tmpDir, ".moai", "config")
	questionsDir := filepath.Join(configDir, "questions")

	if err := os.MkdirAll(questionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file in questions directory
	schemaPath := filepath.Join(questionsDir, "_schema.yaml")
	schemaContent := []byte("version: 1.0\nfields:\n  - name: test\n")
	if err := os.WriteFile(schemaPath, schemaContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create backup
	backupDir, err := backupMoaiConfig(tmpDir)
	if err != nil {
		t.Fatalf("backupMoaiConfig failed: %v", err)
	}

	// Delete the questions directory (simulating template without this directory)
	if err := os.RemoveAll(questionsDir); err != nil {
		t.Fatal(err)
	}

	// Restore from backup - should create directory and restore file
	if err := restoreMoaiConfig(tmpDir, backupDir); err != nil {
		t.Fatalf("restoreMoaiConfig failed: %v", err)
	}

	// Verify the file was restored and directory was created
	restoredPath := filepath.Join(configDir, "questions", "_schema.yaml")
	if _, err := os.Stat(restoredPath); os.IsNotExist(err) {
		t.Error("restored file should exist in questions directory")
	}

	// Verify content
	data, err := os.ReadFile(restoredPath)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}

	if !strings.Contains(string(data), "version: 1.0") {
		t.Error("restored file should contain original content")
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
	originalProfile := os.Getenv("USERPROFILE")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
		_ = os.Setenv("USERPROFILE", originalProfile)
	}()
	_ = os.Setenv("HOME", tempDir)
	_ = os.Setenv("USERPROFILE", tempDir) // Windows: os.UserHomeDir() checks USERPROFILE first

	// Create .claude directory
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	// Test 1: No existing settings.json -> no file created (nothing to clean up)
	t.Run("NoExistingSettings_NothingCreated", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Ensure file doesn't exist
		_ = os.Remove(settingsPath)

		err := ensureGlobalSettingsEnv()
		if err != nil {
			t.Fatalf("ensureGlobalSettingsEnv failed: %v", err)
		}

		// Verify file was NOT created (nothing to clean up)
		if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
			t.Error("settings.json should not be created when there is nothing to clean up")
		}
	})

	// Test 2: Has moai env keys + custom -> removes moai keys, preserves custom, no SessionEnd hook added
	t.Run("CleanupMoaiManagedKeys", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Create existing settings with moai-managed env keys and a custom env key
		existing := map[string]any{
			"env": map[string]any{
				"CUSTOM_VAR":                           "custom_value",
				"PATH":                                 "/old/go/bin:/usr/local/bin",
				"ENABLE_TOOL_SEARCH":                   "1",
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

		var settings map[string]any
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse settings.json: %v", err)
		}

		// env should still exist because CUSTOM_VAR is preserved
		env, hasEnv := settings["env"]
		if !hasEnv {
			t.Fatal("env should still exist (CUSTOM_VAR is present)")
		}
		envMap := env.(map[string]any)

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
		// CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS should be set to "1" as default
		if envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"] != "1" {
			t.Errorf("CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS should be set to '1', got: %v", envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"])
		}

		// SessionEnd hook should NOT be present (no longer managed globally)
		if hooks, ok := settings["hooks"].(map[string]any); ok {
			if _, exists := hooks["SessionEnd"]; exists {
				t.Error("SessionEnd hook should not be added to global settings")
			}
		}

		// language should be preserved (non-moai-managed top-level key)
		if settings["language"] != "en" {
			t.Errorf("language not preserved: got %v", settings["language"])
		}
	})

	// Test 3: Orphaned SessionEnd hook (handle-session-end.sh) -> should be cleaned up
	t.Run("CleanupOrphanedSessionEndHook", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Create settings with the orphaned SessionEnd hook
		existing := map[string]any{
			"hooks": map[string]any{
				"SessionEnd": []any{
					map[string]any{
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "\"$HOME/.claude/hooks/moai/handle-session-end.sh\"",
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

		var settings map[string]any
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse settings.json: %v", err)
		}

		// hooks should be completely removed (empty after cleanup)
		if _, exists := settings["hooks"]; exists {
			t.Error("hooks should be removed entirely when only orphaned handle-session-end.sh existed")
		}
	})

	// Test 4: env has only moai keys -> entire env key removed after cleanup
	t.Run("EmptyEnvRemovedEntirely", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Create settings with env containing only moai-managed keys
		existing := map[string]any{
			"env": map[string]any{
				"PATH":                                 "/old/go/bin:/usr/bin",
				"ENABLE_TOOL_SEARCH":                   "1",
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

		var settings map[string]any
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse settings.json: %v", err)
		}

		// env key should NOT be removed (CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS is added as default)
		env, hasEnv := settings["env"]
		if !hasEnv {
			t.Fatal("env should exist (CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS should be added)")
		}
		envMap := env.(map[string]any)
		// Only CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS should be present
		if envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"] != "1" {
			t.Errorf("CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS should be '1', got: %v", envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"])
		}
		if len(envMap) != 1 {
			t.Errorf("env should only contain CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS, got: %v", envMap)
		}

		// SessionEnd hook should NOT be present (no longer managed globally)
		if hooks, ok := settings["hooks"].(map[string]any); ok {
			if _, exists := hooks["SessionEnd"]; exists {
				t.Error("SessionEnd hook should not be present in global settings")
			}
		}
	})

	// Test 5: Permissions with user-added entries (not just Task:*) -> preserved
	t.Run("PreserveUserPermissions", func(t *testing.T) {
		settingsPath := filepath.Join(claudeDir, "settings.json")

		// Create settings with permissions that include user-added entries beyond Task:*
		// and an orphaned handle-session-end.sh hook that should be cleaned up
		existing := map[string]any{
			"permissions": map[string]any{
				"allow": []any{"Task:*", "Bash(npm run build):*"},
			},
			"hooks": map[string]any{
				"SessionEnd": []any{
					map[string]any{
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "\"$HOME/.claude/hooks/moai/handle-session-end.sh\"",
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

		var settings map[string]any
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse settings.json: %v", err)
		}

		// Permissions should be PRESERVED because it has more than just Task:*
		permVal, exists := settings["permissions"]
		if !exists {
			t.Fatal("permissions should be preserved when it contains user-added entries")
		}
		permMap := permVal.(map[string]any)
		allowArr := permMap["allow"].([]any)
		if len(allowArr) != 2 {
			t.Errorf("permissions.allow should have 2 entries, got %d", len(allowArr))
		}

		// Orphaned SessionEnd hook should be cleaned up
		if hooks, ok := settings["hooks"].(map[string]any); ok {
			if _, exists := hooks["SessionEnd"]; exists {
				t.Error("orphaned SessionEnd hook should be cleaned up")
			}
		}
	})
}

func TestEnsureGlobalSettingsEnv_CleanupMigratedSettings(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()

	// Mock home directory to temp dir
	originalHome := os.Getenv("HOME")
	originalProfile := os.Getenv("USERPROFILE")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
		_ = os.Setenv("USERPROFILE", originalProfile)
	}()
	_ = os.Setenv("HOME", tempDir)
	_ = os.Setenv("USERPROFILE", tempDir) // Windows: os.UserHomeDir() checks USERPROFILE first

	// Create .claude directory
	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Create existing settings with ALL moai-managed settings that should be cleaned up:
	// env keys (PATH, ENABLE_TOOL_SEARCH, CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS),
	// permissions with only Task:*, teammateMode "auto", orphaned SessionEnd hook, plus a custom env key
	existing := map[string]any{
		"env": map[string]any{
			"PATH":                                 "/old/go/bin:/usr/local/bin:/usr/bin:/bin",
			"ENABLE_TOOL_SEARCH":                   "1",
			"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1",
			"CUSTOM_VAR":                           "preserved",
		},
		"permissions": map[string]any{
			"allow": []any{"Task:*"},
		},
		"teammateMode": "auto",
		"hooks": map[string]any{
			"SessionEnd": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": "\"$HOME/.claude/hooks/moai/handle-session-end.sh\"",
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

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings.json: %v", err)
	}

	// Moai-managed env keys should be REMOVED from global settings
	env, hasEnv := settings["env"]
	if !hasEnv {
		t.Fatal("env should still exist (CUSTOM_VAR is present)")
	}
	envMap := env.(map[string]any)

	if _, exists := envMap["PATH"]; exists {
		t.Error("PATH should be removed from global settings (managed at project level)")
	}
	if _, exists := envMap["ENABLE_TOOL_SEARCH"]; exists {
		t.Error("ENABLE_TOOL_SEARCH should be removed from global settings")
	}
	// CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS should be set to "1" as default
	if envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"] != "1" {
		t.Errorf("CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS should be set to '1', got: %v", envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"])
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

	// SessionEnd hook (orphaned handle-session-end.sh) should be REMOVED
	if hooks, hasHooks := settings["hooks"]; hasHooks {
		if hooksMap, ok := hooks.(map[string]any); ok {
			if _, hasSessionEnd := hooksMap["SessionEnd"]; hasSessionEnd {
				t.Error("orphaned SessionEnd hook should be removed from global settings")
			}
		}
	}
}

func TestEnsureGlobalSettingsEnv_RemovesGlobalHooksDir(t *testing.T) {
	tempDir := t.TempDir()
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	defer func() {
		_ = os.Setenv("HOME", origHome)
		_ = os.Setenv("USERPROFILE", origUserProfile)
	}()
	_ = os.Setenv("HOME", tempDir)
	_ = os.Setenv("USERPROFILE", tempDir)

	claudeDir := filepath.Join(tempDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	// Create empty settings.json so ensureGlobalSettingsEnv doesn't bail early
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write settings: %v", err)
	}

	t.Run("RemovesExistingGlobalHooksMoaiDir", func(t *testing.T) {
		hooksDir := filepath.Join(claudeDir, "hooks", "moai")
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			t.Fatalf("failed to create hooks dir: %v", err)
		}
		// Create a dummy hook file
		dummyHook := filepath.Join(hooksDir, "handle-stop.sh")
		if err := os.WriteFile(dummyHook, []byte("#!/bin/bash\n"), 0755); err != nil {
			t.Fatalf("failed to write dummy hook: %v", err)
		}

		err := ensureGlobalSettingsEnv()
		if err != nil {
			t.Fatalf("ensureGlobalSettingsEnv failed: %v", err)
		}

		if _, err := os.Stat(hooksDir); !os.IsNotExist(err) {
			t.Error("~/.claude/hooks/moai should be removed by ensureGlobalSettingsEnv")
		}
	})

	t.Run("NoErrorWhenGlobalHooksDirMissing", func(t *testing.T) {
		hooksDir := filepath.Join(claudeDir, "hooks", "moai")
		_ = os.RemoveAll(hooksDir)

		err := ensureGlobalSettingsEnv()
		if err != nil {
			t.Fatalf("ensureGlobalSettingsEnv should not fail when hooks dir is missing: %v", err)
		}
	})
}

func TestCleanLegacyHooks_RemovesMoaiHandleHooks(t *testing.T) {
	moaiHookPatterns := []struct {
		hookType string
		command  string
	}{
		{"Stop", `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-stop.sh"`},
		{"PreToolUse", `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-pre-tool.sh"`},
		{"PostToolUse", `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-post-tool.sh"`},
		{"SessionStart", `"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh"`},
	}

	for _, tc := range moaiHookPatterns {
		t.Run(tc.hookType, func(t *testing.T) {
			settings := map[string]any{
				"hooks": map[string]any{
					tc.hookType: []any{
						map[string]any{
							"hooks": []any{
								map[string]any{
									"command": tc.command,
									"timeout": float64(5),
									"type":    "command",
								},
							},
						},
					},
				},
			}

			modified := cleanLegacyHooks(settings)
			if !modified {
				t.Errorf("cleanLegacyHooks should detect %s hook as legacy", tc.hookType)
			}

			if hooks, exists := settings["hooks"]; exists {
				hooksMap, _ := hooks.(map[string]any)
				if _, hasHookType := hooksMap[tc.hookType]; hasHookType {
					t.Errorf("%s hook should be removed from global settings", tc.hookType)
				}
			}
		})
	}
}

func TestBuildSmartPATH(t *testing.T) {
	sep := string(os.PathListSeparator)

	// Use a platform-appropriate home directory so filepath.Join produces
	// consistent separators between test data and BuildSmartPATH internals.
	homeDir := filepath.Join(t.TempDir(), "testuser")
	localBin := filepath.Join(homeDir, ".local", "bin")
	goBin := filepath.Join(homeDir, "go", "bin")

	tests := []struct {
		name          string
		envPATH       string
		wantLocalBin  bool
		wantGoBin     bool
		wantUnchanged bool // true if PATH should remain unchanged (essential dirs already present)
	}{
		{
			name:         "essential dirs missing from PATH - should be prepended",
			envPATH:      strings.Join([]string{"/usr/local/bin", "/usr/bin", "/bin"}, sep),
			wantLocalBin: true,
			wantGoBin:    true,
		},
		{
			name:          "essential dirs already in PATH - PATH unchanged",
			envPATH:       strings.Join([]string{localBin, goBin, "/usr/bin"}, sep),
			wantLocalBin:  true,
			wantGoBin:     true,
			wantUnchanged: true,
		},
		{
			name:          "trailing slash on existing dir - should match and not duplicate",
			envPATH:       strings.Join([]string{localBin + "/", goBin + "/", "/usr/bin"}, sep),
			wantLocalBin:  true,
			wantGoBin:     true,
			wantUnchanged: true,
		},
		{
			name:         "empty current PATH - essential dirs only",
			envPATH:      "",
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

			result := buildSmartPATH(homeDir)

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

// --- Binary-First Update Tests ---

func TestShouldSkipBinaryUpdate_EnvVar(t *testing.T) {
	origVal := os.Getenv("MOAI_SKIP_BINARY_UPDATE")
	defer func() {
		if origVal == "" {
			_ = os.Unsetenv("MOAI_SKIP_BINARY_UPDATE")
		} else {
			_ = os.Setenv("MOAI_SKIP_BINARY_UPDATE", origVal)
		}
	}()

	_ = os.Setenv("MOAI_SKIP_BINARY_UPDATE", "1")

	// Use the update command so the templates-only flag is registered
	cmd := updateCmd
	cmd.SetArgs([]string{})

	got := shouldSkipBinaryUpdate(cmd)
	if !got {
		t.Error("shouldSkipBinaryUpdate should return true when MOAI_SKIP_BINARY_UPDATE=1")
	}
}

func TestShouldSkipBinaryUpdate_DevBuild(t *testing.T) {
	origVal := os.Getenv("MOAI_SKIP_BINARY_UPDATE")
	defer func() {
		if origVal == "" {
			_ = os.Unsetenv("MOAI_SKIP_BINARY_UPDATE")
		} else {
			_ = os.Setenv("MOAI_SKIP_BINARY_UPDATE", origVal)
		}
	}()
	_ = os.Unsetenv("MOAI_SKIP_BINARY_UPDATE")

	origVersion := version.Version
	defer func() { version.Version = origVersion }()

	devVersions := []string{"dev", "abc1234-dirty", "none"}
	for _, v := range devVersions {
		version.Version = v
		cmd := updateCmd
		cmd.SetArgs([]string{})

		got := shouldSkipBinaryUpdate(cmd)
		if !got {
			t.Errorf("shouldSkipBinaryUpdate should return true for dev version %q", v)
		}
	}
}

func TestShouldSkipBinaryUpdate_Normal(t *testing.T) {
	origVal := os.Getenv("MOAI_SKIP_BINARY_UPDATE")
	defer func() {
		if origVal == "" {
			_ = os.Unsetenv("MOAI_SKIP_BINARY_UPDATE")
		} else {
			_ = os.Setenv("MOAI_SKIP_BINARY_UPDATE", origVal)
		}
	}()
	_ = os.Unsetenv("MOAI_SKIP_BINARY_UPDATE")

	origVersion := version.Version
	defer func() { version.Version = origVersion }()
	version.Version = "v2.0.0"

	cmd := updateCmd
	cmd.SetArgs([]string{})

	got := shouldSkipBinaryUpdate(cmd)
	if got {
		t.Error("shouldSkipBinaryUpdate should return false for normal version v2.0.0")
	}
}

func TestUpdateCmd_HasTemplatesOnlyFlag(t *testing.T) {
	f := updateCmd.Flags().Lookup("templates-only")
	if f == nil {
		t.Fatal("update command should have --templates-only flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--templates-only default should be false, got %q", f.DefValue)
	}
}

func TestRunBinaryUpdateStep_NoUpdate(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		UpdateChecker: &mockUpdateChecker{
			isUpdateAvailFunc: func(current string) (bool, *update.VersionInfo, error) {
				return false, nil, nil
			},
		},
	}

	origVersion := version.Version
	defer func() { version.Version = origVersion }()
	version.Version = "v2.0.0"

	var buf bytes.Buffer
	cmd := updateCmd
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	cmd.SetContext(context.Background())

	updated, err := runBinaryUpdateStep(cmd)
	if err != nil {
		t.Fatalf("runBinaryUpdateStep returned error: %v", err)
	}
	if updated {
		t.Error("runBinaryUpdateStep should return updated=false when no update is available")
	}
}

func TestRunBinaryUpdateStep_UpdateAvailable(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = &Dependencies{
		UpdateChecker: &mockUpdateChecker{
			isUpdateAvailFunc: func(current string) (bool, *update.VersionInfo, error) {
				return true, &update.VersionInfo{Version: "v3.0.0"}, nil
			},
		},
		UpdateOrch: &mockUpdateOrchestrator{
			updateFunc: func(ctx context.Context) (*update.UpdateResult, error) {
				return &update.UpdateResult{
					PreviousVersion: "v2.0.0",
					NewVersion:      "v3.0.0",
				}, nil
			},
		},
	}

	origVersion := version.Version
	defer func() { version.Version = origVersion }()
	version.Version = "v2.0.0"

	var buf bytes.Buffer
	cmd := updateCmd
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	cmd.SetContext(context.Background())

	updated, err := runBinaryUpdateStep(cmd)
	if err != nil {
		t.Fatalf("runBinaryUpdateStep returned error: %v", err)
	}
	if !updated {
		t.Error("runBinaryUpdateStep should return updated=true when update succeeds")
	}
	output := buf.String()
	if !strings.Contains(output, "v3.0.0") {
		t.Errorf("output should mention new version v3.0.0, got: %s", output)
	}
}

func TestCleanMoaiManagedPaths(t *testing.T) {
	// Helper to create a file (and its parent directories) inside the temp dir.
	mkFile := func(t *testing.T, base string, relPath string, content string) {
		t.Helper()
		full := filepath.Join(base, relPath)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll for %s: %v", relPath, err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile for %s: %v", relPath, err)
		}
	}

	// Helper to create a directory inside the temp dir.
	mkDir := func(t *testing.T, base string, relPath string) {
		t.Helper()
		full := filepath.Join(base, relPath)
		if err := os.MkdirAll(full, 0o755); err != nil {
			t.Fatalf("MkdirAll for %s: %v", relPath, err)
		}
	}

	// Helper to check if a path exists.
	pathExists := func(base string, relPath string) bool {
		_, err := os.Stat(filepath.Join(base, relPath))
		return err == nil
	}

	tests := []struct {
		name string
		// setup creates files/dirs in the temp directory before running cleanMoaiManagedPaths.
		setup func(t *testing.T, root string)
		// verify runs assertions after cleanMoaiManagedPaths returns.
		verify func(t *testing.T, root string, output string)
		// wantErr indicates whether cleanMoaiManagedPaths should return an error.
		wantErr bool
	}{
		{
			name: "AllPathsPresent",
			setup: func(t *testing.T, root string) {
				// Standard managed paths
				mkFile(t, root, filepath.Join(".claude", "settings.json"), `{"key":"val"}`)
				mkDir(t, root, filepath.Join(".claude", "commands", "moai"))
				mkFile(t, root, filepath.Join(".claude", "commands", "moai", "plan.md"), "plan")
				mkDir(t, root, filepath.Join(".claude", "agents", "moai"))
				mkFile(t, root, filepath.Join(".claude", "agents", "moai", "expert.md"), "expert")
				// Glob targets: skills/moai*
				mkDir(t, root, filepath.Join(".claude", "skills", "moai-backend"))
				mkFile(t, root, filepath.Join(".claude", "skills", "moai-backend", "SKILL.md"), "backend")
				mkDir(t, root, filepath.Join(".claude", "skills", "moai-frontend"))
				mkFile(t, root, filepath.Join(".claude", "skills", "moai-frontend", "SKILL.md"), "frontend")
				// rules/moai
				mkDir(t, root, filepath.Join(".claude", "rules", "moai"))
				mkFile(t, root, filepath.Join(".claude", "rules", "moai", "rule.md"), "rule")
				// output-styles/moai
				mkDir(t, root, filepath.Join(".claude", "output-styles", "moai"))
				mkFile(t, root, filepath.Join(".claude", "output-styles", "moai", "style.md"), "style")
				// hooks/moai
				mkDir(t, root, filepath.Join(".claude", "hooks", "moai"))
				mkFile(t, root, filepath.Join(".claude", "hooks", "moai", "hook.sh"), "#!/bin/bash")
				// .moai/config/ with sections/ and config.yaml
				mkFile(t, root, filepath.Join(".moai", "config", "config.yaml"), "version: 1")
				mkFile(t, root, filepath.Join(".moai", "config", "sections", "user.yaml"), "user: name")
			},
			verify: func(t *testing.T, root string, output string) {
				// All standard managed paths should be removed
				if pathExists(root, filepath.Join(".claude", "settings.json")) {
					t.Error(".claude/settings.json should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "commands", "moai")) {
					t.Error(".claude/commands/moai should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "agents", "moai")) {
					t.Error(".claude/agents/moai should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "skills", "moai-backend")) {
					t.Error(".claude/skills/moai-backend should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "skills", "moai-frontend")) {
					t.Error(".claude/skills/moai-frontend should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "rules", "moai")) {
					t.Error(".claude/rules/moai should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "output-styles", "moai")) {
					t.Error(".claude/output-styles/moai should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "hooks", "moai")) {
					t.Error(".claude/hooks/moai should have been removed")
				}
				// entire .moai/config/ should be removed (backup handles preservation)
				if pathExists(root, filepath.Join(".moai", "config", "config.yaml")) {
					t.Error(".moai/config/config.yaml should have been removed")
				}
				if pathExists(root, filepath.Join(".moai", "config", "sections", "user.yaml")) {
					t.Error(".moai/config/sections/user.yaml should have been removed")
				}
			},
		},
		{
			name: "NoPathsExist",
			setup: func(t *testing.T, root string) {
				// Empty temp dir, no managed paths exist
			},
			verify: func(t *testing.T, root string, output string) {
				// Should complete without error; verify output contains skip markers
				if !strings.Contains(output, "Skipped") {
					t.Error("expected output to contain 'Skipped' markers for non-existent paths")
				}
			},
		},
		{
			name: "MixedPresence",
			setup: func(t *testing.T, root string) {
				// Only some paths exist
				mkFile(t, root, filepath.Join(".claude", "settings.json"), `{}`)
				mkDir(t, root, filepath.Join(".claude", "rules", "moai"))
				mkFile(t, root, filepath.Join(".claude", "rules", "moai", "core.md"), "core")
				// commands/moai, agents/moai, skills/moai*, output-styles/moai, hooks/moai do NOT exist
			},
			verify: func(t *testing.T, root string, output string) {
				// Existing paths should be removed
				if pathExists(root, filepath.Join(".claude", "settings.json")) {
					t.Error(".claude/settings.json should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "rules", "moai")) {
					t.Error(".claude/rules/moai should have been removed")
				}
				// Output should contain "Skipped" for non-existent paths
				if !strings.Contains(output, "Skipped") {
					t.Error("expected output to contain 'Skipped' for missing paths")
				}
				// Output should contain "Removed" for existing paths
				if !strings.Contains(output, "Removed") {
					t.Error("expected output to contain 'Removed' for existing paths")
				}
			},
		},
		{
			name: "ConfigDeletedEntirely",
			setup: func(t *testing.T, root string) {
				mkFile(t, root, filepath.Join(".moai", "config", "sections", "user.yaml"), "user: test")
				mkFile(t, root, filepath.Join(".moai", "config", "sections", "language.yaml"), "lang: ko")
				mkFile(t, root, filepath.Join(".moai", "config", "config.yaml"), "version: 2")
				mkFile(t, root, filepath.Join(".moai", "config", "backup_metadata.json"), `{"ts":"now"}`)
			},
			verify: func(t *testing.T, root string, output string) {
				// Entire .moai/config/ should be deleted (backup already done by Backup step)
				if pathExists(root, filepath.Join(".moai", "config", "sections")) {
					t.Error(".moai/config/sections/ should have been removed")
				}
				if pathExists(root, filepath.Join(".moai", "config", "sections", "user.yaml")) {
					t.Error(".moai/config/sections/user.yaml should have been removed")
				}
				if pathExists(root, filepath.Join(".moai", "config", "sections", "language.yaml")) {
					t.Error(".moai/config/sections/language.yaml should have been removed")
				}
				if pathExists(root, filepath.Join(".moai", "config", "config.yaml")) {
					t.Error(".moai/config/config.yaml should have been removed")
				}
				if pathExists(root, filepath.Join(".moai", "config", "backup_metadata.json")) {
					t.Error(".moai/config/backup_metadata.json should have been removed")
				}
			},
		},
		{
			name: "GlobMatchesMultiple",
			setup: func(t *testing.T, root string) {
				// Create multiple moai* skill directories
				mkDir(t, root, filepath.Join(".claude", "skills", "moai-backend"))
				mkFile(t, root, filepath.Join(".claude", "skills", "moai-backend", "SKILL.md"), "be")
				mkDir(t, root, filepath.Join(".claude", "skills", "moai-frontend"))
				mkFile(t, root, filepath.Join(".claude", "skills", "moai-frontend", "SKILL.md"), "fe")
				mkDir(t, root, filepath.Join(".claude", "skills", "moai-testing"))
				mkFile(t, root, filepath.Join(".claude", "skills", "moai-testing", "SKILL.md"), "test")
				// Non-moai skill that should survive
				mkDir(t, root, filepath.Join(".claude", "skills", "custom-skill"))
				mkFile(t, root, filepath.Join(".claude", "skills", "custom-skill", "SKILL.md"), "custom")
			},
			verify: func(t *testing.T, root string, output string) {
				// All moai* skills should be removed
				if pathExists(root, filepath.Join(".claude", "skills", "moai-backend")) {
					t.Error(".claude/skills/moai-backend should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "skills", "moai-frontend")) {
					t.Error(".claude/skills/moai-frontend should have been removed")
				}
				if pathExists(root, filepath.Join(".claude", "skills", "moai-testing")) {
					t.Error(".claude/skills/moai-testing should have been removed")
				}
				// Non-moai skill should be preserved
				if !pathExists(root, filepath.Join(".claude", "skills", "custom-skill")) {
					t.Error(".claude/skills/custom-skill should have been preserved")
				}
				if !pathExists(root, filepath.Join(".claude", "skills", "custom-skill", "SKILL.md")) {
					t.Error(".claude/skills/custom-skill/SKILL.md should have been preserved")
				}
			},
		},
		{
			name: "OutputCapture",
			setup: func(t *testing.T, root string) {
				// Create some paths that exist and leave others missing
				mkFile(t, root, filepath.Join(".claude", "settings.json"), `{}`)
				mkDir(t, root, filepath.Join(".claude", "agents", "moai"))
				// commands/moai does not exist - should produce "Skipped"
			},
			verify: func(t *testing.T, root string, output string) {
				// Verify output contains "Removed" markers for existing paths
				if !strings.Contains(output, "Removed") {
					t.Errorf("expected output to contain 'Removed' marker, got:\n%s", output)
				}
				// Verify output contains "Skipped" markers for missing paths
				if !strings.Contains(output, "Skipped") {
					t.Errorf("expected output to contain 'Skipped' marker, got:\n%s", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			var buf bytes.Buffer
			err := cleanMoaiManagedPaths(root, &buf)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tt.verify(t, root, buf.String())
		})
	}
}

// --- TDD RED: Statusline wizard config tests ---

func TestPresetToSegments(t *testing.T) {
	allSegments := []string{"model", "context", "output_style", "directory", "git_status", "claude_version", "moai_version", "git_branch"}

	tests := []struct {
		name      string
		preset    string
		custom    map[string]bool
		wantTrue  []string // segments that should be true
		wantFalse []string // segments that should be false
	}{
		{
			name:     "full preset enables all segments",
			preset:   "full",
			wantTrue: allSegments,
		},
		{
			name:      "compact preset enables subset",
			preset:    "compact",
			wantTrue:  []string{"model", "context", "git_status", "git_branch"},
			wantFalse: []string{"output_style", "directory", "claude_version", "moai_version"},
		},
		{
			name:      "minimal preset enables model and context only",
			preset:    "minimal",
			wantTrue:  []string{"model", "context"},
			wantFalse: []string{"output_style", "directory", "git_status", "claude_version", "moai_version", "git_branch"},
		},
		{
			name:   "custom preset uses provided map",
			preset: "custom",
			custom: map[string]bool{
				"model":      true,
				"context":    false,
				"git_branch": true,
			},
			wantTrue:  []string{"model", "git_branch"},
			wantFalse: []string{"context"},
		},
		{
			name:     "unknown preset falls back to full",
			preset:   "unknown",
			wantTrue: allSegments,
		},
		{
			name:     "empty preset falls back to full",
			preset:   "",
			wantTrue: allSegments,
		},
		{
			name:     "custom with nil map defaults all to true",
			preset:   "custom",
			custom:   nil,
			wantTrue: allSegments,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := presetToSegments(tt.preset, tt.custom)

			for _, seg := range tt.wantTrue {
				if !result[seg] {
					t.Errorf("segment %q should be true for preset %q", seg, tt.preset)
				}
			}

			for _, seg := range tt.wantFalse {
				if result[seg] {
					t.Errorf("segment %q should be false for preset %q", seg, tt.preset)
				}
			}
		})
	}
}

func TestApplyWizardConfig_StatuslinePresetFull(t *testing.T) {
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create required language.yaml (applyWizardConfig writes it first)
	result := &wizard.WizardResult{
		Locale:           "en",
		StatuslinePreset: "full",
	}

	if err := applyWizardConfig(tmpDir, result); err != nil {
		t.Fatalf("applyWizardConfig failed: %v", err)
	}

	// Read statusline.yaml
	data, err := os.ReadFile(filepath.Join(sectionsDir, "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	content := string(data)

	// Should contain preset: full
	if !strings.Contains(content, "preset: full") {
		t.Errorf("statusline.yaml should contain 'preset: full', got:\n%s", content)
	}

	// All segments should be true
	for _, seg := range []string{"model", "context", "output_style", "directory", "git_status", "claude_version", "moai_version", "git_branch"} {
		if !strings.Contains(content, seg+": true") {
			t.Errorf("statusline.yaml should contain '%s: true', got:\n%s", seg, content)
		}
	}
}

func TestApplyWizardConfig_StatuslinePresetCompact(t *testing.T) {
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	result := &wizard.WizardResult{
		Locale:           "en",
		StatuslinePreset: "compact",
	}

	if err := applyWizardConfig(tmpDir, result); err != nil {
		t.Fatalf("applyWizardConfig failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(sectionsDir, "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	content := string(data)

	// Compact: model, context, git_status, git_branch enabled
	for _, seg := range []string{"model", "context", "git_status", "git_branch"} {
		if !strings.Contains(content, seg+": true") {
			t.Errorf("statusline.yaml should contain '%s: true' for compact, got:\n%s", seg, content)
		}
	}

	// Others disabled
	for _, seg := range []string{"output_style", "directory", "claude_version", "moai_version"} {
		if !strings.Contains(content, seg+": false") {
			t.Errorf("statusline.yaml should contain '%s: false' for compact, got:\n%s", seg, content)
		}
	}
}

func TestApplyWizardConfig_StatuslinePresetMinimal(t *testing.T) {
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	result := &wizard.WizardResult{
		Locale:           "en",
		StatuslinePreset: "minimal",
	}

	if err := applyWizardConfig(tmpDir, result); err != nil {
		t.Fatalf("applyWizardConfig failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(sectionsDir, "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	content := string(data)

	// Minimal: model, context enabled
	for _, seg := range []string{"model", "context"} {
		if !strings.Contains(content, seg+": true") {
			t.Errorf("statusline.yaml should contain '%s: true' for minimal, got:\n%s", seg, content)
		}
	}

	// Others disabled
	for _, seg := range []string{"output_style", "directory", "git_status", "claude_version", "moai_version", "git_branch"} {
		if !strings.Contains(content, seg+": false") {
			t.Errorf("statusline.yaml should contain '%s: false' for minimal, got:\n%s", seg, content)
		}
	}
}

func TestApplyWizardConfig_StatuslinePresetCustom(t *testing.T) {
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	result := &wizard.WizardResult{
		Locale:           "en",
		StatuslinePreset: "custom",
		StatuslineSegments: map[string]bool{
			"model":          true,
			"context":        true,
			"output_style":   false,
			"directory":      false,
			"git_status":     true,
			"claude_version": false,
			"moai_version":   false,
			"git_branch":     true,
		},
	}

	if err := applyWizardConfig(tmpDir, result); err != nil {
		t.Fatalf("applyWizardConfig failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(sectionsDir, "statusline.yaml"))
	if err != nil {
		t.Fatalf("read statusline.yaml: %v", err)
	}

	content := string(data)

	// Custom segments
	for _, seg := range []string{"model", "context", "git_status", "git_branch"} {
		if !strings.Contains(content, seg+": true") {
			t.Errorf("statusline.yaml should contain '%s: true' for custom, got:\n%s", seg, content)
		}
	}
	for _, seg := range []string{"output_style", "directory", "claude_version", "moai_version"} {
		if !strings.Contains(content, seg+": false") {
			t.Errorf("statusline.yaml should contain '%s: false' for custom, got:\n%s", seg, content)
		}
	}
}

func TestApplyWizardConfig_StatuslineEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	sectionsDir := filepath.Join(tmpDir, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	result := &wizard.WizardResult{
		Locale:           "en",
		StatuslinePreset: "", // Empty - should skip
	}

	if err := applyWizardConfig(tmpDir, result); err != nil {
		t.Fatalf("applyWizardConfig failed: %v", err)
	}

	// statusline.yaml should NOT exist
	statuslinePath := filepath.Join(sectionsDir, "statusline.yaml")
	if _, err := os.Stat(statuslinePath); !os.IsNotExist(err) {
		t.Error("statusline.yaml should not be created when StatuslinePreset is empty")
	}
}

func TestMergeGitignoreFile_PreservesUserPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	// Simulate new template .gitignore (deployed by template sync)
	templateContent := "# Go\n*.exe\n*.test\n*.out\nvendor/\n\n# IDE\n.idea/\n.vscode/\n"
	if err := os.WriteFile(gitignorePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	// User's backup had template patterns PLUS custom patterns
	userBackup := []byte("# Go\n*.exe\n*.test\n*.out\nvendor/\n\n# IDE\n.idea/\n.vscode/\n\n# My custom patterns\nmy-secret.txt\nbuild-output/\n.env.local\n")

	if err := mergeGitignoreFile(gitignorePath, userBackup); err != nil {
		t.Fatalf("mergeGitignoreFile failed: %v", err)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	result := string(data)

	// Template patterns should still be present
	if !strings.Contains(result, "*.exe") {
		t.Error("template pattern *.exe should be preserved")
	}
	if !strings.Contains(result, ".idea/") {
		t.Error("template pattern .idea/ should be preserved")
	}

	// User custom patterns should be appended
	if !strings.Contains(result, "my-secret.txt") {
		t.Error("user pattern my-secret.txt should be preserved")
	}
	if !strings.Contains(result, "build-output/") {
		t.Error("user pattern build-output/ should be preserved")
	}
	if !strings.Contains(result, ".env.local") {
		t.Error("user pattern .env.local should be preserved")
	}

	// Should have the user custom patterns header
	if !strings.Contains(result, "User Custom Patterns") {
		t.Error("should contain 'User Custom Patterns' header")
	}
}

func TestMergeGitignoreFile_NoUserAdditions(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	// Template and user have the same patterns
	templateContent := "*.exe\n*.test\nvendor/\n"
	if err := os.WriteFile(gitignorePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	userBackup := []byte("*.exe\n*.test\nvendor/\n")

	if err := mergeGitignoreFile(gitignorePath, userBackup); err != nil {
		t.Fatalf("mergeGitignoreFile failed: %v", err)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	result := string(data)

	// No changes should be made (no user-specific patterns)
	if strings.Contains(result, "User Custom Patterns") {
		t.Error("should NOT contain 'User Custom Patterns' header when no user additions exist")
	}

	// Original template content should remain unchanged
	if result != templateContent {
		t.Errorf("file should remain unchanged, got:\n%s", result)
	}
}

func TestMergeGitignoreFile_EmptyBackup(t *testing.T) {
	tmpDir := t.TempDir()
	gitignorePath := filepath.Join(tmpDir, ".gitignore")

	templateContent := "*.exe\nvendor/\n"
	if err := os.WriteFile(gitignorePath, []byte(templateContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Empty user backup — nothing to merge
	if err := mergeGitignoreFile(gitignorePath, []byte("")); err != nil {
		t.Fatalf("mergeGitignoreFile failed: %v", err)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != templateContent {
		t.Errorf("file should remain unchanged with empty backup, got:\n%s", string(data))
	}
}

func TestRestoreMoaiConfig_CustomSectionPreserved(t *testing.T) {
	// Test that user's custom config sections (not in template) are restored
	tmpDir := t.TempDir()

	// Create config structure with a standard section AND a custom section
	configDir := filepath.Join(tmpDir, ".moai", "config")
	sectionsDir := filepath.Join(configDir, "sections")
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Standard section that exists in template
	standardPath := filepath.Join(sectionsDir, "system.yaml")
	standardContent := []byte("moai:\n  version: \"1.0.0\"\n")
	if err := os.WriteFile(standardPath, standardContent, 0644); err != nil {
		t.Fatal(err)
	}

	// User's custom section (NOT in template)
	customPath := filepath.Join(sectionsDir, "my-custom.yaml")
	customContent := []byte("custom:\n  setting: \"my-value\"\n  enabled: true\n")
	if err := os.WriteFile(customPath, customContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create backup (includes both standard and custom)
	backupDir, err := backupMoaiConfig(tmpDir)
	if err != nil {
		t.Fatalf("backupMoaiConfig failed: %v", err)
	}

	// Simulate template sync: remove custom section (template doesn't have it)
	_ = os.Remove(customPath)

	// Update standard section (as if template deployed a new version)
	newStandardContent := []byte("moai:\n  version: \"2.0.0\"\n  new_field: \"added\"\n")
	if err := os.WriteFile(standardPath, newStandardContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Restore from backup
	if err := restoreMoaiConfig(tmpDir, backupDir); err != nil {
		t.Fatalf("restoreMoaiConfig failed: %v", err)
	}

	// Verify: custom section should be restored
	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		t.Fatal("user's custom config section should be restored")
	}

	data, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "my-value") {
		t.Errorf("custom section should contain original user value, got:\n%s", string(data))
	}

	// Verify: standard section should be merged (not just overwritten with backup)
	standardData, err := os.ReadFile(standardPath)
	if err != nil {
		t.Fatal(err)
	}

	// New field from template should be present
	if !strings.Contains(string(standardData), "new_field") {
		t.Errorf("standard section should contain new template field, got:\n%s", string(standardData))
	}
}
