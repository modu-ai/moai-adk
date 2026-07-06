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

// TestScopeContractNineSections는 9개 user-facing 섹션(git-strategy, llm,
// workflow, harness, ralph, research, feedback, observability, security)이
// 전부 편집 가능 경로로 라우팅되고, 그 경로 분류가 design.md §A.3 표와 일치함을
// 검증한다 (AC-WC11-002 허용 케이스).
func TestScopeContractNineSections(t *testing.T) {
	t.Parallel()

	// typed 경로 2종: git-strategy(dirty-flag Save), llm(typed oneof 검증).
	for _, name := range []string{"git-strategy", "llm"} {
		if got := settings.RouteForSection(name); got != settings.RouteTypedSave {
			t.Errorf("section %q: route = %d, want RouteTypedSave", name, got)
		}
	}

	// seam 전용 7종: typed Save() 경로 부재 — yamlpatch seam이 유일한 쓰기 경로
	// (REQ-WC11-017; workflow.yaml typed re-marshal 금지 REQ-WC11-005).
	for _, name := range []string{
		"workflow", "harness", "ralph", "research",
		"feedback", "observability", "security",
	} {
		if got := settings.RouteForSection(name); got != settings.RouteSeam {
			t.Errorf("section %q: route = %d, want RouteSeam", name, got)
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
// (constitution, context, design, interview) 및 임의 미등재 이름 — 이 전부
// 편집 불가(RouteExcluded)로 거부됨을 검증한다 (AC-WC11-002 거부 케이스,
// REQ-WC11-018).
func TestScopeContractExclusions(t *testing.T) {
	t.Parallel()

	excluded := []string{
		"state", "system", "project", "cache", "sunset",
		"tool-policy", "lsp", "mx",
		"constitution", "context", "design", "interview",
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
