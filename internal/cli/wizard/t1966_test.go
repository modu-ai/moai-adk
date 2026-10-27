package wizard

// Card t1966 — wizard regression tests (item 4).

import "testing"

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
