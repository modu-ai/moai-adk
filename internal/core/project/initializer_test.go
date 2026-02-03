package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk-go/internal/manifest"
)

// --- Mock implementations for testing ---

type mockDeployer struct {
	deployErr error
	deployed  bool
}

func (m *mockDeployer) Deploy(_ context.Context, _ string, _ manifest.Manager) error {
	m.deployed = true
	return m.deployErr
}

func (m *mockDeployer) ExtractTemplate(_ string) ([]byte, error) {
	return nil, nil
}

func (m *mockDeployer) ListTemplates() []string {
	return nil
}

// --- Initializer tests ---

func TestInit_BasicInitialization(t *testing.T) {
	root := t.TempDir()

	mgr := manifest.NewManager()
	init := NewInitializer(nil, mgr, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "test-project",
		Language:        "Go",
		Framework:       "none",
		Features:        []string{},
		UserName:        "testuser",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	result, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify directories were created
	for _, dir := range moaiDirs {
		dirPath := filepath.Join(root, ".moai", dir)
		if !dirExists(dirPath) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	for _, dir := range claudeDirs {
		dirPath := filepath.Join(root, ".claude", dir)
		if !dirExists(dirPath) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	// Verify config files
	assertFileExists(t, filepath.Join(root, ".moai", "config", "sections", "user.yaml"))
	assertFileExists(t, filepath.Join(root, ".moai", "config", "sections", "language.yaml"))
	assertFileExists(t, filepath.Join(root, ".moai", "config", "sections", "quality.yaml"))
	assertFileExists(t, filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"))
	assertFileExists(t, filepath.Join(root, "CLAUDE.md"))

	// Verify result
	if result.DevelopmentMode != "ddd" {
		t.Errorf("DevelopmentMode = %q, want %q", result.DevelopmentMode, "ddd")
	}
	if len(result.CreatedFiles) == 0 {
		t.Error("expected at least one created file")
	}
	if len(result.CreatedDirs) == 0 {
		t.Error("expected at least one created directory")
	}
}

func TestInit_UserYAMLContent(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "my-app",
		Language:        "Python",
		Framework:       "FastAPI",
		UserName:        "GOOS",
		ConvLang:        "ko",
		DevelopmentMode: "tdd",
	}

	_, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify user.yaml
	userPath := filepath.Join(root, ".moai", "config", "sections", "user.yaml")
	var userYAMLData userYAML
	readYAML(t, userPath, &userYAMLData)
	if userYAMLData.User.Name != "GOOS" {
		t.Errorf("user.name = %q, want %q", userYAMLData.User.Name, "GOOS")
	}
}

func TestInit_LanguageYAMLContent(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "my-app",
		Language:        "Python",
		UserName:        "test",
		ConvLang:        "ko",
		DevelopmentMode: "ddd",
	}

	_, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	langPath := filepath.Join(root, ".moai", "config", "sections", "language.yaml")
	var langYAMLData languageYAML
	readYAML(t, langPath, &langYAMLData)

	if langYAMLData.Language.ConversationLanguage != "ko" {
		t.Errorf("conversation_language = %q, want %q", langYAMLData.Language.ConversationLanguage, "ko")
	}
	if langYAMLData.Language.AgentPromptLanguage != "en" {
		t.Errorf("agent_prompt_language = %q, want %q", langYAMLData.Language.AgentPromptLanguage, "en")
	}
}

func TestInit_QualityYAMLContent(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "my-app",
		Language:        "Go",
		UserName:        "test",
		ConvLang:        "en",
		DevelopmentMode: "hybrid",
	}

	_, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	qualityPath := filepath.Join(root, ".moai", "config", "sections", "quality.yaml")
	var qualYAMLData qualityYAML
	readYAML(t, qualityPath, &qualYAMLData)

	if qualYAMLData.Constitution.DevelopmentMode != "hybrid" {
		t.Errorf("development_mode = %q, want %q", qualYAMLData.Constitution.DevelopmentMode, "hybrid")
	}
	if !qualYAMLData.Constitution.EnforceQuality {
		t.Error("enforce_quality should be true")
	}
	if qualYAMLData.Constitution.TestCoverageTarget != 85 {
		t.Errorf("test_coverage_target = %d, want 85", qualYAMLData.Constitution.TestCoverageTarget)
	}
}

func TestInit_InvalidDevelopmentModeFallsBackToDDD(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "my-app",
		Language:        "Go",
		UserName:        "test",
		ConvLang:        "en",
		DevelopmentMode: "invalid-mode",
	}

	_, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	qualityPath := filepath.Join(root, ".moai", "config", "sections", "quality.yaml")
	var qualYAMLData qualityYAML
	readYAML(t, qualityPath, &qualYAMLData)

	if qualYAMLData.Constitution.DevelopmentMode != "ddd" {
		t.Errorf("development_mode = %q, want fallback %q", qualYAMLData.Constitution.DevelopmentMode, "ddd")
	}
}

