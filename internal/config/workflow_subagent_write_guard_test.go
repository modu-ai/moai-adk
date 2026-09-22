package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowSubagentWriteGuardDefaultFalse pins the shipped default: the
// deny layer is OFF — detection and the audit append always run, a maintainer
// opts into denial via local config (SPEC-SUBAGENT-WRITE-SHRINK-GUARD-001
// REQ-SWG-006, family contract). Template neutrality: no `enabled: true`
// anywhere under internal/template/templates/.
func TestWorkflowSubagentWriteGuardDefaultFalse(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultConfig()
	if cfg.Workflow.SubagentWriteGuard.Enabled {
		t.Errorf("Workflow.SubagentWriteGuard.Enabled: got true, want false (shipped default)")
	}
}

// TestWorkflowSubagentWriteGuardLoaderRoundTrip pins the workflow.yaml
// surface: `subagent_write_guard.enabled: true` in a section file overrides
// the default.
func TestWorkflowSubagentWriteGuardLoaderRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yaml := "workflow:\n  subagent_write_guard:\n    enabled: true\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	cfg, err := NewLoader().Load(filepath.Join(dir, ".moai"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Workflow.SubagentWriteGuard.Enabled {
		t.Errorf("subagent_write_guard.enabled: got false, want true after section load")
	}
}
