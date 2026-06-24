// Package toolpolicy implements the declarative tool/permission policy SSOT
// (SPEC-V3R6-TOOL-POLICY-SSOT-001). The YAML at
// .moai/config/sections/tool-policy.yaml is the single source from which the
// settings.json permissions block (enforcement surface) and the YAML header
// comment (audit surface) are generated.
//
// Schema origin: book1 ch08 6-field approval template (action/risk/requestor/
// decision/rationale/impact) made machine-readable. Drift-class analogy:
// harness-namespace-doctrine.md §24.5 (two independently-maintained surfaces
// diverging). Pattern origin: book2 ch4.3-4.4 (execpolicy — contractor with a
// compliance office), appendix B.3 (independently evaluable invariant).
//
// The tool-policy entry schema ({tool, args_pattern, risk_tier, decision,
// owner_agent, audit}) is DISJOINT from the constitution Rule schema
// ({id, zone, zone_class, file, anchor, clause, canary_gate}). The two cannot
// share a struct — see REQ-TPS-006 / design.md §D (D9 decision).
package toolpolicy

// RiskTier classifies the blast radius of a tool invocation.
//
// Enum values: read | write | irreversible. Ordering is risk-ascending so the
// codegen can seed highest-risk rules first per book2 ch8.4 ("make explicit
// which rules must be written down first").
type RiskTier string

const (
	// RiskTierRead is the lowest-risk tier (read-only operations: Read, Glob,
	// Grep, git status/log). Bulk of the current settings.json allow list.
	RiskTierRead RiskTier = "read"
	// RiskTierWrite covers filesystem-mutating operations (Write, Edit,
	// MultiEdit, git commit/push to origin).
	RiskTierWrite RiskTier = "write"
	// RiskTierIrreversible is the highest-risk tier (force-push, rm -rf on
	// project paths). These MUST be written down first per book2 ch8.4.
	RiskTierIrreversible RiskTier = "irreversible"
)

// Decision is the allow/deny/ask verdict for a policy entry.
//
// All three values map directly to the Claude Code settings.json permissions
// arrays: allow → permissions.allow, deny → permissions.deny, ask →
// permissions.ask (verified empirically — the local settings.json already
// carries a non-empty ask array; see design.md §H OQ-2 / D6).
type Decision string

const (
	// DecisionAllow permits the tool invocation without prompting.
	DecisionAllow Decision = "allow"
	// DecisionDeny blocks the tool invocation unconditionally.
	DecisionDeny Decision = "deny"
	// DecisionAsk prompts the user before proceeding.
	DecisionAsk Decision = "ask"
)

// EnvGate expresses an optional environment condition under which a rule
// applies (e.g., the GLM-backend web-tool prohibition). When present, the
// codegen emits the entry as a documentation/audit record rather than a static
// permissions entry, because a static deny would over-block the cg-leader
// exception (design.md §C.4).
type EnvGate struct {
	// BaseURLContains is a substring matched against ANTHROPIC_BASE_URL.
	// Example: "api.z.ai" for the GLM-backend prohibition.
	BaseURLContains string `yaml:"base_url_contains,omitempty" json:"base_url_contains,omitempty"`
	// ExceptionWhen names the exception condition (e.g., "cg_leader" for the
	// moai cg leader pane which runs the Claude backend). Documentation-only;
	// the codegen does not emit a branch for it.
	ExceptionWhen string `yaml:"exception_when,omitempty" json:"exception_when,omitempty"`
}

// PolicyEntry is the 6-field entry schema (plus optional source/env_gate
// metadata) for one tool/permission policy rule.
//
// Required fields (book1 ch08 6-field approval template):
//   - Tool + ArgsPattern  → action
//   - RiskTier            → risk
//   - OwnerAgent          → requestor
//   - Decision            → approval
//   - Audit               → impact + rationale
//
// Optional fields:
//   - Source   → where the entry was extracted from (research.md §C file:line)
//   - EnvGate  → environment condition (GLM-backend rules)
type PolicyEntry struct {
	Tool        string   `yaml:"tool"        json:"tool"`
	ArgsPattern string   `yaml:"args_pattern" json:"args_pattern"`
	RiskTier    RiskTier `yaml:"risk_tier"   json:"risk_tier"`
	Decision    Decision `yaml:"decision"    json:"decision"`
	OwnerAgent  string   `yaml:"owner_agent" json:"owner_agent"`
	Audit       string   `yaml:"audit"       json:"audit"`
	Source      string   `yaml:"source,omitempty"      json:"source,omitempty"`
	EnvGate     *EnvGate `yaml:"env_gate,omitempty"    json:"env_gate,omitempty"`
}

// Metadata carries the YAML header metadata (cross-references + generation
// targets). This block is the audit surface that, together with the generated
// settings.json permissions block (enforcement surface), is derived from the
// same YAML source — structurally preventing YAML↔settings.json drift
// (design.md §D.1 SSOT invariant).
type Metadata struct {
	Version       string   `yaml:"version"        json:"version"`
	GeneratedInto []string `yaml:"generated_into" json:"generated_into"`
	CrossRefs     []string `yaml:"cross_refs,omitempty" json:"cross_refs,omitempty"`
}

// PolicyDocument is the top-level YAML structure.
type PolicyDocument struct {
	Metadata Metadata      `yaml:"metadata" json:"metadata"`
	Entries  []PolicyEntry `yaml:"entries"  json:"entries"`
}

// SettingsSpecifier renders one entry as a Claude Code settings.json
// permission specifier: "Tool" when ArgsPattern is empty or "*", otherwise
// "Tool(args_pattern)". This is the inverse of parsing a settings.json
// permission entry back into (tool, args_pattern) — used by the round-trip
// equivalence test (AC-TPS-004).
func (e PolicyEntry) SettingsSpecifier() string {
	if e.ArgsPattern == "" || e.ArgsPattern == "*" {
		return e.Tool
	}
	return e.Tool + "(" + e.ArgsPattern + ")"
}

// IsValid reports whether a RiskTier is one of the three canonical values.
func (r RiskTier) IsValid() bool {
	switch r {
	case RiskTierRead, RiskTierWrite, RiskTierIrreversible:
		return true
	}
	return false
}

// IsValid reports whether a Decision is one of the three canonical values.
func (d Decision) IsValid() bool {
	switch d {
	case DecisionAllow, DecisionDeny, DecisionAsk:
		return true
	}
	return false
}
