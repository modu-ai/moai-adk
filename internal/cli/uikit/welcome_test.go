package uikit_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
)

// --- uikit.PrintWelcomeMessage tests ---

func TestPrintWelcomeMessage_OutputFormat(t *testing.T) {
	output, err := captureStdout(func() {
		uikit.PrintWelcomeMessage()
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(output) == 0 {
		t.Error("uikit.PrintWelcomeMessage should produce output")
	}

	expectedStrings := []string{
		"Welcome",
		"MoAI-ADK",
		"Initialization",
		"Ctrl+C",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("uikit.PrintWelcomeMessage output should contain %q, got:\n%s", expected, output)
		}
	}
}

func TestPrintWelcomeMessage_MultipleCallsConsistent(t *testing.T) {
	output1, err := captureStdout(func() {
		uikit.PrintWelcomeMessage()
	})
	if err != nil {
		t.Fatal(err)
	}

	output2, err := captureStdout(func() {
		uikit.PrintWelcomeMessage()
	})
	if err != nil {
		t.Fatal(err)
	}

	if output1 != output2 {
		t.Error("uikit.PrintWelcomeMessage should produce consistent output across calls")
	}
}
