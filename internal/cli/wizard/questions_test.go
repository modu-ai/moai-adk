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

	// The first question is now "conversation_language" (drives the render
	// language of every later question); "user_name" is second.
	if questions[0].ID != "conversation_language" {
		t.Errorf("first question should be 'conversation_language', got %q", questions[0].ID)
	}

	expectedIDs := []string{
		"conversation_language",
		"user_name",
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
		"advanced_bridge",
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
		// "locale" stays removed — the language question uses the clearer ID
		// "conversation_language". "user_name" was RE-ADDED as wizard step 2, so
		// it is intentionally absent from this removed list.
		"locale",
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
		"conversation_language",
		"user_name",
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
		"advanced_bridge",
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

// TestConversationLanguageQuestion verifies the language question is first, is a
// Select with the 4 supported locales, defaults to "en", and is unconditional.
func TestConversationLanguageQuestion(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	if questions[0].ID != "conversation_language" {
		t.Fatalf("conversation_language must be the first question, got %q", questions[0].ID)
	}
	q := QuestionByID(questions, "conversation_language")
	if q == nil {
		t.Fatal("conversation_language question not found")
		return
	}
	if q.Type != QuestionTypeSelect {
		t.Errorf("conversation_language should be QuestionTypeSelect, got %v", q.Type)
	}
	wantValues := []string{"en", "ko", "ja", "zh"}
	if len(q.Options) != len(wantValues) {
		t.Fatalf("conversation_language should have %d options, got %d", len(wantValues), len(q.Options))
	}
	for i, v := range wantValues {
		if q.Options[i].Value != v {
			t.Errorf("option %d value = %q, want %q", i, q.Options[i].Value, v)
		}
	}
	if q.Default != "en" {
		t.Errorf("conversation_language default = %q, want %q", q.Default, "en")
	}
	if !q.Required {
		t.Error("conversation_language should be required")
	}
	if q.Condition != nil {
		t.Error("conversation_language should be unconditional (always visible)")
	}
}

// TestUserNameQuestion verifies the user_name question is second, is an optional
// Input, and is unconditional.
func TestUserNameQuestion(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	if questions[1].ID != "user_name" {
		t.Fatalf("user_name must be the second question, got %q", questions[1].ID)
	}
	q := QuestionByID(questions, "user_name")
	if q == nil {
		t.Fatal("user_name question not found")
		return
	}
	if q.Type != QuestionTypeInput {
		t.Errorf("user_name should be QuestionTypeInput, got %v", q.Type)
	}
	if q.Required {
		t.Error("user_name should not be required (empty allowed)")
	}
	if q.Condition != nil {
		t.Error("user_name should be unconditional (always visible)")
	}
}

// TestAdvancedBridgeQuestion verifies the bridge Confirm is last, defaults to No,
// and is hidden once StandardMode is preset (by --standard/--advanced).
func TestAdvancedBridgeQuestion(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	if last := questions[len(questions)-1]; last.ID != "advanced_bridge" {
		t.Fatalf("advanced_bridge must be the last question, got %q", last.ID)
	}
	q := QuestionByID(questions, "advanced_bridge")
	if q == nil {
		t.Fatal("advanced_bridge question not found")
		return
	}
	if q.Type != QuestionTypeConfirm {
		t.Errorf("advanced_bridge should be QuestionTypeConfirm, got %v", q.Type)
	}
	if q.Default != "false" {
		t.Errorf("advanced_bridge default = %q, want %q (No)", q.Default, "false")
	}
	if q.Condition == nil {
		t.Fatal("advanced_bridge must have a condition")
	}
	// Visible in quick mode (StandardMode false), hidden when preset by flag.
	if !q.Condition(&WizardResult{StandardMode: false}) {
		t.Error("advanced_bridge should be visible in quick mode (StandardMode=false)")
	}
	if q.Condition(&WizardResult{StandardMode: true}) {
		t.Error("advanced_bridge should be hidden when StandardMode is preset by flag")
	}
}

// TestSaveAnswerConversationLanguage verifies the answer stores the code AND
// updates the live locale pointer (drives reactive re-render of later questions).
func TestSaveAnswerConversationLanguage(t *testing.T) {
	result := &WizardResult{}
	locale := "en"

	saveAnswer("conversation_language", "ko", result, &locale)
	if result.ConversationLang != "ko" {
		t.Errorf("ConversationLang = %q, want %q", result.ConversationLang, "ko")
	}
	if locale != "ko" {
		t.Errorf("live locale = %q, want %q (must update for reactive rendering)", locale, "ko")
	}
}

// TestSaveAnswerUserName verifies the user_name answer is stored.
func TestSaveAnswerUserName(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	saveAnswer("user_name", "GOOS", result, &locale)
	if result.UserName != "GOOS" {
		t.Errorf("UserName = %q, want %q", result.UserName, "GOOS")
	}
}

// TestSaveBoolAnswerAdvancedBridge verifies the bridge flips StandardMode so the
// gated Phase 1 questions become visible in the same run.
func TestSaveBoolAnswerAdvancedBridge(t *testing.T) {
	result := &WizardResult{}
	saveBoolAnswer("advanced_bridge", true, result)
	if !result.StandardMode {
		t.Error("advanced_bridge=Yes must set StandardMode=true")
	}
	saveBoolAnswer("advanced_bridge", false, result)
	if result.StandardMode {
		t.Error("advanced_bridge=No must set StandardMode=false")
	}
}

// TestNewQuestionTranslationsExist verifies ko/ja/zh translations exist for the
// three new questions (no hardcoded UI strings that belong in the tables).
func TestNewQuestionTranslationsExist(t *testing.T) {
	ids := []string{"conversation_language", "user_name", "advanced_bridge"}
	for _, locale := range []string{"ko", "ja", "zh"} {
		langTrans, ok := translations[locale]
		if !ok {
			t.Fatalf("translations for locale %q not found", locale)
		}
		for _, id := range ids {
			trans, ok := langTrans[id]
			if !ok {
				t.Errorf("translation for %q in locale %q not found", id, locale)
				continue
			}
			if trans.Title == "" {
				t.Errorf("translation for %q in locale %q has empty title", id, locale)
			}
			if trans.Description == "" {
				t.Errorf("translation for %q in locale %q has empty description", id, locale)
			}
		}
	}
	// The conversation_language options must NOT be translated — native language
	// names (Korean (한국어), etc.) stay in the base question.
	q := QuestionByID(DefaultQuestions("/tmp/test"), "conversation_language")
	localized := GetLocalizedQuestion(q, "ko")
	if localized.Options[1].Label != "Korean (한국어)" {
		t.Errorf("language option labels must stay native, got %q", localized.Options[1].Label)
	}
}

// TestPrefillIdentityDefaults verifies the profile user name pre-fills the
// user_name question default (empty leaves it untouched).
func TestPrefillIdentityDefaults(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	prefillIdentityDefaults(questions, "GOOS")
	if q := QuestionByID(questions, "user_name"); q == nil || q.Default != "GOOS" {
		t.Errorf("user_name default should be pre-filled to 'GOOS', got %+v", q)
	}

	questions2 := DefaultQuestions("/tmp/test")
	prefillIdentityDefaults(questions2, "")
	if q := QuestionByID(questions2, "user_name"); q == nil || q.Default != "" {
		t.Errorf("empty userName should leave default empty, got %+v", q)
	}
}

// TestPrefillLocaleDefault verifies the profile/project locale pre-fills the
// conversation_language default (empty leaves the static "en" default).
func TestPrefillLocaleDefault(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	prefillLocaleDefault(questions, "ja")
	if q := QuestionByID(questions, "conversation_language"); q == nil || q.Default != "ja" {
		t.Errorf("conversation_language default should be 'ja', got %+v", q)
	}

	questions2 := DefaultQuestions("/tmp/test")
	prefillLocaleDefault(questions2, "")
	if q := QuestionByID(questions2, "conversation_language"); q == nil || q.Default != "en" {
		t.Errorf("empty locale should keep the static 'en' default, got %+v", q)
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
