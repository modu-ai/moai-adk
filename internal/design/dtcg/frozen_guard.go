package dtcg

import (
	"fmt"
	"strings"
)

// frozenPrefixes: FROZEN 영역 파일 경로 접두사 목록.
// [HARD] 소스 코드에 하드코딩 - config 우회 불가.
// REQ-DPL-011: 설계 헌법 §2, §3.1-3.3, §5, §11, §12 + 브랜드 컨텍스트 보호.
//
// @MX:ANCHOR: [AUTO] Validate 호출 체인의 핵심 보호 경계 - fan_in 증가 예상
// @MX:REASON: FROZEN 영역 우회 불가 요건 (REQ-DPL-011, 설계 헌법 §2)
var frozenPrefixes = []string{
	// 설계 헌법 파일 (§2 Frozen vs Evolvable Zones 전체 포함)
	".claude/rules/moai/design/constitution.md",
	// 브랜드 컨텍스트 디렉토리 (§3.1 Brand Context)
	".moai/project/brand/",
}

// frozenSectionKeywords: 헌법 파일 내 FROZEN 섹션 식별자.
// 파일 경로 + 섹션 앵커 형식의 참조를 차단하기 위해 사용.
var frozenSectionKeywords = []string{
	"#2-frozen-vs-evolvable-zones",
	"#3-brand-context-and-design-brief",
	"#31-brand-context",
	"#32-design-brief",
	"#33-relationship",
	"#5-safety-architecture",
	"#safety-architecture",
	"#11-gan-loop-contract",
	"#gan-loop-contract",
	"#12-evaluator-leniency-prevention",
	"#evaluator-leniency-prevention",
}

// IsFrozen: 지정된 파일 경로가 FROZEN 영역에 속하는지 확인한다.
// [HARD] FROZEN 판정은 소스 코드에 하드코딩된 목록에만 의존 - config로 변경 불가.
func IsFrozen(path string) bool {
	// frozenPrefixes로 시작하는 경로 차단
	for _, prefix := range frozenPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	// constitution.md 파일에 섹션 앵커가 포함된 경우 (예: constitution.md#5-...)
	if strings.Contains(path, "constitution.md") {
		for _, keyword := range frozenSectionKeywords {
			if strings.Contains(path, keyword) {
				return true
			}
		}
	}

	return false
}

// BlockWrite: 지정된 경로가 FROZEN 영역이면 오류를 반환한다.
// 허용된 경로에 대해서는 nil을 반환한다.
//
// path: 쓰기 시도 대상 파일 경로
// reason: 쓰기 시도 이유 (감사 로그용)
func BlockWrite(path, reason string) error {
	if !IsFrozen(path) {
		return nil
	}

	return fmt.Errorf("frozen 영역 쓰기 차단: '%s' 는 FROZEN 영역에 속함 (이유: %s). "+
		"설계 헌법 §2 참조: 헌법 파일 및 브랜드 컨텍스트는 직접 수정 불가", path, reason)
}
