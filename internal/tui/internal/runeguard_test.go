// Package internal provides low-level TUI utilities.
package internal_test

import (
	"testing"

	tuiinternal "github.com/modu-ai/moai-adk/internal/tui/internal"
)

// TestRunewidthBasic verifies StringWidth for ASCII and Korean strings.
func TestRunewidthBasic(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"환경 감지", 9},     // 환(2)+경(2)+space(1)+감(2)+지(2) = 9
		{"Layer 1", 7},
		{"v3.2.4", 6},
		{"한자漢字", 8},       // 한(2)+자(2)+漢(2)+字(2) = 8
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := tuiinternal.StringWidth(tc.input)
			if got != tc.want {
				t.Errorf("StringWidth(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

// TestMixedKoEnAlignment verifies StringWidth + FillRight alignment for mixed strings.
// These cases appear in AC-CLI-TUI-007's mixed alignment test suite.
func TestMixedKoEnAlignment(t *testing.T) {
	cases := []struct {
		input     string
		wantWidth int
	}{
		{"환경 감지", 9},
		{"Layer 1", 7},
		{"환경 감지 Layer 1", 17}, // 9 + space(1) + 7 = 17
		{"한자漢字", 8},
		// U+00B7(·) is Unicode "ambiguous". runeguard pins EastAsianWidth=true so
		// it always renders as 2 cells regardless of the runtime locale.
		{"v3.2.4 · 2026-04-28", 20},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			w := tuiinternal.StringWidth(tc.input)
			if w != tc.wantWidth {
				t.Errorf("StringWidth(%q) = %d, want %d", tc.input, w, tc.wantWidth)
			}
		})
	}
}

// TestTruncate verifies Truncate handles CJK characters without splitting cells.
func TestTruncate(t *testing.T) {
	cases := []struct {
		input string
		n     int
		want  string
	}{
		{"hello world", 5, "hello"},
		{"환경 감지 Layer", 5, "환경 "},   // 환(2)+경(2)+space(1) = 5
		{"abcdef", 3, "abc"},
		{"한한한", 4, "한한"},             // each 한=2 cells, 4 cells = 2 chars
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := tuiinternal.Truncate(tc.input, tc.n)
			w := tuiinternal.StringWidth(got)
			if w > tc.n {
				t.Errorf("Truncate(%q, %d) = %q width=%d, exceeds max %d", tc.input, tc.n, got, w, tc.n)
			}
			if got != tc.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tc.input, tc.n, got, tc.want)
			}
		})
	}
}

// TestFillRight verifies FillRight pads a string to target display width.
func TestFillRight(t *testing.T) {
	cases := []struct {
		input     string
		width     int
		wantWidth int
	}{
		{"hello", 10, 10},
		{"환경", 10, 10},
		{"已知", 8, 8},
		{"ok", 2, 2}, // no padding needed
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := tuiinternal.FillRight(tc.input, tc.width)
			w := tuiinternal.StringWidth(got)
			if w != tc.wantWidth {
				t.Errorf("FillRight(%q, %d) width=%d, want %d", tc.input, tc.width, w, tc.wantWidth)
			}
		})
	}
}
