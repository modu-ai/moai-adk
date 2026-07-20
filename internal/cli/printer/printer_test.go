package printer_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// newBufPrinter constructs a printer writing into fresh out/err buffers.
func newBufPrinter(opts ...printer.Option) (printer.Printer, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	all := append([]printer.Option{printer.WithWriters(out, errBuf)}, opts...)
	return printer.New(all...), out, errBuf
}

// TestInterfaceContract verifies the Printer method set (AC-CTX-011):
// Info/Warn/Error/Success/Data/Step plus spinner and progress handles.
func TestInterfaceContract(t *testing.T) {
	p, _, _ := newBufPrinter(printer.WithMode(printer.ModePlain))

	p.Info("info %d", 1)
	p.Warn("warn %d", 2)
	p.Error("error %d", 3)
	p.Success("success %d", 4)
	if err := p.Data("payload"); err != nil {
		t.Fatalf("Data() error = %v, want nil", err)
	}

	step := p.Step("step-name")
	if step == nil {
		t.Fatal("Step() returned nil handle")
	}
	step.Update("working")
	step.Complete("done")

	sp := p.Spinner("spin-label")
	if sp == nil {
		t.Fatal("Spinner() returned nil handle")
	}
	sp.Update("spinning")
	sp.Done("spun")

	pr := p.Progress("copy", 10)
	if pr == nil {
		t.Fatal("Progress() returned nil handle")
	}
	pr.Set(5)
	pr.Done("copied")
}

// TestChannelDiscipline asserts the HARD channel contract (AC-CTX-012):
// Data goes to the stdout writer ONLY; Info/Warn/Error/Success/Step and
// the spinner/progress handles go to the stderr writer ONLY.
func TestChannelDiscipline(t *testing.T) {
	p, out, errBuf := newBufPrinter(printer.WithMode(printer.ModePlain))

	p.Info("info-msg")
	p.Warn("warn-msg")
	p.Error("error-msg")
	p.Success("success-msg")
	step := p.Step("step-msg")
	step.Update("step-update")
	step.Complete("step-complete")
	sp := p.Spinner("spinner-msg")
	sp.Done("spinner-done")
	pr := p.Progress("progress-msg", 4)
	pr.Set(2)
	pr.Done("progress-done")

	if err := p.Data("data-payload"); err != nil {
		t.Fatalf("Data() error = %v", err)
	}

	// stdout carries the data payload and nothing else.
	if got, want := out.String(), "data-payload\n"; got != want {
		t.Errorf("stdout = %q, want %q (Data only)", got, want)
	}

	// stderr carries every status message.
	errOut := errBuf.String()
	for _, want := range []string{
		"info-msg", "warn-msg", "error-msg", "success-msg",
		"step-msg", "step-update", "step-complete",
		"spinner-msg", "spinner-done",
		"progress-msg", "progress-done",
	} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr missing %q; stderr = %q", want, errOut)
		}
	}
	if strings.Contains(errOut, "data-payload") {
		t.Errorf("stderr must not carry Data output; stderr = %q", errOut)
	}
}

// TestChannelDisciplineJSONMode asserts the same channel split in ModeJSON.
func TestChannelDisciplineJSONMode(t *testing.T) {
	p, out, errBuf := newBufPrinter(printer.WithMode(printer.ModeJSON))

	p.Warn("warn-msg")
	if err := p.Data(map[string]int{"a": 1}); err != nil {
		t.Fatalf("Data() error = %v", err)
	}

	if got, want := strings.TrimSpace(out.String()), `{"a":1}`; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if !strings.Contains(errBuf.String(), "warn-msg") {
		t.Errorf("stderr missing warn event; stderr = %q", errBuf.String())
	}
	if strings.Contains(out.String(), "warn-msg") {
		t.Errorf("stdout must not carry events; stdout = %q", out.String())
	}
}

