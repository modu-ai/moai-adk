package wizard

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

func TestModelNew(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	model := New(questions, nil)

	if model.state != StateRunning {
		t.Errorf("expected state StateRunning, got %v", model.state)
	}
	if model.currentIndex != 0 {
		t.Errorf("expected currentIndex 0, got %d", model.currentIndex)
	}
	if model.visibleIndex != 1 {
		t.Errorf("expected visibleIndex 1, got %d", model.visibleIndex)
	}
	if model.result == nil {
		t.Error("result should not be nil")
	}
	if model.styles == nil {
		t.Error("styles should not be nil")
	}
}

func TestModelInit(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	model := New(questions, nil)

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestModelUpdateCancel(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	model := New(questions, nil)

	// Test Ctrl+C cancellation
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m := newModel.(Model)

	if m.state != StateCancelled {
		t.Errorf("expected state StateCancelled after Ctrl+C, got %v", m.state)
	}
	if cmd == nil {
		t.Error("expected tea.Quit command after cancellation")
	}

	// Test Esc cancellation
	model = New(questions, nil)
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(Model)

	if m.state != StateCancelled {
		t.Errorf("expected state StateCancelled after Esc, got %v", m.state)
	}
}

func TestModelSelectNavigation(t *testing.T) {
	questions := []Question{
		{
			ID:   "test",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
				{Label: "C", Value: "c"},
			},
		},
	}
	model := New(questions, nil)

	// Navigate down
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after KeyDown, got %d", m.cursor)
	}

	// Navigate down again (wrap around)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after wrap, got %d", m.cursor)
	}

	// Navigate up (wrap around)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(Model)
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 after KeyUp wrap, got %d", m.cursor)
	}
}

func TestModelSelectEnter(t *testing.T) {
	questions := []Question{
		{
			ID:   "locale",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "English", Value: "en"},
				{Label: "Korean", Value: "ko"},
			},
		},
		{
			ID:       "project_name",
			Type:     QuestionTypeInput,
			Default:  "test",
			Required: true,
		},
	}
	model := New(questions, nil)

	// Move to second option and select
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := newModel.(Model)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.result.Locale != "ko" {
		t.Errorf("expected Locale 'ko', got %q", m.result.Locale)
	}
	if m.currentIndex != 1 {
		t.Errorf("expected currentIndex 1, got %d", m.currentIndex)
	}
}

func TestModelTextInput(t *testing.T) {
	questions := []Question{
		{
			ID:       "user_name",
			Type:     QuestionTypeInput,
			Default:  "",
			Required: false,
		},
		{
			ID:       "project_name",
			Type:     QuestionTypeInput,
			Default:  "default-project",
			Required: true,
		},
	}
	model := New(questions, nil)

	// Type some text
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("John")})
	m := newModel.(Model)
	if m.inputValue != "John" {
		t.Errorf("expected inputValue 'John', got %q", m.inputValue)
	}

	// Backspace
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = newModel.(Model)
	if m.inputValue != "Joh" {
		t.Errorf("expected inputValue 'Joh', got %q", m.inputValue)
	}
}

func TestModelTextInputSubmit(t *testing.T) {
	questions := []Question{
		{
			ID:       "user_name",
			Type:     QuestionTypeInput,
			Default:  "",
			Required: false,
		},
		{
			ID:       "project_name",
			Type:     QuestionTypeInput,
			Default:  "test",
			Required: true,
		},
	}
	model := New(questions, nil)

	// Submit empty (allowed for non-required)
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(Model)

	if m.result.UserName != "" {
		t.Errorf("expected empty UserName, got %q", m.result.UserName)
	}
	if m.currentIndex != 1 {
		t.Errorf("expected currentIndex 1, got %d", m.currentIndex)
	}
}

func TestModelRequiredField(t *testing.T) {
	questions := []Question{
		{
			ID:       "project_name",
			Type:     QuestionTypeInput,
			Default:  "",
			Required: true,
		},
	}
	model := New(questions, nil)
	model.inputValue = "" // Clear default

	// Try to submit empty required field
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(Model)

	if m.errorMsg == "" {
		t.Error("expected error message for empty required field")
	}
	if m.currentIndex != 0 {
		t.Error("should not advance on validation error")
	}
}

func TestModelView(t *testing.T) {
	questions := []Question{
		{
			ID:          "locale",
			Type:        QuestionTypeSelect,
			Title:       "Select language",
			Description: "Choose your language",
			Options: []Option{
				{Label: "English", Value: "en"},
				{Label: "Korean", Value: "ko"},
			},
		},
	}
	model := New(questions, nil)

	view := model.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// Check for progress indicator
	if !contains(view, "[1/") {
		t.Error("View should contain progress indicator")
	}

	// Check for title
	if !contains(view, "Select language") {
		t.Error("View should contain question title")
	}
}

