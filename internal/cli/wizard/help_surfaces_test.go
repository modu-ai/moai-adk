package wizard

// AC-ITI-013 (REQ-ITI-012) — M7 half: three of the four surfaces (init
// wizard, reconfigure wizard, downgrade confirm) drawn per locale match their
// goldens, and each ko/ja/zh screen shows NO English help action label. The
// fourth surface (the profile wizard) is judged in the cli package where its
// option-label bridge lives (TestProfileWizardGolden_* + NoEnglishLeak).

import (
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/ptycaptest"
)

// enActionLabels are the design.md §7 English action labels that must NOT
// appear on a ko/ja/zh help line.
var enActionLabels = []string{"next", "submit", "back", "select", "up", "down", "filter", "toggle"}

// drawHelpSurface renders one surface's first page in the given locale.
func drawHelpSurface(t *testing.T, surface, locale string) string {
	t.Helper()
	var frame string
	switch surface {
	case "init":
		form := buildUnifiedForm(InitQuestions("/tmp/help-surfaces"), &WizardResult{}, locale)
		frame = ptycaptest.NewFormDriver(t, form).View()
	case "reconfigure":
		form := buildUnifiedForm(ReconfigureQuestions("/tmp/help-surfaces"), &WizardResult{}, locale)
		frame = ptycaptest.NewFormDriver(t, form).View()
	case "downgrade":
		var v bool
		form := NewDowngradeConfirmForm(locale, "v9.9.9", "v1.0.0", &v)
		frame = ptycaptest.NewFormDriver(t, form).View()
	default:
		t.Fatalf("unknown surface %q", surface)
	}
	return ptycaptest.StripANSI(frame)
}

// TestHelpLabels_LocaleGoldenSurfaces draws the three wizard-package surfaces
// in all four locales against saved goldens, and asserts the ko/ja/zh help
// lines carry no English action labels.
func TestHelpLabels_LocaleGoldenSurfaces(t *testing.T) {
	for _, surface := range []string{"init", "reconfigure", "downgrade"} {
		for _, locale := range []string{"en", "ko", "ja", "zh"} {
			name := surface + "-" + locale
			t.Run(name, func(t *testing.T) {
				frame := drawHelpSurface(t, surface, locale)
				if err := ptycaptest.CompareGolden("testdata/help-surfaces", name, frame, *updateAxisGolden); err != nil {
					t.Fatal(err)
				}
				if locale == "en" {
					return
				}
				for _, en := range enActionLabels {
					if slices.Contains(enActionLabels, en) && strings.Contains(frame, en) {
						t.Errorf("%s %s screen leaks the English help action %q", surface, locale, en)
					}
				}
			})
		}
	}
}
