// Package tui — progress_line.go
//
// SPEC-V3R6-UPDATE-PROGRESS-001: Provides ProgressLine, a stateless helper
// that renders a single progress line and transitions it to a Done or Fail
// terminal state without leaving trailing residual characters.
//
// Defect background (Defect #5, 2026-05-23 audit): The legacy pattern
//
//	fmt.Fprintf(out, "  %s Working on long thing...", symProgress())   // progress
//	fmt.Fprintf(out, "\r  %s Working on long thing done\n", symSuccess()) // result
//
// uses a bare carriage return (\r) which only moves the cursor to column 0
// without erasing the line. When the result message is shorter than the
// progress message, residual characters from the progress text leak through
// at the end of the line (the observed "g..." / "n..." / "i..." artifacts).
//
// ProgressLine consolidates the progress -> result transition into a single
// abstraction. On a TTY-backed writer it prefixes the terminal message with
// the ANSI escape sequence "\r\033[2K" (carriage return + CSI Erase in Line
// mode 2 — erase the entire line). On a non-TTY writer it emits the progress
// and result messages on separate lines without any escape sequences, so
// piped output (CI logs, `2>&1 | cat`) stays plain-text readable.
package tui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// progressClearPrefix is the ANSI escape sequence used to clear the entire
// current terminal line before re-rendering. It combines a carriage return
// (move cursor to column 0) with CSI 2K (erase the entire line, including
// characters past the cursor's previous position). This is the only
// mechanism in the package that guarantees no trailing residual characters
// when a short message replaces a longer one. See ECMA-48 / VT100 / ANSI
// X3.64 §8.3.41 (Erase in Line) for the canonical specification.
const progressClearPrefix = "\r\x1b[2K"

// ProgressLineHandle represents an in-flight progress line. It must be
// constructed via ProgressLine and transitioned exactly once via Done or
// Fail. Calling Done or Fail a second time is a programmer error and will
// panic to surface the misuse early (see acceptance.md EC-1).
//
// Update may be called any number of times between construction and the
// terminal Done/Fail transition to re-render the in-flight progress line
// with a new message (REQ-UPR-007).
//
// ProgressLineHandle is stateless apart from the boolean done flag and is
// safe to construct and use within a single goroutine. It is not designed
// for concurrent use across goroutines.
//
// @MX:ANCHOR: [AUTO] ProgressLineHandle is the canonical progress-line
// abstraction; expected fan_in >= 22 across internal/cli/update.go after
// SPEC-V3R6-UPDATE-PROGRESS-001 M1 migration.
// @MX:REASON: [AUTO] Replaces the legacy `\r  %s` pair pattern in 22 call
// sites that previously leaked trailing characters when the result message
// was shorter than the progress message.
type ProgressLineHandle struct {
	out   io.Writer
	isTTY bool
	theme Theme
	done  bool
}

// ProgressLine begins rendering a progress line to out using the provided
// message and theme. If theme is nil, LightTheme is used (matching the
// stateless precedent of Spinner / Progress / Stepper in status.go).
//
// On a TTY-backed writer (out is an *os.File whose file descriptor is a
// terminal), the progress line is emitted in-place without a trailing
// newline, so a subsequent Done / Fail / Update call can rewrite it.
//
// On a non-TTY writer (out is a *bytes.Buffer or a redirected *os.File,
// such as in tests or piped output), the progress line is emitted with a
// trailing newline. The subsequent Done / Fail call then emits the result
// message on a separate line. This preserves readability when the output
// is captured to a log file or piped through a non-terminal-aware tool.
//
// The two-space indent matches the existing CLI layout precedent in
// internal/cli/update.go (which uses "  %s message" pairs).
func ProgressLine(out io.Writer, message string, theme *Theme) *ProgressLineHandle {
	t := LightTheme()
	if theme != nil {
		t = *theme
	}

	h := &ProgressLineHandle{
		out:   out,
		isTTY: isTerminalWriter(out),
		theme: t,
	}

	h.writeProgress(message)
	return h
}

// Done transitions the handle to the success terminal state, rendering
// the success symbol ("✓") followed by the result message. On a TTY this
// clears the entire current line before rendering, guaranteeing that no
// residual characters from the in-flight progress message leak through.
// On a non-TTY this emits the success line on a new line.
//
// Calling Done after the handle has already been transitioned (Done or
// Fail previously called) panics. This matches acceptance.md EC-1
// (Option B: panic on misuse to surface programmer error early).
func (h *ProgressLineHandle) Done(successMsg string) {
	h.assertNotDone()
	h.writeTerminal(h.symSuccess(), successMsg)
	h.done = true
}