func TestModelViewCompleted(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	model := New(questions, nil)
	model.state = StateCompleted

	view := model.View()
	if view != "" {
		t.Error("View() should return empty string when completed")
	}
}

func TestModelViewCancelled(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	model := New(questions, nil)
	model.state = StateCancelled

	view := model.View()
	if view != "" {
		t.Error("View() should return empty string when cancelled")
	}
}

func TestConditionalQuestionSkip(t *testing.T) {
	questions := []Question{
		{
			ID:   "git_mode",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "Manual", Value: "manual"},
				{Label: "Personal", Value: "personal"},
			},
		},
		{
			ID:       "github_username",
			Type:     QuestionTypeInput,
			Default:  "",
			Required: false,
			Condition: func(r *WizardResult) bool {
				return r.GitMode == "personal" || r.GitMode == "team"
			},
		},
		{
			ID:       "next_question",
			Type:     QuestionTypeInput,
			Default:  "test",
			Required: false,
		},
	}
	model := New(questions, nil)

	// Select "manual" mode (index 0)
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(Model)

	// Should skip github_username and go to next_question
	if m.currentIndex != 2 {
		t.Errorf("expected currentIndex 2 (skipping conditional), got %d", m.currentIndex)
	}

	// Verify the current question is next_question
	q := m.currentQuestion()
	if q.ID != "next_question" {
		t.Errorf("expected current question 'next_question', got %q", q.ID)
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

func TestModelResult(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	model := New(questions, nil)

	// Modify the result
	model.result.ProjectName = "my-project"
	model.result.Locale = "ko"

	result := model.Result()
	if result == nil {
		t.Fatal("Result() returned nil")
	}
	if result.ProjectName != "my-project" {
		t.Errorf("expected ProjectName 'my-project', got %q", result.ProjectName)
	}
	if result.Locale != "ko" {
		t.Errorf("expected Locale 'ko', got %q", result.Locale)
	}
}

func TestModelState(t *testing.T) {
	questions := DefaultQuestions("/tmp/test")
	model := New(questions, nil)

	// Test initial state
	if model.State() != StateRunning {
		t.Errorf("expected StateRunning, got %v", model.State())
	}

	// Test completed state
	model.state = StateCompleted
	if model.State() != StateCompleted {
		t.Errorf("expected StateCompleted, got %v", model.State())
	}

	// Test cancelled state
	model.state = StateCancelled
	if model.State() != StateCancelled {
		t.Errorf("expected StateCancelled, got %v", model.State())
	}
}

func TestModelViewWithInputQuestion(t *testing.T) {
	questions := []Question{
		{
			ID:          "user_name",
			Type:        QuestionTypeInput,
			Title:       "Enter your name",
			Description: "This will be used for commits",
			Default:     "defaultuser",
			Required:    false,
		},
	}
	model := New(questions, nil)

	view := model.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// Check for title
	if !contains(view, "Enter your name") {
		t.Error("View should contain question title")
	}

	// Check for input prompt
	if !contains(view, ">") {
		t.Error("View should contain input prompt '>'")
	}
}

func TestModelViewWithInputNoDefault(t *testing.T) {
	questions := []Question{
		{
			ID:       "user_name",
			Type:     QuestionTypeInput,
			Title:    "Enter your name",
			Default:  "",
			Required: false,
		},
	}
	model := New(questions, nil)
	model.inputValue = "" // Ensure empty

	view := model.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestModelViewWithInputValue(t *testing.T) {
	questions := []Question{
		{
			ID:       "user_name",
			Type:     QuestionTypeInput,
			Title:    "Enter your name",
			Default:  "",
			Required: false,
		},
	}
	model := New(questions, nil)
	model.inputValue = "typed-value"

	view := model.View()
	if !contains(view, "typed-value") {
		t.Error("View should contain typed input value")
	}
}

func TestModelViewWithError(t *testing.T) {
	questions := []Question{
		{
			ID:       "project_name",
			Type:     QuestionTypeInput,
			Title:    "Project name",
			Default:  "",
			Required: true,
		},
	}
	model := New(questions, nil)
	model.errorMsg = "This field is required"

	view := model.View()
	if !contains(view, "This field is required") {
		t.Error("View should contain error message")
	}
}

func TestSaveAnswerAllCases(t *testing.T) {
	questions := []Question{
		{ID: "locale", Type: QuestionTypeSelect, Options: []Option{{Value: "ko"}}},
		{ID: "user_name", Type: QuestionTypeInput},
		{ID: "project_name", Type: QuestionTypeInput},
		{ID: "git_mode", Type: QuestionTypeSelect, Options: []Option{{Value: "personal"}}},
		{ID: "git_provider", Type: QuestionTypeSelect, Options: []Option{{Value: "github"}, {Value: "gitlab"}}},
		{ID: "github_username", Type: QuestionTypeInput},
		{ID: "git_commit_lang", Type: QuestionTypeSelect, Options: []Option{{Value: "en"}}},
		{ID: "code_comment_lang", Type: QuestionTypeSelect, Options: []Option{{Value: "en"}}},
		{ID: "doc_lang", Type: QuestionTypeSelect, Options: []Option{{Value: "ko"}}},
		{ID: "agent_teams_mode", Type: QuestionTypeSelect, Options: []Option{{Value: "auto"}}},
		{ID: "max_teammates", Type: QuestionTypeSelect, Options: []Option{{Value: "3"}}},
		{ID: "default_model", Type: QuestionTypeSelect, Options: []Option{{Value: "sonnet"}}},
		{ID: "github_token", Type: QuestionTypeInput},
		{ID: "gitlab_instance_url", Type: QuestionTypeInput},
		{ID: "gitlab_username", Type: QuestionTypeInput},
		{ID: "gitlab_token", Type: QuestionTypeInput},
	}
	model := New(questions, nil)

	// Test all saveAnswer cases
	model.saveAnswer("locale", "ko")
	if model.result.Locale != "ko" {
		t.Errorf("expected Locale 'ko', got %q", model.result.Locale)
	}

	model.saveAnswer("user_name", "testuser")
	if model.result.UserName != "testuser" {
		t.Errorf("expected UserName 'testuser', got %q", model.result.UserName)
	}

	model.saveAnswer("project_name", "myproject")
	if model.result.ProjectName != "myproject" {
		t.Errorf("expected ProjectName 'myproject', got %q", model.result.ProjectName)
	}

	model.saveAnswer("git_mode", "personal")
	if model.result.GitMode != "personal" {
		t.Errorf("expected GitMode 'personal', got %q", model.result.GitMode)
	}

	model.saveAnswer("github_username", "ghuser")
	if model.result.GitHubUsername != "ghuser" {
		t.Errorf("expected GitHubUsername 'ghuser', got %q", model.result.GitHubUsername)
	}

	model.saveAnswer("git_commit_lang", "en")
	if model.result.GitCommitLang != "en" {
		t.Errorf("expected GitCommitLang 'en', got %q", model.result.GitCommitLang)
	}

	model.saveAnswer("code_comment_lang", "en")
	if model.result.CodeCommentLang != "en" {
		t.Errorf("expected CodeCommentLang 'en', got %q", model.result.CodeCommentLang)
	}

	model.saveAnswer("doc_lang", "ko")
	if model.result.DocLang != "ko" {
		t.Errorf("expected DocLang 'ko', got %q", model.result.DocLang)
	}

	model.saveAnswer("agent_teams_mode", "subagent")
	if model.result.AgentTeamsMode != "subagent" {
		t.Errorf("expected AgentTeamsMode 'subagent', got %q", model.result.AgentTeamsMode)
	}

	model.saveAnswer("max_teammates", "4")
	if model.result.MaxTeammates != "4" {
		t.Errorf("expected MaxTeammates '4', got %q", model.result.MaxTeammates)
	}

	model.saveAnswer("default_model", "opus")
	if model.result.DefaultModel != "opus" {
		t.Errorf("expected DefaultModel 'opus', got %q", model.result.DefaultModel)
	}

	model.saveAnswer("github_token", "ghp_test_token")
	if model.result.GitHubToken != "ghp_test_token" {
		t.Errorf("expected GitHubToken 'ghp_test_token', got %q", model.result.GitHubToken)
	}

	model.saveAnswer("git_provider", "gitlab")
	if model.result.GitProvider != "gitlab" {
		t.Errorf("expected GitProvider 'gitlab', got %q", model.result.GitProvider)
	}

	model.saveAnswer("gitlab_instance_url", "https://gitlab.company.com")
	if model.result.GitLabInstanceURL != "https://gitlab.company.com" {
		t.Errorf("expected GitLabInstanceURL 'https://gitlab.company.com', got %q", model.result.GitLabInstanceURL)
	}

	model.saveAnswer("gitlab_username", "gluser")
	if model.result.GitLabUsername != "gluser" {
		t.Errorf("expected GitLabUsername 'gluser', got %q", model.result.GitLabUsername)
	}

	model.saveAnswer("gitlab_token", "glpat-test-token")
	if model.result.GitLabToken != "glpat-test-token" {
		t.Errorf("expected GitLabToken 'glpat-test-token', got %q", model.result.GitLabToken)
	}
}

func TestRenderSelectWithDescription(t *testing.T) {
	questions := []Question{
		{
			ID:   "test",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "Option A", Value: "a", Desc: "Description for A"},
				{Label: "Option B", Value: "b", Desc: "Description for B"},
			},
		},
	}
	model := New(questions, nil)

	view := model.View()
	if !contains(view, "Description for A") {
		t.Error("View should contain option description")
	}
}

