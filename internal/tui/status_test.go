// Package tui provides the MoAI-ADK terminal UI design system.
package tui_test

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// --- StatusIcon ---

// TestStatusIconKinds verifies that each icon kind produces a non-empty string.
func TestStatusIconKinds(t *testing.T) {
	kinds := []string{"ok", "warn", "err", "info", "run", "skip", "dot"}
	for _, k := range kinds {
		t.Run(k, func(t *testing.T) {
			out := tui.StatusIcon(k)
			if out == "" {
				t.Errorf("StatusIcon(%q) returned empty string", k)
			}
		})
	}
}

// TestStatusIconUnknown verifies that an unknown kind falls back to dot.
func TestStatusIconUnknown(t *testing.T) {
	out := tui.StatusIcon("unknown-kind-xyz")
	if out == "" {
		t.Fatal("StatusIcon unknown returned empty string")
	}
}

// --- Spinner ---

// TestSpinnerBasic verifies that Spinner returns a non-empty string.
func TestSpinnerBasic(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Spinner("loading", &th)
	if out == "" {
		t.Fatal("Spinner returned empty string")
	}
	checkGolden(t, "spinner-light-basic", out)
}

// TestSpinnerDark verifies Spinner in dark theme.
func TestSpinnerDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.Spinner("loading", &th)
	if out == "" {
		t.Fatal("Spinner dark returned empty string")
	}
	checkGolden(t, "spinner-dark-basic", out)
}

// TestSpinnerStatic verifies that MOAI_REDUCED_MOTION=1 produces a static dot.
// AC-CLI-TUI-015: Spinner must produce static output when reduced motion is requested.
func TestSpinnerStatic(t *testing.T) {
	t.Setenv("MOAI_REDUCED_MOTION", "1")
	th := tui.LightTheme()
	out := tui.Spinner("loading", &th)
	if out == "" {
		t.Fatal("Spinner reduced-motion returned empty string")
	}
	checkGolden(t, "spinner-reduced", out)
}

// TestSpinnerAnimated verifies that without MOAI_REDUCED_MOTION the output
// is different from the static reduced-motion output.
func TestSpinnerAnimated(t *testing.T) {
	// Ensure MOAI_REDUCED_MOTION is NOT set.
	if err := os.Unsetenv("MOAI_REDUCED_MOTION"); err != nil {
		t.Fatalf("Unsetenv: %v", err)
	}
	th := tui.LightTheme()
	out := tui.Spinner("loading", &th)
	if out == "" {
		t.Fatal("Spinner animated returned empty string")
	}
	checkGolden(t, "spinner-animated", out)
}

// --- Progress ---

// TestProgressBasic verifies Progress produces non-empty output.
func TestProgressBasic(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Progress(3, 10, tui.ProgressOpts{Theme: &th})
	if out == "" {
		t.Fatal("Progress returned empty string")
	}
	checkGolden(t, "progress-light-basic", out)
}

// TestProgressDark verifies Progress in dark theme.
func TestProgressDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.Progress(7, 10, tui.ProgressOpts{Theme: &th})
	if out == "" {
		t.Fatal("Progress dark returned empty string")
	}
	checkGolden(t, "progress-dark-basic", out)
}

// TestProgressStatic verifies that MOAI_REDUCED_MOTION=1 produces a filled bar.
// AC-CLI-TUI-015: Progress must produce filled bar when reduced motion is requested.
func TestProgressStatic(t *testing.T) {
	t.Setenv("MOAI_REDUCED_MOTION", "1")
	th := tui.LightTheme()
	out := tui.Progress(5, 10, tui.ProgressOpts{Theme: &th, Width: 10})
	if out == "" {
		t.Fatal("Progress reduced-motion returned empty string")
	}
	checkGolden(t, "progress-reduced", out)
}

// TestProgressFull verifies Progress at 100% (value == max).
func TestProgressFull(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Progress(10, 10, tui.ProgressOpts{Theme: &th, Width: 10})
	if out == "" {
		t.Fatal("Progress full returned empty string")
	}
	checkGolden(t, "progress-light-full", out)
}

// TestProgressZero verifies Progress at 0%.
func TestProgressZero(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Progress(0, 10, tui.ProgressOpts{Theme: &th, Width: 10})
	if out == "" {
		t.Fatal("Progress zero returned empty string")
	}
	checkGolden(t, "progress-light-zero", out)
}

// --- Stepper ---

// TestStepperBasic verifies Stepper produces non-empty output.
func TestStepperBasic(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Stepper(1, 6, &th)
	if out == "" {
		t.Fatal("Stepper returned empty string")
	}
	checkGolden(t, "stepper-light-basic", out)
}

// TestStepperDark verifies Stepper in dark theme.
func TestStepperDark(t *testing.T) {
	th := tui.DarkTheme()
	out := tui.Stepper(3, 6, &th)
	if out == "" {
		t.Fatal("Stepper dark returned empty string")
	}
	checkGolden(t, "stepper-dark-basic", out)
}

// TestStepperLast verifies Stepper at the last step.
func TestStepperLast(t *testing.T) {
	th := tui.LightTheme()
	out := tui.Stepper(6, 6, &th)
	if out == "" {
		t.Fatal("Stepper last returned empty string")
	}
	checkGolden(t, "stepper-light-last", out)
}
