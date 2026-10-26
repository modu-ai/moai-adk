package wizard

// Card t1966 — wizard regression tests (items 3, 10).

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// Item 3: a pre-filled non-required input (user_name) must be clearable by
// empty input — the default fallback applies to required inputs only.
func TestInputValidator_NonRequiredEmptyClearsPrefill(t *testing.T) {
	result := &WizardResult{UserName: "Alice"}
	locale := "en"
	q := &Question{ID: "user_name", Type: QuestionTypeInput, Default: "Alice", Required: false}

	validate := inputValidator(q, result, &locale)
	if err := validate("   "); err != nil {
		t.Fatalf("validator returned %v for empty non-required input", err)
	}
	if result.UserName != "" {
		t.Errorf("UserName = %q after clearing a pre-filled non-required input, want empty", result.UserName)
	}
	// F3: the explicit-clear flag is recorded by the answer path itself.
	if !result.UserNameCleared {
		t.Error("UserNameCleared = false after an empty answer, want true")
	}
	if err := validate("Bob"); err != nil {
		t.Fatal(err)
	}
	if result.UserNameCleared || result.UserName != "Bob" {
		t.Errorf("after re-entering a value: UserNameCleared=%v UserName=%q, want false/Bob", result.UserNameCleared, result.UserName)
	}
}

// F2: the reconfigure screen does not pre-fill user_name and keeps the
// existing name on an empty answer, so its description keeps the "leave empty
// to skip" meaning in every locale.
func TestReconfigureUserNameDescription_KeepsSkipMeaning(t *testing.T) {
	q := QuestionByID(ReconfigureQuestions(t.TempDir()), "user_name")
	if q == nil {
		t.Fatal("user_name missing from the reconfigure set")
	}
	want := map[string]string{
		"en": "Leave empty to skip.",
		"ko": "비워두면 건너뜁니다.",
		"ja": "空欄でスキップできます。",
		"zh": "留空则跳过。",
	}
	for loc, frag := range want {
		if got := GetLocalizedQuestion(q, loc).Description; !strings.Contains(got, frag) {
			t.Errorf("reconfigure %s user_name description %q lacks %q", loc, got, frag)
		}
	}
}

// Control: required inputs keep the default fallback on empty input.
func TestInputValidator_RequiredEmptyFallsBackToDefault(t *testing.T) {
	result := &WizardResult{}
	locale := "en"
	q := &Question{ID: "project_name", Type: QuestionTypeInput, Default: "my-proj", Required: true}

	if err := inputValidator(q, result, &locale)(""); err != nil {
		t.Fatalf("validator returned %v", err)
	}
	if result.ProjectName != "my-proj" {
		t.Errorf("ProjectName = %q, want default fallback %q", result.ProjectName, "my-proj")
	}
}

// Item 3 wording: the user_name description must describe clearing the
// pre-filled value in every locale (en/ko/ja/zh stay in sync).
func TestUserNameDescription_DescribesClearing(t *testing.T) {
	q := QuestionByID(InitQuestions(t.TempDir()), "user_name")
	if q == nil {
		t.Fatal("user_name question missing")
	}
	want := map[string]string{
		"en": "Clear the field",
		"ko": "입력란을 비우면",
		"ja": "入力欄を空にすると",
		"zh": "清空输入框",
	}
	for loc, frag := range want {
		got := GetLocalizedQuestion(q, loc).Description
		if !strings.Contains(got, frag) {
			t.Errorf("locale %s user_name description %q lacks %q", loc, got, frag)
		}
	}
}

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
