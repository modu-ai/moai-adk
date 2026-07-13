package settings

// 이 파일은 웹 콘솔 편집 가능 섹션의 쓰기 경로 라우팅 SSOT를 담는다
// (SPEC-WEB-CONSOLE-011 M1, REQ-WC11-001/016/017/018 — design.md §A.3 라우팅 표).
//
// 구 scope contract(REQ-WC-012 + REQ-WC3-007)의 "workflow/harness/git-strategy/llm
// 절대 금지" 조항은 SPEC-WEB-CONSOLE-011 REQ-WC11-001이 공식 SUPERSEDE했다.
// 각 섹션의 영속화 경로는 아래 라우팅 테이블이 단일 진실 공급원이다.
// machine/state 섹션·대형 정책 파일·미지명 섹션 (db, research 포함 — 콘솔
// 표면에서 제거됨; db는 settings SSOT, research는 SPEC-WEB-CONSOLE-012
// REQ-WC12-010 폐선)은 계속 쓰기 금지다 (REQ-WC11-018). 단 cache는
// SPEC-WEB-CONSOLE-013 REQ-WC13-001이 REQ-WC11-018을 cache 한정으로 부분
// SUPERSEDE하여 user-facing seam-writable 섹션으로 재분류되었다 (노출 스코프는
// cacheStrategy.enabled + session_ttl 2키 — REQ-WC13-012/015).

// SectionRoute는 config 섹션 파일(.moai/config/sections/<name>.yaml)의 웹 콘솔
// 영속화 경로 분류다.
type SectionRoute int

const (
	// RouteExcluded — 웹 콘솔에서 편집 불가. machine/state 섹션(state, system,
	// project, sunset), 대형 정책 파일(tool-policy, lsp, mx), 미지명 섹션
	// (constitution, context, design, interview 및 그 외 전부)이 여기 속한다
	// (REQ-WC11-018; cache는 SPEC-WEB-CONSOLE-013 REQ-WC13-001에서 제외군을
	// 이탈해 RouteSeam으로 재분류). map 미등재 이름의 zero value이기도 하다 —
	// 새 섹션은 명시적으로 등재되기 전까지 기본 거부된다.
	RouteExcluded SectionRoute = iota

	// RouteTypedSave — config.ConfigManager typed Save 경로로 영속화한다.
	// 완전 typed 섹션만 해당: 기존 34-필드 커버리지(user, language, quality,
	// git-convention) + git-strategy(dirty-flag 경로, SPEC-GITSTRATEGY-SAVE-
	// ISOLATION-001) + llm(oneof 검증 typed 경로; mode/team_mode는 REQ-WC11-013
	// 에 따라 read-only).
	RouteTypedSave

	// RouteSeam — yamlpatch seam(internal/settings/yamlpatch) 전용으로 영속화한다.
	// typed Save() 경로가 없는 8개 섹션: workflow, harness, ralph,
	// feedback, observability, security (REQ-WC11-017) + handoff, cache
	// (SPEC-WEB-CONSOLE-013 REQ-WC13-002/005 — typed HandoffConfig/CacheConfig
	// struct는 read-side 전용 유지). 특히 workflow.yaml은 typed re-marshal이
	// 금지된다 (REQ-WC11-005 — team.patterns 미모델링 + role-profile effort
	// Go-invisible).
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

	// 10섹션 계약 중 typed 경로 (REQ-WC11-010/012) — surviving web-writable.
	"git-strategy": RouteTypedSave,
	"llm":          RouteTypedSave,

	// statusline은 기존 전용 경로 유지 (M6은 노출 fan-out만).
	"statusline": RouteStatusline,

	// report — seam-writable (report.format select, moai-domain-html-report output).
	"report": RouteSeam,

	// SPEC-WEBCONF-SIMPLIFY-001 M3: the 8 former seam sections (workflow, harness,
	// ralph, feedback, observability, security, handoff, cache) are reclassified
	// to RouteExcluded — their tabs are removed and their web write path is gone.
	// They are absent from this map, so RouteForSection returns the zero value
	// (RouteExcluded). Their config keys persist in the baked template YAML for
	// runtime consumption (REQ-WC-003). project/mx were already RouteExcluded
	// (zero-value); quality_extras shares the "quality" file (RouteTypedSave —
	// main SectionQuality D12-unaffected; M4 adds the enable/disable toggle).
}

// RouteForSection은 섹션 파일 base name의 웹 콘솔 쓰기 경로를 반환한다.
// 미등재 이름(제외군 포함 전부)은 RouteExcluded다.
func RouteForSection(name string) SectionRoute {
	return sectionRoutes[name]
}

// SeamSections는 yamlpatch seam이 유일한 쓰기 경로인 섹션을 반환한다.
// SPEC-WEBCONF-SIMPLIFY-001 M3가 기존 8개 seam 섹션을 전부 RouteExcluded로
// 재분류하여 현재 이 목록은 비어 있다. seam 메커니즘(yamlpatch.PatchFile) 자체는
// 잔존하며, 향후 SPEC이 seam 섹션을 재추가하면 여기에 다시 열거한다.
func SeamSections() []string {
	return []string{}
}

// ExcludedSections는 명시적 제외군을 반환한다: machine/state 섹션 + 대형 정책
// 파일 + 미지명 잔여 섹션 + SPEC-WEBCONF-SIMPLIFY-001 M3로 재분류된 8개 전
// seam 섹션. RouteForSection은 이 목록 외의 임의 미등재 이름에도 동일하게
// RouteExcluded를 반환한다 (zero value).
func ExcludedSections() []string {
	return []string{
		// machine/state 섹션.
		"state", "system", "project", "sunset",
		// 대형 정책 파일.
		"tool-policy", "lsp", "mx",
		// 미지명 잔여 섹션.
		"constitution", "context", "design", "interview",
		// SPEC-WEBCONF-SIMPLIFY-001 M3: 8 former seam sections reclassified to
		// RouteExcluded (tabs removed, web write path gone, config keys persist
		// in baked template YAML for runtime consumption — REQ-WC-003).
		"workflow", "harness", "ralph", "feedback", "observability", "security", "handoff", "cache",
	}
}
