package hook

import (
	"context"
	"strings"
	"testing"
)

// SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001 — GLM web-tooling guardrail hook tests.
//
// 이 테스트들은 SessionStart 핸들러가 GLM 백엔드 세션에서만 z.ai MCP 라우팅
// 리마인더를 주입하는지 검증한다. 트리거는 PROCESS env `ANTHROPIC_BASE_URL`의
// `z.ai` 부분문자열이며 (cg_detect.go `hasGLMEnv`와 동일 신호), tmux SESSION env
// 검출기(`sessionEnvHasGLM`)나 `IsCGMode()`를 사용하지 않는다 — 그 둘은 cg-leader
// pane에서 true를 반환하여 carve-out을 깨뜨린다 (D1 위험, REQ-GH-001/005/006).
//
// t.Setenv를 쓰므로 이 테스트들은 non-parallel이다 (OTEL/env 경쟁 회피,
// CLAUDE.local.md §6 / §13).

// glmReminderMarkers는 주입된 리마인더가 반드시 포함해야 하는 3개 z.ai MCP
// 교체 토큰 + ToolSearch preload 토큰이다 (REQ-GH-004 / AC-GH-008).
var glmReminderMarkers = []string{
	"web_search_prime",
	"web_reader",
	"zai-mcp-server",
	"ToolSearch",
}

// setProcessGLMEnv는 PROCESS env에 GLM 백엔드 신호를 심는다.
func setProcessGLMEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ANTHROPIC_BASE_URL", "https://api.z.ai/api/anthropic")
}

// clearProcessGLMEnv는 PROCESS env에서 GLM 신호를 제거한다 (Claude 백엔드 모사).
func clearProcessGLMEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ANTHROPIC_BASE_URL", "")
}

// TestGLMGuardrailReminder_GLMSession은 z.ai PROCESS env가 설정되면 리마인더가
// 4개 마커를 모두 포함하는지 검증한다 (REQ-GH-002 / REQ-GH-004 / AC-GH-008).
func TestGLMGuardrailReminder_GLMSession(t *testing.T) {
	setProcessGLMEnv(t)

	got := glmGuardrailReminder()
	if got == "" {
		t.Fatal("glmGuardrailReminder() returned empty for GLM session; expected reminder")
	}
	for _, marker := range glmReminderMarkers {
		if !strings.Contains(got, marker) {
			t.Errorf("reminder missing required marker %q\nreminder:\n%s", marker, got)
		}
	}
}

// TestGLMGuardrailReminder_NonGLMSession은 z.ai가 없는 PROCESS env에서 빈
// 문자열을 반환하는지 검증한다 (REQ-GH-003 / AC-GH-009).
func TestGLMGuardrailReminder_NonGLMSession(t *testing.T) {
	clearProcessGLMEnv(t)

	if got := glmGuardrailReminder(); got != "" {
		t.Errorf("glmGuardrailReminder() = %q for non-GLM session; want empty", got)
	}
}

// TestGLMGuardrailReminder_CgLeader은 cg-leader pane carve-out을 검증한다 (D1).
// cg-leader는 PROCESS GLM env가 stripped되어 있으나 tmux SESSION env는 여전히
// z.ai를 carry한다. 이 테스트는 PROCESS env만 깨끗하게(z.ai 없이) 두고 SESSION
// GLM 마커를 설정한다 — `sessionEnvHasGLM`/`IsCGMode`에 gate했다면 실패한다
// (REQ-GH-005 / REQ-GH-006 / AC-GH-009b).
func TestGLMGuardrailReminder_CgLeader(t *testing.T) {
	// PROCESS env는 깨끗 (z.ai 없음) — leader pane이 PROCESS GLM env를 strip한 상태.
	clearProcessGLMEnv(t)
	// tmux SESSION GLM 마커 모사 — TMUX가 설정되고 SESSION env가 z.ai를 carry.
	// 만약 hook이 sessionEnvHasGLM/IsCGMode에 gate한다면 이 SESSION 신호로 인해
	// 잘못 주입하게 된다. PROCESS-env 검출기를 쓰면 carve-out이 자동 만족된다.
	t.Setenv("TMUX", "/private/tmp/tmux-501/default,12345,0")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "glm-session-token")

	if got := glmGuardrailReminder(); got != "" {
		t.Errorf("glmGuardrailReminder() = %q on cg-leader (PROCESS env clean); want empty (carve-out)", got)
	}
}

