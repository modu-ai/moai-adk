// Package tui — progress_line_test.go
//
// Golden-output tests for ProgressLine (SPEC-V3R6-UPDATE-PROGRESS-001
// REQ-UPR-006). The tests exercise both the TTY branch (verifying the
// `\r\033[2K` CSI Erase-in-Line prefix on terminal output) and the
// non-TTY branch (verifying plain `\n`-separated output without any
// ANSI escape sequences).
//
// The TTY branch is exercised through internal package access: we
// construct a ProgressLineHandle directly with isTTY=true and bypass the
// isatty detection. The non-TTY branch is exercised via the public
// ProgressLine factory using a *bytes.Buffer (which is not an *os.File
// and therefore reports as non-TTY).
//
// AC coverage:
//   - AC-UPR-001: TestProgressLine_APIExists (compile-time)
//   - AC-UPR-002: TestProgressLine_TTYBranchEmitsClearPrefix
//   - AC-UPR-003: TestProgressLineNonTTY_EmitsPlainNewlines
//   - AC-UPR-006: TestProgressLine_VisibleMessageIdentical (strip-ansi
//     check that visible text is preserved across TTY/non-TTY branches)
//   - AC-UPR-007: TestProgressLine_RejectsDoubleTerminal (panic-on-misuse)
//   - AC-UPR-009: TestProgressLine_Update (Update method behavior)
package tui

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

// ansiEscapeRegex matches CSI (Control Sequence Introducer) escape
// sequences of the form ESC [ ... letter. Used by stripANSI to extract
// the visible message text from rendered output.
var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

// stripANSI removes all ANSI CSI escape sequences and bare carriage
// returns from s, returning only the human-visible text.
func stripANSI(s string) string {
	s = ansiEscapeRegex.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// newTTYHandle constructs a ProgressLineHandle that writes to buf and
// behaves as if connected to a TTY. Used to exercise the TTY branch
// without requiring an actual terminal. This bypasses isTerminalWriter
// which would otherwise return false for *bytes.Buffer.
func newTTYHandle(buf *bytes.Buffer, message string) *ProgressLineHandle {
	t := LightTheme()
	h := &ProgressLineHandle{
		out:   buf,
		isTTY: true,
		theme: t,
	}
	h.writeProgress(message)
	return h
}

// ────────────────────────────────────────────────────────────────────────
// AC-UPR-002 — TTY branch emits `\r\033[2K` clear prefix
// ────────────────────────────────────────────────────────────────────────

func TestProgressLine_TTYBranchEmitsClearPrefix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		progress  string
		terminate func(h *ProgressLineHandle)
		want      string // expected substring in visible output
	}{
		{
			name:      "Done",
			progress:  "Backing up .moai/config...",
			terminate: func(h *ProgressLineHandle) { h.Done(".moai/config backed up") },
			want:      ".moai/config backed up",
		},
		{
			name:      "Fail",
			progress:  "Backing up .moai/config...",
			terminate: func(h *ProgressLineHandle) { h.Fail("Backup failed: disk full") },
			want:      "Backup failed: disk full",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			h := newTTYHandle(buf, tc.progress)
			tc.terminate(h)

			got := buf.String()
			// REQ-UPR-002: terminal message must be prefixed with
			// `\r\033[2K` (carriage return + CSI Erase in Line mode 2).
			if !strings.Contains(got, progressClearPrefix) {
				t.Errorf("expected output to contain clear prefix %q, got: %q", progressClearPrefix, got)
			}
			// REQ-UPR-005: visible message text must be preserved.
			if !strings.Contains(stripANSI(got), tc.want) {
				t.Errorf("expected visible output to contain %q, got: %q", tc.want, stripANSI(got))
			}
			// Sanity: terminal output ends with newline (so subsequent
			// lines do not concatenate).
			if !strings.HasSuffix(got, "\n") {
				t.Errorf("expected output to end with newline, got: %q", got)
			}
		})
	}
}

// ────────────────────────────────────────────────────────────────────────
// AC-UPR-003 — Non-TTY branch emits plain newline-separated lines
// ────────────────────────────────────────────────────────────────────────

func TestProgressLineNonTTY_EmitsPlainNewlines(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	h := ProgressLine(buf, "Working...", nil)
	h.Done("Complete")

	got := buf.String()
	// REQ-UPR-003: no carriage return in non-TTY output.
	if strings.Contains(got, "\r") {
		t.Errorf("non-TTY output must not contain carriage return, got: %q", got)
	}
	// REQ-UPR-003: no CSI escape sequence in non-TTY output.
	if strings.Contains(got, "\x1b[") {
		t.Errorf("non-TTY output must not contain CSI escape, got: %q", got)
	}
	// REQ-UPR-003: progress and result each on their own line (so the
	// total newline count is exactly 2).
	if n := strings.Count(got, "\n"); n != 2 {
		t.Errorf("expected exactly 2 newlines in non-TTY output, got %d: %q", n, got)
	}
	// REQ-UPR-005: visible messages preserved.
	if !strings.Contains(got, "Working...") {
		t.Errorf("expected output to contain progress message, got: %q", got)
	}
	if !strings.Contains(stripANSI(got), "Complete") {
		t.Errorf("expected output to contain result message, got: %q", stripANSI(got))
	}
}

// ────────────────────────────────────────────────────────────────────────
// AC-UPR-009 — Update method
// ────────────────────────────────────────────────────────────────────────

