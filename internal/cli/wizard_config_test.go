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

// TestApplyWizardConfig_TokenNotPersisted is the F1 [CRITICAL] security
// regression guard: a captured github_token / gitlab_token MUST NOT be written
// into the git-tracked user.yaml (or any git-tracked plaintext config). The
// wizard delegates credentials to the gh / glab CLI instead of persisting the
// secret. Non-secret username fields are still persisted. The sentinel values
// below are obviously-fake placeholders, not real credentials.
func TestApplyWizardConfig_TokenNotPersisted(t *testing.T) {
	root := setupSectionsDir(t)
	ghToken := "FAKE-gh-token-do-not-store-" + t.Name()
	glToken := "FAKE-gl-token-do-not-store-" + t.Name()
	result := &wizard.WizardResult{
		GitHubUsername: "ghuser",
		GitHubToken:    ghToken,
		GitLabUsername: "gluser",
		GitLabToken:    glToken,
	}
	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	data, err := os.ReadFile(userPath)
	if err != nil {
		t.Fatalf("read user.yaml: %v", err)
	}
	raw := string(data)
	if strings.Contains(raw, ghToken) {
		t.Errorf("user.yaml MUST NOT contain the GitHub token; file content:\n%s", raw)
	}
	if strings.Contains(raw, glToken) {
		t.Errorf("user.yaml MUST NOT contain the GitLab token; file content:\n%s", raw)
	}

	// Non-secret username fields ARE still persisted.
	parsed := readYAML(t, userPath)
	user, ok := parsed["user"].(map[string]any)
	if !ok {
		t.Fatal("expected user key in user.yaml")
	}
	if user["github_username"] != "ghuser" {
		t.Errorf("github_username = %v, want ghuser", user["github_username"])
	}
	if user["gitlab_username"] != "gluser" {
		t.Errorf("gitlab_username = %v, want gluser", user["gitlab_username"])
	}
	if _, present := user["github_token"]; present {
		t.Errorf("github_token key must be absent from user.yaml, got %v", user["github_token"])
	}
	if _, present := user["gitlab_token"]; present {
		t.Errorf("gitlab_token key must be absent from user.yaml, got %v", user["gitlab_token"])
	}
}

// TestApplyWizardConfig_TokenOnlyCreatesNoUserYAML verifies a token-only input
// (no username, no name) does not create a user.yaml at all — a token is not a
// persistable user field.
func TestApplyWizardConfig_TokenOnlyCreatesNoUserYAML(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitHubToken: "FAKE-gh-token-only",
		GitLabToken: "FAKE-gl-token-only",
	}
	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}
	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	if _, err := os.Stat(userPath); !os.IsNotExist(err) {
		t.Error("user.yaml should not be created when only tokens are provided (tokens are not persisted)")
	}
}

