package project

// initializer_audit_test.go — SPEC-MOAI-MCP-SERVER-001 M4 (REQ-MCP-015 /
// AC-MCP-020). Verifies writeWorkflowAuditYAML persists the audit + codex
// review-gate selection to workflow.yaml, reusing the M3 typed-config yaml keys
// (workflow.audit.model / .gates.* + workflow.codex.review_gate.enabled).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// TestWriteWorkflowAuditYAML_InsertsBlockPreservingNeighbors verifies the
// audit block is inserted under workflow: when absent, preserving every
// unrelated line (token_budget, worktree, comments). Mirrors the deployer path.
func TestWriteWorkflowAuditYAML_InsertsBlockPreservingNeighbors(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	before := `workflow:
  execution_mode: auto
  token_budget:
    plan: 30000
    run: 180000
  worktree:
    auto_create: false
`
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.WorkflowYAML), []byte(before), 0644); err != nil {
		t.Fatal(err)
	}

	opts := InitOptions{
		ProjectRoot:     root,
		AuditConfigSet:  true,
		AuditModel:      config.AuditModelClaude,
		AuditGateClaude: config.AuditGateRequired,
		AuditGateCodex:  config.AuditGateRequired,
		AuditGateGLM:    config.AuditGateAdvisory,
	}
	if err := writeWorkflowAuditYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("writeWorkflowAuditYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.WorkflowYAML))
	for _, want := range []string{
		"audit:",
		"model: claude",
		"gates:",
		"claude: required",
		"codex: required",
		"glm: advisory",
	} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("workflow.yaml missing %q; got: %s", want, got)
		}
	}
	// Unrelated keys preserved.
	for _, want := range []string{"plan: 30000", "run: 180000", "auto_create: false"} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("workflow.yaml lost unrelated line %q; got: %s", want, got)
		}
	}
	// codex.review_gate is NOT written when CodexAuditEnabled is false (opt-in).
	if bytes.Contains(got, []byte("review_gate:")) {
		t.Errorf("workflow.yaml must NOT carry review_gate when CodexAuditEnabled=false; got: %s", got)
	}
}

// TestWriteWorkflowAuditYAML_CodexReviewGateOptIn verifies that
// CodexAuditEnabled=true writes workflow.codex.review_gate.enabled=true (the M2
// review-gate hook's opt-in config key). Default-off is the C6 invariant.
func TestWriteWorkflowAuditYAML_CodexReviewGateOptIn(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	before := "workflow:\n  execution_mode: auto\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.WorkflowYAML), []byte(before), 0644); err != nil {
		t.Fatal(err)
	}

	opts := InitOptions{
		ProjectRoot:       root,
		AuditConfigSet:    true,
		AuditModel:        config.AuditModelCodex,
		CodexAuditEnabled: true,
	}
	if err := writeWorkflowAuditYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("writeWorkflowAuditYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.WorkflowYAML))
	if !bytes.Contains(got, []byte("codex:")) {
		t.Errorf("workflow.yaml missing codex block; got: %s", got)
	}
	if !bytes.Contains(got, []byte("review_gate:")) {
		t.Errorf("workflow.yaml missing review_gate block; got: %s", got)
	}
	if !bytes.Contains(got, []byte("enabled: true")) {
		t.Errorf("workflow.yaml missing enabled: true; got: %s", got)
	}
}

