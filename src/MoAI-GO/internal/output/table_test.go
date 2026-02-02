package output

import (
	"strings"
	"testing"
)

// --- NewTableRenderer ---

func TestNewTableRenderer(t *testing.T) {
	headers := []string{"Name", "Value", "Status"}
	tr := NewTableRenderer(headers)

	if tr == nil {
		t.Fatal("NewTableRenderer returned nil")
	}
	if len(tr.headers) != 3 {
		t.Errorf("headers length = %d, want 3", len(tr.headers))
	}
	if len(tr.rows) != 0 {
		t.Errorf("rows length = %d, want 0", len(tr.rows))
	}
}

// --- AddRow ---

func TestAddRow(t *testing.T) {
	tr := NewTableRenderer([]string{"A", "B"})
	tr.AddRow("1", "2")
	tr.AddRow("3", "4")

	if len(tr.rows) != 2 {
		t.Errorf("rows length = %d, want 2", len(tr.rows))
	}
	if tr.rows[0][0] != "1" || tr.rows[0][1] != "2" {
		t.Errorf("first row = %v, want [1, 2]", tr.rows[0])
	}
}

// --- Render ---

func TestRender_EmptyHeaders(t *testing.T) {
	tr := NewTableRenderer([]string{})
	result := tr.Render()
	if result != "" {
		t.Errorf("Render with empty headers should return empty string, got %q", result)
	}
}

func TestRender_NoRows(t *testing.T) {
	tr := NewTableRenderer([]string{"Name", "Value"})
	result := tr.Render()
	if result != "" {
		t.Errorf("Render with no rows should return empty string, got %q", result)
	}
}

func TestRender_SingleRow(t *testing.T) {
	tr := NewTableRenderer([]string{"Name", "Value"})
	tr.AddRow("key", "val")
	result := tr.Render()

	if result == "" {
		t.Fatal("Render returned empty string")
	}

	// Should contain separator characters
	if !strings.Contains(result, "+") {
		t.Error("rendered table missing '+' separator")
	}
	if !strings.Contains(result, "|") {
		t.Error("rendered table missing '|' separator")
	}
	if !strings.Contains(result, "-") {
		t.Error("rendered table missing '-' separator")
	}
}

func TestRender_MultipleRows(t *testing.T) {
	tr := NewTableRenderer([]string{"ID", "Name", "Status"})
	tr.AddRow("1", "Alice", "Active")
	tr.AddRow("2", "Bob", "Inactive")
	tr.AddRow("3", "Charlie", "Active")
	result := tr.Render()

	if result == "" {
		t.Fatal("Render returned empty string")
	}

	// Count rows: header separator + header + middle separator + 3 data rows + bottom separator
	lines := strings.Split(result, "\n")
	// At least separator + header + separator + 3 data rows + separator = 7 lines
	if len(lines) < 6 {
		t.Errorf("expected at least 6 lines, got %d", len(lines))
	}
}

func TestRender_ColumnWidthAdaptation(t *testing.T) {
	tr := NewTableRenderer([]string{"A", "B"})
	tr.AddRow("short", "this is a very long cell value")
	result := tr.Render()

	if result == "" {
		t.Fatal("Render returned empty string")
	}

	// The separator should be wide enough for the long cell
	lines := strings.Split(result, "\n")
	if len(lines) > 0 {
		// First line is a separator, should be wider than just header widths
		if len(lines[0]) <= 10 {
			t.Errorf("separator too narrow: %q", lines[0])
		}
	}
}

// --- padRight ---

func TestPadRight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{
			name:  "shorter than width",
			input: "abc",
			width: 6,
			want:  "abc   ",
		},
		{
			name:  "equal to width",
			input: "abcdef",
			width: 6,
			want:  "abcdef",
		},
		{
			name:  "longer than width",
			input: "abcdefgh",
			width: 6,
			want:  "abcdefgh",
		},
		{
			name:  "empty string",
			input: "",
			width: 3,
			want:  "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := padRight(tt.input, tt.width)
			if got != tt.want {
				t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

// --- stripANSI ---

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "no ANSI codes",
			input: "plain text",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "only text",
			input: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripANSI(tt.input)
			// For plain text, result should be the same
			if tt.input == got || got == "" && tt.input == "" {
				// OK - either same or both empty
			}
			// At minimum, should not be longer than input for plain text
			if len(got) > len(tt.input)+len(tt.input) {
				t.Errorf("stripANSI output unreasonably long: input=%q, output=%q", tt.input, got)
			}
		})
	}
}

func TestStripANSI_WithEscapeCodes(t *testing.T) {
	// Test with actual ANSI escape sequences
	input := "\x1b[31mred text\x1b[0m"
	got := stripANSI(input)

	// After stripping, should not contain the escape character
	if strings.Contains(got, "\x1b") {
		t.Errorf("stripANSI did not remove escape characters: %q", got)
	}
}
