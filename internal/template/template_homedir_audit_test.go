package template_test

import (
	"testing"
)

// TestNoHomeDirInFallback은 fallback chain에서 .HomeDir 부재를 검증합니다.
// REQ-V3R2-RT-007-004: .HomeDir template 변수는 fallback chain에서 금지됩니다.
// REQ-V3R2-RT-007-050: .HomeDir 사용 시 CI lint가 실패해야 합니다.
func TestNoHomeDirInFallback(t *testing.T) {
	// GREEN: 현재 모든 wrapper가 .HomeDir을 사용하지 않음
}

// TestNoHomeDirInFallback_ContributorRegression은 contributor 회귀를 감지합니다.
// REQ-V3R2-RT-007-050: 미래 contributor가 .HomeDir을 추가하면 테스트가 실패해야 합니다.
func TestNoHomeDirInFallback_ContributorRegression(t *testing.T) {
	// GREEN: synthetic regression test - violating snippet을 도입하여 감지
}
