// Package harness — Apply OUTCOME capture (SPEC-HARNESS-OUTCOME-CAPTURE-001).
//
// This file holds the in-memory OutcomeRecord composition value and the
// RecordOutcome observer method that persists it to usage-log.jsonl. The record
// captures, for one completed harness Apply: the regression-gate verdict
// (kept | rolled-back), the project-health delta (baseline + candidate
// MetricTriple + regressed dimensions), and the proposal_id correlation key
// reused from the lineage record (DD-3).
//
// HONEST FRAMING (spec.md §A.2 / plan.md DD-5): this is capture + persist ONLY.
// The captured delta is typically Δ=0 for the current markdown-only harness write
// surface. The record makes the Apply outcome OBSERVABLE for downstream Phase5
// analysis (failure-signature clustering, canary-effectiveness) — those consumers
// are downstream and out of scope. The capture is a passive observer write of an
// already-decided outcome; it never alters any Apply decision, never "improves"
// the harness, and never "prevents" a regression.
package harness

// OutcomeRecord is the in-memory composition value for one completed harness
// Apply outcome. RecordOutcome maps it onto an Event of type
// EventTypeApplyOutcome and delegates to the shared RecordExtendedEvent append
// path (DD-1 / DD-6) — no new low-level writer is introduced.
//
// @MX:NOTE: [AUTO] OutcomeRecord is composed inside applyWithRegressionGate from
// values the regression gate already computed (verdict + triples + proposal_id);
// the capture is additive and never feeds back into the Apply decision (C10).
type OutcomeRecord struct {
	// Verdict is the Apply verdict: "kept" | "rolled-back" (REQ-OC-002).
	Verdict string

	// Decision is the transition decision: "approved" | "regression-blocked".
	Decision string

	// ProposalID is the lineage correlation key reused from the lineage record
	// (Proposal.ID — DD-3). No lineage_id field is introduced (C11).
	ProposalID string

	// Baseline is the MetricTriple measured BEFORE the apply.
	Baseline MetricTriple

	// Candidate is the MetricTriple measured AFTER the apply.
	Candidate MetricTriple

	// Regressed is the list of regressed dimensions (e.g. ["coverage"]). Empty
	// on a kept outcome.
	Regressed []string
}

// RecordOutcome serializes the outcome record into a single usage-log.jsonl line
// via the existing RecordExtendedEvent low-level append path (REQ-OC-003, DD-6).
// It maps the OutcomeRecord onto an Event of type EventTypeApplyOutcome carrying
// the verdict / decision / proposal_id / baseline+candidate triples /
// regressed-dimensions in the additive omitempty Outcome* fields.
//
// @MX:ANCHOR: [AUTO] RecordOutcome is the observer entry point for Apply outcomes.
// @MX:REASON: [AUTO] fan_in >= 3: outcome_test.go, applier.go (recordOutcome seam), Phase5 substrate
func (o *Observer) RecordOutcome(rec OutcomeRecord) error {
	evt := Event{
		EventType:           EventTypeApplyOutcome,
		Subject:             "apply:" + rec.ProposalID,
		OutcomeVerdict:      rec.Verdict,
		OutcomeDecision:     rec.Decision,
		OutcomeProposalID:   rec.ProposalID,
		OutcomeBaseTests:    rec.Baseline.TestsPassed,
		OutcomeBaseCoverage: rec.Baseline.Coverage,
		OutcomeBaseLint:     rec.Baseline.LintCount,
		OutcomeCandTests:    rec.Candidate.TestsPassed,
		OutcomeCandCoverage: rec.Candidate.Coverage,
		OutcomeCandLint:     rec.Candidate.LintCount,
		OutcomeRegressed:    rec.Regressed,
	}
	return o.RecordExtendedEvent(evt)
}
