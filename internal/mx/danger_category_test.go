package mx

import (
	"testing"
)

// TestDangerCategoryMatcher_Match는 REASON 텍스트와 카테고리 매칭을 테스트합니다.
func TestDangerCategoryMatcher_Match(t *testing.T) {
	matcher := NewDangerCategoryMatcher(DangerCategoryConfig{})

	tests := []struct {
		name     string
		reason   string
		category string
		expected bool
	}{
		{
			name:     "goroutine leak → concurrency",
			reason:   "goroutine leak on panic",
			category: "concurrency",
			expected: true,
		},
		{
			name:     "unbounded channel → concurrency",
			reason:   "unbounded channel usage detected",
			category: "concurrency",
			expected: true,
		},
		{
			name:     "race condition → concurrency",
			reason:   "race condition between goroutines",
			category: "concurrency",
			expected: true,
		},
		{
			name:     "missing Close → resource-leak",
			reason:   "missing Close on io.Reader",
			category: "resource-leak",
			expected: true,
		},
		{
			name:     "fd leak → resource-leak",
			reason:   "fd leak in error path",
			category: "resource-leak",
			expected: true,
		},
		{
			name:     "defer missing → cleanup",
			reason:   "defer missing for cleanup",
			category: "cleanup",
			expected: true,
		},
		{
			name:     "hardcoded credential → security",
			reason:   "hardcoded credential in source",
			category: "security",
			expected: true,
		},
		{
			name:     "sql injection → security",
			reason:   "sql injection vulnerability",
			category: "security",
			expected: true,
		},
		{
			name:     "goroutine leak ≠ resource-leak",
			reason:   "goroutine leak",
			category: "resource-leak",
			expected: false,
		},
		{
			name:     "관련 없는 텍스트",
			reason:   "단순 경고 메시지",
			category: "concurrency",
			expected: false,
		},
		{
			name:     "대소문자 구분 없는 매칭",
			reason:   "Goroutine Leak on panic",
			category: "concurrency",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matcher.Match(tt.reason, tt.category)
			if got != tt.expected {
				t.Errorf("Match(%q, %q): 기대 %v, 실제 %v", tt.reason, tt.category, tt.expected, got)
			}
		})
	}
}

// TestDangerCategoryMatcher_CategoryOf는 REASON 텍스트의 카테고리 추론을 테스트합니다.
func TestDangerCategoryMatcher_CategoryOf(t *testing.T) {
	matcher := NewDangerCategoryMatcher(DangerCategoryConfig{})

	tests := []struct {
		reason   string
		expected string
	}{
		{"goroutine leak detected", "concurrency"},
		{"missing Close on error", "resource-leak"},
		{"defer missing in handler", "cleanup"},
		{"hardcoded credential found", "security"},
		{"관련 없는 텍스트", ""},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			got := matcher.CategoryOf(tt.reason)
			if got != tt.expected {
				t.Errorf("CategoryOf(%q): 기대 %q, 실제 %q", tt.reason, tt.expected, got)
			}
		})
	}
}

// TestDangerCategoryMatcher_ValidateCategory는 카테고리 유효성 검증을 테스트합니다.
func TestDangerCategoryMatcher_ValidateCategory(t *testing.T) {
	matcher := NewDangerCategoryMatcher(DangerCategoryConfig{})

	validCategories := []string{"concurrency", "resource-leak", "cleanup", "security"}
	for _, cat := range validCategories {
		t.Run("valid:"+cat, func(t *testing.T) {
			if !matcher.ValidateCategory(cat) {
				t.Errorf("ValidateCategory(%q): 유효한 카테고리인데 false 반환", cat)
			}
		})
	}

	invalidCategories := []string{"nonexistent", "random", ""}
	for _, cat := range invalidCategories {
		t.Run("invalid:"+cat, func(t *testing.T) {
			if matcher.ValidateCategory(cat) {
				t.Errorf("ValidateCategory(%q): 유효하지 않은 카테고리인데 true 반환", cat)
			}
		})
	}
}

// TestDangerCategoryMatcher_CustomConfig는 커스텀 설정을 테스트합니다.
func TestDangerCategoryMatcher_CustomConfig(t *testing.T) {
	config := DangerCategoryConfig{
		Categories: map[string][]string{
			"custom": {"my-pattern", "another-pattern"},
		},
	}

	matcher := NewDangerCategoryMatcher(config)

	if !matcher.Match("my-pattern found in code", "custom") {
		t.Error("커스텀 카테고리 패턴 매칭 실패")
	}

	// 기본 카테고리는 사용되지 않아야 함
	if matcher.Match("goroutine leak", "concurrency") {
		t.Error("커스텀 설정 시 기본 카테고리 사용됨")
	}
}

// TestDangerCategoryMatcher_KnownCategories는 알려진 카테고리 목록을 테스트합니다.
func TestDangerCategoryMatcher_KnownCategories(t *testing.T) {
	matcher := NewDangerCategoryMatcher(DangerCategoryConfig{})

	categories := matcher.KnownCategories()
	if len(categories) == 0 {
		t.Error("알려진 카테고리 없음")
	}

	// 기본 4개 카테고리가 모두 있어야 함
	expected := map[string]bool{
		"concurrency":   false,
		"resource-leak": false,
		"cleanup":       false,
		"security":      false,
	}

	for _, cat := range categories {
		expected[cat] = true
	}

	for cat, found := range expected {
		if !found {
			t.Errorf("기본 카테고리 누락: %s", cat)
		}
	}
}
