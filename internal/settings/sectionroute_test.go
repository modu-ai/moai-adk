package settings

import "testing"

// TestRouteForSectionTable은 라우팅 테이블의 분류(typed/seam/statusline/excluded)를
// 전수 검증한다. Issue 3이 workflow를 RouteSeam으로 부분 복구했고,
// SPEC-FEEDBACK-AUTO-SUBMIT-001 M7이 feedback을 RouteSeam으로 재개방했다
// (auto_submit 토글 — SPEC-WEBCONF-SIMPLIFY-001 M3 결정 반전). 나머지 6개 전
// seam 섹션은 RouteExcluded로 잔류한다.
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
		// seam (reopened — SPEC-FEEDBACK-AUTO-SUBMIT-001 M7, reversing
		// SPEC-WEBCONF-SIMPLIFY-001 M3 for feedback only).
		"feedback": RouteSeam,
		// 기존 전용 경로.
		"statusline": RouteStatusline,
		// SPEC-WEBCONF-SIMPLIFY-001 M3: 6 former seam sections still reclassified
		// to RouteExcluded (tabs removed, config keys persist — REQ-WC-003).
		// feedback left this group in SPEC-FEEDBACK-AUTO-SUBMIT-001 M7.
		"harness":       RouteExcluded,
		"ralph":         RouteExcluded,
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

// TestSeamSectionsMatchesRoutes는 SeamSections 열거가 workflow와 feedback을
// 포함함을 검증한다 (Issue 3 — workflow 복구; SPEC-FEEDBACK-AUTO-SUBMIT-001 M7 —
// feedback 재개방). 나머지 6개는 RouteExcluded로 잔류한다.
func TestSeamSectionsMatchesRoutes(t *testing.T) {
	t.Parallel()
	seam := SeamSections()
	want := map[string]bool{"workflow": true, "feedback": true}
	if len(seam) != len(want) {
		t.Fatalf("SeamSections() length = %d, want %d (workflow + feedback)", len(seam), len(want))
	}
	for _, name := range seam {
		if !want[name] {
			t.Errorf("SeamSections() carries unexpected section %q", name)
		}
		delete(want, name)
	}
	for name := range want {
		t.Errorf("SeamSections() missing section %q", name)
	}
	// 두 섹션 모두 RouteSeam으로 라우팅되어야 한다.
	for _, name := range []string{"workflow", "feedback"} {
		if got := RouteForSection(name); got != RouteSeam {
			t.Errorf("%s routes to %d, want RouteSeam", name, got)
		}
	}
	// 나머지 6개 전 seam 섹션은 여전히 RouteExcluded.
	for _, name := range []string{"harness", "ralph", "observability", "security", "handoff", "cache"} {
		if got := RouteForSection(name); got != RouteExcluded {
			t.Errorf("former seam section %q routes to %d, want RouteExcluded (M3)", name, got)
		}
	}
}

// TestExcludedSectionsAllRejected는 명시적 제외군 전원이 RouteExcluded임을 검증한다.
// Issue 3이 workflow를, SPEC-FEEDBACK-AUTO-SUBMIT-001 M7이 feedback을 제외군에서
// 제거했다 (기존 19 - 2 = 17).
func TestExcludedSectionsAllRejected(t *testing.T) {
	t.Parallel()
	excluded := ExcludedSections()
	if len(excluded) != 17 {
		t.Fatalf("ExcludedSections() length = %d, want 17 (19 M3 - workflow - feedback)", len(excluded))
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
	// feedback도 제외군에 없어야 한다 (M7 재개방).
	if present["feedback"] {
		t.Error("feedback must NOT be in ExcludedSections() (reopened to RouteSeam — M7)")
	}
	// 나머지 6개 M3 섹션은 제외군에 명시되어 있어야 한다.
	for _, name := range []string{"harness", "ralph", "observability", "security", "handoff", "cache"} {
		if !present[name] {
			t.Errorf("M3-reclassified section %q missing from ExcludedSections()", name)
		}
	}
}
