package cli

import (
	"bytes"
	"encoding/json"
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

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	defer func() { findProjectRootFn = origFn }()

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Override launchClaude to skip actual exec
	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(profile string, args []string) error {
		return nil
	}

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("glm should not error, got: %v", err)
	}
}

// TestGLMCmd_NoSettingsLocalPollution verifies that `moai glm` does NOT write
// GLM env vars to settings.local.json. The previous behavior (injectGLMEnvForTeam
// in applyGLMMode) caused GLM mode to persist after `moai glm` exited, routing all
// subsequent `claude` invocations to GLM unexpectedly (#676).
//
// setGLMEnv() already injects env into the current process, which syscall.Exec
// inherits into `claude` — writing to settings.local.json is redundant and harmful.
func TestGLMCmd_NoSettingsLocalPollution(t *testing.T) {
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

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	defer func() { findProjectRootFn = origFn }()

	// Change to temp dir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Override launchClaude to skip actual exec
	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()
	launchClaudeFunc = func(profile string, args []string) error {
		return nil
	}

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{})
	if err != nil {
		t.Fatalf("glm should not error, got: %v", err)
	}

	// GLM should NOT write GLM env vars to settings.local.json (#676 fix).
	// setGLMEnv() injects env into the current process; no file persistence needed.
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		// File not created at all — that's acceptable and expected.
		if !os.IsNotExist(err) {
			t.Fatalf("unexpected error reading settings.local.json: %v", err)
		}
		return
	}

	content := string(data)
	if strings.Contains(content, "ANTHROPIC_BASE_URL") {
		t.Error("settings.local.json must NOT contain ANTHROPIC_BASE_URL after moai glm (regression: #676 persistence bug)")
	}
	if strings.Contains(content, "DISABLE_PROMPT_CACHING") {
		t.Error("settings.local.json must NOT contain DISABLE_PROMPT_CACHING after moai glm (regression: #676 persistence bug)")
	}
}

func TestGLMCmd_WithProfile(t *testing.T) {
	t.Setenv("GLM_API_KEY", "test-api-key")

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()

	var launchedProfile string
	launchClaudeFunc = func(profile string, args []string) error {
		launchedProfile = profile
		return nil
	}

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	err := glmCmd.RunE(glmCmd, []string{"-p", "work"})
	if err != nil {
		t.Fatalf("glm -p work should not error, got: %v", err)
	}

	if launchedProfile != "work" {
		t.Errorf("profile should be 'work', got %q", launchedProfile)
	}
}

func TestGLMCmd_Setup(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	// Route to setup subcommand
	err := glmCmd.RunE(glmCmd, []string{"setup", "test-key-12345"})
	if err != nil {
		t.Fatalf("glm setup should not error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "GLM API key stored") {
		t.Errorf("output should mention key stored, got %q", output)
	}
}

func TestFindProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	moaiDir := filepath.Join(tmpDir, ".moai")
	if err := os.MkdirAll(moaiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	root, err := findProjectRoot()
	if err != nil {
		t.Fatalf("findProjectRoot should succeed: %v", err)
	}

	expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
	actualRoot, _ := filepath.EvalSymlinks(root)
	if actualRoot != expectedRoot {
		t.Errorf("findProjectRoot returned %q, expected %q", actualRoot, expectedRoot)
	}
}

func TestFindProjectRoot_NotInProject(t *testing.T) {
	tmpDir := t.TempDir()

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
		{name: "backslash", input: `key\value`, expected: `key\\value`},
		{name: "double quote", input: `key"value`, expected: `key\"value`},
		{name: "dollar sign", input: `key$value`, expected: `key\$value`},
		{name: "multiple special chars", input: `key"$value`, expected: `key\"\$value`},
		{name: "no special chars", input: `keyvalue123`, expected: `keyvalue123`},
		{name: "empty string", input: ``, expected: ``},
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
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	testKey := "test-api-key-12345"
	err := saveGLMKey(testKey)
	if err != nil {
		t.Fatalf("saveGLMKey should succeed, got error: %v", err)
	}

	envPath := filepath.Join(tmpHome, ".moai", ".env.glm")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Fatalf("expected .env.glm file to be created at %s", envPath)
	}

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
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	testKey := `key"with$special\chars`
	err := saveGLMKey(testKey)
	if err != nil {
		t.Fatalf("saveGLMKey should succeed with special chars, got error: %v", err)
	}

	loadedKey := loadGLMKey()
	if loadedKey != testKey {
		t.Errorf("loaded key %q does not match saved key %q", loadedKey, testKey)
	}
}

