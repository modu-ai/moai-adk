package mx

import (
	"regexp"
	"strings"
)

// specIDRegex는 태그 본문에서 SPEC ID를 추출하는 정규식입니다 (REQ-SPC-004-006).
// SPEC-[A-Z0-9-]+ 형식을 매칭합니다.
var specIDRegex = regexp.MustCompile(`SPEC-[A-Z0-9][A-Z0-9-]*`)

// SpecAssociator는 @MX TAG와 SPEC ID를 연결하는 역할을 합니다.
// 두 가지 방식으로 연결합니다 (REQ-SPC-004-006):
//  1. 태그의 파일 경로가 SPEC의 module: 프론트매터에 나열된 경로 하위에 있는 경우
//  2. 태그 본문에 명시적으로 SPEC-[A-Z0-9-]+ 패턴이 있는 경우
type SpecAssociator struct {
	// specModules는 specID → []modulePath 매핑입니다.
	// SPEC 문서의 module: 프론트매터에서 읽어옵니다.
	specModules map[string][]string
}

// NewSpecAssociator는 SPEC ID → 모듈 경로 매핑으로 SpecAssociator를 생성합니다.
func NewSpecAssociator(specModules map[string][]string) *SpecAssociator {
	return &SpecAssociator{
		specModules: specModules,
	}
}

// Associate는 태그와 연결된 SPEC ID 목록을 반환합니다 (REQ-SPC-004-006).
// 두 가지 방식(경로 기반, 본문 기반)을 OR 결합하고 중복을 제거합니다.
func (a *SpecAssociator) Associate(tag Tag) []string {
	seen := make(map[string]bool)
	var result []string

	// (a) 경로 기반 연결: 태그 파일 경로가 SPEC의 모듈 경로 하위인 경우
	for specID, modules := range a.specModules {
		if isFileUnderModules(tag.File, modules) && !seen[specID] {
			seen[specID] = true
			result = append(result, specID)
		}
	}

	// (b) 본문 기반 연결: 태그 본문에 명시적 SPEC ID 참조가 있는 경우
	for _, specID := range ExtractSpecIDs(tag.Body) {
		if !seen[specID] {
			seen[specID] = true
			result = append(result, specID)
		}
	}

	return result
}

// ExtractSpecIDs는 태그 본문에서 SPEC ID를 정규식으로 추출합니다.
// "ANCHOR for SPEC-AUTH-001 handler" → ["SPEC-AUTH-001"] (REQ-SPC-004-006 (b))
func ExtractSpecIDs(body string) []string {
	matches := specIDRegex.FindAllString(body, -1)
	if len(matches) == 0 {
		return []string{}
	}

	// 중복 제거
	seen := make(map[string]bool)
	var result []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			result = append(result, m)
		}
	}
	return result
}

// isFileUnderModules는 파일 경로가 모듈 경로 중 하나의 하위 경로인지 확인합니다.
// 경로 접두사 매칭을 사용합니다 (REQ-SPC-004-006 (a)).
func isFileUnderModules(filePath string, modulePaths []string) bool {
	for _, modulePath := range modulePaths {
		if strings.HasPrefix(filePath, modulePath) {
			return true
		}
	}
	return false
}
