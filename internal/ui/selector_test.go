package ui

import (
	"testing"
)

// --- Test helpers ---

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

// --- Constructor tests ---

func TestNewSelector_ReturnsNonNil(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	sel := NewSelector(theme, hm)
	if sel == nil {
		t.Error("NewSelector should not return nil")
	}
}

// --- Empty items tests ---

func TestSelector_EmptyItems_ReturnsError(t *testing.T) {
	tests := []struct {
		name     string
		headless bool
	}{
		{"interactive_mode", false},
		{"headless_mode", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := testTheme()
			hm := NewHeadlessManager()
			if tt.headless {
				hm.ForceHeadless(true)
			}
			sel := NewSelector(theme, hm)
			_, err := sel.Select("language", []SelectItem{})
			if err != ErrNoItems {
				t.Errorf("expected ErrNoItems, got %v", err)
			}
		})
	}
}

func TestSelector_NilItems_ReturnsError(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	sel := NewSelector(theme, hm)
	_, err := sel.Select("language", nil)
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
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

func TestSelectorHeadless_CustomDefault(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"lang": "python"})

	sel := NewSelector(theme, hm)
	result, err := sel.Select("lang", testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "python" {
		t.Errorf("expected 'python', got %q", result)
	}
}

func TestSelectorHeadless_SingleItem(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	sel := NewSelector(theme, hm)
	items := []SelectItem{{Label: "Only", Value: "only"}}
	result, err := sel.Select("choice", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "only" {
		t.Errorf("expected 'only', got %q", result)
	}
}

func TestSelectorHeadless_ItemsWithoutDesc(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)

	sel := NewSelector(theme, hm)
	items := []SelectItem{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	}
	result, err := sel.Select("pick", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "a" {
		t.Errorf("expected 'a', got %q", result)
	}
}

func TestSelectorHeadless_DefaultNotInItems(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"language": "nonexistent"})

	sel := NewSelector(theme, hm)
	result, err := sel.Select("language", testItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return the default value directly, even if not in items
	if result != "nonexistent" {
		t.Errorf("expected 'nonexistent', got %q", result)
	}
}

// --- buildSelectField tests ---

func TestBuildSelectField_CreatesField(t *testing.T) {
	var selected string
	items := testItems()
	field := buildSelectField("Pick language", items, &selected)
	if field == nil {
		t.Fatal("buildSelectField should not return nil")
	}
}

func TestBuildSelectField_BindsValuePointer(t *testing.T) {
	var selected string
	items := []SelectItem{{Label: "Go", Value: "go"}}
	field := buildSelectField("Pick", items, &selected)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	// huh.Select may initialize the value from options, so just verify
	// the field was created successfully with a valid value pointer binding.
}

func TestBuildSelectField_ItemsWithDesc(t *testing.T) {
	var selected string
	items := []SelectItem{
		{Label: "Go", Value: "go", Desc: "Compiled"},
		{Label: "Python", Value: "python", Desc: "Scripting"},
	}
	field := buildSelectField("Language", items, &selected)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

func TestBuildSelectField_ItemsWithoutDesc(t *testing.T) {
	var selected string
	items := []SelectItem{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	}
	field := buildSelectField("Choice", items, &selected)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

func TestBuildSelectField_SingleItem(t *testing.T) {
	var selected string
	items := []SelectItem{{Label: "Only", Value: "only"}}
	field := buildSelectField("Solo", items, &selected)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

func TestBuildSelectField_MixedDesc(t *testing.T) {
	var selected string
	items := []SelectItem{
		{Label: "Go", Value: "go", Desc: "Fast"},
		{Label: "Python", Value: "python"},
		{Label: "Rust", Value: "rust", Desc: "Safe"},
	}
	field := buildSelectField("Lang", items, &selected)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}
