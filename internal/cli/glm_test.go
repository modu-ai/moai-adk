package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestGLMCmd_Exists(t *testing.T) {
	if glmCmd == nil {
		t.Fatal("glmCmd should not be nil")
	}
}

func TestGLMCmd_Use(t *testing.T) {
	if !strings.HasPrefix(glmCmd.Use, "glm") {
		t.Errorf("glmCmd.Use should start with 'glm', got %q", glmCmd.Use)
	}
}

func TestGLMCmd_Short(t *testing.T) {
	if glmCmd.Short == "" {
		t.Error("glmCmd.Short should not be empty")
	}
}

func TestGLMCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "glm" {
			found = true
			break
		}
	}
	if !found {
		t.Error("glm should be registered as a subcommand of root")
	}
}

func TestGLMCmd_NoArgs(t *testing.T) {
	// Set GLM_API_KEY env var
	t.Setenv("GLM_API_KEY", "test-api-key")

	// Create temp project
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("glm should not error, got: %v", err)
	}

	output := buf.String()
	// GLM Team mode should be enabled
	if !strings.Contains(output, "GLM Team mode enabled") {
		t.Errorf("output should mention GLM Team mode enabled, got %q", output)
	}
}

func TestGLMCmd_InjectsEnv(t *testing.T) {
	// Set GLM_API_KEY env var
	t.Setenv("GLM_API_KEY", "test-api-key")

	// Create temp project
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("glm should not error, got: %v", err)
	}

	// GLM Team mode should create settings.local.json with GLM env vars
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings.local.json should be created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "ANTHROPIC_AUTH_TOKEN") {
		t.Error("settings.local.json should contain ANTHROPIC_AUTH_TOKEN")
	}
	if !strings.Contains(content, "ANTHROPIC_BASE_URL") {
		t.Error("settings.local.json should contain ANTHROPIC_BASE_URL")
	}
	if !strings.Contains(content, "CLAUDE_CODE_TEAMMATE_DISPLAY") {
		t.Error("settings.local.json should contain CLAUDE_CODE_TEAMMATE_DISPLAY")
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Create temp project
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("findProjectRoot should succeed: %v", err)
	}

	// Normalize paths for comparison
	expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
	actualRoot, _ := filepath.EvalSymlinks(root)
	if actualRoot != expectedRoot {
		t.Errorf("findProjectRoot returned %q, expected %q", actualRoot, expectedRoot)
	}
}

func TestFindProjectRoot_NotInProject(t *testing.T) {
	// Create temp dir without .moai
	tmpDir := t.TempDir()

	// Verify no .moai exists in the parent chain of tmpDir.
	// When running from within a MoAI project, t.TempDir() may resolve
	// to a path whose ancestor contains .moai, causing findProjectRoot()
	// to succeed unexpectedly.
	dir := tmpDir
	for {
		if _, err := os.Stat(filepath.Join(dir, ".moai")); err == nil {
			t.Skip("temp dir is under a MoAI project directory; skipping test")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	_, err := findProjectRoot()
	if err == nil {
		t.Error("findProjectRoot should error when not in a MoAI project")
	}
}

// --- DDD PRESERVE: Characterization tests for GLM utility functions ---

func TestEscapeDotenvValue_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "backslash",
			input:    `key\value`,
			expected: `key\\value`,
		},
		{
			name:     "double quote",
			input:    `key"value`,
			expected: `key\"value`,
		},
		{
			name:     "dollar sign",
			input:    `key$value`,
			expected: `key\$value`,
		},
		{
			name:     "multiple special chars",
			input:    `key"$value`,
			expected: `key\"\$value`,
		},
		{
			name:     "no special chars",
			input:    `keyvalue123`,
			expected: `keyvalue123`,
		},
		{
			name:     "empty string",
			input:    ``,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeDotenvValue(tt.input)
			if result != tt.expected {
				t.Errorf("escapeDotenvValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSaveGLMKey_Success(t *testing.T) {
	// Create temp home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome) // Windows: os.UserHomeDir() checks USERPROFILE first

	testKey := "test-api-key-12345"

	err := saveGLMKey(testKey)
	if err != nil {
		t.Fatalf("saveGLMKey should succeed, got error: %v", err)
	}

	// Verify file was created
	envPath := filepath.Join(tmpHome, ".moai", ".env.glm")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Fatalf("expected .env.glm file to be created at %s", envPath)
	}

	// Verify file content
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read .env.glm: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "GLM_API_KEY") {
		t.Error("file should contain GLM_API_KEY")
	}
	if !strings.Contains(contentStr, testKey) {
		t.Error("file should contain the API key")
	}
}

func TestSaveGLMKey_SpecialCharacters(t *testing.T) {
	// Create temp home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome) // Windows: os.UserHomeDir() checks USERPROFILE first

	// Key with special characters that need escaping
	testKey := `key"with$special\chars`

	err := saveGLMKey(testKey)
	if err != nil {
		t.Fatalf("saveGLMKey should succeed with special chars, got error: %v", err)
	}

	// Load the key back
	loadedKey := loadGLMKey()
	if loadedKey != testKey {
		t.Errorf("loaded key %q does not match saved key %q", loadedKey, testKey)
	}
}

func TestSaveGLMKey_EmptyKey(t *testing.T) {
	// Create temp home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome) // Windows: os.UserHomeDir() checks USERPROFILE first

	err := saveGLMKey("")
	if err != nil {
		t.Fatalf("saveGLMKey should succeed with empty key, got error: %v", err)
	}

	// Verify file was created
	envPath := filepath.Join(tmpHome, ".moai", ".env.glm")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Fatal("expected .env.glm file to be created")
	}
}

