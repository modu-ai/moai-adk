package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// selectorImpl implements the Selector interface.
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

// selectInteractive runs a bubbletea program for interactive selection.
func (s *selectorImpl) selectInteractive(label string, items []SelectItem) (string, error) {
	m := newSelectorModel(s.theme, label, items)

	finalModel, err := programRunner(m)
	if err != nil {
		return "", fmt.Errorf("selector: %w", err)
	}

	result := finalModel.(selectorModel)
	if result.cancelled {
		return "", ErrCancelled
	}

	return result.selectedValue(), nil
}

// selectorModel is the bubbletea Model for the single-selection component.
type selectorModel struct {
	theme     *Theme
	label     string
	items     []SelectItem
	filtered  []SelectItem
	cursor    int
	filter    string
	cancelled bool
	done      bool
}

// newSelectorModel creates a selectorModel with the initial state.
func newSelectorModel(theme *Theme, label string, items []SelectItem) selectorModel {
	copied := make([]SelectItem, len(items))
	copy(copied, items)
	return selectorModel{
		theme:    theme,
		label:    label,
		items:    copied,
		filtered: copied,
	}
}

// Init is the bubbletea initialization command.
func (m selectorModel) Init() tea.Cmd {
	return nil
}

// Update processes messages and returns the updated model.
func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case tea.KeyBackspace:
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m = m.applyFilter(m.filter)
			}
		case tea.KeyRunes:
			m.filter += string(msg.Runes)
			m = m.applyFilter(m.filter)
		}
	}
	return m, nil
}

// View renders the selector to a string.
func (m selectorModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

	var b strings.Builder

	// Label
	b.WriteString(m.theme.RenderTitle(m.label))
	b.WriteString("\n")

	// Filter prompt
	if m.filter != "" {
		b.WriteString(m.theme.RenderMuted("Filter: "))
		b.WriteString(m.filter)
		b.WriteString("\n")
	}

	// Items
	for i, item := range m.filtered {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		if i == m.cursor {
			b.WriteString(m.theme.RenderHighlight(cursor + item.Label))
			if item.Desc != "" {
				b.WriteString(m.theme.RenderMuted(" - " + item.Desc))
			}
		} else {
			b.WriteString(cursor + item.Label)
			if item.Desc != "" {
				b.WriteString(m.theme.RenderMuted(" - " + item.Desc))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// moveCursorDown advances the cursor, wrapping at the end.
func (m selectorModel) moveCursorDown() selectorModel {
	if len(m.filtered) == 0 {
		return m
	}
	m.cursor = (m.cursor + 1) % len(m.filtered)
	return m
}

// moveCursorUp moves the cursor back, wrapping at the beginning.
func (m selectorModel) moveCursorUp() selectorModel {
	if len(m.filtered) == 0 {
		return m
	}
	m.cursor--
	if m.cursor < 0 {
		m.cursor = len(m.filtered) - 1
	}
	return m
}

// applyFilter performs case-insensitive substring matching on items.
func (m selectorModel) applyFilter(query string) selectorModel {
	if query == "" {
		m.filtered = m.items
		m.cursor = 0
		return m
	}

	q := strings.ToLower(query)
	var result []SelectItem
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.Label), q) ||
			strings.Contains(strings.ToLower(item.Value), q) {
			result = append(result, item)
		}
	}
	m.filtered = result
	m.cursor = 0
	return m
}

// selectedValue returns the value of the item at the current cursor position.
func (m selectorModel) selectedValue() string {
	if len(m.filtered) == 0 {
		return ""
	}
	if m.cursor >= len(m.filtered) {
		return m.filtered[0].Value
	}
	return m.filtered[m.cursor].Value
}
