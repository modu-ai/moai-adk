package ui

import (
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/glamour"
)

// defaultTerminalWidth is used when the terminal width cannot be detected.
const defaultTerminalWidth = 80

// RenderMarkdown renders markdown text with glamour styling.
// It auto-detects dark/light terminal background and applies word wrapping
// based on the current terminal width. In NoColor mode or non-TTY environments,
// it returns the input unchanged.
func RenderMarkdown(md string) (string, error) {
	if noColorFromEnv() {
		return md, nil
	}

	width := terminalWidth()

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", fmt.Errorf("create markdown renderer: %w", err)
	}

	out, err := renderer.Render(md)
	if err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}

	return out, nil
}

// RenderMarkdownWithTheme renders markdown using the Theme's dark/light preference.
// It falls back to plain text when NoColor is set.
func RenderMarkdownWithTheme(md string, theme *Theme) (string, error) {
	if theme.NoColor {
		return md, nil
	}

	width := terminalWidth()

	var style glamour.TermRendererOption
	if theme.IsDark {
		style = glamour.WithStylePath("dark")
	} else {
		style = glamour.WithStylePath("light")
	}

	renderer, err := glamour.NewTermRenderer(
		style,
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", fmt.Errorf("create markdown renderer: %w", err)
	}

	out, err := renderer.Render(md)
	if err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}

	return out, nil
}

// noColorFromEnv checks whether color output is disabled via environment.
func noColorFromEnv() bool {
	if env := os.Getenv("MOAI_NO_COLOR"); env == "true" || env == "1" {
		return true
	}
	if env := os.Getenv("NO_COLOR"); env != "" {
		return true
	}
	return false
}

// terminalWidth returns the current terminal width.
// It reads from the COLUMNS environment variable or falls back to the default.
func terminalWidth() int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if w, err := strconv.Atoi(cols); err == nil && w > 0 {
			return w
		}
	}
	return defaultTerminalWidth
}
