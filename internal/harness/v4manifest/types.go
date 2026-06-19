// Package v4manifest implements the harness-v4 manifest canonical schema and
// the Runner primitive-mapping engine consumed verbatim by the generated
// dynamic-workflow Runner (harness-<name>-run.js).
//
// Schema source of truth: design.md §C (SPEC-V3R6-HARNESS-V4-001). The 8
// top-level fields, the specialist sub-fields, and the Sprint Contract shape
// are reproduced verbatim from that schema. This package performs NO heuristic
// re-derivation of the primitive (REQ-HV4-005 / AC-HV4-005b): the Runner reads
// specialist.primitive and dispatches accordingly.
//
// The package is deliberately separate from the existing learning-subsystem
// internal/harness package (which owns a different manifest.jsonl lineage
// concept) and from internal/manifest (the template-deployment manifest).
package v4manifest

// Manifest is the canonical 8-field harness-v4 manifest (design §C).
// All 8 top-level fields are required (REQ-HV4-006 / AC-HV4-006a).
type Manifest struct {
	// Name is the harness name, derived from the NL request. Matches the
	// /harness:<name> command and the harness-<name>-run.js Runner filename.
	// Constraint: [a-z][a-z0-9-]* (kebab-case, DNS-safe).
	Name string `json:"name"`

	// Domain is a short human-readable domain description.
	Domain string `json:"domain"`

	// SourceRequest is the original natural-language request verbatim,
	// preserved for audit/re-generation.
	SourceRequest string `json:"source_request"`

	// Patterns is the array of pattern names drawn from the 6-pattern
	// catalog (design §E). Selected/combined dynamically by the PLAN phase.
	Patterns []string `json:"patterns"`

	// Specialists is the array of specialist role definitions. MUST be
	// non-empty (>= 1 specialist) per design §C.2.
	Specialists []Specialist `json:"specialists"`

	// SprintContract is the Anthropic-GAN-inspired Sprint Contract
	// (REQ-HV4-008): graded dimensions + thresholds declared pre-coding.
	SprintContract SprintContract `json:"sprint_contract"`

	// EntryCommand is the /harness:<name> string (redundant with Name but
	// explicit for tooling).
	EntryCommand string `json:"entry_command"`

	// RunnerWorkflow is the Runner Workflow filename harness-<name>-run.js.
	RunnerWorkflow string `json:"runner_workflow"`
}

// Specialist is a single specialist role definition within the manifest
// (design §C.1). Each of the 5 sub-fields is required (AC-HV4-005a).
type Specialist struct {
	// Role is the specialist's responsibility (e.g.
	// "template-neutrality-auditor").
	Role string `json:"role"`

	// Primitive is the execution primitive. MUST be exactly one of the
	// 5 primitives (PrimitiveSubAgent / PrimitiveDynamicWorkflow /
	// PrimitiveWorktree / PrimitiveGoal / PrimitiveAdversarialFanOut).
	// The Runner consumes this verbatim (AC-HV4-005b).
	Primitive string `json:"primitive"`

	// Isolation is none (main-tree) or worktree
	// (Agent(isolation:"worktree") sub-agent). Per-specialist, conditional
	// (REQ-HV4-007).
	Isolation string `json:"isolation"`

	// Effort is the reasoning effort level (low/medium/high/xhigh/max).
	Effort string `json:"effort"`

	// Model is the model tier (inherit/haiku/sonnet/opus).
	Model string `json:"model"`
}

// SprintContract is the Generator-Evaluator separation contract
// (REQ-HV4-008). Dimensions is the array of graded dimension names agreed
// pre-coding; Thresholds maps each dimension to its pass threshold.
type SprintContract struct {
	// Dimensions is the array of graded dimension names.
	Dimensions []string `json:"dimensions"`

	// Thresholds maps each dimension name to its pass threshold value.
	// Values are interface{} because threshold types vary by dimension
	// (numeric score, boolean gate, etc.).
	Thresholds map[string]interface{} `json:"thresholds"`
}