// TestWriteWorkflowAuditYAML_FreshFile verifies the no-deployer fallback: when
// no workflow.yaml exists, a minimal block carrying the audit selection is
// created.
func TestWriteWorkflowAuditYAML_FreshFile(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	opts := InitOptions{
		ProjectRoot:     root,
		AuditConfigSet:  true,
		AuditModel:      config.AuditModelGLM,
		AuditGateClaude: config.AuditGateRequired,
	}
	result := &InitResult{}
	if err := writeWorkflowAuditYAML(sectionsDir, opts, result); err != nil {
		t.Fatalf("writeWorkflowAuditYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.WorkflowYAML))
	if !bytes.Contains(got, []byte("model: glm")) {
		t.Errorf("fresh workflow.yaml missing model: glm; got: %s", got)
	}
	if !bytes.Contains(got, []byte("claude: required")) {
		t.Errorf("fresh workflow.yaml missing claude: required; got: %s", got)
	}
}

// TestWriteWorkflowAuditYAML_PatchesExistingBlock verifies the in-place patch
// path: when the audit block ALREADY exists in workflow.yaml, the writer
// patches each leaf in place (it does NOT insert a second block), and the
// codex.review_gate.enabled leaf is patched when its sub-block already exists.
// This covers the writeWorkflowAuditYAML patch branch + auditAnyPatched=true.
func TestWriteWorkflowAuditYAML_PatchesExistingBlock(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	before := `workflow:
  audit:
    model: codex
    gates:
      claude: advisory
      codex: advisory
      glm: off
  codex:
    review_gate:
      enabled: false
  token_budget:
    plan: 30000
`
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.WorkflowYAML), []byte(before), 0644); err != nil {
		t.Fatal(err)
	}

	opts := InitOptions{
		ProjectRoot:       root,
		AuditConfigSet:    true,
		AuditModel:        config.AuditModelClaude,
		AuditGateClaude:   config.AuditGateRequired,
		AuditGateCodex:    config.AuditGateRequired,
		AuditGateGLM:      config.AuditGateAdvisory,
		CodexAuditEnabled: true,
	}
	if err := writeWorkflowAuditYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("writeWorkflowAuditYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.WorkflowYAML))
	// Each leaf patched in place.
	for _, want := range []string{"model: claude", "claude: required", "codex: required", "glm: advisory", "enabled: true"} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("workflow.yaml missing patched leaf %q; got: %s", want, got)
		}
	}
	// Exactly ONE audit block (no duplicate insertion).
	if c := bytes.Count(got, []byte("audit:")); c != 1 {
		t.Errorf("workflow.yaml has %d `audit:` blocks, want 1 (patch must not duplicate); got: %s", c, got)
	}
	// Unrelated key preserved.
	if !bytes.Contains(got, []byte("plan: 30000")) {
		t.Errorf("workflow.yaml lost unrelated token_budget.plan; got: %s", got)
	}
}

// TestWriteWorkflowAuditYAML_SetTrackerSkipsUnset verifies the Set-tracker
// contract: AuditConfigSet=false (--non-interactive, no wizard) MUST NOT touch
// the deployed workflow.yaml — the audit selection is opt-in via the wizard.
func TestWriteWorkflowAuditYAML_SetTrackerSkipsUnset(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	before := "workflow:\n  execution_mode: auto\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.WorkflowYAML), []byte(before), 0644); err != nil {
		t.Fatal(err)
	}

	// AuditConfigSet=false (the zero value). Audit fields populated but ignored.
	opts := InitOptions{
		ProjectRoot:    root,
		AuditConfigSet: false,
		AuditModel:     config.AuditModelMulti,
		MCPProvision:   true,
	}
	if err := writeWorkflowAuditYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("writeWorkflowAuditYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.WorkflowYAML))
	if bytes.Contains(got, []byte("audit:")) {
		t.Errorf("AuditConfigSet=false must NOT write an audit block; got: %s", got)
	}
	if !bytes.Equal(got, []byte(before)) {
		t.Errorf("AuditConfigSet=false must leave workflow.yaml byte-identical; got: %s", got)
	}
}