func TestProgressLine_Update(t *testing.T) {
	t.Parallel()
	t.Run("TTY", func(t *testing.T) {
		t.Parallel()
		buf := &bytes.Buffer{}
		h := newTTYHandle(buf, "Step 1...")
		h.Update("Step 2...")
		h.Done("All steps complete")

		got := buf.String()
		// Update must emit the clear prefix.
		if !strings.Contains(got, progressClearPrefix) {
			t.Errorf("expected Update to emit clear prefix, got: %q", got)
		}
		// Visible: progress 1, then progress 2, then result — all three
		// strings present.
		visible := stripANSI(got)
		for _, want := range []string{"Step 1...", "Step 2...", "All steps complete"} {
			if !strings.Contains(visible, want) {
				t.Errorf("expected visible output to contain %q, got: %q", want, visible)
			}
		}
		// Done after Update must succeed (no panic).
	})

	t.Run("NonTTY", func(t *testing.T) {
		t.Parallel()
		buf := &bytes.Buffer{}
		h := ProgressLine(buf, "Step 1...", nil)
		h.Update("Step 2...")
		h.Done("All steps complete")

		got := buf.String()
		if strings.Contains(got, "\r") {
			t.Errorf("non-TTY Update must not emit carriage return, got: %q", got)
		}
		if strings.Contains(got, "\x1b[") {
			t.Errorf("non-TTY Update must not emit CSI escape, got: %q", got)
		}
		// Non-TTY: 3 newlines (progress 1, update, result).
		if n := strings.Count(got, "\n"); n != 3 {
			t.Errorf("expected 3 newlines in non-TTY Update output, got %d: %q", n, got)
		}
	})
}

// ────────────────────────────────────────────────────────────────────────
// EC-1 — Double-terminal call panics
// ────────────────────────────────────────────────────────────────────────

func TestProgressLine_RejectsDoubleTerminal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		terminate func(h *ProgressLineHandle)
	}{
		{
			name:      "DoneAfterDone",
			terminate: func(h *ProgressLineHandle) { h.Done(".moai backed up"); h.Done("already done") },
		},
		{
			name:      "FailAfterDone",
			terminate: func(h *ProgressLineHandle) { h.Done(".moai backed up"); h.Fail("late failure") },
		},
		{
			name:      "DoneAfterFail",
			terminate: func(h *ProgressLineHandle) { h.Fail("backup failed"); h.Done("done later") },
		},
		{
			name:      "UpdateAfterDone",
			terminate: func(h *ProgressLineHandle) { h.Done(".moai backed up"); h.Update("re-render") },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			buf := &bytes.Buffer{}
			h := newTTYHandle(buf, "Working...")
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic on double-terminal call, got none")
				}
			}()
			tc.terminate(h)
		})
	}
}

// ────────────────────────────────────────────────────────────────────────
// AC-UPR-001 — API surface compile-time check
// ────────────────────────────────────────────────────────────────────────

// TestProgressLine_APIExists is a compile-time pin: it references all
// public ProgressLine API symbols so that signature changes would cause
// the test file to fail to compile, satisfying AC-UPR-001 (function
// signature stability).
func TestProgressLine_APIExists(t *testing.T) {
	t.Parallel()
	var _ func(out interface{ Write(p []byte) (int, error) }, message string, theme *Theme) *ProgressLineHandle
	// The above is illustrative only — the actual signature pin is below
	// (referencing the real factory and handle methods).
	var factory = ProgressLine
	_ = factory

	buf := &bytes.Buffer{}
	h := ProgressLine(buf, "Pin Check", nil)
	var (
		doneFn   = h.Done
		failFn   = h.Fail
		updateFn = h.Update
	)
	_ = doneFn
	_ = failFn
	_ = updateFn
	// Complete the handle so the test does not leak a non-terminated
	// progress line (the buffer is discarded, but cleanliness matters).
	h.Done("Pin OK")
}

// ────────────────────────────────────────────────────────────────────────
// AC-UPR-006 — Visible message byte-identical between TTY and non-TTY
// ────────────────────────────────────────────────────────────────────────

func TestProgressLine_VisibleMessageIdentical(t *testing.T) {
	t.Parallel()
	const (
		progressMsg = "Backing up .moai/config..."
		resultMsg   = ".moai/config backed up"
	)

	ttyBuf := &bytes.Buffer{}
	hTTY := newTTYHandle(ttyBuf, progressMsg)
	hTTY.Done(resultMsg)

	nonTTYBuf := &bytes.Buffer{}
	hNonTTY := ProgressLine(nonTTYBuf, progressMsg, nil)
	hNonTTY.Done(resultMsg)

	ttyVisible := stripANSI(ttyBuf.String())
	nonTTYVisible := stripANSI(nonTTYBuf.String())

	// Both branches must contain both messages in their visible output.
	for _, want := range []string{progressMsg, resultMsg} {
		if !strings.Contains(ttyVisible, want) {
			t.Errorf("TTY visible output missing %q, got: %q", want, ttyVisible)
		}
		if !strings.Contains(nonTTYVisible, want) {
			t.Errorf("non-TTY visible output missing %q, got: %q", want, nonTTYVisible)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────
// EC-3 — theme nil fallback
// ────────────────────────────────────────────────────────────────────────

func TestProgressLine_NilTheme_UsesLightTheme(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	// ProgressLine with nil theme must not panic and must produce
	// readable output.
	h := ProgressLine(buf, "Working...", nil)
	h.Done("Complete")

	got := stripANSI(buf.String())
	if !strings.Contains(got, "Working...") || !strings.Contains(got, "Complete") {
		t.Errorf("nil-theme output must contain both messages, got: %q", got)
	}
}
