package web

// SPEC-WEB-CONSOLE-011 M2b 웹 계층 AC 바인딩 테스트:
// AC-WC11-004(워크플로우 저장의 seam 라우팅 end-to-end), 012(oneof 위반 4xx),
// 013(llm.mode/team_mode read-only), 014(빈 tier placeholder), 016(10섹션 스모크),
// 018(제외군 비노출 + 위조 POST 무시, EC-8), 019(db 3/5 분리 웹 계층),
// 063(raw view 블록 — input 컨트롤 0 + 같은 섹션 스칼라는 폼 필드).

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/settings"
)

// allSectionFixtures는 섹션 fixture 이름이다 (settings testdata 재사용 —
// fixture 중복 없이 실제 파일 사본으로 테스트). SPEC-WEB-CONSOLE-013 M2에서
// handoff/cache 2종이 추가되었다.
var allSectionFixtures = []string{
	"git-strategy", "llm", "quality",
	"workflow", "harness", "ralph", "research", "feedback", "observability", "security",
	"handoff", "cache",
}

// seedWebSections는 internal/settings/testdata/sections의 실제 섹션 fixture를
// 웹 테스트의 임시 프로젝트 루트로 복사한다.
func seedWebSections(t *testing.T, root string, names ...string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		src, err := os.ReadFile(filepath.Join("..", "settings", "testdata", "sections", name+".yaml"))
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, name+".yaml"), src, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// newSchemaTestApp은 섹션 fixture가 시드된 프로젝트 루트와 프로필 no-op seam을
// 가진 앱을 만든다 (프로필 스토어 실쓰기 방지 — 섹션 영속화 경로만 검증).
func newSchemaTestApp(t *testing.T) (*app, string) {
	t.Helper()
	root := t.TempDir()
	seedWebSections(t, root, allSectionFixtures...)
	a := newApp(Config{ProjectRoot: root, ProfileName: "default"})
	a.readPreferences = func(string) (profile.ProfilePreferences, error) {
		return profile.ProfilePreferences{}, nil
	}
	a.writePreferences = func(string, profile.ProfilePreferences) error { return nil }
	a.syncToProject = func(string, profile.ProfilePreferences) error { return nil }
	return a, root
}

// postSave는 loopback Host로 POST /save를 수행한다.
func postSave(t *testing.T, a *app, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/save", strings.NewReader(form.Encode()))
	req.Host = "127.0.0.1:8080"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin") // simulate real browser form POST (REQ-SEC-002)
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	return rec
}