func TestSaveGLMKey_EmptyKey(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	err := saveGLMKey("")
	if err != nil {
		t.Fatalf("saveGLMKey should succeed with empty key, got error: %v", err)
	}

	envPath := filepath.Join(tmpHome, ".moai", ".env.glm")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Fatal("expected .env.glm file to be created")
	}
}

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
			name:       "only High/Medium/Low set",
			models:     config.GLMModels{High: "custom-high", Medium: "custom-medium", Low: "custom-low"},
			wantHigh:   "custom-high",
			wantMedium: "custom-medium",
			wantLow:    "custom-low",
		},
		{
			name:       "only Opus/Sonnet/Haiku set",
			models:     config.GLMModels{Opus: "legacy-opus", Sonnet: "legacy-sonnet", Haiku: "legacy-haiku"},
			wantHigh:   "legacy-opus",
			wantMedium: "legacy-sonnet",
			wantLow:    "legacy-haiku",
		},
		{
			name:       "both set - High/Medium/Low priority",
			models:     config.GLMModels{High: "new-high", Medium: "new-medium", Low: "new-low", Opus: "old-opus", Sonnet: "old-sonnet", Haiku: "old-haiku"},
			wantHigh:   "new-high",
			wantMedium: "new-medium",
			wantLow:    "new-low",
		},
		{
			name:       "neither set - defaults",
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

func TestSaveGLMKey_OverwriteExisting(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	firstKey := "first-key"
	err := saveGLMKey(firstKey)
	if err != nil {
		t.Fatalf("first saveGLMKey failed: %v", err)
	}

	secondKey := "second-key"
	err = saveGLMKey(secondKey)
	if err != nil {
		t.Fatalf("second saveGLMKey failed: %v", err)
	}

	loadedKey := loadGLMKey()
	if loadedKey != secondKey {
		t.Errorf("loaded key %q, want %q", loadedKey, secondKey)
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "1234****6789"},
		{"sk-very-long-api-key-12345", "sk-v****2345"},
	}
	for _, tt := range tests {
		got := maskAPIKey(tt.key)
		if got != tt.want {
			t.Errorf("maskAPIKey(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

// TestInjectGLMEnvForTeam_StatuslineContextSize verifies Issue #742:
// when the GLM High slot maps to a known context window, the function persists
// MOAI_STATUSLINE_CONTEXT_SIZE in settings.local.json so the SessionStart hook
// can later propagate it through tmux env to the statusline.
func TestInjectGLMEnvForTeam_StatuslineContextSize(t *testing.T) {
	cases := []struct {
		name      string
		highModel string
		wantSize  string // "" means key should be absent
	}{
		{"glm-5.1 maps to 200K", "glm-5.1", "200000"},
		{"glm-4.6 maps to 128K", "glm-4.6", "128000"},
		{"glm-4.5-air maps to 128K (longest match)", "glm-4.5-air", "128000"},
		{"unknown model omits the hint", "totally-unknown-model", ""},
		{"claude prefix omits the hint (not GLM)", "claude-opus-4.7", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			claudeDir := filepath.Join(tmpDir, ".claude")
			if err := os.MkdirAll(claudeDir, 0o755); err != nil {
				t.Fatal(err)
			}
			settingsPath := filepath.Join(claudeDir, "settings.local.json")

			glmCfg := &GLMConfigFromYAML{
				BaseURL: "https://api.z.ai",
				EnvVar:  "TEST_KEY",
			}
			glmCfg.Models.High = tc.highModel
			glmCfg.Models.Medium = "ignored"
			glmCfg.Models.Low = "ignored"

			if err := injectGLMEnvForTeam(settingsPath, glmCfg, "test-api-key"); err != nil {
				t.Fatalf("injectGLMEnvForTeam error: %v", err)
			}

			data, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read settings: %v", err)
			}
			var settings SettingsLocal
			if err := json.Unmarshal(data, &settings); err != nil {
				t.Fatalf("parse settings: %v", err)
			}

			got, present := settings.Env["MOAI_STATUSLINE_CONTEXT_SIZE"]
			if tc.wantSize == "" {
				if present {
					t.Errorf("MOAI_STATUSLINE_CONTEXT_SIZE present=%q for unknown model %q (expected absent)",
						got, tc.highModel)
				}
				return
			}
			if !present {
				t.Errorf("MOAI_STATUSLINE_CONTEXT_SIZE absent for known model %q (expected %q)",
					tc.highModel, tc.wantSize)
			}
			if got != tc.wantSize {
				t.Errorf("MOAI_STATUSLINE_CONTEXT_SIZE = %q, want %q", got, tc.wantSize)
			}
		})
	}
}

// TestInjectGLMEnvForTeam_StatuslineContextSize_StaleCleanup verifies that
// switching the High slot from a known model to an unknown model removes the
// previously injected MOAI_STATUSLINE_CONTEXT_SIZE so the stale value is not
// propagated to a different GLM model.
func TestInjectGLMEnvForTeam_StatuslineContextSize_StaleCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	cfg := &GLMConfigFromYAML{BaseURL: "https://api.z.ai", EnvVar: "TEST_KEY"}

	// First call: glm-5.1 -> 200000 written.
	cfg.Models.High = "glm-5.1"
	cfg.Models.Medium = "ignored"
	cfg.Models.Low = "ignored"
	if err := injectGLMEnvForTeam(settingsPath, cfg, "key1"); err != nil {
		t.Fatalf("first inject: %v", err)
	}

	// Second call with an unknown model: stale 200000 must be cleared.
	cfg.Models.High = "unknown-future-model"
	if err := injectGLMEnvForTeam(settingsPath, cfg, "key2"); err != nil {
		t.Fatalf("second inject: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var settings SettingsLocal
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parse settings: %v", err)
	}
	if v, present := settings.Env["MOAI_STATUSLINE_CONTEXT_SIZE"]; present {
		t.Errorf("stale MOAI_STATUSLINE_CONTEXT_SIZE = %q after switch to unknown model (expected absent)", v)
	}
}

// TestRemoveGLMEnv_DropsStatuslineContextSize verifies that switching from GLM
// mode back to Claude mode (moai cc) clears the GLM-specific context size hint.
func TestRemoveGLMEnv_DropsStatuslineContextSize(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Seed settings.local.json with a GLM payload that includes the hint.
	seed := map[string]any{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN":         "glm-key",
			"ANTHROPIC_BASE_URL":           "https://api.z.ai",
			"ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-5.1",
			"MOAI_STATUSLINE_CONTEXT_SIZE": "200000",
		},
	}
	seedData, _ := json.MarshalIndent(seed, "", "  ")
	if err := os.WriteFile(settingsPath, seedData, 0o600); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	if err := removeGLMEnv(settingsPath); err != nil {
		t.Fatalf("removeGLMEnv error: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read after remove: %v", err)
	}
	var settings SettingsLocal
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parse after remove: %v", err)
	}
	if v, present := settings.Env["MOAI_STATUSLINE_CONTEXT_SIZE"]; present {
		t.Errorf("removeGLMEnv left MOAI_STATUSLINE_CONTEXT_SIZE = %q (expected absent)", v)
	}
}
