package cli

import (
	"os"
	"path/filepath"
	"strings"
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

func TestApplyWizardConfig_LanguageOnly(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale: "ko",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	langPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.LanguageYAML)
	data, err := os.ReadFile(langPath)
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "conversation_language: ko") {
		t.Errorf("expected conversation_language: ko in language.yaml, got:\n%s", content)
	}
	if !strings.Contains(content, "conversation_language_name: ko") {
		t.Errorf("expected conversation_language_name: ko in language.yaml, got:\n%s", content)
	}
}

func TestApplyWizardConfig_LanguageEnglish(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale: "en",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	langPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.LanguageYAML)
	data, err := os.ReadFile(langPath)
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "conversation_language: en") {
		t.Errorf("expected conversation_language: en in language.yaml, got:\n%s", content)
	}
}

func TestApplyWizardConfig_NoWorkflowYAMLWhenAgentTeamsModeEmpty(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale:         "en",
		AgentTeamsMode: "", // empty means skip
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	workflowPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML)
	if _, err := os.Stat(workflowPath); !os.IsNotExist(err) {
		t.Error("workflow.yaml should not be created when AgentTeamsMode is empty")
	}
}

func TestApplyWizardConfig_WorkflowTeamEnabled(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale:         "en",
		AgentTeamsMode: "team",
		MaxTeammates:   "5",
		DefaultModel:   "sonnet",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	workflowPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML)
	parsed := readYAML(t, workflowPath)

	wf, ok := parsed["workflow"].(map[string]any)
	if !ok {
		t.Fatal("expected workflow key in parsed YAML")
	}

	if wf["execution_mode"] != "team" {
		t.Errorf("execution_mode = %v, want %q", wf["execution_mode"], "team")
	}

	team, ok := wf["team"].(map[string]any)
	if !ok {
		t.Fatal("expected workflow.team key in parsed YAML")
	}

	if team["enabled"] != true {
		t.Errorf("team.enabled = %v, want true", team["enabled"])
	}

	// MaxTeammates "5" should be stored as int 5.
	if team["max_teammates"] != 5 {
		t.Errorf("team.max_teammates = %v, want 5", team["max_teammates"])
	}

	if team["default_model"] != "sonnet" {
		t.Errorf("team.default_model = %v, want %q", team["default_model"], "sonnet")
	}
}

func TestApplyWizardConfig_WorkflowSubagentMode(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale:         "en",
		AgentTeamsMode: "subagent",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	workflowPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML)
	parsed := readYAML(t, workflowPath)

	wf := parsed["workflow"].(map[string]any)
	if wf["execution_mode"] != "subagent" {
		t.Errorf("execution_mode = %v, want %q", wf["execution_mode"], "subagent")
	}

	team := wf["team"].(map[string]any)
	if team["enabled"] != false {
		t.Errorf("team.enabled = %v, want false (subagent mode)", team["enabled"])
	}
}

func TestApplyWizardConfig_MaxTeammatesInvalidValues(t *testing.T) {
	tests := []struct {
		name         string
		maxTeammates string
		expectKey    bool
	}{
		{"below_min", "1", false},
		{"above_max", "11", false},
		{"non_numeric", "abc", false},
		{"empty", "", false},
		{"min_valid", "2", true},
		{"max_valid", "10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := setupSectionsDir(t)
			result := &wizard.WizardResult{
				Locale:         "en",
				AgentTeamsMode: "team",
				MaxTeammates:   tt.maxTeammates,
			}

			if err := applyWizardConfig(root, result); err != nil {
				t.Fatalf("applyWizardConfig: %v", err)
			}

			workflowPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML)
			parsed := readYAML(t, workflowPath)
			wf := parsed["workflow"].(map[string]any)
			team := wf["team"].(map[string]any)

			_, hasKey := team["max_teammates"]
			if hasKey != tt.expectKey {
				t.Errorf("max_teammates key present = %v, want %v (maxTeammates=%q)",
					hasKey, tt.expectKey, tt.maxTeammates)
			}
		})
	}
}

func TestApplyWizardConfig_DefaultModelValidation(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		expectKey bool
	}{
		{"haiku", "haiku", true},
		{"sonnet", "sonnet", true},
		{"opus", "opus", true},
		{"invalid", "gpt4", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := setupSectionsDir(t)
			result := &wizard.WizardResult{
				Locale:         "en",
				AgentTeamsMode: "team",
				DefaultModel:   tt.model,
			}

			if err := applyWizardConfig(root, result); err != nil {
				t.Fatalf("applyWizardConfig: %v", err)
			}

			workflowPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML)
			parsed := readYAML(t, workflowPath)
			wf := parsed["workflow"].(map[string]any)
			team := wf["team"].(map[string]any)

			_, hasKey := team["default_model"]
			if hasKey != tt.expectKey {
				t.Errorf("default_model key present = %v, want %v (model=%q)",
					hasKey, tt.expectKey, tt.model)
			}
		})
	}
}

