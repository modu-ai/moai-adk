package config

import (
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// project_continuation.go — the single read point for
// workflow.project.continuation (SPEC-PROJECT-CONTINUATION-KEY-001
// REQ-PCK-001 / REQ-PCK-002 / REQ-PCK-003).
//
// The key governs the /moai project Phase 14 completion behaviour: whether a
// derived `[PROJECT] ` backlog card is issued, and how far the recommended
// next-steps option carries the session. It NEVER governs whether the Phase 14
// question is asked, and no value of it pre-authorizes run-phase entry
// (REQ-PCK-007 / REQ-PCK-008) — those are prose invariants the orchestrator
// holds, and this accessor is only the value they read.
//
// @MX:NOTE: [AUTO] the continuation enum is consumed by prose (doc-generation.md), not by a Go caller that branches on it
// @MX:SPEC: SPEC-PROJECT-CONTINUATION-KEY-001

// ProjectContinuation enum (workflow.project.continuation in workflow.yaml).
// The tokens are plain lowercase strings so a YAML round-trip preserves the
// user-authored value verbatim, matching the AuditModel* convention.
const (
	// ProjectContinuationNone skips the Step 4.1.5 card issuance entirely and
	// restores the pre-P1 next-steps option set.
	ProjectContinuationNone = "none"
	// ProjectContinuationCard is the default: issue the card, and recommend a
	// branch that carries the session as far as `/moai plan` and no further.
	ProjectContinuationCard = "card"
	// ProjectContinuationPipeline issues the same card, and recommends a branch
	// that carries the session past `/moai plan` to EMITTING the Implementation
	// Kickoff Approval gate. It emits that gate; it never answers it.
	ProjectContinuationPipeline = "pipeline"
)

// ProjectContinuation resolves workflow.project.continuation to a token of the
// closed domain, and reports the offending value when one was configured but
// matched no token.
//
// The type is a plain string, NOT a *string. WorkflowTodoConfig.Enabled is a
// pointer because a bool has only two states, so "absent" and "explicitly
// false" are otherwise indistinguishable. That reasoning does not transfer: the
// default here is a NAMED token of the domain — card — so absent and `card`
// mean exactly the same thing, and no caller can pose a question whose answer
// requires telling them apart. A pointer would add a nil case nothing reads.
//
// Nil-receiver safe. An unmatched value resolves to card and is REPORTED
// rather than applied or fatal (REQ-PCK-003): stopping a whole /moai project
// run over a mistyped presentation preference is disproportionate — the stop
// in the sync-strategy precedent is bought by push/PR irreversibility this key
// does not have — but a silent fallback would hide the typo for the life of
// the project.
func (c *Config) ProjectContinuation() (value string, unmatched string) {
	if c == nil {
		return ProjectContinuationCard, ""
	}
	configured := c.Workflow.Project.Continuation
	if configured == "" {
		return ProjectContinuationCard, ""
	}
	for _, valid := range ValidProjectContinuations() {
		if configured == valid {
			return configured, ""
		}
	}
	return ProjectContinuationCard, configured
}

// ProjectContinuationForRoot is the project-root convenience form, mirroring
// TodoEnabledForRoot. projectRoot is the project directory (the parent of
// .moai), NOT the .moai directory itself.
//
// Every failure path resolves to card with nothing to report: an empty root, a
// missing .moai tree, and a load error all mean the key could not be read,
// which is the same state as the key being absent.
func ProjectContinuationForRoot(projectRoot string) (value string, unmatched string) {
	if projectRoot == "" {
		return ProjectContinuationCard, ""
	}
	cfg, err := NewLoader().Load(filepath.Join(projectRoot, defs.MoAIDir))
	if err != nil {
		return ProjectContinuationCard, ""
	}
	return cfg.ProjectContinuation()
}
