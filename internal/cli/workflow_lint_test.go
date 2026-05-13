package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper to create a temporary workflow.yaml file
func createTempWorkflowFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	workflowPath := filepath.Join(tmpDir, "workflow.yaml")

	if err := os.WriteFile(workflowPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp workflow file: %v", err)
	}

	return workflowPath
}

// TestWorkflowLint_OrcWorktreeRequiredOnImplementer verifies implementer.isolation=none triggers ORC_WORKTREE_REQUIRED (AC-09)
func TestWorkflowLint_OrcWorktreeRequiredOnImplementer(t *testing.T) {
	content := `workflow:
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
`

	workflowFile := createTempWorkflowFile(t, content)
	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("failed to load workflow YAML: %v", err)
	}

	violations := validateRoleProfiles(cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if violations[0].Rule != SentinelWorktreeRequired {
		t.Errorf("expected rule %s, got %s", SentinelWorktreeRequired, violations[0].Rule)
	}

	if !strings.Contains(violations[0].Path, "implementer") {
		t.Errorf("expected path to contain 'implementer', got %s", violations[0].Path)
	}
}

// TestWorkflowLint_OrcWorktreeRequiredOnTester verifies tester.isolation=none triggers ORC_WORKTREE_REQUIRED
func TestWorkflowLint_OrcWorktreeRequiredOnTester(t *testing.T) {
	content := `workflow:
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
`

	workflowFile := createTempWorkflowFile(t, content)
	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("failed to load workflow YAML: %v", err)
	}

	violations := validateRoleProfiles(cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if violations[0].Rule != SentinelWorktreeRequired {
		t.Errorf("expected rule %s, got %s", SentinelWorktreeRequired, violations[0].Rule)
	}

	if !strings.Contains(violations[0].Path, "tester") {
		t.Errorf("expected path to contain 'tester', got %s", violations[0].Path)
	}
}

// TestWorkflowLint_OrcWorktreeRequiredOnDesigner verifies designer.isolation=none triggers ORC_WORKTREE_REQUIRED
func TestWorkflowLint_OrcWorktreeRequiredOnDesigner(t *testing.T) {
	content := `workflow:
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
`

	workflowFile := createTempWorkflowFile(t, content)
	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("failed to load workflow YAML: %v", err)
	}

	violations := validateRoleProfiles(cfg)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if violations[0].Rule != SentinelWorktreeRequired {
		t.Errorf("expected rule %s, got %s", SentinelWorktreeRequired, violations[0].Rule)
	}

	if !strings.Contains(violations[0].Path, "designer") {
		t.Errorf("expected path to contain 'designer', got %s", violations[0].Path)
	}
}

// TestWorkflowLint_NoViolationsOnCorrectConfig verifies correct config passes (baseline)
func TestWorkflowLint_NoViolationsOnCorrectConfig(t *testing.T) {
	content := `workflow:
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
`

	workflowFile := createTempWorkflowFile(t, content)
	cfg, err := loadWorkflowYAML(workflowFile)
	if err != nil {
		t.Fatalf("failed to load workflow YAML: %v", err)
	}

	violations := validateRoleProfiles(cfg)
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(violations))
	}
}
