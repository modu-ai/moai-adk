package wizard

import (
	"testing"
)

func TestWizardResult(t *testing.T) {
	result := &WizardResult{
		ProjectName:     "test-project",
		Locale:          "ko",
		UserName:        "TestUser",
		GitMode:         "personal",
		GitHubUsername:  "testuser",
		GitCommitLang:   "en",
		CodeCommentLang: "en",
		DocLang:         "ko",
	}

	if result.ProjectName != "test-project" {
		t.Errorf("expected ProjectName 'test-project', got %q", result.ProjectName)
	}
	if result.Locale != "ko" {
		t.Errorf("expected Locale 'ko', got %q", result.Locale)
	}
}

func TestGetLanguageName(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"en", "English"},
		{"ko", "Korean (한국어)"},
		{"ja", "Japanese (日本語)"},
		{"zh", "Chinese (中文)"},
		{"unknown", "English"}, // Default fallback
		{"", "English"},        // Empty string fallback
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := GetLanguageName(tt.code)
			if got != tt.expected {
				t.Errorf("GetLanguageName(%q) = %q, want %q", tt.code, got, tt.expected)
			}
		})
	}
}

func TestNewStyles(t *testing.T) {
	styles := NewStyles()
	if styles == nil {
		t.Fatal("NewStyles() returned nil")
	}

	// Verify styles are initialized
	if styles.Title.String() == "" {
		t.Log("Title style initialized")
	}
}

func TestNoColorStyles(t *testing.T) {
	styles := NoColorStyles()
	if styles == nil {
		t.Fatal("NoColorStyles() returned nil")
	}
}

func TestDefaultQuestions(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	if len(questions) == 0 {
		t.Fatal("DefaultQuestions() returned empty slice")
	}

	// Verify first question is locale selection
	if questions[0].ID != "locale" {
		t.Errorf("expected first question ID 'locale', got %q", questions[0].ID)
	}

	// Verify project name question
	var found bool
	for _, q := range questions {
		if q.ID == "project_name" {
			found = true
			if q.Default != "test-project" {
				t.Errorf("expected project_name default 'test-project', got %q", q.Default)
			}
			break
		}
	}
	if !found {
		t.Error("project_name question not found")
	}

	// Verify development_mode question is NOT in wizard (auto-configured by /moai project)
	for _, q := range questions {
		if q.ID == "development_mode" {
			t.Error("development_mode question should not be in wizard (auto-configured by /moai project)")
		}
	}
}

func TestFilteredQuestions(t *testing.T) {
	questions := []Question{
		{ID: "always_show", Type: QuestionTypeInput},
		{
			ID:   "conditional",
			Type: QuestionTypeInput,
			Condition: func(r *WizardResult) bool {
				return r.GitMode == "team"
			},
		},
	}

	// With GitMode = "manual", conditional question should be filtered out
	result := &WizardResult{GitMode: "manual"}
	filtered := FilteredQuestions(questions, result)
	if len(filtered) != 1 {
		t.Errorf("expected 1 filtered question, got %d", len(filtered))
	}

	// With GitMode = "team", both questions should be included
	result.GitMode = "team"
	filtered = FilteredQuestions(questions, result)
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered questions, got %d", len(filtered))
	}
}

func TestTotalVisibleQuestions(t *testing.T) {
	questions := []Question{
		{ID: "q1", Type: QuestionTypeInput},
		{ID: "q2", Type: QuestionTypeInput},
		{
			ID:   "q3",
			Type: QuestionTypeInput,
			Condition: func(r *WizardResult) bool {
				return false // Always hidden
			},
		},
	}

	result := &WizardResult{}
	total := TotalVisibleQuestions(questions, result)
	if total != 2 {
		t.Errorf("expected 2 visible questions, got %d", total)
	}
}

