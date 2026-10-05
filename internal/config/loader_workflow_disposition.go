package config

// loader_workflow_disposition.go — SPEC-GITSTRAT-WORKFLOW-READER-001 (card
// t656).
//
// The multi-flow extension of the workflow reader in
// loader_integration_branch.go: the 4-value allowed set, the three-way
// disposition, and the D2 flow-scoped integration-target table. The reader
// itself never fails (the t449 fail-open contract) — invalidity is a
// structured disposition a consumer diagnoses, never a load error.

import "strings"

// The allowed workflow values (REQ-GWS-001). Exactly four; `trunk-based` is
// deliberately EXCLUDED by operator decision — a user who sets it gets the
// invalid-value diagnostic, which is where they discover the exclusion.
const (
	WorkflowGitHubFlow  = "github-flow"
	WorkflowGitFlow     = "git-flow"
	WorkflowGitLabFlow  = "gitlab-flow"
	WorkflowReleaseFlow = "release-flow"
)

// githubFlowFixedTarget is github-flow's integration target: the fixed
// default main. github-flow reads no config key for its target.
const githubFlowFixedTarget = "main"

// AllowedWorkflows returns the exactly-four allowed workflow values, in
// their canonical order. Consumers must classify through
// ClassifyWorkflowDisposition rather than re-deriving membership from this
// slice, so validation has one source of truth.
func AllowedWorkflows() []string {
	return []string{WorkflowGitHubFlow, WorkflowGitFlow, WorkflowGitLabFlow, WorkflowReleaseFlow}
}

// WorkflowDisposition is the three-way classification of a
// git_strategy.<mode>.workflow value (REQ-GWS-001). The zero value is NOT
// invalid: it means "unknown" — the file was unreadable or the mode has no
// active profile — which the doctor check must distinguish from "a value is
// present and it is invalid" (acceptance.md §D.3).
type WorkflowDisposition string

const (
	// DispositionGitFlow: the value is exactly git-flow.
	DispositionGitFlow WorkflowDisposition = "git-flow"
	// DispositionValidNonGitFlow: one of the other three allowed flows — a
	// deliberate choice, never warned about (spec.md §C.1 D1).
	DispositionValidNonGitFlow WorkflowDisposition = "valid-non-git-flow"
	// DispositionInvalid: not in the allowed set — typos, trunk-based, an
	// empty value. Consumers fall back exactly as the non-git-flow path does
	// and diagnose (REQ-GWS-003).
	DispositionInvalid WorkflowDisposition = "invalid"
)

// ClassifyWorkflowDisposition classifies a raw workflow value. Matching is
// exact and case-sensitive — the pre-existing git-flow discriminator was
// case-sensitive, and generalizing to the allowed set keeps that contract.
// An empty string classifies invalid (an empty value present), but the
// loader only calls this when an active profile exists, so the zero
// disposition stays reserved for "unknown".
//
// @MX:NOTE: [AUTO] the 3-way disposition SSOT — every consumer of
// git_strategy.<mode>.workflow classifies through here, never by re-matching
// raw values (SPEC-GITSTRAT-WORKFLOW-READER-001 REQ-GWS-001).
func ClassifyWorkflowDisposition(value string) WorkflowDisposition {
	switch value {
	case WorkflowGitFlow:
		return DispositionGitFlow
	case WorkflowGitHubFlow, WorkflowGitLabFlow, WorkflowReleaseFlow:
		return DispositionValidNonGitFlow
	default:
		return DispositionInvalid
	}
}

// WorkflowIntegrationTarget implements the D2 flow-scoped interpretation
// table (REQ-GWS-004): each allowed flow reads exactly one target source, so
// there is no precedence conflict to resolve. Values are trimmed — the t449
// develop_branch trim contract generalized to the new target keys. An empty
// target key yields the empty string, the caller-fallback neutral (REQ-GWS-006).
//
// The develop_branch MANUAL-mode gate (REQ-GWS-005) is NOT applied here —
// this function is the pure value table. LoadGitFlowIntegrationConfig applies
// the gate when it projects the table onto GitFlowIntegrationConfig, exactly
// as it gates DevelopBranch today.
//
// @MX:NOTE: [AUTO] the D2 interpretation table — the doctor check (REQ-GWS-009)
// consumes this through the loader; acquire/automerge stay unwired
// (REQ-GWS-008) and are a follow-up card candidate.
func WorkflowIntegrationTarget(workflow string, profile ModeProfile) string {
	switch workflow {
	case WorkflowGitHubFlow:
		return githubFlowFixedTarget
	case WorkflowGitFlow:
		return strings.TrimSpace(profile.DevelopBranch)
	case WorkflowGitLabFlow:
		return strings.TrimSpace(profile.Environment)
	case WorkflowReleaseFlow:
		return strings.TrimSpace(profile.ReleaseBranchPrefix)
	default:
		return ""
	}
}
