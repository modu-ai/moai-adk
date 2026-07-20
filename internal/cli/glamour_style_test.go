package cli

// M4a/M4b contract tests for the glamour render layer (SPEC-CLI-TUX-V3-004
// REQ-TUX4-004/005). TDD RED-first: these tests define the glamour_style.go
// contract before implementation.
//
// Style-source contract (AC-TUX4-005): every colour in the glamour style
// config must reference an internal/tui Theme token — no hex literals are
// authored in glamour_style.go (grep-verified by the AC command).

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestGlamourStyleFromTheme_TokenMapping verifies the tui.Theme → glamour
// ansi.StyleConfig mapping references theme tokens for every colour-bearing
// element (REQ-TUX4-004: "style derived from internal/tui tokens").
func TestGlamourStyleFromTheme_TokenMapping(t *testing.T) {
	for _, tc := range []struct {
		name string
		th   tui.Theme
	}{
		{"light", tui.LightTheme()},
		{"dark", tui.DarkTheme()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := glamourStyleFromTheme(tc.th)

			if cfg.Document.Color == nil || *cfg.Document.Color != tc.th.Body {
				t.Errorf("Document.Color = %v, want Body token %q", cfg.Document.Color, tc.th.Body)
			}
			if cfg.Heading.Color == nil || *cfg.Heading.Color != tc.th.Accent {
				t.Errorf("Heading.Color = %v, want Accent token %q", cfg.Heading.Color, tc.th.Accent)
			}
			if cfg.Strong.Color == nil || *cfg.Strong.Color != tc.th.Fg {
				t.Errorf("Strong.Color = %v, want Fg token %q", cfg.Strong.Color, tc.th.Fg)
			}
			if cfg.Link.Color == nil || *cfg.Link.Color != tc.th.Info {
				t.Errorf("Link.Color = %v, want Info token %q", cfg.Link.Color, tc.th.Info)
			}
			if cfg.Code.Color == nil || *cfg.Code.Color != tc.th.Info {
				t.Errorf("Code.Color = %v, want Info token %q", cfg.Code.Color, tc.th.Info)
			}
			if cfg.BlockQuote.Color == nil || *cfg.BlockQuote.Color != tc.th.Dim {
				t.Errorf("BlockQuote.Color = %v, want Dim token %q", cfg.BlockQuote.Color, tc.th.Dim)
			}
			if cfg.HorizontalRule.Color == nil || *cfg.HorizontalRule.Color != tc.th.Rule {
				t.Errorf("HorizontalRule.Color = %v, want Rule token %q", cfg.HorizontalRule.Color, tc.th.Rule)
			}
		})
	}
}

// TestGlamourStyleFromTheme_LightDarkDiffer guards against a hard-coded
// single-palette config: the style must follow the injected theme.
func TestGlamourStyleFromTheme_LightDarkDiffer(t *testing.T) {
	light := glamourStyleFromTheme(tui.LightTheme())
	dark := glamourStyleFromTheme(tui.DarkTheme())
	if *light.Heading.Color == *dark.Heading.Color {
		t.Errorf("light and dark heading colours are identical (%q) — style must derive from the injected theme", *light.Heading.Color)
	}
}

// TestMarkdownRichEnabled_Matrix verifies the rich/plain routing decision
// (REQ-TUX4-005: non-TTY or NO_COLOR → plain markdown passthrough).
func TestMarkdownRichEnabled_Matrix(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
		tty     bool
		want    bool
	}{
		{"tty_color", false, true, true},
		{"tty_nocolor", true, true, false},
		{"pipe_color", false, false, false},
		{"pipe_nocolor", true, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := markdownRichEnabled(tt.noColor, tt.tty); got != tt.want {
				t.Errorf("markdownRichEnabled(noColor=%v, tty=%v) = %v, want %v", tt.noColor, tt.tty, got, tt.want)
			}
		})
	}
}

// TestRenderMarkdown_PlainPassthroughNonTTY verifies the non-TTY fallback is a
// byte-stable passthrough with zero ANSI escapes (REQ-TUX4-005, AC-TUX4-006).
func TestRenderMarkdown_PlainPassthroughNonTTY(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	buf := new(bytes.Buffer) // non-*os.File writer → non-TTY
	md := "# Project Status\n\n- **Project**: demo\n"
	got := renderMarkdown(buf, md)
	if got != md {
		t.Errorf("non-TTY renderMarkdown must pass markdown through unchanged\ngot:  %q\nwant: %q", got, md)
	}
	if strings.Contains(got, "\x1b") {
		t.Error("non-TTY passthrough must not contain ANSI escape sequences")
	}
}

// TestRenderMarkdown_NoColorForcesPassthrough verifies NO_COLOR wins even
// before TTY detection (series-common isEnvSet semantics: non-empty = set).
func TestRenderMarkdown_NoColorForcesPassthrough(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	buf := new(bytes.Buffer)
	md := "## Heading\n\nbody\n"
	if got := renderMarkdown(buf, md); got != md {
		t.Errorf("NO_COLOR renderMarkdown must pass markdown through unchanged, got %q", got)
	}
}

// TestGlamourRender_RichOutputStyled verifies the rich path actually routes
// through glamour with the token-derived style: output is transformed and
// carries ANSI styling (AC-TUX4-004 reachability — glamour is really wired,
// not just imported).
func TestGlamourRender_RichOutputStyled(t *testing.T) {
	md := "# Hello\n\nworld\n"
	out, err := glamourRender(md, tui.DarkTheme())
	if err != nil {
		t.Fatalf("glamourRender: %v", err)
	}
	if out == md {
		t.Error("rich glamour render should transform the markdown, got identical output")
	}
	if !strings.Contains(out, "Hello") {
		t.Errorf("rendered output lost heading text: %q", out)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Error("rich glamour render should carry ANSI styling from theme tokens")
	}
}
