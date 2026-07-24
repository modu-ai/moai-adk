package printer_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// Characterization tests locking the spinner-residue contract of the live
// printer step/spinner handles (SPEC-CLI-TUX-INIT-UPDATE-001 M1, AC-TUXIU-006 /
// AC-TUXIU-007). On a TTY a finished step clears its in-flight line via the
// ANSI erase-in-line sequence and prints exactly one clean result line — no
// residual step marker ("○") or spinner frame ("⠋") survives the erase. On a
// non-TTY the split is preserved: a single plain result line, no ANSI escape.

const eraseLine = "\r\x1b[2K"

func lastEraseTail(s string) string {
	i := strings.LastIndex(s, eraseLine)
	if i < 0 {
		return s
	}
	return s[i+len(eraseLine):]
}

// AC-TUXIU-006 — TTY Step: cleared line + one clean ✓ result line, no residue.
func TestPrinterStepResidue_TTY_CleanResultLine(t *testing.T) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	p := printer.New(
		printer.WithWriters(out, errBuf),
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithAnimatedHandles(false),
	)

	step := p.Step("Deploy Templates: Deploying template files...")
	step.Complete("Deploy Templates complete")

	got := errBuf.String()
	if !strings.Contains(got, eraseLine) {
		t.Fatalf("expected erase-line sequence in TTY step output, got: %q", got)
	}
	tail := lastEraseTail(got)
	if n := strings.Count(tail, "\n"); n != 1 {
		t.Errorf("expected exactly 1 result line after erase, got %d: %q", n, tail)
	}
	if !strings.Contains(tail, string(tui.GlyphDone)) {
		t.Errorf("expected success glyph in cleared result line, got: %q", tail)
	}
	if strings.ContainsRune(tail, tui.GlyphSkip) {
		t.Errorf("residual step marker %q leaked past the erase: %q", string(tui.GlyphSkip), tail)
	}
	if strings.Contains(tail, "⠋") {
		t.Errorf("residual spinner-frame fragment leaked past the erase: %q", tail)
	}
}

// AC-TUXIU-006 — TTY Spinner: same clean-clear contract on the spinner handle.
func TestPrinterSpinnerResidue_TTY_CleanResultLine(t *testing.T) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	p := printer.New(
		printer.WithWriters(out, errBuf),
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithAnimatedHandles(false),
	)

	sp := p.Spinner("Syncing templates")
	sp.Done("Templates synced")

	tail := lastEraseTail(errBuf.String())
	if !strings.Contains(tail, string(tui.GlyphDone)) {
		t.Errorf("expected success glyph in cleared spinner result line, got: %q", tail)
	}
	if strings.ContainsRune(tail, tui.GlyphSkip) || strings.Contains(tail, "⠋") {
		t.Errorf("residual spinner fragment leaked past the erase: %q", tail)
	}
}

// AC-TUXIU-007 — non-TTY Step: no ANSI erase sequence, single ✓ result line.
func TestPrinterStepResidue_NonTTY_NoAnsi(t *testing.T) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	p := printer.New(
		printer.WithWriters(out, errBuf),
		printer.WithMode(printer.ModePlain),
	)

	step := p.Step("Deploy Templates: Deploying template files...")
	step.Complete("Deploy Templates complete")

	got := errBuf.String()
	if strings.Contains(got, "\x1b[") {
		t.Errorf("non-TTY step output must contain zero ANSI CSI sequences, got: %q", got)
	}
	if strings.Contains(got, "\r") {
		t.Errorf("non-TTY step output must contain zero carriage returns, got: %q", got)
	}
	if c := strings.Count(got, string(tui.GlyphDone)); c != 1 {
		t.Errorf("expected exactly one ✓ result line, got %d: %q", c, got)
	}
}
