// agent_start_test.go: verifies the SubagentStart hook's retired-agent rejection logic.
// Mapping: REQ-RA-004, REQ-RA-007, REQ-RA-008, REQ-RA-009, REQ-RA-012.
//
// M3 GREEN-2 phase: agentStartHandler is implemented in subagent_start.go, so
// every t.Skip is converted into a real assertion.
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

// agentStartInterface is a type-assertion helper used at compile time to confirm
// that the AgentStartHandler struct exists in this package.
type agentStartInterface interface {
	Handle(ctx context.Context, input *HookInput) (*HookOutput, error)
	EventType() EventType
}

// buildRetiredAgentDir creates a test temp directory and writes
// .claude/agents/core/<agentName>.md with retired frontmatter.
// (post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: agents are split into core/expert/meta/harness)
func buildRetiredAgentDir(t *testing.T, agentName, replacement, paramHint string) string {
	t.Helper()

	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".claude", "agents", "core")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("임시 에이전트 디렉터리 생성 실패: %v", err)
	}

	// Standardized retired stub frontmatter (REQ-RA-002)
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

// buildActiveAgentDir creates a test temp directory and writes
// .claude/agents/core/<agentName>.md with active frontmatter.
// (post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: agents are split into core/expert/meta/harness)
func buildActiveAgentDir(t *testing.T, agentName string) string {
	t.Helper()

	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".claude", "agents", "core")
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

// buildHookInputForAgent builds a HookInput for the given agent name.
// On SubagentStart events the AgentName field is used (types.go line 233).
func buildHookInputForAgent(agentName, projectDir string) *HookInput {
	return &HookInput{
		HookEventName: string(EventSubagentStart),
		AgentName:     agentName,
		CWD:           projectDir,
		AgentID:       "test-agent-id-" + agentName,
	}
}

// TestAgentStartHandler_RoutesViaFactory verifies that the NewAgentStartHandler()
// constructor returns a non-nil Handler whose EventType is EventSubagentStart.
//
// REQ-RA-004: register the new SubagentStart handler
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

// TestAgentStartHandler_BlocksRetiredAgent verifies that decision=block is returned
// when invoked with an agent name whose frontmatter has retired:true.
//
// REQ-RA-007: block decision + reason when spawning a retired agent
func TestAgentStartHandler_BlocksRetiredAgent(t *testing.T) {
	t.Parallel()

	projectDir := buildRetiredAgentDir(t, "manager-tdd", "manager-develop", "cycle_type=tdd")
	handler := NewAgentStartHandler()
	input := buildHookInputForAgent("manager-tdd", projectDir)

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() 오류: %v", err)
	}
	if output == nil {
		t.Fatal("Handle() returned nil output")
	}
	// Verify decision=block (REQ-RA-007)
	if output.Decision != DecisionBlock && output.Decision != DecisionDeny {
		t.Errorf("retired agent에 대해 block/deny 결정이 없음, got: %q", output.Decision)
	}
	// reason must include the replacement agent name (REQ-RA-007)
	if !strings.Contains(output.Reason, "manager-develop") {
		t.Errorf("reason에 'manager-develop' 없음: %q", output.Reason)
	}
	// reason must include the cycle_type=tdd hint (REQ-RA-007)
	if !strings.Contains(output.Reason, "cycle_type") && !strings.Contains(output.Reason, "tdd") {
		t.Errorf("reason에 cycle_type 힌트 없음: %q", output.Reason)
	}
}

// TestAgentStartHandler_AllowsActiveAgent verifies that an active agent
// is allowed without rejection.
//
// REQ-RA-008: non-retired agents are allowed to spawn
func TestAgentStartHandler_AllowsActiveAgent(t *testing.T) {
	t.Parallel()

	projectDir := buildActiveAgentDir(t, "manager-develop")
	handler := NewAgentStartHandler()
	input := buildHookInputForAgent("manager-develop", projectDir)

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() 오류: %v", err)
	}
	// Must have no block/deny
	if output != nil && (output.Decision == DecisionBlock || output.Decision == DecisionDeny) {
		t.Errorf("활성 에이전트에 대해 block/deny 결정이 반환됨: %q (이유: %q)",
			output.Decision, output.Reason)
	}
}

// TestAgentStartHandler_AllowsUnknownAgent verifies that an agent name with no
// corresponding file is allowed (non-MoAI agent bypass).
//
// REQ-RA-008: unknown agent names are bypassed (exit 0)
func TestAgentStartHandler_AllowsUnknownAgent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // Empty directory with no agent files
	handler := NewAgentStartHandler()
	input := buildHookInputForAgent("non-existent-agent-xyz", dir)

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() 오류: %v", err)
	}
	// Unknown agents must pass without block/deny
	if output != nil && (output.Decision == DecisionBlock || output.Decision == DecisionDeny) {
		t.Errorf("알 수 없는 에이전트에 대해 block/deny 결정이 반환됨 (REQ-RA-008): %q", output.Decision)
	}
}

// TestAgentStartHandler_PerformanceUnder500ms verifies that the mean response time
// across 100 calls — including retired stub frontmatter parsing — is at most 500ms.
//
// REQ-RA-012: SubagentStart guard must respond in ≤500ms
func TestAgentStartHandler_PerformanceUnder500ms(t *testing.T) {
	if testing.Short() {
		t.Skip("성능 테스트: -short 플래그로 스킵")
	}

	projectDir := buildRetiredAgentDir(t, "manager-tdd", "manager-develop", "cycle_type=tdd")
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

// TestAgentStartHandler_OutputFormat verifies the block-decision output is in
// the correct JSON form (compatible with the hook stdout protocol).
//
// REQ-RA-007: JSON form {"decision":"block","reason":"..."}
func TestAgentStartHandler_OutputFormat(t *testing.T) {
	t.Parallel()

	projectDir := buildRetiredAgentDir(t, "manager-tdd", "manager-develop", "cycle_type=tdd")
	handler := NewAgentStartHandler()
	input := buildHookInputForAgent("manager-tdd", projectDir)

	output, err := handler.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("Handle() 오류: %v", err)
	}

	// Verify JSON serialization succeeds
	jsonBytes, marshalErr := json.Marshal(output)
	if marshalErr != nil {
		t.Fatalf("HookOutput JSON 직렬화 실패: %v", marshalErr)
	}
	t.Logf("block output JSON: %s", jsonBytes)

	// Verify the decision and reason fields exist
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
	_ = strings.Contains // avoid compiler warning
}
