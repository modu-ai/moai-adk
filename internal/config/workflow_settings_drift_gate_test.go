package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWorkflowSettingsDriftGateDefaultFalse pins the shipped default: the
// REFUSAL layer of the pre-merge settings.json drift assertion is OFF.
// Detection, preservation and the ledger row are not gated by this flag and
// run on every acquire; a maintainer opts into refusal via local config.
// Template neutrality: no `enabled: true` anywhere under
// internal/template/templates/.
func TestWorkflowSettingsDriftGateDefaultFalse(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultConfig()
	if cfg.Workflow.SettingsDriftGate.Enabled {
		t.Errorf("Workflow.SettingsDriftGate.Enabled: got true, want false (shipped default)")
	}
}

// TestWorkflowSettingsDriftGateLoaderRoundTrip pins the workflow.yaml surface:
// `settings_drift_gate.enabled: true` in a section file overrides the default.
// The refusal-path fixtures (AC-PSD-009 / AC-PSD-010) depend on this path
// working, so it is asserted here rather than assumed there.
func TestWorkflowSettingsDriftGateLoaderRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yaml := "workflow:\n  settings_drift_gate:\n    enabled: true\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	cfg, err := NewLoader().Load(filepath.Join(dir, ".moai"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Workflow.SettingsDriftGate.Enabled {
		t.Errorf("settings_drift_gate.enabled: got false, want true after section load")
	}
}

// TestWorkflowSettingsDriftGateIsNotIntegrationLock pins the separation D2
// argued for: the two flags are independent keys, so turning one off says
// nothing about the other. A sub-key implementation would fail this.
func TestWorkflowSettingsDriftGateIsNotIntegrationLock(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yaml := "workflow:\n  settings_drift_gate:\n    enabled: true\n  integration_lock:\n    enabled: false\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	cfg, err := NewLoader().Load(filepath.Join(dir, ".moai"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Workflow.SettingsDriftGate.Enabled {
		t.Errorf("settings_drift_gate.enabled: got false, want true")
	}
	if cfg.Workflow.IntegrationLock.Enabled {
		t.Errorf("integration_lock.enabled: got true, want false (the two keys are independent)")
	}
}
