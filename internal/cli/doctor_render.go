package cli

// doctor result-table render layer (SPEC-CLI-TUX-V3-004 M4c, REQ-TUX4-002/003).
//
// Separated from doctor.go so the verdict logic diff there stays limited to
// the progress reporter seam (AC-TUX4-014 render-layer separation proof).
//
// Rich path (TTY + colour): bubbles v2 table styled from internal/tui tokens.
// Plain path (non-TTY or NO_COLOR): aligned plain-text table, zero ANSI —
// golden-test deterministic. No hex literals in this file (AC-CLI-TUI-013).

import (
	"fmt"
	"io"
	"strings"

	btable "charm.land/bubbles/v2/table"
	lipglossv2 "charm.land/lipgloss/v2"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// doctorTableHeaders are the per-section result table column titles.
var doctorTableHeaders = [3]string{"STATUS", "CHECK", "MESSAGE"}

// groupCounts tallies ok/warn/fail for one check group.
func groupCounts(g checkGroup) (ok, warn, fail int) {
	for _, c := range g.checks {
		switch c.Status {
		case uikit.CheckOK:
			ok++
		case uikit.CheckWarn:
			warn++
		case uikit.CheckFail:
			fail++
		}
	}
	return ok, warn, fail
}

// doctorColumnWidths computes the status/check column widths from content.
func doctorColumnWidths(g checkGroup) (statusW, checkW int) {
	statusW = len(doctorTableHeaders[0])
	checkW = len(doctorTableHeaders[1])
	for _, c := range g.checks {
		statusW = max(statusW, len(string(c.Status)))
		checkW = max(checkW, len(c.Name))
	}
	return statusW, checkW
}

// renderDoctorPlainTable renders one group as an aligned plain-text table
// (REQ-TUX4-002 non-TTY pair; zero ANSI).
func renderDoctorPlainTable(g checkGroup) []string {
	statusW, checkW := doctorColumnWidths(g)
	var lines []string
	lines = append(lines, fmt.Sprintf("  %-*s  %-*s  %s",
		statusW, doctorTableHeaders[0], checkW, doctorTableHeaders[1], doctorTableHeaders[2]))
	for _, c := range g.checks {
		lines = append(lines, fmt.Sprintf("  %-*s  %-*s  %s",
			statusW, string(c.Status), checkW, c.Name, c.Message))
	}
	return lines
}

// renderDoctorRichTable renders one group through the bubbles v2 table
// component styled from tui tokens (REQ-TUX4-002 TTY pair; AC-TUX4-002
// bubbles-v2 reachability).
func renderDoctorRichTable(g checkGroup, th tui.Theme) string {
	statusW, checkW := doctorColumnWidths(g)
	msgW := len(doctorTableHeaders[2])
	for _, c := range g.checks {
		msgW = max(msgW, len(c.Message))
	}

	cols := []btable.Column{
		{Title: doctorTableHeaders[0], Width: statusW},
		{Title: doctorTableHeaders[1], Width: checkW},
		{Title: doctorTableHeaders[2], Width: msgW},
	}
	rows := make([]btable.Row, 0, len(g.checks))
	for _, c := range g.checks {
		rows = append(rows, btable.Row{string(c.Status), c.Name, c.Message})
	}

	cellStyle := lipglossv2.NewStyle().Foreground(lipglossv2.Color(th.Body)).Padding(0, 1)
	styles := btable.Styles{
		Header:   lipglossv2.NewStyle().Foreground(lipglossv2.Color(th.Accent)).Bold(true).Padding(0, 1),
		Cell:     cellStyle,
		Selected: lipglossv2.NewStyle(), // static render: no selection highlight
	}

	// Total width: column content widths + per-cell horizontal padding (2 each).
	// The table viewport defaults to width 0 and renders no rows without an
	// explicit width, so this is load-bearing, not cosmetic.
	totalW := statusW + checkW + msgW + 6

	t := btable.New(
		btable.WithColumns(cols),
		btable.WithRows(rows),
		btable.WithHeight(len(rows)+1),
		btable.WithWidth(totalW),
		btable.WithStyles(styles),
	)
	return strings.TrimRight(t.View(), "\n ")
}

// renderDoctorGroups renders the grouped doctor results as per-section
// pass/fail tables with per-section and overall counts, wrapped in the
// System Diagnostics box (REQ-TUX4-002). The table backend is selected by
// the same rich/plain predicate as the glamour surfaces: bubbles v2 table on
// a colour-capable terminal, aligned plain text otherwise (REQ-TUX4-003).
func renderDoctorGroups(out io.Writer, groups []checkGroup, verbose bool, th tui.Theme) string {
	rich := markdownRichEnabled((tui.OSEnv{}).NoColor(), writerIsTerminal(out))

	var bodyLines []string
	okTotal, warnTotal, failTotal := 0, 0, 0
	for _, g := range groups {
		if len(g.checks) == 0 {
			continue
		}
		ok, warn, fail := groupCounts(g)
		okTotal += ok
		warnTotal += warn
		failTotal += fail

		bodyLines = append(bodyLines, tui.Section(g.title, tui.SectionOpts{Theme: &th}))
		if rich {
			bodyLines = append(bodyLines, renderDoctorRichTable(g, th))
		} else {
			bodyLines = append(bodyLines, renderDoctorPlainTable(g)...)
		}
		if verbose {
			for _, c := range g.checks {
				if c.Detail != "" {
					bodyLines = append(bodyLines, "    "+c.Name+": "+c.Detail)
				}
			}
		}
		// Per-section counts (REQ-TUX4-002).
		bodyLines = append(bodyLines, fmt.Sprintf("  %d ok, %d warn, %d fail", ok, warn, fail))
		bodyLines = append(bodyLines, "")
	}

	// Overall summary pill row (Pass/Warn/Fail counts).
	pPass := tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: fmt.Sprintf("Pass %d", okTotal), Theme: &th})
	pWarn := tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: fmt.Sprintf("Warn %d", warnTotal), Theme: &th})
	pErr := tui.Pill(tui.PillOpts{Kind: tui.PillErr, Solid: false, Label: fmt.Sprintf("Fail %d", failTotal), Theme: &th})
	bodyLines = append(bodyLines, pPass+"  "+pWarn+"  "+pErr)

	return tui.Box(tui.BoxOpts{
		Title: "System Diagnostics",
		Body:  strings.Join(bodyLines, "\n"),
		Theme: &th,
	})
}
