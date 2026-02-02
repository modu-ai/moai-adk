package output

import (
	"testing"
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
	styles := map[string]func(string) string{
		"BannerStyle":     BannerStyle.Render,
		"VersionStyle":    VersionStyle.Render,
		"SubtitleStyle":   SubtitleStyle.Render,
		"WelcomeStyle":    WelcomeStyle.Render,
		"WizardInfoStyle": WizardInfoStyle.Render,
	}

	for name, renderFn := range styles {
		t.Run(name, func(t *testing.T) {
			result := renderFn("test")
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