func TestInit_CLAUDEMDContent(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "awesome-app",
		Language:        "TypeScript",
		Framework:       "Next.js",
		UserName:        "test",
		ConvLang:        "en",
		DevelopmentMode: "tdd",
	}

	_, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	claudeMDPath := filepath.Join(root, "CLAUDE.md")
	data, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "awesome-app") {
		t.Error("CLAUDE.md should contain project name")
	}
	if !strings.Contains(content, "TypeScript") {
		t.Error("CLAUDE.md should contain language")
	}
	if !strings.Contains(content, "Next.js") {
		t.Error("CLAUDE.md should contain framework")
	}
	if !strings.Contains(content, "tdd") {
		t.Error("CLAUDE.md should contain development mode")
	}
}

func TestInit_NoDynamicTokensInGeneratedFiles(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "test",
		Language:        "Go",
		UserName:        "test",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	_, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// REQ-N-003: No unexpanded dynamic tokens
	dynamicTokens := []string{"$VAR", "${", "{{", "}}"}
	files := []string{
		filepath.Join(root, ".moai", "config", "sections", "user.yaml"),
		filepath.Join(root, ".moai", "config", "sections", "language.yaml"),
		filepath.Join(root, ".moai", "config", "sections", "quality.yaml"),
		filepath.Join(root, "CLAUDE.md"),
	}

	for _, f := range files {
		data, readErr := os.ReadFile(f)
		if readErr != nil {
			t.Fatalf("read %s: %v", filepath.Base(f), readErr)
		}
		content := string(data)
		for _, token := range dynamicTokens {
			if strings.Contains(content, token) {
				t.Errorf("file %s contains forbidden token %q", filepath.Base(f), token)
			}
		}
	}
}

func TestInit_YAMLFilesAreParseable(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "test",
		Language:        "Go",
		UserName:        "test",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	_, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// REQ-N-002: All YAML files must be parseable (generated via struct serialization)
	yamlFiles := []string{
		filepath.Join(root, ".moai", "config", "sections", "user.yaml"),
		filepath.Join(root, ".moai", "config", "sections", "language.yaml"),
		filepath.Join(root, ".moai", "config", "sections", "quality.yaml"),
		filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"),
	}

	for _, f := range yamlFiles {
		data, readErr := os.ReadFile(f)
		if readErr != nil {
			t.Fatalf("read %s: %v", filepath.Base(f), readErr)
		}
		var raw any
		if yamlErr := yaml.Unmarshal(data, &raw); yamlErr != nil {
			t.Errorf("file %s is not valid YAML: %v", filepath.Base(f), yamlErr)
		}
	}
}

func TestInit_ManifestIsValidJSON(t *testing.T) {
	root := t.TempDir()
	mgr := manifest.NewManager()
	init := NewInitializer(nil, mgr, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "test",
		Language:        "Go",
		UserName:        "test",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	_, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// REQ-U-005: JSON files validated with json.Valid()
	manifestPath := filepath.Join(root, ".moai", "manifest.json")
	data, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		t.Fatalf("read manifest.json: %v", readErr)
	}
	if !json.Valid(data) {
		t.Error("manifest.json is not valid JSON")
	}
}

