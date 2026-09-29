package agentlint

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// ============================================================================
// moai workflow lint: the model_routing_profiles closed-set check (SPEC-AGENT-
// TEAM-RETIRE-001 AC-ATR-014) was retired by SPEC-AGENT-MODEL-INHERIT-001 D12; the
// remaining tests pin the parse and exit-code contract of the subcommand.
// ============================================================================

// writeWorkflowYAML writes a workflow.yaml fixture and returns its path.
func writeWorkflowYAML(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "workflow.yaml")
	if err := os.WriteFile(workflowFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}
	return workflowFile
}

// TestWorkflowLint_RunECleanExitPath verifies the clean-config path returns
// nil (exit 0 contract preserved).
func TestWorkflowLint_RunECleanExitPath(t *testing.T) {
	workflowFile := writeWorkflowYAML(t, `workflow:
  model_routing_profiles:
    medium:
      L-run:
        model: glm
        effort: max
`)

	cmd := &cobra.Command{Use: "lint", RunE: runWorkflowLint}
	cmd.Flags().String("path", workflowFile, "")
	cmd.Flags().String("format", "text", "")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := runWorkflowLint(cmd, nil); err != nil {
		t.Fatalf("expected nil for clean config, got: %v", err)
	}
}
