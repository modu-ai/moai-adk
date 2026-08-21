package wizard

import "testing"

// TestTodoEnabledQuestion is AC-T-006: the confirm lives in the Page-3
// "Quality & Workflow" group, is unconditional, and defaults to true.
//
// The default is the load-bearing part. Every other Page-3 confirm mirrors a
// default-OFF config gate; this one mirrors a default-ON one, so a copy-pasted
// "false" would silently invert the operator's answer for anyone who accepts
// the default.
func TestTodoEnabledQuestion(t *testing.T) {
	q := QuestionByID(InitQuestions("/tmp/test-project"), "todo_enabled")
	if q == nil {
		t.Fatal("todo_enabled question not found in InitQuestions")
	}
	if q.Type != QuestionTypeConfirm {
		t.Errorf("todo_enabled type = %v, want QuestionTypeConfirm", q.Type)
	}
	if q.Default != "true" {
		t.Errorf("todo_enabled default = %q, want %q (workflow.todo.enabled is default-ON)", q.Default, "true")
	}
	if q.Group != "Quality & Workflow" {
		t.Errorf("todo_enabled Group = %q, want %q", q.Group, "Quality & Workflow")
	}
	if q.Condition != nil {
		t.Error("todo_enabled must be unconditional (always visible on page 3)")
	}
}

// TestSaveBoolAnswerTodoEnabled pins the capture branch. WizardResult carries a
// *bool rather than a bool so "the wizard never ran" (nil) stays distinct from
// "the operator said no" — the same absent-is-enabled distinction the config
// field makes, carried through the answer path so a non-interactive init
// cannot write enabled: false by falling through a zero value.
func TestSaveBoolAnswerTodoEnabled(t *testing.T) {
	r := &WizardResult{}
	if r.TodoEnabled != nil {
		t.Fatal("TodoEnabled must start nil (question not asked)")
	}

	saveBoolAnswer("todo_enabled", false, r)
	if r.TodoEnabled == nil || *r.TodoEnabled {
		t.Error("saveBoolAnswer(todo_enabled=false) did not record an explicit false")
	}

	r2 := &WizardResult{}
	saveBoolAnswer("todo_enabled", true, r2)
	if r2.TodoEnabled == nil || !*r2.TodoEnabled {
		t.Error("saveBoolAnswer(todo_enabled=true) did not record an explicit true")
	}
}

// TestTodoEnabledTranslationsExist is the per-question half of AC-T-007; the
// suite-wide half is TestWizardQuestionTranslationCompleteness.
func TestTodoEnabledTranslationsExist(t *testing.T) {
	for _, locale := range localizableLocales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}
		trans, ok := langTrans["todo_enabled"]
		if !ok {
			t.Errorf("locale %q: todo_enabled translation missing", locale)
			continue
		}
		if trans.Title == "" {
			t.Errorf("locale %q: todo_enabled has empty title", locale)
		}
		if trans.Description == "" {
			t.Errorf("locale %q: todo_enabled has empty description", locale)
		}
	}
}
