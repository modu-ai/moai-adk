package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// SPEC-V3R2-ORC-004: moai workflow lint tests (AC-09)
// RED tests — fail until workflow_lint.go is implemented
// ============================================================================

// TestWorkflowLint_OrcWorktreeRequiredOnImplementer tests that workflow.yaml
// with implementer.isolation=none triggers ORC_WORKTREE_REQUIRED (AC-09).
func TestWorkflowLint_OrcWorktreeRequiredOnImplementer(t *testing.T) {
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "workflow.yaml")
	err := os.WriteFile(workflowFile, []byte(`workflow:
  team:
    role_profiles:
      implementer:
        mode: acceptEdits
        isolation: none
      tester:
        mode: acceptEdits
        isolation: worktree
      designer:
        mode: acceptEdits
        isolation: worktree
`), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("loadWorkflowYAML: %v", err)
	}

	violations := validateRoleProfiles(cfg)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if violations[0].Rule != SentinelWorktreeRequired {
		t.Errorf("rule = %s, want %s", violations[0].Rule, SentinelWorktreeRequired)
	}
	if !strings.Contains(violations[0].Path, "implementer") {
		t.Errorf("path should contain 'implementer', got: %s", violations[0].Path)
	}
	if violations[0].Severity != string(SeverityError) {
		t.Errorf("severity = %s, want error", violations[0].Severity)
	}
}

// TestWorkflowLint_OrcWorktreeRequiredOnTester tests that tester.isolation=none
// triggers ORC_WORKTREE_REQUIRED.
func TestWorkflowLint_OrcWorktreeRequiredOnTester(t *testing.T) {
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "workflow.yaml")
	err := os.WriteFile(workflowFile, []byte(`workflow:
  team:
    role_profiles:
      implementer:
        mode: acceptEdits
        isolation: worktree
      tester:
        mode: acceptEdits
        isolation: none
      designer:
        mode: acceptEdits
        isolation: worktree
`), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("loadWorkflowYAML: %v", err)
	}

	violations := validateRoleProfiles(cfg)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if violations[0].Rule != SentinelWorktreeRequired {
		t.Errorf("rule = %s, want %s", violations[0].Rule, SentinelWorktreeRequired)
	}
	if !strings.Contains(violations[0].Path, "tester") {
		t.Errorf("path should contain 'tester', got: %s", violations[0].Path)
	}
}

// TestWorkflowLint_OrcWorktreeRequiredOnDesigner tests that designer.isolation=none
// triggers ORC_WORKTREE_REQUIRED.
func TestWorkflowLint_OrcWorktreeRequiredOnDesigner(t *testing.T) {
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "workflow.yaml")
	err := os.WriteFile(workflowFile, []byte(`workflow:
  team:
    role_profiles:
      implementer:
        mode: acceptEdits
        isolation: worktree
      tester:
        mode: acceptEdits
        isolation: worktree
      designer:
        mode: acceptEdits
        isolation: none
`), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("loadWorkflowYAML: %v", err)
	}

	violations := validateRoleProfiles(cfg)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if violations[0].Rule != SentinelWorktreeRequired {
		t.Errorf("rule = %s, want %s", violations[0].Rule, SentinelWorktreeRequired)
	}
	if !strings.Contains(violations[0].Path, "designer") {
		t.Errorf("path should contain 'designer', got: %s", violations[0].Path)
	}
}

// TestWorkflowLint_NoViolationsOnCorrectConfig tests that correctly configured
// workflow.yaml (implementer/tester/designer all with isolation:worktree)
// produces no violations (AC-09 clean-config scenario).
func TestWorkflowLint_NoViolationsOnCorrectConfig(t *testing.T) {
	tmpDir := t.TempDir()
	workflowFile := filepath.Join(tmpDir, "workflow.yaml")
	err := os.WriteFile(workflowFile, []byte(`workflow:
  team:
    role_profiles:
      implementer:
        mode: acceptEdits
        isolation: worktree
      tester:
        mode: acceptEdits
        isolation: worktree
      designer:
        mode: acceptEdits
        isolation: worktree
      researcher:
        mode: plan
        isolation: none
      analyst:
        mode: plan
        isolation: none
`), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("loadWorkflowYAML: %v", err)
	}

	violations := validateRoleProfiles(cfg)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for correct config, got %d: %v", len(violations), violations)
	}
}
