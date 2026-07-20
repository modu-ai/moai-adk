// Package printer provides the single output gateway for all moai CLI
// commands (SPEC-CLI-TUX-V3-001 REQ-CTX-011..014).
//
// # Channel discipline (HARD)
//
// Data output goes to stdout; Info/Warn/Error/Success/Step/Spinner/Progress
// output goes to stderr (internal/cli/CLAUDE.md "Output streams" convention,
// REQ-CTX-012). The discipline is mode-independent: every render mode routes
// Data to stdout and everything else to stderr.
//
// # Render modes (selected once at construction, REQ-CTX-013)
//
//   - ModeTTY:   rich output — status markers coloured from internal/tui
//     string tokens, in-place erase-line updates for step/spinner/progress.
//   - ModePlain: non-TTY/CI output — no ANSI escape sequences at all,
//     one line per event.
//   - ModeJSON:  machine output — one JSON object per line on stderr for
//     events, JSON payloads on stdout for Data.
//
// Mode resolution: explicit WithMode override > isatty detection on the
// stderr writer. NO_COLOR (any non-empty value) forces unstyled output:
// a resolved ModeTTY degrades to ModePlain so that no ANSI escape sequence
// is emitted on either channel (REQ-CTX-014).
//
// # Style source
//
// All colours come from internal/tui string tokens (Theme fields). This
// package contains no hex colour literals (AC-CLI-TUI-013 succession,
// REQ-CTX-008).
package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/mattn/go-isatty"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// Mode is the render mode selected at construction time (REQ-CTX-013).
type Mode int

const (
	// ModePlain renders unstyled, escape-free, line-per-event output.
	// It is the zero value: the safest default for pipes and CI.
	ModePlain Mode = iota
	// ModeTTY renders rich output with tui-token colours and in-place
	// erase-line updates.
	ModeTTY
	// ModeJSON renders one-line JSON events (stderr) and JSON payloads
	// (stdout).
	ModeJSON
)

// Printer is the single output gateway for all CLI commands (REQ-CTX-011).
type Printer interface {
	// Info writes an informational status line to stderr.
	Info(format string, args ...any)
	// Warn writes a "Warning: ..." status line to stderr.
	Warn(format string, args ...any)
	// Error writes an "Error: ..." status line to stderr.
	Error(format string, args ...any)
	// Success writes a success status line to stderr.
	Success(format string, args ...any)
	// Data writes machine-consumable output to stdout. In ModeJSON the
	// value is marshalled as a single JSON line; in text modes strings are
	// written verbatim and other values via fmt.
	Data(v any) error
	// Step begins a long-running step feedback line on stderr.
	Step(name string) StepHandle
	// Spinner begins a spinner feedback line on stderr. Under
	// MOAI_REDUCED_MOTION or non-TTY modes the spinner degrades to
	// static lines (REQ-CTX-007 semantics delegated from internal/tui).
	Spinner(label string) SpinnerHandle
	// Progress begins a progress-bar feedback line on stderr.
	Progress(label string, total int) ProgressHandle
	// Mode reports the resolved render mode.
	Mode() Mode
}

// StepHandle is an in-flight step line. On a TTY it is updated in place
// with erase-line sequences; otherwise each event is its own plain line
// (tui.ProgressLine semantics). After Complete or Fail the handle is
// terminal: further calls are no-ops.
type StepHandle interface {
	Update(format string, args ...any)
	Complete(format string, args ...any)
	Fail(err error)
}

// SpinnerHandle is an in-flight spinner line. Terminal-once like StepHandle.
type SpinnerHandle interface {
	Update(label string)
	Done(format string, args ...any)
	Fail(format string, args ...any)
}

// ProgressHandle is an in-flight progress-bar line. Terminal-once like
// StepHandle.
type ProgressHandle interface {
	Set(current int)
	Done(format string, args ...any)
	Fail(format string, args ...any)
}

// Option configures New.
type Option func(*settings)

type settings struct {
	stdout       io.Writer
	stderr       io.Writer
	mode         *Mode
	theme        *tui.Theme
	noColor      *bool
	animate      *bool
	animInterval *time.Duration
}

// WithWriters overrides the stdout / stderr writers (default: os.Stdout,
// os.Stderr).
func WithWriters(stdout, stderr io.Writer) Option {
	return func(s *settings) {
		s.stdout = stdout
		s.stderr = stderr
	}
}

