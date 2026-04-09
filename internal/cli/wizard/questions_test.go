package wizard

import "testing"

// TestDevelopmentModeQuestion verifies the development_mode question is present and valid.
func TestDevelopmentModeQuestion(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	q := QuestionByID(questions, "development_mode")
	if q == nil {
		t.Fatal("development_mode question not found")
		return // staticcheck SA5011 guard
	}

	if q.Type != QuestionTypeSelect {
		t.Errorf("development_mode should be QuestionTypeSelect, got %v", q.Type)
	}

	if len(q.Options) != 2 {
		t.Fatalf("development_mode should have 2 options, got %d", len(q.Options))
	}

	expectedValues := []string{"tdd", "ddd"}
	for i, expected := range expectedValues {
		if q.Options[i].Value != expected {
			t.Errorf("option %d value = %q, want %q", i, q.Options[i].Value, expected)
		}
	}

	if q.Default != "tdd" {
		t.Errorf("development_mode default = %q, want %q", q.Default, "tdd")
	}

	if q.Condition != nil {
		t.Error("development_mode should have no condition (always visible)")
	}

	if !q.Required {
		t.Error("development_mode should be required")
	}
}

// TestQuestionOrder verifies the new streamlined question order.
func TestQuestionOrder(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	// The first question should now be "project_name"
	if questions[0].ID != "project_name" {
		t.Errorf("first question should be 'project_name', got %q", questions[0].ID)
	}

	expectedIDs := []string{
		"project_name",
		"model_policy",
		"development_mode",
		"git_mode",
		"git_provider",
		"gitlab_instance_url",
		"github_username",
		"github_token",
		"gitlab_username",
		"gitlab_token",
	}

	for i, expectedID := range expectedIDs {
		if i >= len(questions) {
			t.Fatalf("expected question at index %d (%s), but only %d questions", i, expectedID, len(questions))
		}
		if questions[i].ID != expectedID {
			t.Errorf("question[%d].ID = %q, want %q", i, questions[i].ID, expectedID)
		}
	}
}

// TestSaveAnswerDevelopmentMode verifies saveAnswer stores the development_mode correctly.
func TestSaveAnswerDevelopmentMode(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	saveAnswer("development_mode", "tdd", result, &locale)
	if result.DevelopmentMode != "tdd" {
		t.Errorf("expected DevelopmentMode 'tdd', got %q", result.DevelopmentMode)
	}

	saveAnswer("development_mode", "ddd", result, &locale)
	if result.DevelopmentMode != "ddd" {
		t.Errorf("expected DevelopmentMode 'ddd', got %q", result.DevelopmentMode)
	}
}

// TestDevelopmentModeTranslationsExist verifies translations exist for the new question.
func TestDevelopmentModeTranslationsExist(t *testing.T) {
	locales := []string{"ko", "ja", "zh"}

	for _, locale := range locales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}

		trans, ok := langTrans["development_mode"]
		if !ok {
			t.Errorf("translation for 'development_mode' in locale %q not found", locale)
			continue
		}

		if trans.Title == "" {
			t.Errorf("translation for 'development_mode' in locale %q has empty title", locale)
		}
		if trans.Description == "" {
			t.Errorf("translation for 'development_mode' in locale %q has empty description", locale)
		}
		if len(trans.Options) != 2 {
			t.Errorf("locale %q: development_mode should have 2 option translations, got %d", locale, len(trans.Options))
		}
	}
}

// TestRemovedQuestionsAbsent verifies that removed user-level questions are no longer present.
func TestRemovedQuestionsAbsent(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	removedIDs := []string{
		"locale",
		"user_name",
		"git_commit_lang",
		"code_comment_lang",
		"doc_lang",
		// model_policy intentionally NOT listed here: it was re-added as a project-level question
		"agent_teams_mode",
		"max_teammates",
		"default_model",
		"teammate_display",
		"statusline_preset",
		"statusline_seg_model",
		"statusline_seg_context",
		"statusline_seg_output_style",
	}

	for _, id := range removedIDs {
		if q := QuestionByID(questions, id); q != nil {
			t.Errorf("question %q should have been removed from DefaultQuestions", id)
		}
	}
}

// TestQuestionsAllPresent verifies all expected questions are present.
func TestQuestionsAllPresent(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	expectedIDs := []string{
		"project_name",
		"model_policy",
		"development_mode",
		"git_mode",
		"git_provider",
		"gitlab_instance_url",
		"github_username",
		"github_token",
		"gitlab_username",
		"gitlab_token",
	}

	for _, id := range expectedIDs {
		if q := QuestionByID(questions, id); q == nil {
			t.Errorf("question %q not found in DefaultQuestions", id)
		}
	}
}

// TestGitConditionalFilteredByMode verifies conditional git questions hide when manual.
func TestGitConditionalFilteredByMode(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	// When git_mode is "manual", git provider questions should be hidden
	result := &WizardResult{GitMode: "manual"}
	filtered := FilteredQuestions(questions, result)

	for _, q := range filtered {
		if q.ID == "git_provider" || q.ID == "github_username" || q.ID == "github_token" {
			t.Errorf("question %q should be hidden when git_mode is 'manual'", q.ID)
		}
	}

	// When git_mode is "team", git provider question should be visible
	result = &WizardResult{GitMode: "team"}
	filtered = FilteredQuestions(questions, result)

	providerFound := false
	for _, q := range filtered {
		if q.ID == "git_provider" {
			providerFound = true
		}
	}
	if !providerFound {
		t.Error("git_provider should be visible when git_mode is 'team'")
	}
}
