package ui

import (
	"strings"
	"testing"
)

func checkboxItems() []SelectItem {
	return []SelectItem{
		{Label: "LSP", Value: "lsp"},
		{Label: "Quality Gates", Value: "quality"},
		{Label: "Git Hooks", Value: "hooks"},
		{Label: "Statusline", Value: "statusline"},
	}
}

func TestNewCheckboxModel(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Select features", checkboxItems())
	if m.label != "Select features" {
		t.Errorf("expected label 'Select features', got %q", m.label)
	}
	if len(m.items) != 4 {
		t.Errorf("expected 4 items, got %d", len(m.items))
	}
	if m.cursor != 0 {
		t.Error("initial cursor should be 0")
	}
}

func TestCheckboxModel_View_ShowsLabel(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	view := m.View()
	if !strings.Contains(view, "Features") {
		t.Error("View should contain the label")
	}
}

func TestCheckboxModel_View_ShowsAllItems(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	view := m.View()
	for _, item := range checkboxItems() {
		if !strings.Contains(view, item.Label) {
			t.Errorf("View should contain item label %q", item.Label)
		}
	}
}

func TestCheckboxModel_View_ShowsUnchecked(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	view := m.View()
	if !strings.Contains(view, "[ ]") {
		t.Error("View should show unchecked boxes '[ ]'")
	}
}

func TestCheckboxModel_Toggle(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m = m.toggleCurrent()
	if !m.selected[0] {
		t.Error("first item should be selected after toggle")
	}
	view := m.View()
	if !strings.Contains(view, "[x]") {
		t.Error("View should show checked box '[x]' after toggle")
	}
}

func TestCheckboxModel_ToggleTwice_Deselects(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m = m.toggleCurrent()
	m = m.toggleCurrent()
	if m.selected[0] {
		t.Error("first item should be deselected after double toggle")
	}
}

func TestCheckboxModel_CursorDown(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m = m.moveCursorDown()
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.cursor)
	}
}

func TestCheckboxModel_CursorDown_Wraps(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m.cursor = 3
	m = m.moveCursorDown()
	if m.cursor != 0 {
		t.Errorf("cursor should wrap to 0, got %d", m.cursor)
	}
}

func TestCheckboxModel_CursorUp_Wraps(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m = m.moveCursorUp()
	if m.cursor != 3 {
		t.Errorf("cursor should wrap to 3, got %d", m.cursor)
	}
}

func TestCheckboxModel_SelectedValues(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	// Select first and third items
	m = m.toggleCurrent()                   // lsp
	m = m.moveCursorDown().moveCursorDown() // move to hooks
	m = m.toggleCurrent()                   // hooks
	vals := m.selectedValues()
	if len(vals) != 2 {
		t.Fatalf("expected 2 selected, got %d", len(vals))
	}
	if vals[0] != "lsp" {
		t.Errorf("expected 'lsp', got %q", vals[0])
	}
	if vals[1] != "hooks" {
		t.Errorf("expected 'hooks', got %q", vals[1])
	}
}

func TestCheckboxModel_SelectedValues_None(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	vals := m.selectedValues()
	if len(vals) != 0 {
		t.Errorf("expected 0 selected, got %d", len(vals))
	}
}

func TestCheckboxModel_ToggleAll(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m = m.toggleAll()
	vals := m.selectedValues()
	if len(vals) != 4 {
		t.Errorf("expected 4 selected after toggle all, got %d", len(vals))
	}
}

func TestCheckboxModel_ToggleAll_Twice_DeselectsAll(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m = m.toggleAll()
	m = m.toggleAll()
	vals := m.selectedValues()
	if len(vals) != 0 {
		t.Errorf("expected 0 selected after double toggle all, got %d", len(vals))
	}
}

func TestCheckboxModel_ToggleAll_PartialSelection(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m = m.toggleCurrent() // select first
	m = m.toggleAll()     // should select all (some already selected)
	vals := m.selectedValues()
	if len(vals) != 4 {
		t.Errorf("expected 4 selected, got %d", len(vals))
	}
}

// --- Headless checkbox tests ---

func TestCheckboxHeadless_ReturnsDefaults(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"features": "lsp,quality"})

	cb := NewCheckbox(theme, hm)
	result, err := cb.MultiSelect("features", checkboxItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0] != "lsp" || result[1] != "quality" {
		t.Errorf("expected [lsp, quality], got %v", result)
	}
}

func TestCheckboxHeadless_NoDefaults_ReturnsEmpty(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	cb := NewCheckbox(theme, hm)
	result, err := cb.MultiSelect("features", checkboxItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestCheckboxHeadless_EmptyItems_ReturnsError(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	cb := NewCheckbox(theme, hm)
	_, err := cb.MultiSelect("features", []SelectItem{})
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}

func TestCheckbox_EmptyItems_ReturnsError(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()

	cb := NewCheckbox(theme, hm)
	_, err := cb.MultiSelect("features", []SelectItem{})
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}
