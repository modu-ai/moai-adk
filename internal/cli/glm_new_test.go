package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// --- Tests for loadGLMConfig ---

func TestLoadGLMConfig_FallbackDefaults(t *testing.T) {
	// When deps is nil, loadGLMConfig should return fallback defaults.
	origDeps := deps
	deps = nil
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig("/nonexistent")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error with nil deps, got: %v", err)
	}
	if cfg.BaseURL != "https://api.z.ai/api/anthropic" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://api.z.ai/api/anthropic")
	}
	if cfg.Models.Low != "glm-4.7-flashx" {
		t.Errorf("Models.Low = %q, want %q", cfg.Models.Low, "glm-4.7-flashx")
	}
	if cfg.Models.Medium != "glm-4.7" {
		t.Errorf("Models.Medium = %q, want %q", cfg.Models.Medium, "glm-4.7")
	}
	if cfg.Models.High != "glm-5" {
		t.Errorf("Models.High = %q, want %q", cfg.Models.High, "glm-5")
	}
	if cfg.EnvVar != "GLM_API_KEY" {
		t.Errorf("EnvVar = %q, want %q", cfg.EnvVar, "GLM_API_KEY")
	}
}

func TestLoadGLMConfig_DepsWithNilConfig(t *testing.T) {
	// When deps.Config is nil, loadGLMConfig should return fallback defaults.
	origDeps := deps
	deps = &Dependencies{Config: nil}
	defer func() { deps = origDeps }()

	cfg, err := loadGLMConfig("/nonexistent")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error with nil Config, got: %v", err)
	}
	if cfg.BaseURL != "https://api.z.ai/api/anthropic" {
		t.Errorf("BaseURL = %q, want fallback default", cfg.BaseURL)
	}
}

func TestLoadGLMConfig_DepsWithEmptyBaseURL(t *testing.T) {
	// When deps.Config returns a config with empty GLM BaseURL,
	// loadGLMConfig should return fallback defaults.
	origDeps := deps
	mgr := config.NewConfigManager()
	deps = &Dependencies{Config: mgr}
	defer func() { deps = origDeps }()

	// ConfigManager.Get() returns nil when not loaded, so fallback is used.
	cfg, err := loadGLMConfig("/nonexistent")
	if err != nil {
		t.Fatalf("loadGLMConfig should not error, got: %v", err)
	}
	if cfg.BaseURL != "https://api.z.ai/api/anthropic" {
		t.Errorf("BaseURL = %q, want fallback default", cfg.BaseURL)
	}
}

// --- Tests for injectGLMEnv ---

func TestInjectGLMEnv_Success(t *testing.T) {
	tmpDir := t.TempDir()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Save a test API key so getGLMAPIKey finds it.
	moaiDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(moaiDir, ".env.glm"),
		[]byte("GLM_API_KEY=\"my-test-key\"\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	settingsPath := filepath.Join(tmpDir, ".claude", "settings.local.json")
	glmConfig := &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		Models: struct {
			High   string
			Medium string
			Low    string
		}{
			High:   "glm-5",
			Medium: "glm-4.7",
			Low:    "glm-4.7-flashx",
		},
		EnvVar: "GLM_API_KEY",
	}

	err := injectGLMEnv(settingsPath, glmConfig)
	if err != nil {
		t.Fatalf("injectGLMEnv should succeed, got: %v", err)
	}

	// Verify file was created and contains expected env vars.
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings file: %v", err)
	}

	var settings SettingsLocal
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to unmarshal settings: %v", err)
	}

	expectedVars := map[string]string{
		"ANTHROPIC_AUTH_TOKEN":           "my-test-key",
		"ANTHROPIC_BASE_URL":            "https://api.z.ai/api/anthropic",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-flashx",
		"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
	}
	for k, v := range expectedVars {
		got, ok := settings.Env[k]
		if !ok {
			t.Errorf("settings.Env missing key %q", k)
		} else if got != v {
			t.Errorf("settings.Env[%q] = %q, want %q", k, got, v)
		}
	}
}

func TestInjectGLMEnv_NoAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	// Clear GLM_API_KEY env var.
	t.Setenv("GLM_API_KEY", "")

	settingsPath := filepath.Join(tmpDir, ".claude", "settings.local.json")
	glmConfig := &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		EnvVar:  "GLM_API_KEY",
	}

	err := injectGLMEnv(settingsPath, glmConfig)
	if err == nil {
		t.Fatal("injectGLMEnv should error when no API key is available")
	}
	if !strings.Contains(err.Error(), "GLM API key not found") {
		t.Errorf("error should mention API key not found, got: %v", err)
	}
}