func TestRenderHelpForInput(t *testing.T) {
	questions := []Question{
		{
			ID:       "test",
			Type:     QuestionTypeInput,
			Title:    "Test input",
			Required: false,
		},
	}
	model := New(questions, nil)

	view := model.View()
	if !contains(view, "Enter to confirm") {
		t.Error("View should contain input help text")
	}
}

func TestAdvanceToCompletion(t *testing.T) {
	questions := []Question{
		{
			ID:   "only_question",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "Yes", Value: "yes"},
			},
		},
	}
	model := New(questions, nil)

	// Select and advance (should complete since only one question)
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(Model)

	if m.state != StateCompleted {
		t.Errorf("expected StateCompleted, got %v", m.state)
	}
	if cmd == nil {
		t.Error("expected tea.Quit command on completion")
	}
}

func TestCurrentQuestionNil(t *testing.T) {
	questions := []Question{
		{
			ID:   "only",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "Yes", Value: "yes"},
			},
		},
	}
	model := New(questions, nil)
	model.currentIndex = 999 // Beyond questions

	q := model.currentQuestion()
	if q != nil {
		t.Error("expected nil when currentIndex is beyond questions")
	}
}

func TestModelViewNilQuestion(t *testing.T) {
	questions := []Question{
		{
			ID:   "only",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "Yes", Value: "yes"},
			},
		},
	}
	model := New(questions, nil)
	model.currentIndex = 999 // Beyond questions

	view := model.View()
	if view != "" {
		t.Error("View should return empty string when no current question")
	}
}

