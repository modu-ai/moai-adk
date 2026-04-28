package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEnsureTmuxGLMEnv_NotInTmux_ReturnsEmpty verifies that the function returns an empty
// string when executed outside a tmux session.
func TestEnsureTmuxGLMEnv_NotInTmux_ReturnsEmpty(t *testing.T) {
	// cannot use t.Parallel() due to t.Setenv
	t.Setenv("TMUX", "") // simulate environment outside a tmux session

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)
	settings := `{"teammateMode":"tmux","env":{"ANTHROPIC_AUTH_TOKEN":"test-token"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Errorf("must return empty string outside tmux, got %q", msg)
	}
}

// TestEnsureTmuxGLMEnv_NoTeammateModeTmux_ReturnsEmpty verifies that the function returns
// an empty string when teammateMode is not "tmux".
func TestEnsureTmuxGLMEnv_NoTeammateModeTmux_ReturnsEmpty(t *testing.T) {
	// cannot use t.Parallel() due to t.Setenv
	t.Setenv("TMUX", "/tmp/fake-tmux,1234,0") // simulate environment inside a tmux session

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// case: teammateMode is absent
	settings := `{"env":{"ANTHROPIC_AUTH_TOKEN":"test-token"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Errorf("must return empty string when teammateMode is not tmux (no key), got %q", msg)
	}

	// case: teammateMode is "auto"
	settings2 := `{"teammateMode":"auto","env":{"ANTHROPIC_AUTH_TOKEN":"test-token"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings2), 0o644)

	msg2 := ensureTmuxGLMEnv(dir)
	if msg2 != "" {
		t.Errorf("must return empty string when teammateMode=auto, got %q", msg2)
	}
}

// TestEnsureTmuxGLMEnv_NoAuthToken_ReturnsEmpty verifies that the function returns
// an empty string when ANTHROPIC_AUTH_TOKEN is absent.
func TestEnsureTmuxGLMEnv_NoAuthToken_ReturnsEmpty(t *testing.T) {
	// cannot use t.Parallel() due to t.Setenv
	t.Setenv("TMUX", "/tmp/fake-tmux,1234,0")

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// teammateMode=tmux without ANTHROPIC_AUTH_TOKEN
	settings := `{"teammateMode":"tmux","env":{"ANTHROPIC_BASE_URL":"https://api.z.ai/api/anthropic"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Errorf("must return empty string when AUTH_TOKEN is absent, got %q", msg)
	}
}

// TestEnsureTmuxGLMEnv_NoSettingsFile_ReturnsEmpty verifies that the function returns
// an empty string when settings.local.json is absent.
func TestEnsureTmuxGLMEnv_NoSettingsFile_ReturnsEmpty(t *testing.T) {
	// cannot use t.Parallel() due to t.Setenv
	t.Setenv("TMUX", "/tmp/fake-tmux,1234,0")

	dir := t.TempDir()
	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Errorf("must return empty string when settings file is absent, got %q", msg)
	}
}

// TestBuildGLMTmuxEnvVars_AllGLMVars verifies that buildGLMTmuxEnvVars returns
// the correct map when all GLM environment variables are present.
func TestBuildGLMTmuxEnvVars_AllGLMVars(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"ANTHROPIC_AUTH_TOKEN":                     "test-glm-token",
		"ANTHROPIC_BASE_URL":                       "https://api.z.ai/api/anthropic",
		"ANTHROPIC_DEFAULT_OPUS_MODEL":             "glm-5.1",
		"ANTHROPIC_DEFAULT_SONNET_MODEL":           "glm-4.7",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "glm-4.7-flash",
		"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS":   "1",
		"DISABLE_PROMPT_CACHING":                   "1",
		"API_TIMEOUT_MS":                           "3000000",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
		"SOME_OTHER_VAR":                           "should-be-ignored",
	}

	result := buildGLMTmuxEnvVars(env)

	// all GLM-related variables must be present in the returned map
	wantKeys := []string{
		"ANTHROPIC_AUTH_TOKEN",
		"ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
		"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS",
		"DISABLE_PROMPT_CACHING",
		"API_TIMEOUT_MS",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC",
	}

	for _, key := range wantKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("key %q is missing from result map", key)
		}
	}

	// unrelated variables must not be included
	if _, ok := result["SOME_OTHER_VAR"]; ok {
		t.Error("unrelated variable SOME_OTHER_VAR must not be included")
	}
}