// WithMode sets the render mode explicitly, bypassing isatty detection.
// NO_COLOR still forces a ModeTTY override down to ModePlain (REQ-CTX-014).
func WithMode(m Mode) Option {
	return func(s *settings) { s.mode = &m }
}

// WithTheme injects the tui token source used for ModeTTY colouring
// (default: tui.ResolveOS() when ModeTTY is resolved).
func WithTheme(th tui.Theme) Option {
	return func(s *settings) { s.theme = &th }
}

// WithNoColor overrides NO_COLOR environment detection (tests).
func WithNoColor(no bool) Option {
	return func(s *settings) { s.noColor = &no }
}

// WithAnimatedHandles overrides the spinner-animation capability detection
// (SPEC-CLI-TUX-V3-002 REQ-TUX2-010). By default the frame-cycling animator
// runs only when the resolved mode is ModeTTY, motion is allowed, AND stderr
// is a real terminal — buffer/pipe writers keep the legacy static
// single-frame behavior so captured output stays deterministic. Reduced
// motion and non-TTY fallbacks (REQ-TUX2-011) still win over a true
// override: animation never runs when the handle is not in-place.
func WithAnimatedHandles(on bool) Option {
	return func(s *settings) { s.animate = &on }
}

// WithAnimationInterval overrides the spinner frame interval (tests).
// The default is the bubbles v2 MiniDot cadence.
func WithAnimationInterval(d time.Duration) Option {
	return func(s *settings) { s.animInterval = &d }
}

// New constructs a Printer. Without options it writes to os.Stdout /
// os.Stderr and resolves the mode from the environment.
func New(opts ...Option) Printer {
	s := &settings{stdout: os.Stdout, stderr: os.Stderr}
	for _, opt := range opts {
		opt(s)
	}

	mode := resolveMode(s)

	// Animation capability (REQ-TUX2-010): frame cycling requires a real
	// terminal on stderr; the explicit option overrides detection (tests).
	animate := mode == ModeTTY && isTTYFile(s.stderr)
	if s.animate != nil {
		animate = *s.animate
	}
	animInterval := spinner.MiniDot.FPS
	if s.animInterval != nil {
		animInterval = *s.animInterval
	}

	// Theme resolution is lazy: tokens are only needed for ModeTTY
	// colouring. tui.ResolveOS honours NO_COLOR / MOAI_THEME and may query
	// the terminal background only in auto mode (M1a background-detection
	// decision).
	th := tui.MonochromeTheme()
	if mode == ModeTTY {
		if s.theme != nil {
			th = *s.theme
		} else {
			th = tui.ResolveOS()
		}
	}

	stderr := s.stderr
	if mode == ModeTTY && isTTYFile(stderr) {
		// lipgloss v2 renders full-fidelity ANSI; downsampling for the
		// actual terminal capabilities is the writer's job (Charm v2
		// architecture). Only a real terminal writer is wrapped, so test
		// buffers observe the raw contract output.
		stderr = colorprofile.NewWriter(stderr, os.Environ())
	}

	return &printerImpl{
		stdout:       s.stdout,
		stderr:       stderr,
		mode:         mode,
		theme:        th,
		animate:      animate,
		animInterval: animInterval,
	}
}

// resolveMode applies the construction-time mode contract (REQ-CTX-013):
// explicit override > isatty detection on stderr; NO_COLOR forces a TTY
// resolution down to plain so no ANSI is emitted (REQ-CTX-014).
func resolveMode(s *settings) Mode {
	noColor := noColorEnv()
	if s.noColor != nil {
		noColor = *s.noColor
	}

	var mode Mode
	switch {
	case s.mode != nil:
		mode = *s.mode
	case isTTYFile(s.stderr):
		mode = ModeTTY
	default:
		mode = ModePlain
	}

	if mode == ModeTTY && noColor {
		return ModePlain
	}
	return mode
}

// noColorEnv reports whether NO_COLOR is set to any non-empty value
// (https://no-color.org/ — mirrors tui detect.go isEnvSet semantics,
// acceptance.md §C: NO_COLOR=0 counts as set).
func noColorEnv() bool {
	v, ok := os.LookupEnv("NO_COLOR")
	return ok && v != ""
}

// reducedMotion mirrors internal/tui's MOAI_REDUCED_MOTION semantics
// (non-empty, non-"0" value; AC-CLI-TUI-015). It degrades spinner and
// progress handles to static lines; Step erase-line updates are not
// motion-gated (REQ-CTX-007 binds spinner/progress only).
func reducedMotion() bool {
	v := os.Getenv("MOAI_REDUCED_MOTION")
	return v != "" && v != "0"
}

