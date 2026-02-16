package ui

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
)

// selectorImpl implements the Selector interface using huh.Select.
type selectorImpl struct {
	theme    *Theme
	headless *HeadlessManager
}

// NewSelector creates a Selector backed by the given theme and headless manager.
func NewSelector(theme *Theme, hm *HeadlessManager) Selector {
	return &selectorImpl{theme: theme, headless: hm}
}

// Select displays items for the user to choose from and returns the selected value.
// In headless mode it returns the default or the first item immediately.
func (s *selectorImpl) Select(label string, items []SelectItem) (string, error) {
	if len(items) == 0 {
		return "", ErrNoItems
	}

	if s.headless.IsHeadless() {
		return s.selectHeadless(label, items)
	}

	return s.selectInteractive(label, items)
}

// selectHeadless returns the default value or the first item when no default is set.
func (s *selectorImpl) selectHeadless(label string, items []SelectItem) (string, error) {
	if val, ok := s.headless.GetDefault(label); ok {
		return val, nil
	}
	return items[0].Value, nil
}

// buildSelectField creates a configured huh.Select field for single selection.
// It converts SelectItems to huh.Options and binds the selected value pointer.
// This function is separated from selectInteractive to enable unit testing
// of field configuration without requiring a TTY.
func buildSelectField(label string, items []SelectItem, selected *string) *huh.Select[string] {
	opts := make([]huh.Option[string], len(items))
	for i, item := range items {
		key := item.Label
		if item.Desc != "" {
			key = item.Label + " - " + item.Desc
		}
		opts[i] = huh.NewOption(key, item.Value)
	}

	return huh.NewSelect[string]().
		Title(label).
		Options(opts...).
		Value(selected)
}


// selectInteractive runs a huh.Select form for interactive selection.
func (s *selectorImpl) selectInteractive(label string, items []SelectItem) (string, error) {
	var selected string
	sel := buildSelectField(label, items, &selected)

	form := huh.NewForm(huh.NewGroup(sel)).
		WithTheme(NewMoAIHuhTheme(s.theme.NoColor)).
		WithAccessible(s.headless.IsHeadless())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", ErrCancelled
		}
		return "", fmt.Errorf("selector: %w", err)
	}

	return selected, nil
}
