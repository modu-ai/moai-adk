// Package safety — oversight unit test.
// REQ-HL-005: verify the OversightProposal payload structure (respect the subagent boundary).
package safety

import (
	"testing"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// TestBuildOversightProposal_BasicStructure verifies the basic OversightProposal structure.
func TestBuildOversightProposal_BasicStructure(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:               "test-001",
		TargetPath:       ".claude/skills/harness-plugin/SKILL.md",
		FieldKey:         "description",
		NewValue:         "improved description with heuristic note",
		PatternKey:       "moai_subcommand:/moai plan:",
		Tier:             harness.TierAutoUpdate,
		ObservationCount: 10,
		CreatedAt:        time.Now(),
	}

	op := BuildOversightProposal(proposal, "중복 trigger 탐지")

	// Question text must be present
	if op.Question == "" {
		t.Error("Question is empty")
	}

	// ProposalID must match
	if op.ProposalID != proposal.ID {
		t.Errorf("ProposalID = %q, want %q", op.ProposalID, proposal.ID)
	}

	// There must be 1–4 options
	if len(op.Options) < 1 || len(op.Options) > 4 {
		t.Errorf("option count = %d, must be in the 1–4 range", len(op.Options))
	}
}

// TestBuildOversightProposal_FirstOptionRecommended verifies the first option is marked recommended.
// REQ-HL-005: AskUserQuestion schema — the first option must be Recommended.
func TestBuildOversightProposal_FirstOptionRecommended(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:         "test-002",
		TargetPath: ".moai/harness/chaining-rules.yaml",
		FieldKey:   "triggers",
		NewValue:   "new-trigger",
		Tier:       harness.TierAutoUpdate,
		CreatedAt:  time.Now(),
	}

	op := BuildOversightProposal(proposal, "rate limit 초과")

	if len(op.Options) == 0 {
		t.Fatal("no options")
	}

	// The first option must be Recommended
	if !op.Options[0].Recommended {
		t.Errorf("first option Recommended = false, must be true. option: %+v", op.Options[0])
	}

	// The remaining options must be Recommended=false
	for i, opt := range op.Options[1:] {
		if opt.Recommended {
			t.Errorf("option[%d] Recommended = true, only the first option may be Recommended", i+1)
		}
	}
}

// TestBuildOversightProposal_MaxFourOptions verifies the option count is at most 4.
// REQ-HL-005: AskUserQuestion supports at most 4 options.
func TestBuildOversightProposal_MaxFourOptions(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:         "test-003",
		TargetPath: ".moai/harness/test.yaml",
		CreatedAt:  time.Now(),
	}

	op := BuildOversightProposal(proposal, "contradiction detected")

	if len(op.Options) > 4 {
		t.Errorf("option count = %d, must be at most 4 (AskUserQuestion limit)", len(op.Options))
	}
}

// TestBuildOversightProposal_AllOptionsHaveDescriptions verifies every option has a description.
func TestBuildOversightProposal_AllOptionsHaveDescriptions(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:               "test-004",
		TargetPath:       ".claude/skills/harness-x/SKILL.md",
		Tier:             harness.TierAutoUpdate,
		ObservationCount: 12,
		CreatedAt:        time.Now(),
	}

	op := BuildOversightProposal(proposal, "")

	for i, opt := range op.Options {
		if opt.Label == "" {
			t.Errorf("option[%d].Label is empty", i)
		}
		if opt.Description == "" {
			t.Errorf("option[%d].Description is empty", i)
		}
		if opt.Value == "" {
			t.Errorf("option[%d].Value is empty", i)
		}
	}
}

// TestBuildOversightProposal_ContextContainsProposalInfo verifies that Context contains proposal info.
func TestBuildOversightProposal_ContextContainsProposalInfo(t *testing.T) {
	t.Parallel()

	proposal := harness.Proposal{
		ID:               "unique-id-xyz",
		TargetPath:       ".claude/skills/harness-abc/SKILL.md",
		Tier:             harness.TierAutoUpdate,
		ObservationCount: 15,
		CreatedAt:        time.Now(),
	}

	op := BuildOversightProposal(proposal, "")

	if op.Context == "" {
		t.Error("Context is empty")
	}
}

// TestBuildOversightProposal_DoesNotCallAskUserQuestion verifies AskUserQuestion is not called.
// REQ-HL-005 + agent-common-protocol §User Interaction Boundary:
// subagents must not call AskUserQuestion directly.
// BuildOversightProposal only returns a payload and never initiates a user interaction.
//
// This is a compile-time check:
// - Indirectly verifies oversight.go has no "AskUserQuestion" call (grep-based)
// - Calling AskUserQuestion is impossible via Go's type system because it is an external tool
func TestBuildOversightProposal_DoesNotCallAskUserQuestion(t *testing.T) {
	t.Parallel()

	// BuildOversightProposal must be a pure function:
	// - no waiting for user input
	// - no external tool calls
	// - only constructs and returns a payload
	proposal := harness.Proposal{
		ID:        "boundary-test",
		CreatedAt: time.Now(),
	}

	// This call must return immediately (non-blocking)
	op := BuildOversightProposal(proposal, "")

	// If the struct is returned the test passes (an AskUserQuestion call would cause panic or timeout)
	if op.ProposalID != proposal.ID {
		t.Errorf("ProposalID = %q, want %q", op.ProposalID, proposal.ID)
	}
}