func TestApplyWizardConfig_GitHubUsername(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
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

// TestApplyWizardConfig_GitHubToken verifies the F1 security fix: a token-only
// answer persists nothing (a token is not a persistable user field, and the
// secret is never written to a git-tracked file).
func TestApplyWizardConfig_GitHubToken(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitHubToken: "FAKE-gh-token-value",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	if _, err := os.Stat(userPath); !os.IsNotExist(err) {
		t.Error("user.yaml should not be created for a token-only answer (F1: tokens are not persisted)")
	}
}

func TestApplyWizardConfig_GitHubUsernameAndToken(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitHubUsername: "myuser",
		GitHubToken:    "FAKE-gh-token-value",
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
	// F1 security fix: the token MUST NOT be persisted.
	if _, present := user["github_token"]; present {
		t.Errorf("user.github_token must be absent, got %v", user["github_token"])
	}
}

func TestApplyWizardConfig_NoUserYAMLWhenGitHubFieldsEmpty(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
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
		GitHubToken:    "FAKE-gh-token-value",
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
	// F1 security fix: the token MUST NOT be persisted.
	if _, present := user["github_token"]; present {
		t.Errorf("github_token must be absent, got %v", user["github_token"])
	}
}

// --- REQ-4: applyWizardConfig git-strategy.yaml save tests ---

func TestApplyWizardConfig_GitStrategyYAML(t *testing.T) {
	root := setupSectionsDir(t)

	// Pre-create git-strategy.yaml
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	gitStratPath := filepath.Join(sectionsDir, defs.GitStrategyYAML)
	existing := "git_strategy:\n  mode: manual\n  provider: github\n"
	if err := os.WriteFile(gitStratPath, []byte(existing), defs.FilePerm); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}

	result := &wizard.WizardResult{
		GitMode:     "team",
		GitProvider: "github",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	parsed := readYAML(t, gitStratPath)
	gs, ok := parsed["git_strategy"].(map[string]any)
	if !ok {
		t.Fatal("expected git_strategy key in git-strategy.yaml")
	}
	if gs["mode"] != "team" {
		t.Errorf("git_strategy.mode = %v, want %q", gs["mode"], "team")
	}
	if gs["provider"] != "github" {
		t.Errorf("git_strategy.provider = %v, want %q", gs["provider"], "github")
	}
}

func TestApplyWizardConfig_GitStrategyYAML_NoFileCreatedWhenEmpty(t *testing.T) {
	root := setupSectionsDir(t)

	result := &wizard.WizardResult{
		GitMode:     "",
		GitProvider: "",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	gitStratPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.GitStrategyYAML)
	if _, err := os.Stat(gitStratPath); !os.IsNotExist(err) {
		t.Error("git-strategy.yaml should not be created when GitMode and GitProvider are empty")
	}
}

func TestApplyWizardConfig_QualityYAML(t *testing.T) {
	root := setupSectionsDir(t)

	// Pre-create quality.yaml
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	qualityPath := filepath.Join(sectionsDir, defs.QualityYAML)
	existing := "constitution:\n  development_mode: tdd\n  enforce_quality: true\n"
	if err := os.WriteFile(qualityPath, []byte(existing), defs.FilePerm); err != nil {
		t.Fatalf("write quality.yaml: %v", err)
	}

	result := &wizard.WizardResult{
		DevelopmentMode: "ddd",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	parsed := readYAML(t, qualityPath)
	constitution, ok := parsed["constitution"].(map[string]any)
	if !ok {
		t.Fatal("expected constitution key in quality.yaml")
	}
	if constitution["development_mode"] != "ddd" {
		t.Errorf("constitution.development_mode = %v, want %q", constitution["development_mode"], "ddd")
	}
	// Verify existing fields are preserved
	if constitution["enforce_quality"] != true {
		t.Errorf("constitution.enforce_quality should be preserved, got %v", constitution["enforce_quality"])
	}
}

func TestApplyWizardConfig_QualityYAML_NoFileCreatedWhenEmpty(t *testing.T) {
	root := setupSectionsDir(t)

	result := &wizard.WizardResult{
		DevelopmentMode: "",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	qualityPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.QualityYAML)
	if _, err := os.Stat(qualityPath); !os.IsNotExist(err) {
		t.Error("quality.yaml should not be created when DevelopmentMode is empty")
	}
}

// --- REQ-5: GitLab credential save tests ---

func TestApplyWizardConfig_GitLabCredentials(t *testing.T) {
	root := setupSectionsDir(t)

	result := &wizard.WizardResult{
		GitLabUsername: "gluser",
		GitLabToken:    "FAKE-gl-token-value",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	parsed := readYAML(t, userPath)
	user, ok := parsed["user"].(map[string]any)
	if !ok {
		t.Fatal("expected user key in user.yaml")
	}
	if user["gitlab_username"] != "gluser" {
		t.Errorf("user.gitlab_username = %v, want %q", user["gitlab_username"], "gluser")
	}
	// F1 security fix: the token MUST NOT be persisted.
	if _, present := user["gitlab_token"]; present {
		t.Errorf("user.gitlab_token must be absent, got %v", user["gitlab_token"])
	}
}

func TestApplyWizardConfig_GitLabInstanceURL(t *testing.T) {
	root := setupSectionsDir(t)

	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	gitStratPath := filepath.Join(sectionsDir, defs.GitStrategyYAML)
	existing := "git_strategy:\n  mode: personal\n  provider: gitlab\n  gitlab:\n    instance_url: https://gitlab.com\n"
	if err := os.WriteFile(gitStratPath, []byte(existing), defs.FilePerm); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}

	result := &wizard.WizardResult{
		GitMode:           "personal",
		GitProvider:       "gitlab",
		GitLabInstanceURL: "https://gitlab.company.com",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	parsed := readYAML(t, gitStratPath)
	gs, ok := parsed["git_strategy"].(map[string]any)
	if !ok {
		t.Fatal("expected git_strategy key")
	}
	gitlab, ok := gs["gitlab"].(map[string]any)
	if !ok {
		t.Fatal("expected gitlab key in git_strategy")
	}
	if gitlab["instance_url"] != "https://gitlab.company.com" {
		t.Errorf("gitlab.instance_url = %v, want %q", gitlab["instance_url"], "https://gitlab.company.com")
	}
}

func TestApplyWizardConfig_AllFields(t *testing.T) {
	root := setupSectionsDir(t)
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)

	// Pre-create existing files
	qualityExisting := "constitution:\n  development_mode: tdd\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.QualityYAML), []byte(qualityExisting), defs.FilePerm); err != nil {
		t.Fatalf("write quality.yaml: %v", err)
	}
	gitStratExisting := "git_strategy:\n  mode: manual\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.GitStrategyYAML), []byte(gitStratExisting), defs.FilePerm); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}

	result := &wizard.WizardResult{
		GitHubUsername:    "ghuser",
		GitHubToken:       "FAKE-gh-token-value",
		GitLabUsername:    "gluser",
		GitLabToken:       "FAKE-gl-token-value",
		GitLabInstanceURL: "https://self-hosted.gl.com",
		GitMode:           "team",
		GitProvider:       "github",
		DevelopmentMode:   "ddd",
	}

	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	// Verify user.yaml
	userParsed := readYAML(t, filepath.Join(sectionsDir, defs.UserYAML))
	user := userParsed["user"].(map[string]any)
	if user["github_username"] != "ghuser" {
		t.Errorf("github_username = %v, want ghuser", user["github_username"])
	}
	if user["gitlab_username"] != "gluser" {
		t.Errorf("gitlab_username = %v, want gluser", user["gitlab_username"])
	}
	// F1 security fix: neither token is persisted to user.yaml.
	if _, present := user["github_token"]; present {
		t.Errorf("github_token must be absent, got %v", user["github_token"])
	}
	if _, present := user["gitlab_token"]; present {
		t.Errorf("gitlab_token must be absent, got %v", user["gitlab_token"])
	}

	// Verify quality.yaml
	qualityParsed := readYAML(t, filepath.Join(sectionsDir, defs.QualityYAML))
	constitution := qualityParsed["constitution"].(map[string]any)
	if constitution["development_mode"] != "ddd" {
		t.Errorf("development_mode = %v, want ddd", constitution["development_mode"])
	}

	// Verify git-strategy.yaml
	gsParsed := readYAML(t, filepath.Join(sectionsDir, defs.GitStrategyYAML))
	gs := gsParsed["git_strategy"].(map[string]any)
	if gs["mode"] != "team" {
		t.Errorf("git_strategy.mode = %v, want team", gs["mode"])
	}
	if gs["provider"] != "github" {
		t.Errorf("git_strategy.provider = %v, want github", gs["provider"])
	}
}

// --- ConversationLang persistence tests (reconfigure wizard language question) ---

func TestApplyWizardConfig_ConversationLang(t *testing.T) {
	root := setupSectionsDir(t)

	// Pre-create language.yaml with the full schema to verify sibling keys survive.
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	langPath := filepath.Join(sectionsDir, defs.LanguageYAML)
	existing := "language:\n  conversation_language: en\n  conversation_language_name: en\n  agent_prompt_language: en\n  code_comments: en\n"
	if err := os.WriteFile(langPath, []byte(existing), defs.FilePerm); err != nil {
		t.Fatalf("write language.yaml: %v", err)
	}

	result := &wizard.WizardResult{ConversationLang: "ko"}
	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	parsed := readYAML(t, langPath)
	language, ok := parsed["language"].(map[string]any)
	if !ok {
		t.Fatal("expected language key in language.yaml")
	}
	if language["conversation_language"] != "ko" {
		t.Errorf("conversation_language = %v, want %q", language["conversation_language"], "ko")
	}
	if language["conversation_language_name"] != "ko" {
		t.Errorf("conversation_language_name = %v, want %q", language["conversation_language_name"], "ko")
	}
	// Sibling keys must be preserved.
	if language["agent_prompt_language"] != "en" {
		t.Errorf("agent_prompt_language should be preserved, got %v", language["agent_prompt_language"])
	}
	if language["code_comments"] != "en" {
		t.Errorf("code_comments should be preserved, got %v", language["code_comments"])
	}
}

func TestApplyWizardConfig_ConversationLang_NoFileCreatedWhenEmpty(t *testing.T) {
	root := setupSectionsDir(t)

	result := &wizard.WizardResult{ConversationLang: ""}
	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	langPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.LanguageYAML)
	if _, err := os.Stat(langPath); !os.IsNotExist(err) {
		t.Error("language.yaml should not be created when ConversationLang is empty")
	}
}

