package update

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// preview_tui.go implements the Bubble Tea v2 TUI surface for the change
// preview (AC-TUX3-008/009). It is a two-view model:
//   - previewTableView: a bubbles v2 table listing every file with its class
//     label, plus a per-class count summary card. Keyboard navigation is the
//     table component's default keymap (up/down/enter).
//   - previewDiffView: a bubbles v2 viewport showing the selected file's diff.
//     Esc returns to the table.
//
// The model consumes the SAME classification (classifyAll → Classify) as the
// text fallback (REQ-TUX3-001/002 single source of truth — plan.md §G
// anti-pattern defense).
//
// # Presentation
//
// Every colour decision resolves from the internal/tui design tokens
// (REQ-TUIM-010/011/014). The table is built with an explicit table.Styles
// derived from the resolved Theme instead of the bubbles default palette, the
// per-class count summary is rendered through the tui.Box structural primitive,
// and key hints render through tui.HelpBar. No hex literal appears in this file
// (constraint D1).
//
// # Residue policy — HYBRID inline / alternate screen (REQ-TUIM-019/019a/022/022a)
//
// The two sub-views render on different screens, and the mechanism is the
// per-view AltScreen field on the value returned by View() — bubbletea v2 has no
// tea.WithAltScreen program option, so the toggle is naturally per-sub-view:
//
//   - previewTableView renders INLINE (AltScreen: false) so the identity band
//     and classification card printed above the program stay visible while the
//     user decides. Making the whole program alt-screen would hide the very
//     context the confirmation depends on.
//   - previewDiffView renders on the ALTERNATE screen (AltScreen: true) so a
//     full-height diff gets the whole viewport and leaves no scrollback residue.
//   - Leaving the diff (esc) restores the inline table. Prior scrollback is
//     preserved by DEC private mode 1049 itself (the renderer emits the
//     save-cursor / switch-buffer pair on transition), so backToTable stays a
//     pure state transition and performs no manual redraw.
//   - On exit the interactive frame is replaced by a single-line result summary
//     (resultLine). The renderer's inline close path writes EraseScreenBelow
//     after the final frame, so a one-line final view leaves exactly one line in
//     scrollback — that is why the summary is the final View() rather than a
//     post-program print.
//
// CONSTRAINT — a quit MUST resolve to the inline sub-view first. The renderer's
// close path branches on the LAST view and discards the frame when that view is
// an alt-screen view. Because ctrl+c / q / y are handled before the sub-view
// check, they are reachable FROM the diff sub-view; returning tea.Quit while
// still in previewDiffView would silently discard the exit summary. Every quit
// path therefore sets view = previewTableView before returning tea.Quit
// (REQ-TUIM-022a). This compiles cleanly and no content-level test catches it.
//
// `n` is deliberately NOT diff-reachable: the diff branch returns before the
// table-only key cases, so `n` falls through to the viewport. That asymmetry
// matches the documented keymap and is preserved.

// previewView identifies which sub-view the TUI model currently renders.
type previewView int

const (
	previewTableView previewView = iota
	previewDiffView
)

// previewResolveTheme is the preview's theme-resolution entry point. It defaults
// to tui.ResolveOS so the preview INHERITS the canonical precedence chain
// (NO_COLOR > MOAI_THEME light/dark > DetectDark > dark-default; internal/tui
// detect.go Resolve) rather than re-implementing or re-ordering any part of it
// (REQ-TUIM-015).
//
// It is a package-level var so tests can force a specific axis without mutating
// the process environment, mirroring cli.huhThemeIsDark and wizard.wizardIsDark.
// Forcing it BYPASSES the precedence chain, so a test that verifies the chain
// itself must leave this var at its default and drive the environment instead.
var previewResolveTheme = tui.ResolveOS

// previewModel is the Bubble Tea v2 model for the change preview TUI. It
// satisfies tea.Model for runPreviewTUI; its accessor methods (tableView /
// diffView / currentView / selectRow / backToTable) are the contract-test
// surface used by preview_test.go.
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
	header    string    // per-class count summary card rendered above the table
	theme     tui.Theme // resolved design tokens for every colour decision
}

// newPreviewModel constructs the TUI model from the preview inputs. It
// classifies via classifyAll (single source of truth) and builds the table
// rows + viewport. The model starts in previewTableView.
func newPreviewModel(in []FilePreviewInput, isUserOwned UserOwnedPredicate, opts PreviewOptions) *previewModel {
	th := previewResolveTheme()

	classes := classifyAll(in, isUserOwned)
	counts := countByClass(classes)

	// Build a card summarizing per-class counts. This surfaces every class
	// label (including `preserved (user-owned)`) above the table — AC-TUX3-008
	// per-class counts + AC-TUX3-014 label visibility.
	header := buildClassCountHeader(counts, th)

	paths := make([]string, 0, len(classes))
	rows := make([]table.Row, 0, len(classes))
	diffs := make(map[string]string, len(in))
	for _, f := range in {
		diffs[f.RelPath] = f.Diff
	}
	for _, c := range classes {
		paths = append(paths, c.RelPath)
		// The class label carries its semantic colour role (REQ-TUIM-014); the
		// label STRING stays byte-identical to ChangeClass.String().
		rows = append(rows, table.Row{paint(classRoleToken(th, c.Class), c.Class.String()), c.RelPath})
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

	// WithStyles precedes WithHeight deliberately: table options are applied in
	// order, and WithHeight subtracts the rendered header height, which the
	// header's bottom border changes.
	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithStyles(previewTableStyles(th)),
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
		theme:     th,
	}
}

