package update

import "strings"

// previewFileColumnWidth is the File column width for a terminal of termWidth
// columns: the fixed Class column (24) and the table's padding (4) come off the
// top, and a zero or negative width falls back to the 80-column default.
func previewFileColumnWidth(termWidth int) int {
	if termWidth <= 0 {
		termWidth = 80
	}
	return termWidth - 24 - 4
}

// elidePath fits a relative path into width columns without cutting its file
// name. The table cuts an overlong cell from the right, which would drop the one
// part that identifies the row, so the leading directories are shortened to "…"
// and the name stays whole.
func elidePath(p string, width int) string {
	runes := []rune(p)
	if len(runes) <= width {
		return p
	}
	if width < 2 {
		return "…"
	}
	slash := strings.LastIndex(p, "/")
	if slash < 0 {
		return "…" + string(runes[len(runes)-(width-1):])
	}
	tail := []rune("…/" + p[slash+1:])
	if len(tail) >= width {
		return "…" + string(runes[len(runes)-(width-1):])
	}
	return string(runes[:width-len(tail)]) + string(tail)
}
