package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowAgentStopGuardDefaultFalse pins the shipped default: the
// enforcement gate (deny layer) is OFF — observation and advisory always run,
// a maintainer opts into denial via local config. Template neutrality: no
// `enabled: true` anywhere under internal/template/templates/.
func TestWorkflowAgentStopGuardDefaultFalse(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultConfig()
	if cfg.Workflow.AgentStopGuard.Enabled {
		t.Errorf("Workflow.AgentStopGuard.Enabled: got true, want false (shipped default)")
	}
}

// TestWorkflowAgentStopGuardLoaderRoundTrip pins the workflow.yaml surface:
// `agent_stop_guard.enabled: true` in a section file overrides the default.
func TestWorkflowAgentStopGuardLoaderRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yaml := "workflow:\n  agent_stop_guard:\n    enabled: true\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	cfg, err := NewLoader().Load(filepath.Join(dir, ".moai"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Workflow.AgentStopGuard.Enabled {
		t.Errorf("agent_stop_guard.enabled: got false, want true after section load")
	}
}
