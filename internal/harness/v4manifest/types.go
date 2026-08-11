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

	// Schedule is the OPTIONAL recurring-schedule declaration. Nil means the
	// harness is one-shot (user-initiated only) — the manifest is
	// shape-identical to the pre-schedule baseline (omitempty: no empty/null
	// key is emitted). When present, all three sub-fields are required and
	// Mode MUST be the exact literal "discovery-only" (see Schedule godoc).
	Schedule *Schedule `json:"schedule,omitempty"`

	// Learning is the OPTIONAL harness-run → learning-subsystem wiring
	// declaration (REQ-HRR-001). Nil means the harness does NOT emit
	// improvement findings from its runs — the manifest is shape-identical to
	// the pre-learning baseline (omitempty: no empty/null key is emitted) and
	// the harness is treated as a valid legacy harness (REQ-HRR-010). When
	// present, the block declares that this harness's Runner/specialist emit
	// structured findings that route to the learning subsystem's Tier-4
	// proposal gate via the reserved-namespace `harness_run:` producer.
	Learning *LearningBlock `json:"learning,omitempty"`
}

// LearningBlock declares that a harness emits structured improvement findings
// from its runs and how those findings route into the learning subsystem
// (REQ-HRR-001). The block is OPTIONAL on Manifest; a nil *LearningBlock means
// "this harness does not participate in the run→learning wiring" and is a
// valid legacy harness (REQ-HRR-010).
//
// The 4 fields are the minimal shape agreed in plan.md §D-D1 (minimal-shape
// principle, Enforce Simplicity). The Tier vocabulary is DERIVED from
// harness.Tier.String() (REQ-HRR-002) — see the LearningTier* constants in
// schema.go. Validate does not reject a partially-populated block: zero-value
// fields are accepted at schema level and defaults are applied downstream in
// the M2 findings→proposal mapping (EC-1 policy). The only hard schema-level
// rejection is a non-empty Tier value outside the SSOT vocabulary (parallel
// vocabulary re-introduction, AP-1).
type LearningBlock struct {
	// Enabled reports whether this harness emits improvement findings from
	// its runs. When false, the Runner/specialist findings path is inert.
	Enabled bool `json:"enabled"`

	// Tier is the proposal tier findings from this harness are promoted into.
	// Valid values are the harness.Tier.String() vocabulary
	// {observation, heuristic, rule, auto_update} (REQ-HRR-002). The empty
	// string is accepted as "unset" and defaulted downstream (EC-1). A
	// non-empty value outside the SSOT vocabulary is rejected by Validate
	// (AP-1: no parallel vocabulary like "recommendation"/"approval_required").
	Tier string `json:"tier"`

	// ConfidenceFloor is the minimum confidence a finding must carry to
	// qualify as a proposal candidate. The confidence_floor boundary aligns
	// with the learning subsystem's confidenceThreshold=0.70
	// (plan.md §D-D1). Schema-level Validate does NOT range-check this field
	// — range validation is the doctor's responsibility (REQ-HRR-005, M3).
	ConfidenceFloor float64 `json:"confidence_floor"`

	// MaxFindingsPerRun caps the number of findings emitted per run (noise
	// prevention, plan.md §D-D1). Schema-level Validate does NOT range-check
	// this field; truncation policy is applied downstream in the M2
	// findings→proposal mapping (EC-2).
	MaxFindingsPerRun int `json:"max_findings_per_run"`
}

// Schedule declares an optional recurring schedule for a harness. Scheduled
// runs execute in discovery-only mode: read-only analysis with findings
// persisted to a queue surface — never commits, never pushes, never enters
// run-phase. The Mode field is an explicit machine-checkable invariant
// marker: it is declared verbatim in the manifest and never defaulted or
// inferred by the decoder, so any future write-mode proposal fails
// validation loudly instead of passing silently.
type Schedule struct {
	// Interval is the recurrence interval — a duration form ("30m"), a
	// named cadence ("nightly"), or a cron expression ("0 3 * * *").
	// Non-empty required.
	Interval string `json:"interval"`

	// Mechanism selects the scheduling mechanism: "loop" (native /loop,
	// session-scoped — dies with the session) or "cron" (Cron tools,
	// persistent across sessions).
	Mechanism string `json:"mechanism"`

	// Mode is the execution-mode invariant marker. MUST be the exact
	// literal "discovery-only"; any other value is rejected by Validate.
	Mode string `json:"mode"`
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
