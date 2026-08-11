// Package harnessrun is the harness-run findings → proposal producer.
//
// It is the third producer that reaches the Tier-4 orchestrator approval gate,
// alongside the tier-ladder producer (proposalgen.MapPromotions, keyed by
// PatternBearingEventTypes) and the delegation-map producer
// (delegationmap.BuildCandidates, reserved namespace "delegation_map:"). This
// producer carries harness-run-time artifact friction surfaced by a Runner or
// specialist at run time, keyed by the reserved namespace "harness_run:".
//
// Sibling pattern: the package mirrors delegationmap.BuildCandidates — direct
// ProposalCandidate construction (NOT via proposalgen.MapPromotions), a
// reserved namespace the existing mapper regex rejects by construction, and
// producer-specific fields carried via the ProposalCandidate.Evidence map seam.
//
// SPEC: SPEC-HARNESS-EVO-RUN-REPORT-001 (REQ-HRR-003, REQ-HRR-004, REQ-HRR-007).
package harnessrun

// PatternNamespace is the reserved pattern-key prefix for this producer's
// proposals. It is deliberately NOT a member of
// harness.PatternBearingEventTypes(), so the existing proposalgen format regex
// — derived from that SSOT — rejects every key emitted here by construction.
// Widening the event-type SSOT to make the existing mapper accept these keys
// would silently widen the observer's taxonomy, which is a different subject
// (plan.md §A.3, AP-11). This mirrors delegationmap's "delegation_map:"
// isolation (delegationmap/proposal.go:21).
const PatternNamespace = "harness_run"

// Finding kind vocabulary (REQ-HRR-003). The four kinds are the closed set a
// Runner/specialist may emit; BuildHarnessRunCandidates does not reject an
// unknown kind at the mapping boundary (it surfaces in the pattern_key and
// evidence for review), but the Runner contract limits emission to these.
const (
	KindDrift    = "drift"    // doc/code divergence
	KindGap      = "gap"      // missing capability/step
	KindFriction = "friction" // repeated manual step
	KindDefect   = "defect"   // incorrect behavior
)

// ConservativeConfidenceFloor is the floor-aligned default confidence a Runner
// emits when it has signal but no precise run-time measurement (REQ-HRR-004).
// The numeric value aligns with proposalgen.ConfidenceThreshold (0.70), the
// boundary below which a candidate does not qualify — so a floor-default
// finding remains a borderline candidate the Tier-4 reviewer judges on its
// merits rather than being silently auto-qualified.
//
// This constant is defined HERE, not imported. REQ-HRR-004 forbids reusing
// learner.go's hardcoded defaultConfidence (1.0): the learner's 1.0 is a
// pre-measurement placeholder on a different subsystem's promotion path, and
// routing it through harness-run findings would disguise an unmeasured signal
// as a measured one (AP-2, verification-claim-integrity §1).
const ConservativeConfidenceFloor = 0.70

// Finding is a single improvement signal a harness Runner or specialist emits
// at run time (REQ-HRR-003). The 5-field shape is the standard findings
// contract every Runner return object carries; it maps to a
// proposalgen.ProposalCandidate via BuildHarnessRunCandidates.
//
// Confidence is a run-time value. The producer carries it verbatim and never
// substitutes learner.go's defaultConfidence. When the Runner has no
// measurement basis, it emits ConservativeConfidenceFloor and the Tier-4
// reviewer treats the value as estimated (REQ-HRR-004).
type Finding struct {
	// Surface names the file or artifact the improvement applies to (e.g.,
	// ".claude/workflows/hns-release-update-run.js").
	Surface string `json:"surface"`

	// Kind is the finding category, one of {drift, gap, friction, defect}.
	Kind string `json:"kind"`

	// Summary is a one-line human-readable description of the improvement.
	Summary string `json:"summary"`

	// Confidence is the run-time measured/estimated confidence in [0,1]. It
	// is NOT derived from learner.go's defaultConfidence.
	Confidence float64 `json:"confidence"`

	// SuggestedTier is the proposal tier the finding should be promoted into,
	// one of the actionable harness.Tier.String() vocabulary values
	// {rule, auto_update} (REQ-HRR-002 SSOT derivation).
	SuggestedTier string `json:"suggested_tier"`
}