func TestQuestionByID(t *testing.T) {
	questions := []Question{
		{ID: "q1", Title: "Question 1"},
		{ID: "q2", Title: "Question 2"},
	}

	q := QuestionByID(questions, "q1")
	if q == nil {
		t.Fatal("QuestionByID('q1') returned nil")
	}
	if q.Title != "Question 1" {
		t.Errorf("expected Title 'Question 1', got %q", q.Title)
	}

	q = QuestionByID(questions, "nonexistent")
	if q != nil {
		t.Error("QuestionByID('nonexistent') should return nil")
	}
}

func TestRunWithEmptyQuestions(t *testing.T) {
	_, err := Run(nil, nil)
	if err != ErrNoQuestions {
		t.Errorf("expected ErrNoQuestions, got %v", err)
	}

	_, err = Run([]Question{}, nil)
	if err != ErrNoQuestions {
		t.Errorf("expected ErrNoQuestions, got %v", err)
	}
}

func TestErrors(t *testing.T) {
	if ErrCancelled.Error() != "wizard cancelled by user" {
		t.Errorf("unexpected ErrCancelled message: %q", ErrCancelled.Error())
	}
	if ErrNoQuestions.Error() != "no questions provided" {
		t.Errorf("unexpected ErrNoQuestions message: %q", ErrNoQuestions.Error())
	}
	if ErrInvalidQuestion.Error() != "invalid question index" {
		t.Errorf("unexpected ErrInvalidQuestion message: %q", ErrInvalidQuestion.Error())
	}
}

// --- saveAnswer standalone function tests ---

func TestSaveAnswer_AllFields(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	tests := []struct {
		id       string
		value    string
		checkFn  func() bool
		expected string
	}{
		{"locale", "ko", func() bool { return result.Locale == "ko" && locale == "ko" }, "Locale=ko, locale pointer=ko"},
		{"user_name", "testuser", func() bool { return result.UserName == "testuser" }, "UserName=testuser"},
		{"project_name", "myproject", func() bool { return result.ProjectName == "myproject" }, "ProjectName=myproject"},
		{"git_mode", "personal", func() bool { return result.GitMode == "personal" }, "GitMode=personal"},
		{"git_provider", "gitlab", func() bool { return result.GitProvider == "gitlab" }, "GitProvider=gitlab"},
		{"github_username", "ghuser", func() bool { return result.GitHubUsername == "ghuser" }, "GitHubUsername=ghuser"},
		{"github_token", "ghp_token", func() bool { return result.GitHubToken == "ghp_token" }, "GitHubToken=ghp_token"},
		{"gitlab_instance_url", "https://gl.co", func() bool { return result.GitLabInstanceURL == "https://gl.co" }, "GitLabInstanceURL"},
		{"gitlab_username", "gluser", func() bool { return result.GitLabUsername == "gluser" }, "GitLabUsername=gluser"},
		{"gitlab_token", "glpat-token", func() bool { return result.GitLabToken == "glpat-token" }, "GitLabToken=glpat-token"},
		{"git_commit_lang", "en", func() bool { return result.GitCommitLang == "en" }, "GitCommitLang=en"},
		{"code_comment_lang", "en", func() bool { return result.CodeCommentLang == "en" }, "CodeCommentLang=en"},
		{"doc_lang", "ko", func() bool { return result.DocLang == "ko" }, "DocLang=ko"},
		{"model_policy", "high", func() bool { return result.ModelPolicy == "high" }, "ModelPolicy=high"},
		{"agent_teams_mode", "auto", func() bool { return result.AgentTeamsMode == "auto" }, "AgentTeamsMode=auto"},
		{"max_teammates", "3", func() bool { return result.MaxTeammates == "3" }, "MaxTeammates=3"},
		{"default_model", "sonnet", func() bool { return result.DefaultModel == "sonnet" }, "DefaultModel=sonnet"},
		{"teammate_display", "tmux", func() bool { return result.TeammateDisplay == "tmux" }, "TeammateDisplay=tmux"},
		{"statusline_preset", "compact", func() bool { return result.StatuslinePreset == "compact" }, "StatuslinePreset=compact"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			saveAnswer(tt.id, tt.value, result, &locale)
			if !tt.checkFn() {
				t.Errorf("saveAnswer(%q, %q) did not set expected: %s", tt.id, tt.value, tt.expected)
			}
		})
	}
}