// isTTYFile reports whether w is a terminal-backed *os.File (conservative
// policy shared with tui.isTerminalWriter: buffers, pipes and redirected
// files are non-TTY).
func isTTYFile(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd())
}

// Rendering constants. All glyphs are in the AC-CLI-TUI-017 whitelist;
// eraseLine matches tui's progressClearPrefix (ECMA-48 Erase in Line 2).
const (
	eraseLine     = "\r\x1b[2K"
	stepMarker    = "○" // in-flight step marker (ConsoleReporter precedent)
	spinnerFrame  = "⠋" // frame 0, matches tui.Spinner
	spinnerStatic = "●" // reduced-motion static marker, matches tui.Spinner
)

// printerImpl is the concrete Printer.
type printerImpl struct {
	stdout       io.Writer
	stderr       io.Writer
	mode         Mode
	theme        tui.Theme
	animate      bool
	animInterval time.Duration
}

var _ Printer = (*printerImpl)(nil)

func (p *printerImpl) Mode() Mode { return p.mode }

func (p *printerImpl) Info(format string, args ...any) {
	p.status("info", p.theme.Info, tui.StatusIcon("info"), "", format, args...)
}

func (p *printerImpl) Warn(format string, args ...any) {
	p.status("warn", p.theme.Warning, tui.StatusIcon("warn"), "Warning: ", format, args...)
}

func (p *printerImpl) Error(format string, args ...any) {
	p.status("error", p.theme.Danger, tui.StatusIcon("err"), "Error: ", format, args...)
}

func (p *printerImpl) Success(format string, args ...any) {
	p.status("success", p.theme.Success, tui.StatusIcon("ok"), "", format, args...)
}

// status renders one status event: a JSON line in ModeJSON, otherwise a
// "<glyph> <prefix><msg>" stderr line with the glyph coloured from the
// theme token in ModeTTY.
func (p *printerImpl) status(level, token, glyph, prefix, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if p.mode == ModeJSON {
		p.event(jsonEvent{Level: level, Msg: msg})
		return
	}
	_, _ = fmt.Fprintf(p.stderr, "%s %s%s\n", p.paint(token, glyph), prefix, msg)
}

// paint colours s with the given tui token in ModeTTY. Empty tokens
// (MonochromeTheme) and non-TTY modes render s unchanged, so plain/json
// output never carries colour sequences.
func (p *printerImpl) paint(token, s string) string {
	if p.mode != ModeTTY || token == "" {
		return s
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(token)).Render(s)
}

// Data writes machine-consumable output to stdout (REQ-CTX-012).
func (p *printerImpl) Data(v any) error {
	if p.mode == ModeJSON {
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("printer: marshal data: %w", err)
		}
		_, err = fmt.Fprintln(p.stdout, string(b))
		return err
	}
	_, err := fmt.Fprintln(p.stdout, v)
	return err
}

// jsonEvent is the one-line JSON event schema for ModeJSON stderr output.
type jsonEvent struct {
	Level   string `json:"level"`
	Msg     string `json:"msg,omitempty"`
	Name    string `json:"name,omitempty"`
	Event   string `json:"event,omitempty"`
	Current int    `json:"current,omitempty"`
	Total   int    `json:"total,omitempty"`
}

func (p *printerImpl) event(ev jsonEvent) {
	b, err := json.Marshal(ev)
	if err != nil {
		return // static schema: cannot fail in practice
	}
	_, _ = fmt.Fprintln(p.stderr, string(b))
}

// lineState is the shared in-flight feedback line machinery
// (tui.ProgressLine semantics: TTY in-place erase-line updates, non-TTY
// one plain line per event). Terminal-once: after finish, the handle is
// inert (no panic — output gateways must not crash the CLI on misuse).
type lineState struct {
	p       *printerImpl
	inPlace bool
	done    bool
}

func (l *lineState) start(body string) {
	if l.inPlace {
		_, _ = fmt.Fprintf(l.p.stderr, "  %s", body)
		return
	}
	_, _ = fmt.Fprintf(l.p.stderr, "  %s\n", body)
}

func (l *lineState) update(body string) {
	if l.done {
		return
	}
	if l.inPlace {
		_, _ = fmt.Fprintf(l.p.stderr, "%s  %s", eraseLine, body)
		return
	}
	_, _ = fmt.Fprintf(l.p.stderr, "  %s\n", body)
}

