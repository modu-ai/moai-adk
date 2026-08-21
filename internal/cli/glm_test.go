package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
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
			gotHigh, gotMedium, gotLow, _ := resolveGLMModels(tt.models)
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

// TestGLMReasoningEnvVarsForEffort verifies the prefs-derived GLM reasoning
// env injection for the main session. It is the main-session wire point that
// derives ANTHROPIC_REASONING_EFFORT from the web-set effort (z.ai honors
// reasoning_effort, NOT Claude's 5-step CLAUDE_CODE_EFFORT_LEVEL). Distinct from
// glmReasoningEnvVars() (the hardcoded session default used for
// sub-agents / empty-effort fallback), this helper carries prefs.EffortLevel
// through to z.ai.
func TestGLMReasoningEnvVarsForEffort(t *testing.T) {
	cases := []struct {
		name          string
		effort        string
		wantPresent   bool
		wantReasoning string
	}{
		{"low → reasoning low (glm-5.3 cannot disable thinking)", template.EffortLevelLow, true, template.GLMReasoningEffortLow},
		{"medium → reasoning max", template.EffortLevelMedium, true, template.GLMReasoningEffortMax},
		{"high → reasoning max", template.EffortLevelHigh, true, template.GLMReasoningEffortMax},
		{"xhigh → reasoning max", template.EffortLevelXHigh, true, template.GLMReasoningEffortMax},
		{"max → reasoning max", template.EffortLevelMax, true, template.GLMReasoningEffortMax},
		{"empty → session default (reasoning max)", "", true, template.GLMReasoningEffortMax},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := glmReasoningEnvVarsForEffort(tc.effort)
			v, present := got[config.EnvAnthropicReasoningEffort]
			if present != tc.wantPresent {
				t.Errorf("glmReasoningEnvVarsForEffort(%q): %s present=%v, want %v (got=%v)",
					tc.effort, config.EnvAnthropicReasoningEffort, present, tc.wantPresent, got)
				return
			}
			if tc.wantPresent && v != tc.wantReasoning {
				t.Errorf("glmReasoningEnvVarsForEffort(%q)[%s] = %q, want %q",
					tc.effort, config.EnvAnthropicReasoningEffort, v, tc.wantReasoning)
			}
		})
	}
}

// ── SPEC-FACTORY-MODE-001 M5 ──

// TestGLM_KanbanFlagParity is AC-FM-005: `moai glm --kanban` reaches the
// launcher in glm mode with the kanban signal published, confirming parity
// with `moai cc`. Both launchers are single-backend, so both support the mode.
func TestGLM_KanbanFlagParity(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedMode string
	var capturedArgs []string
	var kanbanAtLaunch, specAtLaunch string
	unifiedLaunchFunc = func(_ string, mode string, args []string) error {
		capturedMode = mode
		capturedArgs = args
		kanbanAtLaunch = os.Getenv(config.EnvMoaiKanban)
		specAtLaunch = os.Getenv(config.EnvMoaiKanbanSpec)
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	buf := new(bytes.Buffer)
	glmCmd.SetOut(buf)
	glmCmd.SetErr(buf)

	if err := runGLM(glmCmd, []string{"--kanban", "SPEC-PLACEHOLDER"}); err != nil {
		t.Fatalf("AC-FM-005: runGLM(--kanban) should not error, got: %v", err)
	}
	if capturedMode != "glm" {
		t.Errorf("AC-FM-005: mode = %q, want %q", capturedMode, "glm")
	}
	for _, a := range capturedArgs {
		if a == "--kanban" || a == "-k" {
			t.Errorf("AC-FM-005: kanban token must not reach the launcher, got %v", capturedArgs)
		}
	}
	if kanbanAtLaunch != "1" {
		t.Errorf("AC-FM-005: %s must be set at launch, got %q", config.EnvMoaiKanban, kanbanAtLaunch)
	}
	if specAtLaunch != "SPEC-PLACEHOLDER" {
		t.Errorf("AC-FM-005: %s must carry the identifier at launch, got %q", config.EnvMoaiKanbanSpec, specAtLaunch)
	}
}