// TestGLMGuardrailReminder_GLMNonAPIHost은 D5 superset 엣지를 검증한다:
// z.ai를 포함하나 api.z.ai는 아닌 host도 부분문자열 매칭으로 GLM 취급된다
// (advisory-only이므로 false-positive는 무해, AC §B edge).
func TestGLMGuardrailReminder_GLMNonAPIHost(t *testing.T) {
	t.Setenv("ANTHROPIC_BASE_URL", "https://gateway.z.ai/custom")

	if got := glmGuardrailReminder(); got == "" {
		t.Error("glmGuardrailReminder() returned empty for a z.ai-substring host; want reminder (substring match, D5)")
	}
}

// TestSessionStartHandler_Handle_GLMGuardrailInjected는 통합 검증이다:
// GLM 백엔드 세션에서 Handle()의 HookSpecificOutput.AdditionalContext가 GLM
// 리마인더 마커들을 포함하는지 확인한다 (REQ-GH-002, end-to-end).
func TestSessionStartHandler_Handle_GLMGuardrailInjected(t *testing.T) {
	setProcessGLMEnv(t)

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	h := NewSessionStartHandler(cfg)

	input := &HookInput{
		SessionID:     "sess-glm-guardrail",
		CWD:           t.TempDir(),
		HookEventName: "SessionStart",
		ProjectDir:    t.TempDir(),
	}

	got, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("expected HookSpecificOutput to be set on GLM session")
	}
	ctx := got.HookSpecificOutput.AdditionalContext
	for _, marker := range glmReminderMarkers {
		if !strings.Contains(ctx, marker) {
			t.Errorf("AdditionalContext missing GLM marker %q\ncontext:\n%s", marker, ctx)
		}
	}
}

// TestSessionStartHandler_Handle_NonGLMNoGuardrail는 비-GLM 세션에서 GLM
// 리마인더가 AdditionalContext에 들어가지 않는지 검증한다 (REQ-GH-003).
// (UUID attribution context는 여전히 주입되나 GLM 마커는 없어야 한다.)
func TestSessionStartHandler_Handle_NonGLMNoGuardrail(t *testing.T) {
	clearProcessGLMEnv(t)

	cfg := &mockConfigProvider{cfg: newTestConfig()}
	h := NewSessionStartHandler(cfg)

	input := &HookInput{
		SessionID:     "sess-claude-backend",
		CWD:           t.TempDir(),
		HookEventName: "SessionStart",
		ProjectDir:    t.TempDir(),
	}

	got, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.HookSpecificOutput == nil {
		t.Fatal("expected HookSpecificOutput (UUID attribution) to be set")
	}
	ctx := got.HookSpecificOutput.AdditionalContext
	// GLM 전용 마커 중 하나라도 있으면 비-GLM 세션에 잘못 주입된 것.
	for _, marker := range []string{"web_search_prime", "web_reader", "zai-mcp-server"} {
		if strings.Contains(ctx, marker) {
			t.Errorf("non-GLM session leaked GLM marker %q into AdditionalContext:\n%s", marker, ctx)
		}
	}
}

// TestGLMGuardrailReminder_AllowOnError은 REQ-GH-012를 검증한다: 검출은 절대
// 블로킹하지 않는다. glmGuardrailReminder는 env 읽기 실패 시에도 (빈 env →
// 빈 문자열) 안전하게 빈 문자열을 반환하며 panic하지 않는다.
func TestGLMGuardrailReminder_AllowOnError(t *testing.T) {
	// ANTHROPIC_BASE_URL을 명시적으로 빈 값으로 설정 (읽기 실패/부재 모사).
	t.Setenv("ANTHROPIC_BASE_URL", "")

	// panic 없이 빈 문자열을 반환해야 한다 (non-blocking, allow).
	if got := glmGuardrailReminder(); got != "" {
		t.Errorf("glmGuardrailReminder() = %q on absent env; want empty (non-blocking allow)", got)
	}
}
