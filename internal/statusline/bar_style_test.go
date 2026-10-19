package statusline

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Card t1822 follow-up: the usage-bar gradient follows the theme's pastel
// palette stops (green → yellow → peach → red) instead of pure RGB, and an
// n/a usage window keeps an empty, muted placeholder bar in place.

var trueColorFG = regexp.MustCompile(`38;2;(\d+);(\d+);(\d+)m`)

func trueColorCodes(t *testing.T, s string) [][3]int {
	t.Helper()
	var out [][3]int
	for _, m := range trueColorFG.FindAllStringSubmatch(s, -1) {
		var c [3]int
		for i := range 3 {
			v, err := strconv.Atoi(m[i+1])
			if err != nil {
				t.Fatalf("parse %q: %v", m[0], err)
			}
			c[i] = v
		}
		out = append(out, c)
	}
	return out
}

func hasColorNear(codes [][3]int, want [3]int) bool {
	for _, c := range codes {
		if abs(c[0]-want[0]) <= 1 && abs(c[1]-want[1]) <= 1 && abs(c[2]-want[2]) <= 1 {
			return true
		}
	}
	return false
}

func fullContextData() *StatusData {
	return &StatusData{Memory: MemoryData{TokensUsed: 200000, TokenBudget: 200000, Available: true}}
}

func TestRender_GradientFollowsThemePastelStops(t *testing.T) {
	tests := []struct {
		theme      string
		start, end [3]int
	}{
		{theme: "catppuccin-mocha", start: [3]int{0xa6, 0xe3, 0xa1}, end: [3]int{0xf3, 0x8b, 0xa8}},
		{theme: "catppuccin-latte", start: [3]int{0x40, 0xa0, 0x2b}, end: [3]int{0xd2, 0x0f, 0x39}},
	}
	for _, tt := range tests {
		t.Run(tt.theme, func(t *testing.T) {
			t.Setenv("COLORTERM", "truecolor")
			got := NewRenderer(tt.theme, false, nil).Render(fullContextData(), ModeDefault)
			codes := trueColorCodes(t, got)
			if !hasColorNear(codes, tt.start) {
				t.Errorf("first bar block must use the theme green %v, codes %v", tt.start, codes)
			}
			if !hasColorNear(codes, tt.end) {
				t.Errorf("last bar block must use the theme red %v, codes %v", tt.end, codes)
			}
			for _, saturated := range [][3]int{{0, 255, 0}, {255, 0, 0}, {255, 255, 0}} {
				if hasColorNear(codes, saturated) {
					t.Errorf("pure RGB %v must no longer appear, codes %v", saturated, codes)
				}
			}
		})
	}
}

var placeholderPattern = regexp.MustCompile(`(5H|7D): (░+) n/a`)
var cwBarPattern = regexp.MustCompile(`CW: ([█░]+) `)

func halfContextData() *StatusData {
	return &StatusData{Memory: MemoryData{TokensUsed: 100000, TokenBudget: 200000, Available: true}}
}

func TestRender_QuotaUnknown_KeepsPlaceholderBar(t *testing.T) {
	for _, noColor := range []bool{true, false} {
		t.Run("noColor="+strconv.FormatBool(noColor), func(t *testing.T) {
			t.Setenv("COLORTERM", "truecolor")
			r := NewRenderer("catppuccin-mocha", noColor, nil)
			got := r.Render(halfContextData(), ModeDefault)
			plain := stripSGR(got)

			cw := cwBarPattern.FindStringSubmatch(plain)
			if cw == nil {
				t.Fatalf("CW bar missing: %q", plain)
			}
			width := len([]rune(cw[1]))

			matches := placeholderPattern.FindAllStringSubmatch(plain, -1)
			if len(matches) != 2 {
				t.Fatalf("5H and 7D must each render a placeholder bar before n/a, got %q", plain)
			}
			for _, m := range matches {
				if n := len([]rune(m[2])); n != width {
					t.Errorf("%s placeholder width %d, want the fresh bar width %d", m[1], n, width)
				}
			}
			for _, seg := range strings.Split(plain, " | ") {
				if !strings.HasPrefix(seg, "5H:") && !strings.HasPrefix(seg, "7D:") {
					continue
				}
				if rest := seg[len("5H:"):]; strings.ContainsAny(rest, "0123456789%") {
					t.Errorf("n/a segment must carry no digits or percent after its label, got %q", seg)
				}
			}
			if noColor {
				if strings.Contains(got, "\x1b[") {
					t.Errorf("NoColor placeholder must be plain, got %q", got)
				}
				return
			}
			muted := foregroundSequence(newForcedColorRenderer("truecolor"), NewTheme("catppuccin-mocha").Muted())
			if !strings.Contains(got, muted+strings.Repeat("░", width)) {
				t.Errorf("placeholder blocks must be drawn in the theme muted color %q, got %q", muted, got)
			}
		})
	}
}

// Operator decision 2026-10-16 (card t1822): CW behaves like 5H/7D — missing
// context-window data keeps an empty placeholder bar plus "n/a" instead of
// dropping the segment.
func TestRender_ContextUnknown_KeepsPlaceholderBar(t *testing.T) {
	for _, noColor := range []bool{true, false} {
		t.Run("noColor="+strconv.FormatBool(noColor), func(t *testing.T) {
			t.Setenv("COLORTERM", "truecolor")
			got := NewRenderer("catppuccin-mocha", noColor, nil).Render(&StatusData{}, ModeDefault)
			plain := stripSGR(got)
			m := regexp.MustCompile(`(?:^|\n)CW: (░+) n/a \| 5H: (░+) n/a`).FindStringSubmatch(plain)
			if m == nil {
				t.Fatalf("CW must render a placeholder bar before n/a, got %q", plain)
			}
			if len([]rune(m[1])) != len([]rune(m[2])) {
				t.Errorf("CW placeholder width %d, want the 5H width %d", len([]rune(m[1])), len([]rune(m[2])))
			}
			if !noColor {
				muted := foregroundSequence(newForcedColorRenderer("truecolor"), NewTheme("catppuccin-mocha").Muted())
				if !strings.Contains(got, "CW: "+muted+m[1]) {
					t.Errorf("CW placeholder must be muted-colored, got %q", got)
				}
			}
		})
	}
}
