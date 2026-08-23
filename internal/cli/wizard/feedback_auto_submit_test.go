package wizard

import "testing"

// feedback_auto_submit_test.go replicates the worktree_auto_create 3-test set
// for the feedback.auto_submit confirm question: registration, answer capture,
// and translation completeness.

// TestFeedbackAutoSubmitQuestion verifies the feedback_auto_submit confirm
// question is registered in the init set (Page 3 "Quality & Workflow"),
// defaults to false (matching the config default), and is always visible.
func TestFeedbackAutoSubmitQuestion(t *testing.T) {
	q := QuestionByID(InitQuestions("/tmp/test-project"), "feedback_auto_submit")
	if q == nil {
		t.Fatal("feedback_auto_submit question not found in InitQuestions")
	}
	if q.Type != QuestionTypeConfirm {
		t.Errorf("feedback_auto_submit type = %v, want QuestionTypeConfirm", q.Type)
	}
	if q.Default != "false" {
		t.Errorf("feedback_auto_submit default = %q, want %q (config default — internal/config/defaults.go DefaultFeedbackAutoSubmit)", q.Default, "false")
	}
	if q.Group != "Quality & Workflow" {
		t.Errorf("feedback_auto_submit Group = %q, want %q", q.Group, "Quality & Workflow")
	}
	if q.Condition != nil {
		t.Error("feedback_auto_submit must be unconditional (always visible on page 3)")
	}
}

// TestSaveBoolAnswerFeedbackAutoSubmit verifies the boolean answer is captured
// into WizardResult.FeedbackAutoSubmit. The field is a pointer because absence
// (the question was never asked) must stay distinguishable from an explicit
// "no": an unasked question writes nothing.
func TestSaveBoolAnswerFeedbackAutoSubmit(t *testing.T) {
	r := &WizardResult{}
	saveBoolAnswer("feedback_auto_submit", false, r)
	if r.FeedbackAutoSubmit == nil {
		t.Fatal("saveBoolAnswer(feedback_auto_submit=false) left FeedbackAutoSubmit nil")
	}
	if *r.FeedbackAutoSubmit {
		t.Error("saveBoolAnswer(feedback_auto_submit=false) recorded true")
	}

	r2 := &WizardResult{}
	saveBoolAnswer("feedback_auto_submit", true, r2)
	if r2.FeedbackAutoSubmit == nil {
		t.Fatal("saveBoolAnswer(feedback_auto_submit=true) left FeedbackAutoSubmit nil")
	}
	if !*r2.FeedbackAutoSubmit {
		t.Error("saveBoolAnswer(feedback_auto_submit=true) recorded false")
	}

	// An untouched result keeps nil — the non-interactive path.
	r3 := &WizardResult{}
	if r3.FeedbackAutoSubmit != nil {
		t.Error("zero WizardResult must leave FeedbackAutoSubmit nil")
	}
}

// TestFeedbackAutoSubmitTranslationsExist verifies ko/ja/zh translations exist
// for the feedback_auto_submit confirm question. English-only would fail
// TestWizardQuestionTranslationCompleteness by design; this test names the
// question so the failure points at the right entry.
func TestFeedbackAutoSubmitTranslationsExist(t *testing.T) {
	for _, locale := range localizableLocales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}
		trans, ok := langTrans["feedback_auto_submit"]
		if !ok {
			t.Errorf("locale %q: feedback_auto_submit translation missing", locale)
			continue
		}
		if trans.Title == "" {
			t.Errorf("locale %q: feedback_auto_submit has empty title", locale)
		}
		if trans.Description == "" {
			t.Errorf("locale %q: feedback_auto_submit has empty description", locale)
		}
	}
}
