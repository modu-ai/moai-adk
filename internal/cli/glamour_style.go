package cli

// glamour render layer for markdown-bearing CLI surfaces (SPEC-CLI-TUX-V3-004
// REQ-TUX4-004/005, M4a).
//
// # Style source (AC-TUX4-005)
//
// Every colour in the glamour style config references an internal/tui Theme
// token. No hex literal is authored in this file (AC-CLI-TUI-013 succession);
// the AC grep `#[0-9A-Fa-f]{6}` over this file must stay at zero.
//
// # Rich/plain routing (REQ-TUX4-005)
//
// The rich glamour path runs only when the destination writer is a real
// terminal AND NO_COLOR is unset. Non-TTY (pipes, CI, golden-test buffers)
// and NO_COLOR fall back to plain markdown passthrough with zero ANSI
// escape sequences, keeping golden output deterministic.

import (
	"io"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// glamourWordWrap is the fixed render width for the rich path. A fixed width
// (rather than live terminal-width probing) keeps line-wrap positions stable
// across environments (acceptance.md §C "glamour 폭 계산" edge: fixed-width
// fallback so goldens cannot destabilise).
const glamourWordWrap = 100

// strptr / boolptr adapt tui string tokens to glamour's pointer-based
// ansi.StylePrimitive fields.
func strptr(s string) *string { return &s }
func boolptr(b bool) *bool    { return &b }

// glamourStyleFromTheme maps internal/tui design tokens onto a glamour
// ansi.StyleConfig. Mapping decision (recorded in progress.md §E.2):
//
//	Document / body text  → Theme.Body
//	Headings (all levels) → Theme.Accent, bold
//	Strong                → Theme.Fg, bold
//	Emph                  → Theme.Body, italic
//	Link / LinkText       → Theme.Info (link underlined)
//	Inline code           → Theme.Info
//	Code block            → Theme.Body
//	BlockQuote            → Theme.Dim
//	HorizontalRule        → Theme.Rule
//	List item / enum      → Theme.Body
//	Table                 → Theme.Body
func glamourStyleFromTheme(th tui.Theme) ansi.StyleConfig {
	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: strptr(th.Body)},
			Margin:         uintptr2(0),
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:       strptr(th.Accent),
				Bold:        boolptr(true),
				BlockSuffix: "\n",
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strptr(th.Accent),
				Bold:   boolptr(true),
				Prefix: "# ",
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strptr(th.Accent),
				Bold:   boolptr(true),
				Prefix: "## ",
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strptr(th.Accent),
				Bold:   boolptr(true),
				Prefix: "### ",
			},
		},
		Text: ansi.StylePrimitive{Color: strptr(th.Body)},
		Strong: ansi.StylePrimitive{
			Color: strptr(th.Fg),
			Bold:  boolptr(true),
		},
		Emph: ansi.StylePrimitive{
			Color:  strptr(th.Body),
			Italic: boolptr(true),
		},
		HorizontalRule: ansi.StylePrimitive{Color: strptr(th.Rule)},
		Item:           ansi.StylePrimitive{Color: strptr(th.Body)},
		Enumeration:    ansi.StylePrimitive{Color: strptr(th.Body)},
		Link: ansi.StylePrimitive{
			Color:     strptr(th.Info),
			Underline: boolptr(true),
		},
		LinkText: ansi.StylePrimitive{Color: strptr(th.Info)},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: strptr(th.Dim)},
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{Color: strptr(th.Info)},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{Color: strptr(th.Body)},
			},
		},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{Color: strptr(th.Body)},
			},
		},
	}
}

// uintptr2 returns a pointer to a uint (glamour margin fields).
func uintptr2(v uint) *uint { return &v }

// markdownRichEnabled reports whether the rich glamour path is active.
// Pure function for testability: NO_COLOR (any non-empty value, series
// isEnvSet semantics) or a non-terminal writer forces the plain path.
func markdownRichEnabled(noColor, isTerminal bool) bool {
	return !noColor && isTerminal
}

// writerIsTerminal reports whether w is a character-device-backed *os.File
// (mirrors spec_status.go stdinIsTerminal; buffers, pipes and redirected
// files are non-TTY, so golden-test captures stay deterministic).
func writerIsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// glamourRender always renders through glamour with the token-derived style
// (the rich path body). Split from renderMarkdown so tests can exercise the
// glamour wiring without a real terminal.
func glamourRender(md string, th tui.Theme) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(glamourStyleFromTheme(th)),
		glamour.WithWordWrap(glamourWordWrap),
	)
	if err != nil {
		return "", err
	}
	return r.Render(md)
}

// renderMarkdown is the shared markdown display gateway for status/spec view
// (REQ-TUX4-004). Rich path: glamour with tui-token style. Plain path
// (non-TTY or NO_COLOR): byte-stable markdown passthrough — zero ANSI
// (REQ-TUX4-005). Render failures degrade to passthrough, never error the
// command (display-layer-only constraint: data must still reach the user).
func renderMarkdown(w io.Writer, md string) string {
	if !markdownRichEnabled((tui.OSEnv{}).NoColor(), writerIsTerminal(w)) {
		return md
	}
	out, err := glamourRender(md, resolveTheme())
	if err != nil {
		return md
	}
	return out
}
