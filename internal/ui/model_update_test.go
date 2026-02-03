package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// === Selector Update tests ===

func TestSelectorModel_Init(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil cmd")
	}
}

func TestSelectorModel_Update_KeyDown(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil {
		t.Error("expected nil cmd for KeyDown")
	}
	result := updated.(selectorModel)
	if result.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", result.cursor)
	}
}

func TestSelectorModel_Update_KeyUp(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	result := updated.(selectorModel)
	if result.cursor != 2 {
		t.Errorf("expected cursor 2 (wrapped), got %d", result.cursor)
	}
}

func TestSelectorModel_Update_Enter(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(selectorModel)
	if !result.done {
		t.Error("expected done=true after Enter")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd after Enter")
	}
}

func TestSelectorModel_Update_Escape(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(selectorModel)
	if !result.cancelled {
		t.Error("expected cancelled=true after ESC")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd after ESC")
	}
}

func TestSelectorModel_Update_CtrlC(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(selectorModel)
	if !result.cancelled {
		t.Error("expected cancelled=true after Ctrl+C")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd after Ctrl+C")
	}
}

func TestSelectorModel_Update_Runes_Filter(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("py")})
	result := updated.(selectorModel)
	if result.filter != "py" {
		t.Errorf("expected filter 'py', got %q", result.filter)
	}
	if len(result.filtered) != 1 {
		t.Errorf("expected 1 filtered item, got %d", len(result.filtered))
	}
}

func TestSelectorModel_Update_Backspace(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.filter = "py"
	m.filtered = []SelectItem{testItems()[1]}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	result := updated.(selectorModel)
	if result.filter != "p" {
		t.Errorf("expected filter 'p', got %q", result.filter)
	}
}

func TestSelectorModel_Update_Backspace_Empty(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	result := updated.(selectorModel)
	if result.filter != "" {
		t.Errorf("expected empty filter, got %q", result.filter)
	}
}

func TestSelectorModel_View_Done(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.done = true
	view := m.View()
	if view != "" {
		t.Error("done view should be empty")
	}
}

func TestSelectorModel_View_Cancelled(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.cancelled = true
	view := m.View()
	if view != "" {
		t.Error("cancelled view should be empty")
	}
}

func TestSelectorModel_SelectedValue_EmptyFiltered(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.filtered = nil
	val := m.selectedValue()
	if val != "" {
		t.Errorf("expected empty value for nil filtered, got %q", val)
	}
}

func TestSelectorModel_SelectedValue_CursorOutOfRange(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.cursor = 100
	val := m.selectedValue()
	if val != "go" {
		t.Errorf("expected fallback to first item 'go', got %q", val)
	}
}

func TestSelectorModel_MoveCursorDown_EmptyItems(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", nil)
	m.filtered = nil
	m = m.moveCursorDown()
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 for empty items, got %d", m.cursor)
	}
}

func TestSelectorModel_MoveCursorUp_EmptyItems(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", nil)
	m.filtered = nil
	m = m.moveCursorUp()
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 for empty items, got %d", m.cursor)
	}
}

// === Checkbox Update tests ===

func TestCheckboxModel_Init(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil cmd")
	}
}

func TestCheckboxModel_Update_KeyDown(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := updated.(checkboxModel)
	if result.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", result.cursor)
	}
}

func TestCheckboxModel_Update_KeyUp(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	result := updated.(checkboxModel)
	if result.cursor != 3 {
		t.Errorf("expected cursor 3 (wrapped), got %d", result.cursor)
	}
}

func TestCheckboxModel_Update_Space_Toggle(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	result := updated.(checkboxModel)
	if !result.selected[0] {
		t.Error("expected first item selected after space")
	}
}