// TestModeResolution verifies construction-time mode selection
// (REQ-CTX-013): explicit override > isatty detection, and NO_COLOR
// forcing unstyled output.
func TestModeResolution(t *testing.T) {
	t.Run("non-tty writers default to plain", func(t *testing.T) {
		p, _, _ := newBufPrinter(printer.WithNoColor(false))
		if got := p.Mode(); got != printer.ModePlain {
			t.Errorf("Mode() = %v, want ModePlain", got)
		}
	})

	t.Run("explicit mode overrides detection", func(t *testing.T) {
		p, _, _ := newBufPrinter(printer.WithMode(printer.ModeJSON))
		if got := p.Mode(); got != printer.ModeJSON {
			t.Errorf("Mode() = %v, want ModeJSON", got)
		}
	})

	t.Run("explicit tty honored without NO_COLOR", func(t *testing.T) {
		p, _, _ := newBufPrinter(
			printer.WithMode(printer.ModeTTY),
			printer.WithNoColor(false),
			printer.WithTheme(tui.MonochromeTheme()),
		)
		if got := p.Mode(); got != printer.ModeTTY {
			t.Errorf("Mode() = %v, want ModeTTY", got)
		}
	})

	t.Run("NO_COLOR forces tty down to plain", func(t *testing.T) {
		p, _, _ := newBufPrinter(
			printer.WithMode(printer.ModeTTY),
			printer.WithNoColor(true),
		)
		if got := p.Mode(); got != printer.ModePlain {
			t.Errorf("Mode() = %v, want ModePlain (NO_COLOR)", got)
		}
	})

	t.Run("NO_COLOR keeps json as json", func(t *testing.T) {
		p, _, _ := newBufPrinter(
			printer.WithMode(printer.ModeJSON),
			printer.WithNoColor(true),
		)
		if got := p.Mode(); got != printer.ModeJSON {
			t.Errorf("Mode() = %v, want ModeJSON", got)
		}
	})

	t.Run("NO_COLOR env detected when not overridden", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		p, _, _ := newBufPrinter(printer.WithMode(printer.ModeTTY))
		if got := p.Mode(); got != printer.ModePlain {
			t.Errorf("Mode() = %v, want ModePlain (NO_COLOR env)", got)
		}
	})
}

// TestModeTTYStyledFromTokens verifies that ModeTTY colours status markers
// from injected tui tokens (AC-CTX-013): a coloured theme produces ANSI,
// the monochrome theme (empty tokens) produces none.
func TestModeTTYStyledFromTokens(t *testing.T) {
	t.Run("dark theme tokens emit ANSI colour", func(t *testing.T) {
		p, _, errBuf := newBufPrinter(
			printer.WithMode(printer.ModeTTY),
			printer.WithNoColor(false),
			printer.WithTheme(tui.DarkTheme()),
		)
		p.Success("styled")
		got := errBuf.String()
		if !strings.Contains(got, "\x1b[") {
			t.Errorf("expected ANSI colour from theme tokens, got %q", got)
		}
		if !strings.Contains(got, "✓") || !strings.Contains(got, "styled") {
			t.Errorf("expected success glyph + message, got %q", got)
		}
	})

	t.Run("monochrome tokens emit no colour", func(t *testing.T) {
		p, _, errBuf := newBufPrinter(
			printer.WithMode(printer.ModeTTY),
			printer.WithNoColor(false),
			printer.WithTheme(tui.MonochromeTheme()),
		)
		p.Success("plain-glyph")
		got := errBuf.String()
		if strings.Contains(got, "\x1b[38") || strings.Contains(got, "\x1b[3") {
			t.Errorf("monochrome theme must not emit colour SGR, got %q", got)
		}
		if !strings.Contains(got, "✓ plain-glyph") {
			t.Errorf("expected plain glyph line, got %q", got)
		}
	})
}

