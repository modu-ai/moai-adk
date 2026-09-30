package web

// t1280 — 설정웹 제목 고착 버그와 고급 접기 영역의 회귀 고정.
//
// 결함(t1280, 구스 스크린샷 #27·#28): 좌측 subnav 탭 전환이 클라이언트에서
// 일어나는 동안 settings-page-head(설정 상세 머리글)는 로드 시점 탭의 제목에
// 고착됐다 — 모든 페이지의 제목이 '에이전트'(진입 탭)로 남는 결함. 수리는 두
// 층이다: (1) shell.templ subnav 행이 탭 메타를 data-page-* 로 실어 주고,
// (2) app.js wireTabs 가 전환 시 updateSettingsPageHead 로 머리글을 맞춘다.
//
// 검증은 저장소 어법을 따른다: 렌더 HTML 문자열 단언 + 내장 app.js 문자열 단언.
// 브라우저 왕복 실측(title-check.mjs)은 판정서 증거로 남긴다.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/profile"
)

// TestSubnavRowsCarryPageHeadMeta — subnav 행 하나하나가 자기 탭의 제목·설명·
// 효과 배지 키를 data-page-* 로 실고 나가는지 본다. updateSettingsPageHead 의
// 유일한 데이터 원천이라 이 속성이 빠지면 수리가 무력화된다.
func TestSubnavRowsCarryPageHeadMeta(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{UserName: "jline"})

	for _, attr := range []string{
		`data-page-title-key=`,
		`data-page-title=`,
		`data-page-desc-key=`,
		`data-page-desc=`,
		`data-page-effect-key=`,
		`data-page-effect=`,
	} {
		if n := strings.Count(body, attr); n < 13 {
			t.Errorf("subnav rows carry %q %d times, want >= 13 (one per tab)", attr, n)
		}
	}

	// crosssession 행이 자기 섹션 키를 실는지 — 결함 보고(#27)의 그 페이지다.
	row := subnavRowFor(t, body, "crosssession")
	for _, want := range []string{
		`data-page-title-key="sec.crosssession.title"`,
		`data-page-desc-key="sec.crosssession.desc"`,
	} {
		if !strings.Contains(row, want) {
			t.Errorf("crosssession subnav row missing %q; row:\n%s", want, row)
		}
	}
}

// TestWireTabsUpdatesPageHead — app.js 가 탭 전환 시 머리글을 갱신하는지 본다.
// 갱신 함수가 존재하고, wireTabs 전환 경로에서 불리며, 갱신 대상 세 요소
// (제목·설명·효과 배지)의 data-i18n 을 되풀이해 바꾸는지 문자열로 고정한다.
func TestWireTabsUpdatesPageHead(t *testing.T) {
	js := readEmbeddedAsset(t, "app.js")

	for _, want := range []string{
		"function updateSettingsPageHead(",
		"updateSettingsPageHead(this);",
		`getElementById("settings-page-title")`,
		".settings-page-head__desc",
		`.settings-page-head__meta .badge`,
		"applyI18n(readPersistedLang())",
	} {
		if !strings.Contains(js, want) {
			t.Errorf("app.js missing page-head update marker %q", want)
		}
	}
}

// TestWorkflowAdvancedFold — 기본값 유지가 정답인 운영 한도 필드(루프 방지 3종·
// 완성 루프 반복)가 workflow 패널의 <details class="panel__advanced"> 안으로
// 내려가고, 판단 축인 두 모드 필드는 접기 밖(보임 영역)에 남는지 본다. 접혀도
// 필드는 같은 폼 안에 있어야 한다 — atomic Save 계약 (settings_shell.go 상단 주석).
func TestWorkflowAdvancedFold(t *testing.T) {
	body := renderIndexBody(t, profile.ProfilePreferences{UserName: "jline"})

	if n := strings.Count(body, `class="panel__advanced"`); n != 1 {
		t.Fatalf("panel__advanced rendered %d times, want 1 (workflow only)", n)
	}
	workflow := panelBlockFor(t, body, "workflow")
	detailsIdx := strings.Index(workflow, `<details class="panel__advanced">`)
	if detailsIdx < 0 {
		t.Fatalf("workflow panel has no panel__advanced details block")
	}
	before := strings.Index(workflow, `name="workflow.default_mode"`)
	after := strings.Index(workflow, `name="workflow.loop_prevention.max_iterations"`)
	if before < 0 || before >= detailsIdx {
		t.Errorf("workflow.default_mode must stay visible (before the fold); index=%d details=%d", before, detailsIdx)
	}
	if after >= 0 && after < detailsIdx {
		t.Errorf("workflow.loop_prevention.max_iterations must render inside the fold; index=%d details=%d", after, detailsIdx)
	}

	// 접힌 필드도 렌더는 된다 — 미렌더는 미제출로 읽혀 저장 계약이 깨진다.
	for _, name := range []string{
		"workflow.agentic_loop.max_iterations",
		"workflow.loop_prevention.failure_pattern_detection",
		"workflow.loop_prevention.max_retries_per_operation",
	} {
		if !strings.Contains(workflow, `name="`+name+`"`) {
			t.Errorf("folded field %q missing from rendered workflow panel", name)
		}
	}
}

// subnavRowFor 는 렌더 HTML 에서 지정 탭의 subnav 행 조각을 잘라 돌려준다.
func subnavRowFor(t *testing.T, body, tabID string) string {
	t.Helper()
	marker := `data-tab="` + tabID + `"`
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("subnav row for %q not found", tabID)
	}
	start := strings.LastIndex(body[:i], "<a ")
	end := i + len(marker) + strings.Index(body[i+len(marker):], "</a>")
	return body[start:end]
}

// panelBlockFor 는 지정 패널의 data-panel 블록을 잘라 돌려준다 (다음 tabpanel 경계까지).
func panelBlockFor(t *testing.T, body, panelID string) string {
	t.Helper()
	marker := `data-panel="` + panelID + `"`
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("panel %q not found", panelID)
	}
	rest := body[i:]
	if j := strings.Index(rest[1:], `data-panel="`); j >= 0 {
		rest = rest[:j+1]
	}
	return rest
}
