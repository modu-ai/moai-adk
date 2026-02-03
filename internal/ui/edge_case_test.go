package ui

import (
	"context"
	"strings"
	"testing"
)

// --- renderASCIIBar edge cases ---

func TestRenderASCIIBar_OverTotal(t *testing.T) {
	bar := renderASCIIBar(15, 10, 20)
	// ratio capped at 1.0, so should be all filled
	if !strings.HasPrefix(bar, "[") || !strings.HasSuffix(bar, "]") {
		t.Error("bar should have brackets")
	}
	filled := strings.Count(bar, "=")
	if filled != 20 {
		t.Errorf("expected 20 '=' for overflow, got %d", filled)
	}
}

func TestRenderASCIIBar_ExactlyFull(t *testing.T) {
	bar := renderASCIIBar(10, 10, 20)
	filled := strings.Count(bar, "=")
	if filled != 20 {
		t.Errorf("expected 20 '=' for 100%%, got %d in %q", filled, bar)
	}
}

func TestRenderASCIIBar_ZeroWidth(t *testing.T) {
	bar := renderASCIIBar(5, 10, 0)
	if bar != "[]" {
		t.Errorf("expected '[]' for zero width, got %q", bar)
	}
}

// --- IsHeadless TTY detection paths ---

func TestHeadlessManager_IsHeadless_UnforcedPath(t *testing.T) {
	hm := NewHeadlessManager()
	// Just exercise the TTY detection path without forcing
	// The result depends on whether tests run in a TTY
	_ = hm.IsHeadless()
}

func TestHeadlessManager_ForceHeadless_ThenUnforce(t *testing.T) {
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	if !hm.IsHeadless() {
		t.Error("expected headless after force true")
	}
	hm.ForceHeadless(false)
	// Now it falls back to TTY detection
	_ = hm.IsHeadless()
}

// --- Wizard runHeadless context check ---

func TestWizardHeadless_ContextCancelled_AfterDefaultsCheck(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{
		"project_name": "proj",
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w := NewWizard(theme, hm)
	result, err := w.Run(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

// --- Selector View with filter showing filtered items ---

func TestSelectorModel_View_FilterPromptShown(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.filter = "go"
	m = m.applyFilter("go")
	view := m.View()
	if !strings.Contains(view, "Filter:") {
		t.Error("View should show filter prompt when filter is active")
	}
	if !strings.Contains(view, "go") {
		t.Error("View should show filter text")
	}
}

func TestSelectorModel_View_NoDesc(t *testing.T) {
	items := []SelectItem{
		{Label: "Option A", Value: "a"},
		{Label: "Option B", Value: "b"},
	}
	m := newSelectorModel(testTheme(), "Pick", items)
	view := m.View()
	if !strings.Contains(view, "Option A") {
		t.Error("View should contain Option A")
	}
	// Should not have " - " separator since no Desc
	lines := strings.Split(view, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Option A") {
			if strings.Contains(line, " - ") {
				t.Error("Option A line should not contain desc separator")
			}
		}
	}
}

// --- Checkbox view with cursor on non-first item ---

func TestCheckboxModel_View_CursorOnSecondItem(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m.cursor = 1
	view := m.View()
	if !strings.Contains(view, ">") {
		t.Error("View should contain cursor indicator")
	}
}

// --- Selector with items that have descriptions for cursor item ---

func TestSelectorModel_View_CursorItemWithDesc(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.cursor = 0
	view := m.View()
	// First item (Go) has desc "Compiled language"
	if !strings.Contains(view, "Compiled language") {
		t.Error("cursor item should show its description")
	}
}

// --- Non-interactive path for Select when not headless ---

func TestSelector_NonHeadless_EmptyItems(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	// Don't force headless, but still test empty items validation
	sel := NewSelector(theme, hm)
	_, err := sel.Select("language", []SelectItem{})
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}

func TestCheckbox_NonHeadless_EmptyItems(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	cb := NewCheckbox(theme, hm)
	_, err := cb.MultiSelect("features", []SelectItem{})
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}

// --- Prompt Confirm view with both true and false values ---

func TestConfirmModel_View_DefaultTrue(t *testing.T) {
	m := newConfirmModel(testTheme(), "Proceed?", true)
	view := m.View()
	if !strings.Contains(view, "Yes") {
		t.Error("should contain Yes")
	}
}

func TestConfirmModel_View_DefaultFalse(t *testing.T) {
	m := newConfirmModel(testTheme(), "Proceed?", false)
	view := m.View()
	if !strings.Contains(view, "No") {
		t.Error("should contain No")
	}
}

// --- Prompt addRunes clears error ---

func TestPromptModel_AddRunes_ClearsError(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	m.errMsg = "some error"
	m = m.addRunes([]rune("a"))
	if m.errMsg != "" {
		t.Errorf("expected errMsg cleared, got %q", m.errMsg)
	}
}

func TestPromptModel_Backspace_ClearsError(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	m.value = "abc"
	m.errMsg = "some error"
	m = m.backspace()
	if m.errMsg != "" {
		t.Errorf("expected errMsg cleared, got %q", m.errMsg)
	}
}

// --- ValidateInput with nil validator ---

func TestPromptModel_ValidateInput_NilValidator(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	err := m.validateInput()
	if err != nil {
		t.Errorf("expected nil error for nil validator, got %v", err)
	}
}