func TestApplyWizardConfig_NoStatuslineWhenPresetEmpty(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale:           "en",
		StatuslinePreset: "",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	statuslinePath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.StatuslineYAML)
	if _, err := os.Stat(statuslinePath); !os.IsNotExist(err) {
		t.Error("statusline.yaml should not be created when StatuslinePreset is empty")
	}
}

func TestApplyWizardConfig_StatuslineFullPresetAllEnabled(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale:           "en",
		StatuslinePreset: "full",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	statuslinePath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.StatuslineYAML)
	parsed := readYAML(t, statuslinePath)

	sl, ok := parsed["statusline"].(map[string]any)
	if !ok {
		t.Fatal("expected statusline key in parsed YAML")
	}

	if sl["preset"] != "full" {
		t.Errorf("statusline.preset = %v, want %q", sl["preset"], "full")
	}

	segments, ok := sl["segments"].(map[string]any)
	if !ok {
		t.Fatal("expected statusline.segments map")
	}

	// "full" preset enables all segments.
	for _, seg := range allStatuslineSegments {
		val, exists := segments[seg]
		if !exists {
			t.Errorf("segment %q missing from statusline.segments", seg)
			continue
		}
		if val != true {
			t.Errorf("segment %q = %v, want true for full preset", seg, val)
		}
	}
}

func TestApplyWizardConfig_StatuslineCompactPresetValue(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale:           "en",
		StatuslinePreset: "compact",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	statuslinePath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.StatuslineYAML)
	parsed := readYAML(t, statuslinePath)

	sl := parsed["statusline"].(map[string]any)
	if sl["preset"] != "compact" {
		t.Errorf("statusline.preset = %v, want %q", sl["preset"], "compact")
	}
}

func TestApplyWizardConfig_StatuslineMinimalPresetValue(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale:           "en",
		StatuslinePreset: "minimal",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	statuslinePath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.StatuslineYAML)
	parsed := readYAML(t, statuslinePath)

	sl := parsed["statusline"].(map[string]any)
	if sl["preset"] != "minimal" {
		t.Errorf("statusline.preset = %v, want %q", sl["preset"], "minimal")
	}
}

func TestApplyWizardConfig_GitHubUsername(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale:         "en",
		GitHubUsername: "testuser",
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
		Locale:      "en",
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
		Locale:         "en",
		GitHubUsername: "myuser",
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
		Locale:         "en",
		GitHubUsername: "",
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

func TestApplyWizardConfig_ExistingWorkflowYAMLPreserved(t *testing.T) {
	root := setupSectionsDir(t)

	// Pre-create workflow.yaml with existing content.
	workflowPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML)
	existingContent := "workflow:\n  some_existing_key: preserved_value\n"
	if err := os.WriteFile(workflowPath, []byte(existingContent), defs.FilePerm); err != nil {
		t.Fatalf("write existing workflow.yaml: %v", err)
	}

	result := &wizard.WizardResult{
		Locale:         "en",
		AgentTeamsMode: "auto",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	parsed := readYAML(t, workflowPath)
	wf := parsed["workflow"].(map[string]any)

	// The new execution_mode should be set.
	if wf["execution_mode"] != "auto" {
		t.Errorf("execution_mode = %v, want %q", wf["execution_mode"], "auto")
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
		Locale:         "en",
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

func TestApplyWizardConfig_LanguageWriteError(t *testing.T) {
	// Use a non-existent path to trigger a write error for language.yaml.
	result := &wizard.WizardResult{
		Locale: "en",
	}

	err := applyWizardConfig("/nonexistent/path/root", result)
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
	if !strings.Contains(err.Error(), "write language.yaml") {
		t.Errorf("expected 'write language.yaml' in error, got: %v", err)
	}
}

func TestApplyWizardConfig_AllFieldsPopulated(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		Locale:           "ja",
		AgentTeamsMode:   "team",
		MaxTeammates:     "8",
		DefaultModel:     "opus",
		StatuslinePreset: "compact",
		GitHubUsername:   "fulluser",
		GitHubToken:      "ghp_full",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)

	// Verify language.yaml
	langData, err := os.ReadFile(filepath.Join(sectionsDir, defs.LanguageYAML))
	if err != nil {
		t.Fatalf("read language.yaml: %v", err)
	}
	if !strings.Contains(string(langData), "conversation_language: ja") {
		t.Errorf("language.yaml missing ja locale")
	}

	// Verify workflow.yaml
	wfParsed := readYAML(t, filepath.Join(sectionsDir, defs.WorkflowYAML))
	wf := wfParsed["workflow"].(map[string]any)
	team := wf["team"].(map[string]any)
	if team["enabled"] != true {
		t.Error("team.enabled should be true")
	}
	if team["max_teammates"] != 8 {
		t.Errorf("team.max_teammates = %v, want 8", team["max_teammates"])
	}
	if team["default_model"] != "opus" {
		t.Errorf("team.default_model = %v, want opus", team["default_model"])
	}

	// Verify statusline.yaml
	slParsed := readYAML(t, filepath.Join(sectionsDir, defs.StatuslineYAML))
	sl := slParsed["statusline"].(map[string]any)
	if sl["preset"] != "compact" {
		t.Errorf("statusline.preset = %v, want compact", sl["preset"])
	}

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
