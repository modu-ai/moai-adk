// Package pr783_head은 PR #783 이후 상태를 나타냅니다 (BOTH data + test 번역됨).
// Package pr783_head represents the state after PR #783 translation (BOTH data + test translated).
package pr783_head_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// mockReleaseData는 릴리스 오류 메시지 맵입니다 (영어 번역 버전).
// mockReleaseData is the release error message map (English translated version).
var mockReleaseData = map[string]string{
	"yaml_error": "Not a valid YAML document",
}

// TestReleaseFlow는 YAML 파싱 오류 메시지를 검증합니다.
func TestReleaseFlow(t *testing.T) {
	assert.Equal(t, "Not a valid YAML document", mockReleaseData["yaml_error"])
}
