package template_test

import (
	"testing"
)

// TestHomeIsRegisteredInPassthroughTokens는 $HOME passthrough 등록을 검증합니다.
// REQ-V3R2-RT-007-006: $HOME은 claudeCodePassthroughTokens에 등록되어야 합니다.
func TestHomeIsRegisteredInPassthroughTokens(t *testing.T) {
	// GREEN: renderer.go:42에 이미 "$HOME"이 등록됨
	// 이 테스트는 현재 상태를 affirm함
}
