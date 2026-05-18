package router

import (
	"regexp"
	"strings"
	"sync"
)

// securityKeywords는 force-thorough를 발동하는 보안 관련 키워드 목록입니다.
// REQ-HRN-001-008, spec.md §4 Assumptions.
//
// @MX:WARN: [AUTO] FROZEN 키워드 집합 — 변경 시 CON-002 amendment 필요
// @MX:REASON: design-constitution §5 FROZEN floor; 보안 키워드 집합 변경은 스키마 변경에 준함
var securityKeywords = []string{
	"auth", "crypto", "encrypt", "oauth", "jwt", "session", "password", "rbac", "acl",
}

// paymentKeywords는 force-thorough를 발동하는 결제 관련 키워드 목록입니다.
// REQ-HRN-001-008, spec.md §4 Assumptions.
var paymentKeywords = []string{
	"payment", "billing", "subscription", "invoice", "charge", "stripe", "paypal",
}

// matchForceThoroughKeywords는 SPEC의 title과 Requirements 섹션 본문에서
// 보안/결제 키워드를 매칭합니다.
// REQ-HRN-001-008: title OR any requirement body에 키워드가 있으면 force-thorough.
// 주의: tags 필드는 키워드 매칭 대상에서 제외합니다 (false-positive 방지).
// 단어 경계 매칭으로 "author" → "auth" 오탐 방지.
// 매칭된 키워드 목록을 반환합니다 (없으면 빈 슬라이스).
func matchForceThoroughKeywords(doc *SPECInput) []string {
	// 매칭 대상: title + Requirements 섹션 본문 (대소문자 무관)
	// tags는 제외 (false-positive 방지)
	searchText := strings.ToLower(doc.Title + " " + extractRequirementsSection(doc.Body))

	var matched []string

	for _, kw := range securityKeywords {
		if matchKeywordBoundary(searchText, kw) {
			matched = append(matched, kw)
		}
	}
	for _, kw := range paymentKeywords {
		if matchKeywordBoundary(searchText, kw) {
			matched = append(matched, kw)
		}
	}

	if matched == nil {
		return []string{}
	}
	return matched
}

// keywordPatternCache는 컴파일된 regex 패턴 캐시입니다.
// sync.Map으로 goroutine-safe 캐시를 구현합니다.
var keywordPatternCache sync.Map

// matchKeywordBoundary는 텍스트에 단어 경계를 고려하여 키워드가 있는지 확인합니다.
// "auth" → "authentication" 허용, "author" 오탐 방지.
func matchKeywordBoundary(lowerText, keyword string) bool {
	// 캐시된 패턴 조회
	if v, ok := keywordPatternCache.Load(keyword); ok {
		return v.(*regexp.Regexp).MatchString(lowerText)
	}

	// 패턴 컴파일
	patStr := `\b` + regexp.QuoteMeta(keyword) + `\b`
	compiled, err := regexp.Compile(patStr)
	if err != nil {
		// 컴파일 실패 시 단순 포함 검색으로 폴백
		return strings.Contains(lowerText, keyword)
	}

	// 캐시에 저장 (동시 write도 안전)
	keywordPatternCache.Store(keyword, compiled)
	return compiled.MatchString(lowerText)
}

// extractRequirementsSection은 SPEC 본문에서 Requirements 섹션 텍스트를 추출합니다.
// "## 5. Requirements" 섹션부터 다음 H2 섹션까지의 내용을 반환합니다.
// 섹션이 없으면 전체 본문을 반환합니다.
func extractRequirementsSection(body string) string {
	if body == "" {
		return ""
	}

	// Requirements 섹션 찾기
	lines := strings.Split(body, "\n")
	inReqs := false
	var reqLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") && strings.Contains(strings.ToLower(line), "requirement") {
			inReqs = true
			continue
		}
		if inReqs {
			// 다음 H2 섹션 시작 시 중단
			if strings.HasPrefix(line, "## ") {
				break
			}
			reqLines = append(reqLines, line)
		}
	}

	if len(reqLines) == 0 {
		return body
	}
	return strings.Join(reqLines, "\n")
}
