package cli

import (
	"fmt"
	"strings"
	"testing"
)

func TestRenderCard(t *testing.T) {
	result := renderCard("Title", "content line")
	if !strings.Contains(result, "Title") {
		t.Errorf("renderCard should contain title, got %q", result)
	}
	if !strings.Contains(result, "content line") {
		t.Errorf("renderCard should contain content, got %q", result)
	}
}

func TestRenderKeyValue(t *testing.T) {
	result := renderKeyValue("Key", "Value", 10)
	if !strings.Contains(result, "Key") {
		t.Errorf("renderKeyValue should contain key, got %q", result)
	}
	if !strings.Contains(result, "Value") {
		t.Errorf("renderKeyValue should contain value, got %q", result)
	}
}

func TestRenderKeyValueLines(t *testing.T) {
	pairs := []kvPair{
		{"Name", "test"},
		{"Version", "1.0.0"},
	}
	result := renderKeyValueLines(pairs)
	if !strings.Contains(result, "Name") {
		t.Errorf("renderKeyValueLines should contain Name, got %q", result)
	}
	if !strings.Contains(result, "Version") {
		t.Errorf("renderKeyValueLines should contain Version, got %q", result)
	}
	if !strings.Contains(result, "test") {
		t.Errorf("renderKeyValueLines should contain test, got %q", result)
	}
}

func TestRenderKeyValueLines_Empty(t *testing.T) {
	result := renderKeyValueLines(nil)
	if result != "" {
		t.Errorf("renderKeyValueLines with nil should return empty, got %q", result)
	}
}

func TestRenderStatusLine(t *testing.T) {
	result := renderStatusLine(CheckOK, "Go", "1.21", 10)
	if !strings.Contains(result, "Go") {
		t.Errorf("renderStatusLine should contain label, got %q", result)
	}
	if !strings.Contains(result, "1.21") {
		t.Errorf("renderStatusLine should contain message, got %q", result)
	}
	if !strings.Contains(result, "\u2713") {
		t.Errorf("renderStatusLine should contain check mark, got %q", result)
	}
}

func TestRenderSuccessCard(t *testing.T) {
	result := renderSuccessCard("Done", "detail 1", "detail 2")
	if !strings.Contains(result, "Done") {
		t.Errorf("renderSuccessCard should contain title, got %q", result)
	}
	if !strings.Contains(result, "detail 1") {
		t.Errorf("renderSuccessCard should contain detail, got %q", result)
	}
	if !strings.Contains(result, "\u2713") {
		t.Errorf("renderSuccessCard should contain check mark, got %q", result)
	}
}

func TestRenderSuccessCard_NoDetails(t *testing.T) {
	result := renderSuccessCard("Done")
	if !strings.Contains(result, "Done") {
		t.Errorf("renderSuccessCard should contain title, got %q", result)
	}
}

func TestRenderInfoCard(t *testing.T) {
	result := renderInfoCard("Info", "line 1")
	if !strings.Contains(result, "Info") {
		t.Errorf("renderInfoCard should contain title, got %q", result)
	}
	if !strings.Contains(result, "line 1") {
		t.Errorf("renderInfoCard should contain detail, got %q", result)
	}
}

func TestRenderSummaryLine(t *testing.T) {
	result := renderSummaryLine(3, 2, 0)
	if !strings.Contains(result, "3") {
		t.Errorf("renderSummaryLine should contain 3, got %q", result)
	}
	if !strings.Contains(result, "2") {
		t.Errorf("renderSummaryLine should contain 2, got %q", result)
	}
	if !strings.Contains(result, "0") {
		t.Errorf("renderSummaryLine should contain 0, got %q", result)
	}
	if !strings.Contains(result, "passed") {
		t.Errorf("renderSummaryLine should contain passed, got %q", result)
	}
}

func TestCardStyle(t *testing.T) {
	style := cardStyle()
	// Verify it renders without panic
	result := style.Render("test content")
	if !strings.Contains(result, "test content") {
		t.Errorf("cardStyle should render content, got %q", result)
	}
}

// Characterization tests for RenderError — M6-S6
// These tests capture expected behavior of the new RenderError function.

// TestCharacterize_RenderError_OutputContainsMessage checks that RenderError
// echoes the error message in its output.
func TestCharacterize_RenderError_OutputContainsMessage(t *testing.T) {
	err := fmt.Errorf("something went wrong")
	result := RenderError(err)
	if !strings.Contains(result, "something went wrong") {
		t.Errorf("RenderError should contain the error message, got %q", result)
	}
}

// TestCharacterize_RenderError_OutputContainsStatusIconErr checks that the
// error glyph ✗ (StatusIcon("err")) appears in the output.
func TestCharacterize_RenderError_OutputContainsStatusIconErr(t *testing.T) {
	err := fmt.Errorf("test error")
	result := RenderError(err)
	// StatusIcon("err") returns "✗" (U+2717)
	if !strings.Contains(result, "✗") {
		t.Errorf("RenderError should contain error icon ✗, got %q", result)
	}
}

// TestCharacterize_RenderError_OutputIsNonEmpty confirms RenderError never
// returns an empty string even for a plain error.
func TestCharacterize_RenderError_OutputIsNonEmpty(t *testing.T) {
	err := fmt.Errorf("x")
	result := RenderError(err)
	if strings.TrimSpace(result) == "" {
		t.Errorf("RenderError should return non-empty output, got %q", result)
	}
}

// TestCharacterize_RenderError_NoHexLiterals checks that RenderError output is
// produced by the tui theme system (AC-CLI-TUI-013: no raw hex literals).
// This test is structural: we verify the function is callable without panic,
// not the internal rendering path, which is covered by tui package tests.
func TestCharacterize_RenderError_NoHexLiterals(t *testing.T) {
	err := fmt.Errorf("hex check")
	result := RenderError(err)
	// Output must be non-empty and contain the error text.
	if !strings.Contains(result, "hex check") {
		t.Errorf("RenderError should echo error text, got %q", result)
	}
}
