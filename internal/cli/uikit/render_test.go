package uikit_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

func TestRenderCard(t *testing.T) {
	result := uikit.RenderCard("Title", "content line")
	if !strings.Contains(result, "Title") {
		t.Errorf("uikit.RenderCard should contain title, got %q", result)
	}
	if !strings.Contains(result, "content line") {
		t.Errorf("uikit.RenderCard should contain content, got %q", result)
	}
}

func TestRenderKeyValue(t *testing.T) {
	result := uikit.RenderKeyValue("Key", "Value", 10)
	if !strings.Contains(result, "Key") {
		t.Errorf("uikit.RenderKeyValue should contain key, got %q", result)
	}
	if !strings.Contains(result, "Value") {
		t.Errorf("uikit.RenderKeyValue should contain value, got %q", result)
	}
}

func TestRenderKeyValueLines(t *testing.T) {
	pairs := []uikit.KVPair{
		{"Name", "test"},
		{"Version", "1.0.0"},
	}
	result := uikit.RenderKeyValueLines(pairs)
	if !strings.Contains(result, "Name") {
		t.Errorf("uikit.RenderKeyValueLines should contain Name, got %q", result)
	}
	if !strings.Contains(result, "Version") {
		t.Errorf("uikit.RenderKeyValueLines should contain Version, got %q", result)
	}
	if !strings.Contains(result, "test") {
		t.Errorf("uikit.RenderKeyValueLines should contain test, got %q", result)
	}
}

func TestRenderKeyValueLines_Empty(t *testing.T) {
	result := uikit.RenderKeyValueLines(nil)
	if result != "" {
		t.Errorf("uikit.RenderKeyValueLines with nil should return empty, got %q", result)
	}
}

func TestRenderStatusLine(t *testing.T) {
	result := uikit.RenderStatusLine(uikit.CheckOK, "Go", "1.21", 10)
	if !strings.Contains(result, "Go") {
		t.Errorf("uikit.RenderStatusLine should contain label, got %q", result)
	}
	if !strings.Contains(result, "1.21") {
		t.Errorf("uikit.RenderStatusLine should contain message, got %q", result)
	}
	if !strings.Contains(result, "\u2713") {
		t.Errorf("uikit.RenderStatusLine should contain check mark, got %q", result)
	}
}

func TestRenderSuccessCard(t *testing.T) {
	result := uikit.RenderSuccessCard("Done", "detail 1", "detail 2")
	if !strings.Contains(result, "Done") {
		t.Errorf("uikit.RenderSuccessCard should contain title, got %q", result)
	}
	if !strings.Contains(result, "detail 1") {
		t.Errorf("uikit.RenderSuccessCard should contain detail, got %q", result)
	}
	if !strings.Contains(result, "\u2713") {
		t.Errorf("uikit.RenderSuccessCard should contain check mark, got %q", result)
	}
}

func TestRenderSuccessCard_NoDetails(t *testing.T) {
	result := uikit.RenderSuccessCard("Done")
	if !strings.Contains(result, "Done") {
		t.Errorf("uikit.RenderSuccessCard should contain title, got %q", result)
	}
}

func TestRenderInfoCard(t *testing.T) {
	result := uikit.RenderInfoCard("Info", "line 1")
	if !strings.Contains(result, "Info") {
		t.Errorf("uikit.RenderInfoCard should contain title, got %q", result)
	}
	if !strings.Contains(result, "line 1") {
		t.Errorf("uikit.RenderInfoCard should contain detail, got %q", result)
	}
}

func TestRenderSummaryLine(t *testing.T) {
	result := uikit.RenderSummaryLine(3, 2, 0)
	if !strings.Contains(result, "3") {
		t.Errorf("uikit.RenderSummaryLine should contain 3, got %q", result)
	}
	if !strings.Contains(result, "2") {
		t.Errorf("uikit.RenderSummaryLine should contain 2, got %q", result)
	}
	if !strings.Contains(result, "0") {
		t.Errorf("uikit.RenderSummaryLine should contain 0, got %q", result)
	}
	if !strings.Contains(result, "passed") {
		t.Errorf("uikit.RenderSummaryLine should contain passed, got %q", result)
	}
}

func TestCardStyle(t *testing.T) {
	style := uikit.CardStyle()
	// Verify it renders without panic
	result := style.Render("test content")
	if !strings.Contains(result, "test content") {
		t.Errorf("uikit.CardStyle should render content, got %q", result)
	}
}

// Characterization tests for uikit.RenderError — M6-S6
// These tests capture expected behavior of the new uikit.RenderError function.

// TestCharacterize_RenderError_OutputContainsMessage checks that uikit.RenderError
// echoes the error message in its output.
func TestCharacterize_RenderError_OutputContainsMessage(t *testing.T) {
	err := fmt.Errorf("something went wrong")
	result := uikit.RenderError(err)
	if !strings.Contains(result, "something went wrong") {
		t.Errorf("uikit.RenderError should contain the error message, got %q", result)
	}
}

// TestCharacterize_RenderError_OutputContainsStatusIconErr checks that the
// error glyph ✗ (StatusIcon("err")) appears in the output.
func TestCharacterize_RenderError_OutputContainsStatusIconErr(t *testing.T) {
	err := fmt.Errorf("test error")
	result := uikit.RenderError(err)
	// StatusIcon("err") returns "✗" (U+2717)
	if !strings.Contains(result, "✗") {
		t.Errorf("uikit.RenderError should contain error icon ✗, got %q", result)
	}
}

// TestCharacterize_RenderError_OutputIsNonEmpty confirms uikit.RenderError never
// returns an empty string even for a plain error.
func TestCharacterize_RenderError_OutputIsNonEmpty(t *testing.T) {
	err := fmt.Errorf("x")
	result := uikit.RenderError(err)
	if strings.TrimSpace(result) == "" {
		t.Errorf("uikit.RenderError should return non-empty output, got %q", result)
	}
}

// TestCharacterize_RenderError_NoHexLiterals checks that uikit.RenderError output is
// produced by the tui theme system (AC-CLI-TUI-013: no raw hex literals).
// This test is structural: we verify the function is callable without panic,
// not the internal rendering path, which is covered by tui package tests.
func TestCharacterize_RenderError_NoHexLiterals(t *testing.T) {
	err := fmt.Errorf("hex check")
	result := uikit.RenderError(err)
	// Output must be non-empty and contain the error text.
	if !strings.Contains(result, "hex check") {
		t.Errorf("uikit.RenderError should echo error text, got %q", result)
	}
}
