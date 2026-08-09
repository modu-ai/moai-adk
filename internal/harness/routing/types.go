// Package routing implements Loop 0 (observation / Generator) of the
// self-evolving harness: the routing observation ledger.
//
// This package observes /moai routing decisions — request -> subcommand -> mode
// -> tier/harness-level -> eventual outcome — and appends them to
// .moai/state/routing-ledger.jsonl (schema v1, append-only JSONL). It is a
// SEPARATE observation surface from internal/harness/observer.go, which watches
// generated harness-* artifact usage. Per REQ-HEV-009 this package never writes
// to the generated-artifact usage log and never imports that observer's Event
// schema types — the two subjects stay decoupled.
package routing

import "time"

// SchemaVersion is the routing-ledger row schema version stamped on every row
// (REQ-HEV-001). The versioned-row marker lets Loop 1 consumers evolve the
// schema without re-reading unversioned rows.
const SchemaVersion = 1

// LedgerFileName is the canonical ledger file name under .moai/state/. The full
// path .moai/state/routing-ledger.jsonl is covered by the existing .moai/state/
// gitignore rule (REQ-HEV-001) — the ledger DATA is never committed.
const LedgerFileName = "routing-ledger.jsonl"

// pendingPrefix / pendingSuffix bound the per-session pending-row file name:
// routing-pending-<session>.json (REQ-HEV-011, D5 per-session isolation).
const (
	pendingPrefix = "routing-pending-"
	pendingSuffix = ".json"
)

// StalenessThreshold is the pinned age past which a FOREIGN or
// unresolvable-session pending row is swept as abort (REQ-HEV-014). The current
// session's own pending row is never swept — it reroutes (REQ-HEV-010).
const StalenessThreshold = 24 * time.Hour

// Outcome is the closed terminal-outcome enum (REQ-HEV-002).
type Outcome string

const (
	// OutcomeSuccess is derived by DeriveOutcome from a terminal passing signal.
	OutcomeSuccess Outcome = "success"
	// OutcomeFail is derived by DeriveOutcome from a non-zero gate_exit signal.
	OutcomeFail Outcome = "fail"
	// OutcomeAbort is assigned by DeriveOutcome (abort-kind evidence) OR directly
	// by the lazy stale sweep (REQ-HEV-014), bypassing DeriveOutcome.
	OutcomeAbort Outcome = "abort"
	// OutcomeReroute is assigned directly by the same-session reroute
	// (REQ-HEV-010), bypassing DeriveOutcome.
	OutcomeReroute Outcome = "reroute"
)

// EvidenceKind is the closed machine-signal kind enum (REQ-HEV-006). Model
// self-reported prose is NOT an accepted kind — outcome derives from machine
// signals only.
type EvidenceKind string

const (
	// KindGateExit is a gate/command exit code (value "0" terminal-passing;
	// non-zero -> fail).
	KindGateExit EvidenceKind = "gate_exit"
	// KindAuditScore is an auditor verdict score reference.
	KindAuditScore EvidenceKind = "audit_score"
	// KindVerifyPath is a verification-log path reference.
	KindVerifyPath EvidenceKind = "verify_path"
	// KindAbort is the explicit abort marker — a structurally observable abort
	// artifact (killed/interrupted delegation, user interrupt). Inherently
	// terminal (DeriveOutcome precedence 1).
	KindAbort EvidenceKind = "abort"
)

// ValidEvidenceKind reports whether k is a member of the closed kind enum. The
// CLI evidence verb rejects any kind outside this set so the ledger never
// carries a free-text (model-invented) evidence kind.
func ValidEvidenceKind(k EvidenceKind) bool {
	switch k {
	case KindGateExit, KindAuditScore, KindVerifyPath, KindAbort:
		return true
	default:
		return false
	}
}

// EvidenceRef is a single machine evidence reference (REQ-HEV-006). Terminal
// marks a signal eligible to close the row via DeriveOutcome (REQ-HEV-013).
type EvidenceRef struct {
	Kind     EvidenceKind `json:"kind"`
	Value    string       `json:"value"`
	Ref      string       `json:"ref"`
	Terminal bool         `json:"terminal,omitempty"`
}

