package output

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// --- Constants ---

func TestColorTerraCotta(t *testing.T) {
	if ColorTerraCotta == "" {
		t.Error("ColorTerraCotta is empty")
	}
	if ColorTerraCotta[0] != '#' {
		t.Errorf("ColorTerraCotta = %q, expected hex color starting with #", ColorTerraCotta)
	}
}

func TestMoaiBanner(t *testing.T) {
	if MoaiBanner == "" {
		t.Error("MoaiBanner is empty")
	}
	// Banner should contain MoAI text elements
	if len(MoaiBanner) < 50 {
		t.Error("MoaiBanner seems too short to be valid ASCII art")
	}
}

// --- Banner styles ---

func TestBannerStyles(t *testing.T) {
	styles := map[string]lipgloss.Style{
		"BannerStyle":     BannerStyle,
		"VersionStyle":    VersionStyle,
		"SubtitleStyle":   SubtitleStyle,
		"WelcomeStyle":    WelcomeStyle,
		"WizardInfoStyle": WizardInfoStyle,
	}

	for name, style := range styles {
		t.Run(name, func(t *testing.T) {
			result := style.Render("test")
			if result == "" {
				t.Errorf("%s.Render returned empty string", name)
			}
		})
	}
}

// --- PrintBanner ---

func TestPrintBanner_DoesNotPanic(t *testing.T) {
	// PrintBanner writes to stdout; verify it does not panic
	PrintBanner("1.0.0")
}

func TestPrintBanner_DifferentVersions(t *testing.T) {
	versions := []string{"dev", "1.0.0", "2.5.3-beta", ""}
	for _, v := range versions {
		// Should not panic with any version string
		PrintBanner(v)
	}
}

// --- PrintWelcomeMessage ---

func TestPrintWelcomeMessage_DoesNotPanic(t *testing.T) {
	// PrintWelcomeMessage writes to stdout; verify it does not panic
	PrintWelcomeMessage()
}
