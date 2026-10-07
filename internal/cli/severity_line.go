package cli

// severity_line.go — the single severity-prefix vocabulary for ad-hoc
// init/update output lines (card t1527 D4).
//
// Before this helper the init/update surface emitted a zoo of raw text
// prefixes — ✓ / · / note: / Note: / Tip: / advisory: / Advisory: /
// Warning: / warning: / hint: — each Fprintln'd with its own casing, some on
// stdout, some on stderr. This file owns the mapping from a severity to ONE
// glyph (resolved from the tui glyph SSOT, never a redeclared rune) and ONE
// line shape ("<glyph> <message>"), so a line's urgency is readable from its
// leading glyph alone:
//
//	sevOK    ✓  the step succeeded
//	sevWarn  !  needs attention, run continues
//	sevErr   ✗  a failure the operator must act on
//	sevNote  ·  informational
//	sevRun   ●  in progress
//
// Colour rides paintToken (Theme tokens; empty token under NO_COLOR renders
// plain), matching the update_tux.go styling gateway.

import (
	"fmt"
	"io"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// updateSeverity enumerates the ad-hoc line severities of the init/update
// surface. Doctor rows and progress steps have their own primitives
// (tui.CheckLine, tui.ProgressLine); this vocabulary is for the single lines
// between them.
type updateSeverity int

const (
	sevOK updateSeverity = iota
	sevWarn
	sevErr
	sevNote
	sevRun
)

// severityGlyph resolves the glyph for a severity from the tui.StatusIcon SSOT
// and paints it with the matching Theme token. Under NO_COLOR the token is
// empty and the bare glyph is returned (paintToken contract).
func severityGlyph(s updateSeverity, th tui.Theme) string {
	switch s {
	case sevOK:
		return paintToken(tui.StatusIcon("ok"), th.Success, false)
	case sevWarn:
		return paintToken(tui.StatusIcon("warn"), th.Warning, true)
	case sevErr:
		return paintToken(tui.StatusIcon("err"), th.Danger, true)
	case sevRun:
		return paintToken(tui.StatusIcon("run"), th.Accent, false)
	default: // sevNote
		return paintToken(tui.StatusIcon("info"), th.Faint, false)
	}
}

// emitSeverityLine writes one "<glyph> <message>" line. The message is
// fmt.Sprintf-formatted; no trailing prefix is added — the glyph IS the
// prefix. Errors from the writer are ignored by contract (every caller here is
// a best-effort presentation path that has nowhere to escalate a report
// failure).
func emitSeverityLine(w io.Writer, s updateSeverity, th tui.Theme, format string, args ...any) {
	_, _ = fmt.Fprintf(w, "%s %s\n", severityGlyph(s, th), fmt.Sprintf(format, args...))
}
