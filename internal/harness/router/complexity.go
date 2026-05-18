package router

import (
	"regexp"
	"strings"
)

// ComplexitySignals는 SPEC 복잡도 추정에 사용되는 신호 모음입니다.
// REQ-HRN-001-007: 자동 감지 규칙에 입력으로 사용됩니다.
type ComplexitySignals struct {
	// FileCount는 SPEC에서 추정된 관련 파일 수입니다.
	// acceptance.md + plan.md의 internal/... 경로 언급 수로 추정합니다.
	FileCount int
	// DomainCount는 SPEC의 관련 도메인 수입니다.
	// tags 필드의 쉼표 구분 토큰 수로 추정합니다.
	DomainCount int
	// SpecType은 추정된 SPEC 유형입니다.
	// tags 또는 title에서 {bugfix, docs, config, feature, refactor} 중 하나를 추정합니다.
	SpecType string
	// HasSecurityKeyword는 보안 키워드 존재 여부입니다.
	// REQ-HRN-001-008: security_keywords 목록과 대소문자 무관 매칭.
	HasSecurityKeyword bool
	// HasPaymentKeyword는 결제 키워드 존재 여부입니다.
	// REQ-HRN-001-008: payment_keywords 목록과 대소문자 무관 매칭.
	HasPaymentKeyword bool
}

// internalPathPattern은 internal/ 또는 .moai/ 경로 패턴입니다.
// file_count 추정에 사용됩니다.
var internalPathPattern = regexp.MustCompile(`(?i)\b(internal/|\.moai/)\S+\.(go|yaml|yml|md)`)

// ExtractSignals는 SPECInput에서 복잡도 신호를 추출합니다.
// REQ-HRN-001-007/008/012.
func ExtractSignals(doc *SPECInput) ComplexitySignals {
	signals := ComplexitySignals{}

	// file_count: 본문에서 internal/ 또는 .moai/ 경로 언급 수
	if doc.Body != "" {
		matches := internalPathPattern.FindAllString(doc.Body, -1)
		// 유니크 파일 경로 카운트
		uniquePaths := make(map[string]bool)
		for _, m := range matches {
			uniquePaths[m] = true
		}
		signals.FileCount = len(uniquePaths)
	}

	// domain_count: tags 필드의 쉼표 구분 토큰 수
	if doc.Tags != "" {
		parts := strings.Split(doc.Tags, ",")
		count := 0
		for _, p := range parts {
			if strings.TrimSpace(p) != "" {
				count++
			}
		}
		signals.DomainCount = count
	}
	if signals.DomainCount == 0 {
		signals.DomainCount = 1 // 기본값: 단일 도메인
	}

	// spec_type: tags 또는 title 기반 추정
	signals.SpecType = inferSpecType(doc.Tags, doc.Title)

	// 키워드 매칭 (title + Requirements 섹션 본문, tags 제외)
	searchText := strings.ToLower(doc.Title + " " + extractRequirementsSection(doc.Body))
	signals.HasSecurityKeyword = hasAnyKeyword(searchText, securityKeywords)
	signals.HasPaymentKeyword = hasAnyKeyword(searchText, paymentKeywords)

	return signals
}

// inferSpecType은 tags와 title에서 SPEC 유형을 추정합니다.
func inferSpecType(tags, title string) string {
	combined := strings.ToLower(tags + " " + title)

	// bugfix 유형
	if containsAny(combined, []string{"bugfix", "fix", "bug", "hotfix"}) {
		return "bugfix"
	}
	// docs 유형
	if containsAny(combined, []string{"docs", "documentation", "readme"}) {
		return "docs"
	}
	// config 유형
	if containsAny(combined, []string{"config", "yaml", "configuration", "settings"}) {
		return "config"
	}
	// refactor 유형
	if containsAny(combined, []string{"refactor", "refactoring", "cleanup", "clean"}) {
		return "refactor"
	}
	// feature 유형
	if containsAny(combined, []string{"feature", "feat", "impl", "implement", "add"}) {
		return "feature"
	}

	return "other"
}

// hasAnyKeyword는 text에 keywords 중 하나라도 포함되어 있는지 확인합니다.
// 대소문자 무관 (text는 이미 소문자 변환되어야 함).
// 단어 경계 매칭으로 false-positive 방지.
func hasAnyKeyword(lowerText string, keywords []string) bool {
	for _, kw := range keywords {
		if matchKeywordBoundary(lowerText, kw) {
			return true
		}
	}
	return false
}

// containsAny는 s에 items 중 하나라도 포함되어 있는지 확인합니다.
func containsAny(s string, items []string) bool {
	for _, item := range items {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}
