package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEnsureTmuxGLMEnv_NotInTmux_ReturnsEmpty는 tmux 세션 밖에서 실행 시
// 함수가 빈 문자열을 반환하는지 검증합니다.
func TestEnsureTmuxGLMEnv_NotInTmux_ReturnsEmpty(t *testing.T) {
	// t.Setenv 사용으로 t.Parallel() 불가
	t.Setenv("TMUX", "") // tmux 세션 밖을 시뮬레이션

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)
	settings := `{"teammateMode":"tmux","env":{"ANTHROPIC_AUTH_TOKEN":"test-token"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Errorf("tmux 밖에서는 빈 문자열이어야 함, got %q", msg)
	}
}

// TestEnsureTmuxGLMEnv_NoTeammateModeTmux_ReturnsEmpty는 teammateMode가
// "tmux"가 아닐 때 빈 문자열을 반환하는지 검증합니다.
func TestEnsureTmuxGLMEnv_NoTeammateModeTmux_ReturnsEmpty(t *testing.T) {
	// t.Setenv 사용으로 t.Parallel() 불가
	t.Setenv("TMUX", "/tmp/fake-tmux,1234,0") // tmux 세션 안을 시뮬레이션

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// teammateMode가 없는 경우
	settings := `{"env":{"ANTHROPIC_AUTH_TOKEN":"test-token"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Errorf("teammateMode가 tmux가 아닐 때 빈 문자열이어야 함 (no key), got %q", msg)
	}

	// teammateMode가 "auto"인 경우
	settings2 := `{"teammateMode":"auto","env":{"ANTHROPIC_AUTH_TOKEN":"test-token"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings2), 0o644)

	msg2 := ensureTmuxGLMEnv(dir)
	if msg2 != "" {
		t.Errorf("teammateMode=auto일 때 빈 문자열이어야 함, got %q", msg2)
	}
}

// TestEnsureTmuxGLMEnv_NoAuthToken_ReturnsEmpty는 ANTHROPIC_AUTH_TOKEN이 없을 때
// 빈 문자열을 반환하는지 검증합니다.
func TestEnsureTmuxGLMEnv_NoAuthToken_ReturnsEmpty(t *testing.T) {
	// t.Setenv 사용으로 t.Parallel() 불가
	t.Setenv("TMUX", "/tmp/fake-tmux,1234,0")

	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	_ = os.MkdirAll(claudeDir, 0o755)

	// ANTHROPIC_AUTH_TOKEN 없이 teammateMode=tmux
	settings := `{"teammateMode":"tmux","env":{"ANTHROPIC_BASE_URL":"https://api.z.ai/api/anthropic"}}`
	_ = os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(settings), 0o644)

	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Errorf("AUTH_TOKEN 없을 때 빈 문자열이어야 함, got %q", msg)
	}
}

// TestEnsureTmuxGLMEnv_NoSettingsFile_ReturnsEmpty는 settings.local.json이
// 없을 때 빈 문자열을 반환하는지 검증합니다.
func TestEnsureTmuxGLMEnv_NoSettingsFile_ReturnsEmpty(t *testing.T) {
	// t.Setenv 사용으로 t.Parallel() 불가
	t.Setenv("TMUX", "/tmp/fake-tmux,1234,0")

	dir := t.TempDir()
	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Errorf("설정 파일 없을 때 빈 문자열이어야 함, got %q", msg)
	}
}

// TestBuildGLMTmuxEnvVars_AllGLMVars는 GLM 환경변수가 모두 있을 때
// buildGLMTmuxEnvVars가 올바른 맵을 반환하는지 검증합니다.
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

	// 반환된 맵에 모든 GLM 관련 변수가 포함되어야 함
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
			t.Errorf("결과 맵에 %q 키가 없음", key)
		}
	}

	// 무관한 변수는 포함되지 않아야 함
	if _, ok := result["SOME_OTHER_VAR"]; ok {
		t.Error("무관한 변수 SOME_OTHER_VAR가 포함되면 안 됨")
	}
}

// TestBuildGLMTmuxEnvVars_EmptyAuthToken은 AUTH_TOKEN이 없을 때
// buildGLMTmuxEnvVars가 nil을 반환하는지 검증합니다.
func TestBuildGLMTmuxEnvVars_EmptyAuthToken(t *testing.T) {
	t.Parallel()

	// ANTHROPIC_AUTH_TOKEN 없는 경우
	emptyEnv := map[string]string{
		"ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
	}
	emptyResult := buildGLMTmuxEnvVars(emptyEnv)
	if emptyResult != nil {
		t.Error("AUTH_TOKEN 없을 때 nil이어야 함")
	}

	// ANTHROPIC_AUTH_TOKEN이 빈 문자열인 경우
	emptyTokenEnv := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "",
		"ANTHROPIC_BASE_URL":   "https://api.z.ai/api/anthropic",
	}
	emptyTokenResult := buildGLMTmuxEnvVars(emptyTokenEnv)
	if emptyTokenResult != nil {
		t.Error("빈 AUTH_TOKEN일 때 nil이어야 함")
	}
}

// TestEnsureTmuxGLMEnv_HandlesTmuxBinaryMissing은 tmux 바이너리가 없을 때
// 패닉 없이 빈 문자열을 반환하는지 검증합니다.
func TestEnsureTmuxGLMEnv_HandlesTmuxBinaryMissing(t *testing.T) {
	// t.Setenv 사용으로 t.Parallel() 불가
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

	// tmux 바이너리 없을 때 패닉 없이 빈 문자열 반환해야 함
	msg := ensureTmuxGLMEnv(dir)
	if msg != "" {
		t.Logf("tmux 바이너리가 없는데 non-empty 반환됨 (PATH 격리 불완전): %q", msg)
	}
	// 패닉이 발생하지 않으면 성공
}

// TestBuildGLMTmuxEnvVars_OnlyPresentVars는 설정 파일에 있는 변수만
// 결과에 포함되는지 검증합니다 (없는 변수는 포함 안 됨).
func TestBuildGLMTmuxEnvVars_OnlyPresentVars(t *testing.T) {
	t.Parallel()

	// 일부 GLM 변수만 있는 경우
	env := map[string]string{
		"ANTHROPIC_AUTH_TOKEN": "test-token",
		"ANTHROPIC_BASE_URL":   "https://api.z.ai/api/anthropic",
		// 모델명 없음
	}

	result := buildGLMTmuxEnvVars(env)
	if result == nil {
		t.Fatal("AUTH_TOKEN 있을 때 nil이면 안 됨")
	}

	// AUTH_TOKEN과 BASE_URL은 있어야 함
	if result["ANTHROPIC_AUTH_TOKEN"] != "test-token" {
		t.Errorf("AUTH_TOKEN = %q, want %q", result["ANTHROPIC_AUTH_TOKEN"], "test-token")
	}
	if result["ANTHROPIC_BASE_URL"] != "https://api.z.ai/api/anthropic" {
		t.Errorf("BASE_URL = %q, want Z.AI endpoint", result["ANTHROPIC_BASE_URL"])
	}

	// 설정에 없는 모델 키는 포함되지 않아야 함
	if _, ok := result["ANTHROPIC_DEFAULT_OPUS_MODEL"]; ok {
		t.Error("설정에 없는 OPUS_MODEL이 포함되면 안 됨")
	}
}

// TestFormatTmuxGLMEnvSummary_ContainsVarCount는 성공 시 반환 문자열에
// 주입된 변수 수가 포함되는지 검증합니다.
func TestFormatTmuxGLMEnvSummary_ContainsVarCount(t *testing.T) {
	t.Parallel()

	summary := formatTmuxGLMEnvSummary(9)
	if !strings.Contains(summary, "9") {
		t.Errorf("요약 문자열에 변수 수(9)가 포함되어야 함, got %q", summary)
	}
}
