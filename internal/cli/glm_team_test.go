package cli

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestBuildGLMEnvVars verifies that buildGLMEnvVars produces the correct
// map of environment variables for GLM hybrid mode.
func TestBuildGLMEnvVars(t *testing.T) {
	tests := []struct {
		name      string
		glmConfig *GLMConfigFromYAML
		apiKey    string
		wantKeys  []string
		wantVals  map[string]string
	}{
		{
			name: "default_glm_config",
			glmConfig: &GLMConfigFromYAML{
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
			},
			apiKey: "test-key-123",
			wantKeys: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
			},
			wantVals: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "test-key-123",
				"ANTHROPIC_BASE_URL":             "https://api.z.ai/api/anthropic",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-flashx",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
			},
		},
		{
			name: "custom_config",
			glmConfig: &GLMConfigFromYAML{
				BaseURL: "https://custom.glm.api/v1",
				Models: struct {
					Haiku  string
					Sonnet string
					Opus   string
				}{
					Haiku:  "custom-haiku",
					Sonnet: "custom-sonnet",
					Opus:   "custom-opus",
				},
				EnvVar: "CUSTOM_API_KEY",
			},
			apiKey: "custom-api-key-xyz",
			wantKeys: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
			},
			wantVals: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "custom-api-key-xyz",
				"ANTHROPIC_BASE_URL":             "https://custom.glm.api/v1",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "custom-haiku",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "custom-sonnet",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "custom-opus",
			},
		},
		{
			name: "empty_api_key",
			glmConfig: &GLMConfigFromYAML{
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
			},
			apiKey: "",
			wantKeys: []string{
				"ANTHROPIC_AUTH_TOKEN",
				"ANTHROPIC_BASE_URL",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL",
				"ANTHROPIC_DEFAULT_SONNET_MODEL",
				"ANTHROPIC_DEFAULT_OPUS_MODEL",
			},
			wantVals: map[string]string{
				"ANTHROPIC_AUTH_TOKEN":           "",
				"ANTHROPIC_BASE_URL":             "https://api.z.ai/api/anthropic",
				"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-4.7-flashx",
				"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildGLMEnvVars(tt.glmConfig, tt.apiKey)

			// Verify the map has exactly 5 keys
			if len(got) != 5 {
				t.Errorf("buildGLMEnvVars() returned %d keys, want 5", len(got))
			}

			// Verify all required keys exist
			for _, key := range tt.wantKeys {
				if _, ok := got[key]; !ok {
					t.Errorf("buildGLMEnvVars() missing key %q", key)
				}
			}

			// Verify specific values
			for key, wantVal := range tt.wantVals {
				if gotVal, ok := got[key]; !ok {
					t.Errorf("buildGLMEnvVars() missing key %q", key)
				} else if gotVal != wantVal {
					t.Errorf("buildGLMEnvVars()[%q] = %q, want %q", key, gotVal, wantVal)
				}
			}
		})
	}
}

// TestCheckTmuxAvailable verifies that checkTmuxAvailable returns an error
// when tmux is not in PATH.
func TestCheckTmuxAvailable(t *testing.T) {
	tests := []struct {
		name        string
		pathVal     string
		wantErr     bool
		errContains string
	}{
		{
			name:        "tmux_not_in_path",
			pathVal:     t.TempDir(), // empty directory with no tmux binary
			wantErr:     true,
			errContains: "tmux is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override PATH to a directory that doesn't contain tmux
			t.Setenv("PATH", tt.pathVal)

			err := checkTmuxAvailable()

			if tt.wantErr {
				if err == nil {
					t.Error("checkTmuxAvailable() should return error when tmux not in PATH")
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("checkTmuxAvailable() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("checkTmuxAvailable() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestCheckTmuxAvailable_TmuxPresent verifies checkTmuxAvailable succeeds
// when tmux is installed. Skips if tmux is not available on the test machine.
func TestCheckTmuxAvailable_TmuxPresent(t *testing.T) {
	// Check if tmux is actually installed before testing this path
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux is not installed on this machine, skipping tmux-present test")
	}

	err := checkTmuxAvailable()
	if err != nil {
		t.Errorf("checkTmuxAvailable() should succeed when tmux is installed, got: %v", err)
	}
}

// TestGLMTeamFlag verifies that the --team flag is correctly registered
// on the glm command with the expected default value.
func TestGLMTeamFlag(t *testing.T) {
	// Verify the --team flag exists on glmCmd
	flag := glmCmd.Flags().Lookup("team")
	if flag == nil {
		t.Fatal("--team flag should be registered on glmCmd")
	}

	// Verify the flag type is bool
	if flag.Value.Type() != "bool" {
		t.Errorf("--team flag type = %q, want %q", flag.Value.Type(), "bool")
	}

	// Verify the default value is false
	if flag.DefValue != "false" {
		t.Errorf("--team flag default = %q, want %q", flag.DefValue, "false")
	}

	// Verify the flag usage text mentions tmux
	if !strings.Contains(strings.ToLower(flag.Usage), "tmux") {
		t.Errorf("--team flag usage should mention tmux, got: %q", flag.Usage)
	}
}

// TestRunGLMTeam_NoTmux verifies that runGLMTeam returns an error
// containing the expected message when tmux is not found in PATH.
func TestRunGLMTeam_NoTmux(t *testing.T) {
	t.Setenv("MOAI_TEST_MODE", "1")

	// Override PATH to a directory that does not contain tmux
	emptyDir := t.TempDir()
	t.Setenv("PATH", emptyDir)

	buf := new(bytes.Buffer)
	cmd := glmCmd
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := runGLMTeam(cmd)
	if err == nil {
		t.Fatal("runGLMTeam() should return error when tmux is not available")
	}

	if !strings.Contains(err.Error(), "tmux") {
		t.Errorf("runGLMTeam() error should mention tmux, got: %q", err.Error())
	}
}

// TestRunGLMTeam_NoAPIKey verifies that runGLMTeam returns an error
// when tmux is available but no API key is configured.
// Skips if tmux is not installed on the test machine.
func TestRunGLMTeam_NoAPIKey(t *testing.T) {
	// Skip if tmux is not available
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux is not installed on this machine, skipping API key test")
	}

	t.Setenv("MOAI_TEST_MODE", "1")

	// Use a temp directory as the home dir so no saved key is found
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	// Clear any API key environment variable
	t.Setenv("GLM_API_KEY", "")

	// Create a temporary project directory so findProjectRoot succeeds
	projectRoot := t.TempDir()
	if err := os.MkdirAll(projectRoot+"/.moai", 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}

	// Set deps to nil so loadGLMConfig uses fallback (which reads GLM_API_KEY)
	origDeps := deps
	deps = nil
	defer func() { deps = origDeps }()

	buf := new(bytes.Buffer)
	cmd := glmCmd
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := runGLMTeam(cmd)
	if err == nil {
		t.Fatal("runGLMTeam() should return error when no API key is configured")
	}

	if !strings.Contains(err.Error(), "GLM API key not found") {
		t.Errorf("runGLMTeam() error should mention missing API key, got: %q", err.Error())
	}
}
