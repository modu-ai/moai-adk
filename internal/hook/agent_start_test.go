// agent_start_test.go: SubagentStart hook의 retired-agent 거부 로직 검증.
// REQ-RA-004, REQ-RA-007, REQ-RA-008, REQ-RA-009, REQ-RA-012 매핑.
//
// M3 GREEN-2 phase: agentStartHandler가 subagent_start.go에 구현되었으므로
// 모든 t.Skip이 실제 assertion으로 전환된다.
package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// agentStartHandlerExists는 AgentStartHandler 구조체가 이 패키지에 존재하는지
// 컴파일 타임에 확인하기 위한 타입 단언 헬퍼 인터페이스.
type agentStartInterface interface {
	Handle(ctx context.Context, input *HookInput) (*HookOutput, error)
	EventType() EventType
}

// buildRetiredAgentDir는 테스트용 임시 디렉터리를 생성하고
// .claude/agents/moai/<agentName>.md 파일을 retired frontmatter와 함께 작성한다.
func buildRetiredAgentDir(t *testing.T, agentName, replacement, paramHint string) string {
	t.Helper()

	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".claude", "agents", "moai")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("임시 에이전트 디렉터리 생성 실패: %v", err)
	}

	// retired stub frontmatter 표준화 (REQ-RA-002)
	content := strings.Join([]string{
		"---",
		"name: " + agentName,
		"description: \"Retired — use " + replacement + "\"",
		"retired: true",
		"retired_replacement: " + replacement,
		"retired_param_hint: \"" + paramHint + "\"",
		"tools: []",
		"skills: []",
		"---",
		"",
		"# " + agentName + " (은퇴됨)",
		"",
		"이 에이전트는 은퇴했습니다. `" + replacement + "`를 사용하세요.",
		"",
		"마이그레이션 명령: Agent({subagent_type: \"" + replacement + "\", " + paramHint + "})",
	}, "\n")

	agentFile := filepath.Join(agentDir, agentName+".md")
	if err := os.WriteFile(agentFile, []byte(content), 0o644); err != nil {
		t.Fatalf("에이전트 파일 쓰기 실패: %v", err)
	}

	return dir
}

// buildActiveAgentDir는 테스트용 임시 디렉터리를 생성하고
// .claude/agents/moai/<agentName>.md 파일을 활성 frontmatter와 함께 작성한다.
func buildActiveAgentDir(t *testing.T, agentName string) string {
	t.Helper()

	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".claude", "agents", "moai")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("임시 에이전트 디렉터리 생성 실패: %v", err)
	}

	content := strings.Join([]string{
		"---",
		"name: " + agentName,
		"description: \"Unified DDD/TDD implementation cycle\"",
		"tools: \"Read, Write, Edit, Bash, Grep, Glob\"",
		"model: sonnet",
		"permissionMode: bypassPermissions",
		"---",
		"",
		"# " + agentName,
		"",
		"활성 에이전트 정의.",
	}, "\n")

	agentFile := filepath.Join(agentDir, agentName+".md")
	if err := os.WriteFile(agentFile, []byte(content), 0o644); err != nil {
		t.Fatalf("에이전트 파일 쓰기 실패: %v", err)
	}

	return dir
}

// buildHookInputForAgent는 주어진 에이전트 이름으로 HookInput을 생성한다.
// SubagentStart 이벤트에서 AgentName 필드 사용 (types.go line 233).
func buildHookInputForAgent(agentName, projectDir string) *HookInput {
	return &HookInput{
		HookEventName: string(EventSubagentStart),
		AgentName:     agentName,
		CWD:           projectDir,
		AgentID:       "test-agent-id-" + agentName,
	}
}

// TestAgentStartHandler_RoutesViaFactory는 NewAgentStartHandler() 생성자가
// 비-nil Handler를 반환하고 EventType이 EventSubagentStart인지 검증한다.
//
// REQ-RA-004: SubagentStart handler 신규 등록
// REQ-RA-009: factory dispatch for agent-start event
func TestAgentStartHandler_RoutesViaFactory(t *testing.T) {
	t.Parallel()

	handler := NewAgentStartHandler()
	if handler == nil {
		t.Fatal("NewAgentStartHandler() returned nil")
	}
	if handler.EventType() != EventSubagentStart {
		t.Errorf("EventType() = %q, want %q", handler.EventType(), EventSubagentStart)
	}
	_, ok := handler.(agentStartInterface)
	if !ok {
		t.Error("AgentStartHandler does not implement agentStartInterface")
	}
}

// TestAgentStartHandler_BlocksRetiredAgent는 retired:true frontmatter를 가진
// 에이전트 이름으로 호출 시 decision=block이 반환되는지 검증한다.
//
// REQ-RA-007: retired agent spawn 시 block decision + reason
func TestAgentStartHandler_BlocksRetiredAgent(t *testing.T) {
	t.Parallel()

	projectDir := buildRetiredAgentDir(t, "manager-tdd", "manager-cycle", "cycle_type=tdd")
	handler := NewAgentStartHandler()
	input := buildHookInputForAgent("manager-tdd", projectDir)

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() 오류: %v", err)
	}
	if output == nil {
		t.Fatal("Handle() returned nil output")
	}
	// decision=block 검증 (REQ-RA-007)
	if output.Decision != DecisionBlock && output.Decision != DecisionDeny {
		t.Errorf("retired agent에 대해 block/deny 결정이 없음, got: %q", output.Decision)
	}
	// reason에 replacement 에이전트 이름 포함 (REQ-RA-007)
	if !strings.Contains(output.Reason, "manager-cycle") {
		t.Errorf("reason에 'manager-cycle' 없음: %q", output.Reason)
	}
	// reason에 cycle_type=tdd 힌트 포함 (REQ-RA-007)
	if !strings.Contains(output.Reason, "cycle_type") && !strings.Contains(output.Reason, "tdd") {
		t.Errorf("reason에 cycle_type 힌트 없음: %q", output.Reason)
	}
}

