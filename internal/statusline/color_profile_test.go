package statusline

import (
	"regexp"
	"strings"
	"testing"
)

// Card t1822: Claude Code captures the statusline command's stdout through a
// pipe, so TTY-based color detection always resolved to the Ascii profile and
// the bar reached Claude Code uncolored. Since Claude Code 2.1.299-2.1.301
// renders uncolored statusline text genuinely dim, the renderer must force a
// color profile when color is enabled, give every visible rune an explicit
// foreground, and open each line with SGR 22 (normal intensity).

var sgrPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

// colorProfileFixture exercises the plain-text segments (model, versions,
// directory, branch) and the gradient bars in one render.
func colorProfileFixture() *StatusData {
	return &StatusData{
		Metrics: MetricsData{Model: "Opus 5.5", Available: true},
		Memory:  MemoryData{TokensUsed: 120000, TokenBudget: 200000, Available: true},
		Version: VersionData{Current: "2.1.301", Available: true},
		Git:     GitStatusData{Branch: "main", Available: true},
	}
}

func stripSGR(s string) string { return sgrPattern.ReplaceAllString(s, "") }

// testBarStyle maps a legacy noColor flag onto the usage-bar helpers' style
// argument: nil renders plain, otherwise a forced-profile Mocha style.
func testBarStyle(noColor bool) *barStyle {
	if noColor {
		return nil
	}
	return newBarStyle(newForcedColorRenderer(""), NewTheme(""))
}

func TestRender_ColorEnabled_ForcesProfileOnNonTTY(t *testing.T) {
	t.Setenv("COLORTERM", "")
	r := NewRenderer("catppuccin-mocha", false, nil)
	got := r.Render(colorProfileFixture(), ModeDefault)

	if !strings.Contains(got, "\x1b[38;") {
		t.Fatalf("color-enabled render on a non-TTY must carry SGR foreground codes, got %q", got)
	}
	for i, line := range strings.Split(got, "\n") {
		if !strings.HasPrefix(line, "\x1b[22m") {
			t.Errorf("line %d must start with SGR 22 (normal intensity), got %q", i, line)
		}
	}
}

func TestRender_ColorProfile_FollowsCOLORTERM(t *testing.T) {
	tests := []struct {
		colorterm string
		want      string
		forbid    string
	}{
		{colorterm: "truecolor", want: "38;2;", forbid: "38;5;"},
		{colorterm: "24bit", want: "38;2;", forbid: "38;5;"},
		{colorterm: "", want: "38;5;", forbid: "38;2;"},
		{colorterm: "yes", want: "38;5;", forbid: "38;2;"},
	}
	for _, tt := range tests {
		t.Run("COLORTERM="+tt.colorterm, func(t *testing.T) {
			t.Setenv("COLORTERM", tt.colorterm)
			got := NewRenderer("catppuccin-mocha", false, nil).Render(colorProfileFixture(), ModeDefault)
			if !strings.Contains(got, tt.want) {
				t.Errorf("want %q codes, got %q", tt.want, got)
			}
			if strings.Contains(got, tt.forbid) {
				t.Errorf("must not carry %q codes, got %q", tt.forbid, got)
			}
		})
	}
}

func TestRender_NoColor_EmitsNoEscapesEvenWithTruecolor(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	got := NewRenderer("catppuccin-mocha", true, nil).Render(colorProfileFixture(), ModeDefault)
	if strings.Contains(got, "\x1b[") {
		t.Errorf("NoColor render must not contain escape codes, got %q", got)
	}
}

// Every visible (non-space) rune must sit under an explicit foreground so none
// falls back to Claude Code's default/secondary text rendering.
func TestRender_ColorEnabled_EveryVisibleRuneHasForeground(t *testing.T) {
	for _, theme := range []string{"catppuccin-mocha", "catppuccin-latte"} {
		t.Run(theme, func(t *testing.T) {
			t.Setenv("COLORTERM", "truecolor")
			got := NewRenderer(theme, false, nil).Render(colorProfileFixture(), ModeDefault)
			for i, line := range strings.Split(got, "\n") {
				assertEveryRuneColored(t, i, line)
			}
		})
	}
}

func assertEveryRuneColored(t *testing.T, lineNo int, line string) {
	t.Helper()
	fg := false
	rest := line
	for rest != "" {
		if loc := sgrPattern.FindStringIndex(rest); loc != nil && loc[0] == 0 {
			params := rest[2 : loc[1]-1]
			for _, p := range strings.Split(params, ";") {
				switch p {
				case "", "0", "39":
					fg = false
				}
			}
			if strings.HasPrefix(params, "38;") || strings.Contains(params, ";38;") {
				fg = true
			}
			rest = rest[loc[1]:]
			continue
		}
		var r rune
		for _, rr := range rest {
			r = rr
			break
		}
		if r != ' ' && !fg {
			t.Errorf("line %d: rune %q has no explicit foreground; line %q", lineNo, r, line)
			return
		}
		rest = rest[len(string(r)):]
	}
}

// Color must not change visible width: the colored render, stripped of SGR,
// equals the NoColor render byte for byte.
func TestRender_ColorEnabled_VisibleTextMatchesNoColor(t *testing.T) {
	t.Setenv("COLORTERM", "")
	colored := NewRenderer("catppuccin-mocha", false, nil).Render(colorProfileFixture(), ModeDefault)
	plain := NewRenderer("catppuccin-mocha", true, nil).Render(colorProfileFixture(), ModeDefault)
	if stripSGR(colored) != plain {
		t.Errorf("visible text diverged:\n colored(stripped)=%q\n plain=%q", stripSGR(colored), plain)
	}
}
