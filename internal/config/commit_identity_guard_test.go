package config

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestCommitIdentityGuard_DefaultOff covers the config half of AC-CIG-008 /
// AC-CIG-009: the engine default is off (REQ-CIG-006), a tree with no
// configuration file loads it off, a workflow section that never names the key
// keeps the default, and the explicit `enabled: true` + `deny_emails` rows
// bind where the local config writes them (the positive control proving the
// false rows are not produced by a key that never binds).
func TestCommitIdentityGuard_DefaultOff(t *testing.T) {
	t.Run("defaults_struct_is_off", func(t *testing.T) {
		if NewDefaultWorkflowConfig().CommitIdentityGuard.Enabled {
			t.Fatal("Workflow.CommitIdentityGuard.Enabled = true in the engine default, want false")
		}
	})

	t.Run("tree_without_config_loads_off", func(t *testing.T) {
		cfg, err := NewConfigManager().Load(t.TempDir())
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Workflow.CommitIdentityGuard.Enabled {
			t.Fatal("a tree with no configuration file loaded commit_identity_guard.enabled = true")
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
		if wrapper.Workflow.CommitIdentityGuard.Enabled {
			t.Fatal("a workflow section that never names commit_identity_guard turned it on")
		}
	})

	t.Run("control_explicit_true_and_deny_emails_bind", func(t *testing.T) {
		var wrapper struct {
			Workflow WorkflowConfig `yaml:"workflow"`
		}
		text := "workflow:\n  commit_identity_guard:\n    enabled: true\n    deny_emails:\n      - ops-bot@corp.invalid\n"
		if err := yaml.Unmarshal([]byte(text), &wrapper); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !wrapper.Workflow.CommitIdentityGuard.Enabled {
			t.Fatal("workflow.commit_identity_guard.enabled: true did not bind")
		}
		if got, want := wrapper.Workflow.CommitIdentityGuard.DenyEmails, []string{"ops-bot@corp.invalid"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("deny_emails = %v, want %v", got, want)
		}
	})
}