func TestSaveAnswer_Locale_UpdatesPointer(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	saveAnswer("locale", "ja", result, &locale)

	if result.Locale != "ja" {
		t.Errorf("expected Locale 'ja', got %q", result.Locale)
	}
	if locale != "ja" {
		t.Errorf("expected locale pointer 'ja', got %q", locale)
	}
}

func TestSaveAnswer_StatuslineSegments(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	// Initially StatuslineSegments should be nil
	if result.StatuslineSegments != nil {
		t.Error("StatuslineSegments should be nil initially")
	}

	// Save first segment - should initialize map
	saveAnswer("statusline_seg_model", "true", result, &locale)
	if result.StatuslineSegments == nil {
		t.Fatal("StatuslineSegments should be initialized after saving a segment")
	}
	if !result.StatuslineSegments["model"] {
		t.Error("StatuslineSegments['model'] should be true")
	}

	// Save second segment
	saveAnswer("statusline_seg_context", "false", result, &locale)
	if result.StatuslineSegments["context"] {
		t.Error("StatuslineSegments['context'] should be false")
	}
}

func TestSaveAnswer_UnknownID(t *testing.T) {
	result := &WizardResult{}
	locale := ""

	// Should not panic for unknown IDs
	saveAnswer("unknown_field", "value", result, &locale)
}

// --- Localization tests ---

func TestGetLocalizedQuestion(t *testing.T) {
	q := &Question{
		ID:          "locale",
		Type:        QuestionTypeSelect,
		Title:       "Select conversation language",
		Description: "This determines the language Claude will use.",
		Options: []Option{
			{Label: "Korean (한국어)", Value: "ko", Desc: "Korean"},
			{Label: "English", Value: "en", Desc: "English"},
		},
	}

	// Test English (default, no translation)
	localizedEn := GetLocalizedQuestion(q, "en")
	if localizedEn.Title != q.Title {
		t.Errorf("expected English title %q, got %q", q.Title, localizedEn.Title)
	}

	// Test Korean translation
	localizedKo := GetLocalizedQuestion(q, "ko")
	if localizedKo.Title == q.Title {
		t.Error("Korean title should be different from English")
	}
	if localizedKo.Title != "대화 언어 선택" {
		t.Errorf("expected Korean title '대화 언어 선택', got %q", localizedKo.Title)
	}

	// Test Japanese translation
	localizedJa := GetLocalizedQuestion(q, "ja")
	if localizedJa.Title != "会話言語を選択" {
		t.Errorf("expected Japanese title '会話言語を選択', got %q", localizedJa.Title)
	}

	// Test unknown locale (should return original)
	localizedUnknown := GetLocalizedQuestion(q, "xx")
	if localizedUnknown.Title != q.Title {
		t.Errorf("unknown locale should return original title")
	}
}

func TestGetUIStrings(t *testing.T) {
	// Test English
	enStrings := GetUIStrings("en")
	if enStrings.HelpSelect == "" {
		t.Error("English HelpSelect should not be empty")
	}

	// Test Korean
	koStrings := GetUIStrings("ko")
	if koStrings.HelpSelect == enStrings.HelpSelect {
		t.Error("Korean HelpSelect should be different from English")
	}
	if koStrings.ErrorRequired != "필수 입력 항목입니다" {
		t.Errorf("expected Korean error '필수 입력 항목입니다', got %q", koStrings.ErrorRequired)
	}

	// Test unknown locale (should return English)
	unknownStrings := GetUIStrings("xx")
	if unknownStrings.HelpSelect != enStrings.HelpSelect {
		t.Error("unknown locale should return English strings")
	}
}

// --- Git provider conditional tests ---

