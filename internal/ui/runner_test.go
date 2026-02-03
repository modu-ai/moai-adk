package ui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// mockRunner returns a function that simulates tea.Program.Run()
// by returning the model as-is (already in final state).
func mockRunner(m tea.Model) (tea.Model, error) {
	return m, nil
}

// mockRunnerError returns a function that simulates a tea.Program failure.
func mockRunnerError(m tea.Model) (tea.Model, error) {
	return nil, errors.New("program failed")
}

// withMockRunner temporarily replaces programRunner and restores it after the test.
func withMockRunner(t *testing.T, runner func(tea.Model) (tea.Model, error)) {
	t.Helper()
	original := programRunner
	programRunner = runner
	t.Cleanup(func() { programRunner = original })
}

// === Selector interactive path tests ===

func TestSelectorInteractive_ReturnsSelectedValue(t *testing.T) {
	// Mock runner that simulates user selecting the second item
	withMockRunner(t, func(m tea.Model) (tea.Model, error) {
		sm := m.(selectorModel)
		sm.cursor = 1
		sm.done = true
		return sm, nil
	})

	theme := testTheme()
	hm := NewHeadlessManager()
	// Not headless, so it goes to interactive path
	sel := &selectorImpl{theme: theme, headless: hm}
	result, err := sel.selectInteractive("language", testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "python" {
		t.Errorf("expected 'python', got %q", result)
	}
}

func TestSelectorInteractive_Cancelled(t *testing.T) {
	withMockRunner(t, func(m tea.Model) (tea.Model, error) {
		sm := m.(selectorModel)
		sm.cancelled = true
		return sm, nil
	})

	theme := testTheme()
	hm := NewHeadlessManager()
	sel := &selectorImpl{theme: theme, headless: hm}
	_, err := sel.selectInteractive("language", testItems())
	if err != ErrCancelled {
		t.Errorf("expected ErrCancelled, got %v", err)
	}
}

func TestSelectorInteractive_ProgramError(t *testing.T) {
	withMockRunner(t, mockRunnerError)

	theme := testTheme()
	hm := NewHeadlessManager()
	sel := &selectorImpl{theme: theme, headless: hm}
	_, err := sel.selectInteractive("language", testItems())
	if err == nil {
		t.Error("expected error from program failure")
	}
}

// === Checkbox interactive path tests ===

func TestCheckboxInteractive_ReturnsSelectedValues(t *testing.T) {
	withMockRunner(t, func(m tea.Model) (tea.Model, error) {
		cm := m.(checkboxModel)
		cm.selected[0] = true
		cm.selected[2] = true
		cm.done = true
		return cm, nil
	})

	theme := testTheme()
	hm := NewHeadlessManager()
	cb := &checkboxImpl{theme: theme, headless: hm}
	result, err := cb.multiSelectInteractive("features", checkboxItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0] != "lsp" || result[1] != "hooks" {
		t.Errorf("expected [lsp, hooks], got %v", result)
	}
}

func TestCheckboxInteractive_Cancelled(t *testing.T) {
	withMockRunner(t, func(m tea.Model) (tea.Model, error) {
		cm := m.(checkboxModel)
		cm.cancelled = true
		return cm, nil
	})

	theme := testTheme()
	hm := NewHeadlessManager()
	cb := &checkboxImpl{theme: theme, headless: hm}
	_, err := cb.multiSelectInteractive("features", checkboxItems())
	if err != ErrCancelled {
		t.Errorf("expected ErrCancelled, got %v", err)
	}
}

func TestCheckboxInteractive_ProgramError(t *testing.T) {
	withMockRunner(t, mockRunnerError)

	theme := testTheme()
	hm := NewHeadlessManager()
	cb := &checkboxImpl{theme: theme, headless: hm}
	_, err := cb.multiSelectInteractive("features", checkboxItems())
	if err == nil {
		t.Error("expected error from program failure")
	}
}

// === Prompt interactive path tests ===

func TestPromptInteractive_ReturnsValue(t *testing.T) {
	withMockRunner(t, func(m tea.Model) (tea.Model, error) {
		pm := m.(promptModel)
		pm.value = "my-project"
		pm.done = true
		return pm, nil
	})

	theme := testTheme()
	hm := NewHeadlessManager()
	p := &promptImpl{theme: theme, headless: hm}
	result, err := p.inputInteractive("name", inputConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "my-project" {
		t.Errorf("expected 'my-project', got %q", result)
	}
}

func TestPromptInteractive_Cancelled(t *testing.T) {
	withMockRunner(t, func(m tea.Model) (tea.Model, error) {
		pm := m.(promptModel)
		pm.cancelled = true
		return pm, nil
	})

	theme := testTheme()
	hm := NewHeadlessManager()
	p := &promptImpl{theme: theme, headless: hm}
	_, err := p.inputInteractive("name", inputConfig{})
	if err != ErrCancelled {
		t.Errorf("expected ErrCancelled, got %v", err)
	}
}

func TestPromptInteractive_ProgramError(t *testing.T) {
	withMockRunner(t, mockRunnerError)

	theme := testTheme()
	hm := NewHeadlessManager()
	p := &promptImpl{theme: theme, headless: hm}
	_, err := p.inputInteractive("name", inputConfig{})
	if err == nil {
		t.Error("expected error from program failure")
	}
}

// === Confirm interactive path tests ===

func TestConfirmInteractive_ReturnsTrue(t *testing.T) {
	withMockRunner(t, func(m tea.Model) (tea.Model, error) {
		cm := m.(confirmModel)
		cm.value = true
		cm.done = true
		return cm, nil
	})

	theme := testTheme()
	hm := NewHeadlessManager()
	p := &promptImpl{theme: theme, headless: hm}
	result, err := p.confirmInteractive("Sure?", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true")
	}
}

func TestConfirmInteractive_Cancelled(t *testing.T) {
	withMockRunner(t, func(m tea.Model) (tea.Model, error) {
		cm := m.(confirmModel)
		cm.cancelled = true
		return cm, nil
	})

	theme := testTheme()
	hm := NewHeadlessManager()
	p := &promptImpl{theme: theme, headless: hm}
	_, err := p.confirmInteractive("Sure?", true)
	if err != ErrCancelled {
		t.Errorf("expected ErrCancelled, got %v", err)
	}
}

func TestConfirmInteractive_ProgramError(t *testing.T) {
	withMockRunner(t, mockRunnerError)

	theme := testTheme()
	hm := NewHeadlessManager()
	p := &promptImpl{theme: theme, headless: hm}
	_, err := p.confirmInteractive("Sure?", true)
	if err == nil {
		t.Error("expected error from program failure")
	}
}

// === Wizard interactive path via mock ===

func TestWizardInteractive_FullFlow(t *testing.T) {
	callCount := 0
	withMockRunner(t, func(m tea.Model) (tea.Model, error) {
		callCount++
		switch model := m.(type) {
		case promptModel:
			if callCount == 1 {
				model.value = "test-project"
			} else {
				model.value = "TestUser"
			}
			model.done = true
			return model, nil
		case selectorModel:
			model.done = true
			return model, nil
		case checkboxModel:
			model.done = true
			return model, nil
		}
		return m, nil
	})

	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(false) // Force interactive mode so sub-components use mock runner
	w := &wizardImpl{theme: theme, headless: hm}
	result, err := w.runInteractive(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ProjectName != "test-project" {
		t.Errorf("expected 'test-project', got %q", result.ProjectName)
	}
}

func TestWizardInteractive_StepError(t *testing.T) {
	withMockRunner(t, mockRunnerError)

	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(false) // Force interactive mode so sub-components use mock runner
	w := &wizardImpl{theme: theme, headless: hm}
	result, err := w.runInteractive(context.Background())
	if err == nil {
		t.Error("expected error")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestWizardInteractive_ContextCancelled(t *testing.T) {
	withMockRunner(t, mockRunner)

	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(false) // Force interactive mode
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w := &wizardImpl{theme: theme, headless: hm}
	result, err := w.runInteractive(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

// === defaultProgramRunner test ===

func TestDefaultProgramRunner_Exists(t *testing.T) {
	// Verify the default runner is a valid function by assigning it.
	// We cannot actually call it because it needs a TTY.
	var fn func(tea.Model) (tea.Model, error) = defaultProgramRunner
	_ = fn
}