// Fail transitions the handle to the error terminal state, rendering the
// error symbol ("✗") followed by the error message. Semantics match Done
// but with the danger-colored error symbol.
//
// Calling Fail after the handle has already been transitioned panics.
func (h *ProgressLineHandle) Fail(errMsg string) {
	h.assertNotDone()
	h.writeTerminal(h.symError(), errMsg)
	h.done = true
}

// Update re-renders the in-flight progress line with a new message
// (REQ-UPR-007). On a TTY it clears the current line and rewrites it in
// place. On a non-TTY it emits the new progress message on a new line.
//
// Update may be called any number of times before Done or Fail. Calling
// Update after Done or Fail panics.
func (h *ProgressLineHandle) Update(message string) {
	h.assertNotDone()
	if h.isTTY {
		// Clear the current line and re-emit the progress prefix in place.
		_, _ = fmt.Fprintf(h.out, "%s  %s %s", progressClearPrefix, h.symProgress(), message)
		return
	}
	// Non-TTY: a new line for each update so that progress history is
	// preserved in the captured output.
	_, _ = fmt.Fprintf(h.out, "  %s %s\n", h.symProgress(), message)
}

// writeProgress emits the initial in-flight progress message. On TTY it
// is emitted without a trailing newline so a subsequent Done/Fail/Update
// can rewrite the same line. On non-TTY it is emitted with a trailing
// newline so the progress message is preserved as its own log entry.
func (h *ProgressLineHandle) writeProgress(message string) {
	if h.isTTY {
		_, _ = fmt.Fprintf(h.out, "  %s %s", h.symProgress(), message)
		return
	}
	_, _ = fmt.Fprintf(h.out, "  %s %s\n", h.symProgress(), message)
}

// writeTerminal emits the terminal (Done or Fail) line. On TTY it prefixes
// the message with progressClearPrefix (`\r\033[2K`) to guarantee the full
// previous line is cleared. On non-TTY it emits a plain "  symbol message"
// line (no escape sequences, no carriage return).
func (h *ProgressLineHandle) writeTerminal(symbol, message string) {
	if h.isTTY {
		_, _ = fmt.Fprintf(h.out, "%s  %s %s\n", progressClearPrefix, symbol, message)
		return
	}
	_, _ = fmt.Fprintf(h.out, "  %s %s\n", symbol, message)
}

// assertNotDone panics if the handle has already been transitioned. The
// panic surfaces programmer error (e.g., Done() called twice) rather than
// silently double-rendering or no-op'ing, matching the "panic on misuse"
// design decision documented in acceptance.md EC-1.
func (h *ProgressLineHandle) assertNotDone() {
	if h.done {
		panic("tui: ProgressLineHandle terminal state already reached (Done/Fail already called)")
	}
}

// symProgress returns the in-flight progress glyph ("○") rendered in the
// dim (muted) theme color. Mirrors the cliMuted / symProgress() helper in
// internal/cli/update.go to preserve visual consistency.
func (h *ProgressLineHandle) symProgress() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.theme.Dim)).
		Render("○")
}

// symSuccess returns the success glyph ("✓") rendered in the theme's
// success color.
func (h *ProgressLineHandle) symSuccess() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.theme.Success)).
		Render("✓")
}

// symError returns the error glyph ("✗") rendered in the theme's danger
// color.
func (h *ProgressLineHandle) symError() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(h.theme.Danger)).
		Render("✗")
}

// isTerminalWriter reports whether out is a terminal-backed writer. The
// determination is conservative: only *os.File values whose file
// descriptor reports as a terminal are treated as TTYs. All other writers
// (bytes.Buffer, custom io.Writer implementations, pipes, redirected
// files) are treated as non-TTYs.
//
// This conservative policy is what guarantees REQ-UPR-003 (non-TTY
// fallback): when output is captured for tests, piped to another command,
// or redirected to a log file, no ANSI escape sequences are emitted.
//
// The dependency on mattn/go-isatty is already present in
// internal/cli/update.go (line 23), so adding it to internal/tui does not
// expand the module's external dependency surface. Layering is one-way
// (cli → tui → isatty); no cycle is introduced.
func isTerminalWriter(out io.Writer) bool {
	f, ok := out.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd())
}
