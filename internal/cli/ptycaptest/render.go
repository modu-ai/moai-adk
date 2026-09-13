package ptycaptest

// Render helpers shared by golden (View()) tests and pty capture tests: ANSI
// stripping, display-width column measurement, and golden-file comparison.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

// ansiRe matches CSI sequences (colour, cursor movement, private modes) and
// OSC sequences terminated by BEL or ST (hyperlinks, title).
var ansiRe = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)

// StripANSI removes terminal escape sequences, leaving the painted text.
func StripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// columnWidth measures terminal display width with a fixed condition: East
// Asian wide/fullwidth runes count 2, ambiguous-width runes (box drawing)
// count 1. Fixing the condition keeps the measurement independent of the
// locale environment variables go-runewidth otherwise consults.
var columnWidth = &runewidth.Condition{EastAsianWidth: false, StrictEmojiNeutral: true}

// DisplayColumn returns the 0-based display column at which sub first starts
// in line, or -1 when sub is absent. Columns are display cells, not runes or
// bytes, so a Hangul/Kanji/Hanzi rune advances the column by 2.
func DisplayColumn(line, sub string) int {
	idx := strings.Index(line, sub)
	if idx < 0 {
		return -1
	}
	return columnWidth.StringWidth(line[:idx])
}

// CompareGolden compares got with <dir>/<name>.golden. With update set it
// (re)writes the file and returns nil. A missing golden is an error, never an
// implicit pass. A mismatch error names every differing line with both sides.
// Golden tests pass it an ANSI-stripped View() frame (StripANSI); the caller
// owns the golden directory and its update flag.
func CompareGolden(dir, name, got string, update bool) error {
	path := filepath.Join(dir, name+".golden")
	if update {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("golden %s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			return fmt.Errorf("golden %s: %w", path, err)
		}
		return nil
	}
	want, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("golden %s missing; rerun with -update-golden to create it", path)
	}
	if err != nil {
		return fmt.Errorf("golden %s: %w", path, err)
	}
	if string(want) == got {
		return nil
	}
	wl, gl := strings.Split(string(want), "\n"), strings.Split(got, "\n")
	var b strings.Builder
	fmt.Fprintf(&b, "golden %s differs:", path)
	for i := 0; i < max(len(wl), len(gl)); i++ {
		var w, g string
		if i < len(wl) {
			w = wl[i]
		}
		if i < len(gl) {
			g = gl[i]
		}
		if w != g || i >= len(wl) || i >= len(gl) {
			fmt.Fprintf(&b, "\n  line %d\n    want %q\n    got  %q", i+1, w, g)
		}
	}
	return errors.New(b.String())
}

// RequireLines is the positive-existence gate: every want must appear on some
// line of frame before any absence or layout property is judged.
func RequireLines(tb testing.TB, frame string, wants ...string) {
	tb.Helper()
	var missing []string
	for _, w := range wants {
		if !strings.Contains(frame, w) {
			missing = append(missing, w)
		}
	}
	if len(missing) > 0 {
		tb.Fatalf("frame lacks %q; frame:\n%s", missing, frame)
	}
}
