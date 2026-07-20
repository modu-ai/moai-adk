package update

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

// preview_tui.go implements the Bubble Tea v2 TUI surface for the change
// preview (AC-TUX3-008/009). It is a two-view model:
//   - previewTableView: a bubbles v2 table listing every file with its class
//     label, plus a per-class count summary header. Keyboard navigation is the
//     table component's default keymap (up/down/enter).
//   - previewDiffView: a bubbles v2 viewport showing the selected file's diff.
//     Esc returns to the table.
//
// The model consumes the SAME classification (classifyAll → Classify) as the
// text fallback (REQ-TUX3-001/002 single source of truth — plan.md §G
// anti-pattern defense).

// previewView identifies which sub-view the TUI model currently renders.
type previewView int

const (
	previewTableView previewView = iota
	previewDiffView
)

// previewModel is the Bubble Tea v2 model for the change preview TUI. It
// satisfies tea.Model for runPreviewTUI; its exported accessor methods
// (tableView / diffView / currentView / selectRow / backToTable) are the
// contract-test surface used by preview_test.go.
type previewModel struct {
	table     table.Model
	viewport  viewport.Model
	classes   []FileClassification
	diffs     map[string]string // relPath → diff content
	paths     []string          // parallel to table rows; index → relPath
	view      previewView
	width     int
	height    int
	confirmed bool
	done      bool
	header    string // per-class count summary rendered above the table
}

// newPreviewModel constructs the TUI model from the preview inputs. It
// classifies via classifyAll (single source of truth) and builds the table
// rows + viewport. The model starts in previewTableView.
func newPreviewModel(in []FilePreviewInput, isUserOwned UserOwnedPredicate, opts PreviewOptions) *previewModel {
	classes := classifyAll(in, isUserOwned)
	counts := countByClass(classes)

	// Build a header summarizing per-class counts. This surfaces every class
	// label (including `preserved (user-owned)`) above the table — AC-TUX3-008
	// per-class counts + AC-TUX3-014 label visibility.
	header := buildClassCountHeader(counts)

	paths := make([]string, 0, len(classes))
	rows := make([]table.Row, 0, len(classes))
	diffs := make(map[string]string, len(in))
	for _, f := range in {
		diffs[f.RelPath] = f.Diff
	}
	for _, c := range classes {
		paths = append(paths, c.RelPath)
		rows = append(rows, table.Row{c.Class.String(), c.RelPath})
	}

	width := opts.Width
	if width <= 0 {
		width = 80
	}
	height := opts.Height
	if height <= 0 {
		height = 24
	}

	columns := []table.Column{
		{Title: "Class", Width: 24},
		{Title: "File", Width: width - 24 - 4},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height-8),
		table.WithWidth(width),
	)

	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(height-4),
	)

	return &previewModel{
		table:     tbl,
		viewport:  vp,
		classes:   classes,
		diffs:     diffs,
		paths:     paths,
		view:      previewTableView,
		width:     width,
		height:    height,
		confirmed: false,
		done:      false,
		header:    header,
	}
}

// buildClassCountHeader renders the per-class count summary that sits above
// the table. Every class with count >= 1 appears, so the preserve class label
// is visible whenever a user-owned file is present (AC-TUX3-008/014).
func buildClassCountHeader(counts map[ChangeClass]int) string {
	var b strings.Builder
	b.WriteString("Classification summary:\n")
	for _, class := range classOrder {
		n := counts[class]
		if n == 0 {
			continue
		}
		fmt.Fprintf(&b, "  %s: %d\n", class.String(), n)
	}
	return b.String()
}

// currentView returns the active sub-view (contract test accessor).
func (m *previewModel) currentView() previewView { return m.view }

// tableView renders the table view as a string (contract test accessor). The
// rendered output includes the per-class count header + every class label +
// every file row, satisfying AC-TUX3-008 (counts + rows) and AC-TUX3-014
// (preserve label visibility).
func (m *previewModel) tableView() string {
	var b strings.Builder
	b.WriteString(m.header)
	b.WriteString("\n")
	b.WriteString(m.table.View())
	b.WriteString("\n[enter] view diff  [y] confirm  [n/q] cancel\n")
	return b.String()
}

// diffView renders the diff viewport content as a string (contract test
// accessor). AC-TUX3-009 — selecting a row reaches this view.
func (m *previewModel) diffView() string {
	return m.viewport.View()
}

// selectRow transitions from the table view to the diff view for the
// currently-selected row (contract test accessor simulating the Enter
// keypress). If the selected row has no diff content, the viewport shows a
// placeholder notice rather than an empty pane.
func (m *previewModel) selectRow() *previewModel {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.paths) {
		return m
	}
	path := m.paths[idx]
	diff := m.diffs[path]
	content := buildDiffViewContent(path, diff)
	m.viewport.SetContent(content)
	m.view = previewDiffView
	return m
}

// backToTable transitions from the diff view back to the table view (contract
// test accessor simulating the Esc keypress).
func (m *previewModel) backToTable() *previewModel {
	m.view = previewTableView
	return m
}

// buildDiffViewContent renders the content displayed in the diff viewport for
// a file. Empty diff content produces a notice line so the viewport is never
// blank.
func buildDiffViewContent(path, diff string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Diff for %s\n\n", path)
	if strings.TrimSpace(diff) == "" {
		b.WriteString("(no textual diff — this file is added or preserved without content change)\n")
	} else {
		b.WriteString(diff)
		b.WriteString("\n")
	}
	b.WriteString("\n[esc] back to table\n")
	return b.String()
}

// ---- tea.Model interface (for runPreviewTUI) ----

// Init returns the initial command (nil — no startup command needed).
func (m *previewModel) Init() tea.Cmd { return nil }

// Update handles messages. In table view: up/down navigate (handled by the
// table's own Update), enter → selectRow, 'y' → confirm+quit, 'n'/'q'/ctrl+c
// → cancel+quit. In diff view: esc → backToTable, viewport paging handled by
// the viewport's own Update.
func (m *previewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Keystroke() {
		case "ctrl+c", "q":
			m.done = true
			m.confirmed = false
			return m, tea.Quit
		case "y":
			m.done = true
			m.confirmed = true
			return m, tea.Quit
		}
		if m.view == previewDiffView {
			switch msg.Keystroke() {
			case "esc":
				m.backToTable()
				return m, nil
			}
			vp, cmd := m.viewport.Update(msg)
			m.viewport = vp
			return m, cmd
		}
		// table view
		switch msg.Keystroke() {
		case "enter":
			return m.selectRow(), nil
		case "n":
			m.done = true
			m.confirmed = false
			return m, tea.Quit
		}
		tbl, cmd := m.table.Update(msg)
		m.table = tbl
		return m, cmd
	}
	return m, nil
}

// View renders the current sub-view (satisfies tea.Model).
func (m *previewModel) View() tea.View {
	if m.view == previewDiffView {
		return tea.NewView(m.diffView())
	}
	return tea.NewView(m.tableView())
}

// runPreviewTUI launches the Bubble Tea v2 program and blocks until the user
// confirms or cancels. Returns the confirmed flag.
func runPreviewTUI(in []FilePreviewInput, isUserOwned UserOwnedPredicate, opts PreviewOptions) (bool, error) {
	model := newPreviewModel(in, isUserOwned, opts)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return false, err
	}
	return model.confirmed, nil
}
