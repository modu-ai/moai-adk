package mx

import (
	"context"
)

// FanInCounter는 @MX:ANCHOR 태그의 코드 참조 수를 계산하는 인터페이스입니다.
// LSP 기반 구현과 텍스트 폴백 구현을 동일한 인터페이스로 추상화합니다 (REQ-SPC-004-003).
type FanInCounter interface {
	// Count는 주어진 태그의 fan-in(코드 참조 수)을 계산합니다.
	// excludeTests가 true이면 테스트 파일의 참조는 제외합니다 (REQ-SPC-004-040).
	// 반환값: count (참조 수), method ("lsp" 또는 "textual"), err
	Count(ctx context.Context, tag Tag, projectRoot string, excludeTests bool) (count int, method string, err error)
}

// TextualFanInCounter는 텍스트 기반 grep 방식으로 fan-in을 계산하는 구현체입니다.
// LSP 서버가 없는 언어에 대한 폴백으로 사용됩니다 (REQ-SPC-004-020).
//
// @MX:WARN: [AUTO] TextualFanInCounter — 텍스트 검색 방식은 문자열/주석의 오탐(false positive) 위험이 있습니다
// @MX:REASON: 소스 코드가 아닌 문자열 리터럴이나 주석에서도 심볼 이름이 발견될 수 있어 fan-in이 과대 계산될 수 있습니다
type TextualFanInCounter struct {
	// ProjectRoot는 프로젝트 루트 디렉토리 경로입니다.
	ProjectRoot string
}

// Count는 텍스트 검색을 통해 AnchorID의 참조 수를 계산합니다.
// 결과의 fan_in_method는 항상 "textual"입니다.
func (c *TextualFanInCounter) Count(_ context.Context, tag Tag, projectRoot string, excludeTests bool) (int, string, error) {
	// RED 단계: 미구현 stub
	// GREEN 단계에서 실제 구현으로 교체됩니다
	return 0, "textual", nil
}
