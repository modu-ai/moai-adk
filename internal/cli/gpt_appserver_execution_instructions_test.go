package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/codextools"
)

func TestManagedGPTAsyncAgentGuidancePreservesCallerInstructions(t *testing.T) {
	original := "Keep user instructions and report observed evidence."
	if got := managedGPTExecutionInstructions(original, nil); got != original {
		t.Fatal("tool-free request changed")
	}
	got := managedGPTExecutionInstructions(original, []codextools.Definition{{Name: "Agent"}})
	for _, part := range []string{original, "run asynchronously", "Do not use native sleep", "completion notifications", "user explicitly requests waiting", "yielded execution", "Never report an agent task as completed"} {
		if !strings.Contains(got, part) {
			t.Errorf("missing instruction %q", part)
		}
	}
	if !strings.HasPrefix(got, original+"\n\n") {
		t.Fatal("caller instructions were replaced")
	}
}
