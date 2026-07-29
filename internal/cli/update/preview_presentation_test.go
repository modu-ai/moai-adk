package update

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// preview_presentation_test.go is MECHANISM 2 of the split presentation-regression
// coverage (SPEC-CLI-TUI-MODERNIZE-001 D-5): token-application unit tests that
// assert each semantic role resolves to the correct internal/tui theme token.
//
// The whole regression-detection burden for the defect class this SPEC repairs
// sits HERE, not on the structure goldens in preview_golden_test.go. Those
// goldens are palette-insensitive by design — removing the theme wiring would
// leave them unchanged and green. Only these tests trip.
//
// # Assertion method — SGR parameter substring
//
// lipgloss v2 converts a hex token to a decimal RGB SGR parameter run before
// emission, so two intuitive assertions are FALSE for a correct implementation:
//
//  1. strings.Contains(rendered, th.Accent) — the hex literal never appears.
//  2. the full-CSI prefix "\x1b[38;2;191;101;71m" — v2 merges every SGR
//     parameter into ONE CSI, so a bold+coloured cell renders
//     "\x1b[1;38;2;191;101;71m" and the foreground-only prefix is not a
//     substring. This is not hypothetical: the selected table row is styled
//     Bold(true).Foreground(...).
//
// Both are ruled out by asserting the PARAMETER RUN alone (sgrParams below).
// Reading the token value from the resolved Theme rather than hard-coding a hex
// string also keeps these tests inside the no-hex-literals-outside-internal/tui
// constraint.

// sgrParams returns the SGR parameter run lipgloss v2 emits for a theme token —
// e.g. "38;2;191;101;71" for the light Accent. It is the substring that survives
// being merged into a multi-attribute CSI.
func sgrParams(t *testing.T, token string) string {
	t.Helper()
	if token == "" {
		t.Fatalf("sgrParams called with an empty token; the theme is monochrome")
	}
	probe := lipgloss.NewStyle().Foreground(lipgloss.Color(token)).Render("x")
	params := strings.TrimSuffix(strings.TrimPrefix(probe, "\x1b["), "x\x1b[m")
	params = strings.TrimSuffix(params, "m")
	if params == "" || params == probe {
		t.Fatalf("could not extract an SGR parameter run from probe %q (token %q)", probe, token)
	}
	return params
}

// withPreviewTheme forces the preview's theme resolution to a fixed palette for
// the duration of the test, bypassing the environment-driven precedence chain.
// Tests using it MUST NOT be parallel: previewResolveTheme is package state.
func withPreviewTheme(t *testing.T, theme tui.Theme) {
	t.Helper()
	prev := previewResolveTheme
	previewResolveTheme = func() tui.Theme { return theme }
	t.Cleanup(func() { previewResolveTheme = prev })
}

// newFixtureModel builds a preview model over the four-class fixture.
func newFixtureModel() *previewModel {
	return newPreviewModel(
		allFourClassesInputs(),
		allFourClassesPredicate(),
		PreviewOptions{Interactive: true, Width: 80, Height: 24},
	)
}

// key builds a synthetic key-press message for driving Update headlessly.
func key(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text}
}

// -------------------------------------------------------------------
// AC-TUIM-004 — the Accent token reaches tableView() in rendered form
// -------------------------------------------------------------------

func TestPreviewTableAppliesAccentToken(t *testing.T) {
	th := tui.LightTheme()
	withPreviewTheme(t, th)

	rendered := newFixtureModel().tableView()
	want := sgrParams(t, th.Accent)

	if !strings.Contains(rendered, want) {
		t.Errorf("tableView output does not carry the Accent SGR parameter run %q "+
			"(AC-TUIM-004: the resolved theme must reach the rendered table); rendered:\n%q", want, rendered)
	}
}

// -------------------------------------------------------------------
// AC-TUIM-008 — each class label carries a DISTINCT semantic colour role
// -------------------------------------------------------------------

func TestPreviewClassLabelsCarryDistinctSemanticRoles(t *testing.T) {
	th := tui.LightTheme()
	withPreviewTheme(t, th)

	rendered := newFixtureModel().tableView()

	seen := make(map[string]ChangeClass, 4)
	for _, class := range classOrder {
		params := sgrParams(t, classRoleToken(th, class))
		if !strings.Contains(rendered, params) {
			t.Errorf("class %q does not carry its semantic role SGR run %q in tableView output (AC-TUIM-008)",
				class.String(), params)
			continue
		}
		if other, dup := seen[params]; dup {
			t.Errorf("class %q shares SGR run %q with class %q — the four roles must be distinct (AC-TUIM-008)",
				class.String(), params, other.String())
		}
		seen[params] = class
	}
}

