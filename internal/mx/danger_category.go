package mx

import (
	"strings"
)

// DangerCategoryConfig는 mx.yaml의 danger_categories 설정을 나타냅니다.
// 카테고리 이름을 WARN REASON 텍스트 패턴 목록에 매핑합니다 (REQ-SPC-004-012).
//
// @MX:NOTE: [AUTO] DangerCategoryConfig — 기본 카테고리 패턴은 보수적으로 설계됨
// 기본값: concurrency(동시성), resource-leak(리소스 누수), cleanup(정리), security(보안)
// 사용자가 mx.yaml에서 커스터마이즈 가능하며, 언어별 확장을 지원합니다
type DangerCategoryConfig struct {
	// Categories는 카테고리 이름 → 패턴 목록 매핑입니다.
	// 패턴은 대소문자 구분 없이 부분 문자열 매칭을 사용합니다.
	Categories map[string][]string `yaml:"danger_categories"`

	// TestPaths는 fan-in 계산 시 제외할 테스트 파일 경로 패턴 목록입니다 (REQ-SPC-004-040).
	TestPaths []string `yaml:"test_paths"`
}

// DefaultDangerCategories는 mx.yaml에 설정이 없을 때 사용하는 기본 위험 카테고리 매핑입니다.
// REQ-SPC-004-012에 정의된 4가지 기본 카테고리를 포함합니다.
var DefaultDangerCategories = map[string][]string{
	"concurrency": {
		"goroutine leak",
		"unbounded channel",
		"race condition",
	},
	"resource-leak": {
		"missing Close",
		"fd leak",
	},
	"cleanup": {
		"defer missing",
		"Close not called",
	},
	"security": {
		"hardcoded credential",
		"sql injection",
		"xss",
	},
}

// DangerCategoryMatcher는 WARN REASON 텍스트와 위험 카테고리를 매칭합니다.
type DangerCategoryMatcher struct {
	config DangerCategoryConfig
}

// NewDangerCategoryMatcher는 주어진 설정으로 DangerCategoryMatcher를 생성합니다.
// 설정에 Categories가 없으면 기본 카테고리를 사용합니다.
func NewDangerCategoryMatcher(config DangerCategoryConfig) *DangerCategoryMatcher {
	if len(config.Categories) == 0 {
		config.Categories = DefaultDangerCategories
	}
	return &DangerCategoryMatcher{config: config}
}

// Match는 reason 텍스트가 category의 패턴 중 하나와 매칭되는지 확인합니다.
// 패턴 매칭은 대소문자 구분 없는 부분 문자열 검색을 사용합니다 (REQ-SPC-004-012).
func (m *DangerCategoryMatcher) Match(reason, category string) bool {
	patterns, ok := m.config.Categories[category]
	if !ok {
		return false
	}

	lowerReason := strings.ToLower(reason)
	for _, pattern := range patterns {
		if strings.Contains(lowerReason, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// CategoryOf는 reason 텍스트에 매칭되는 첫 번째 카테고리를 반환합니다.
// 매칭되는 카테고리가 없으면 빈 문자열을 반환합니다.
func (m *DangerCategoryMatcher) CategoryOf(reason string) string {
	lowerReason := strings.ToLower(reason)
	for cat, patterns := range m.config.Categories {
		for _, pattern := range patterns {
			if strings.Contains(lowerReason, strings.ToLower(pattern)) {
				return cat
			}
		}
	}
	return ""
}

// ValidateCategory는 주어진 category가 알려진 카테고리인지 확인합니다.
// 알려진 카테고리가 아니면 false를 반환합니다.
func (m *DangerCategoryMatcher) ValidateCategory(category string) bool {
	_, ok := m.config.Categories[category]
	return ok
}

// KnownCategories는 설정에 정의된 모든 카테고리 이름을 반환합니다.
func (m *DangerCategoryMatcher) KnownCategories() []string {
	result := make([]string, 0, len(m.config.Categories))
	for cat := range m.config.Categories {
		result = append(result, cat)
	}
	return result
}
