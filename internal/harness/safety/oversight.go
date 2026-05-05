// Package safety — Layer 5: Human Oversight (REQ-HL-005).
// Generates proposal payload to delegate automatic update approval to orchestrator.
//
// [HARD] This file does not call AskUserQuestion directly.
// agent-common-protocol §User Interaction Boundary:
// Subagents cannot directly interact with users. Orchestrator (MoAI) receives this payload
// and displays it to users via AskUserQuestion.
package safety

import (
	"fmt"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// BuildOversightProposal generates an OversightProposal for the orchestrator to use in AskUserQuestion.
// REQ-HL-005: Tier 4 pattern automatic updates require user approval.
//
// reason is why this proposal requests human oversight (e.g., "contradiction detected", "rate limit exceeded").
//
// [HARD] This function does not call AskUserQuestion. Only returns the payload.
// The orchestrator receives the return value and displays it to users via AskUserQuestion(options=payload.Options).
//
// @MX:ANCHOR: [AUTO] BuildOversightProposal is the L5 subagent→orchestrator interface.
// @MX:REASON: [AUTO] fan_in >= 3: oversight_test.go, pipeline.go, Phase 4 coordinator
func BuildOversightProposal(proposal harness.Proposal, reason string) harness.OversightProposal {
	// Build context information
	contextStr := buildContext(proposal, reason)

	// Question text
	question := fmt.Sprintf(
		"Auto-learning update approval request: modify '%s' field in file '%s'. (%d observations)",
		proposal.FieldKey, proposal.TargetPath, proposal.ObservationCount,
	)
	if proposal.TargetPath == "" {
		question = "Auto-learning update approval request: approve pattern-based configuration change?"
	}

	// AskUserQuestion options configuration (max 4, first recommended)
	// [HARD] First option must have Recommended=true
	options := []harness.OversightOption{
		{
			Label:       "Approve (Recommended)",
			Description: "Approve the proposed auto-update. The system will apply changes immediately.",
			Value:       "approve",
			Recommended: true,
		},
		{
			Label:       "Reject",
			Description: "Reject this update. The system will maintain current settings.",
			Value:       "reject",
			Recommended: false,
		},
		{
			Label:       "Defer 7 days",
			Description: "Reconsider this update after 7 days. Current settings will be maintained.",
			Value:       "defer",
			Recommended: false,
		},
	}

	return harness.OversightProposal{
		Question:   question,
		Options:    options,
		ProposalID: proposal.ID,
		Context:    contextStr,
	}
}

// buildContext builds the Context field of OversightProposal.
func buildContext(proposal harness.Proposal, reason string) string {
	ctx := fmt.Sprintf(
		"Pattern key: %s | Tier: %s | Observations: %d",
		proposal.PatternKey, proposal.Tier.String(), proposal.ObservationCount,
	)
	if proposal.TargetPath != "" {
		ctx += fmt.Sprintf(" | Target file: %s", proposal.TargetPath)
	}
	if proposal.FieldKey != "" {
		ctx += fmt.Sprintf(" | Field to modify: %s", proposal.FieldKey)
	}
	if proposal.NewValue != "" {
		ctx += fmt.Sprintf(" | New value: %s", proposal.NewValue)
	}
	if reason != "" {
		ctx += fmt.Sprintf(" | Reason: %s", reason)
	}
	return ctx
}