// TestModeTTYStepEraseLine verifies tui.ProgressLine-compatible semantics
// on a TTY: in-place erase-line updates (AC-CTX-013).
func TestModeTTYStepEraseLine(t *testing.T) {
	p, _, errBuf := newBufPrinter(
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithTheme(tui.MonochromeTheme()),
	)

	h := p.Step("Detect")
	if got := errBuf.String(); strings.HasSuffix(got, "\n") {
		t.Errorf("TTY step start must not end with newline, got %q", got)
	}
	if !strings.Contains(errBuf.String(), "○ Detect") {
		t.Errorf("step start missing in-flight marker, got %q", errBuf.String())
	}

	h.Update("scanning")
	if !strings.Contains(errBuf.String(), "\r\x1b[2K") {
		t.Errorf("TTY step update must erase the line, got %q", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "○ scanning") {
		t.Errorf("step update missing message, got %q", errBuf.String())
	}

	h.Complete("Detect: completed")
	got := errBuf.String()
	if !strings.Contains(got, "✓ Detect: completed") || !strings.HasSuffix(got, "\n") {
		t.Errorf("step complete must render success glyph + newline, got %q", got)
	}

	// Terminal-once: a second Complete/Fail is a silent no-op.
	before := errBuf.String()
	h.Complete("again")
	h.Fail(errors.New("late"))
	h.Update("late-update")
	if errBuf.String() != before {
		t.Errorf("terminal handle must be inert, extra output %q",
			strings.TrimPrefix(errBuf.String(), before))
	}
}

// TestModePlainRendering verifies the plain-mode line contract
// (AC-CTX-013): fixed textual prefixes, one line per event, no escapes.
func TestModePlainRendering(t *testing.T) {
	p, _, errBuf := newBufPrinter(printer.WithMode(printer.ModePlain))

	p.Info("hello")
	p.Warn("disk low")
	p.Error("boom")
	p.Success("ok")

	want := "· hello\n" +
		"! Warning: disk low\n" +
		"✗ Error: boom\n" +
		"✓ ok\n"
	if got := errBuf.String(); got != want {
		t.Errorf("plain rendering:\n got %q\nwant %q", got, want)
	}

	errBuf.Reset()
	h := p.Step("Deploy")
	h.Update("uploading")
	h.Complete("Deploy: completed")
	got := errBuf.String()
	wantStep := "  ○ Deploy\n  ○ uploading\n  ✓ Deploy: completed\n"
	if got != wantStep {
		t.Errorf("plain step:\n got %q\nwant %q", got, wantStep)
	}
	if strings.Contains(got, "\r") || strings.Contains(got, "\x1b") {
		t.Errorf("plain mode must not emit CR or escapes, got %q", got)
	}
}

// TestModeJSONRendering verifies the json-mode event contract (AC-CTX-013).
func TestModeJSONRendering(t *testing.T) {
	p, out, errBuf := newBufPrinter(printer.WithMode(printer.ModeJSON))

	p.Info("i")
	p.Warn("w %d", 1)
	p.Error("e")
	p.Success("s")
	h := p.Step("Deploy")
	h.Update("uploading")
	h.Complete("done")
	h2 := p.Step("Fails")
	h2.Fail(errors.New("bad"))

	lines := strings.Split(strings.TrimSpace(errBuf.String()), "\n")
	if len(lines) != 9 {
		t.Fatalf("expected 9 JSON event lines, got %d: %q", len(lines), lines)
	}

	type event struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
		Name  string `json:"name"`
		Event string `json:"event"`
	}
	var evs []event
	for i, ln := range lines {
		var ev event
		if err := json.Unmarshal([]byte(ln), &ev); err != nil {
			t.Fatalf("line %d not JSON: %q (%v)", i, ln, err)
		}
		evs = append(evs, ev)
	}

	checks := []struct {
		level, msg, name, evt string
	}{
		{"info", "i", "", ""},
		{"warn", "w 1", "", ""},
		{"error", "e", "", ""},
		{"success", "s", "", ""},
		{"step", "", "Deploy", "start"},
		{"step", "uploading", "", "update"},
		{"step", "done", "", "complete"},
		{"step", "", "Fails", "start"},
		{"step", "bad", "", "fail"},
	}
	for i, want := range checks[:len(evs)] {
		if evs[i].Level != want.level || evs[i].Msg != want.msg ||
			evs[i].Name != want.name || evs[i].Event != want.evt {
			t.Errorf("event[%d] = %+v, want %+v", i, evs[i], want)
		}
	}

	// Data payload is a single JSON line on stdout.
	if err := p.Data(map[string]string{"k": "v"}); err != nil {
		t.Fatalf("Data() error = %v", err)
	}
	if got, want := strings.TrimSpace(out.String()), `{"k":"v"}`; got != want {
		t.Errorf("json Data = %q, want %q", got, want)
	}

	// Unmarshalable payloads surface an error.
	if err := p.Data(func() {}); err == nil {
		t.Error("Data(func) expected marshal error, got nil")
	}
}

// TestNoANSIPlainAndJSONModes asserts REQ-CTX-014: plain and json modes
// emit zero ANSI escape sequences on either channel across every method.
func TestNoANSIPlainAndJSONModes(t *testing.T) {
	for _, mode := range []printer.Mode{printer.ModePlain, printer.ModeJSON} {
		p, out, errBuf := newBufPrinter(printer.WithMode(mode))

		p.Info("i")
		p.Warn("w")
		p.Error("e")
		p.Success("s")
		_ = p.Data("d")
		h := p.Step("st")
		h.Update("u")
		h.Complete("c")
		sp := p.Spinner("spin")
		sp.Update("spin2")
		sp.Done("spun")
		pr := p.Progress("prog", 3)
		pr.Set(1)
		pr.Done("progged")

		for name, buf := range map[string]*bytes.Buffer{"stdout": out, "stderr": errBuf} {
			if strings.Contains(buf.String(), "\x1b") {
				t.Errorf("mode %v: %s contains ANSI escape: %q", mode, name, buf.String())
			}
		}
	}
}