func TestGitProviderQuestion(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	// Verify git_provider question exists
	q := QuestionByID(questions, "git_provider")
	if q == nil {
		t.Fatal("git_provider question not found")
	}
	if q.Default != "github" {
		t.Errorf("expected git_provider default 'github', got %q", q.Default)
	}

	// Verify condition: should show for personal/team modes
	result := &WizardResult{GitMode: "personal"}
	if !q.Condition(result) {
		t.Error("git_provider should be visible for personal mode")
	}
	result.GitMode = "team"
	if !q.Condition(result) {
		t.Error("git_provider should be visible for team mode")
	}
	result.GitMode = "manual"
	if q.Condition(result) {
		t.Error("git_provider should be hidden for manual mode")
	}
}

func TestGitLabQuestionsConditional(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	// gitlab_instance_url should only show for gitlab provider
	q := QuestionByID(questions, "gitlab_instance_url")
	if q == nil {
		t.Fatal("gitlab_instance_url question not found")
	}
	result := &WizardResult{GitMode: "personal", GitProvider: "gitlab"}
	if !q.Condition(result) {
		t.Error("gitlab_instance_url should be visible for gitlab provider")
	}
	result.GitProvider = "github"
	if q.Condition(result) {
		t.Error("gitlab_instance_url should be hidden for github provider")
	}

	// gitlab_username should only show for gitlab provider
	q = QuestionByID(questions, "gitlab_username")
	if q == nil {
		t.Fatal("gitlab_username question not found")
	}
	result.GitProvider = "gitlab"
	if !q.Condition(result) {
		t.Error("gitlab_username should be visible for gitlab provider")
	}
	result.GitProvider = "github"
	if q.Condition(result) {
		t.Error("gitlab_username should be hidden for github provider")
	}

	// gitlab_token should only show for gitlab provider
	q = QuestionByID(questions, "gitlab_token")
	if q == nil {
		t.Fatal("gitlab_token question not found")
	}
	result.GitProvider = "gitlab"
	if !q.Condition(result) {
		t.Error("gitlab_token should be visible for gitlab provider")
	}
	result.GitProvider = "github"
	if q.Condition(result) {
		t.Error("gitlab_token should be hidden for github provider")
	}
}

func TestGitHubQuestionsHiddenForGitLab(t *testing.T) {
	questions := DefaultQuestions("/tmp/test-project")

	// github_username should be hidden for gitlab provider
	q := QuestionByID(questions, "github_username")
	if q == nil {
		t.Fatal("github_username question not found")
	}
	result := &WizardResult{GitMode: "personal", GitProvider: "gitlab"}
	if q.Condition(result) {
		t.Error("github_username should be hidden for gitlab provider")
	}
	result.GitProvider = "github"
	if !q.Condition(result) {
		t.Error("github_username should be visible for github provider")
	}

	// github_token should be hidden for gitlab provider
	q = QuestionByID(questions, "github_token")
	if q == nil {
		t.Fatal("github_token question not found")
	}
	result.GitProvider = "gitlab"
	if q.Condition(result) {
		t.Error("github_token should be hidden for gitlab provider")
	}
	result.GitProvider = "github"
	if !q.Condition(result) {
		t.Error("github_token should be visible for github provider")
	}
}

func TestWizardResultGitLabFields(t *testing.T) {
	result := &WizardResult{
		ProjectName:       "test-project",
		Locale:            "en",
		GitMode:           "personal",
		GitProvider:       "gitlab",
		GitLabInstanceURL: "https://gitlab.company.com",
		GitLabUsername:    "gluser",
		GitLabToken:       "glpat-test-token",
	}

	if result.GitProvider != "gitlab" {
		t.Errorf("expected GitProvider 'gitlab', got %q", result.GitProvider)
	}
	if result.GitLabInstanceURL != "https://gitlab.company.com" {
		t.Errorf("expected GitLabInstanceURL 'https://gitlab.company.com', got %q", result.GitLabInstanceURL)
	}
	if result.GitLabUsername != "gluser" {
		t.Errorf("expected GitLabUsername 'gluser', got %q", result.GitLabUsername)
	}
	if result.GitLabToken != "glpat-test-token" {
		t.Errorf("expected GitLabToken 'glpat-test-token', got %q", result.GitLabToken)
	}
}

