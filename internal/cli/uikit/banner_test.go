package uikit_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// updateBannerGolden controls golden snapshot regeneration. Set via UPDATE_GOLDEN=1.
var updateBannerGolden = os.Getenv("UPDATE_GOLDEN") == "1"

// bannerGoldenPath returns the path to a golden snapshot file under testdata/.
func bannerGoldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

// checkBannerGolden compares got to a golden file, regenerating it if UPDATE_GOLDEN=1.
func checkBannerGolden(t *testing.T, name, got string) {
	t.Helper()
	path := bannerGoldenPath(name)
	if updateBannerGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		t.Logf("updated golden: %s", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with UPDATE_GOLDEN=1 to generate)", path, err)
	}
	if got != string(want) {
		t.Errorf("banner output mismatch for %s\ngot:\n%s\nwant:\n%s", name, got, string(want))
	}
}

// captureStdout captures stdout during function execution.
// Returns captured output string and any error encountered.
func captureStdout(fn func()) (string, error) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	fn()

	// Close writer to signal end of capture
	if err := w.Close(); err != nil {
		os.Stdout = old
		return "", err
	}
	os.Stdout = old

	// Read all captured output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	// Close reader to release resources
	if err := r.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// --- DDD PRESERVE: Characterization tests for banner functions ---

// TestPrintBanner_OutputFormat verifies the banner output contains expected elements.
func TestPrintBanner_OutputFormat(t *testing.T) {
	output, err := captureStdout(func() {
		uikit.PrintBanner("1.2.3")
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify output contains expected strings.
	// SPEC-CLI-TUX-V3-004 M4d: the "Version:" label line is retired with the
	// large ASCII logo — the version now rides the pill row (v1.2.3).
	// SPEC-CLI-TUX-INIT-UPDATE-001 M3 (logo-aware): the large logo is RESTORED and
	// stacks above the compact band, so the composed PrintBanner surface now also
	// carries the logo's first-row signature (███╗   ███╗) — REQ-TUXIU-050/054.
	expectedStrings := []string{
		"███╗   ███╗",  // Restored logo signature (stacked above the compact band)
		"MoAI",         // Banner should contain MoAI
		"v1.2.3",       // Version pill
		"Agentic",      // Description text
		"Development",  // Description text
		"Kit",          // Description text
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("uikit.PrintBanner output should contain %q, got:\n%s", expected, output)
		}
	}

	// Verify output is not empty
	if len(output) == 0 {
		t.Error("uikit.PrintBanner should produce output")
	}
}

// TestPrintBanner_WithVersion verifies banner displays version correctly.
func TestPrintBanner_WithVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "normal version",
			version: "1.0.0",
		},
		{
			name:    "dev version",
			version: "1.0.0-dev",
		},
		{
			name:    "long version",
			version: "1.2.3-beta.1+build.20240101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := captureStdout(func() {
				uikit.PrintBanner(tt.version)
			})
			if err != nil {
				t.Fatal(err)
			}

			// Verify version is in output
			if !strings.Contains(output, tt.version) {
				t.Errorf("uikit.PrintBanner should contain version %q, got:\n%s", tt.version, output)
			}
		})
	}
}

// TestPrintBanner_OptionalLeadingV is the regression guard for the vv3.0.0
// banner bug: version.GetVersion() returns "v3.0.0" (already v-prefixed) and the
// pill previously prepended a second "v", rendering "vv3.0.0". The pill must
// normalize an optional leading "v" — both a prefixed ("v3.0.0") and a bare
// ("3.0.0") version yield exactly a single-"v" pill ("v3.0.0"), never "vv3.0.0".
func TestPrintBanner_OptionalLeadingV(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"prefixed version (v3.0.0)", "v3.0.0"},
		{"bare version (3.0.0)", "3.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := captureStdout(func() {
				uikit.PrintBanner(tt.input)
			})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output, "v3.0.0") {
				t.Errorf("banner should contain single-v %q, got:\n%s", "v3.0.0", output)
			}
			if strings.Contains(output, "vv3.0.0") {
				t.Errorf("banner must NOT contain doubled-v %q, got:\n%s", "vv3.0.0", output)
			}
		})
	}
}

// TestPrintBanner_EmptyVersion verifies banner handles empty version gracefully.
func TestPrintBanner_EmptyVersion(t *testing.T) {
	output, err := captureStdout(func() {
		uikit.PrintBanner("")
	})
	if err != nil {
		t.Fatal(err)
	}

	// Should still produce output (banner and description)
	if len(output) == 0 {
		t.Error("uikit.PrintBanner with empty version should still produce output")
	}

	// Should contain MoAI branding
	if !strings.Contains(output, "MoAI") {
		t.Error("uikit.PrintBanner should contain MoAI branding")
	}
}

