package web

// SPEC-WEB-CONSOLE-011 M3 sub-agent frontmatter AC 바인딩 테스트:
// AC-WC11-020(잔존 표면 렌더), 025(frontmatter round-trip),
// 026(workflow_agents 숨김 + taxonomy 참조), 028(지속 경고 + i18n ×4),
// 029(검증/부재 보존).
// (구 AC-WC11-023 컴파일 계층 테스트는 SPEC-AGENT-TEAM-RETIRE-001 M1에서 제거.
// team.role_profiles 웹 렌더 표면(구 surface (b), AC-WC11-022/024/071)은
// SPEC-AGENT-TEAM-RETIRE-001 M2에서 제거 — 웹 콘솔은 더 이상 Agent Teams 설정을
// 렌더하지 않는다. sub-agent frontmatter 편집(agentfm)은 무관하게 유지된다.)

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 테스트용 agent 파일 2종: effort 보유 + effort 부재 (manager-docs 선례, EC-7).
const testAgentWithEffort = `---
name: dev-a
model: inherit
effort: xhigh
memory: project
---
# dev-a body

Sentinel body content (byte-preservation).
`

const testAgentNoEffort = `---
name: docs-b
model: haiku
memory: project
---
docs-b body sentinel.
`

// newAgentTestApp은 섹션 fixture + agent 파일이 시드된 앱을 만든다.
func newAgentTestApp(t *testing.T) (*app, string) {
	t.Helper()
	a, root := newSchemaTestApp(t)
	agentsDir := filepath.Join(root, ".claude", "agents", "moai")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "dev-a.md"), []byte(testAgentWithEffort), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "docs-b.md"), []byte(testAgentNoEffort), 0o644); err != nil {
		t.Fatal(err)
	}
	return a, root
}

func getIndex(t *testing.T, a *app) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	return rec.Body.String()
}

// TestAgentSettingsFourSurfacesRendered는 AC-WC11-020을 검증한다: 단일 뷰에
// 잔존 표면 렌더 — (a) llm tiers, (c) sub-agent frontmatter, (d) workflow_agents.
// 구 surface (b) team.role_profiles(7)는 SPEC-AGENT-TEAM-RETIRE-001 M2에서 제거
// (웹 콘솔은 더 이상 Agent Teams 설정을 렌더하지 않는다). M5-a B1부터 (d)
// workflow_agents는 웹 렌더에서 숨김 — 폼 컨트롤이 렌더되지 않는다 (struct/yaml
// 키는 유지). AC-WC11-026의 taxonomy 참조와 AC-WC11-028의 지속 경고 렌더도 함께
// 고정한다.
func TestAgentSettingsFourSurfacesRendered(t *testing.T) {
	a, _ := newAgentTestApp(t)
	body := getIndex(t, a)

	// (a) llm tiers — M4 다이어트 후 GLM tier 매핑만 잔류 (claude_models 제거됨).
	for _, marker := range []string{`name="llm.glm.models.high"`, `data-i18n="agentfm.llmnote"`} {
		if !strings.Contains(body, marker) {
			t.Errorf("surface (a) marker missing: %s", marker)
		}
	}
	// 구 surface (b) team.role_profiles 렌더 검증은 Agent Teams 정적 레이어와 함께
	// 제거되었다 (SPEC-AGENT-TEAM-RETIRE-001 M2 — workflow.team.role_profiles.* 폼
	// 컨트롤은 더 이상 렌더되지 않는다). 잔존 확인은 아래 TestNoTeamRoleProfileRender.
	// (c) sub-agent frontmatter rows.
	for _, marker := range []string{`name="agentfm.dev-a.model"`, `name="agentfm.docs-b.effort"`} {
		if !strings.Contains(body, marker) {
			t.Errorf("surface (c) marker missing: %s", marker)
		}
	}
	// (d) workflow_agents — M5-a B1부터 웹 렌더에서 숨김. 7 purposes의 폼 컨트롤이
	// 렌더되지 않는다 (struct 필드 + yaml 키는 유지, dynamic-workflow JS가 yaml
	// 파일을 직접 읽는다).
	for _, purpose := range []string{
		"read-only-extract", "mechanical-transform", "synthesize",
		"research", "verify-judge", "implement", "design-architecture",
	} {
		if strings.Contains(body, `name="workflow.workflow_agents.`+purpose+`.model"`) {
			t.Errorf("surface (d) purpose %q control should be hidden (M5-a B1)", purpose)
		}
	}
	// AC-WC11-026: taxonomy 참조 잔존 (agentfm 섹션 설명에 잔류).
	if !strings.Contains(body, ".claude/rules/moai/workflow/dynamic-workflows.md") {
		t.Error("dynamic-workflows.md taxonomy reference missing")
	}
	// AC-WC11-028: 지속 경고 렌더.
	if !strings.Contains(body, `data-i18n="agentfm.warn"`) {
		t.Error("persistent moai-update warning missing")
	}
	// M5-a B5: effort 필드 "(Go 미독)" 배지 렌더 (agentfm effort — 구 role_profiles.effort
	// 표면은 SPEC-AGENT-TEAM-RETIRE-001 M2에서 제거, 배지는 agentfm이 계속 렌더).
	if !strings.Contains(body, `data-i18n="hint.effort.go_unbound"`) {
		t.Error("effort (Go 미독) hint badge missing")
	}
}

