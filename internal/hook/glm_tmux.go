package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/tmux"
)

// glmTmuxKeys는 tmux 세션에 주입할 GLM 관련 환경변수 목록입니다.
// 새 tmux 팬(팀원)이 GLM API를 사용하려면 이 변수들이 세션 수준에서 설정되어야 합니다.
var glmTmuxKeys = []string{
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

// buildGLMTmuxEnvVars는 settings.local.json의 env 맵에서 GLM 관련 변수만 추출합니다.
//
// 규칙:
//   - ANTHROPIC_AUTH_TOKEN이 비어 있으면 nil 반환 (주입 불필요)
//   - glmTmuxKeys에 포함된 키만 결과에 포함됨
//   - 설정에 없는 키는 결과에 포함되지 않음
func buildGLMTmuxEnvVars(env map[string]string) map[string]string {
	if env["ANTHROPIC_AUTH_TOKEN"] == "" {
		return nil
	}

	result := make(map[string]string)
	for _, key := range glmTmuxKeys {
		if val, ok := env[key]; ok && val != "" {
			result[key] = val
		}
	}

	// ANTHROPIC_AUTH_TOKEN이 있으면 반드시 포함
	if token := env["ANTHROPIC_AUTH_TOKEN"]; token != "" {
		result["ANTHROPIC_AUTH_TOKEN"] = token
	}

	return result
}

// formatTmuxGLMEnvSummary는 tmux 세션 환경변수 주입 결과 요약 문자열을 반환합니다.
func formatTmuxGLMEnvSummary(n int) string {
	return fmt.Sprintf("GLM 팀원을 위해 tmux 세션 환경변수 주입 완료 (%d개 변수)", n)
}

// ensureTmuxGLMEnv는 SessionStart 훅에서 호출되어, GLM 모드에서 팀원 tmux 팬이
// ANTHROPIC_AUTH_TOKEN을 상속받을 수 있도록 현재 tmux 세션에 환경변수를 주입합니다.
//
// 동작:
//  1. TMUX 환경변수가 없으면 no-op (tmux 세션 밖)
//  2. settings.local.json에서 teammateMode가 "tmux"가 아니면 no-op
//  3. ANTHROPIC_AUTH_TOKEN이 없으면 no-op
//  4. GLM 관련 환경변수를 tmux set-environment로 현재 세션에 주입
//
// 에러 처리: 모든 에러는 경고 로그만 출력하고 빈 문자열을 반환합니다.
// SessionStart 훅 실패를 절대 일으키지 않습니다.
//
// @MX:ANCHOR: [AUTO] SessionStart 훅의 GLM+팀 모드 tmux 환경변수 주입 진입점
// @MX:REASON: Handle에서 직접 호출되며 session_start_glm_tmux_test.go에서 3개 이상 테스트 케이스가 검증
func ensureTmuxGLMEnv(projectDir string) string {
	// 1. tmux 세션 여부 확인
	if os.Getenv("TMUX") == "" {
		return ""
	}

	// 2. settings.local.json 읽기
	settingsPath := filepath.Join(projectDir, ".claude", "settings.local.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil || len(data) == 0 {
		return ""
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}

	// 3. teammateMode == "tmux" 확인
	var teammateMode string
	if v, ok := raw["teammateMode"]; ok {
		_ = json.Unmarshal(v, &teammateMode)
	}
	if teammateMode != "tmux" {
		return ""
	}

	// 4. env 맵 추출
	envRaw, ok := raw["env"]
	if !ok {
		return ""
	}
	var env map[string]string
	if err := json.Unmarshal(envRaw, &env); err != nil {
		return ""
	}

	// 5. GLM 환경변수 맵 구성 (AUTH_TOKEN 없으면 nil 반환)
	vars := buildGLMTmuxEnvVars(env)
	if len(vars) == 0 {
		return ""
	}

	// 6. tmux set-environment 실행
	mgr := tmux.NewSessionManager()
	if err := mgr.InjectEnv(context.Background(), vars); err != nil {
		slog.Warn("ensureTmuxGLMEnv: tmux 환경변수 주입 실패",
			"error", err.Error(),
			"hint", "tmux 바이너리가 없거나 세션 외부일 수 있음",
		)
		return ""
	}

	summary := formatTmuxGLMEnvSummary(len(vars))
	slog.Info("ensureTmuxGLMEnv: tmux 세션 GLM 환경변수 주입",
		"vars", len(vars),
	)
	return summary
}
