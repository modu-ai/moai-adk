package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

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
		PrintBanner("1.2.3")
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify output contains expected strings
	expectedStrings := []string{
		"MoAI",        // Banner should contain MoAI
		"Version",     // Version label
		"1.2.3",       // Actual version
		"Agentic",     // Description text
		"Development", // Description text
		"Kit",         // Description text
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("PrintBanner output should contain %q, got:\n%s", expected, output)
		}
	}

	// Verify output is not empty
	if len(output) == 0 {
		t.Error("PrintBanner should produce output")
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
				PrintBanner(tt.version)
			})
			if err != nil {
				t.Fatal(err)
			}

			// Verify version is in output
			if !strings.Contains(output, tt.version) {
				t.Errorf("PrintBanner should contain version %q, got:\n%s", tt.version, output)
			}
		})
	}
}

// TestPrintBanner_EmptyVersion verifies banner handles empty version gracefully.
func TestPrintBanner_EmptyVersion(t *testing.T) {
	output, err := captureStdout(func() {
		PrintBanner("")
	})
	if err != nil {
		t.Fatal(err)
	}

	// Should still produce output (banner and description)
	if len(output) == 0 {
		t.Error("PrintBanner with empty version should still produce output")
	}

	// Should contain MoAI branding
	if !strings.Contains(output, "MoAI") {
		t.Error("PrintBanner should contain MoAI branding")
	}
}
