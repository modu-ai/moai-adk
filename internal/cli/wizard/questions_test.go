package wizard

import "testing"

// TestReportFormatQuestion verifies the report_format question is present and valid.
func TestReportFormatQuestion(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	q := QuestionByID(questions, "report_format")
	if q == nil {
		t.Fatal("report_format question not found")
		return // staticcheck SA5011 guard
	}

	if q.Type != QuestionTypeSelect {
		t.Errorf("report_format should be QuestionTypeSelect, got %v", q.Type)
	}

	if len(q.Options) != 2 {
		t.Fatalf("report_format should have 2 options, got %d", len(q.Options))
	}

	// The closed value set mirrors internal/settings reportFormatValues.
	expectedValues := []string{"html+md", "md"}
	for i, expected := range expectedValues {
		if q.Options[i].Value != expected {
			t.Errorf("option %d value = %q, want %q", i, q.Options[i].Value, expected)
		}
	}

	if q.Default != "html+md" {
		t.Errorf("report_format default = %q, want %q", q.Default, "html+md")
	}

	if q.Condition != nil {
		t.Error("report_format should have no condition (always visible)")
	}

	if !q.Required {
		t.Error("report_format should be required")
	}
}

// TestSaveAnswerReportFormat verifies saveAnswer stores the report_format correctly.
func TestSaveAnswerReportFormat(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	saveAnswer("report_format", "html+md", result, &locale)
	if result.ReportFormat != "html+md" {
		t.Errorf("expected ReportFormat 'html+md', got %q", result.ReportFormat)
	}

	saveAnswer("report_format", "md", result, &locale)
	if result.ReportFormat != "md" {
		t.Errorf("expected ReportFormat 'md', got %q", result.ReportFormat)
	}
}

// TestReportFormatTranslationsExist verifies translations exist for report_format.
func TestReportFormatTranslationsExist(t *testing.T) {
	locales := []string{"ko", "ja", "zh"}

	for _, locale := range locales {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}
		trans, ok := langTrans["report_format"]
		if !ok {
			t.Errorf("translation for 'report_format' in locale %q not found", locale)
			continue
		}
		if trans.Title == "" {
			t.Errorf("translation for 'report_format' in locale %q has empty title", locale)
		}
		if trans.Description == "" {
			t.Errorf("translation for 'report_format' in locale %q has empty description", locale)
		}
		if len(trans.Options) != 2 {
			t.Errorf("locale %q: report_format should have 2 option translations, got %d", locale, len(trans.Options))
		}
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
		"report_format",
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
		// plan_type and development_mode were removed as interactive questions; they
		// now default silently (subscription / tdd) and are flag-only overrides.
		"plan_type",
		"development_mode",
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
		"report_format",
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

// TestSaveAnswerModelPolicy verifies saveAnswer routes the model_policy answer
// (the profile selection) into WizardResult.ModelPolicy.
func TestSaveAnswerModelPolicy(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	saveAnswer("model_policy", "high", result, &locale)
	if result.ModelPolicy != "high" {
		t.Errorf("expected ModelPolicy 'high', got %q", result.ModelPolicy)
	}
	saveAnswer("model_policy", "low", result, &locale)
	if result.ModelPolicy != "low" {
		t.Errorf("expected ModelPolicy 'low', got %q", result.ModelPolicy)
	}
}
