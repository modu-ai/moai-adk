package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/version"
)

// updateVersionGolden controls golden snapshot regeneration. Set via UPDATE_GOLDEN=1.
var updateVersionGolden = os.Getenv("UPDATE_GOLDEN") == "1"

// versionGoldenPath returns the path to a version golden snapshot file under testdata/.
func versionGoldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

// checkVersionGolden compares got to a golden file, regenerating it if UPDATE_GOLDEN=1.
func checkVersionGolden(t *testing.T, name, got string) {
	t.Helper()
	path := versionGoldenPath(name)
	if updateVersionGolden {
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
		t.Errorf("version output mismatch for %s\ngot:\n%s\nwant:\n%s", name, got, string(want))
	}
}

// captureVersionCmd executes versionCmd and returns stdout as a string.
// version, commit, date are pinned via pkg/version variables for deterministic output.
func captureVersionCmd(t *testing.T) string {
	t.Helper()
	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)
	versionCmd.SetErr(buf)
	if err := versionCmd.RunE(versionCmd, []string{}); err != nil {
		t.Fatalf("versionCmd.RunE: %v", err)
	}
	// Reset cobra output writers after capture to avoid test pollution.
	versionCmd.SetOut(nil)
	versionCmd.SetErr(nil)
	return buf.String()
}

// --- DDD PRESERVE: Characterization tests for version command behavior ---

func TestVersionCmd_Exists(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd should not be nil")
	}
}

func TestVersionCmd_Use(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("versionCmd.Use = %q, want %q", versionCmd.Use, "version")
	}
}

func TestVersionCmd_Short(t *testing.T) {
	if versionCmd.Short == "" {
		t.Error("versionCmd.Short should not be empty")
	}
}

func TestVersionCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("version should be registered as a subcommand of root")
	}
}

func TestVersionCmd_OutputContainsBranding(t *testing.T) {
	// Pin build-time variables for deterministic output.
	orig, origC, origD := version.Version, version.Commit, version.Date
	version.Version = "v0.0.0-test"
	version.Commit = "abc1234"
	version.Date = "2026-01-01"
	defer func() {
		version.Version = orig
		version.Commit = origC
		version.Date = origD
	}()

	output := captureVersionCmd(t)

	// moai-adk branding must appear in Box title.
	if !strings.Contains(output, "moai-adk") {
		t.Errorf("version output should contain 'moai-adk', got %q", output)
	}
	// Version string must appear (version.Version already contains "v" prefix).
	if !strings.Contains(output, "v0.0.0-test") {
		t.Errorf("version output should contain 'v0.0.0-test', got %q", output)
	}
	// Commit (shortened) must appear.
	if !strings.Contains(output, "abc1234") {
		t.Errorf("version output should contain commit hash, got %q", output)
	}
	// Built date must appear.
	if !strings.Contains(output, "built 2026-01-01") {
		t.Errorf("version output should contain 'built 2026-01-01', got %q", output)
	}
}

func TestVersionCmd_HasRunE(t *testing.T) {
	if versionCmd.RunE == nil {
		t.Error("versionCmd.RunE should not be nil")
	}
}

// --- shortCommit characterization tests (doctor.go shared helper) ---
// shortCommit is defined in doctor.go and shared across the cli package.
// These tests characterize its behavior as used by version rendering.

func TestShortCommit_NonePassthrough(t *testing.T) {
	// "none" is shorter than 9 chars, so passthrough.
	if got := shortCommit("none"); got != "none" {
		t.Errorf("shortCommit(\"none\") = %q, want %q", got, "none")
	}
}

func TestShortCommit_ShortPassthrough(t *testing.T) {
	// strings shorter than 9 chars pass through unchanged.
	if got := shortCommit("abc123"); got != "abc123" {
		t.Errorf("shortCommit(\"abc123\") = %q, want %q", got, "abc123")
	}
}

func TestShortCommit_Truncate9(t *testing.T) {
	// strings >= 9 chars are truncated to first 9.
	if got := shortCommit("abcdef1234567890"); got != "abcdef123" {
		t.Errorf("shortCommit(\"abcdef1234567890\") = %q, want %q", got, "abcdef123")
	}
}

func TestShortCommit_ExactlyNinePassthrough(t *testing.T) {
	if got := shortCommit("abcdef123"); got != "abcdef123" {
		t.Errorf("shortCommit(\"abcdef123\") = %q, want %q", got, "abcdef123")
	}
}

// --- DDD IMPROVE: Golden-snapshot characterization tests ---
//
// These tests capture the AFTER state of versionCmd output (tui.Box + tui.Pill rendering).
// They serve as the regression baseline for future DDD cycles.
//
// lipgloss AdaptiveColor behaviour under cmd.SetOut(buf):
//   - lipgloss detects non-TTY writer and disables ANSI colour output.
//   - Therefore NO_COLOR=0/1 and MOAI_THEME=light/dark produce identical ANSI-stripped output.
//   - Each env combination receives its own golden file for clarity and future-proofing.
//
// To regenerate snapshots:
//   UPDATE_GOLDEN=1 go test ./internal/cli/ -run "TestVersion_Current" -count=1

// TestVersion_Current_Light captures versionCmd output with light-theme env.
func TestVersion_Current_Light(t *testing.T) {
	t.Setenv("NO_COLOR", "0")
	t.Setenv("MOAI_THEME", "light")

	// Pin build-time variables for deterministic golden output.
	orig, origC, origD := version.Version, version.Commit, version.Date
	version.Version = "v0.0.0-test"
	version.Commit = "abc1234def567890"
	version.Date = "2026-01-01"
	defer func() {
		version.Version = orig
		version.Commit = origC
		version.Date = origD
	}()

	got := captureVersionCmd(t)
	if len(got) == 0 {
		t.Fatal("versionCmd produced no output")
	}
	checkVersionGolden(t, "version-light", got)
}

// TestVersion_Current_Dark captures versionCmd output with dark-theme env.
func TestVersion_Current_Dark(t *testing.T) {
	t.Setenv("NO_COLOR", "0")
	t.Setenv("MOAI_THEME", "dark")

	orig, origC, origD := version.Version, version.Commit, version.Date
	version.Version = "v0.0.0-test"
	version.Commit = "abc1234def567890"
	version.Date = "2026-01-01"
	defer func() {
		version.Version = orig
		version.Commit = origC
		version.Date = origD
	}()

	got := captureVersionCmd(t)
	if len(got) == 0 {
		t.Fatal("versionCmd produced no output")
	}
	checkVersionGolden(t, "version-dark", got)
}

// TestVersion_NoColor captures versionCmd output with NO_COLOR=1.
// tui.MonochromeTheme() is used; Pill degrades to [label] plain text.
func TestVersion_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	orig, origC, origD := version.Version, version.Commit, version.Date
	version.Version = "v0.0.0-test"
	version.Commit = "abc1234def567890"
	version.Date = "2026-01-01"
	defer func() {
		version.Version = orig
		version.Commit = origC
		version.Date = origD
	}()

	got := captureVersionCmd(t)
	if len(got) == 0 {
		t.Fatal("versionCmd produced no output")
	}
	checkVersionGolden(t, "version-nocolor", got)
}
