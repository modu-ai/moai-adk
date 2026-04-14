package cli

// F1 재현 테스트: filterByLang 미구현 pass-through
// Finding.Language 필드와 filterByLang 함수를 직접 테스트한다.
// 패키지 내부 테스트 (package cli)로 작성하여 비공개 함수에 접근한다.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// TestFilterByLang_KeepsMatchingLanguage: 언어 필터가 일치하는 항목과 언어 정보 없는 항목만 반환하는지 검증
func TestFilterByLang_KeepsMatchingLanguage(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Language: "go"},
		{RuleID: "r2", Language: "python"},
		{RuleID: "r3", Language: ""},
	}
	got := filterByLang(findings, "python")
	// python + 빈 언어 = 2개 반환
	if len(got) != 2 {
		t.Errorf("filterByLang(python) len = %d, want 2 (python 항목 + 언어 정보 없는 항목)", len(got))
	}
	for _, f := range got {
		if strings.ToLower(f.Language) != "python" && f.Language != "" {
			t.Errorf("filterByLang(python)에 잘못된 언어가 포함됨: %q", f.Language)
		}
	}
}

// TestFilterByLang_EmptyLang_ReturnsAll: 언어 필터가 비어있으면 전부 반환하는지 검증
func TestFilterByLang_EmptyLang_ReturnsAll(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Language: "go"},
		{RuleID: "r2", Language: "python"},
		{RuleID: "r3", Language: ""},
	}
	got := filterByLang(findings, "")
	if len(got) != len(findings) {
		t.Errorf("filterByLang(\"\") len = %d, want %d (전체 반환)", len(got), len(findings))
	}
}

// TestFilterByLang_CaseInsensitive: 언어 필터가 대소문자를 무시하는지 검증
func TestFilterByLang_CaseInsensitive(t *testing.T) {
	findings := []astgrep.Finding{
		{RuleID: "r1", Language: "Go"},
		{RuleID: "r2", Language: "PYTHON"},
		{RuleID: "r3", Language: ""},
	}
	got := filterByLang(findings, "go")
	// Go (대문자) + 빈 언어 = 2개
	if len(got) != 2 {
		t.Errorf("filterByLang(go) 대소문자 무시 테스트: len = %d, want 2", len(got))
	}
}
