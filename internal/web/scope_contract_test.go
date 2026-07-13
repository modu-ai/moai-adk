package web

// SPEC-WEB-CONSOLE-011 M1 guard tests (AC-WC11-001/002).
//
// 구 scope contract — REQ-WC-012(server.go 패키지 doc) + REQ-WC3-007
// (projectconfig.go @MX:WARN)의 "workflow/harness/git-strategy/llm 절대 금지" —
// 는 REQ-WC11-001이 10섹션 계약으로 공식 SUPERSEDE했다. 본 파일의 guard 2종은
// 신규 계약을 기계적으로 고정한다: 10개 user-facing 섹션 허용(경로 분류 포함) +
// 제외군(machine/state 섹션, 대형 정책 파일, 미지명 섹션) 거부.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings"
)

// TestScopeContractEditableSections는 8개 user-facing 섹션(git-strategy, llm,
// workflow, harness, ralph, feedback, observability, security)이
// 전부 편집 가능 경로로 라우팅되고, 그 경로 분류가 design.md §A.3 표와 일치함을
// 검증한다 (AC-WC11-002 허용 케이스; research는 SPEC-WEB-CONSOLE-012 M1에서
// 폐선 — REQ-WC12-010, 제외군 테스트로 이동).
func TestScopeContractEditableSections(t *testing.T) {
	t.Parallel()

	// typed 경로 2종: git-strategy(dirty-flag Save), llm(typed oneof 검증).
	for _, name := range []string{"git-strategy", "llm"} {
		if got := settings.RouteForSection(name); got != settings.RouteTypedSave {
			t.Errorf("section %q: route = %d, want RouteTypedSave", name, got)
		}
	}

	// SPEC-WEBCONF-SIMPLIFY-001 M3: 8 former seam sections reclassified to
	// RouteExcluded (tabs removed, web write path gone, config keys persist in
	// baked template YAML — REQ-WC-003).
	for _, name := range []string{
		"workflow", "harness", "ralph",
		"feedback", "observability", "security",
		"handoff", "cache",
	} {
		if got := settings.RouteForSection(name); got != settings.RouteExcluded {
			t.Errorf("section %q: route = %d, want RouteExcluded (M3 reclassified)", name, got)
		}
	}

	// 기존 34-필드 typed 커버리지는 계약 개정 후에도 유지된다.
	for _, name := range []string{"user", "language", "quality", "git-convention"} {
		if got := settings.RouteForSection(name); got != settings.RouteTypedSave {
			t.Errorf("legacy typed section %q: route = %d, want RouteTypedSave", name, got)
		}
	}

	// statusline은 기존 전용 동기화 경로를 유지한다 (seam 대상 아님).
	if got := settings.RouteForSection("statusline"); got != settings.RouteStatusline {
		t.Errorf("statusline route = %d, want RouteStatusline", got)
	}
}

// TestScopeContractExclusions는 제외군 — machine/state 섹션(state, system,
// project, cache, sunset), 대형 정책 파일(tool-policy, lsp, mx), 미지명 섹션
// (constitution, context, design, interview), 폐선 섹션(db, research) 및 임의
// 미등재 이름 — 이 전부 편집 불가(RouteExcluded)로 거부됨을 검증한다
// (AC-WC11-002 거부 케이스, REQ-WC11-018; research 폐선은 REQ-WC12-010/012).
func TestScopeContractExclusions(t *testing.T) {
	t.Parallel()

	// cache는 SPEC-WEB-CONSOLE-013 M1에서 seam-writable로 재분류 (REQ-WC13-001 —
	// REQ-WC11-018의 cache 한정 부분 supersede). 잔여 제외군은 전원 유지.
	excluded := []string{
		"state", "system", "project", "sunset",
		"tool-policy", "lsp", "mx",
		"constitution", "context", "design", "interview",
		"db", "research",
		// SPEC-WEBCONF-SIMPLIFY-001 M3: 8 former seam sections reclassified.
		"workflow", "harness", "ralph", "feedback", "observability", "security", "handoff", "cache",
	}
	for _, name := range excluded {
		if got := settings.RouteForSection(name); got != settings.RouteExcluded {
			t.Errorf("excluded section %q: route = %d, want RouteExcluded", name, got)
		}
	}

	// 미등재 임의 이름도 기본 거부 (zero-value 방어).
	for _, name := range []string{"", "nonexistent", "llm2", "Workflow"} {
		if got := settings.RouteForSection(name); got != settings.RouteExcluded {
			t.Errorf("unknown section %q: route = %d, want RouteExcluded", name, got)
		}
	}
}
