package settings

import "testing"

// TestRouteForSectionTable은 라우팅 테이블의 분류(typed/statusline/excluded)를
// 전수 검증한다. SPEC-WEBCONF-SIMPLIFY-001 M3가 기존 8개 seam 섹션을
// RouteExcluded로 재분류하여 현재 RouteSeam에 등록된 섹션은 없다
// (REQ-WC-003 — config keys persist, web write path removed).
func TestRouteForSectionTable(t *testing.T) {
	t.Parallel()

	cases := map[string]SectionRoute{
		// typed (surviving web-writable).
		"user":           RouteTypedSave,
		"language":       RouteTypedSave,
		"quality":        RouteTypedSave,
		"git-convention": RouteTypedSave,
		"git-strategy":   RouteTypedSave,
		"llm":            RouteTypedSave,
		// 기존 전용 경로.
		"statusline": RouteStatusline,
		// SPEC-WEBCONF-SIMPLIFY-001 M3: 8 former seam sections reclassified to
		// RouteExcluded (tabs removed, config keys persist — REQ-WC-003).
		"workflow":      RouteExcluded,
		"harness":       RouteExcluded,
		"ralph":         RouteExcluded,
		"feedback":      RouteExcluded,
		"observability": RouteExcluded,
		"security":      RouteExcluded,
		"handoff":       RouteExcluded,
		"cache":         RouteExcluded,
		// 기존 제외군 + 미등재 (zero-value RouteExcluded).
		"state":       RouteExcluded,
		"sunset":      RouteExcluded,
		"tool-policy": RouteExcluded,
		"mx":          RouteExcluded,
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

// TestSeamSectionsMatchesRoutes는 SeamSections 열거가 M3 이후 비어 있음을 검증한다
// (SPEC-WEBCONF-SIMPLIFY-001 M3 — 8개 seam 섹션이 전부 RouteExcluded로
// 재분류되어 현재 seam-writable 섹션은 없다).
func TestSeamSectionsMatchesRoutes(t *testing.T) {
	t.Parallel()
	seam := SeamSections()
	if len(seam) != 0 {
		t.Fatalf("SeamSections() length = %d, want 0 (M3 reclassified all 8 seam sections to RouteExcluded)", len(seam))
	}
	// M3 이전에 seam이었던 8개 섹션이 전부 RouteExcluded로 라우팅되는지 확인.
	for _, name := range []string{"workflow", "harness", "ralph", "feedback", "observability", "security", "handoff", "cache"} {
		if got := RouteForSection(name); got != RouteExcluded {
			t.Errorf("former seam section %q routes to %d, want RouteExcluded (M3)", name, got)
		}
	}
}

// TestExcludedSectionsAllRejected는 명시적 제외군 전원이 RouteExcluded임을 검증한다.
// SPEC-WEBCONF-SIMPLIFY-001 M3가 8개 전 seam 섹션을 제외군에 추가했다
// (기존 11 + M3 8 = 19).
func TestExcludedSectionsAllRejected(t *testing.T) {
	t.Parallel()
	excluded := ExcludedSections()
	if len(excluded) != 19 {
		t.Fatalf("ExcludedSections() length = %d, want 19 (11 original + 8 M3-reclassified)", len(excluded))
	}
	for _, name := range excluded {
		if got := RouteForSection(name); got != RouteExcluded {
			t.Errorf("excluded section %q routes to %d, want RouteExcluded", name, got)
		}
	}
	// M3로 추가된 8개 섹션이 제외군에 명시되어 있는지 확인.
	present := map[string]bool{}
	for _, name := range excluded {
		present[name] = true
	}
	for _, name := range []string{"workflow", "harness", "ralph", "feedback", "observability", "security", "handoff", "cache"} {
		if !present[name] {
			t.Errorf("M3-reclassified section %q missing from ExcludedSections()", name)
		}
	}
}