// --- UserName persistence tests (reconfigure wizard user_name question) ---

func TestApplyWizardConfig_UserName(t *testing.T) {
	root := setupSectionsDir(t)

	result := &wizard.WizardResult{UserName: "GOOS"}
	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	parsed := readYAML(t, userPath)
	user, ok := parsed["user"].(map[string]any)
	if !ok {
		t.Fatal("expected user key in user.yaml")
	}
	if user["name"] != "GOOS" {
		t.Errorf("user.name = %v, want %q", user["name"], "GOOS")
	}
}

// TestApplyWizardConfig_UserNameWithoutGitCredentials verifies the user.yaml
// write triggers on UserName alone — no git credentials present.
func TestApplyWizardConfig_UserNameWithoutGitCredentials(t *testing.T) {
	root := setupSectionsDir(t)

	result := &wizard.WizardResult{UserName: "Solo"}
	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	if _, err := os.Stat(userPath); err != nil {
		t.Fatalf("user.yaml should be created when only UserName is set: %v", err)
	}
	parsed := readYAML(t, userPath)
	user := parsed["user"].(map[string]any)
	if user["name"] != "Solo" {
		t.Errorf("user.name = %v, want %q", user["name"], "Solo")
	}
}

// TestApplyWizardConfig_UserNamePreservesGitCredentials verifies user.name does
// not clobber github/gitlab credentials written in the same call.
func TestApplyWizardConfig_UserNamePreservesGitCredentials(t *testing.T) {
	root := setupSectionsDir(t)

	result := &wizard.WizardResult{
		UserName:       "Both",
		GitHubUsername: "ghuser",
	}
	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("applyWizardConfig: %v", err)
	}

	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	parsed := readYAML(t, userPath)
	user := parsed["user"].(map[string]any)
	if user["name"] != "Both" {
		t.Errorf("user.name = %v, want %q", user["name"], "Both")
	}
	if user["github_username"] != "ghuser" {
		t.Errorf("github_username = %v, want %q", user["github_username"], "ghuser")
	}
}