// AgentUnattributed is the marker recorded in Delegation.Agent when the stopping
// subagent supplied no agent identity at all (REQ-HLE-008). It is deliberately
// NOT the empty string: an unattributed delegation must stay countable, and an
// empty value is indistinguishable from an attributed one that serialized badly.
// The angle brackets keep it outside the space of real agent and spawn names, so
// no consumer can mistake it for one.
const AgentUnattributed = "<unattributed>"

// OutcomeUnknownDelegation is the marker recorded in Delegation.Outcome when no
// outcome signal was observable for a stopping subagent (REQ-HLE-009). Absence
// of a failure signal is not evidence of success, so the seam records this
// rather than inferring "success" — see plan.md §G AP-3.
//
// It is a Delegation-level marker and intentionally distinct from the row-level
// Outcome enum above: a delegation whose outcome was never observed must not be
// confusable with a row that terminated.
const OutcomeUnknownDelegation = "unknown"

// Delegation is one subagent delegation trajectory entry (REQ-HEV-004, design
// delta A4). Blocker carries the structured blocker category when the delegation
// returned a blocker report (a structurally observable artifact); null otherwise.
//
// Agent holds the hook input's agent_type VERBATIM (REQ-HLE-007) — it is not a
// validated enum. Two observed shapes make that explicit:
//
//   - a named spawn puts the spawn NAME in agent_type, not the agent type, so a
//     value like "audit-hle" is a legitimate recording of a plan-auditor spawn;
//   - roughly a third of observed subagent_stop events carry no agent_type, and
//     those are recorded as AgentUnattributed.
//
// Normalizing an unrecognized value onto a catalog name, or dropping the entry,
// is prohibited at the seam (plan.md §G AP-7): discrimination is the consumer's
// job, and doing it here would erase the distinction the consumer needs.
type Delegation struct {
	Agent     string  `json:"agent"`
	CycleType string  `json:"cycle_type,omitempty"`
	Outcome   string  `json:"outcome"`
	Blocker   *string `json:"blocker"`
}

// RoutingPatch carries the six annotatable routing-metadata fields as optional
// values (REQ-HLE-005). A nil field means "leave the existing value untouched";
// only non-nil fields are written. Pointer-optionality is what makes the
// distinction expressible at all — a plain string cannot separate "clear this"
// from "do not touch this".
type RoutingPatch struct {
	MatchedSubcommand *string
	ModeSelected      *string
	Tier              *string
	HarnessLevel      *string
	ClarifyRounds     *int
	ModelClass        *string
}

// IsEmpty reports whether the patch would change nothing.
func (p RoutingPatch) IsEmpty() bool {
	return p.MatchedSubcommand == nil &&
		p.ModeSelected == nil &&
		p.Tier == nil &&
		p.HarnessLevel == nil &&
		p.ClarifyRounds == nil &&
		p.ModelClass == nil
}

// apply writes the patch's non-nil fields onto row and reports whether anything
// changed. An empty string in a non-nil field is treated as "unset" and leaves
// the existing value alone, matching REQ-HLE-005's stated semantics that a
// supplied empty value is not a clear operation.
func (p RoutingPatch) apply(row *PendingRow) bool {
	changed := false
	if p.MatchedSubcommand != nil && *p.MatchedSubcommand != "" {
		row.MatchedSubcommand = *p.MatchedSubcommand
		changed = true
	}
	if p.ModeSelected != nil && *p.ModeSelected != "" {
		v := *p.ModeSelected
		row.ModeSelected = &v
		changed = true
	}
	if p.Tier != nil && *p.Tier != "" {
		v := *p.Tier
		row.Tier = &v
		changed = true
	}
	if p.HarnessLevel != nil && *p.HarnessLevel != "" {
		v := *p.HarnessLevel
		row.HarnessLevel = &v
		changed = true
	}
	if p.ClarifyRounds != nil {
		row.ClarifyRounds = *p.ClarifyRounds
		changed = true
	}
	if p.ModelClass != nil && *p.ModelClass != "" {
		row.ModelClass = *p.ModelClass
		changed = true
	}
	return changed
}

