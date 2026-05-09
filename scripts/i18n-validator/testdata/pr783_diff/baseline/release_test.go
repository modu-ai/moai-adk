// Package pr783_baseline은 PR #783 기준선 상태를 나타냅니다.
// Package pr783_baseline represents the baseline state before PR #783 translation.
package pr783_baseline_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// mockReleaseData는 릴리스 오류 메시지 맵입니다 (한국어 버전).
// mockReleaseData is the release error message map (Korean version).
var mockReleaseData = map[string]string{
	"yaml_error": "유효한 YAML 문서가 아닙니다",
}

// TestReleaseFlow는 YAML 파싱 오류 메시지를 검증합니다.
func TestReleaseFlow(t *testing.T) {
	assert.Equal(t, "유효한 YAML 문서가 아닙니다", mockReleaseData["yaml_error"])
}
