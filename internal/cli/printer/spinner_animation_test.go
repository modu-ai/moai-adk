package printer_test

// SPEC-CLI-TUX-V3-002 M2d — animated spinner + live progress contracts
// (REQ-TUX2-010/011, AC-TUX2-010/011).
//
// The Spinner handle gains a bubbles v2 frame-cycling backend: on a
// motion-allowed TTY it emits multiple animation frames over time; every
// fallback surface (non-TTY, NO_COLOR, MOAI_REDUCED_MOTION) stays a static
// single frame with zero ANSI escapes. Complete/Fail stops the animator —
// no frame emission after the terminal event (goroutine-lifecycle edge case;
// -race required).

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// syncBuffer is a mutex-guarded bytes.Buffer safe for concurrent writes from
// the animator goroutine and reads from the test goroutine.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// miniDotFrames are the bubbles v2 MiniDot frames the animated backend
// cycles through (frame 0 matches the legacy static glyph).
var miniDotFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func countDistinctFrames(s string) int {
	distinct := 0
	for _, f := range miniDotFrames {
		if strings.Contains(s, f) {
			distinct++
		}
	}
	return distinct
}

// TestSpinnerAnimated_MultiFrameEmission asserts the TTY+motion spinner
// emits MULTIPLE distinct animation frames over time (AC-TUX2-010) and stops
// emitting after Done (no frames after the terminal event).
func TestSpinnerAnimated_MultiFrameEmission(t *testing.T) {
	t.Setenv("MOAI_REDUCED_MOTION", "")
	t.Setenv("NO_COLOR", "")

	out := &bytes.Buffer{}
	errBuf := &syncBuffer{}
	p := printer.New(
		printer.WithWriters(out, errBuf),
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithTheme(tui.MonochromeTheme()),
		printer.WithAnimatedHandles(true),
		printer.WithAnimationInterval(5*time.Millisecond),
	)

	sp := p.Spinner("deploying templates")

	// Give the animator a few intervals to cycle frames.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if countDistinctFrames(errBuf.String()) >= 3 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if got := countDistinctFrames(errBuf.String()); got < 3 {
		t.Fatalf("animated spinner must emit >= 3 distinct frames, got %d in %q", got, errBuf.String())
	}

	sp.Done("deployed")
	settled := errBuf.String()
	time.Sleep(50 * time.Millisecond)
	if errBuf.String() != settled {
		t.Errorf("no frame emission allowed after Done: output grew from %d to %d bytes",
			len(settled), len(errBuf.String()))
	}
	if out.Len() != 0 {
		t.Errorf("spinner must not write to stdout, got %q", out.String())
	}
}

