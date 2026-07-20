package agentlint

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ============================================================================
// SPEC-AGENT-TEAM-RETIRE-001 REQ-ATR-009: moai workflow lint repurposed to
// model_routing_profiles closed-set validation (AC-ATR-014). The former
// Agent Teams role-profiles isolation checks are retired with the static
// team layer.
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

// TestWorkflowLint_ModelRoutingValidConfig verifies a fully-valid
// model_routing_profiles block produces zero violations.
func TestWorkflowLint_ModelRoutingValidConfig(t *testing.T) {
	workflowFile := writeWorkflowYAML(t, `workflow:
  model_routing_profiles:
    max:
      S-run:
        model: sonnet
        effort: high
      L-plan:
        model: opus
        effort: xhigh
    low:
      M-sync:
        model: inherit
        effort: medium
`)

	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("loadWorkflowYAML: %v", err)
	}

	violations := validateModelRoutingProfiles(cfg)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for valid config, got %d: %v", len(violations), violations)
	}
}

// TestWorkflowLint_ModelRoutingAbsentBlockIsValid verifies an absent
// model_routing_profiles block is valid (every lookup falls back).
func TestWorkflowLint_ModelRoutingAbsentBlockIsValid(t *testing.T) {
	workflowFile := writeWorkflowYAML(t, `workflow:
  execution_mode: solo
`)

	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("loadWorkflowYAML: %v", err)
	}

	violations := validateModelRoutingProfiles(cfg)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for absent block, got %d: %v", len(violations), violations)
	}
}

// TestWorkflowLint_ModelRoutingInvalidModel verifies a model outside the
// closed set {inherit, sonnet, opus, glm} triggers MODEL_ROUTING_INVALID.
func TestWorkflowLint_ModelRoutingInvalidModel(t *testing.T) {
	workflowFile := writeWorkflowYAML(t, `workflow:
  model_routing_profiles:
    max:
      S-run:
        model: haiku
        effort: high
`)

	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("loadWorkflowYAML: %v", err)
	}

	violations := validateModelRoutingProfiles(cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	v := violations[0]
	if v.Rule != SentinelModelRoutingInvalid {
		t.Errorf("rule = %s, want %s", v.Rule, SentinelModelRoutingInvalid)
	}
	if v.Severity != string(SeverityError) {
		t.Errorf("severity = %s, want error", v.Severity)
	}
	if !strings.Contains(v.Path, "S-run") {
		t.Errorf("path should contain the offending key 'S-run', got: %s", v.Path)
	}
	if v.Actual != "haiku" {
		t.Errorf("actual = %q, want %q", v.Actual, "haiku")
	}
}

// TestWorkflowLint_ModelRoutingInvalidPerfTier verifies a perfTier outside
// {max, medium, low} triggers MODEL_ROUTING_INVALID.
func TestWorkflowLint_ModelRoutingInvalidPerfTier(t *testing.T) {
	workflowFile := writeWorkflowYAML(t, `workflow:
  model_routing_profiles:
    turbo:
      S-run:
        model: sonnet
        effort: high
`)

	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("loadWorkflowYAML: %v", err)
	}

	violations := validateModelRoutingProfiles(cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if !strings.Contains(violations[0].Path, "turbo") {
		t.Errorf("path should contain the offending perfTier 'turbo', got: %s", violations[0].Path)
	}
}

// TestWorkflowLint_ModelRoutingInvalidKeyShape verifies a key that is not
// <TIER>-<phase> triggers MODEL_ROUTING_INVALID.
func TestWorkflowLint_ModelRoutingInvalidKeyShape(t *testing.T) {
	workflowFile := writeWorkflowYAML(t, `workflow:
  model_routing_profiles:
    max:
      XL-deploy:
        model: sonnet
        effort: high
`)

	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("loadWorkflowYAML: %v", err)
	}

	violations := validateModelRoutingProfiles(cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if !strings.Contains(violations[0].Path, "XL-deploy") {
		t.Errorf("path should contain the offending key 'XL-deploy', got: %s", violations[0].Path)
	}
}

// TestWorkflowLint_RunEViolationExitPath verifies the cobra RunE contract:
// a violating config returns errLintViolations (mapped to a non-zero exit by
// cmd/moai/main.go — AC-ATR-014 violation-path evidence: the subcommand did
// NOT become a no-op after the role-profiles check removal).
func TestWorkflowLint_RunEViolationExitPath(t *testing.T) {
	workflowFile := writeWorkflowYAML(t, `workflow:
  model_routing_profiles:
    max:
      S-run:
        model: haiku
        effort: high
`)

	cmd := &cobra.Command{Use: "lint", RunE: runWorkflowLint}
	cmd.Flags().String("path", workflowFile, "")
	cmd.Flags().String("format", "text", "")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := runWorkflowLint(cmd, nil)
	if err == nil {
		t.Fatal("expected errLintViolations for violating config, got nil")
	}
	if !errors.Is(err, errLintViolations) {
		t.Errorf("error = %v, want errLintViolations", err)
	}
	if !strings.Contains(out.String(), SentinelModelRoutingInvalid) {
		t.Errorf("output should name %s, got: %s", SentinelModelRoutingInvalid, out.String())
	}
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