// TestPlainNoEscapeUnderNoColor asserts REQ-CTX-014 for the NO_COLOR path:
// even an explicit TTY request degrades to escape-free output.
func TestPlainNoEscapeUnderNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	p, out, errBuf := newBufPrinter(printer.WithMode(printer.ModeTTY))
	if got := p.Mode(); got != printer.ModePlain {
		t.Fatalf("Mode() = %v, want ModePlain under NO_COLOR", got)
	}

	p.Warn("careful")
	p.Success("fine")
	h := p.Step("work")
	h.Complete("worked")
	sp := p.Spinner("spin")
	sp.Done("spun")
	pr := p.Progress("prog", 2)
	pr.Set(1)
	pr.Done("progged")
	_ = p.Data("payload")

	for name, buf := range map[string]*bytes.Buffer{"stdout": out, "stderr": errBuf} {
		if strings.Contains(buf.String(), "\x1b") {
			t.Errorf("NO_COLOR: %s contains ANSI escape: %q", name, buf.String())
		}
	}
}

// TestSpinnerReducedMotionStatic verifies MOAI_REDUCED_MOTION degrades the
// TTY spinner and progress bar to static lines (REQ-CTX-007 semantics).
func TestSpinnerReducedMotionStatic(t *testing.T) {
	t.Setenv("MOAI_REDUCED_MOTION", "1")

	p, _, errBuf := newBufPrinter(
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithTheme(tui.MonochromeTheme()),
	)

	sp := p.Spinner("loading")
	got := errBuf.String()
	if !strings.Contains(got, "●") {
		t.Errorf("reduced-motion spinner must use static marker, got %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("reduced-motion spinner start must be a full line, got %q", got)
	}
	sp.Done("loaded")
	if strings.Contains(errBuf.String(), "\x1b[2K") {
		t.Errorf("reduced motion must not erase lines, got %q", errBuf.String())
	}

	errBuf.Reset()
	pr := p.Progress("copy", 10)
	pr.Set(3)
	got = errBuf.String()
	if !strings.Contains(got, "█") || strings.Contains(got, "░") {
		t.Errorf("reduced-motion bar must render fully filled, got %q", got)
	}
	if strings.Contains(got, "\x1b[2K") {
		t.Errorf("reduced motion must not erase lines, got %q", got)
	}
	pr.Done("copied")
}

// TestSpinnerTTYAnimatedFrame verifies the animated frame + erase-line
// updates when motion is allowed.
func TestSpinnerTTYAnimatedFrame(t *testing.T) {
	t.Setenv("MOAI_REDUCED_MOTION", "")

	p, _, errBuf := newBufPrinter(
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithTheme(tui.MonochromeTheme()),
	)

	sp := p.Spinner("loading")
	got := errBuf.String()
	if !strings.Contains(got, "⠋ loading") {
		t.Errorf("TTY spinner must render frame glyph, got %q", got)
	}
	if strings.HasSuffix(got, "\n") {
		t.Errorf("TTY spinner start must stay in place, got %q", got)
	}

	sp.Update("still loading")
	if !strings.Contains(errBuf.String(), "\r\x1b[2K") {
		t.Errorf("TTY spinner update must erase line, got %q", errBuf.String())
	}

	sp.Done("loaded")
	if !strings.Contains(errBuf.String(), "✓ loaded") {
		t.Errorf("spinner Done must render success glyph, got %q", errBuf.String())
	}
}

// TestProgressBarPlain verifies plain-mode progress rendering.
func TestProgressBarPlain(t *testing.T) {
	p, _, errBuf := newBufPrinter(printer.WithMode(printer.ModePlain))

	pr := p.Progress("copy", 4)
	pr.Set(2)
	got := errBuf.String()
	if !strings.Contains(got, "copy") || !strings.Contains(got, "2/4") {
		t.Errorf("progress line missing label or counter, got %q", got)
	}
	pr.Done("copied")
	if !strings.Contains(errBuf.String(), "✓ copied") {
		t.Errorf("progress Done must render success line, got %q", errBuf.String())
	}

	errBuf.Reset()
	pr2 := p.Progress("fetch", 2)
	pr2.Fail("network down")
	if !strings.Contains(errBuf.String(), "✗ network down") {
		t.Errorf("progress Fail must render error line, got %q", errBuf.String())
	}
}

// TestModeJSONSpinnerProgressEvents verifies spinner/progress event
// emission and terminal-once semantics in ModeJSON.
func TestModeJSONSpinnerProgressEvents(t *testing.T) {
	p, _, errBuf := newBufPrinter(printer.WithMode(printer.ModeJSON))

	sp := p.Spinner("load")
	sp.Update("loading")
	sp.Done("loaded")
	sp2 := p.Spinner("load2")
	sp2.Fail("broke")

	pr := p.Progress("copy", 5)
	pr.Set(2)
	pr.Done("copied")
	pr2 := p.Progress("fetch", 3)
	pr2.Fail("net down")

	lines := strings.Split(strings.TrimSpace(errBuf.String()), "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 event lines, got %d: %q", len(lines), lines)
	}
	for i, ln := range lines {
		if !json.Valid([]byte(ln)) {
			t.Errorf("line %d not JSON: %q", i, ln)
		}
	}
	joined := errBuf.String()
	for _, want := range []string{
		`"level":"spinner"`, `"event":"done"`, `"event":"fail"`,
		`"level":"progress"`, `"event":"set"`, `"current":2`, `"total":5`,
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("events missing %q; got %q", want, joined)
		}
	}

	// Terminal-once in json mode: further calls emit nothing.
	before := errBuf.String()
	sp.Update("late")
	sp.Done("late")
	sp.Fail("late")
	pr.Set(9)
	pr.Done("late")
	pr.Fail("late")
	if errBuf.String() != before {
		t.Errorf("terminal handles must be inert in json mode, extra %q",
			strings.TrimPrefix(errBuf.String(), before))
	}
}

