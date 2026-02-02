package output

import (
	"testing"
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
	styles := map[string]func(string) string{
		"HeaderStyle":      HeaderStyle.Render,
		"SuccessStyle":     SuccessStyle.Render,
		"WarningStyle":     WarningStyle.Render,
		"ErrorStyle":       ErrorStyle.Render,
		"InfoStyle":        InfoStyle.Render,
		"MutedStyle":       MutedStyle.Render,
		"CheckmarkStyle":   CheckmarkStyle.Render,
		"CrossmarkStyle":   CrossmarkStyle.Render,
		"BulletStyle":      BulletStyle.Render,
		"TableHeaderStyle": TableHeaderStyle.Render,
		"TableCellStyle":   TableCellStyle.Render,
		"TableBorderStyle": TableBorderStyle.Render,
		"SectionStyle":     SectionStyle.Render,
		"SubsectionStyle":  SubsectionStyle.Render,
		"CodeStyle":        CodeStyle.Render,
	}

	for name, renderFn := range styles {
		t.Run(name, func(t *testing.T) {
			result := renderFn("test text")
			if result == "" {
				t.Errorf("%s.Render returned empty string", name)
			}
		})
	}
}

func TestStylesRenderContent(t *testing.T) {
	// Verify styles preserve the input text content
	input := "Hello World"

	tests := []struct {
		name   string
		render func(string) string
	}{
		{"HeaderStyle", HeaderStyle.Render},
		{"SuccessStyle", SuccessStyle.Render},
		{"WarningStyle", WarningStyle.Render},
		{"ErrorStyle", ErrorStyle.Render},
		{"InfoStyle", InfoStyle.Render},
		{"MutedStyle", MutedStyle.Render},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.render(input)
			// The rendered output should contain the original text
			// (it may also contain ANSI escape codes)
			if len(result) < len(input) {
				t.Errorf("%s.Render(%q) = %q, shorter than input", tt.name, input, result)
			}
		})
	}
}

func TestStylesRenderEmptyString(t *testing.T) {
	// Verify styles handle empty strings without panic
	styles := []func(string) string{
		HeaderStyle.Render,
		SuccessStyle.Render,
		WarningStyle.Render,
		ErrorStyle.Render,
		InfoStyle.Render,
		MutedStyle.Render,
	}

	for i, renderFn := range styles {
		// Should not panic
		_ = renderFn("")
		_ = i
	}
}