// readSectionFile은 임시 루트의 섹션 파일을 읽는다.
func readSectionFile(t *testing.T, root, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", name+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// TestSchemaSectionsRenderSmoke는 AC-WC11-016(GET half) + AC-WC11-014 + AC-WC11-018
// (렌더 half) + AC-WC11-063을 검증한다: 10섹션 각각 최소 1개 스칼라 폼 필드 렌더,
// 빈 tier의 "(runtime default)" placeholder, 제외군 폼 컨트롤 0, raw view 블록은
// input 없이 렌더 + read-only 표시 렌더.
func TestSchemaSectionsRenderSmoke(t *testing.T) {
	a, _ := newSchemaTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "127.0.0.1:8080"
	rec := httptest.NewRecorder()
	a.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	// SPEC-WEBCONF-SIMPLIFY-001 M3: only the 2 surviving schema-rendered sections
	// (git_strategy, llm) render form controls. The 10 removed sections' fields are
	// schema-preserved but web-not-rendered (tabs removed — REQ-WC-003).
	for _, name := range []string{
		"git_strategy.mode",
		"llm.glm.models.high",
	} {
		if !strings.Contains(body, `name="`+name+`"`) {
			t.Errorf("rendered page missing form control %q (surviving section)", name)
		}
	}

	// 제외군 + M3-removed 섹션 폼 컨트롤 0 (AC-WC11-018 렌더 half + M3 reclassified).
	for _, prefix := range []string{
		"state.", "system.", "sunset.", "tool-policy.", "lsp.", "mx.", "constitution.", "context.", "interview.",
		// SPEC-WEBCONF-SIMPLIFY-001 M3: 8 former seam sections removed from UI.
		"workflow.", "harness.", "ralph.", "feedback.", "observability.", "security.", "handoff.", "cacheStrategy.",
	} {
		if strings.Contains(body, `name="`+prefix) {
			t.Errorf("excluded/removed section control rendered: name=%q...", prefix)
		}
	}

	// read-only 표시 (REQ-WC11-013/019): 편집 컨트롤 없음 + 표시 렌더.
	for _, ro := range []string{"llm.mode", "llm.team_mode"} {
		if strings.Contains(body, `name="`+ro+`"`) {
			t.Errorf("read-only key %q rendered as a form control", ro)
		}
	}
	if !strings.Contains(body, "read-only (runtime-managed)") {
		t.Error("read-only note not rendered")
	}

	// SPEC-WEBCONF-SIMPLIFY-001 M3: raw view blocks belonged to removed sections
	// (harness.levels, security.*, mx.*) — none render after M3. The rawview
	// input-control guard below is retained (no-op when no rawview blocks exist).
	rest := body
	for {
		start := strings.Index(rest, `<details class="rawview">`)
		if start < 0 {
			break
		}
		end := strings.Index(rest[start:], "</details>")
		if end < 0 {
			t.Fatal("unclosed rawview details block")
		}
		block := rest[start : start+end]
		if strings.Contains(block, "<input") || strings.Contains(block, "<select") {
			t.Errorf("raw view block contains a form control:\n%s", block)
		}
		rest = rest[start+end:]
	}
}

// TestSaveWorkflowRoutesThroughSeam은 SPEC-WEBCONF-SIMPLIFY-001 M3 이후 workflow
// 섹션이 RouteExcluded로 재분류되어 web write path가 제거되었음을 검증한다.
// POST /save의 workflow 필드 제출은 WriteSectionViaSeam에서 거부되고, 디스크의
// workflow.yaml이 바이트 단위로 무변경이다 (REQ-WC-003 — config keys persist,
// web write path removed).
func TestSaveWorkflowRoutesThroughSeam(t *testing.T) {
	a, root := newSchemaTestApp(t)
	before := readSectionFile(t, root, "workflow")

	postSave(t, a, url.Values{"workflow.token_budget.plan": {"31000"}})

	after := readSectionFile(t, root, "workflow")
	if after != before {
		t.Errorf("workflow.yaml mutated by a removed-section POST (M3 RouteExcluded must block the write):\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestSaveLLMModeReadOnlyIgnored는 AC-WC11-013을 검증한다: mode/team_mode 변경
// 제출은 무시되고 llm.yaml의 두 키는 불변이다 (runtime-managed read-only).
func TestSaveLLMModeReadOnlyIgnored(t *testing.T) {
	a, root := newSchemaTestApp(t)
	before := readSectionFile(t, root, "llm")

	rec := postSave(t, a, url.Values{
		"llm.mode":      {"glm"},
		"llm.team_mode": {"glm"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save status = %d, want 200", rec.Code)
	}
	after := readSectionFile(t, root, "llm")
	if before != after {
		t.Errorf("llm.yaml changed by a read-only key submission:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestSaveInvalidSchemaValueRejected는 닫힌 옵션 위반 제출의 4xx atomic reject를
// 검증한다 (AC-WC11-012의 oneof 위반 4xx + EC-2 — 파일 불변).
func TestSaveInvalidSchemaValueRejected(t *testing.T) {
	a, root := newSchemaTestApp(t)
	beforeGS := readSectionFile(t, root, "git-strategy")
	beforeLLM := readSectionFile(t, root, "llm")

	// git_strategy.mode oneof 위반만으로 4xx atomic reject 발생 (M4 다이어트로
	// llm.performance_tier 필드가 제거되어 turbo 제출은 무시된다 — schema 미등록).
	rec := postSave(t, a, url.Values{
		"git_strategy.mode": {"bogus"},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /save status = %d, want 400", rec.Code)
	}
	if got := readSectionFile(t, root, "git-strategy"); got != beforeGS {
		t.Error("git-strategy.yaml changed despite validation reject (atomic reject violated)")
	}
	if got := readSectionFile(t, root, "llm"); got != beforeLLM {
		t.Error("llm.yaml changed despite validation reject (atomic reject violated)")
	}
}

// TestSaveExcludedSectionForgedPost는 EC-8 / AC-WC11-018 쓰기 half를 검증한다:
// 제외군 섹션 이름을 위조한 제출은 무시되고 어떤 섹션 파일도 생성/변경되지 않는다.
func TestSaveExcludedSectionForgedPost(t *testing.T) {
	a, root := newSchemaTestApp(t)
	dir := filepath.Join(root, ".moai", "config", "sections")
	entriesBefore, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	workflowBefore := readSectionFile(t, root, "workflow")

	rec := postSave(t, a, url.Values{
		"tool-policy.max_calls": {"999"},
		"state.hacked":          {"true"},
		"constitution.evil":     {"x"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save status = %d, want 200 (ignore semantics)", rec.Code)
	}
	entriesAfter, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entriesAfter) != len(entriesBefore) {
		t.Errorf("forged POST created section files: %d → %d", len(entriesBefore), len(entriesAfter))
	}
	if got := readSectionFile(t, root, "workflow"); got != workflowBefore {
		t.Error("forged POST mutated an unrelated section file")
	}
}

// TestSaveSchemaSmokeAllSections는 SPEC-WEBCONF-SIMPLIFY-001 M3 이후 surviving
// schema-rendered sections(git_strategy, llm)의 저장 라운드트립을 검증한다.
// M3가 8개 seam 섹션을 RouteExcluded로 재분류하여 이들의 저장은 거부된다 —
// 본 스모크는 surviving typed 섹션만 다룬다.
func TestSaveSchemaSmokeAllSections(t *testing.T) {
	a, root := newSchemaTestApp(t)

	form := url.Values{
		"llm.glm.models.high": {"glm-test"},
	}
	rec := postSave(t, a, form)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /save status = %d, want 200 (body: %.300s)", rec.Code, rec.Body.String())
	}

	values, err := settings.SchemaCurrentValues(root)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"llm.glm.models.high": "glm-test",
	} {
		if got := values[name]; got != want {
			t.Errorf("persisted %q = %q, want %q", name, got, want)
		}
	}
}

// TestHandoffModeValidation은 AC-WC13-012를 검증한다: handoff.mode 닫힌 집합
// {manual, auto} 밖의 값 제출은 4xx atomic reject되고 handoff.yaml이 바이트
// 무변경이며(EC-2), 유효 값(auto)은 seam으로 반영되고 미노출/주석이 보존된다.
func TestHandoffModeValidation(t *testing.T) {
	a, root := newSchemaTestApp(t)
	before := readSectionFile(t, root, "handoff")

	// (1) 닫힌 집합 밖 값 → 4xx + 파일 무변경.
	postSave(t, a, url.Values{"handoff.mode": {"bogus"}})
	if got := readSectionFile(t, root, "handoff"); got != before {
		t.Error("handoff.yaml changed despite M3 RouteExcluded (write must be blocked)")
	}

	// (2) 유효 값 auto → 200 + seam 반영 + guide/주석 보존.
	postSave(t, a, url.Values{"handoff.mode": {"auto"}})
	if got := readSectionFile(t, root, "handoff"); got != before {
		t.Error("handoff.yaml changed by auto POST despite M3 RouteExcluded (write must be blocked)")
	}
}

// TestCacheSessionTTLValidation은 AC-WC13-013 웹 계층을 검증한다:
// cacheStrategy.session_ttl 닫힌 집합 {1h,5m,off} 밖의 값 제출은 4xx atomic
// reject되고 cache.yaml이 무변경이며, 미노출 키(spec_ttl/min_cacheable_tokens)가
// 보존된다 (REQ-WC13-006/013 — acceptance.md §D.2 시나리오 2).
func TestCacheSessionTTLValidation(t *testing.T) {
	a, root := newSchemaTestApp(t)
	before := readSectionFile(t, root, "cache")

	postSave(t, a, url.Values{"cacheStrategy.session_ttl": {"2h"}})
	if got := readSectionFile(t, root, "cache"); got != before {
		t.Error("cache.yaml changed despite M3 RouteExcluded (write must be blocked)")
	}

	// 유효 값 off + enabled 토글 → 반영 + 미노출 키 보존.
	postSave(t, a, url.Values{
		"cacheStrategy.session_ttl":      {"off"},
		"cacheStrategy.enabled__present": {"1"},
		"cacheStrategy.enabled":          {"on"},
	})
	if got := readSectionFile(t, root, "cache"); got != before {
		t.Error("cache.yaml changed by off/enabled POST despite M3 RouteExcluded (write must be blocked)")
	}
}
