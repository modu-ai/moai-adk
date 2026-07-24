package update

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// preview_fallback.go implements the plain-text classification summary path
// (AC-TUX3-010). The fallback is structurally color-free: it NEVER emits ANSI
// escape sequences, regardless of NO_COLOR or TTY state, because color
// rendering belongs to the TUI surface only. This makes the `\x1b` absence
// assertion (AC-TUX3-010, AC-TUIM-011) a structural guarantee rather than a
// conditional branch — which is precisely why this file imports neither
// lipgloss nor internal/tui, and why the guarantee must never be re-expressed
// as `if !noColor { ... }` (REQ-TUIM-018).
//
// The layout is drawn with ASCII box characters and whitespace only: an
// ASCII-bordered summary card plus a column-aligned file list whose class
// column has a uniform width across every row.
//
// The fallback consumes the SAME classification model (classifyAll → Classify)
// as the TUI (REQ-TUX3-001/002 single source of truth).

// classOrder is the stable display order for the per-class count summary.
var classOrder = []ChangeClass{ClassAdd, ClassUpdate, ClassPreserveUserOwned, ClassConflict}

// countByClass tallies classifications per ChangeClass.
func countByClass(classes []FileClassification) map[ChangeClass]int {
	counts := make(map[ChangeClass]int, 4)
	for _, c := range classes {
		counts[c.Class]++
	}
	return counts
}

// displayWidth returns the column width of s for the fallback's alignment
// arithmetic. It counts runes rather than terminal cells: every string this
// file aligns on (class labels, counts, the ASCII card chrome) is ASCII, and
// the one non-ASCII character in the layout — the em dash in the title — is a
// single-cell rune. A double-width CJK codepoint inside a file path would
// widen the row beyond the computed padding, but file paths are printed last
// on their line and so cannot disturb any other column.
func displayWidth(s string) int { return utf8.RuneCountInString(s) }

// padRight pads s with spaces to the given display width.
func padRight(s string, width int) string {
	if pad := width - displayWidth(s); pad > 0 {
		return s + strings.Repeat(" ", pad)
	}
	return s
}

// renderCard draws an ASCII-bordered card around a title and a body block.
// The card width adapts to its widest line. Every character it emits is ASCII;
// no escape sequence is produced.
func renderCard(title string, lines []string) string {
	inner := displayWidth(title)
	for _, l := range lines {
		inner = max(inner, displayWidth(l))
	}

	edge := "+" + strings.Repeat("-", inner+2) + "+"
	rule := "|" + strings.Repeat("-", inner+2) + "|"

	var b strings.Builder
	b.WriteString(edge + "\n")
	b.WriteString("| " + padRight(title, inner) + " |\n")
	if len(lines) > 0 {
		b.WriteString(rule + "\n")
		for _, l := range lines {
			b.WriteString("| " + padRight(l, inner) + " |\n")
		}
	}
	b.WriteString(edge + "\n")
	return b.String()
}

// renderFallback renders the plain-text classification summary. It is
// structurally color-free (no ANSI escapes ever) — the noColor parameter is
// accepted for API symmetry with the TUI surface and to document the caller's
// intent; the output is identical regardless of its value.
//
// The summary contains:
//   - A card carrying a per-class count line for every class with count >= 1
//     (each line carries the class label, satisfying AC-TUX3-014 for the
//     preserve class)
//   - One aligned row per file: "<class label padded><relPath>", where the
//     class column width is uniform across every row (AC-TUIM-013)
func renderFallback(classes []FileClassification, noColor bool) string {
	_ = noColor // structurally color-free; see the file comment

	counts := countByClass(classes)

	// One class-column width shared by the summary card and the file list, so
	// the two blocks line up with each other as well as internally.
	classWidth := 0
	for _, class := range classOrder {
		classWidth = max(classWidth, displayWidth(class.String()))
	}

	summary := make([]string, 0, len(classOrder))
	for _, class := range classOrder {
		n := counts[class]
		if n == 0 {
			continue
		}
		summary = append(summary, fmt.Sprintf("%s  %d", padRight(class.String(), classWidth), n))
	}

	var b strings.Builder
	b.WriteString("moai update — change preview\n")
	b.WriteString("\n")
	b.WriteString(renderCard("Classification summary", summary))
	b.WriteString("\n")
	b.WriteString("Files\n")
	for _, c := range classes {
		fmt.Fprintf(&b, "  %s  %s\n", padRight(c.Class.String(), classWidth), c.RelPath)
	}
	return b.String()
}

// renderFallbackToStdout writes the fallback summary to stdout and returns
// confirmed=true (the non-interactive path always proceeds — there is no
// interactive confirmation gate when --yes is set or the process is non-TTY).
// The return value satisfies the PreviewClassification signature.
func renderFallbackToStdout(classes []FileClassification, opts PreviewOptions) bool {
	_, _ = io.WriteString(os.Stdout, renderFallback(classes, opts.NoColor))
	return true
}
