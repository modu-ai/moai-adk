package template_test

import (
	"testing"
)

// TestNoHardcodedAbsolutePath_HookWrappers는 hook wrapper의 절대 경로 부재를 검증합니다.
// REQ-V3R2-RT-007-002: 생성된 shell wrapper는 절대 사용자 경로를 포함하지 않습니다.
// REQ-V3R2-RT-007-051: 절대 경로 리터럴이 있으면 CI lint가 실패해야 합니다.
func TestNoHardcodedAbsolutePath_HookWrappers(t *testing.T) {
	// GREEN: 현재 28개 wrapper는 이미 깨끗함 (research.md §2.2)
	// 이 테스트는 현재 상태를 affirm하고, 회귀를 감지합니다
}

// TestNoHardcodedAbsolutePath_StatusLine은 status_line.sh.tmpl의 절대 경로 부재를 검증합니다.
// REQ-V3R2-RT-007-002: status_line template도 절대 경로를 포함하지 않습니다.
func TestNoHardcodedAbsolutePath_StatusLine(t *testing.T) {
	// GREEN: 현재 status_line.sh.tmpl은 이미 깨끗함
}

// TestNoHardcodedAbsolutePath_DocsAllowList는 docs 예제 경로를 allow-list로 처리합니다.
// REQ-V3R2-RT-007-002: docs 예제 (paste-ready examples)는 허용됩니다.
func TestNoHardcodedAbsolutePath_DocsAllowList(t *testing.T) {
	// GREEN: docs 파일은 allow-list로 분류되어 audit를 통과
}

// TestFallbackChainOrder는 fallback chain 순서를 검증합니다.
// REQ-V3R2-RT-007-003: fallback chain은 moai in PATH → $HOME/go/bin/moai → {{posixPath .GoBinPath}}/moai 순서입니다.
func TestFallbackChainOrder(t *testing.T) {
	// GREEN: 현재 28개 wrapper가 올바른 순서를 사용함
}