// -------------------------------------------------------------------
// AC-TUIM-030b — mechanism 2: semantic role → internal/tui token map
// -------------------------------------------------------------------

// TestPreviewSemanticRolesResolveToThemeTokens is the load-bearing
// regression test for this SPEC. Removing the theme wiring from preview_tui.go
// (the AC-TUIM-030c self-trip) must make this test FAIL.
func TestPreviewSemanticRolesResolveToThemeTokens(t *testing.T) {
	th := tui.LightTheme()
	withPreviewTheme(t, th)

	rendered := newFixtureModel().tableView()

	roles := []struct {
		name  string
		token string
	}{
		{"conflict label → Danger", th.Danger},
		{"table border → ChromeBorder", th.ChromeBorder},
		{"selected row → Accent", th.Accent},
		{"add label → Success", th.Success},
		{"update label → Info", th.Info},
		{"preserve label → Dim", th.Dim},
	}

	for _, r := range roles {
		params := sgrParams(t, r.token)
		if !strings.Contains(rendered, params) {
			t.Errorf("%s: SGR parameter run %q absent from tableView output (AC-TUIM-030b); rendered:\n%q",
				r.name, params, rendered)
		}
	}
}

// -------------------------------------------------------------------
// AC-TUIM-003 — the light/dark axis is consumed, not resolved and discarded
// -------------------------------------------------------------------

func TestPreviewTableLightAndDarkAxesDiffer(t *testing.T) {
	withPreviewTheme(t, tui.LightTheme())
	light := newFixtureModel().tableView()

	withPreviewTheme(t, tui.DarkTheme())
	dark := newFixtureModel().tableView()

	if light == dark {
		t.Errorf("tableView output is identical under the light and dark axes — the resolved theme is not consumed (AC-TUIM-003)")
	}
}

// -------------------------------------------------------------------
// AC-TUIM-005 — the count summary uses an internal/tui structural primitive
// -------------------------------------------------------------------

func TestPreviewSummaryRendersThroughStructuralPrimitive(t *testing.T) {
	withPreviewTheme(t, tui.LightTheme())

	rendered := newFixtureModel().tableView()

	// tui.Box draws a rounded lipgloss border; its corner rune is the
	// structural signal that the bare text block was replaced by a card.
	if !strings.ContainsAny(rendered, "╭╮╰╯─│") {
		t.Errorf("tableView output carries no box-border rune — the per-class summary is not rendered "+
			"through a tui structural primitive (AC-TUIM-005); rendered:\n%q", rendered)
	}
	if !strings.Contains(rendered, "Classification summary") {
		t.Errorf("tableView output is missing the summary card title; rendered:\n%q", rendered)
	}
}

// -------------------------------------------------------------------
// AC-TUIM-009 — zero ANSI under the monochrome axis
// -------------------------------------------------------------------

func TestPreviewViewsEmitZeroANSIUnderMonochrome(t *testing.T) {
	withPreviewTheme(t, tui.MonochromeTheme())

	m := newFixtureModel()
	if n := strings.Count(m.tableView(), "\x1b["); n != 0 {
		t.Errorf("tableView emitted %d ANSI escape sequences under the monochrome axis; want 0 (AC-TUIM-009)", n)
	}
	m = m.selectRow()
	if n := strings.Count(m.diffView(), "\x1b["); n != 0 {
		t.Errorf("diffView emitted %d ANSI escape sequences under the monochrome axis; want 0 (AC-TUIM-009)", n)
	}
	if n := strings.Count(m.resultLine(), "\x1b["); n != 0 {
		t.Errorf("resultLine emitted %d ANSI escape sequences under the monochrome axis; want 0 (AC-TUIM-009)", n)
	}
}

// -------------------------------------------------------------------
// AC-TUIM-039 — the canonical axis-precedence chain is inherited, not re-implemented
// -------------------------------------------------------------------