// --- F3: input validation tests ---

// TestApplyWizardConfig_RejectsMalformedGitLabURL verifies a gitlab_instance_url
// without a scheme is rejected and nothing is persisted.
func TestApplyWizardConfig_RejectsMalformedGitLabURL(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitMode:           "personal",
		GitProvider:       "gitlab",
		GitLabInstanceURL: "not-a-url",
	}
	if err := applyWizardConfig(root, result); err == nil {
		t.Fatal("expected error for malformed gitlab_instance_url, got nil")
	}
	// No config file should be written when validation fails.
	gitStratPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.GitStrategyYAML)
	if _, err := os.Stat(gitStratPath); !os.IsNotExist(err) {
		t.Error("git-strategy.yaml must not be written when input is invalid")
	}
}

// TestApplyWizardConfig_RejectsHTTPGitLabURL verifies a plaintext http:// URL is
// rejected (https required).
func TestApplyWizardConfig_RejectsHTTPGitLabURL(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitMode:           "team",
		GitProvider:       "gitlab",
		GitLabInstanceURL: "http://gitlab.company.com",
	}
	if err := applyWizardConfig(root, result); err == nil {
		t.Fatal("expected error for http:// gitlab_instance_url, got nil")
	}
}

// TestApplyWizardConfig_RejectsMalformedGitHubUsername verifies a username with
// illegal characters is rejected and user.yaml is not written.
func TestApplyWizardConfig_RejectsMalformedGitHubUsername(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitHubUsername: "bad user!",
	}
	if err := applyWizardConfig(root, result); err == nil {
		t.Fatal("expected error for malformed github username, got nil")
	}
	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	if _, err := os.Stat(userPath); !os.IsNotExist(err) {
		t.Error("user.yaml must not be written when username is invalid")
	}
}

// TestApplyWizardConfig_RejectsLeadingHyphenUsername verifies a leading-hyphen
// GitHub username is rejected.
func TestApplyWizardConfig_RejectsLeadingHyphenUsername(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{GitHubUsername: "-badstart"}
	if err := applyWizardConfig(root, result); err == nil {
		t.Fatal("expected error for leading-hyphen github username, got nil")
	}
}

// TestApplyWizardConfig_AcceptsValidIdentityInput verifies well-formed input is
// accepted and persisted.
func TestApplyWizardConfig_AcceptsValidIdentityInput(t *testing.T) {
	root := setupSectionsDir(t)
	result := &wizard.WizardResult{
		GitMode:           "personal",
		GitProvider:       "gitlab",
		GitHubUsername:    "octo-cat",
		GitLabUsername:    "octo.cat_1",
		GitLabInstanceURL: "https://gitlab.example.com",
	}
	if err := applyWizardConfig(root, result); err != nil {
		t.Fatalf("valid input should be accepted: %v", err)
	}
	userPath := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML)
	parsed := readYAML(t, userPath)
	user := parsed["user"].(map[string]any)
	if user["github_username"] != "octo-cat" {
		t.Errorf("github_username = %v, want octo-cat", user["github_username"])
	}
}

// presetToSegments tests removed (SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001):
// the lowercase presetToSegments wrapper in update.go and the capital
// statusline.PresetToSegments function were both deleted. Named presets no
// longer exist as a configuration surface — segments are configured directly.
