package settings

import "testing"

// TestRouteForSectionTable은 라우팅 테이블의 분류(typed/seam/statusline/excluded)를
// 전수 검증한다. Issue 3이 workflow를 RouteSeam으로 부분 복구했다 (worktree
// auto-create 토글). 나머지 7개 전 seam 섹션은 RouteExcluded로 잔류한다.
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
		// seam (restored — Issue 3).
		"workflow": RouteSeam,
		// 기존 전용 경로.
		"statusline": RouteStatusline,
		// SPEC-WEBCONF-SIMPLIFY-001 M3: 7 former seam sections still reclassified
		// to RouteExcluded (tabs removed, config keys persist — REQ-WC-003).
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

// TestSeamSectionsMatchesRoutes는 SeamSections 열거가 workflow만 포함함을 검증한다
// (Issue 3 — workflow가 RouteSeam으로 복구되었고, 나머지 7개는 RouteExcluded).
func TestSeamSectionsMatchesRoutes(t *testing.T) {
	t.Parallel()
	seam := SeamSections()
	if len(seam) != 1 {
		t.Fatalf("SeamSections() length = %d, want 1 (workflow restored — Issue 3)", len(seam))
	}
	if seam[0] != "workflow" {
		t.Errorf("SeamSections()[0] = %q, want %q", seam[0], "workflow")
	}
	// workflow는 RouteSeam으로 라우팅되어야 한다.
	if got := RouteForSection("workflow"); got != RouteSeam {
		t.Errorf("workflow routes to %d, want RouteSeam (Issue 3 restored)", got)
	}
	// 나머지 7개 전 seam 섹션은 여전히 RouteExcluded.
	for _, name := range []string{"harness", "ralph", "feedback", "observability", "security", "handoff", "cache"} {
		if got := RouteForSection(name); got != RouteExcluded {
			t.Errorf("former seam section %q routes to %d, want RouteExcluded (M3)", name, got)
		}
	}
}

// TestExcludedSectionsAllRejected는 명시적 제외군 전원이 RouteExcluded임을 검증한다.
// Issue 3이 workflow를 제외군에서 제거했다 (기존 19 - 1 = 18).
func TestExcludedSectionsAllRejected(t *testing.T) {
	t.Parallel()
	excluded := ExcludedSections()
	if len(excluded) != 18 {
		t.Fatalf("ExcludedSections() length = %d, want 18 (19 M3 - workflow restored — Issue 3)", len(excluded))
	}
	for _, name := range excluded {
		if got := RouteForSection(name); got != RouteExcluded {
			t.Errorf("excluded section %q routes to %d, want RouteExcluded", name, got)
		}
	}
	// workflow는 제외군에 없어야 한다 (Issue 3 복구).
	present := map[string]bool{}
	for _, name := range excluded {
		present[name] = true
	}
	if present["workflow"] {
		t.Error("workflow must NOT be in ExcludedSections() (restored to RouteSeam — Issue 3)")
	}
	// 나머지 7개 M3 섹션은 제외군에 명시되어 있어야 한다.
	for _, name := range []string{"harness", "ralph", "feedback", "observability", "security", "handoff", "cache"} {
		if !present[name] {
			t.Errorf("M3-reclassified section %q missing from ExcludedSections()", name)
		}
	}
}