func TestInjectGLMEnv_MergesExistingSettings(t *testing.T) {
	tmpDir := t.TempDir()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Save a test API key.
	moaiDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(moaiDir, ".env.glm"),
		[]byte("GLM_API_KEY=\"merge-key\"\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	// Create existing settings file with some env vars.
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := SettingsLocal{
		Env: map[string]string{
			"EXISTING_VAR": "keep-me",
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	glmConfig := &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		Models: struct {
			High   string
			Medium string
			Low    string
		}{
			High:   "o",
			Medium: "s",
			Low:    "h",
		},
		EnvVar: "GLM_API_KEY",
	}

	err := injectGLMEnv(settingsPath, glmConfig)
	if err != nil {
		t.Fatalf("injectGLMEnv should succeed, got: %v", err)
	}

	// Verify existing vars are preserved.
	data, _ = os.ReadFile(settingsPath)
	var result SettingsLocal
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if result.Env["EXISTING_VAR"] != "keep-me" {
		t.Error("existing EXISTING_VAR should be preserved after inject")
	}
	if result.Env["ANTHROPIC_AUTH_TOKEN"] != "merge-key" {
		t.Error("ANTHROPIC_AUTH_TOKEN should be set")
	}
}

func TestInjectGLMEnv_InvalidExistingJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Save a test API key.
	moaiDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(moaiDir, ".env.glm"),
		[]byte("GLM_API_KEY=\"test\"\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	// Create invalid JSON file.
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	if err := os.WriteFile(settingsPath, []byte("{invalid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	glmConfig := &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		EnvVar:  "GLM_API_KEY",
	}

	err := injectGLMEnv(settingsPath, glmConfig)
	if err == nil {
		t.Fatal("injectGLMEnv should error on invalid existing JSON")
	}
	if !strings.Contains(err.Error(), "parse settings.local.json") {
		t.Errorf("error should mention parsing, got: %v", err)
	}
}

func TestInjectGLMEnv_FromEnvironmentVariable(t *testing.T) {
	tmpDir := t.TempDir()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	// Set API key via environment variable (no .env.glm file).
	t.Setenv("MY_GLM_KEY", "env-api-key")

	settingsPath := filepath.Join(tmpDir, ".claude", "settings.local.json")
	glmConfig := &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		Models: struct {
			High   string
			Medium string
			Low    string
		}{
			High:   "o",
			Medium: "s",
			Low:    "h",
		},
		EnvVar: "MY_GLM_KEY",
	}

	err := injectGLMEnv(settingsPath, glmConfig)
	if err != nil {
		t.Fatalf("injectGLMEnv should succeed with env var, got: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
	var settings SettingsLocal
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if settings.Env["ANTHROPIC_AUTH_TOKEN"] != "env-api-key" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q, want %q", settings.Env["ANTHROPIC_AUTH_TOKEN"], "env-api-key")
	}
}

// --- Tests for getGLMEnvPath ---

func TestGetGLMEnvPath_ReturnsExpectedPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	got := getGLMEnvPath()
	expected := filepath.Join(tmpHome, ".moai", ".env.glm")
	if got != expected {
		t.Errorf("getGLMEnvPath() = %q, want %q", got, expected)
	}
}

func TestGetGLMEnvPath_ContainsMoaiDir(t *testing.T) {
	path := getGLMEnvPath()
	if path == "" {
		t.Skip("cannot determine home directory")
	}
	if !strings.Contains(path, ".moai") {
		t.Errorf("getGLMEnvPath() = %q, should contain '.moai'", path)
	}
	if !strings.HasSuffix(path, ".env.glm") {
		t.Errorf("getGLMEnvPath() = %q, should end with '.env.glm'", path)
	}
}

// --- Tests for saveGLMKey (additional coverage for edge cases) ---

func TestSaveGLMKey_DirectoryCreation(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// .moai directory does not exist yet.
	err := saveGLMKey("new-key")
	if err != nil {
		t.Fatalf("saveGLMKey should create directory, got: %v", err)
	}

	envPath := filepath.Join(tmpHome, ".moai", ".env.glm")
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("env file should exist: %v", err)
	}
	// Verify file permissions (owner-only read/write).
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestSaveGLMKey_FileContainsHeader(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	if err := saveGLMKey("header-test"); err != nil {
		t.Fatal(err)
	}

	envPath := filepath.Join(tmpHome, ".moai", ".env.glm")
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatal(err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "# GLM API Key") {
		t.Error("file should contain header comment")
	}
	if !strings.Contains(contentStr, "Generated by moai glm") {
		t.Error("file should contain generation comment")
	}
}

// --- Tests for isTestEnvironment ---

func TestIsTestEnvironment_InGoTest(t *testing.T) {
	// When running under `go test`, isTestEnvironment should return true
	// because os.Args[0] ends with ".test".
	result := isTestEnvironment()
	if !result {
		t.Error("isTestEnvironment() should return true when running under go test")
	}
}

func TestIsTestEnvironment_WithEnvVar(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")
	if !isTestEnvironment() {
		t.Error("isTestEnvironment() should return true when MOAI_TEST_MODE=1")
	}
}

func TestIsTestEnvironment_WithEnvVarNotSet(t *testing.T) {
	// Even without MOAI_TEST_MODE, running under go test should detect test env.
	t.Setenv("MOAI_TEST_MODE", "")
	result := isTestEnvironment()
	// Under go test, os.Args should contain a ".test" suffix.
	if !result {
		t.Skip("not running under go test binary (unusual environment)")
	}
}

// --- Tests for loadGLMKey ---

func TestLoadGLMKey_ValidFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	envDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(envDir, ".env.glm"),
		[]byte("# comment\nGLM_API_KEY=\"loaded-key\"\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	key := loadGLMKey()
	if key != "loaded-key" {
		t.Errorf("loadGLMKey() = %q, want %q", key, "loaded-key")
	}
}

func TestLoadGLMKey_MissingFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	key := loadGLMKey()
	if key != "" {
		t.Errorf("loadGLMKey() = %q, want empty string for missing file", key)
	}
}

func TestLoadGLMKey_SingleQuoted(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	envDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(envDir, ".env.glm"),
		[]byte("GLM_API_KEY='single-quoted-key'\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	key := loadGLMKey()
	if key != "single-quoted-key" {
		t.Errorf("loadGLMKey() = %q, want %q", key, "single-quoted-key")
	}
}

func TestLoadGLMKey_Unquoted(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	envDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(envDir, ".env.glm"),
		[]byte("GLM_API_KEY=unquoted-key\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	key := loadGLMKey()
	if key != "unquoted-key" {
		t.Errorf("loadGLMKey() = %q, want %q", key, "unquoted-key")
	}
}

func TestLoadGLMKey_EmptyFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	envDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(envDir, ".env.glm"),
		[]byte(""),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	key := loadGLMKey()
	if key != "" {
		t.Errorf("loadGLMKey() = %q, want empty string for empty file", key)
	}
}

func TestLoadGLMKey_OnlyComments(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	envDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(envDir, ".env.glm"),
		[]byte("# just a comment\n# another comment\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	key := loadGLMKey()
	if key != "" {
		t.Errorf("loadGLMKey() = %q, want empty string for comments-only file", key)
	}
}

// --- Tests for getGLMAPIKey ---

func TestGetGLMAPIKey_FromSavedFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("GLM_API_KEY", "env-fallback")

	envDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(envDir, ".env.glm"),
		[]byte("GLM_API_KEY=\"file-key\"\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	// File key should take priority over env var.
	key := getGLMAPIKey("GLM_API_KEY")
	if key != "file-key" {
		t.Errorf("getGLMAPIKey() = %q, want %q (file takes priority)", key, "file-key")
	}
}

func TestGetGLMAPIKey_FromEnvFallback(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("MY_KEY", "from-env")

	// No saved file.
	key := getGLMAPIKey("MY_KEY")
	if key != "from-env" {
		t.Errorf("getGLMAPIKey() = %q, want %q (env fallback)", key, "from-env")
	}
}

func TestGetGLMAPIKey_NoSource(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("NONEXISTENT_KEY", "")

	key := getGLMAPIKey("NONEXISTENT_KEY")
	if key != "" {
		t.Errorf("getGLMAPIKey() = %q, want empty string when no source", key)
	}
}

// --- Tests for escapeDotenvValue / unescapeDotenvValue round-trip ---

func TestEscapeUnescapeDotenvValue_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"plain", "simple-key-123"},
		{"backslash", `key\with\backslash`},
		{"quotes", `key"with"quotes`},
		{"dollar", `key$with$dollar`},
		{"all_special", `k\e"y$val`},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escaped := escapeDotenvValue(tt.input)
			unescaped := unescapeDotenvValue(escaped)
			if unescaped != tt.input {
				t.Errorf("round-trip failed: input=%q, escaped=%q, unescaped=%q", tt.input, escaped, unescaped)
			}
		})
	}
}

func TestUnescapeDotenvValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"escaped_dollar", `\$HOME`, `$HOME`},
		{"escaped_quote", `\"hello\"`, `"hello"`},
		{"escaped_backslash", `\\path\\to`, `\path\to`},
		{"no_escapes", `plaintext`, `plaintext`},
		{"empty", ``, ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unescapeDotenvValue(tt.input)
			if result != tt.expected {
				t.Errorf("unescapeDotenvValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// --- TAG-006: GLM isolation - only modify settings.local.json, never settings.json ---

func TestInjectGLMEnvForTeam_NeverModifiesSettingsJson(t *testing.T) {
	tmpDir := t.TempDir()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Save a test API key.
	moaiDir := filepath.Join(tmpHome, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(moaiDir, ".env.glm"),
		[]byte("GLM_API_KEY=\"isolation-test\"\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	// Create both settings.json and settings.local.json
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// settings.json content (team-shared, should NOT be modified)
	settingsJSONPath := filepath.Join(claudeDir, "settings.json")
	originalSettingsJSON := map[string]any{
		"permissions": map[string]any{
			"allow": []string{"Read", "Write"},
		},
	}
	settingsData, _ := json.MarshalIndent(originalSettingsJSON, "", "  ")
	if err := os.WriteFile(settingsJSONPath, settingsData, 0o644); err != nil {
		t.Fatal(err)
	}

	// settings.local.json path (should be modified)
	settingsLocalPath := filepath.Join(claudeDir, "settings.local.json")

	glmConfig := &GLMConfigFromYAML{
		BaseURL: "https://api.z.ai/api/anthropic",
		Models: struct {
			High   string
			Medium string
			Low    string
		}{
			High:   "glm-5",
			Medium: "glm-4.7",
			Low:    "glm-4.7-flashx",
		},
		EnvVar: "GLM_API_KEY",
	}

	// Run injectGLMEnvForTeam
	err := injectGLMEnvForTeam(settingsLocalPath, glmConfig, "isolation-test")
	if err != nil {
		t.Fatalf("injectGLMEnvForTeam should succeed, got: %v", err)
	}

	// Verify settings.json was NOT modified
	settingsJSONAfter, err := os.ReadFile(settingsJSONPath)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	var settingsAfter map[string]any
	if err := json.Unmarshal(settingsJSONAfter, &settingsAfter); err != nil {
		t.Fatalf("failed to unmarshal settings.json: %v", err)
	}

	// Verify original permissions are unchanged
	perms, ok := settingsAfter["permissions"].(map[string]any)
	if !ok {
		t.Fatal("settings.json should still have permissions")
	}
	allow, ok := perms["allow"].([]any)
	if !ok || len(allow) != 2 {
		t.Errorf("settings.json permissions.allow should be unchanged, got: %v", allow)
	}

	// Verify settings.json does NOT have GLM env vars
	if env, exists := settingsAfter["env"]; exists {
		if envMap, ok := env.(map[string]any); ok {
			if _, hasToken := envMap["ANTHROPIC_AUTH_TOKEN"]; hasToken {
				t.Error("settings.json should NOT contain ANTHROPIC_AUTH_TOKEN (GLM isolation)")
			}
		}
	}

	// Verify settings.local.json WAS modified
	settingsLocalAfter, err := os.ReadFile(settingsLocalPath)
	if err != nil {
		t.Fatalf("failed to read settings.local.json: %v", err)
	}
	var settingsLocal SettingsLocal
	if err := json.Unmarshal(settingsLocalAfter, &settingsLocal); err != nil {
		t.Fatalf("failed to unmarshal settings.local.json: %v", err)
	}
	if settingsLocal.Env["ANTHROPIC_AUTH_TOKEN"] != "isolation-test" {
		t.Error("settings.local.json should contain ANTHROPIC_AUTH_TOKEN")
	}
}
