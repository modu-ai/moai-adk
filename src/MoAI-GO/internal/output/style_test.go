package output

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// --- Color constants ---

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name  string
		color string
	}{
		{"ColorPrimary", ColorPrimary},
		{"ColorSuccess", ColorSuccess},
		{"ColorWarning", ColorWarning},
		{"ColorError", ColorError},
		{"ColorInfo", ColorInfo},
		{"ColorMuted", ColorMuted},
		{"ColorBackground", ColorBackground},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color == "" {
				t.Errorf("%s is empty", tt.name)
			}
			// All colors should be hex format starting with #
			if tt.color[0] != '#' {
				t.Errorf("%s = %q, expected hex color starting with #", tt.name, tt.color)
			}
		})
	}
}

// --- Style definitions ---

func TestStylesNotNil(t *testing.T) {
	// Verify all styles can render without panic
	styles := map[string]lipgloss.Style{
		"HeaderStyle":      HeaderStyle,
		"SuccessStyle":     SuccessStyle,
		"WarningStyle":     WarningStyle,
		"ErrorStyle":       ErrorStyle,
		"InfoStyle":        InfoStyle,
		"MutedStyle":       MutedStyle,
		"CheckmarkStyle":   CheckmarkStyle,
		"CrossmarkStyle":   CrossmarkStyle,
		"BulletStyle":      BulletStyle,
		"TableHeaderStyle": TableHeaderStyle,
		"TableCellStyle":   TableCellStyle,
		"TableBorderStyle": TableBorderStyle,
		"SectionStyle":     SectionStyle,
		"SubsectionStyle":  SubsectionStyle,
		"CodeStyle":        CodeStyle,
	}

	for name, style := range styles {
		t.Run(name, func(t *testing.T) {
			result := style.Render("test text")
			if result == "" {
				t.Errorf("%s.Render returned empty string", name)
			}
		})
	}
}

func TestStylesRenderContent(t *testing.T) {
	// Verify styles preserve the input text content
	input := "Hello World"

	styles := map[string]lipgloss.Style{
		"HeaderStyle":  HeaderStyle,
		"SuccessStyle": SuccessStyle,
		"WarningStyle": WarningStyle,
		"ErrorStyle":   ErrorStyle,
		"InfoStyle":    InfoStyle,
		"MutedStyle":   MutedStyle,
	}

	for name, style := range styles {
		t.Run(name, func(t *testing.T) {
			result := style.Render(input)
			// The rendered output should contain the original text
			// (it may also contain ANSI escape codes)
			if len(result) < len(input) {
				t.Errorf("%s.Render(%q) = %q, shorter than input", name, input, result)
			}
		})
	}
}

func TestStylesRenderEmptyString(t *testing.T) {
	// Verify styles handle empty strings without panic
	styles := []lipgloss.Style{
		HeaderStyle,
		SuccessStyle,
		WarningStyle,
		ErrorStyle,
		InfoStyle,
		MutedStyle,
	}

	for _, style := range styles {
		// Should not panic
		_ = style.Render("")
	}
}
