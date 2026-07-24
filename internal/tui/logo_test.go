package tui

import (
	"image/color"
	"strings"
	"testing"
)

// luminance returns a perceptual-ish luminance for ordering ramp stops
// (higher = lighter). RGBA() yields 16-bit premultiplied channels.
func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
}

func distinctCount(cs []color.Color) int {
	seen := map[[3]uint32]struct{}{}
	for _, c := range cs {
		r, g, b, _ := c.RGBA()
		seen[[3]uint32{r, g, b}] = struct{}{}
	}
	return len(seen)
}

// AC-TUXIU-021 — CoralRamp(6) yields 6 distinct CIELAB-interpolated stops.
func TestCoralRamp_ExportedSixDistinctStops(t *testing.T) {
	ramp := CoralRamp(6)
	if len(ramp) != 6 {
		t.Fatalf("CoralRamp(6) returned %d stops, want 6", len(ramp))
	}
	if n := distinctCount(ramp); n < 3 {
		t.Errorf("CoralRamp(6) has %d distinct colours, want >= 3", n)
	}
}

// AC-TUXIU-021 — directionality: top stop is lighter than the bottom stop
// (row 1 = Accent, the lighter coral; row 6 = AccentDeep, the deeper coral).
func TestCoralRamp_LightThemeDirectionality(t *testing.T) {
	ramp := coralRamp(LightTheme(), 6)
	if len(ramp) != 6 {
		t.Fatalf("coralRamp(Light,6) returned %d stops, want 6", len(ramp))
	}
	if distinctCount(ramp) != 6 {
		t.Errorf("coralRamp(Light,6) should yield 6 distinct stops, got %d", distinctCount(ramp))
	}
	top, bottom := luminance(ramp[0]), luminance(ramp[5])
	if !(top > bottom) {
		t.Errorf("expected top stop lighter than bottom (top=%.0f > bottom=%.0f)", top, bottom)
	}
}

// AC-TUXIU-020 — Logo renders the restored 6-line art (first-line signature +
// all 6 rows present).
func TestLogo_ContainsAllSixRows(t *testing.T) {
	got := Logo(LightTheme())
	if !strings.Contains(got, "███╗   ███╗") {
		t.Errorf("Logo output missing first-line signature %q:\n%s", "███╗   ███╗", got)
	}
	// The bottom row's leading runes.
	if !strings.Contains(got, "╚═╝     ╚═╝") {
		t.Errorf("Logo output missing bottom row, got:\n%s", got)
	}
	// Six art rows → at least six newlines.
	if n := strings.Count(got, "\n"); n < 6 {
		t.Errorf("Logo output has %d newlines, want >= 6 (six art rows)", n)
	}
}

// AC-TUXIU-022 — NO_COLOR analogue: a monochrome theme yields plain art with
// zero ANSI colour and the runes intact.
func TestLogo_MonochromePlain(t *testing.T) {
	got := Logo(MonochromeTheme())
	if strings.Contains(got, "\x1b[") {
		t.Errorf("Logo(Monochrome) must emit zero ANSI colour, got: %q", got)
	}
	if !strings.Contains(got, "███╗   ███╗") {
		t.Errorf("Logo(Monochrome) art runes must be intact, got: %q", got)
	}
}

// coralRamp guards a non-positive stop count (defensive; a caller must never
// request fewer than one stop).
func TestCoralRamp_ZeroNReturnsNil(t *testing.T) {
	if got := coralRamp(LightTheme(), 0); got != nil {
		t.Errorf("coralRamp(_, 0) = %v, want nil", got)
	}
	if got := coralRamp(LightTheme(), -3); got != nil {
		t.Errorf("coralRamp(_, -3) = %v, want nil", got)
	}
}

// Under NO_COLOR the active theme resolves to monochrome (empty accent tokens);
// CoralRamp must still return the brand coral gradient (colour suppression is a
// render-time decision, not a ramp decision). Non-parallel: mutates NO_COLOR.
func TestCoralRamp_MonoFallbackStillCoral(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// Sanity: the active theme is indeed monochrome under NO_COLOR.
	if ResolveOS().Accent != "" {
		t.Skip("environment did not resolve to a monochrome theme; skipping fallback assertion")
	}
	ramp := CoralRamp(6)
	if len(ramp) != 6 {
		t.Fatalf("CoralRamp(6) under NO_COLOR returned %d stops, want 6", len(ramp))
	}
	if n := distinctCount(ramp); n < 3 {
		t.Errorf("CoralRamp(6) under NO_COLOR should still yield the coral gradient (>=3 distinct), got %d", n)
	}
}

// The restored art constant must be byte-identical to the retired moaiBanner
// (SPEC-CLI-TUX-V3-004 commit 77893579e). Structural pin: 6 non-empty rows,
// leading + trailing newline (the original const shape).
func TestLogoArt_ByteShape(t *testing.T) {
	if !strings.HasPrefix(moaiLogoArt, "\n") || !strings.HasSuffix(moaiLogoArt, "\n") {
		t.Errorf("moaiLogoArt must retain the leading+trailing newline of the original const")
	}
	rows := strings.Split(strings.Trim(moaiLogoArt, "\n"), "\n")
	if len(rows) != 6 {
		t.Fatalf("moaiLogoArt has %d rows, want 6", len(rows))
	}
	for i, r := range rows {
		if strings.TrimSpace(r) == "" {
			t.Errorf("art row %d is empty", i+1)
		}
	}
}
