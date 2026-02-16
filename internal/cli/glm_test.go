package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	// Enable test mode to prevent modifying actual settings files
	t.Setenv("MOAI_TEST_MODE", "1")
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
	// In test mode, the command should skip settings modification
	if !strings.Contains(output, "Test environment detected") {
		t.Errorf("output should mention test environment, got %q", output)
	}
}

func TestGLMCmd_InjectsEnv(t *testing.T) {
	// Enable test mode to prevent modifying actual settings files
	t.Setenv("MOAI_TEST_MODE", "1")
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

	// In test mode, settings.local.json should NOT be created
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Error("settings.local.json should not be created in test mode")
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

// --- Tests for project-level .env.glm (issue #384) ---

func TestCreateProjectEnvGLM(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Save a test API key so getGLMAPIKey can find it
	if err := os.MkdirAll(filepath.Join(tmpHome, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(tmpHome, ".moai", ".env.glm"),
		[]byte("GLM_API_KEY=\"test-api-key\"\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	// Create temp project root with .moai directory
	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	glmConfig := &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		Models: struct {
			Haiku  string
			Sonnet string
			Opus   string
		}{
			Haiku:  "glm-4.7-flashx",
			Sonnet: "glm-4.7",
			Opus:   "glm-5",
		},
		EnvVar: "GLM_API_KEY",
	}

	err := createProjectEnvGLM(glmConfig, projectRoot)
	if err != nil {
		t.Fatalf("createProjectEnvGLM should succeed, got: %v", err)
	}

	// Verify file was created
	envPath := filepath.Join(projectRoot, ".moai", ".env.glm")
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read project .env.glm: %v", err)
	}

	contentStr := string(content)

	// Verify all ANTHROPIC_* export statements are present
	expectedVars := []string{
		`export ANTHROPIC_AUTH_TOKEN="test-api-key"`,
		`export ANTHROPIC_BASE_URL="https://api.z.ai/api/anthropic"`,
		`export ANTHROPIC_DEFAULT_HAIKU_MODEL="glm-4.7-flashx"`,
		`export ANTHROPIC_DEFAULT_SONNET_MODEL="glm-4.7"`,
		`export ANTHROPIC_DEFAULT_OPUS_MODEL="glm-5"`,
	}
	for _, expected := range expectedVars {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("project .env.glm should contain %q", expected)
		}
	}
}

func TestCreateProjectEnvGLM_CreatesDirectory(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Save a test API key
	if err := os.MkdirAll(filepath.Join(tmpHome, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(tmpHome, ".moai", ".env.glm"),
		[]byte("GLM_API_KEY=\"test-key\"\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	// Create project root WITHOUT .moai directory
	projectRoot := t.TempDir()

	glmConfig := &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		Models: struct {
			Haiku  string
			Sonnet string
			Opus   string
		}{
			Haiku:  "glm-4.7-flashx",
			Sonnet: "glm-4.7",
			Opus:   "glm-5",
		},
		EnvVar: "GLM_API_KEY",
	}

	err := createProjectEnvGLM(glmConfig, projectRoot)
	if err != nil {
		t.Fatalf("createProjectEnvGLM should create .moai dir and succeed, got: %v", err)
	}

	// Verify file exists
	envPath := filepath.Join(projectRoot, ".moai", ".env.glm")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Fatal("expected .moai/.env.glm to be created")
	}
}
