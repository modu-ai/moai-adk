package wizard

// Card t1966 — wizard regression tests (items 4, 10).

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// Item 4: pins the counts the questions.go / wizard.go doc comments state
// (four shared questions, five init questions on four pages).
func TestQuestionSetCounts_MatchDocComments(t *testing.T) {
	root := t.TempDir()
	if n := len(DefaultQuestions(root)); n != 4 {
		t.Errorf("DefaultQuestions = %d questions, doc comment says four", n)
	}
	init := InitQuestions(root)
	if len(init) != 5 {
		t.Errorf("InitQuestions = %d questions, doc comment says five", len(init))
	}
	pages := map[string]bool{}
	for i := range init {
		pages[init[i].Group] = true
	}
	if len(pages) != 4 {
		t.Errorf("InitQuestions spans %d pages, doc comment says four: %v", len(pages), pages)
	}
}

// Item 10: the select height is static in huh, so it must cover the LONGEST
// localized description. A height computed from the English text alone hides
// options when another locale's description wraps to more rows.
func TestSelectHeight_CoversLongestLocaleDescription(t *testing.T) {
	const id = "t1966_height_probe"
	long := strings.Repeat("긴 설명 문장이 여러 줄로 감싸집니다 ", 10)
	translations["ko"][id] = QuestionTranslation{Title: "t", Description: long}
	t.Cleanup(func() { delete(translations["ko"], id) })

	q := &Question{
		ID:          id,
		Type:        QuestionTypeSelect,
		Title:       "t",
		Description: "short",
		Options:     []Option{{Label: "a", Value: "a"}, {Label: "b", Value: "b"}},
	}

	rows := func(desc string) int {
		return lipgloss.Height(lipgloss.NewStyle().Width(selectFieldUsableWidth).Render(desc))
	}
	need := 1 + rows(long) + len(q.Options)
	if rows(long) <= rows("short") {
		t.Fatalf("precondition: localized description must wrap to more rows than English (%d vs %d)", rows(long), rows("short"))
	}
	if got := selectHeight(q); got < need {
		t.Errorf("selectHeight = %d, want >= %d (longest locale description)", got, need)
	}
}
