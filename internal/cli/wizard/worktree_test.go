package wizard

import "testing"

// TestWorktreeAutoCreateQuestion (Issue 3) verifies the worktree_auto_create
// confirm question is registered in the init set (Page 3 "Quality & Workflow"),
// defaults to false (matching the config default), and is always visible.
func TestWorktreeAutoCreateQuestion(t *testing.T) {
	q := QuestionByID(InitQuestions("/tmp/test-project"), "worktree_auto_create")
	if q == nil {
		t.Fatal("worktree_auto_create question not found in InitQuestions (Issue 3)")
	}
	if q.Type != QuestionTypeConfirm {
		t.Errorf("worktree_auto_create type = %v, want QuestionTypeConfirm", q.Type)
	}
	if q.Default != "false" {
		t.Errorf("worktree_auto_create default = %q, want %q (config default — internal/config/defaults.go AutoCreate: false)", q.Default, "false")
	}
	if q.Group != "Quality & Workflow" {
		t.Errorf("worktree_auto_create Group = %q, want %q", q.Group, "Quality & Workflow")
	}
	if q.Condition != nil {
		t.Error("worktree_auto_create must be unconditional (always visible on page 3)")
	}
}

// TestSaveBoolAnswerWorktreeAutoCreate (Issue 3) verifies the boolean answer is
// captured into WizardResult.WorktreeAutoCreate.
func TestSaveBoolAnswerWorktreeAutoCreate(t *testing.T) {
	// Default / false path.
	r := &WizardResult{}
	saveBoolAnswer("worktree_auto_create", false, r)
	if r.WorktreeAutoCreate {
		t.Error("saveBoolAnswer(worktree_auto_create=false) set WorktreeAutoCreate true")
	}

	// True path.
	r2 := &WizardResult{}
	saveBoolAnswer("worktree_auto_create", true, r2)
	if !r2.WorktreeAutoCreate {
		t.Error("saveBoolAnswer(worktree_auto_create=true) did not set WorktreeAutoCreate")
	}
}

// TestWorktreeAutoCreateTranslationsExist (Issue 3) verifies ko/ja/zh
// translations exist for the worktree_auto_create confirm question.
func TestWorktreeAutoCreateTranslationsExist(t *testing.T) {
	for _, locale := range localizableLocales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}
		trans, ok := langTrans["worktree_auto_create"]
		if !ok {
			t.Errorf("locale %q: worktree_auto_create translation missing", locale)
			continue
		}
		if trans.Title == "" {
			t.Errorf("locale %q: worktree_auto_create has empty title", locale)
		}
		if trans.Description == "" {
			t.Errorf("locale %q: worktree_auto_create has empty description", locale)
		}
	}
}
