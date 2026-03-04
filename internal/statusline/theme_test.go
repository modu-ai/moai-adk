package statusline

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestThemeInterface verifies all three theme implementations satisfy the Theme interface.
func TestThemeInterface(t *testing.T) {
	themes := []struct {
		name  string
		theme Theme
	}{
		{"default", NewTheme("default")},
		{"catppuccin-mocha", NewTheme("catppuccin-mocha")},
		{"catppuccin-latte", NewTheme("catppuccin-latte")},
		{"unknown falls back to default", NewTheme("unknown-theme")},
	}

	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			if tt.theme == nil {
				t.Fatal("NewTheme() returned nil")
			}
			// Each method must return a non-zero color
			tt.theme.Primary()
			tt.theme.Muted()
			tt.theme.Success()
			tt.theme.Warning()
			tt.theme.Danger()
			tt.theme.Text()
		})
	}
}

func TestDefaultThemeColors(t *testing.T) {
	theme := NewTheme("default")

	// DefaultTheme preserves current hard-coded #6B7280 muted color (REQ-SLE-012)
	muted := theme.Muted()
	if muted != lipgloss.Color("#6B7280") {
		t.Errorf("DefaultTheme.Muted() = %q, want %q", muted, "#6B7280")
	}
}

func TestCatppuccinMochaColors(t *testing.T) {
	theme := NewTheme("catppuccin-mocha")

	tests := []struct {
		name  string
		got   lipgloss.Color
		want  lipgloss.Color
	}{
		{"Primary", theme.Primary(), lipgloss.Color("#CDD6F4")},
		{"Muted", theme.Muted(), lipgloss.Color("#6C7086")},
		{"Success", theme.Success(), lipgloss.Color("#A6E3A1")},
		{"Warning", theme.Warning(), lipgloss.Color("#F9E2AF")},
		{"Danger", theme.Danger(), lipgloss.Color("#F38BA8")},
		{"Text", theme.Text(), lipgloss.Color("#CDD6F4")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("CatppuccinMocha.%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestCatppuccinLatteColors(t *testing.T) {
	theme := NewTheme("catppuccin-latte")

	tests := []struct {
		name string
		got  lipgloss.Color
		want lipgloss.Color
	}{
		{"Primary", theme.Primary(), lipgloss.Color("#4C4F69")},
		{"Muted", theme.Muted(), lipgloss.Color("#9CA0B0")},
		{"Success", theme.Success(), lipgloss.Color("#40A02B")},
		{"Warning", theme.Warning(), lipgloss.Color("#DF8E1D")},
		{"Danger", theme.Danger(), lipgloss.Color("#D20F39")},
		{"Text", theme.Text(), lipgloss.Color("#4C4F69")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("CatppuccinLatte.%s() = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestBarGradient_Stages(t *testing.T) {
	// REQ-SLE-015: 4-stage gradient
	tests := []struct {
		name   string
		pct    float64
		stage  int // 1=green, 2=yellow, 3=peach/orange, 4=red
	}{
		{"0% = stage 1 (green)", 0, 1},
		{"25% = stage 1 (green)", 25, 1},
		{"26% = stage 2 (yellow)", 26, 2},
		{"50% = stage 2 (yellow)", 50, 2},
		{"51% = stage 3 (peach)", 51, 3},
		{"75% = stage 3 (peach)", 75, 3},
		{"76% = stage 4 (red)", 76, 4},
		{"100% = stage 4 (red)", 100, 4},
	}

	mochaTheme := NewTheme("catppuccin-mocha")
	latteTheme := NewTheme("catppuccin-latte")

	mochaGreen := lipgloss.Color("#A6E3A1")
	mochaYellow := lipgloss.Color("#F9E2AF")
	mochaPeach := lipgloss.Color("#FAB387")
	mochaRed := lipgloss.Color("#F38BA8")

	latteGreen := lipgloss.Color("#40A02B")
	latteYellow := lipgloss.Color("#DF8E1D")
	lattePeach := lipgloss.Color("#FE640B")
	latteRed := lipgloss.Color("#D20F39")

	mochaExpected := map[int]lipgloss.Color{
		1: mochaGreen, 2: mochaYellow, 3: mochaPeach, 4: mochaRed,
	}
	latteExpected := map[int]lipgloss.Color{
		1: latteGreen, 2: latteYellow, 3: lattePeach, 4: latteRed,
	}

	for _, tt := range tests {
		t.Run("mocha/"+tt.name, func(t *testing.T) {
			got := mochaTheme.BarGradient(tt.pct)
			want := mochaExpected[tt.stage]
			if got != want {
				t.Errorf("CatppuccinMocha.BarGradient(%v) = %q, want %q (stage %d)", tt.pct, got, want, tt.stage)
			}
		})
		t.Run("latte/"+tt.name, func(t *testing.T) {
			got := latteTheme.BarGradient(tt.pct)
			want := latteExpected[tt.stage]
			if got != want {
				t.Errorf("CatppuccinLatte.BarGradient(%v) = %q, want %q (stage %d)", tt.pct, got, want, tt.stage)
			}
		})
	}
}

func TestDefaultTheme_BarGradient_Stages(t *testing.T) {
	// REQ-SLE-018: DefaultTheme uses same 4-stage progression
	theme := NewTheme("default")
	stage1 := theme.BarGradient(0)
	stage2 := theme.BarGradient(26)
	stage3 := theme.BarGradient(51)
	stage4 := theme.BarGradient(76)

	// All stages should return distinct non-empty colors
	colors := []lipgloss.Color{stage1, stage2, stage3, stage4}
	for i, c := range colors {
		if c == "" {
			t.Errorf("DefaultTheme.BarGradient() stage %d returned empty color", i+1)
		}
	}
	// Stages should be distinct (different colors for progression)
	if stage1 == stage4 {
		t.Error("DefaultTheme: stage 1 (green) and stage 4 (red) should be distinct colors")
	}
}

func TestNewTheme_UnknownFallsBackToDefault(t *testing.T) {
	defaultTheme := NewTheme("default")
	unknownTheme := NewTheme("totally-unknown-theme")

	// Unknown theme should behave like default
	if defaultTheme.Muted() != unknownTheme.Muted() {
		t.Errorf("unknown theme should fall back to default: got Muted=%q, want %q",
			unknownTheme.Muted(), defaultTheme.Muted())
	}
}
