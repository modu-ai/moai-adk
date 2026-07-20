package profile

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/statusline"
)

// TestSegmentListSSOT는 profile 의 정규 세그먼트 시드 맵(defaultStatuslineSegments)
// 키 집합이 SSOT 인 statusline.CanonicalSegments 와 정확히 일치함을 검증한다.
// 세그먼트가 CanonicalSegments 에만 추가되고 이 시드 맵에서 누락되면(또는 그 반대)
// 테스트가 실패하여 3-way orphan 재발을 차단한다(REQ-WC11-051, design.md §F).
//
// settings / TUI(cli) 목록은 각 패키지의 동명 TestSegmentListSSOT 가 담당하며,
// 셋 다 statusline.CanonicalSegments 를 SSOT 앵커로 사용하므로 네 목록이
// 이행적으로 집합-동일해진다.
func TestSegmentListSSOT(t *testing.T) {
	got := defaultStatuslineSegments()

	want := statusline.CanonicalSegments
	wantSet := make(map[string]bool, len(want))
	for _, s := range want {
		wantSet[s] = true
	}

	if len(got) != len(wantSet) {
		t.Errorf("profile.defaultStatuslineSegments has %d keys, SSOT statusline.CanonicalSegments has %d", len(got), len(wantSet))
	}
	for s := range wantSet {
		if _, ok := got[s]; !ok {
			t.Errorf("profile.defaultStatuslineSegments: MISSING SSOT segment %q (present in CanonicalSegments but orphaned)", s)
		}
	}
	for s := range got {
		if !wantSet[s] {
			t.Errorf("profile.defaultStatuslineSegments: EXTRA segment %q (not in SSOT statusline.CanonicalSegments)", s)
		}
	}
}
