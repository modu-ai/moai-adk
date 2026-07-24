// Package tui — progress_line_residue_test.go
//
// Characterization tests locking the spinner-residue contract of the
// ProgressLine primitive (SPEC-CLI-TUX-INIT-UPDATE-001 M1, AC-TUXIU-006 /
// AC-TUXIU-007). A finished step MUST clear its in-flight line on a TTY and
// leave exactly one clean result line, with NO residual progress glyph
// (GlyphSkip "○") or spinner-frame fragment after the erase; on a non-TTY the
// result is a single plain result line with no ANSI erase sequence.
//
// The residue fix itself lives in the erase-line prefix (progressClearPrefix =
// "\r\x1b[2K") applied by writeTerminal on the TTY branch; these tests pin that
// contract so a regression (e.g. reverting to a bare "\r") is caught.
package tui

import (
	"bytes"
	"strings"
	"testing"
)

// lastEraseTail returns the segment of s AFTER the final progressClearPrefix
// occurrence — i.e. what a real terminal displays on the line after the last
// erase-in-line. When no erase prefix is present the whole string is returned.
func lastEraseTail(s string) string {
	i := strings.LastIndex(s, progressClearPrefix)
	if i < 0 {
		return s
	}
	return s[i+len(progressClearPrefix):]
}

// AC-TUXIU-006 — TTY: finished step clears the line, one clean ✓ result line,
// zero residual progress-glyph / spinner-frame fragments after the erase.
func TestStepResidue_TTY_CleanSingleResultLine(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	h := newTTYHandle(buf, "Deploying templates...")
	h.Done("Templates deployed")

	got := buf.String()

	// The spinner line is cleared via the ANSI erase-in-line sequence.
	if !strings.Contains(got, progressClearPrefix) {
		t.Fatalf("expected erase-line prefix %q in TTY output, got: %q", progressClearPrefix, got)
	}
	tail := lastEraseTail(got)
	// Exactly one result line after the last erase (one trailing newline).
	if n := strings.Count(tail, "\n"); n != 1 {
		t.Errorf("expected exactly 1 result line after erase, got %d: %q", n, tail)
	}
	// The clean result line carries the success glyph...
	if !strings.Contains(tail, string(GlyphDone)) {
		t.Errorf("expected success glyph %q in cleared result line, got: %q", string(GlyphDone), tail)
	}
	// ...and NO residual progress glyph or spinner-frame fragment.
	if strings.ContainsRune(tail, GlyphSkip) {
		t.Errorf("residual progress glyph %q leaked past the erase: %q", string(GlyphSkip), tail)
	}
	if strings.Contains(tail, "⠋") {
		t.Errorf("residual spinner-frame fragment leaked past the erase: %q", tail)
	}
}

// AC-TUXIU-007 — non-TTY: single newline-terminated result line, NO ANSI erase
// sequence (the TTY/non-TTY split is preserved).
func TestStepResidue_NonTTY_NoAnsiSingleResultLine(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	h := ProgressLine(buf, "Deploying templates...", nil)
	h.Done("Templates deployed")

	got := buf.String()

	// No ANSI CSI sequence at all on the non-TTY path.
	if strings.Contains(got, "\x1b[") {
		t.Errorf("non-TTY output must contain zero ANSI CSI sequences, got: %q", got)
	}
	if strings.Contains(got, "\r") {
		t.Errorf("non-TTY output must contain zero carriage returns, got: %q", got)
	}
	// Exactly one ✓ result line.
	if c := strings.Count(got, string(GlyphDone)); c != 1 {
		t.Errorf("expected exactly one ✓ result line, got %d: %q", c, got)
	}
	if !strings.Contains(got, "Templates deployed") {
		t.Errorf("expected result message in output, got: %q", got)
	}
}

// The error path also clears without residue (edge case: a step that fails
// mid-deploy still clears the spinner line — acceptance.md §D.3).
func TestStepResidue_TTY_FailPathClears(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	h := newTTYHandle(buf, "Deploying templates...")
	h.Fail("deploy failed")

	tail := lastEraseTail(buf.String())
	if !strings.Contains(tail, string(GlyphErr)) {
		t.Errorf("expected error glyph %q in cleared result line, got: %q", string(GlyphErr), tail)
	}
	if strings.ContainsRune(tail, GlyphSkip) {
		t.Errorf("residual progress glyph leaked past the erase on the fail path: %q", tail)
	}
}