// TestAgentStartHandler_AllowsActiveAgent는 활성 에이전트에 대해
// 거부 없이 allow가 반환되는지 검증한다.
//
// REQ-RA-008: 비-retired 에이전트는 spawn 허용
func TestAgentStartHandler_AllowsActiveAgent(t *testing.T) {
	t.Parallel()

	projectDir := buildActiveAgentDir(t, "manager-cycle")
	handler := NewAgentStartHandler()
	input := buildHookInputForAgent("manager-cycle", projectDir)

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() 오류: %v", err)
	}
	// block/deny가 없어야 함
	if output != nil && (output.Decision == DecisionBlock || output.Decision == DecisionDeny) {
		t.Errorf("활성 에이전트에 대해 block/deny 결정이 반환됨: %q (이유: %q)",
			output.Decision, output.Reason)
	}
}

// TestAgentStartHandler_AllowsUnknownAgent는 파일이 없는 에이전트 이름에 대해
// allow가 반환되는지 검증한다 (non-MoAI 에이전트 bypass).
//
// REQ-RA-008: 알 수 없는 에이전트 이름은 bypass (exit 0)
func TestAgentStartHandler_AllowsUnknownAgent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // 에이전트 파일 없는 빈 디렉터리
	handler := NewAgentStartHandler()
	input := buildHookInputForAgent("non-existent-agent-xyz", dir)

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() 오류: %v", err)
	}
	// unknown agent는 block/deny 없이 통과해야 함
	if output != nil && (output.Decision == DecisionBlock || output.Decision == DecisionDeny) {
		t.Errorf("알 수 없는 에이전트에 대해 block/deny 결정이 반환됨 (REQ-RA-008): %q", output.Decision)
	}
}

// TestAgentStartHandler_PerformanceUnder500ms는 retired stub frontmatter 파싱 포함
// 100회 호출의 평균 응답 시간이 500ms 이하인지 검증한다.
//
// REQ-RA-012: SubagentStart guard ≤500ms 응답 시간
func TestAgentStartHandler_PerformanceUnder500ms(t *testing.T) {
	if testing.Short() {
		t.Skip("성능 테스트: -short 플래그로 스킵")
	}

	projectDir := buildRetiredAgentDir(t, "manager-tdd", "manager-cycle", "cycle_type=tdd")
	handler := NewAgentStartHandler()
	input := buildHookInputForAgent("manager-tdd", projectDir)

	const iterations = 100
	start := time.Now()
	for i := range iterations {
		_, err := handler.Handle(context.Background(), input)
		if err != nil {
			t.Fatalf("iteration %d Handle() 오류: %v", i, err)
		}
	}
	elapsed := time.Since(start)
	avgMs := elapsed.Milliseconds() / iterations

	const maxAvgMs = 500
	if avgMs > maxAvgMs {
		t.Errorf("평균 응답 시간 %dms가 %dms 초과 (REQ-RA-012: ≤500ms 필요)",
			avgMs, maxAvgMs)
	}
	t.Logf("성능: %d회 평균 %dms (총 %v)", iterations, avgMs, elapsed)
}

// TestAgentStartHandler_OutputFormat은 block decision 출력이
// 올바른 JSON 형식인지 검증한다 (hook stdout 프로토콜 호환).
//
// REQ-RA-007: {"decision":"block","reason":"..."} JSON 형식
func TestAgentStartHandler_OutputFormat(t *testing.T) {
	t.Parallel()

	projectDir := buildRetiredAgentDir(t, "manager-tdd", "manager-cycle", "cycle_type=tdd")
	handler := NewAgentStartHandler()
	input := buildHookInputForAgent("manager-tdd", projectDir)

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() 오류: %v", err)
	}

	// JSON 직렬화 가능 여부 검증
	jsonBytes, marshalErr := json.Marshal(output)
	if marshalErr != nil {
		t.Fatalf("HookOutput JSON 직렬화 실패: %v", marshalErr)
	}
	t.Logf("block output JSON: %s", jsonBytes)

	// decision, reason 필드 존재 확인
	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("JSON 역직렬화 실패: %v", err)
	}

	if decision, ok := parsed["decision"]; !ok || decision == "" {
		t.Errorf("JSON에 'decision' 필드가 없거나 비어있음: %s", jsonBytes)
	}
	if reason, ok := parsed["reason"]; !ok || reason == "" {
		t.Errorf("JSON에 'reason' 필드가 없거나 비어있음: %s", jsonBytes)
	}
	_ = strings.Contains // 컴파일러 경고 방지
}
