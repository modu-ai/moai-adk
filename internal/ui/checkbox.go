package ui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// checkboxImpl implements the Checkbox interface using huh.MultiSelect.
type checkboxImpl struct {
	theme    *Theme
	headless *HeadlessManager
}

// NewCheckbox creates a Checkbox backed by the given theme and headless manager.
func NewCheckbox(theme *Theme, hm *HeadlessManager) Checkbox {
	return &checkboxImpl{theme: theme, headless: hm}
}

// MultiSelect displays items for the user to toggle and returns selected values.
// In headless mode it returns comma-separated defaults or an empty slice.
func (c *checkboxImpl) MultiSelect(label string, items []SelectItem) ([]string, error) {
	if len(items) == 0 {
		return nil, ErrNoItems
	}

	if c.headless.IsHeadless() {
		return c.multiSelectHeadless(label)
	}

	return c.multiSelectInteractive(label, items)
}

// multiSelectHeadless returns comma-separated default values or an empty slice.
func (c *checkboxImpl) multiSelectHeadless(label string) ([]string, error) {
	if val, ok := c.headless.GetDefault(label); ok && val != "" {
		parts := strings.Split(val, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result, nil
	}
	return []string{}, nil
}

// buildMultiSelectField creates a configured huh.MultiSelect field for multi-selection.
// It converts SelectItems to huh.Options and binds the selected values pointer.
// This function is separated from multiSelectInteractive to enable unit testing
// of field configuration without requiring a TTY.
func buildMultiSelectField(label string, items []SelectItem, selected *[]string) *huh.MultiSelect[string] {
	opts := make([]huh.Option[string], len(items))
	for i, item := range items {
		key := item.Label
		if item.Desc != "" {
			key = item.Label + " - " + item.Desc
		}
		opts[i] = huh.NewOption(key, item.Value)
	}

	return huh.NewMultiSelect[string]().
		Title(label).
		Options(opts...).
		Value(selected)
}

// multiSelectInteractive runs a huh.MultiSelect form for interactive multi-selection.
func (c *checkboxImpl) multiSelectInteractive(label string, items []SelectItem) ([]string, error) {
	var selected []string
	ms := buildMultiSelectField(label, items, &selected)

	form := huh.NewForm(huh.NewGroup(ms)).
		WithTheme(NewMoAIHuhTheme(c.theme.NoColor)).
		WithAccessible(c.headless.IsHeadless())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrCancelled
		}
		return nil, fmt.Errorf("checkbox: %w", err)
	}

	if selected == nil {
		return []string{}, nil
	}
	return selected, nil
}