func TestDevelopmentModeRemovedFromWizard(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")

	for _, q := range questions {
		if q.ID == "development_mode" {
			t.Fatal("development_mode question should not be in wizard; it is auto-configured by /moai project")
		}
	}
}

// --- buildQuestionGroup tests ---

func TestBuildQuestionGroup_Select(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:   "test_select",
		Type: QuestionTypeSelect,
		Options: []Option{
			{Label: "A", Value: "a"},
			{Label: "B", Value: "b"},
		},
	}

	g := buildQuestionGroup(q, result, &locale)
	if g == nil {
		t.Error("expected non-nil group")
	}
}

func TestBuildQuestionGroup_Input(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:      "test_input",
		Type:    QuestionTypeInput,
		Default: "default-val",
	}

	g := buildQuestionGroup(q, result, &locale)
	if g == nil {
		t.Error("expected non-nil group")
	}
}

func TestBuildQuestionGroup_WithCondition(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:   "conditional_q",
		Type: QuestionTypeInput,
		Condition: func(r *WizardResult) bool {
			return r.GitMode == "team"
		},
	}

	g := buildQuestionGroup(q, result, &locale)
	if g == nil {
		t.Error("expected non-nil group even with condition")
	}
}

// --- newMoAIWizardTheme tests ---

func TestNewMoAIWizardTheme_ReturnsNonNil(t *testing.T) {
	theme := newMoAIWizardTheme()
	if theme == nil {
		t.Error("expected non-nil theme")
	}
}

// --- GetLocalizedQuestion edge cases ---

func TestGetLocalizedQuestion_EmptyLocale(t *testing.T) {
	q := &Question{
		ID:    "test_q",
		Title: "Original Title",
	}
	result := GetLocalizedQuestion(q, "")
	if result.Title != "Original Title" {
		t.Errorf("empty locale should return original title, got %q", result.Title)
	}
}

func TestGetLocalizedQuestion_KnownLocaleUnknownID(t *testing.T) {
	q := &Question{
		ID:    "nonexistent_question_id",
		Title: "Original Title",
	}
	result := GetLocalizedQuestion(q, "ko")
	if result.Title != "Original Title" {
		t.Errorf("unknown question ID should return original title, got %q", result.Title)
	}
}

func TestGetLocalizedQuestion_OptionTranslation(t *testing.T) {
	q := &Question{
		ID:    "locale",
		Type:  QuestionTypeSelect,
		Title: "Select language",
		Options: []Option{
			{Label: "Korean (한국어)", Value: "ko", Desc: "Korean"},
			{Label: "English", Value: "en", Desc: "English"},
		},
	}

	// Korean translation should translate option labels
	result := GetLocalizedQuestion(q, "ko")
	if len(result.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(result.Options))
	}
	// Values should be preserved
	if result.Options[0].Value != "ko" {
		t.Errorf("option[0].Value should be 'ko', got %q", result.Options[0].Value)
	}
	if result.Options[1].Value != "en" {
		t.Errorf("option[1].Value should be 'en', got %q", result.Options[1].Value)
	}
}

func TestGetLocalizedQuestion_MismatchedOptionCount(t *testing.T) {
	// Create a question with 3 options but use a locale that has 2 option translations
	q := &Question{
		ID:    "locale",
		Type:  QuestionTypeSelect,
		Title: "Select language",
		Options: []Option{
			{Label: "Korean", Value: "ko"},
			{Label: "English", Value: "en"},
			{Label: "Extra", Value: "extra"}, // Extra option not in translations
		},
	}

	// When option count doesn't match translation count, options should NOT be translated
	result := GetLocalizedQuestion(q, "ko")
	// Title should still be translated
	if result.Title == q.Title {
		t.Error("title should be translated even with mismatched option count")
	}
	// Options should remain original since count mismatch
	if len(result.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(result.Options))
	}
	if result.Options[0].Label != "Korean" {
		t.Errorf("mismatched options should keep original label, got %q", result.Options[0].Label)
	}
}