func TestHandleKeyMsgNilQuestion(t *testing.T) {
	questions := []Question{
		{
			ID:   "only",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "Yes", Value: "yes"},
			},
		},
	}
	model := New(questions, nil)
	model.currentIndex = 999 // Beyond questions

	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(Model)

	if m.state != StateCompleted {
		t.Errorf("expected StateCompleted when no questions left, got %v", m.state)
	}
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

func TestTabNavigation(t *testing.T) {
	questions := []Question{
		{
			ID:   "test",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
				{Label: "C", Value: "c"},
			},
		},
	}
	model := New(questions, nil)

	// Tab should move down
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	m := newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after Tab, got %d", m.cursor)
	}

	// ShiftTab should move up
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after ShiftTab, got %d", m.cursor)
	}
}

func TestSubmitTextInputWithDefault(t *testing.T) {
	questions := []Question{
		{
			ID:       "project_name",
			Type:     QuestionTypeInput,
			Default:  "default-project",
			Required: true,
		},
		{
			ID:      "next",
			Type:    QuestionTypeInput,
			Default: "next-default",
		},
	}
	model := New(questions, nil)
	model.inputValue = "" // Empty, should use default

	// Submit with empty input (should use default)
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(Model)

	if m.result.ProjectName != "default-project" {
		t.Errorf("expected ProjectName 'default-project', got %q", m.result.ProjectName)
	}
}

func TestAdvanceWithDefaultOption(t *testing.T) {
	questions := []Question{
		{
			ID:   "first",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
			},
		},
		{
			ID:      "second",
			Type:    QuestionTypeSelect,
			Default: "b",
			Options: []Option{
				{Label: "A", Value: "a"},
				{Label: "B", Value: "b"},
			},
		},
	}
	model := New(questions, nil)

	// Select first option and advance
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(Model)

	// Cursor should be set to default option index (1 for "b")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 (default option 'b'), got %d", m.cursor)
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

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

func TestLocaleTransitionInWizard(t *testing.T) {
	questions := []Question{
		{
			ID:   "locale",
			Type: QuestionTypeSelect,
			Options: []Option{
				{Label: "Korean (한국어)", Value: "ko"},
				{Label: "English", Value: "en"},
			},
		},
		{
			ID:          "user_name",
			Type:        QuestionTypeInput,
			Title:       "Enter your name",
			Description: "This will be used in configuration files.",
			Default:     "",
			Required:    false,
		},
	}
	model := New(questions, nil)

	// Initially locale is empty, so English is used
	view1 := model.View()
	if !contains(view1, "Korean") {
		t.Error("initial view should show locale options")
	}

	// Select Korean and advance
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Select first option (ko)
	m := newModel.(Model)

	if m.locale != "ko" {
		t.Errorf("expected locale 'ko', got %q", m.locale)
	}

	// Next question should be in Korean
	view2 := m.View()
	if !contains(view2, "이름 입력") {
		t.Errorf("second question should be in Korean, got: %s", view2)
	}
}

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