func TestCheckboxModel_Update_Enter(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(checkboxModel)
	if !result.done {
		t.Error("expected done=true after Enter")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestCheckboxModel_Update_Escape(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(checkboxModel)
	if !result.cancelled {
		t.Error("expected cancelled=true after ESC")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestCheckboxModel_Update_CtrlC(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(checkboxModel)
	if !result.cancelled {
		t.Error("expected cancelled=true after Ctrl+C")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestCheckboxModel_Update_RuneA_ToggleAll(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	result := updated.(checkboxModel)
	vals := result.selectedValues()
	if len(vals) != 4 {
		t.Errorf("expected 4 selected after 'a', got %d", len(vals))
	}
}

func TestCheckboxModel_Update_OtherRune_Ignored(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	result := updated.(checkboxModel)
	vals := result.selectedValues()
	if len(vals) != 0 {
		t.Errorf("expected 0 selected for 'x' rune, got %d", len(vals))
	}
}

func TestCheckboxModel_View_Done(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m.done = true
	if m.View() != "" {
		t.Error("done view should be empty")
	}
}

func TestCheckboxModel_View_Cancelled(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", checkboxItems())
	m.cancelled = true
	if m.View() != "" {
		t.Error("cancelled view should be empty")
	}
}

func TestCheckboxModel_MoveCursorDown_EmptyItems(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", nil)
	m = m.moveCursorDown()
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

func TestCheckboxModel_MoveCursorUp_EmptyItems(t *testing.T) {
	m := newCheckboxModel(testTheme(), "Features", nil)
	m = m.moveCursorUp()
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

// === Prompt Update tests ===

func TestPromptModel_Init(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil cmd")
	}
}

func TestPromptModel_Update_Enter_Valid(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	m.value = "hello"
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(promptModel)
	if !result.done {
		t.Error("expected done=true after Enter with valid input")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestPromptModel_Update_Enter_Invalid(t *testing.T) {
	validate := func(s string) error {
		if s == "" {
			return errCancelled{}
		}
		return nil
	}
	m := newPromptModel(testTheme(), "Name", inputConfig{validate: validate})
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(promptModel)
	if result.done {
		t.Error("expected done=false after Enter with invalid input")
	}
	if result.errMsg == "" {
		t.Error("expected error message for invalid input")
	}
	if cmd != nil {
		t.Error("expected nil cmd when validation fails")
	}
}

func TestPromptModel_Update_Escape(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(promptModel)
	if !result.cancelled {
		t.Error("expected cancelled=true")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestPromptModel_Update_CtrlC(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := updated.(promptModel)
	if !result.cancelled {
		t.Error("expected cancelled=true")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestPromptModel_Update_Runes(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("abc")})
	result := updated.(promptModel)
	if result.value != "abc" {
		t.Errorf("expected 'abc', got %q", result.value)
	}
}

func TestPromptModel_Update_Backspace(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	m.value = "abc"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	result := updated.(promptModel)
	if result.value != "ab" {
		t.Errorf("expected 'ab', got %q", result.value)
	}
}

func TestPromptModel_View_Done(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	m.done = true
	if m.View() != "" {
		t.Error("done view should be empty")
	}
}

func TestPromptModel_View_Cancelled(t *testing.T) {
	m := newPromptModel(testTheme(), "Name", inputConfig{})
	m.cancelled = true
	if m.View() != "" {
		t.Error("cancelled view should be empty")
	}
}

// === Confirm Update tests ===

func TestConfirmModel_Init(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", true)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil cmd")
	}
}

func TestConfirmModel_Update_Enter(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", true)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := updated.(confirmModel)
	if !result.done {
		t.Error("expected done=true")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestConfirmModel_Update_Escape(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", true)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := updated.(confirmModel)
	if !result.cancelled {
		t.Error("expected cancelled=true")
	}
	if cmd == nil {
		t.Error("expected tea.Quit cmd")
	}
}

func TestConfirmModel_Update_LeftRight_Toggle(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", true)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	result := updated.(confirmModel)
	if result.value {
		t.Error("expected false after Left toggle")
	}
	updated, _ = result.Update(tea.KeyMsg{Type: tea.KeyRight})
	result = updated.(confirmModel)
	if !result.value {
		t.Error("expected true after Right toggle")
	}
}

func TestConfirmModel_Update_RuneY(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", false)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	result := updated.(confirmModel)
	if !result.value {
		t.Error("expected true after 'y'")
	}
}

func TestConfirmModel_Update_RuneN(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", true)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	result := updated.(confirmModel)
	if result.value {
		t.Error("expected false after 'n'")
	}
}

func TestConfirmModel_Update_RuneYUppercase(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", false)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")})
	result := updated.(confirmModel)
	if !result.value {
		t.Error("expected true after 'Y'")
	}
}

func TestConfirmModel_Update_RuneNUppercase(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", true)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	result := updated.(confirmModel)
	if result.value {
		t.Error("expected false after 'N'")
	}
}

func TestConfirmModel_View_Done(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", true)
	m.done = true
	if m.View() != "" {
		t.Error("done view should be empty")
	}
}

func TestConfirmModel_View_Cancelled(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", true)
	m.cancelled = true
	if m.View() != "" {
		t.Error("cancelled view should be empty")
	}
}

func TestConfirmModel_View_ShowsYesHighlightedWhenTrue(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", true)
	view := m.View()
	if !strings.Contains(view, "Yes") {
		t.Error("view should contain 'Yes'")
	}
	if !strings.Contains(view, "No") {
		t.Error("view should contain 'No'")
	}
}

func TestConfirmModel_View_ShowsNoHighlightedWhenFalse(t *testing.T) {
	m := newConfirmModel(testTheme(), "Sure?", false)
	view := m.View()
	if !strings.Contains(view, "Yes") {
		t.Error("view should contain 'Yes'")
	}
	if !strings.Contains(view, "No") {
		t.Error("view should contain 'No'")
	}
}

// === Framework options full coverage ===

func TestFrameworkOptions_AllLanguages(t *testing.T) {
	languages := []string{"Go", "Python", "TypeScript", "Java", "Rust", "PHP", "UnknownLang"}
	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			opts := frameworkOptions(lang)
			if len(opts) == 0 {
				t.Errorf("expected non-empty options for %s", lang)
			}
		})
	}
}

func TestFrameworkOptions_TypeScript(t *testing.T) {
	opts := frameworkOptions("TypeScript")
	found := false
	for _, o := range opts {
		if o.Value == "Next.js" {
			found = true
		}
	}
	if !found {
		t.Error("expected Next.js in TypeScript frameworks")
	}
}

func TestFrameworkOptions_Java(t *testing.T) {
	opts := frameworkOptions("Java")
	found := false
	for _, o := range opts {
		if o.Value == "Spring Boot" {
			found = true
		}
	}
	if !found {
		t.Error("expected Spring Boot in Java frameworks")
	}
}

func TestFrameworkOptions_Rust(t *testing.T) {
	opts := frameworkOptions("Rust")
	found := false
	for _, o := range opts {
		if o.Value == "Axum" {
			found = true
		}
	}
	if !found {
		t.Error("expected Axum in Rust frameworks")
	}
}

func TestFrameworkOptions_PHP(t *testing.T) {
	opts := frameworkOptions("PHP")
	found := false
	for _, o := range opts {
		if o.Value == "Laravel" {
			found = true
		}
	}
	if !found {
		t.Error("expected Laravel in PHP frameworks")
	}
}

// === Selector View with desc on non-cursor items ===

func TestSelectorModel_View_NonCursorItemWithDesc(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.cursor = 0 // first item is cursor
	view := m.View()
	// Non-cursor items (index 1, 2) should show their desc
	if !strings.Contains(view, "Scripting language") {
		t.Error("View should show desc for non-cursor items")
	}
}