// TestResolveGLMModels verifies backward compatibility fallback logic.
func TestResolveGLMModels(t *testing.T) {
	defaults := config.NewDefaultLLMConfig()

	tests := []struct {
		name       string
		models     config.GLMModels
		wantHigh   string
		wantMedium string
		wantLow    string
	}{
		{
			name: "only High/Medium/Low set - uses them directly",
			models: config.GLMModels{
				High:   "custom-high",
				Medium: "custom-medium",
				Low:    "custom-low",
			},
			wantHigh:   "custom-high",
			wantMedium: "custom-medium",
			wantLow:    "custom-low",
		},
		{
			name: "only Opus/Sonnet/Haiku set - falls back to legacy fields",
			models: config.GLMModels{
				Opus:   "legacy-opus",
				Sonnet: "legacy-sonnet",
				Haiku:  "legacy-haiku",
			},
			wantHigh:   "legacy-opus",
			wantMedium: "legacy-sonnet",
			wantLow:    "legacy-haiku",
		},
		{
			name: "both set - High/Medium/Low takes priority over Opus/Sonnet/Haiku",
			models: config.GLMModels{
				High:   "new-high",
				Medium: "new-medium",
				Low:    "new-low",
				Opus:   "old-opus",
				Sonnet: "old-sonnet",
				Haiku:  "old-haiku",
			},
			wantHigh:   "new-high",
			wantMedium: "new-medium",
			wantLow:    "new-low",
		},
		{
			name:       "neither set - falls back to config defaults",
			models:     config.GLMModels{},
			wantHigh:   defaults.GLM.Models.High,
			wantMedium: defaults.GLM.Models.Medium,
			wantLow:    defaults.GLM.Models.Low,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHigh, gotMedium, gotLow := resolveGLMModels(tt.models)
			if gotHigh != tt.wantHigh {
				t.Errorf("high = %q, want %q", gotHigh, tt.wantHigh)
			}
			if gotMedium != tt.wantMedium {
				t.Errorf("medium = %q, want %q", gotMedium, tt.wantMedium)
			}
			if gotLow != tt.wantLow {
				t.Errorf("low = %q, want %q", gotLow, tt.wantLow)
			}
		})
	}
}

// TestLoadGLMConfig_BackwardCompatibility reproduces Issue #409:
// Legacy users with llm.yaml configured using old field names (opus, sonnet, haiku)
// should get correct model mapping via resolveGLMModels fallback.
// NOTE: This test is now covered by glm_compat_test.go tests.
func TestLoadGLMConfig_BackwardCompatibility(t *testing.T) {
	t.Skip("Covered by glm_compat_test.go tests - kept for historical reference")
}

// configManagerMock is a minimal mock of config.ConfigManager for testing.
type configManagerMock struct {
	cfg *config.Config
}

func (m *configManagerMock) Get() *config.Config {
	return m.cfg
}

func (m *configManagerMock) Reload() error {
	return nil
}

func TestSaveGLMKey_OverwriteExisting(t *testing.T) {
	// Create temp home directory
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome) // Windows: os.UserHomeDir() checks USERPROFILE first

	// Save first key
	firstKey := "first-key"
	err := saveGLMKey(firstKey)
	if err != nil {
		t.Fatalf("first saveGLMKey failed: %v", err)
	}

	// Save second key (should overwrite)
	secondKey := "second-key"
	err = saveGLMKey(secondKey)
	if err != nil {
		t.Fatalf("second saveGLMKey failed: %v", err)
	}

	// Verify second key was saved
	loadedKey := loadGLMKey()
	if loadedKey != secondKey {
		t.Errorf("loaded key %q, want %q", loadedKey, secondKey)
	}
	if loadedKey == firstKey {
		t.Error("first key should be overwritten")
	}
}