// TestModeJSONStepTerminalOnce verifies step terminal-once in ModeJSON.
func TestModeJSONStepTerminalOnce(t *testing.T) {
	p, _, errBuf := newBufPrinter(printer.WithMode(printer.ModeJSON))

	h := p.Step("Deploy")
	h.Complete("done")
	before := errBuf.String()
	h.Update("late")
	h.Complete("late")
	h.Fail(errors.New("late"))
	if errBuf.String() != before {
		t.Errorf("terminal step must be inert in json mode, extra %q",
			strings.TrimPrefix(errBuf.String(), before))
	}

	// Fail path with nil error falls back to a generic message.
	h2 := p.Step("Nil")
	h2.Fail(nil)
	if !strings.Contains(errBuf.String(), `"event":"fail"`) {
		t.Errorf("expected fail event, got %q", errBuf.String())
	}
}

// TestSpinnerFailPlain verifies the plain-mode spinner failure line.
func TestSpinnerFailPlain(t *testing.T) {
	p, _, errBuf := newBufPrinter(printer.WithMode(printer.ModePlain))

	sp := p.Spinner("connect")
	sp.Fail("refused: %s", "ECONNREFUSED")
	got := errBuf.String()
	if !strings.Contains(got, "✗ refused: ECONNREFUSED") {
		t.Errorf("spinner Fail line missing, got %q", got)
	}

	// Step Fail with nil error in plain mode.
	h := p.Step("nilstep")
	h.Fail(nil)
	if !strings.Contains(errBuf.String(), "✗ Error: Error") {
		t.Errorf("nil-error step Fail must render generic error, got %q", errBuf.String())
	}
}

// TestDataTextMode verifies text-mode Data rendering on stdout.
func TestDataTextMode(t *testing.T) {
	p, out, _ := newBufPrinter(printer.WithMode(printer.ModePlain))

	if err := p.Data("raw-line"); err != nil {
		t.Fatalf("Data(string) error = %v", err)
	}
	if got, want := out.String(), "raw-line\n"; got != want {
		t.Errorf("Data(string) = %q, want %q", got, want)
	}

	out.Reset()
	if err := p.Data(42); err != nil {
		t.Fatalf("Data(int) error = %v", err)
	}
	if got, want := out.String(), "42\n"; got != want {
		t.Errorf("Data(int) = %q, want %q", got, want)
	}
}
