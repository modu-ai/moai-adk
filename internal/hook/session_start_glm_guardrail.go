// SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001 — GLM web-tooling 가드레일 훅.
//
// glm-web-tooling.md 규칙을 always-load 세트에서 제거하고, 그 라우팅 요약을
// GLM 백엔드 세션에서만 SessionStart 시점에 on-demand로 주입한다. 비-GLM 세션은
// 아무것도 로드하지 않고, GLM 세션은 작은 리마인더 하나만 받는다.
//
// 트리거는 PROCESS env `ANTHROPIC_BASE_URL`의 `z.ai` 부분문자열이며 — 이는
// cg_detect.go의 PROCESS-env GLM 검출기(`hasGLMEnv`)와 동일한 신호다 (REQ-GH-001).
// cg_detect.go의 tmux SESSION-env 검출기와 CG-모드 검출기는 cg-leader pane에서
// true를 반환하여 carve-out(REQ-GH-005/006)을 깨뜨리므로 절대 사용하지 않는다
// (D1 위험). 본 파일은 PROCESS env만 직접 읽으므로 cg-leader carve-out이
// 자동으로 만족된다.
package hook

import (
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// glmProcessEnvSubstring은 GLM 백엔드 신호 부분문자열이다. cg_detect.go
// `hasGLMEnv`가 매칭하는 `z.ai`와 동일하며 (REQ-GH-001 단일 신호 재사용),
// 정규 GLM 기본 URL(config.DefaultGLMBaseURL = "https://api.z.ai/api/anthropic")의
// host를 포함한다. `z.ai`는 `api.z.ai`의 strict superset이므로 모든 실 GLM
// 세션을 포착하며, advisory-only이므로 superset 엣지(D5)는 무해하다.
const glmProcessEnvSubstring = "z.ai"

// hookProcessEnvHasGLM은 PROCESS env `ANTHROPIC_BASE_URL`이 GLM 신호를
// 포함하는지 보고한다 (REQ-GH-001). cg_detect.go `hasGLMEnv`의 base-URL
// disjunct와 동일한 검사를 PROCESS env에 대해 수행한다. 본 함수는 PROCESS env만
// 읽으므로 cg-leader pane(PROCESS GLM env stripped)에서 false를 반환하여
// carve-out을 자동 만족시킨다 (REQ-GH-005/006, D1). os.Getenv는 부재 변수에
// 대해 ""를 반환하므로 읽기 실패가 panic이나 블로킹을 유발하지 않는다
// (REQ-GH-012, non-blocking).
//
// config.EnvAnthropicBaseURL 상수를 재사용하여 env 키 리터럴 하드코딩을 피한다
// (CLAUDE.local.md §14).
func hookProcessEnvHasGLM() bool {
	return strings.Contains(os.Getenv(config.EnvAnthropicBaseURL), glmProcessEnvSubstring)
}

// glmGuardrailReminderText는 GLM 백엔드에서 주입되는 간결한 라우팅 리마인더다.
// 7.5 KB 전체 규칙이 아닌 HARD 라우팅 테이블 요약만 담는다 (REQ-GH-004):
// 3개 z.ai MCP 교체(web search / web fetch / image read) + ToolSearch preload
// note + 전체 규칙 포인터.
const glmGuardrailReminderText = "[GLM backend detected] Route web tooling to z.ai MCP (not built-in):\n" +
	"WebSearch → mcp__web_search_prime__webSearchPrime\n" +
	"WebFetch  → mcp__web_reader__webReader\n" +
	"Read-on-image → mcp__zai-mcp-server__* (pass a local file path, not base64)\n" +
	"Preload deferred MCP schemas first: ToolSearch(query: \"select:<tool>\").\n" +
	"Full rule: .claude/rules/moai/core/glm-web-tooling.md"

// glmGuardrailReminder는 GLM 백엔드 세션이면 간결한 라우팅 리마인더 문자열을,
// 아니면 빈 문자열을 반환한다 (REQ-GH-002 / REQ-GH-003 / REQ-GH-004). 검출은
// PROCESS-env 신호(hookProcessEnvHasGLM)에만 의존하므로 cg-leader carve-out이
// 자동 만족된다 (REQ-GH-005/006). 검출 실패는 빈 문자열을 반환하며 절대
// 블로킹하지 않는다 (REQ-GH-012).
//
// 본 함수명은 AC-GH-011 커버리지 grep('glm.*[Gg]uardrail|[Gg]uardrail.*[Rr]eminder')에
// 매칭된다 (D8).
func glmGuardrailReminder() string {
	if !hookProcessEnvHasGLM() {
		return ""
	}
	return glmGuardrailReminderText
}
