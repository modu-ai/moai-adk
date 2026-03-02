package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// setupSectionsDir creates the .moai/config/sections/ directory tree in a temp dir
// and returns the temp dir root path.
func setupSectionsDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, defs.DirPerm); err != nil {
		t.Fatalf("create sections dir: %v", err)
	}
	return root
}

// readYAML reads a YAML file and unmarshals it into a map.
func readYAML(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return result
}

// --- applyWizardConfig tests ---

func TestApplyWizardConfig_GitHubUsername(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitHubUsername:"testuser",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	parsed := readYAML(t, userPath)

	user, ok := parsed["user"].(map[string]any)
	if !ok {
		t.Fatal("expected user key in parsed YAML")
	}

	if user["github_username"] != "testuser" {
		t.Errorf("user.github_username = %v, want %q", user["github_username"], "testuser")
	}
}

func TestApplyWizardConfig_GitHubToken(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitHubToken: "ghp_testtoken123",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	parsed := readYAML(t, userPath)

	user := parsed["user"].(map[string]any)
	if user["github_token"] != "ghp_testtoken123" {
		t.Errorf("user.github_token = %v, want %q", user["github_token"], "ghp_testtoken123")
	}
}

func TestApplyWizardConfig_GitHubUsernameAndToken(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitHubUsername:"myuser",
		GitHubToken:    "ghp_tok",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	parsed := readYAML(t, userPath)

	user := parsed["user"].(map[string]any)
	if user["github_username"] != "myuser" {
		t.Errorf("user.github_username = %v, want %q", user["github_username"], "myuser")
	}
	if user["github_token"] != "ghp_tok" {
		t.Errorf("user.github_token = %v, want %q", user["github_token"], "ghp_tok")
	}
}

func TestApplyWizardConfig_NoUserYAMLWhenGitHubFieldsEmpty(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitHubUsername:"",
		GitHubToken:    "",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	if _, err := os.Stat(userPath); !os.IsNotExist(err) {
		t.Error("user.yaml should not be created when both GitHubUsername and GitHubToken are empty")
	}
}

func TestApplyWizardConfig_ExistingUserYAMLPreserved(t *testing.T) {
	root := setupSectionsDir(t)

	// Pre-create user.yaml with existing content.
	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	existingContent := "user:\n  name: existing-user\n"
	if err := os.WriteFile(userPath, []byte(existingContent), defs.FilePerm); err != nil {
		t.Fatalf("write existing user.yaml: %v", err)
	}

	result := &wizard.WizardResult{
		GitHubUsername: "newuser",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	parsed := readYAML(t, userPath)
	user := parsed["user"].(map[string]any)

	// New field should be added.
	if user["github_username"] != "newuser" {
		t.Errorf("github_username = %v, want %q", user["github_username"], "newuser")
	}

	// Existing field should be preserved.
	if user["name"] != "existing-user" {
		t.Errorf("user.name = %v, want %q", user["name"], "existing-user")
	}
}

func TestApplyWizardConfig_AllFieldsPopulated(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitHubUsername: "fulluser",
		GitHubToken:   "ghp_full",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)

	// Verify user.yaml
	userParsed := readYAML(t, filepath.Join(sectionsDir, defs.UserYAML))
	user := userParsed["user"].(map[string]any)
	if user["github_username"] != "fulluser" {
		t.Errorf("github_username = %v, want fulluser", user["github_username"])
	}
	if user["github_token"] != "ghp_full" {
		t.Errorf("github_token = %v, want ghp_full", user["github_token"])
	}
}

// --- presetToSegments tests ---

func TestPresetToSegments_Full(t *testing.T) {
	segments := presetToSegments("full", nil)
	for _, seg := range allStatuslineSegments {
		if !segments[seg] {
			t.Errorf("segment %q should be true for full preset", seg)
		}
	}
}

func TestPresetToSegments_Unknown(t *testing.T) {
	segments := presetToSegments("unknown-preset", nil)
	for _, seg := range allStatuslineSegments {
		if !segments[seg] {
			t.Errorf("segment %q should be true for unknown preset (falls back to full)", seg)
		}
	}
}

func TestPresetToSegments_Compact(t *testing.T) {
	segments := presetToSegments("compact", nil)
	if !segments["model"] {
		t.Error("compact preset should enable model segment")
	}
	if !segments["context"] {
		t.Error("compact preset should enable context segment")
	}
	if !segments["git_status"] {
		t.Error("compact preset should enable git_status segment")
	}
	if !segments["git_branch"] {
		t.Error("compact preset should enable git_branch segment")
	}
	if segments["output_style"] {
		t.Error("compact preset should disable output_style segment")
	}
	if segments["directory"] {
		t.Error("compact preset should disable directory segment")
	}
}

func TestPresetToSegments_Minimal(t *testing.T) {
	segments := presetToSegments("minimal", nil)
	if !segments["model"] {
		t.Error("minimal preset should enable model segment")
	}
	if !segments["context"] {
		t.Error("minimal preset should enable context segment")
	}
	if segments["git_status"] {
		t.Error("minimal preset should disable git_status segment")
	}
	if segments["directory"] {
		t.Error("minimal preset should disable directory segment")
	}
}

func TestPresetToSegments_CustomWithNilMap(t *testing.T) {
	segments := presetToSegments("custom", nil)
	// When custom map is nil, all segments default to true.
	for _, seg := range allStatuslineSegments {
		if !segments[seg] {
			t.Errorf("segment %q should be true for custom preset with nil map", seg)
		}
	}
}

func TestPresetToSegments_CustomWithPartialMap(t *testing.T) {
	custom := map[string]bool{
		"model":   false,
		"context": true,
	}
	segments := presetToSegments("custom", custom)

	if segments["model"] {
		t.Error("model should be false per custom map")
	}
	if !segments["context"] {
		t.Error("context should be true per custom map")
	}
	// Segments not in custom map should default to true.
	if !segments["directory"] {
		t.Error("directory should default to true when missing from custom map")
	}
}
