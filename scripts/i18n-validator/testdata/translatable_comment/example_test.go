// Package translatable_comment은 i18n:translatable 매직 코멘트 예시를 제공합니다.
// Package translatable_comment provides example of i18n:translatable magic comment.
package translatable_comment_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// errMsg는 번역 가능한 CLI 메시지입니다.
// errMsg is a translatable CLI message.
const errMsg = "Failed to load config" // i18n:translatable

// TestConfig는 설정 로드 오류 메시지를 검증합니다.
func TestConfig(t *testing.T) {
	assert.Equal(t, errMsg, "Failed to load config")
}