// Row is a finalized routing-ledger row (schema v1, spec.md §D.1). Every field
// whose machine signal may be absent is nullable and defaults to null/zero,
// never to an inferred value (REQ-HEV-006 evidence-or-null).
type Row struct {
	SchemaVersion     int           `json:"schema_version"`
	TS                string        `json:"ts"`
	SessionID         string        `json:"session_id"`
	ModelClass        string        `json:"model_class"`
	RequestDigest     string        `json:"request_digest"`
	RequestClass      string        `json:"request_class"`
	MatchedSubcommand string        `json:"matched_subcommand"`
	ModeSelected      *string       `json:"mode_selected"`
	Tier              *string       `json:"tier"`
	HarnessLevel      *string       `json:"harness_level"`
	ClarifyRounds     int           `json:"clarify_rounds"`
	Outcome           Outcome       `json:"outcome"`
	LoopIterations    int           `json:"loop_iterations"`
	GoalConverged     *bool         `json:"goal_converged"`
	ConvergenceClass  *string       `json:"convergence_class"`
	Delegations       []Delegation  `json:"delegations"`
	EvidenceRefs      []EvidenceRef `json:"evidence_refs"`
}

// PendingRow is the dispatch-time observation state persisted per session at
// .moai/state/routing-pending-<session>.json until finalized (REQ-HEV-011).
// CreatedAt is the staleness anchor consumed by the age-guarded sweep
// (REQ-HEV-014); it is NOT part of the finalized ledger Row.
type PendingRow struct {
	SchemaVersion     int           `json:"schema_version"`
	CreatedAt         time.Time     `json:"created_at"`
	TS                string        `json:"ts"`
	SessionID         string        `json:"session_id"`
	ModelClass        string        `json:"model_class"`
	RequestDigest     string        `json:"request_digest"`
	RequestClass      string        `json:"request_class"`
	MatchedSubcommand string        `json:"matched_subcommand"`
	ModeSelected      *string       `json:"mode_selected"`
	Tier              *string       `json:"tier"`
	HarnessLevel      *string       `json:"harness_level"`
	ClarifyRounds     int           `json:"clarify_rounds"`
	LoopIterations    int           `json:"loop_iterations"`
	GoalConverged     *bool         `json:"goal_converged"`
	ConvergenceClass  *string       `json:"convergence_class"`
	Delegations       []Delegation  `json:"delegations"`
	EvidenceRefs      []EvidenceRef `json:"evidence_refs"`
}

// Finalize materializes a terminal ledger Row from the pending row plus the
// derived (or directly-assigned) outcome. Delegations / evidence_refs are
// normalized to non-nil slices so they serialize as [] rather than null.
func (p PendingRow) Finalize(outcome Outcome) Row {
	dels := p.Delegations
	if dels == nil {
		dels = []Delegation{}
	}
	refs := p.EvidenceRefs
	if refs == nil {
		refs = []EvidenceRef{}
	}
	return Row{
		SchemaVersion:     SchemaVersion,
		TS:                p.TS,
		SessionID:         p.SessionID,
		ModelClass:        p.ModelClass,
		RequestDigest:     p.RequestDigest,
		RequestClass:      p.RequestClass,
		MatchedSubcommand: p.MatchedSubcommand,
		ModeSelected:      p.ModeSelected,
		Tier:              p.Tier,
		HarnessLevel:      p.HarnessLevel,
		ClarifyRounds:     p.ClarifyRounds,
		Outcome:           outcome,
		LoopIterations:    p.LoopIterations,
		GoalConverged:     p.GoalConverged,
		ConvergenceClass:  p.ConvergenceClass,
		Delegations:       dels,
		EvidenceRefs:      refs,
	}
}
