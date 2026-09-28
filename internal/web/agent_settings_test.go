package web

// Settings-page tests that outlived the agent-settings tab: the retired Agent
// Teams role_profiles controls stay unrendered, and a forged workflow_agents
// submission stays ignored. The per-agent model/effort panel itself was removed
// (SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-011).

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newAgentTestApp builds the schema test app used by the settings-page tests.
func newAgentTestApp(t *testing.T) (*app, string) {
	t.Helper()
	return newSchemaTestApp(t)
}

func getIndex(t *testing.T, a *app) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	return rec.Body.String()
}

func TestNoTeamRoleProfileRender(t *testing.T) {
	a, _ := newAgentTestApp(t)
	body := getIndex(t, a)
	for _, p := range []string{"analyst", "architect", "designer", "implementer", "researcher", "reviewer", "tester"} {
		for _, field := range []string{"model", "effort", "isolation", "mode"} {
			marker := `name="workflow.team.role_profiles.` + p + `.` + field + `"`
			if strings.Contains(body, marker) {
				t.Errorf("Agent Teams role_profile control still rendered: %s", marker)
			}
		}
	}
	// team.max_teammates 등 team seam 필드도 렌더되지 않는다.
	if strings.Contains(body, `name="workflow.team.max_teammates"`) {
		t.Error("Agent Teams workflow.team.max_teammates control still rendered")
	}
}

// TestAgentFMWarnI18nParity는 AC-WC11-028의 4-locale half다: agentfm 신규 키가

// TestWorkflowAgentsWebSubmissionIgnored는 M5-a B1의 행동 완결이다: workflow_agents
// 폼 제출은 웹에서 더 이상 렌더/쓰기하지 않으므로 무시된다 — 블록은 생성되지 않고
// 기존 workflow.yaml 내용은 불변이다. struct 필드(config.Workflow.WorkflowAgents)와
// yaml 키는 유지되며, dynamic-workflow JS가 yaml 파일을 직접 읽는 소비자다
// (dynamic-workflows.md §Config surface). 본 테스트는 선행 TestWorkflowAgentsUpsertGolden
// (AC-WC11-072 웹 쓰기 경로)을 M5-a B1 이후 행동으로 대체한다.
func TestWorkflowAgentsWebSubmissionIgnored(t *testing.T) {
	a, root := newAgentTestApp(t)
	before := readSectionFile(t, root, "workflow")
	if strings.Contains(before, "workflow_agents") {
		t.Fatal("fixture precondition violated: workflow_agents already present")
	}

	rec := postSave(t, a, url.Values{
		"workflow.workflow_agents.implement.model":  {"sonnet"},
		"workflow.workflow_agents.implement.effort": {"xhigh"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save status = %d, want 200 (body: %.300s)", rec.Code, rec.Body.String())
	}
	after := readSectionFile(t, root, "workflow")
	if strings.Contains(after, "workflow_agents") {
		t.Errorf("workflow_agents block created by web submission — should be ignored (M5-a B1):\n%s", after)
	}
	// 무시된 제출은 workflow.yaml을 byte 단위로 불변으로 남긴다 (동일 패턴:
	// TestSaveExcludedSectionForgedPost). "model: sonnet"/"effort: xhigh" 값 자체는
	// role_profiles fixture에 이미 존재하므로 값 문자열 검사가 아닌 byte 동등성으로
	// 검증한다.
	if before != after {
		t.Errorf("workflow.yaml mutated by ignored workflow_agents submission (M5-a B1)")
	}
	for _, keep := range []string{"patterns:", "role_profile_keys:"} {
		if !strings.Contains(after, keep) {
			t.Errorf("preserved content %q lost", keep)
		}
	}
}
