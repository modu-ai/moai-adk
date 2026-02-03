package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// checkboxImpl implements the Checkbox interface.
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

// multiSelectInteractive runs a bubbletea program for interactive multi-selection.
func (c *checkboxImpl) multiSelectInteractive(label string, items []SelectItem) ([]string, error) {
	m := newCheckboxModel(c.theme, label, items)

	finalModel, err := programRunner(m)
	if err != nil {
		return nil, fmt.Errorf("checkbox: %w", err)
	}

	result := finalModel.(checkboxModel)
	if result.cancelled {
		return nil, ErrCancelled
	}

	return result.selectedValues(), nil
}

// checkboxModel is the bubbletea Model for multi-selection.
type checkboxModel struct {
	theme     *Theme
	label     string
	items     []SelectItem
	selected  map[int]bool
	cursor    int
	cancelled bool
	done      bool
}

// newCheckboxModel creates a checkboxModel with the initial state.
func newCheckboxModel(theme *Theme, label string, items []SelectItem) checkboxModel {
	copied := make([]SelectItem, len(items))
	copy(copied, items)
	return checkboxModel{
		theme:    theme,
		label:    label,
		items:    copied,
		selected: make(map[int]bool),
	}
}

// Init is the bubbletea initialization command.
func (m checkboxModel) Init() tea.Cmd {
	return nil
}

// Update processes messages and returns the updated model.
func (m checkboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyUp:
			m = m.moveCursorUp()
		case tea.KeyDown:
			m = m.moveCursorDown()
		case tea.KeySpace:
			m = m.toggleCurrent()
		case tea.KeyRunes:
			if len(msg.Runes) == 1 && msg.Runes[0] == 'a' {
				m = m.toggleAll()
			}
		}
	}
	return m, nil
}

// View renders the checkbox list to a string.
func (m checkboxModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.theme.RenderTitle(m.label))
	b.WriteString("\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}

		line := cursor + check + " " + item.Label
		if i == m.cursor {
			b.WriteString(m.theme.RenderHighlight(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString(m.theme.RenderMuted("Space: toggle, a: toggle all, Enter: confirm"))
	return b.String()
}

// moveCursorDown advances the cursor, wrapping at the end.
func (m checkboxModel) moveCursorDown() checkboxModel {
	if len(m.items) == 0 {
		return m
	}
	m.cursor = (m.cursor + 1) % len(m.items)
	return m
}

// moveCursorUp moves the cursor back, wrapping at the beginning.
func (m checkboxModel) moveCursorUp() checkboxModel {
	if len(m.items) == 0 {
		return m
	}
	m.cursor--
	if m.cursor < 0 {
		m.cursor = len(m.items) - 1
	}
	return m
}

// toggleCurrent toggles the selection state of the item at the cursor.
func (m checkboxModel) toggleCurrent() checkboxModel {
	// Copy the map to maintain immutability for bubbletea
	newSelected := make(map[int]bool, len(m.selected))
	for k, v := range m.selected {
		newSelected[k] = v
	}
	newSelected[m.cursor] = !m.selected[m.cursor]
	m.selected = newSelected
	return m
}

// toggleAll selects all items if any are unselected, or deselects all if all are selected.
func (m checkboxModel) toggleAll() checkboxModel {
	allSelected := true
	for i := range m.items {
		if !m.selected[i] {
			allSelected = false
			break
		}
	}

	newSelected := make(map[int]bool, len(m.items))
	if allSelected {
		// Deselect all
		for i := range m.items {
			newSelected[i] = false
		}
	} else {
		// Select all
		for i := range m.items {
			newSelected[i] = true
		}
	}
	m.selected = newSelected
	return m
}

// selectedValues returns the values of all selected items in order.
func (m checkboxModel) selectedValues() []string {
	var result []string
	for i, item := range m.items {
		if m.selected[i] {
			result = append(result, item.Value)
		}
	}
	if result == nil {
		return []string{}
	}
	return result
}
