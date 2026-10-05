package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestDefaults_JevDisabled is the AC-JEVC-005 compiled half: the shipped
// default for workflow.jev.enabled is false, and internal/config/defaults.go is
// the source of truth for it (REQ-JEVC-016).
func TestDefaults_JevDisabled(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultWorkflowConfig()
	if cfg.Jev.Enabled {
		t.Error("Workflow.Jev.Enabled = true, want false (the capability ships OFF)")
	}
}

// TestJevConfig_KeyShape asserts the key binds at workflow.jev.enabled inside
// the existing workflow section — no new config section file (REQ-JEVC-015).
func TestJevConfig_KeyShape(t *testing.T) {
	t.Parallel()
	var wrapper struct {
		Workflow WorkflowConfig `yaml:"workflow"`
	}
	src := "workflow:\n" +
		"  jev:\n" +
		"    enabled: true\n"
	if err := yaml.Unmarshal([]byte(src), &wrapper); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !wrapper.Workflow.Jev.Enabled {
		t.Error("workflow.jev.enabled: true did not bind to Workflow.Jev.Enabled")
	}
}

// TestJevConfig_AbsentKeyKeepsDefault: a workflow section that never mentions
// jev leaves the capability off. The loader unmarshals onto the
// default-populated struct, so this is the branch that must not silently flip.
func TestJevConfig_AbsentKeyKeepsDefault(t *testing.T) {
	t.Parallel()
	cfg := NewDefaultWorkflowConfig()
	src := "workflow:\n  default_mode: personal\n"
	var wrapper struct {
		Workflow *WorkflowConfig `yaml:"workflow"`
	}
	wrapper.Workflow = &cfg
	if err := yaml.Unmarshal([]byte(src), &wrapper); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if wrapper.Workflow.Jev.Enabled {
		t.Error("Workflow.Jev.Enabled = true after a workflow section that never names jev")
	}
}

// TestTemplateWorkflowYAML_JevShipsOff reads the distributed template and
// asserts the shipped value is false — the template-side half of AC-JEVC-005,
// and the template-neutrality guard against an `enabled: true` shipping to
// every user.
func TestTemplateWorkflowYAML_JevShipsOff(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "template", "templates", ".moai", "config", "sections", "workflow.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read template workflow.yaml: %v", err)
	}
	if !strings.Contains(string(raw), "jev:") {
		t.Fatalf("template workflow.yaml carries no jev block — the shipped default is undocumented")
	}

	var wrapper struct {
		Workflow WorkflowConfig `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(raw, &wrapper); err != nil {
		t.Fatalf("unmarshal template workflow.yaml: %v", err)
	}
	if wrapper.Workflow.Jev.Enabled {
		t.Error("template workflow.yaml ships workflow.jev.enabled: true — the capability MUST ship OFF")
	}
	// Positive control: the same decode reads a sibling opt-in switch, so a
	// false here is attributable to the value rather than to a failed decode.
	if wrapper.Workflow.SlotLease.DefaultMaxDuration == "" {
		t.Fatal("positive control failed: the template decode produced no slot_lease.default_max_duration, so the jev assertion above is unattributable")
	}
}

// TestConfigCacheSchemaVersion_BumpedForJev: adding a Workflow field requires a
// cache schema bump. A cache written by an older binary omits the new key, and
// json.Unmarshal leaves it at its zero value — so a fingerprint-valid cache
// would silently serve enabled=false over an enabled:true workflow.yaml. This
// is the observed AgentStopGuard / SettingsDriftGate / SlotLease defect.
func TestConfigCacheSchemaVersion_BumpedForJev(t *testing.T) {
	t.Parallel()
	if configCacheSchemaVersion < 5 {
		t.Errorf("configCacheSchemaVersion = %d, want >= 5 — Workflow gained Jev without a cache schema bump", configCacheSchemaVersion)
	}
}
