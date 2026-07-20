package update

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// preview_fallback.go implements the plain-text classification summary path
// (AC-TUX3-010). The fallback is structurally color-free: it NEVER emits ANSI
// escape sequences, regardless of NO_COLOR or TTY state, because color
// rendering belongs to the TUI surface only. This makes the `\x1b[` absence
// assertion (AC-TUX3-010) a structural guarantee rather than a conditional
// branch.
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

// renderFallback renders the plain-text classification summary. It is
// structurally color-free (no ANSI escapes ever) — the noColor parameter is
// accepted for API symmetry with the TUI surface and to document the caller's
// intent; the output is identical regardless of its value.
//
// The summary contains:
//   - A per-class count line for every class with count >= 1 (each line
//     carries the class label, satisfying AC-TUX3-014 for the preserve class)
//   - One row per file: "<class label>\t<relPath>"
func renderFallback(classes []FileClassification, noColor bool) string {
	counts := countByClass(classes)
	var b strings.Builder

	b.WriteString("moai update — change preview\n")
	b.WriteString("\n")
	b.WriteString("Classification summary:\n")
	for _, class := range classOrder {
		n := counts[class]
		if n == 0 {
			continue
		}
		fmt.Fprintf(&b, "  %s: %d\n", class.String(), n)
	}

	b.WriteString("\n")
	b.WriteString("Files:\n")
	for _, c := range classes {
		fmt.Fprintf(&b, "  %s\t%s\n", c.Class.String(), c.RelPath)
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