func (l *lineState) finish(body string) {
	if l.done {
		return
	}
	l.done = true
	if l.inPlace {
		_, _ = fmt.Fprintf(l.p.stderr, "%s  %s\n", eraseLine, body)
		return
	}
	_, _ = fmt.Fprintf(l.p.stderr, "  %s\n", body)
}

// Step begins a long-running step feedback line (REQ-CTX-011).
func (p *printerImpl) Step(name string) StepHandle {
	h := &stepHandle{line: lineState{p: p, inPlace: p.mode == ModeTTY}}
	if p.mode == ModeJSON {
		p.event(jsonEvent{Level: "step", Event: "start", Name: name})
		return h
	}
	h.line.start(p.paint(p.theme.Dim, stepMarker) + " " + name)
	return h
}

type stepHandle struct {
	line lineState
}

func (h *stepHandle) Update(format string, args ...any) {
	p := h.line.p
	msg := fmt.Sprintf(format, args...)
	if p.mode == ModeJSON {
		if h.line.done {
			return
		}
		p.event(jsonEvent{Level: "step", Event: "update", Msg: msg})
		return
	}
	h.line.update(p.paint(p.theme.Dim, stepMarker) + " " + msg)
}

func (h *stepHandle) Complete(format string, args ...any) {
	p := h.line.p
	msg := fmt.Sprintf(format, args...)
	if p.mode == ModeJSON {
		if h.line.done {
			return
		}
		h.line.done = true
		p.event(jsonEvent{Level: "step", Event: "complete", Msg: msg})
		return
	}
	h.line.finish(p.paint(p.theme.Success, tui.StatusIcon("ok")) + " " + msg)
}

func (h *stepHandle) Fail(err error) {
	p := h.line.p
	msg := "Error"
	if err != nil {
		msg = err.Error()
	}
	if p.mode == ModeJSON {
		if h.line.done {
			return
		}
		h.line.done = true
		p.event(jsonEvent{Level: "step", Event: "fail", Msg: msg})
		return
	}
	h.line.finish(p.paint(p.theme.Danger, tui.StatusIcon("err")) + " Error: " + msg)
}

// Spinner begins a spinner feedback line. Reduced motion and non-TTY
// modes degrade to static, line-per-event output (REQ-CTX-007). On a
// motion-allowed real terminal the handle cycles the bubbles v2 MiniDot
// frames in a background animator (SPEC-CLI-TUX-V3-002 REQ-TUX2-010);
// Done/Fail stops the animator before the final line — no frame is ever
// emitted after the terminal event.
func (p *printerImpl) Spinner(label string) SpinnerHandle {
	h := &spinnerHandle{
		line:  lineState{p: p, inPlace: p.mode == ModeTTY && !reducedMotion()},
		label: label,
	}
	if p.mode == ModeJSON {
		p.event(jsonEvent{Level: "spinner", Event: "start", Name: label})
		return h
	}
	h.line.start(p.spinnerMarker() + " " + label)
	// The animator requires an in-place line (TTY + motion allowed) AND the
	// printer-level animation capability (real terminal or test override).
	if h.line.inPlace && p.animate {
		h.startAnimator(spinner.MiniDot.Frames, p.animInterval)
	}
	return h
}

// spinnerMarker returns the in-flight spinner glyph: the animated frame
// character on a motion-allowed TTY, the static dot otherwise.
func (p *printerImpl) spinnerMarker() string {
	if p.mode == ModeTTY && !reducedMotion() {
		return p.paint(p.theme.Accent, spinnerFrame)
	}
	return p.paint(p.theme.Accent, spinnerStatic)
}

// spinnerHandle is the in-flight spinner line. The mutex serialises the
// animator goroutine against Update/Done/Fail callers; the animator is
// stopped (and joined) before the finish line is printed, so Complete/Fail
// is a hard emission barrier (acceptance.md §C goroutine-lifecycle edge).
type spinnerHandle struct {
	line     lineState
	mu       sync.Mutex
	label    string
	frameIdx int
	frames   []string
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// startAnimator launches the frame-cycling goroutine (bubbles v2 backend).
func (h *spinnerHandle) startAnimator(frames []string, interval time.Duration) {
	if interval <= 0 {
		return
	}
	h.frames = frames
	h.stopCh = make(chan struct{})
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-h.stopCh:
				return
			case <-ticker.C:
				h.mu.Lock()
				if h.line.done {
					h.mu.Unlock()
					return
				}
				h.frameIdx = (h.frameIdx + 1) % len(h.frames)
				p := h.line.p
				h.line.update(p.paint(p.theme.Accent, h.frames[h.frameIdx]) + " " + h.label)
				h.mu.Unlock()
			}
		}
	}()
}