// TestPreviewInheritsCanonicalAxisPrecedence drives the REAL resolution path
// (previewResolveTheme left at its tui.ResolveOS default) through the process
// environment, so it verifies the chain itself. Forcing the indirection var
// would bypass exactly what is under test. Non-parallel per t.Setenv.
func TestPreviewInheritsCanonicalAxisPrecedence(t *testing.T) {
	cases := []struct {
		name     string
		noColor  string
		setColor bool
		moaiTh   string
		setTheme bool
		want     tui.Theme
	}{
		{name: "NO_COLOR beats MOAI_THEME=dark", noColor: "1", setColor: true, moaiTh: "dark", setTheme: true, want: tui.MonochromeTheme()},
		{name: "MOAI_THEME=light", moaiTh: "light", setTheme: true, want: tui.LightTheme()},
		{name: "MOAI_THEME=dark", moaiTh: "dark", setTheme: true, want: tui.DarkTheme()},
		// Non-TTY detection returns the safe dark default for both auto and unset.
		{name: "MOAI_THEME=auto defers to detection", moaiTh: "auto", setTheme: true, want: tui.DarkTheme()},
		{name: "MOAI_THEME unset defers to detection", want: tui.DarkTheme()},
		// The default: branch short-circuits to dark WITHOUT querying the terminal.
		{name: "invalid MOAI_THEME short-circuits to dark", moaiTh: "purple", setTheme: true, want: tui.DarkTheme()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", "")
			t.Setenv("MOAI_THEME", "")
			if tc.setColor {
				t.Setenv("NO_COLOR", tc.noColor)
			}
			if tc.setTheme {
				t.Setenv("MOAI_THEME", tc.moaiTh)
			}

			got := newFixtureModel().theme
			if got != tc.want {
				t.Errorf("preview resolved the wrong theme through the precedence chain (AC-TUIM-039)\n got Accent=%q Fg=%q\nwant Accent=%q Fg=%q",
					got.Accent, got.Fg, tc.want.Accent, tc.want.Fg)
			}
		})
	}
}

// -------------------------------------------------------------------
// AC-TUIM-014a/b/c — hybrid inline / alternate-screen residue policy
// -------------------------------------------------------------------

func TestPreviewTableSubViewRendersInline(t *testing.T) {
	withPreviewTheme(t, tui.LightTheme())

	m := newFixtureModel()
	if m.currentView() != previewTableView {
		t.Fatalf("model did not start in the table sub-view")
	}
	if v := m.View(); v.AltScreen {
		t.Errorf("table sub-view returned AltScreen=true; the table must render inline (AC-TUIM-014a)")
	}
}

func TestPreviewDiffSubViewRendersOnAlternateScreen(t *testing.T) {
	withPreviewTheme(t, tui.LightTheme())

	m := newFixtureModel().selectRow()
	if m.currentView() != previewDiffView {
		t.Fatalf("selectRow did not enter the diff sub-view")
	}
	if v := m.View(); !v.AltScreen {
		t.Errorf("diff sub-view returned AltScreen=false; the diff must render on the alternate screen (AC-TUIM-014b)")
	}
}

func TestPreviewEscRestoresInlineTableSubView(t *testing.T) {
	withPreviewTheme(t, tui.LightTheme())

	m := newFixtureModel().selectRow()
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got, ok := updated.(*previewModel)
	if !ok {
		t.Fatalf("Update returned %T; want *previewModel", updated)
	}
	if got.currentView() != previewTableView {
		t.Errorf("esc from the diff sub-view did not restore the table sub-view (AC-TUIM-014c)")
	}
	if v := got.View(); v.AltScreen {
		t.Errorf("after esc the emitted view still carries AltScreen=true (AC-TUIM-014c)")
	}
}

// -------------------------------------------------------------------
// AC-TUIM-014d — every DIFF-REACHABLE quit key resolves to the inline sub-view
// -------------------------------------------------------------------

