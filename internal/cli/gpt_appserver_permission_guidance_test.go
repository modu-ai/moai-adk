package cli

import (
	"github.com/modu-ai/moai-adk/internal/codextools"
	"strings"
	"testing"
)

func TestManagedGPTPermissionGuidanceDistinguishesNativeAndClaudeTools(t *testing.T) {
	for _, name := range []string{"Write", "Edit", "Bash", "EnterWorktree", "Read", "Agent"} {
		got := managedGPTExecutionInstructions("Preserve caller rules.", []codextools.Definition{{Name: name}})
		for _, part := range []string{"native filesystem sandbox", "Claude Code permissions", "Do not infer", "only tools present", "actual tool result", "Never bypass"} {
			if !strings.Contains(got, part) {
				t.Errorf("tool=%s missing %q", name, part)
			}
		}
		if !strings.HasPrefix(got, "Preserve caller rules.") {
			t.Fatal("caller instructions replaced")
		}
		for _, forbidden := range []string{"read-only", "danger-full-access", "workspace-write", "externalSandbox"} {
			if strings.Contains(got, forbidden) {
				t.Errorf("guidance asserted a sandbox value or requested privilege escalation: %s", forbidden)
			}
		}
		if name == "Agent" && !strings.Contains(got, "run asynchronously") {
			t.Fatal("Agent guidance lost during permission guidance composition")
		}
	}
	if got := managedGPTExecutionInstructions("No tools allowed.", nil); got != "No tools allowed." {
		t.Fatal("tool-free authority changed")
	}
}
