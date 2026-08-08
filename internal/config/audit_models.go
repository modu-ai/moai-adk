package config

// audit_models.go — multi-model audit enum tokens + the AuditConfig surface
// consumed by the SPEC-MOAI-MCP-SERVER-001 / SPEC-AUDIT-MULTI-MODEL-001
// backends (internal/cli/mcp_audit.go, mcp_audit_multi.go, mcp_convergence.go,
// codex_review_gate.go).
//
// The enum values are plain lowercase strings so that YAML round-trip
// (workflow.yaml → WorkflowConfig → yaml.Marshal) preserves the user-authored
// tokens verbatim. The locked distributed-default profile is claude+codex
// required, glm advisory (see NewDefaultWorkflowConfig).
//
// @MX:NOTE: [AUTO] audit_models.go — multi-model audit enum + AuditConfig surface for the MCP audit backends
// @MX:SPEC: SPEC-MOAI-MCP-SERVER-001

// AuditModel enum (audit_model in workflow.yaml). `multi` is a DECLARED token
// only — its convergence logic is owned by SPEC-AUDIT-MULTI-MODEL; the M3
// surface accepts the token but does NOT orchestrate the parallel fan-out.
const (
	// AuditModelClaude is the default single-backend audit model.
	AuditModelClaude = "claude"
	// AuditModelCodex selects the codex JSON-RPC reviewer as the active backend.
	AuditModelCodex = "codex"
	// AuditModelGLM selects the GLM reviewer as the active backend.
	AuditModelGLM = "glm"
	// AuditModelMulti declares the multi-auditor convergence model. Accepted as
	// a stored value; its convergence logic is deferred to
	// SPEC-AUDIT-MULTI-MODEL (AP-8).
	AuditModelMulti = "multi"
)

// AuditGate enum (per-auditor audit_gate in workflow.yaml).
const (
	// AuditGateOff disables the auditor entirely.
	AuditGateOff = "off"
	// AuditGateAdvisory runs the auditor but never blocks convergence on its
	// verdict (the default for GLM, which is user-enabled).
	AuditGateAdvisory = "advisory"
	// AuditGateRequired makes the auditor's fail verdict block overall
	// convergence (the default for claude and codex).
	AuditGateRequired = "required"
)

// AuditGates is the per-auditor gate configuration. Each field is an AuditGate
// enum token (off|advisory|required). The zero value (empty string) is treated
// by the convergence engine as "unset — apply the distributed default".
type AuditGates struct {
	// Claude gates the always-available anchor verdict.
	Claude string `yaml:"claude,omitempty" json:"claude,omitempty"`
	// Codex gates the codex JSON-RPC reviewer.
	Codex string `yaml:"codex,omitempty" json:"codex,omitempty"`
	// GLM gates the GLM reviewer (advisory by default — user-enabled).
	GLM string `yaml:"glm,omitempty" json:"glm,omitempty"`
}

// AuditConfig is the workflow.audit block: the active audit_model plus the
// per-auditor gates. Loaded from workflow.yaml; defaults live in
// NewDefaultWorkflowConfig.
type AuditConfig struct {
	// Model is the active audit_model token (claude|codex|glm|multi).
	Model string `yaml:"model,omitempty" json:"model,omitempty"`
	// Gates carries the per-auditor gate tokens.
	Gates AuditGates `yaml:"gates,omitempty" json:"gates,omitempty"`
}
