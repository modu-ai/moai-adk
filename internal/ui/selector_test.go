package ui

import (
	"strings"
	"testing"
)

// --- selectorModel unit tests ---

func testTheme() *Theme {
	return NewTheme(ThemeConfig{NoColor: true})
}

func testItems() []SelectItem {
	return []SelectItem{
		{Label: "Go", Value: "go", Desc: "Compiled language"},
		{Label: "Python", Value: "python", Desc: "Scripting language"},
		{Label: "TypeScript", Value: "ts", Desc: "Typed JavaScript"},
	}
}

func TestNewSelectorModel(t *testing.T) {
	m := newSelectorModel(testTheme(), "Choose language", testItems())
	if m.label != "Choose language" {
		t.Errorf("expected label 'Choose language', got %q", m.label)
	}
	if len(m.items) != 3 {
		t.Errorf("expected 3 items, got %d", len(m.items))
	}
	if m.cursor != 0 {
		t.Error("initial cursor should be 0")
	}
}

func TestSelectorModel_View_ShowsLabel(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick one", testItems())
	view := m.View()
	if !strings.Contains(view, "Pick one") {
		t.Error("View should contain the label")
	}
}

func TestSelectorModel_View_ShowsAllItems(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	view := m.View()
	for _, item := range testItems() {
		if !strings.Contains(view, item.Label) {
			t.Errorf("View should contain item label %q", item.Label)
		}
	}
}

func TestSelectorModel_View_ShowsCursor(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	view := m.View()
	// The cursor indicator should appear for the first item
	if !strings.Contains(view, ">") {
		t.Error("View should contain cursor indicator '>'")
	}
}

func TestSelectorModel_CursorDown(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m = m.moveCursorDown()
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.cursor)
	}
}

func TestSelectorModel_CursorDown_Wraps(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.cursor = 2
	m = m.moveCursorDown()
	if m.cursor != 0 {
		t.Errorf("cursor should wrap to 0, got %d", m.cursor)
	}
}

func TestSelectorModel_CursorUp(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.cursor = 1
	m = m.moveCursorUp()
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}
}

func TestSelectorModel_CursorUp_Wraps(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m = m.moveCursorUp()
	if m.cursor != 2 {
		t.Errorf("cursor should wrap to 2, got %d", m.cursor)
	}
}

func TestSelectorModel_SelectCurrent(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.cursor = 1
	val := m.selectedValue()
	if val != "python" {
		t.Errorf("expected 'python', got %q", val)
	}
}

func TestSelectorModel_FuzzyFilter(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m = m.applyFilter("typ")
	if len(m.filtered) != 1 {
		t.Fatalf("expected 1 filtered item, got %d", len(m.filtered))
	}
	if m.filtered[0].Label != "TypeScript" {
		t.Errorf("expected 'TypeScript', got %q", m.filtered[0].Label)
	}
}

func TestSelectorModel_FuzzyFilter_CaseInsensitive(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m = m.applyFilter("GO")
	if len(m.filtered) != 1 {
		t.Fatalf("expected 1 filtered item, got %d", len(m.filtered))
	}
	if m.filtered[0].Value != "go" {
		t.Errorf("expected 'go', got %q", m.filtered[0].Value)
	}
}

func TestSelectorModel_FuzzyFilter_Empty(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m = m.applyFilter("")
	if len(m.filtered) != 3 {
		t.Errorf("empty filter should show all items, got %d", len(m.filtered))
	}
}

func TestSelectorModel_FuzzyFilter_NoMatch(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m = m.applyFilter("xyz")
	if len(m.filtered) != 0 {
		t.Errorf("expected 0 filtered items for no match, got %d", len(m.filtered))
	}
}

func TestSelectorModel_FuzzyFilter_ResetsCursor(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m.cursor = 2
	m = m.applyFilter("py")
	if m.cursor != 0 {
		t.Errorf("filter should reset cursor to 0, got %d", m.cursor)
	}
}

func TestSelectorModel_SelectedValueFromFiltered(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m = m.applyFilter("py")
	val := m.selectedValue()
	if val != "python" {
		t.Errorf("expected 'python', got %q", val)
	}
}

func TestSelectorModel_ViewWithFilter(t *testing.T) {
	m := newSelectorModel(testTheme(), "Pick", testItems())
	m = m.applyFilter("py")
	view := m.View()
	if !strings.Contains(view, "Python") {
		t.Error("filtered view should contain 'Python'")
	}
	if strings.Contains(view, "TypeScript") {
		t.Error("filtered view should not contain 'TypeScript'")
	}
}

// --- Headless selector tests ---

func TestSelectorHeadless_ReturnsDefault(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"language": "go"})

	sel := NewSelector(theme, hm)
	result, err := sel.Select("language", testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "go" {
		t.Errorf("expected 'go', got %q", result)
	}
}

func TestSelectorHeadless_NoDefault_ReturnsFirstItem(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	sel := NewSelector(theme, hm)
	result, err := sel.Select("language", testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "go" {
		t.Errorf("expected first item 'go', got %q", result)
	}
}

func TestSelectorHeadless_EmptyItems_ReturnsError(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	sel := NewSelector(theme, hm)
	_, err := sel.Select("language", []SelectItem{})
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}

func TestSelector_EmptyItems_ReturnsError(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(false)

	sel := NewSelector(theme, hm)
	_, err := sel.Select("language", []SelectItem{})
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
	}
}
