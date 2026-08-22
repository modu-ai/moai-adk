package cli

// SPEC-INIT-WIZARD-REPAIR-001 chain ③ runInit wiring test (M3, AC-009): the
// interactive wizard's audit + codex review-gate answers persist into the
// deployed workflow.yaml through writeWorkflowAuditYAML on the deployer path,
// without re-parenting the template's existing workflow children.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// yamlUnmarshalWorkflow is a thin alias so the parse helper reads uniformly.
func yamlUnmarshalWorkflow(data []byte, out *map[string]any) error {
	return yaml.Unmarshal(data, out)
}

// TestRunInit_WizardAuditSelectionPersists asserts AC-009 on the deployer
// path: a wizard result carrying an audit selection (AuditConfigSet=true via
// applyWizardPage3ToOpts) lands in the deployed workflow.yaml, and the insert
// does not re-parent the template's pre-existing workflow children (the
// 4-space-indent sibling hazard).
func TestRunInit_WizardAuditSelectionPersists(t *testing.T) {
	wiz := &wizard.WizardResult{
		AuditModel:        "claude",
		AuditGateClaude:   "required",
		AuditGateCodex:    "advisory",
		AuditGateGLM:      "off",
		CodexAuditEnabled: true,
	}
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	projectDir := runInitForAutonomyAtHome(t, homeDir, wiz, nil)

	workflowPath := filepath.Join(projectDir, ".moai", "config", "sections", defs.WorkflowYAML)
	got, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read deployed workflow.yaml: %v", err)
	}
	for _, want := range []string{
		"audit:",
		"model: claude",
		"claude: required",
		"codex: advisory",
		"glm: off",
		"review_gate:",
		"enabled: true",
	} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("deployed workflow.yaml missing %q; got:\n%s", want, got)
		}
	}

	// Nesting guard (boundary check): the shipped template indents workflow
	// children at 4 spaces; an insert at the wrong indentation silently
	// re-parents every later workflow child under the audit block. Parse and
	// assert the audit block AND the pre-existing siblings keep their paths.
	wf := parseWorkflowYAMLT(t, got)
	audit, _ := wf["audit"].(map[string]any)
	if audit == nil || audit["model"] != "claude" {
		t.Errorf("audit must parse as workflow.audit.model=claude; got:\n%s", got)
	}
	if gates, _ := audit["gates"].(map[string]any); gates == nil || gates["claude"] != "required" {
		t.Errorf("gates must parse as workflow.audit.gates.claude=required; got:\n%s", got)
	}
	if wf["default_mode"] != "" {
		t.Errorf("workflow.default_mode must stay a direct workflow child; got:\n%s", got)
	}
	if tb, _ := wf["token_budget"].(map[string]any); tb == nil || tb["plan"] != 30000 {
		t.Errorf("workflow.token_budget must stay a direct workflow child; got:\n%s", got)
	}
	if wt, _ := wf["worktree"].(map[string]any); wt == nil || wt["auto_create"] != false {
		t.Errorf("workflow.worktree must stay a direct workflow child; got:\n%s", got)
	}
	rg, _ := wf["codex"].(map[string]any)
	if rg == nil {
		t.Fatalf("codex must parse as workflow.codex; got:\n%s", got)
	}
	rgg, _ := rg["review_gate"].(map[string]any)
	if rgg == nil || rgg["enabled"] != true {
		t.Errorf("codex.review_gate.enabled must parse true; got:\n%s", got)
	}
}

// parseWorkflowYAMLT mirrors the project-package helper (small parse helper
// for the nesting guard).
func parseWorkflowYAMLT(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := yamlUnmarshalWorkflow(data, &doc); err != nil {
		t.Fatalf("workflow.yaml no longer parses: %v\ngot:\n%s", err, data)
	}
	wf, _ := doc["workflow"].(map[string]any)
	if wf == nil {
		t.Fatalf("workflow mapping lost; got:\n%s", data)
	}
	return wf
}
