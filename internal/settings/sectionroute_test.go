package settings

import "testing"

// TestRouteForSectionTable은 라우팅 테이블의 4분류(typed/seam/statusline/excluded)
// 를 전수 검증한다 (REQ-WC11-016/017/018 — design.md §A.3).
func TestRouteForSectionTable(t *testing.T) {
	t.Parallel()

	cases := map[string]SectionRoute{
		// typed.
		"user":           RouteTypedSave,
		"language":       RouteTypedSave,
		"quality":        RouteTypedSave,
		"git-convention": RouteTypedSave,
		"git-strategy":   RouteTypedSave,
		"llm":            RouteTypedSave,
		// seam ×6 (research는 SPEC-WEB-CONSOLE-012 M1에서 폐선 — FieldDef 0개
		// + 콘솔 탭 미등재의 유령 쓰기 경로였다, REQ-WC12-010).
		"workflow":      RouteSeam,
		"harness":       RouteSeam,
		"ralph":         RouteSeam,
		"feedback":      RouteSeam,
		"observability": RouteSeam,
		"security":      RouteSeam,
		// 기존 전용 경로.
		"statusline": RouteStatusline,
		// 제외군 + 미등재 (research/db는 미등재 → RouteExcluded, 동일 선례).
		"state":       RouteExcluded,
		"tool-policy": RouteExcluded,
		"interview":   RouteExcluded,
		"research":    RouteExcluded,
		"db":          RouteExcluded,
		"no-such":     RouteExcluded,
	}
	for name, want := range cases {
		if got := RouteForSection(name); got != want {
			t.Errorf("RouteForSection(%q) = %d, want %d", name, got, want)
		}
	}
}

// TestSeamSectionsMatchesRoutes는 SeamSections 열거와 라우팅 테이블의 정합을
// 검증한다 — 열거에 있는 섹션은 전부 RouteSeam이어야 한다.
func TestSeamSectionsMatchesRoutes(t *testing.T) {
	t.Parallel()
	seam := SeamSections()
	if len(seam) != 6 {
		t.Fatalf("SeamSections() length = %d, want 6", len(seam))
	}
	for _, name := range seam {
		if got := RouteForSection(name); got != RouteSeam {
			t.Errorf("seam section %q routes to %d, want RouteSeam", name, got)
		}
		if name == "research" {
			t.Error("research must not be enumerated as a seam section (REQ-WC12-010)")
		}
	}
}

// TestExcludedSectionsAllRejected는 명시적 제외군 전원이 RouteExcluded임을
// 검증한다 (REQ-WC11-018).
func TestExcludedSectionsAllRejected(t *testing.T) {
	t.Parallel()
	excluded := ExcludedSections()
	if len(excluded) != 12 {
		t.Fatalf("ExcludedSections() length = %d, want 12", len(excluded))
	}
	for _, name := range excluded {
		if got := RouteForSection(name); got != RouteExcluded {
			t.Errorf("excluded section %q routes to %d, want RouteExcluded", name, got)
		}
	}
}
