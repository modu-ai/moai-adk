package cli

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/statusline"
)

// TestSegmentListSSOT는 TUI MultiSelect 가 제공하는 statuslineAllSegments 목록이
// SSOT 인 statusline.CanonicalSegments 와 집합으로 정확히 일치함을 검증한다.
// 세그먼트가 CanonicalSegments 에만 추가되고 TUI 목록에서 누락되면(또는 그 반대)
// 테스트가 실패하여, renderer 에는 있으나 노출 목록에는 없던 cache_hit 류의
// 3-way orphan 재발을 차단한다(REQ-WC11-051, design.md §F).
//
// settings / profile 목록은 각 패키지의 동명 TestSegmentListSSOT 가 담당한다
// (cli→settings import 사이클 때문에 하나의 중앙 테스트로 통합할 수 없어 패키지별
// white-box 로 분리; 셋 다 동일한 statusline.CanonicalSegments 를 SSOT 앵커로 사용).
//
// 추가 불변식: MultiSelect Options 목록(profile_setup.go)과 statuslineAllSegments
// 는 KEEP-IN-SYNC 여야 하므로, 두 목록이 어긋나면 이 테스트가 함께 실패한다.
func TestSegmentListSSOT(t *testing.T) {
	want := statusline.CanonicalSegments
	wantSet := make(map[string]bool, len(want))
	for _, s := range want {
		wantSet[s] = true
	}

	gotSet := make(map[string]bool, len(statuslineAllSegments))
	for _, s := range statuslineAllSegments {
		if gotSet[s] {
			t.Errorf("cli.statuslineAllSegments: duplicate segment %q", s)
		}
		gotSet[s] = true
	}

	if len(gotSet) != len(wantSet) {
		t.Errorf("cli.statuslineAllSegments has %d unique segments, SSOT statusline.CanonicalSegments has %d", len(gotSet), len(wantSet))
	}
	for s := range wantSet {
		if !gotSet[s] {
			t.Errorf("cli.statuslineAllSegments: MISSING SSOT segment %q (present in CanonicalSegments but orphaned)", s)
		}
	}
	for s := range gotSet {
		if !wantSet[s] {
			t.Errorf("cli.statuslineAllSegments: EXTRA segment %q (not in SSOT statusline.CanonicalSegments)", s)
		}
	}
}
