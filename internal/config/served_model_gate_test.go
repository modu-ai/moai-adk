package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestServedModelGate_DefaultOff is AC-SMA-010's compiled half: a tree with no
// configuration file loads workflow.served_model_gate.enabled as false, the key
// binds where the template and the local config write it, and a workflow
// section that never names it keeps the default. The positive control (the
// explicit `enabled: true` row) shows the binding is live, so the false rows
// are not produced by a key that never binds.
func TestServedModelGate_DefaultOff(t *testing.T) {
	t.Run("defaults_struct_is_off", func(t *testing.T) {
		if NewDefaultWorkflowConfig().ServedModelGate.Enabled {
			t.Fatal("Workflow.ServedModelGate.Enabled = true in the engine default, want false")
		}
	})

	t.Run("tree_without_config_loads_off", func(t *testing.T) {
		cfg, err := NewConfigManager().Load(t.TempDir())
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Workflow.ServedModelGate.Enabled {
			t.Fatal("a tree with no configuration file loaded served_model_gate.enabled = true")
		}
	})

	t.Run("absent_key_keeps_default", func(t *testing.T) {
		wc := NewDefaultWorkflowConfig()
		var wrapper struct {
			Workflow *WorkflowConfig `yaml:"workflow"`
		}
		wrapper.Workflow = &wc
		if err := yaml.Unmarshal([]byte("workflow:\n  default_mode: personal\n"), &wrapper); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if wrapper.Workflow.ServedModelGate.Enabled {
			t.Fatal("a workflow section that never names served_model_gate turned it on")
		}
	})

	t.Run("control_explicit_true_binds", func(t *testing.T) {
		var wrapper struct {
			Workflow WorkflowConfig `yaml:"workflow"`
		}
		if err := yaml.Unmarshal([]byte("workflow:\n  served_model_gate:\n    enabled: true\n"), &wrapper); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !wrapper.Workflow.ServedModelGate.Enabled {
			t.Fatal("workflow.served_model_gate.enabled: true did not bind to Workflow.ServedModelGate.Enabled")
		}
	})
}
