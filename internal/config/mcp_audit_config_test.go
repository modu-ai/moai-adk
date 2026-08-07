package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// These tests cover the SPEC-MOAI-MCP-SERVER-001 M3 audit config surface
// (REQ-MCP-010 / AC-MCP-012): the 3-way audit_model enum, the per-auditor
// audit_gate enum, and the locked default profile (claude + codex required,
// glm advisory). They are RED until types.go + defaults.go carry the new
// AuditConfig.

func TestAuditModel_EnumValues(t *testing.T) {
	cases := map[string]string{
		"claude": AuditModelClaude,
		"codex":  AuditModelCodex,
		"glm":    AuditModelGLM,
		"multi":  AuditModelMulti,
	}
	for want, got := range cases {
		if got != want {
			t.Errorf("AuditModel %q = %q, want %q", want, got, want)
		}
	}
}

func TestAuditGate_EnumValues(t *testing.T) {
	cases := map[string]string{
		"off":      AuditGateOff,
		"advisory": AuditGateAdvisory,
		"required": AuditGateRequired,
	}
	for want, got := range cases {
		if got != want {
			t.Errorf("AuditGate %q = %q, want %q", want, got, want)
		}
	}
}

func TestAuditConfig_DefaultProfile(t *testing.T) {
	cfg := NewDefaultConfig()
	a := cfg.Workflow.Audit
	if a.Model != AuditModelClaude {
		t.Errorf("default Audit.Model = %q, want %q (claude)", a.Model, AuditModelClaude)
	}
	if a.Gates.Claude != AuditGateRequired {
		t.Errorf("default Gates.Claude = %q, want %q (required)", a.Gates.Claude, AuditGateRequired)
	}
	if a.Gates.Codex != AuditGateRequired {
		t.Errorf("default Gates.Codex = %q, want %q (required)", a.Gates.Codex, AuditGateRequired)
	}
	if a.Gates.GLM != AuditGateAdvisory {
		t.Errorf("default Gates.GLM = %q, want %q (advisory — user-enabled)", a.Gates.GLM, AuditGateAdvisory)
	}
}

func TestAuditConfig_YAMLRoundTrip(t *testing.T) {
	// A workflow.yaml carrying the audit block mirrors the locked schema
	// (progress.md §G.3). Round-tripping through yaml.v3 must preserve every
	// field, including the `multi` token (accepted-but-not-orchestrated).
	src := []byte(`
workflow:
  audit:
    model: glm
    gates:
      claude: off
      codex: advisory
      glm: required
`)
	var wrap struct {
		Workflow WorkflowConfig `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(src, &wrap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	a := wrap.Workflow.Audit
	if a.Model != AuditModelGLM {
		t.Errorf("Model = %q, want glm", a.Model)
	}
	if a.Gates.Claude != AuditGateOff {
		t.Errorf("Gates.Claude = %q, want off", a.Gates.Claude)
	}
	if a.Gates.Codex != AuditGateAdvisory {
		t.Errorf("Gates.Codex = %q, want advisory", a.Gates.Codex)
	}
	if a.Gates.GLM != AuditGateRequired {
		t.Errorf("Gates.GLM = %q, want required", a.Gates.GLM)
	}

	// `multi` token must parse without error even though its convergence logic
	// is owned by a future SPEC (AP-8).
	multi := []byte(`
workflow:
  audit:
    model: multi
`)
	var w2 struct {
		Workflow WorkflowConfig `yaml:"workflow"`
	}
	if err := yaml.Unmarshal(multi, &w2); err != nil {
		t.Fatalf("unmarshal multi token: %v", err)
	}
	if w2.Workflow.Audit.Model != AuditModelMulti {
		t.Errorf("multi token Model = %q, want multi", w2.Workflow.Audit.Model)
	}
}