// TestSpinnerAnimated_UpdateCarriesLiveLabel asserts Update() swaps the live
// label mid-animation (the file-count readout path used by the init
// template-deploy surface).
func TestSpinnerAnimated_UpdateCarriesLiveLabel(t *testing.T) {
	t.Setenv("MOAI_REDUCED_MOTION", "")
	t.Setenv("NO_COLOR", "")

	out := &bytes.Buffer{}
	errBuf := &syncBuffer{}
	p := printer.New(
		printer.WithWriters(out, errBuf),
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithTheme(tui.MonochromeTheme()),
		printer.WithAnimatedHandles(true),
		printer.WithAnimationInterval(5*time.Millisecond),
	)

	sp := p.Spinner("deploying")
	sp.Update("42/96 files")

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(errBuf.String(), "42/96 files") {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !strings.Contains(errBuf.String(), "42/96 files") {
		t.Fatalf("animated spinner must render the updated live label, got %q", errBuf.String())
	}
	sp.Done("done")
}

// TestSpinnerAnimated_FailStopsAnimator asserts Fail() is terminal for the
// animator exactly like Done (goroutine lifecycle edge case).
func TestSpinnerAnimated_FailStopsAnimator(t *testing.T) {
	t.Setenv("MOAI_REDUCED_MOTION", "")
	t.Setenv("NO_COLOR", "")

	out := &bytes.Buffer{}
	errBuf := &syncBuffer{}
	p := printer.New(
		printer.WithWriters(out, errBuf),
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithTheme(tui.MonochromeTheme()),
		printer.WithAnimatedHandles(true),
		printer.WithAnimationInterval(5*time.Millisecond),
	)

	sp := p.Spinner("working")
	time.Sleep(20 * time.Millisecond)
	sp.Fail("broke")
	settled := errBuf.String()
	time.Sleep(50 * time.Millisecond)
	if errBuf.String() != settled {
		t.Errorf("no frame emission allowed after Fail: output grew")
	}
	if !strings.Contains(settled, "broke") {
		t.Errorf("Fail message missing, got %q", settled)
	}
	_ = out
}

// TestSpinnerAnimated_DefaultOffForBuffers asserts animation-capability
// detection: without WithAnimatedHandles(true), a buffer-backed ModeTTY
// printer keeps the legacy static single-frame behavior (existing
// characterization tests depend on it; real terminals enable animation via
// isatty detection at construction).
func TestSpinnerAnimated_DefaultOffForBuffers(t *testing.T) {
	t.Setenv("MOAI_REDUCED_MOTION", "")
	t.Setenv("NO_COLOR", "")

	out := &bytes.Buffer{}
	errBuf := &syncBuffer{}
	p := printer.New(
		printer.WithWriters(out, errBuf),
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithTheme(tui.MonochromeTheme()),
	)

	sp := p.Spinner("static")
	time.Sleep(30 * time.Millisecond)
	if got := countDistinctFrames(errBuf.String()); got > 1 {
		t.Errorf("buffer-backed printer must not self-animate, got %d frames in %q", got, errBuf.String())
	}
	sp.Done("ok")
}

// TestProgressLive_FileCountUpdates asserts the Progress handle renders a
// live k/N readout as Set() advances (AC-TUX2-010 file-count surface).
func TestProgressLive_FileCountUpdates(t *testing.T) {
	t.Setenv("MOAI_REDUCED_MOTION", "")
	t.Setenv("NO_COLOR", "")

	out := &bytes.Buffer{}
	errBuf := &syncBuffer{}
	p := printer.New(
		printer.WithWriters(out, errBuf),
		printer.WithMode(printer.ModeTTY),
		printer.WithNoColor(false),
		printer.WithTheme(tui.MonochromeTheme()),
	)

	pr := p.Progress("files", 96)
	pr.Set(42)
	pr.Set(96)
	got := errBuf.String()
	for _, want := range []string{"0/96", "42/96", "96/96"} {
		if !strings.Contains(got, want) {
			t.Errorf("live progress must render %q, got %q", want, got)
		}
	}
	pr.Done("96 files deployed")
	if !strings.Contains(errBuf.String(), "96 files deployed") {
		t.Errorf("Done message missing, got %q", errBuf.String())
	}
	if out.Len() != 0 {
		t.Errorf("progress must not write to stdout, got %q", out.String())
	}
}

// TestPlainFallback_NoANSINoAnimation asserts the fallback matrix
// (REQ-TUX2-011, AC-TUX2-011): non-TTY plain mode, NO_COLOR-forced plain,
// and reduced-motion TTY all emit zero ANSI escapes and no animation frames
// beyond the single static marker — even when animation is force-enabled.
func TestPlainFallback_NoANSINoAnimation(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T) []printer.Option
	}{
		{
			name: "non-tty-plain",
			setup: func(t *testing.T) []printer.Option {
				t.Setenv("MOAI_REDUCED_MOTION", "")
				return []printer.Option{printer.WithMode(printer.ModePlain), printer.WithAnimatedHandles(true)}
			},
		},
		{
			name: "no-color-forces-plain",
			setup: func(t *testing.T) []printer.Option {
				t.Setenv("MOAI_REDUCED_MOTION", "")
				return []printer.Option{printer.WithMode(printer.ModeTTY), printer.WithNoColor(true), printer.WithAnimatedHandles(true)}
			},
		},
		{
			name: "reduced-motion-tty",
			setup: func(t *testing.T) []printer.Option {
				t.Setenv("MOAI_REDUCED_MOTION", "1")
				return []printer.Option{
					printer.WithMode(printer.ModeTTY),
					printer.WithNoColor(false),
					printer.WithTheme(tui.MonochromeTheme()),
					printer.WithAnimatedHandles(true),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			errBuf := &syncBuffer{}
			opts := append([]printer.Option{printer.WithWriters(out, errBuf)}, tc.setup(t)...)
			p := printer.New(opts...)

			sp := p.Spinner("fallback")
			sp.Update("21/96 files")
			time.Sleep(30 * time.Millisecond)
			sp.Done("ok")

			pr := p.Progress("files", 8)
			pr.Set(4)
			pr.Done("done")

			got := errBuf.String()
			if strings.Contains(got, "\x1b[") {
				t.Errorf("%s: zero ANSI escapes required, got %q", tc.name, got)
			}
			if n := countDistinctFrames(got); n > 1 {
				t.Errorf("%s: no animation frames allowed (max 1 static glyph), got %d in %q", tc.name, n, got)
			}
			if out.Len() != 0 {
				t.Errorf("%s: stdout must stay clean, got %q", tc.name, out.String())
			}
		})
	}
}
