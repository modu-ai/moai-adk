package output

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TableRenderer renders styled tables
type TableRenderer struct {
	headers []string
	rows    [][]string
	style   *lipgloss.Style
}

// NewTableRenderer creates a new table renderer
func NewTableRenderer(headers []string) *TableRenderer {
	return &TableRenderer{
		headers: headers,
		rows:    make([][]string, 0),
		style:   &TableCellStyle,
	}
}

// AddRow adds a row to the table
func (t *TableRenderer) AddRow(cells ...string) {
	t.rows = append(t.rows, cells)
}

// Render renders the table as a string
func (t *TableRenderer) Render() string {
	if len(t.headers) == 0 || len(t.rows) == 0 {
		return ""
	}

	// Calculate column widths
	colWidths := make([]int, len(t.headers))
	for i, header := range t.headers {
		colWidths[i] = len(header)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Build separator
	separator := "+"
	for _, width := range colWidths {
		separator += strings.Repeat("-", width+2) + "+"
	}

	// Build header row
	headerRow := "|"
	for i, header := range t.headers {
		padded := TableHeaderStyle.Render(header)
		headerRow += " " + padRight(padded, colWidths[i]) + " |"
	}

	// Build data rows
	var dataRows []string
	for _, row := range t.rows {
		dataRow := "|"
		for i, cell := range row {
			if i < len(colWidths) {
				styled := t.style.Render(cell)
				dataRow += " " + padRight(styled, colWidths[i]) + " |"
			}
		}
		dataRows = append(dataRows, dataRow)
	}

	// Combine all parts
	result := separator + "\n"
	result += headerRow + "\n"
	result += separator + "\n"
	for _, dataRow := range dataRows {
		result += dataRow + "\n"
	}
	result += separator

	return result
}

// padRight pads a string to the right with spaces
func padRight(s string, width int) string {
	// Remove ANSI codes for width calculation
	plain := stripANSI(s)
	if len(plain) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(plain))
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	// Simple ANSI code removal (not exhaustive)
	result := s
	inEscape := false
	for i := 0; i < len(result); i++ {
		if result[i] == '\x1b' {
			inEscape = true
		}
		if inEscape && (result[i] == 'm' || result[i] == 'A' || result[i] == 'K') {
			inEscape = false
			result = result[:i] + " " + result[i+1:]
		}
		if inEscape {
			result = result[:i] + result[i+1:]
			i--
		}
	}
	return result
}