func TestGetLocalizedQuestion_PreservesNonTranslatedFields(t *testing.T) {
	q := &Question{
		ID:       "locale",
		Type:     QuestionTypeSelect,
		Title:    "Select language",
		Default:  "en",
		Required: true,
		Options: []Option{
			{Label: "Korean", Value: "ko"},
			{Label: "English", Value: "en"},
		},
	}

	result := GetLocalizedQuestion(q, "ko")
	if result.Default != "en" {
		t.Errorf("Default should be preserved, got %q", result.Default)
	}
	if !result.Required {
		t.Error("Required should be preserved")
	}
	if result.Type != QuestionTypeSelect {
		t.Errorf("Type should be preserved, got %v", result.Type)
	}
}

// --- buildSelectField edge cases ---

func TestBuildSelectField_WithDefault(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:      "test_sel",
		Type:    QuestionTypeSelect,
		Title:   "Pick one",
		Default: "b",
		Options: []Option{
			{Label: "A", Value: "a"},
			{Label: "B", Value: "b"},
			{Label: "C", Value: "c"},
		},
	}

	field := buildSelectField(q, result, &locale)
	if field == nil {
		t.Error("expected non-nil select field")
	}
}

func TestBuildSelectField_WithoutDefault(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:   "test_sel_nodef",
		Type: QuestionTypeSelect,
		Options: []Option{
			{Label: "X", Value: "x"},
			{Label: "Y", Value: "y"},
		},
	}

	field := buildSelectField(q, result, &locale)
	if field == nil {
		t.Error("expected non-nil select field")
	}
}

func TestBuildSelectField_WithDescriptions(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:   "test_sel_desc",
		Type: QuestionTypeSelect,
		Options: []Option{
			{Label: "Option A", Value: "a", Desc: "Description A"},
			{Label: "Option B", Value: "b", Desc: ""},
		},
	}

	field := buildSelectField(q, result, &locale)
	if field == nil {
		t.Error("expected non-nil select field")
	}
}

// --- buildInputField edge cases ---

func TestBuildInputField_WithDefault(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:      "test_inp_def",
		Type:    QuestionTypeInput,
		Title:   "Enter name",
		Default: "default-name",
	}

	field := buildInputField(q, result, &locale)
	if field == nil {
		t.Error("expected non-nil input field")
	}
}

func TestBuildInputField_WithoutDefault(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:   "test_inp_nodef",
		Type: QuestionTypeInput,
	}

	field := buildInputField(q, result, &locale)
	if field == nil {
		t.Error("expected non-nil input field")
	}
}

func TestBuildInputField_Required(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:       "test_inp_req",
		Type:     QuestionTypeInput,
		Required: true,
	}

	field := buildInputField(q, result, &locale)
	if field == nil {
		t.Error("expected non-nil input field")
	}
}

func TestBuildInputField_RequiredWithDefault(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:       "test_inp_reqdef",
		Type:     QuestionTypeInput,
		Required: true,
		Default:  "fallback",
	}

	field := buildInputField(q, result, &locale)
	if field == nil {
		t.Error("expected non-nil input field")
	}
}

// --- DefaultQuestions comprehensive coverage ---

func TestDefaultQuestions_AllQuestionTypesValid(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	for _, q := range questions {
		if q.Type != QuestionTypeSelect && q.Type != QuestionTypeInput {
			t.Errorf("question %q has invalid type %v", q.ID, q.Type)
		}
		if q.ID == "" {
			t.Error("question has empty ID")
		}
		if q.Title == "" {
			t.Errorf("question %q has empty title", q.ID)
		}
	}
}

func TestDefaultQuestions_SelectQuestionsHaveOptions(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	for _, q := range questions {
		if q.Type == QuestionTypeSelect && len(q.Options) == 0 {
			t.Errorf("select question %q has no options", q.ID)
		}
	}
}