// isMonochrome reports whether the resolved theme carries no colour at all —
// the tui.MonochromeTheme returned under NO_COLOR. Every styling decision in
// this file is skipped on that path so the rendered output stays structurally
// free of ANSI sequences (REQ-TUIM-016).
func isMonochrome(th tui.Theme) bool { return th == tui.MonochromeTheme() }

// paint applies a foreground colour to s. An empty token (the MonochromeTheme
// case) returns s unchanged, so no ANSI sequence is emitted under NO_COLOR.
//
// lipgloss v2 renders the colour as a decimal RGB SGR parameter run; the hex
// token itself never reaches the output.
func paint(token, s string) string {
	if token == "" {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(token)).Render(s)
}

// classRoleToken maps each ChangeClass to its semantic colour role in the
// resolved theme (REQ-TUIM-014): add is a positive outcome, update is
// informational, a preserved user-owned file is de-emphasized, and a conflict is
// a negative outcome. The four tokens are distinct in both palettes.
func classRoleToken(th tui.Theme, c ChangeClass) string {
	switch c {
	case ClassAdd:
		return th.Success
	case ClassUpdate:
		return th.Info
	case ClassPreserveUserOwned:
		return th.Dim
	case ClassConflict:
		return th.Danger
	}
	return th.Fg
}

// previewTableStyles derives the bubbles table styles from the resolved theme
// (REQ-TUIM-011). It replaces table.DefaultStyles(), whose palette belongs to
// the component library rather than to MoAI — the mechanical root of the dated
// look this SPEC repairs.
//
// The selected row keeps the bubbles baseline's Bold so the cursor stays legible
// on a low-contrast terminal. Bold combined with a foreground colour renders as
// a single merged CSI, which is why token-application tests assert the SGR
// parameter run rather than a whole escape prefix.
func previewTableStyles(th tui.Theme) table.Styles {
	if isMonochrome(th) {
		// No colour and no attribute: NO_COLOR output must carry zero ANSI. The
		// header's bottom RULE is kept — box-drawing runes are structure, not
		// colour — so the monochrome structure golden still pins the table's
		// border geometry.
		return table.Styles{
			Header: lipgloss.NewStyle().
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true),
			Cell:     lipgloss.NewStyle().Padding(0, 1),
			Selected: lipgloss.NewStyle(),
		}
	}
	return table.Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Foreground(lipgloss.Color(th.Accent)).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(th.ChromeBorder)).
			BorderBottom(true),
		Cell: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color(th.Fg)),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(th.Accent)),
	}
}

// buildClassCountHeader renders the per-class count summary card that sits above
// the table. Every class with count >= 1 appears, so the preserve class label is
// visible whenever a user-owned file is present (AC-TUX3-008/014).
//
// The card is a tui.Box structural primitive (REQ-TUIM-012) rather than the
// former bare `Classification summary:` text block; each class label carries its
// semantic colour role and the counts align in a fixed column.
func buildClassCountHeader(counts map[ChangeClass]int, th tui.Theme) string {
	labelWidth := 0
	for _, class := range classOrder {
		if counts[class] == 0 {
			continue
		}
		labelWidth = max(labelWidth, len(class.String()))
	}

	var b strings.Builder
	for _, class := range classOrder {
		n := counts[class]
		if n == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		label := class.String()
		// Pad from the PLAIN label width: a colour sequence has zero display
		// width, so measuring the painted string would over-pad.
		b.WriteString(paint(classRoleToken(th, class), label))
		b.WriteString(strings.Repeat(" ", labelWidth-len(label)+4))
		fmt.Fprintf(&b, "%d", n)
	}

	return tui.Box(tui.BoxOpts{
		Title:  "Classification summary",
		Body:   b.String(),
		Theme:  &th,
		Accent: true,
	})
}

// previewTableHints is the key-hint row for the table sub-view (REQ-TUIM-013).
func previewTableHints(th tui.Theme) string {
	return tui.HelpBar([]tui.KeyHint{
		{Key: "enter", Label: "view diff"},
		{Key: "y", Label: "confirm"},
		{Key: "n/q", Label: "cancel"},
	}, &th)
}

// previewDiffHints is the key-hint row for the diff sub-view (REQ-TUIM-013).
func previewDiffHints(th tui.Theme) string {
	return tui.HelpBar([]tui.KeyHint{
		{Key: "esc", Label: "back to table"},
	}, &th)
}

// currentView returns the active sub-view (contract test accessor).
func (m *previewModel) currentView() previewView { return m.view }