func TestInit_ContextCancellation(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "test",
		Language:        "Go",
		UserName:        "test",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	_, err := init.Init(ctx, opts)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestInit_FilePathsAreClean(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root + "/./", // intentionally messy path
		ProjectName:     "test",
		Language:        "Go",
		UserName:        "test",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	result, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// REQ-U-001: All paths use filepath.Clean()
	for _, f := range result.CreatedFiles {
		cleaned := filepath.Clean(f)
		if f != cleaned {
			t.Errorf("file path %q is not clean (expected %q)", f, cleaned)
		}
	}

	for _, d := range result.CreatedDirs {
		cleaned := filepath.Clean(d)
		if d != cleaned {
			t.Errorf("dir path %q is not clean (expected %q)", d, cleaned)
		}
	}
}

func TestResolveLanguageName(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"en", "English"},
		{"ko", "Korean (한국어)"},
		{"ja", "Japanese (日本語)"},
		{"zh", "Chinese (中文)"},
		{"es", "Spanish (Español)"},
		{"fr", "French (Français)"},
		{"de", "German (Deutsch)"},
		{"unknown", "English"}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := resolveLanguageName(tt.code)
			if got != tt.want {
				t.Errorf("resolveLanguageName(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestInit_WithDeployer(t *testing.T) {
	root := t.TempDir()

	deployer := &mockDeployer{}
	mgr := manifest.NewManager()
	init := NewInitializer(deployer, mgr, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "test-deploy",
		Language:        "Go",
		Framework:       "none",
		UserName:        "testuser",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	result, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if !deployer.deployed {
		t.Error("expected deployer.Deploy to have been called")
	}

	if len(result.CreatedFiles) == 0 {
		t.Error("expected at least one created file")
	}
}

func TestInit_WithDeployerError(t *testing.T) {
	root := t.TempDir()

	deployer := &mockDeployer{deployErr: fmt.Errorf("deploy failed")}
	mgr := manifest.NewManager()
	init := NewInitializer(deployer, mgr, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "test-deploy-err",
		Language:        "Go",
		Framework:       "none",
		UserName:        "testuser",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	result, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v (should be non-fatal warning)", err)
	}

	// Deploy error should be recorded as warning, not a fatal error
	if len(result.Warnings) == 0 {
		t.Error("expected at least one warning for deploy failure")
	}

	foundWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "template deployment") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("expected template deployment warning, got %v", result.Warnings)
	}
}

func TestInit_WithDeployerNoManifest(t *testing.T) {
	root := t.TempDir()

	deployer := &mockDeployer{}
	// Pass deployer but no manifest manager
	init := NewInitializer(deployer, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "test-no-manifest",
		Language:        "Go",
		Framework:       "none",
		UserName:        "testuser",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	result, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Deploy should still attempt and fail (needs manifest), recorded as warning
	if len(result.Warnings) == 0 {
		t.Error("expected warning when deployer has no manifest manager")
	}
}

func TestInit_WorkflowYAMLContent(t *testing.T) {
	root := t.TempDir()
	init := NewInitializer(nil, nil, nil)

	opts := InitOptions{
		ProjectRoot:     root,
		ProjectName:     "test",
		Language:        "Go",
		UserName:        "test",
		ConvLang:        "en",
		DevelopmentMode: "ddd",
	}

	_, err := init.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	workflowPath := filepath.Join(root, ".moai", "config", "sections", "workflow.yaml")
	var wfYAMLData workflowYAML
	readYAML(t, workflowPath, &wfYAMLData)

	if !wfYAMLData.Workflow.AutoClear {
		t.Error("auto_clear should be true")
	}
	if wfYAMLData.Workflow.PlanTokens != 30000 {
		t.Errorf("plan_tokens = %d, want 30000", wfYAMLData.Workflow.PlanTokens)
	}
	if wfYAMLData.Workflow.RunTokens != 180000 {
		t.Errorf("run_tokens = %d, want 180000", wfYAMLData.Workflow.RunTokens)
	}
	if wfYAMLData.Workflow.SyncTokens != 40000 {
		t.Errorf("sync_tokens = %d, want 40000", wfYAMLData.Workflow.SyncTokens)
	}
}

// --- Test helpers ---

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", path)
	}
}

func readYAML(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := yaml.Unmarshal(data, target); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
}