// stopAnimator signals and joins the animator goroutine. Must be called
// WITHOUT h.mu held (the animator acquires it per frame).
func (h *spinnerHandle) stopAnimator() {
	if h.stopCh == nil {
		return
	}
	close(h.stopCh)
	h.wg.Wait()
	h.stopCh = nil
}

func (h *spinnerHandle) Update(label string) {
	p := h.line.p
	if p.mode == ModeJSON {
		if h.line.done {
			return
		}
		p.event(jsonEvent{Level: "spinner", Event: "update", Msg: label})
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.label = label
	marker := p.spinnerMarker()
	if h.frames != nil {
		marker = p.paint(p.theme.Accent, h.frames[h.frameIdx])
	}
	h.line.update(marker + " " + label)
}

func (h *spinnerHandle) Done(format string, args ...any) {
	p := h.line.p
	msg := fmt.Sprintf(format, args...)
	if p.mode == ModeJSON {
		if h.line.done {
			return
		}
		h.line.done = true
		p.event(jsonEvent{Level: "spinner", Event: "done", Msg: msg})
		return
	}
	h.stopAnimator()
	h.mu.Lock()
	defer h.mu.Unlock()
	h.line.finish(p.paint(p.theme.Success, tui.StatusIcon("ok")) + " " + msg)
}

func (h *spinnerHandle) Fail(format string, args ...any) {
	p := h.line.p
	msg := fmt.Sprintf(format, args...)
	if p.mode == ModeJSON {
		if h.line.done {
			return
		}
		h.line.done = true
		p.event(jsonEvent{Level: "spinner", Event: "fail", Msg: msg})
		return
	}
	h.stopAnimator()
	h.mu.Lock()
	defer h.mu.Unlock()
	h.line.finish(p.paint(p.theme.Danger, tui.StatusIcon("err")) + " " + msg)
}

// Progress begins a progress-bar feedback line. The bar itself is
// rendered by tui.Progress, which owns the reduced-motion (static filled
// bar) and monochrome (no ANSI) semantics.
func (p *printerImpl) Progress(label string, total int) ProgressHandle {
	h := &progressHandle{
		line:  lineState{p: p, inPlace: p.mode == ModeTTY && !reducedMotion()},
		label: label,
		total: total,
	}
	if p.mode == ModeJSON {
		p.event(jsonEvent{Level: "progress", Event: "start", Name: label, Total: total})
		return h
	}
	h.line.start(h.body(0))
	return h
}

type progressHandle struct {
	line  lineState
	label string
	total int
}

// body composes "<label> <bar> <current>/<total>".
func (h *progressHandle) body(current int) string {
	th := h.line.p.theme
	bar := tui.Progress(current, h.total, tui.ProgressOpts{Theme: &th})
	return fmt.Sprintf("%s %s %d/%d", h.label, bar, current, h.total)
}

func (h *progressHandle) Set(current int) {
	p := h.line.p
	if p.mode == ModeJSON {
		if h.line.done {
			return
		}
		p.event(jsonEvent{Level: "progress", Event: "set", Current: current, Total: h.total})
		return
	}
	h.line.update(h.body(current))
}

func (h *progressHandle) Done(format string, args ...any) {
	p := h.line.p
	msg := fmt.Sprintf(format, args...)
	if p.mode == ModeJSON {
		if h.line.done {
			return
		}
		h.line.done = true
		p.event(jsonEvent{Level: "progress", Event: "done", Msg: msg})
		return
	}
	h.line.finish(p.paint(p.theme.Success, tui.StatusIcon("ok")) + " " + msg)
}

func (h *progressHandle) Fail(format string, args ...any) {
	p := h.line.p
	msg := fmt.Sprintf(format, args...)
	if p.mode == ModeJSON {
		if h.line.done {
			return
		}
		h.line.done = true
		p.event(jsonEvent{Level: "progress", Event: "fail", Msg: msg})
		return
	}
	h.line.finish(p.paint(p.theme.Danger, tui.StatusIcon("err")) + " " + msg)
}
