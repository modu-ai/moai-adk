// Package internal provides low-level TUI utilities used exclusively by the
// parent tui package. Callers outside internal/tui/ must not import this
// package directly; they should use the tui package's public API instead.
//
// # Korean / CJK Width Handling
//
// Terminal rendering of CJK characters requires accurate display-width
// calculation because each CJK glyph occupies 2 terminal cells while ASCII
// occupies 1. Mismatching these values causes box borders to mis-align.
//
// This package wraps github.com/mattn/go-runewidth so that it is the ONLY
// file in internal/tui/ that imports go-runewidth directly (REQ-CLI-TUI-003).
// All other tui files call StringWidth / Truncate / FillRight via this package.
package internal

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// StringWidth returns the display width of s in terminal cells.
// ASCII characters count as 1 cell; CJK characters count as 2 cells.
// This is a thin wrapper around runewidth.StringWidth.
func StringWidth(s string) int {
	return runewidth.StringWidth(s)
}

// Truncate shortens s to at most n terminal cells, never splitting a
// multi-cell character mid-glyph. The returned string fits within n cells.
func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	w := 0
	var b strings.Builder
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > n {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String()
}

// FillRight pads s with spaces on the right until its display width equals
// width. If s is already wider than width, it is returned unchanged.
// This enables column alignment of mixed-width (ASCII + CJK) text.
func FillRight(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}