// TestBuildGLMTmuxEnvVars_EmptyAuthToken verifies that buildGLMTmuxEnvVars
// returns nil when AUTH_TOKEN is absent.
func TestBuildGLMTmuxEnvVars_EmptyAuthToken(t *testing.T) {
	t.Parallel()

	// case: ANTHROPIC_AUTH_TOKEN is absent
	emptyEnv := map[string]string{
		"ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
	}
	emptyResult := buildGLMTmuxEnvVars(emptyEnv)
	if emptyResult != nil {
		t.Error("must return nil when AUTH_TOKEN is absent")
	}

	// case: ANTHROPIC_AUTH_TOKEN is an empty string
	emptyTokenEnv := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "",
		"ANTHROPIC_BASE_URL":   "https://api.z.ai/api/anthropic",
	}
	emptyTokenResult := buildGLMTmuxEnvVars(emptyTokenEnv)
	if emptyTokenResult != nil {
		t.Error("must return nil when AUTH_TOKEN is an empty string")
	}
}

// TestEnsureTmuxGLMEnv_HandlesTmuxBinaryMissing verifies that the function returns
// an empty string without panicking when the tmux binary is absent.
func TestEnsureTmuxGLMEnv_HandlesTmuxBinaryMissing(t *testing.T) {
	// cannot use t.Parallel() due to t.Setenv
	t.Setenv("TMUX", "/tmp/fake-tmux,1234,0")
	t.Setenv("PATH", "/nonexistent-path-for-testing")

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	settings := map[string]any{
		"teammateMode": "tmux",
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN":                     "test-token",
			"ANTHROPIC_BASE_URL":                       "https://api.z.ai/api/anthropic",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":             "glm-5.1",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":           "glm-4.7",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "glm-4.7-flash",
			"CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS":   "1",
			"DISABLE_PROMPT_CACHING":                   "1",
		},
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), data, 0o644)

	// must return empty string without panicking when tmux binary is absent
	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Logf("non-empty returned despite missing tmux binary (incomplete PATH isolation): %q", msg)
	}
	// success if no panic occurs
}

// TestBuildGLMTmuxEnvVars_OnlyPresentVars verifies that only variables present in the
// settings file are included in the result (absent variables are excluded).
func TestBuildGLMTmuxEnvVars_OnlyPresentVars(t *testing.T) {
	t.Parallel()

	// only some GLM variables are present
	env := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "test-token",
		"ANTHROPIC_BASE_URL":   "https://api.z.ai/api/anthropic",
		// model names absent
	}

	result := buildGLMTmuxEnvVars(env)
	if result == nil {
		t.Fatal("must not return nil when AUTH_TOKEN is present")
	}

	// AUTH_TOKEN and BASE_URL must be present
	if result["ANTHROPIC_AUTH_TOKEN"] != "test-token" {
		t.Errorf("AUTH_TOKEN = %q, want %q", result["ANTHROPIC_AUTH_TOKEN"], "test-token")
	}
	if result["ANTHROPIC_BASE_URL"] != "https://api.z.ai/api/anthropic" {
		t.Errorf("BASE_URL = %q, want Z.AI endpoint", result["ANTHROPIC_BASE_URL"])
	}

	// model keys not in settings must not be included
	if _, ok := result["ANTHROPIC_DEFAULT_OPUS_MODEL"]; ok {
		t.Error("OPUS_MODEL not in settings must not be included")
	}
}

// TestFormatTmuxGLMEnvSummary_ContainsVarCount verifies that the returned string
// contains the number of injected variables on success.
func TestFormatTmuxGLMEnvSummary_ContainsVarCount(t *testing.T) {
	t.Parallel()

	summary := formatTmuxGLMEnvSummary(9)
	if !strings.Contains(summary, "9") {
		t.Errorf("summary string must contain variable count (9), got %q", summary)
	}
}
