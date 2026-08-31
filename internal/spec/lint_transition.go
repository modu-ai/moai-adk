package spec

import (
	"fmt"
	"strings"
)

// StatusTransitionValidityRule judges the (previousStatus, currentStatus) pair
// of a SPEC's most recent `status:` transition against the canonical lifecycle
// edge set, independently of who authored the commit.
//
// It is deliberately NOT a variant of OwnershipTransitionRule, which sits
// beside it in this package and asks a different question — *did the right
// agent sign it?* — and which returns nil before it can judge anything in two
// cases this rule must not inherit:
//
//   - `expected == ownerNone` (the pair is undefined in the ownership matrix),
//     which is exactly the shape a status reversal or a phase skip produces;
//   - `rec.AuthoredByAgent == ""` (the commit carries no trailer), which
//     silently exempts every legacy and non-MoAI commit.
//
// This rule reads the same extracted record but reaches its verdict from the
// pair alone (REQ-STV-002), so a trailer-less commit is judged identically to
// one carrying a trailer.
//
// Two codes, because two different facts are being reported:
//
//   - StatusTransitionInvalid  — the pair is outside the canonical set
//   - StatusTokenUnrecognized  — a side names a token outside the 8-value enum
//
// Reporting the second as the first would make the message misstate its own
// cause. Both are emitted at warning severity with NO emission-site Advisory
// flag (REQ-STV-009), so a finding on a modern-era, non-terminal SPEC is
// escalated by --strict.
//
// Observation-only: never mutates a SPEC, and reuses the existing git lookup
// rather than adding a second history walk (REQ-STV-012).
//
// SPEC-STATUS-TRANSITION-VALIDITY-001 (card t376), REQ-STV-001..015.
//
// @MX:ANCHOR: [AUTO] StatusTransitionValidityRule — judges the transition pair, not its author
// @MX:REASON: the ownership matrix delegates every matrix-undefined pair to a rule whose signature carries no previous status, so `draft → completed` and `completed → draft` were both silent; this rule is the only consumer that judges the pair itself
type StatusTransitionValidityRule struct{}

// Code returns the primary lint finding code. The rule also emits
// StatusTokenUnrecognized; see the type comment for why the two are distinct.
func (r *StatusTransitionValidityRule) Code() string { return "StatusTransitionInvalid" }

// canonicalStatusTransitions is the set of legal lifecycle edges, derived from
// `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status
// Transition Ownership Matrix. It introduces no parallel definition of the DAG;
// each entry cites the matrix row it comes from.
//
// Edges with a terminal TARGET (`* → superseded|archived|rejected`) are a
// target-side test rather than pair entries — see isCanonicalStatusTransition.
// Edges naming the legacy-optional `planned` status are tolerated by the same
// function, because § Status Enum declares it to have no active-flow owner.
var canonicalStatusTransitions = map[[2]string]bool{
	{"draft", "in-progress"}:       true, // matrix row 2
	{"in-progress", "implemented"}: true, // matrix row 3 (in-progress → implemented → completed)
	{"implemented", "completed"}:   true, // matrix row 3
	{"in-progress", "completed"}:   true, // matrix row 3 + its single-sync-commit note — spec.md §A.5 D4
	{"completed", "in-progress"}:   true, // matrix row 7 — the declared amendment edge
	// spec.md §A.5 D1: adopted from expectedOwnerForTransition (lint_ownership.go),
	// which already assigns `draft|planned → implemented` to manager-develop.
	// The SSOT matrix carries NO such row; the divergence is recorded in
	// spec.md §C as a later card, not closed here.
	{"draft", "implemented"}: true,
}

// terminalTransitionTargets are the lifecycle end states reachable from any
// status (matrix rows 4, 5, 6).
var terminalTransitionTargets = map[string]bool{
	"superseded": true,
	"archived":   true,
	"rejected":   true,
}

// isCanonicalStatusTransition reports whether the (prev, curr) pair is a legal
// lifecycle edge. Both sides are assumed to be recognized enum members — the
// token check runs first (see Check).
func isCanonicalStatusTransition(prev, curr string) bool {
	// `* → superseded|archived|rejected` — matrix rows 4-6.
	if terminalTransitionTargets[curr] {
		return true
	}
	// The legacy-optional `planned` status has no active-flow owner
	// (§ Status Enum), and grandfathered history carries it on both sides.
	if prev == "planned" || curr == "planned" {
		return true
	}
	return canonicalStatusTransitions[[2]string{prev, curr}]
}

// Check inspects one SPEC document's most recent recorded status transition.
//
// Check order is load-bearing, because it decides which code a document gets:
//
//  1. `(none)` skip — an added `status:` line with no removed counterpart is
//     the shape a squash merge and a truncated history both produce, so the
//     pair is not evidence of a transition at all (REQ-STV-014, §A.5 D5).
//  2. token recognition — a side outside the 8-value enum gets its own code
//     and stops here (REQ-STV-013, §A.5 D3).
//  3. pair validity — the remaining pairs are judged against the canonical set.
//
// Behavior at the edges:
//   - empty id/status: skipped (FrontmatterSchemaRule's responsibility)
//   - git unreachable, or no transition in history: silent, at no severity
//     (REQ-STV-007). OwnershipTransitionRule already reports unreachable git as
//     an Info finding; duplicating it here would report one fact twice.
func (r *StatusTransitionValidityRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	fm := doc.Frontmatter
	if fm.ID == "" || fm.Status == "" {
		return nil
	}

	rec, err := getOwnershipTransitionRunner(doc.Path, fm.ID)
	if err != nil || rec == nil {
		// REQ-STV-007: unreachable git and no-transition are both silent here.
		return nil
	}

	// (1) `(none) → X` is unjudgeable — skip both checks (§A.5 D5).
	if rec.PreviousStatus == "" {
		return nil
	}

	// (2) An unrecognized token is a different fact with a different remedy.
	var unrecognized []string
	for _, tok := range []string{rec.PreviousStatus, rec.CurrentStatus} {
		if !IsValidStatus(tok) {
			unrecognized = append(unrecognized, fmt.Sprintf("%q", tok))
		}
	}
	if len(unrecognized) > 0 {
		return []Finding{
			{
				File:     doc.Path,
				Line:     1,
				Severity: SeverityWarning,
				Code:     "StatusTokenUnrecognized",
				Message: fmt.Sprintf(
					"SPEC %s status transition %q → %q names %s outside the canonical 8-value enum %v (commit %s)",
					fm.ID,
					rec.PreviousStatus,
					rec.CurrentStatus,
					strings.Join(unrecognized, " and "),
					ValidStatuses,
					rec.CommitSHA,
				),
			},
		}
	}

	// (3) The pair itself.
	if isCanonicalStatusTransition(rec.PreviousStatus, rec.CurrentStatus) {
		return nil
	}

	return []Finding{
		{
			File:     doc.Path,
			Line:     1,
			Severity: SeverityWarning,
			Code:     "StatusTransitionInvalid",
			Message: fmt.Sprintf(
				"SPEC %s status transition %q → %q is not a canonical lifecycle edge (commit %s)",
				fm.ID,
				rec.PreviousStatus,
				rec.CurrentStatus,
				rec.CommitSHA,
			),
		},
	}
}