// TestPreviewDiffReachableQuitKeysResolveToInline exercises the discard hazard:
// the renderer's close path drops the final frame when the last view is an
// alt-screen view, so a quit entered FROM the diff sub-view would silently lose
// the exit summary. `n` is deliberately excluded — the diff branch returns
// before its case, so it is not diff-reachable by design (AC-TUIM-014f covers
// it on the table path).
func TestPreviewDiffReachableQuitKeysResolveToInline(t *testing.T) {
	withPreviewTheme(t, tui.LightTheme())

	cases := []struct {
		name          string
		msg           tea.KeyPressMsg
		wantConfirmed bool
	}{
		{"y from diff", key('y', "y"), true},
		{"q from diff", key('q', "q"), false},
		{"ctrl+c from diff", tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newFixtureModel().selectRow()
			if m.currentView() != previewDiffView {
				t.Fatalf("precondition failed: not in the diff sub-view")
			}

			updated, cmd := m.Update(tc.msg)
			got, ok := updated.(*previewModel)
			if !ok {
				t.Fatalf("Update returned %T; want *previewModel", updated)
			}
			if cmd == nil {
				t.Errorf("quit key %q returned a nil command; want tea.Quit", tc.name)
			}
			if got.currentView() != previewTableView {
				t.Errorf("quit from the diff sub-view left currentView=%v; want previewTableView — "+
					"the renderer discards an alt-screen final frame (AC-TUIM-014d)", got.currentView())
			}
			if got.confirmed != tc.wantConfirmed {
				t.Errorf("confirmed = %v; want %v", got.confirmed, tc.wantConfirmed)
			}

			assertSingleLineExitFrame(t, got)
		})
	}
}

// -------------------------------------------------------------------
// AC-TUIM-014f — the table-only cancel key `n`
// -------------------------------------------------------------------

func TestPreviewTableCancelKeyResolvesToInline(t *testing.T) {
	withPreviewTheme(t, tui.LightTheme())

	m := newFixtureModel()
	updated, cmd := m.Update(key('n', "n"))
	got, ok := updated.(*previewModel)
	if !ok {
		t.Fatalf("Update returned %T; want *previewModel", updated)
	}
	if cmd == nil {
		t.Errorf("`n` returned a nil command; want tea.Quit")
	}
	if got.currentView() != previewTableView {
		t.Errorf("`n` left currentView=%v; want previewTableView (AC-TUIM-014f)", got.currentView())
	}
	if got.confirmed {
		t.Errorf("`n` set confirmed=true; want false (AC-TUIM-014f)")
	}

	assertSingleLineExitFrame(t, got)
	if !strings.Contains(got.resultLine(), "cancelled") {
		t.Errorf("the `n` exit line does not name the cancelled outcome: %q", got.resultLine())
	}
}

// assertSingleLineExitFrame verifies the exit frame contract: the final view is
// inline, exactly one line, and carries neither the class-count card nor any
// file row.
func assertSingleLineExitFrame(t *testing.T, m *previewModel) {
	t.Helper()

	v := m.View()
	if v.AltScreen {
		t.Errorf("the exit frame carries AltScreen=true; it must be inline so the renderer preserves it")
	}
	if n := strings.Count(strings.TrimRight(v.Content, "\n"), "\n"); n != 0 {
		t.Errorf("the exit frame spans %d newlines; want a single line. Content:\n%q", n, v.Content)
	}
	if strings.Contains(v.Content, "Classification summary") {
		t.Errorf("the exit frame re-renders the classification card: %q", v.Content)
	}
	for _, f := range allFourClassesInputs() {
		if strings.Contains(v.Content, f.RelPath) {
			t.Errorf("the exit frame re-renders file row %q: %q", f.RelPath, v.Content)
		}
	}
}

// -------------------------------------------------------------------
// AC-TUIM-013 (structural half) — the fallback's class column is uniform
// -------------------------------------------------------------------

func TestPreviewFallbackClassColumnHasUniformWidth(t *testing.T) {
	classes := classifyAll(allFourClassesInputs(), allFourClassesPredicate())
	out := renderFallback(classes, true)

	offsets := make(map[int][]string)
	for _, c := range classes {
		line := fallbackLineFor(t, out, c.RelPath)
		idx := strings.Index(line, c.RelPath)
		if idx < 0 {
			t.Fatalf("file row for %q does not contain the path: %q", c.RelPath, line)
		}
		offsets[idx] = append(offsets[idx], c.RelPath)
	}
	if len(offsets) != 1 {
		t.Errorf("file rows start their path column at %d different offsets; the class column is not uniform "+
			"(AC-TUIM-013): %v\noutput:\n%s", len(offsets), offsets, out)
	}
}

// fallbackLineFor returns the single output line carrying the given path.
func fallbackLineFor(t *testing.T, out, path string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, path) {
			return line
		}
	}
	t.Fatalf("no fallback line carries path %q; output:\n%s", path, out)
	return ""
}