func TestDefaultQuestions_ConditionalQuestionsHaveConditions(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	conditionalIDs := map[string]bool{
		"git_provider":        true,
		"gitlab_instance_url": true,
		"github_username":     true,
		"github_token":        true,
		"gitlab_username":     true,
		"gitlab_token":        true,
	}

	for _, q := range questions {
		if conditionalIDs[q.ID] && q.Condition == nil {
			t.Errorf("question %q should have a condition", q.ID)
		}
	}
}

func TestDefaultQuestions_UniqueIDs(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	seen := make(map[string]bool)
	for _, q := range questions {
		if seen[q.ID] {
			t.Errorf("duplicate question ID: %q", q.ID)
		}
		seen[q.ID] = true
	}
}

// --- RunWithDefaults test ---

func TestRunWithDefaults_CreatesQuestions(t *testing.T) {
	// RunWithDefaults calls Run which needs TTY, but will fail gracefully.
	// We just verify it doesn't panic and creates questions correctly.
	// Since we can't run the form without TTY, we test that it at least
	// processes questions by verifying DefaultQuestions returns valid data.
	questions := DefaultQuestions("/tmp/run-defaults-test")
	if len(questions) == 0 {
		t.Error("DefaultQuestions should return non-empty for RunWithDefaults")
	}
}

// --- Additional translation coverage ---

func TestGetUIStrings_AllLocales(t *testing.T) {
	locales := []string{"en", "ko", "ja", "zh"}
	for _, locale := range locales {
		str := GetUIStrings(locale)
		if str.HelpSelect == "" {
			t.Errorf("locale %q: HelpSelect should not be empty", locale)
		}
		if str.HelpInput == "" {
			t.Errorf("locale %q: HelpInput should not be empty", locale)
		}
		if str.ErrorRequired == "" {
			t.Errorf("locale %q: ErrorRequired should not be empty", locale)
		}
	}
}

func TestGetLocalizedQuestion_AllLocales_ForLocaleQuestion(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	q := QuestionByID(questions, "locale")
	if q == nil {
		t.Fatal("locale question not found")
	}

	locales := []string{"ko", "ja", "zh"}
	for _, locale := range locales {
		result := GetLocalizedQuestion(q, locale)
		if result.Title == "" {
			t.Errorf("locale %q: title should not be empty", locale)
		}
		if result.Description == "" {
			t.Errorf("locale %q: description should not be empty", locale)
		}
	}
}

// --- buildQuestionGroup tests ---

func TestBuildQuestionGroup_SelectType(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:    "test_group_sel",
		Type:  QuestionTypeSelect,
		Title: "Group select",
		Options: []Option{
			{Label: "A", Value: "a"},
		},
	}

	group := buildQuestionGroup(q, result, &locale)
	if group == nil {
		t.Fatal("expected non-nil group for select question")
	}
}

func TestBuildQuestionGroup_InputType(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:    "test_group_inp",
		Type:  QuestionTypeInput,
		Title: "Group input",
	}

	group := buildQuestionGroup(q, result, &locale)
	if group == nil {
		t.Fatal("expected non-nil group for input question")
	}
}

func TestBuildQuestionGroup_WithCondition_Personal(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:    "conditional_personal",
		Type:  QuestionTypeInput,
		Title: "Conditional Personal",
		Condition: func(r *WizardResult) bool {
			return r.GitMode == "personal"
		},
	}

	group := buildQuestionGroup(q, result, &locale)
	if group == nil {
		t.Fatal("expected non-nil group for conditional question")
	}
}

func TestBuildQuestionGroup_WithoutCondition(t *testing.T) {
	result := &WizardResult{}
	locale := ""
	q := &Question{
		ID:    "no_cond_q",
		Type:  QuestionTypeSelect,
		Title: "Always visible",
		Options: []Option{
			{Label: "Yes", Value: "yes"},
		},
	}

	group := buildQuestionGroup(q, result, &locale)
	if group == nil {
		t.Fatal("expected non-nil group")
	}
}