// --- DDD PRESERVE Phase: Golden-snapshot characterization tests ---
//
// These tests capture the BEFORE state of uikit.PrintBanner / uikit.PrintWelcomeMessage output
// (terra cotta / purple lipgloss styles) and serve as the regression baseline for
// Step 2 DDD IMPROVE, which replaces the body with tui-derived rendering.
//
// lipgloss AdaptiveColor behaviour under os.Pipe() stdout capture:
//   - lipgloss detects non-TTY stdout and disables ANSI colour output.
//   - Therefore NO_COLOR=0/1 and MOAI_THEME=light/dark produce identical byte output
//     when captured via os.Pipe(). The env vars are set to document intent and to
//     remain robust if banner.go is later extended to honour MOAI_THEME explicitly.
//   - Each env combination receives its own golden file for clarity.
//
// To regenerate snapshots: UPDATE_GOLDEN=1 go test ./internal/cli/ -run "TestBanner_Current|TestWelcome_Current" -count=1

// TestBanner_Current_Light captures uikit.PrintBanner output with light-theme env.
// Characteristics: Claude coral Accent color (tui.Theme().Accent), MoAI ASCII art banner + 3 tui.Pill.
// Note: Go version is embedded in the golden snapshot — re-run UPDATE_GOLDEN=1 when Go toolchain updates.
func TestBanner_Current_Light(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("MOAI_THEME", "light")
	t.Setenv("CLAUDE_CODE_VERSION", "1.0.18")      // pinned for deterministic golden
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.26.0") // pin Go version for cross-toolchain deterministic golden

	got, err := captureStdout(func() {
		uikit.PrintBanner("1.0.0")
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("uikit.PrintBanner produced no output")
	}
	checkBannerGolden(t, "banner-current-light", got)
}

// TestBanner_Current_Dark captures uikit.PrintBanner output with dark-theme env.
// Characteristics: Claude coral Accent color (tui.DarkTheme().Accent), MoAI ASCII art banner + 3 tui.Pill.
// Note: Go version is embedded in the golden snapshot — re-run UPDATE_GOLDEN=1 when Go toolchain updates.
func TestBanner_Current_Dark(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("MOAI_THEME", "dark")
	t.Setenv("CLAUDE_CODE_VERSION", "1.0.18")      // pinned for deterministic golden
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.26.0") // pin Go version for cross-toolchain deterministic golden

	got, err := captureStdout(func() {
		uikit.PrintBanner("1.0.0")
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("uikit.PrintBanner produced no output")
	}
	checkBannerGolden(t, "banner-current-dark", got)
}

// TestBanner_NoColor captures uikit.PrintBanner output with NO_COLOR=1 (no ANSI escape).
// tui.MonochromeTheme() is used; all colours are empty; Pill degrades to [label] plain text.
func TestBanner_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("CLAUDE_CODE_VERSION", "1.0.18")      // pinned for deterministic golden
	t.Setenv("MOAI_GO_VERSION_OVERRIDE", "1.26.0") // pin Go version for cross-toolchain deterministic golden

	got, err := captureStdout(func() {
		uikit.PrintBanner("1.0.0")
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("uikit.PrintBanner produced no output")
	}
	checkBannerGolden(t, "banner-current-nocolor", got)
}

// TestWelcome_Current_Light captures uikit.PrintWelcomeMessage output with light-theme env.
// Characteristics: Claude coral Accent color (tui.LightTheme().Accent), bold title.
func TestWelcome_Current_Light(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("MOAI_THEME", "light")

	got, err := captureStdout(func() {
		uikit.PrintWelcomeMessage()
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("uikit.PrintWelcomeMessage produced no output")
	}
	checkBannerGolden(t, "welcome-current-light", got)
}

// TestWelcome_Current_Dark captures uikit.PrintWelcomeMessage output with dark-theme env.
// Characteristics: Claude coral Accent color (tui.DarkTheme().Accent), bold title.
func TestWelcome_Current_Dark(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("MOAI_THEME", "dark")

	got, err := captureStdout(func() {
		uikit.PrintWelcomeMessage()
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("uikit.PrintWelcomeMessage produced no output")
	}
	checkBannerGolden(t, "welcome-current-dark", got)
}

// TestWelcome_NoColor captures uikit.PrintWelcomeMessage output with NO_COLOR=1.
// tui.MonochromeTheme() is used; all colours and Bold are suppressed; output is plain text only.
func TestWelcome_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got, err := captureStdout(func() {
		uikit.PrintWelcomeMessage()
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("uikit.PrintWelcomeMessage produced no output")
	}
	checkBannerGolden(t, "welcome-current-nocolor", got)
}
