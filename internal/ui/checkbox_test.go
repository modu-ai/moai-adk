package ui

import (
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

// --- Constructor tests ---

func TestNewCheckbox_ReturnsNonNil(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	cb := NewCheckbox(theme, hm)
	if cb == nil {
		t.Error("NewCheckbox should not return nil")
	}
}

// --- Empty items tests ---

func TestCheckbox_EmptyItems_ReturnsError(t *testing.T) {
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
			cb := NewCheckbox(theme, hm)
			_, err := cb.MultiSelect("features", []SelectItem{})
			if err != ErrNoItems {
				t.Errorf("expected ErrNoItems, got %v", err)
			}
		})
	}
}

func TestCheckbox_NilItems_ReturnsError(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	cb := NewCheckbox(theme, hm)
	_, err := cb.MultiSelect("features", nil)
	if err != ErrNoItems {
		t.Errorf("expected ErrNoItems, got %v", err)
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

func TestCheckboxHeadless_EmptyDefaultValue(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"features": ""})

	cb := NewCheckbox(theme, hm)
	result, err := cb.MultiSelect("features", checkboxItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result for empty default, got %v", result)
	}
}

func TestCheckboxHeadless_SingleDefault(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"features": "hooks"})

	cb := NewCheckbox(theme, hm)
	result, err := cb.MultiSelect("features", checkboxItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "hooks" {
		t.Errorf("expected [hooks], got %v", result)
	}
}

func TestCheckboxHeadless_DefaultsWithSpaces(t *testing.T) {
	theme := testTheme()
	hm := NewHeadlessManager()
	hm.ForceHeadless(true)
	hm.SetDefaults(map[string]string{"features": " lsp , quality , hooks "})

	cb := NewCheckbox(theme, hm)
	result, err := cb.MultiSelect("features", checkboxItems())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
	if result[0] != "lsp" || result[1] != "quality" || result[2] != "hooks" {
		t.Errorf("expected [lsp, quality, hooks], got %v", result)
	}
}

// --- buildMultiSelectField tests ---

func TestBuildMultiSelectField_CreatesField(t *testing.T) {
	var selected []string
	items := checkboxItems()
	field := buildMultiSelectField("Pick features", items, &selected)
	if field == nil {
		t.Fatal("buildMultiSelectField should not return nil")
	}
}

func TestBuildMultiSelectField_BindsValuePointer(t *testing.T) {
	var selected []string
	items := checkboxItems()
	field := buildMultiSelectField("Features", items, &selected)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
	// Value pointer should be bound but not yet populated
	if selected != nil {
		t.Errorf("expected nil initial value, got %v", selected)
	}
}

func TestBuildMultiSelectField_ItemsWithDesc(t *testing.T) {
	var selected []string
	items := []SelectItem{
		{Label: "LSP", Value: "lsp", Desc: "Language Server"},
		{Label: "Hooks", Value: "hooks", Desc: "Git Hooks"},
	}
	field := buildMultiSelectField("Features", items, &selected)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

func TestBuildMultiSelectField_ItemsWithoutDesc(t *testing.T) {
	var selected []string
	items := []SelectItem{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	}
	field := buildMultiSelectField("Pick", items, &selected)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}

func TestBuildMultiSelectField_SingleItem(t *testing.T) {
	var selected []string
	items := []SelectItem{{Label: "Only", Value: "only"}}
	field := buildMultiSelectField("Solo", items, &selected)
	if field == nil {
		t.Fatal("expected non-nil field")
	}
}
