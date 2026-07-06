package settings

// 이 파일은 웹 콘솔 편집 가능 섹션의 쓰기 경로 라우팅 SSOT를 담는다
// (SPEC-WEB-CONSOLE-011 M1, REQ-WC11-001/016/017/018 — design.md §A.3 라우팅 표).
//
// 구 scope contract(REQ-WC-012 + REQ-WC3-007)의 "workflow/harness/git-strategy/llm
// 절대 금지" 조항은 SPEC-WEB-CONSOLE-011 REQ-WC11-001이 공식 SUPERSEDE했다. 이제
// 9개 user-facing 섹션이 편집 가능하며, 각 섹션의 영속화 경로는 아래 라우팅
// 테이블이 단일 진실 공급원이다. machine/state 섹션·대형 정책 파일·미지명 섹션
// (db 포함 — 콘솔 표면에서 제거됨, settings SSOT)은 계속 쓰기 금지다 (REQ-WC11-018).

// SectionRoute는 config 섹션 파일(.moai/config/sections/<name>.yaml)의 웹 콘솔
// 영속화 경로 분류다.
type SectionRoute int

const (
	// RouteExcluded — 웹 콘솔에서 편집 불가. machine/state 섹션(state, system,
	// project, cache, sunset), 대형 정책 파일(tool-policy, lsp, mx), 미지명 섹션
	// (constitution, context, design, interview 및 그 외 전부)이 여기 속한다
	// (REQ-WC11-018). map 미등재 이름의 zero value이기도 하다 — 새 섹션은
	// 명시적으로 등재되기 전까지 기본 거부된다.
	RouteExcluded SectionRoute = iota

	// RouteTypedSave — config.ConfigManager typed Save 경로로 영속화한다.
	// 완전 typed 섹션만 해당: 기존 34-필드 커버리지(user, language, quality,
	// git-convention) + git-strategy(dirty-flag 경로, SPEC-GITSTRATEGY-SAVE-
	// ISOLATION-001) + llm(oneof 검증 typed 경로; mode/team_mode는 REQ-WC11-013
	// 에 따라 read-only).
	RouteTypedSave

	// RouteSeam — yamlpatch seam(internal/settings/yamlpatch) 전용으로 영속화한다.
	// typed Save() 경로가 없는 7개 섹션: workflow, harness, ralph, research,
	// feedback, observability, security (REQ-WC11-017). 특히 workflow.yaml은
	// typed re-marshal이 금지된다 (REQ-WC11-005 — team.patterns 미모델링 +
	// role-profile effort Go-invisible).
	RouteSeam

	// RouteStatusline — statusline은 기존 전용 동기화 경로(internal/profile
	// sync.go 직접 marshal)를 유지한다. 본 SPEC의 seam 대상이 아니다
	// (design.md §A.3 — 경로 재설계는 범위 밖).
	RouteStatusline
)

// sectionRoutes는 섹션 파일 base name(확장자 제외, 파일명 그대로 dash 표기) →
// 쓰기 경로 매핑이다. 여기 없는 이름은 전부 RouteExcluded (zero value).
var sectionRoutes = map[string]SectionRoute{
	// 기존 typed 커버리지 (SPEC-WEB-CONSOLE-003/007/010 확립).
	"user":           RouteTypedSave,
	"language":       RouteTypedSave,
	"quality":        RouteTypedSave,
	"git-convention": RouteTypedSave,

	// 10섹션 계약 중 typed 경로 (REQ-WC11-010/012).
	"git-strategy": RouteTypedSave,
	"llm":          RouteTypedSave,

	// 10섹션 계약 중 seam 전용 7섹션 (REQ-WC11-017). db는 콘솔 표면에서 제거되어
	// (settings SSOT) 더 이상 seam-writable이 아니다 — 미등재 → RouteExcluded.
	"workflow":      RouteSeam,
	"harness":       RouteSeam,
	"ralph":         RouteSeam,
	"research":      RouteSeam,
	"feedback":      RouteSeam,
	"observability": RouteSeam,
	"security":      RouteSeam,

	// statusline은 기존 전용 경로 유지 (M6은 노출 fan-out만).
	"statusline": RouteStatusline,
}

// RouteForSection은 섹션 파일 base name의 웹 콘솔 쓰기 경로를 반환한다.
// 미등재 이름(제외군 포함 전부)은 RouteExcluded다.
func RouteForSection(name string) SectionRoute {
	return sectionRoutes[name]
}

// SeamSections는 yamlpatch seam이 유일한 쓰기 경로인 7개 섹션을 반환한다
// (REQ-WC11-017). 순서는 spec.md §1.1 결정 3의 열거 순서를 따른다 (db는 콘솔
// 표면에서 제거됨 — settings SSOT).
func SeamSections() []string {
	return []string{"workflow", "harness", "ralph", "research", "feedback", "observability", "security"}
}

// ExcludedSections는 명시적 제외군(REQ-WC11-018)을 반환한다: machine/state 섹션 +
// 대형 정책 파일 + 미지명 잔여 섹션(2026-07-03 실측 4종). RouteForSection은 이
// 목록 외의 임의 미등재 이름에도 동일하게 RouteExcluded를 반환한다.
func ExcludedSections() []string {
	return []string{
		// machine/state 섹션.
		"state", "system", "project", "cache", "sunset",
		// 대형 정책 파일.
		"tool-policy", "lsp", "mx",
		// 미지명 잔여 섹션.
		"constitution", "context", "design", "interview",
	}
}
