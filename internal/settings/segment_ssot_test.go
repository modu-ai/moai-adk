package settings

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/statusline"
)

// TestSegmentListSSOT는 settings 의 정규 세그먼트 노출 목록(statuslineSegmentKeys)이
// 단일 진실 원천(SSOT)인 statusline.CanonicalSegments 와 집합으로 정확히 일치함을
// 검증한다. 세그먼트가 CanonicalSegments 에만 추가되고 이 노출 목록에서 누락되면
// (또는 그 반대) 테스트가 실패하여, renderer 에는 있으나 노출 목록에는 없던
// cache_hit 류의 3-way orphan 재발을 컴파일-CI 시점에 차단한다
// (REQ-WC11-051, design.md §F — SSOT = statusline.CanonicalSegments).
//
// TUI(cli) 목록은 cli→settings import 사이클 때문에 이 패키지에서 참조할 수 없어
// internal/cli 의 동명 TestSegmentListSSOT 가 담당하고, profile 목록은
// internal/profile 의 동명 테스트가 담당한다. 세 테스트 모두 동일한
// statusline.CanonicalSegments 를 SSOT 앵커로 사용하므로 네 목록이 이행적으로
// 집합-동일해진다. `go test -run TestSegmentListSSOT ./...` 로 셋 다 실행된다.
func TestSegmentListSSOT(t *testing.T) {
	assertSegmentSetEqualToSSOT(t, "settings.statuslineSegmentKeys", statuslineSegmentKeys())
}

// assertSegmentSetEqualToSSOT는 got 목록이 statusline.CanonicalSegments 와 집합으로
// 동일한지 검증한다(순서 무시; 중복·누락·초과 모두 실패로 보고).
func assertSegmentSetEqualToSSOT(t *testing.T, name string, got []string) {
	t.Helper()
	want := statusline.CanonicalSegments

	wantSet := make(map[string]bool, len(want))
	for _, s := range want {
		wantSet[s] = true
	}
	gotSet := make(map[string]bool, len(got))
	for _, s := range got {
		if gotSet[s] {
			t.Errorf("%s: duplicate segment %q", name, s)
		}
		gotSet[s] = true
	}
	if len(gotSet) != len(wantSet) {
		t.Errorf("%s: has %d unique segments, SSOT statusline.CanonicalSegments has %d", name, len(gotSet), len(wantSet))
	}
	for s := range wantSet {
		if !gotSet[s] {
			t.Errorf("%s: MISSING SSOT segment %q (present in CanonicalSegments but orphaned from this list)", name, s)
		}
	}
	for s := range gotSet {
		if !wantSet[s] {
			t.Errorf("%s: EXTRA segment %q (not in SSOT statusline.CanonicalSegments)", name, s)
		}
	}
}
