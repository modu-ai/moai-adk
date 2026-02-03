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
		DevelopmentMode: "ddd",
	}

	if result.ProjectName != "test-project" {
		t.Errorf("expected ProjectName 'test-project', got %q", result.ProjectName)
	}
	if result.Locale != "ko" {
		t.Errorf("expected Locale 'ko', got %q", result.Locale)
	}
	if result.DevelopmentMode != "ddd" {
		t.Errorf("expected DevelopmentMode 'ddd', got %q", result.DevelopmentMode)
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

	// Verify development_mode question exists
	found = false
	for _, q := range questions {
		if q.ID == "development_mode" {
			found = true
			if q.Type != QuestionTypeSelect {
				t.Errorf("expected development_mode type QuestionTypeSelect, got %v", q.Type)
			}
			break
		}
	}
	if !found {
		t.Error("development_mode question not found")
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
	newModel, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
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