// TestWriteWorkflowAuditYAML_InsertsMissingLeavesIntoPinOnlyAuditBlock is the
// SPEC-V3R6-AUDIT-MODEL-PIN-001 F2 regression: the distributed template's
// audit block now ships with ONLY the codex/glm pin sub-keys (no model, no
// gates), so workflowHasAuditBlock routes writeWorkflowAuditYAML to the patch
// path — and the wizard's audit selection MUST still land. Without an
// insertion fallback in patchAuditLeaves the leaves are silently dropped
// (SPEC-INIT-WIZARD-REPAIR-001 AC-009 broken).
func TestWriteWorkflowAuditYAML_InsertsMissingLeavesIntoPinOnlyAuditBlock(t *testing.T) {
	t.Parallel()
	root, sectionsDir := setupSectionsDir(t)

	// The NEW template shape: audit carries only the pin sub-keys.
	before := "workflow:\n" +
		"    default_mode: \"\"\n" +
		"    audit:\n" +
		"        codex:\n" +
		"            model: \"\"\n" +
		"            effort: \"\"\n" +
		"        glm:\n" +
		"            model: \"\"\n" +
		"            effort: \"\"\n" +
		"    token_budget:\n" +
		"        plan: 30000\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.WorkflowYAML), []byte(before), 0644); err != nil {
		t.Fatal(err)
	}

	opts := InitOptions{
		ProjectRoot:       root,
		AuditConfigSet:    true,
		AuditModel:        config.AuditModelClaude,
		AuditGateClaude:   config.AuditGateRequired,
		AuditGateCodex:    config.AuditGateAdvisory,
		AuditGateGLM:      config.AuditGateOff,
		CodexAuditEnabled: true,
	}
	if err := writeWorkflowAuditYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("writeWorkflowAuditYAML: %v", err)
	}

	got, _ := os.ReadFile(filepath.Join(sectionsDir, defs.WorkflowYAML))
	s := string(got)
	for _, want := range []string{"model: claude", "claude: required", "codex: advisory", "glm: off"} {
		if !strings.Contains(s, want) {
			t.Errorf("workflow.yaml missing wizard selection %q (insertion fallback did not fire); got: %s", want, s)
		}
	}
	// Exactly ONE audit block — insertion must nest, not duplicate.
	if c := strings.Count(s, "audit:"); c != 1 {
		t.Errorf("workflow.yaml has %d `audit:` blocks, want 1; got: %s", c, s)
	}
	// The pre-existing pin sub-keys survive untouched.
	if !strings.Contains(s, `effort: ""`) {
		t.Errorf("workflow.yaml lost the pin sub-keys; got: %s", s)
	}

	// Structural guard: the inserted leaves land at the right NESTING (a
	// mis-indented insert would re-parent siblings). Parse and assert paths.
	wf := parseWorkflowYAMLForTest(t, s)
	audit, _ := wf["audit"].(map[string]any)
	if audit == nil || audit["model"] != "claude" {
		t.Errorf("workflow.audit.model must parse as claude; got: %s", s)
	}
	gates, _ := audit["gates"].(map[string]any)
	if gates == nil || gates["claude"] != "required" || gates["codex"] != "advisory" || gates["glm"] != "off" {
		t.Errorf("workflow.audit.gates must parse with all three wizard gates; got: %s", s)
	}
	codexPin, _ := audit["codex"].(map[string]any)
	if codexPin == nil || codexPin["model"] != "" {
		t.Errorf("workflow.audit.codex pin block must survive as a map; got: %s", s)
	}
	if tb, _ := wf["token_budget"].(map[string]any); tb == nil || tb["plan"] != 30000 {
		t.Errorf("workflow.token_budget must stay a direct workflow child; got: %s", s)
	}
}

// parseWorkflowYAMLForTest parses a workflow.yaml body into its top-level
// mapping (test-only; fails the test on a parse error).
func parseWorkflowYAMLForTest(t *testing.T, s string) map[string]any {
	t.Helper()
	var wf struct {
		Workflow map[string]any `yaml:"workflow"`
	}
	if err := yaml.Unmarshal([]byte(s), &wf); err != nil {
		t.Fatalf("parse workflow.yaml: %v\n%s", err, s)
	}
	return wf.Workflow
}