// tableView renders the table view as a string (contract test accessor). The
// rendered output includes the per-class count card + every class label + every
// file row, satisfying AC-TUX3-008 (counts + rows) and AC-TUX3-014 (preserve
// label visibility).
func (m *previewModel) tableView() string {
	var b strings.Builder
	b.WriteString(m.header)
	b.WriteString("\n\n")
	b.WriteString(m.table.View())
	b.WriteString("\n")
	b.WriteString(previewTableHints(m.theme))
	b.WriteString("\n")
	return b.String()
}

// diffView renders the diff viewport content as a string (contract test
// accessor). AC-TUX3-009 — selecting a row reaches this view.
func (m *previewModel) diffView() string {
	return m.viewport.View()
}

// resultLine is the single-line exit summary that replaces the interactive frame
// (REQ-TUIM-022). It names the outcome and the file count and carries NEITHER
// the class-count card NOR any file row — re-rendering those is the
// duplicate-render symptom the hybrid residue policy removes.
//
// The glyph resolves from the canonical tui.Glyph* constants (REQ-TUIM-017); no
// raw glyph rune literal appears in this file.
func (m *previewModel) resultLine() string {
	glyph := string(tui.GlyphSkip)
	token := m.theme.Dim
	outcome := "update cancelled"
	if m.confirmed {
		glyph = string(tui.GlyphDone)
		token = m.theme.Success
		outcome = "update confirmed"
	}
	return paint(token, fmt.Sprintf("%s %s %s %d files", glyph, outcome, string(tui.GlyphInfo), len(m.classes)))
}

// selectRow transitions from the table view to the diff view for the
// currently-selected row (contract test accessor simulating the Enter keypress).
// If the selected row has no diff content, the viewport shows a placeholder
// notice rather than an empty pane.
func (m *previewModel) selectRow() *previewModel {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.paths) {
		return m
	}
	path := m.paths[idx]
	diff := m.diffs[path]
	content := buildDiffViewContent(path, diff, m.theme)
	m.viewport.SetContent(content)
	m.view = previewDiffView
	return m
}

// backToTable transitions from the diff view back to the table view (contract
// test accessor simulating the Esc keypress). It is a pure state transition: the
// terminal's DEC mode 1049 restores the main screen buffer, so no manual redraw
// is required (REQ-TUIM-019a).
func (m *previewModel) backToTable() *previewModel {
	m.view = previewTableView
	return m
}

// buildDiffViewContent renders the content displayed in the diff viewport for a
// file. Empty diff content produces a notice line so the viewport is never
// blank.
//
// The diff body itself is left unpainted: it is user data whose own +/- markers
// carry the meaning, and wrapping it in a single foreground token inside a
// scrolling viewport buys nothing while risking the content assertions that pin
// the diff text. Only the header, the placeholder notice, and the key hints are
// theme-painted.
func buildDiffViewContent(path, diff string, th tui.Theme) string {
	var b strings.Builder
	b.WriteString(paint(th.Accent, "Diff for "+path))
	b.WriteString("\n\n")
	if strings.TrimSpace(diff) == "" {
		b.WriteString(paint(th.Dim, "(no textual diff — this file is added or preserved without content change)"))
		b.WriteString("\n")
	} else {
		b.WriteString(diff)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(previewDiffHints(th))
	b.WriteString("\n")
	return b.String()
}

// ---- tea.Model interface (for runPreviewTUI) ----

// Init returns the initial command (nil — no startup command needed).
func (m *previewModel) Init() tea.Cmd { return nil }

// quit marks the model done with the given outcome and resolves the sub-view
// back to the inline table BEFORE the program tears down.
//
// The view reset is load-bearing, not cosmetic: the renderer's close path
// discards the final frame when the last view is an alt-screen view, so a quit
// entered from the diff sub-view would drop the exit summary entirely
// (REQ-TUIM-022a).
func (m *previewModel) quit(confirmed bool) (tea.Model, tea.Cmd) {
	m.done = true
	m.confirmed = confirmed
	m.view = previewTableView
	return m, tea.Quit
}

// Update handles messages. In table view: up/down navigate (handled by the
// table's own Update), enter → selectRow, 'y' → confirm+quit, 'n'/'q'/ctrl+c →
// cancel+quit. In diff view: esc → backToTable, viewport paging handled by the
// viewport's own Update.
func (m *previewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Keystroke() {
		case "ctrl+c", "q":
			return m.quit(false)
		case "y":
			return m.quit(true)
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
			return m.quit(false)
		}
		tbl, cmd := m.table.Update(msg)
		m.table = tbl
		return m, cmd
	}
	return m, nil
}

// View renders the current sub-view (satisfies tea.Model). See the residue
// policy documented at the top of this file for why AltScreen is a per-sub-view
// decision and why the exit summary is the final view rather than a teardown
// print.
func (m *previewModel) View() tea.View {
	if m.done {
		v := tea.NewView(m.resultLine())
		v.AltScreen = false
		return v
	}
	if m.view == previewDiffView {
		v := tea.NewView(m.diffView())
		v.AltScreen = true
		return v
	}
	v := tea.NewView(m.tableView())
	v.AltScreen = false
	return v
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