// TestNoTeamRoleProfileRender는 SPEC-AGENT-TEAM-RETIRE-001 M2의 절대-부재를
// 검증한다: 웹 콘솔은 더 이상 Agent Teams team.role_profiles 폼 컨트롤을 렌더하지
// 않는다 (REQ-ATR-008 — "shall not render Agent Teams configuration").
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
// 4개 locale 전부에 존재한다.
func TestAgentFMWarnI18nParity(t *testing.T) {
	dict := readEmbeddedAsset(t, "i18n.js")
	for _, key := range []string{
		"agentfm.warn", "agentfm.keep", "agentfm.absent", "agentfm.llmnote",
		"agentfm.taxonomy", "agentfm.unavailable",
		"sec.agentfm.title", "sec.agentfm.desc",
		"sec.agent_settings.title", "sec.agent_settings.desc",
	} {
		if strings.Count(dict, `"`+key+`":`) < 4 {
			t.Errorf("i18n key %q not present in all 4 locales", key)
		}
	}
}

// NOTE: TestRoleProfileEditRoutesThroughSeam (구 AC-WC11-022 role_profile seam
// diff)와 TestAgentSettingsEnumReject (구 AC-WC11-024/071 role_profiles +
// workflow_agents enum reject)는 SPEC-AGENT-TEAM-RETIRE-001 M2에서 제거되었다 —
// 두 테스트의 대상 표면(workflow.team.role_profiles 편집)이 Agent Teams 정적
// 레이어와 함께 retired. agentfm enum reject 커버리지는
// TestAgentFMValidationAndAbsent가 유지한다.

// TestAgentFMEditRoundTrip은 AC-WC11-025 + GWT-6을 검증한다: frontmatter 편집이
// 파일에 반영되고 body는 byte-identical이며 재렌더에 현재 값이 반영된다.
func TestAgentFMEditRoundTrip(t *testing.T) {
	a, root := newAgentTestApp(t)
	agentPath := filepath.Join(root, ".claude", "agents", "moai", "dev-a.md")

	rec := postSave(t, a, url.Values{"agentfm.dev-a.effort": {"high"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save status = %d, want 200 (body: %.300s)", rec.Code, rec.Body.String())
	}
	raw, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	if !strings.Contains(content, "effort: high") {
		t.Errorf("frontmatter effort not patched:\n%s", content)
	}
	if !strings.Contains(content, "Sentinel body content (byte-preservation).") {
		t.Error("body content lost")
	}
	// 재렌더에 현재 값 반영 (selected 옵션으로 표시).
	body := getIndex(t, a)
	if !strings.Contains(body, `<option value="high" selected`) {
		t.Error("re-render does not reflect the patched effort value (selected option)")
	}
}

// TestAgentFMValidationAndAbsent는 AC-WC11-029를 검증한다: (i) out-of-set
// model/effort → 4xx + 파일 불변; (ii) effort 부재 agent에 model만 저장 시
// effort 키 미주입 (EC-7).
func TestAgentFMValidationAndAbsent(t *testing.T) {
	a, root := newAgentTestApp(t)
	devPath := filepath.Join(root, ".claude", "agents", "moai", "dev-a.md")
	docsPath := filepath.Join(root, ".claude", "agents", "moai", "docs-b.md")
	devBefore, _ := os.ReadFile(devPath)

	// (i) 거부.
	rec := postSave(t, a, url.Values{
		"agentfm.dev-a.model":  {"gpt5"},
		"agentfm.dev-a.effort": {"superhigh"},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("out-of-set frontmatter submission status = %d, want 400", rec.Code)
	}
	if devAfter, _ := os.ReadFile(devPath); string(devAfter) != string(devBefore) {
		t.Error("agent file changed despite validation reject")
	}

	// (ii) effort 부재 agent에 model만 변경 — effort 키 미주입.
	rec = postSave(t, a, url.Values{"agentfm.docs-b.model": {"sonnet"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("model-only patch status = %d, want 200", rec.Code)
	}
	docsAfter, _ := os.ReadFile(docsPath)
	if strings.Contains(string(docsAfter), "effort:") {
		t.Errorf("effort key injected into effort-absent agent:\n%s", docsAfter)
	}
	if !strings.Contains(string(docsAfter), "model: sonnet") {
		t.Errorf("model not patched:\n%s", docsAfter)
	}
}

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
