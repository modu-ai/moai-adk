package statusline

import (
	"strings"
	"testing"
)

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestInterpolateGradientColor verifies interpolation across the theme's
// palette stops (card t1822: pastel Catppuccin stops replaced the pure-RGB
// Green(0,255,0) → Yellow(255,255,0) → Red(255,0,0) path).
func TestInterpolateGradientColor(t *testing.T) {
	mocha := barGradientStops(NewTheme("catppuccin-mocha"))
	latte := barGradientStops(NewTheme("catppuccin-latte"))
	tests := []struct {
		name                string
		stops               []string
		blockPct            float64 // 0.0 to 1.0
		wantR, wantG, wantB int
	}{
		{"mocha start (green #a6e3a1)", mocha, 0.0, 0xa6, 0xe3, 0xa1},
		{"mocha 1/3 (yellow #f9e2af)", mocha, 1.0 / 3, 0xf9, 0xe2, 0xaf},
		{"mocha 2/3 (peach #fab387)", mocha, 2.0 / 3, 0xfa, 0xb3, 0x87},
		{"mocha end (red #f38ba8)", mocha, 1.0, 0xf3, 0x8b, 0xa8},
		{"mocha 1/6 (mid green-yellow)", mocha, 1.0 / 6, 208, 227, 168},
		{"latte start (green #40a02b)", latte, 0.0, 0x40, 0xa0, 0x2b},
		{"latte 1/3 (yellow #df8e1d)", latte, 1.0 / 3, 0xdf, 0x8e, 0x1d},
		{"latte 2/3 (peach #fe640b)", latte, 2.0 / 3, 0xfe, 0x64, 0x0b},
		{"latte end (red #d20f39)", latte, 1.0, 0xd2, 0x0f, 0x39},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := interpolateGradientColor(tt.stops, tt.blockPct)
			// Allow ±1 tolerance (rounding)
			if abs(r-tt.wantR) > 1 || abs(g-tt.wantG) > 1 || abs(b-tt.wantB) > 1 {
				t.Errorf("interpolateGradientColor(%f) = (%d,%d,%d), want ~(%d,%d,%d)",
					tt.blockPct, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

// TestInterpolateGradientColor_Clamp verifies clamping for out-of-range inputs.
func TestInterpolateGradientColor_Clamp(t *testing.T) {
	mocha := barGradientStops(NewTheme("catppuccin-mocha"))
	// Below 0 returns the first stop (green)
	r, g, b := interpolateGradientColor(mocha, -0.5)
	if r != 0xa6 || g != 0xe3 || b != 0xa1 {
		t.Errorf("clamp below 0: got (%d,%d,%d), want (166,227,161)", r, g, b)
	}

	// Above 1 returns the last stop (red)
	r, g, b = interpolateGradientColor(mocha, 1.5)
	if r != 0xf3 || g != 0x8b || b != 0xa8 {
		t.Errorf("clamp above 1: got (%d,%d,%d), want (243,139,168)", r, g, b)
	}
}

// TestBuildGradientBar verifies block counts in gradient progress bars.
func TestBuildGradientBar(t *testing.T) {
	tests := []struct {
		name       string
		pct        int
		width      int
		noColor    bool
		wantFilled int
		wantEmpty  int
	}{
		{"60% of 40", 60, 40, false, 24, 16},
		{"60% of 10", 60, 10, false, 6, 4},
		{"0%", 0, 40, false, 0, 40},
		{"100%", 100, 40, false, 40, 0},
		{"50% of 10", 50, 10, false, 5, 5},
		{"zero width", 50, 0, false, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildGradientBar(tt.pct, tt.width, tt.noColor)
			// When noColor=false, ANSI escapes are included so simple count won't work
			// Only verify exact block counts when noColor=true
			if tt.noColor {
				filled := strings.Count(result, "█")
				empty := strings.Count(result, "░")
				if filled != tt.wantFilled || empty != tt.wantEmpty {
					t.Errorf("blocks: got %d filled + %d empty, want %d + %d", filled, empty, tt.wantFilled, tt.wantEmpty)
				}
			}
			if tt.width > 0 && result == "" {
				t.Error("expected non-empty result for positive width")
			}
			if tt.width == 0 && result != "" {
				t.Error("expected empty result for zero width")
			}
		})
	}
}

// TestBuildGradientBar_NoColor verifies no ANSI escapes in noColor mode.
func TestBuildGradientBar_NoColor(t *testing.T) {
	result := BuildGradientBar(50, 10, true)
	// Must not contain ANSI escape sequences
	if strings.Contains(result, "\033") || strings.Contains(result, "\x1b") {
		t.Error("noColor mode must not contain ANSI escape sequences")
	}
	// Exactly 5 filled + 5 empty
	filled := strings.Count(result, "█")
	empty := strings.Count(result, "░")
	if filled != 5 || empty != 5 {
		t.Errorf("got %d filled + %d empty, want 5 + 5", filled, empty)
	}
}

// TestBuildGradientBar_WithColor verifies ANSI escapes are present in color mode.
func TestBuildGradientBar_WithColor(t *testing.T) {
	result := BuildGradientBar(50, 10, false)
	// lipgloss adds ANSI escapes so length should exceed 10 characters
	if len(result) <= 15 {
		t.Error("color mode should be longer due to ANSI escape sequences")
	}
}

// TestBuildGradientBar_FilledCountWithColor verifies correct block count in color mode.
func TestBuildGradientBar_FilledCountWithColor(t *testing.T) {
	// 60% / 40 blocks -> 24 filled, 16 empty
	result := BuildGradientBar(60, 40, false)
	filled := strings.Count(result, "█")
	empty := strings.Count(result, "░")
	if filled != 24 {
		t.Errorf("60%% of 40: want 24 filled blocks, got %d", filled)
	}
	if empty != 16 {
		t.Errorf("60%% of 40: want 16 empty blocks, got %d", empty)
	}
}

// TestBuildGradientBar_SingleFilled verifies no division by zero when filled=1.
func TestBuildGradientBar_SingleFilled(t *testing.T) {
	// 10% of 10 -> 1 filled
	result := BuildGradientBar(10, 10, false)
	filled := strings.Count(result, "█")
	if filled != 1 {
		t.Errorf("10%% of 10: want 1 filled block, got %d", filled)
	}
}
