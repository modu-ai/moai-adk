package web

// SPEC-WEB-CONSOLE-011 M3 agent-settings AC 바인딩 테스트:
// AC-WC11-020(4표면 렌더), 022(role_profile seam diff), 023(RoleProfileEntry
// Effort 부재), 024/071(enum reject 4xx + 파일 불변), 025(frontmatter round-trip),
// 026(workflow_agents 반영 + taxonomy 참조), 028(지속 경고 + i18n ×4),
// 029(검증/부재 보존), 072(workflow_agents upsert golden).

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
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
// 4표면 전부 렌더 — (a) llm tiers, (b) role_profiles(7), (c) sub-agent
// frontmatter, (d) workflow_agents(7 purposes). AC-WC11-026의 taxonomy 참조와
// AC-WC11-028의 지속 경고 렌더도 함께 고정한다.
func TestAgentSettingsFourSurfacesRendered(t *testing.T) {
	a, _ := newAgentTestApp(t)
	body := getIndex(t, a)

	// (a) llm tiers (LLM 섹션 + 교차 참조 노트).
	for _, marker := range []string{`name="llm.claude_models.high"`, `data-i18n="agentfm.llmnote"`} {
		if !strings.Contains(body, marker) {
			t.Errorf("surface (a) marker missing: %s", marker)
		}
	}
	// (b) role_profiles — 7 profiles × model/effort/isolation/mode.
	for _, p := range []string{"analyst", "architect", "designer", "implementer", "researcher", "reviewer", "tester"} {
		if !strings.Contains(body, `name="workflow.team.role_profiles.`+p+`.model"`) {
			t.Errorf("surface (b) profile %q model control missing", p)
		}
		if !strings.Contains(body, `name="workflow.team.role_profiles.`+p+`.effort"`) {
			t.Errorf("surface (b) profile %q effort control missing", p)
		}
	}
	// (c) sub-agent frontmatter rows.
	for _, marker := range []string{`name="agentfm.dev-a.model"`, `name="agentfm.docs-b.effort"`} {
		if !strings.Contains(body, marker) {
			t.Errorf("surface (c) marker missing: %s", marker)
		}
	}
	// (d) workflow_agents — 7 purposes.
	for _, purpose := range []string{
		"read-only-extract", "mechanical-transform", "synthesize",
		"research", "verify-judge", "implement", "design-architecture",
	} {
		if !strings.Contains(body, `name="workflow.workflow_agents.`+purpose+`.model"`) {
			t.Errorf("surface (d) purpose %q control missing", purpose)
		}
	}
	// AC-WC11-026: taxonomy 참조 잔존.
	if !strings.Contains(body, ".claude/rules/moai/workflow/dynamic-workflows.md") {
		t.Error("dynamic-workflows.md taxonomy reference missing")
	}
	// AC-WC11-028: 지속 경고 렌더.
	if !strings.Contains(body, `data-i18n="agentfm.warn"`) {
		t.Error("persistent moai-update warning missing")
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

// TestRoleProfileEditRoutesThroughSeam은 AC-WC11-022를 검증한다: role_profile
// model 편집이 seam 경유로 대상 스칼라 1라인만 변경하고 주석/`team.patterns`를
// 보존한다 (GWT-2).
func TestRoleProfileEditRoutesThroughSeam(t *testing.T) {
	a, root := newAgentTestApp(t)
	before := readSectionFile(t, root, "workflow")

	rec := postSave(t, a, url.Values{"workflow.team.role_profiles.implementer.model": {"opus"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save status = %d, want 200 (body: %.300s)", rec.Code, rec.Body.String())
	}
	after := readSectionFile(t, root, "workflow")

	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")
	if len(beforeLines) != len(afterLines) {
		t.Fatalf("workflow.yaml line count drifted: %d → %d", len(beforeLines), len(afterLines))
	}
	changed := 0
	for i := range beforeLines {
		if beforeLines[i] != afterLines[i] {
			changed++
			if !strings.Contains(afterLines[i], "model: opus") {
				t.Errorf("unexpected changed line: %q", afterLines[i])
			}
		}
	}
	if changed != 1 {
		t.Errorf("changed lines = %d, want exactly 1", changed)
	}
	for _, keep := range []string{"patterns:", "design_implementation:", "# effort: declarative metadata"} {
		if !strings.Contains(after, keep) {
			t.Errorf("preserved content %q lost", keep)
		}
	}
}

// TestRoleProfileEntryHasNoEffortField는 AC-WC11-023의 컴파일 계층 증거다:
// RoleProfileEntry struct 블록에 Effort 필드가 없다 (REQ-WEM-006 유지 — 신설
// WorkflowAgentEntry.Effort는 본 assertion 범위 밖).
func TestRoleProfileEntryHasNoEffortField(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf(config.RoleProfileEntry{})
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Name == "Effort" {
			t.Error("RoleProfileEntry gained an Effort field — REQ-WEM-006 reversal prohibited (REQ-WC11-023)")
		}
	}
}

// TestAgentSettingsEnumReject는 AC-WC11-024 + AC-WC11-071을 검증한다:
// v4manifest closed set 밖 값 제출 → 4xx + 파일 불변.
func TestAgentSettingsEnumReject(t *testing.T) {
	a, root := newAgentTestApp(t)
	before := readSectionFile(t, root, "workflow")

	rec := postSave(t, a, url.Values{
		"workflow.team.role_profiles.implementer.effort": {"superhigh"},
		"workflow.workflow_agents.implement.model":       {"gpt5"},
		"workflow.workflow_agents.implement.effort":      {"ultra"},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("out-of-set submission status = %d, want 400", rec.Code)
	}
	if got := readSectionFile(t, root, "workflow"); got != before {
		t.Error("workflow.yaml changed despite enum reject (atomic reject violated)")
	}
}

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
	// 재렌더에 현재 값 반영 (선택 마킹 ✓).
	body := getIndex(t, a)
	if !strings.Contains(body, "high ✓") {
		t.Error("re-render does not reflect the patched effort value")
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

// TestWorkflowAgentsUpsertGolden은 AC-WC11-072를 검증한다: 블록 부재 fixture에
// workflow_agents 최초 기록(upsert) → 기존 라인 전부 보존 + 신규 블록만 추가
// (GWT-7).
func TestWorkflowAgentsUpsertGolden(t *testing.T) {
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

	if !strings.HasPrefix(after, strings.TrimSuffix(before, "\n")) {
		// additive-only: 기존 내용은 prefix로 보존되어야 한다.
		t.Error("upsert was not additive-only (existing lines drifted)")
	}
	for _, want := range []string{"workflow_agents:", "implement:", "model: sonnet", "effort: xhigh"} {
		if !strings.Contains(after, want) {
			t.Errorf("upserted block missing %q", want)
		}
	}
	for _, keep := range []string{"patterns:", "role_profile_keys:"} {
		if !strings.Contains(after, keep) {
			t.Errorf("preserved content %q lost across upsert", keep)
		}
	}
}
