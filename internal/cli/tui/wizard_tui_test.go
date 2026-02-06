// Package tui provides Bubble Tea TUI components for MoAI-ADK.
package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewWizardModel verifies the wizard model is created correctly.
func TestNewWizardModel(t *testing.T) {
	model := NewWizardModel()

	if model.currentScreen != ScreenWelcome {
		t.Errorf("Expected screen to be ScreenWelcome, got %v", model.currentScreen)
	}

	if model.textInput.Placeholder != "my-awesome-project" {
		t.Errorf("Expected placeholder 'my-awesome-project', got '%s'", model.textInput.Placeholder)
	}

	if model.width != 80 {
		t.Errorf("Expected width 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("Expected height 24, got %d", model.height)
	}
}

// TestWizardModelInit verifies Init returns a valid command.
func TestWizardModelInit(t *testing.T) {
	model := NewWizardModel()
	cmd := model.Init()

	if cmd == nil {
		t.Error("Expected Init to return a command, got nil")
	}
}

// TestWizardModelUpdateQuit verifies quit key handling.
func TestWizardModelUpdate(t *testing.T) {
	model := NewWizardModel()

	// Test 'q' key
	m := model
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newModel, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("Expected quit command for key 'q', got nil")
	}
	if !newModel.(WizardModel).quitting {
		t.Error("Expected quitting to be true for key 'q'")
	}

	// Test ctrl+c
	m = model
	msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	newModel, cmd = m.Update(msg)
	if cmd == nil {
		t.Error("Expected quit command for ctrl+c, got nil")
	}
	if !newModel.(WizardModel).quitting {
		t.Error("Expected quitting to be true for ctrl+c")
	}

	// Test esc
	m = model
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd = m.Update(msg)
	if cmd == nil {
		t.Error("Expected quit command for esc, got nil")
	}
	if !newModel.(WizardModel).quitting {
		t.Error("Expected quitting to be true for esc")
	}
}

// TestValidateProjectName verifies project name validation.
func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "my-project", false},
		{"valid with numbers", "project123", false},
		{"valid with underscores", "my_project", false},
		{"empty", "", true},
		{"too short", "a", true},
		{"too long", "this-is-a-very-long-project-name-that-exceeds-fifty-characters", true},
		{"invalid characters", "my project!", true},
		{"special characters", "project@123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// TestMin verifies the min helper function.
func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{0, 10, 0},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := min(tt.a, tt.b); got != tt.want {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestStyles verifies the style set is initialized.
func TestStyles(t *testing.T) {
	if styles == nil {
		t.Fatal("Expected styles to be initialized")
	}

	// Test that styles are set by rendering a simple string
	if styles.title.Render("Test") == "" {
		t.Error("Expected title style to render text")
	}

	if styles.error.Render("Test") == "" {
		t.Error("Expected error style to render text")
	}

	if styles.successTitle.Render("Test") == "" {
		t.Error("Expected successTitle style to render text")
	}
}

// TestScreenString verifies screen string representation.
func TestScreenString(t *testing.T) {
	tests := []struct {
		screen Screen
		want   string
	}{
		{ScreenWelcome, "Welcome"},
		{ScreenProjectName, "ProjectName"},
		{ScreenSuccess, "Success"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.screen.String(); got != tt.want {
				t.Errorf("Screen.String() = %s, want %s", got, tt.want)
			}
		})
	}
}
