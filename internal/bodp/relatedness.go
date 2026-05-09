// Package bodp implements the Branch Origin Decision Protocol (BODP).
//
// @MX:NOTE BODP는 새 슬래시 명령어/CLI 서브명령어 ZERO 원칙. 3개 entry point
// 공유 라이브러리 (`/moai plan --branch`, `/moai plan --worktree`, `moai worktree new`).
// SPEC: SPEC-V3R3-CI-AUTONOMY-001 W7-T01 (BODP relatedness check).
package bodp

import (
	"os/exec"
)

// Choice represents the BODP recommendation outcome.
type Choice string

// Choice values.
const (
	ChoiceMain     Choice = "main"
	ChoiceStacked  Choice = "stacked"
	ChoiceContinue Choice = "continue"
)

// EntryPoint identifies which invocation path triggered the BODP check.
type EntryPoint string

// EntryPoint values.
const (
	EntryPlanBranch   EntryPoint = "plan-branch"
	EntryPlanWorktree EntryPoint = "plan-worktree"
	EntryWorktreeCLI  EntryPoint = "worktree-cli"
)

// CheckInput is the input passed to Check().
type CheckInput struct {
	CurrentBranch string
	NewSpecID     string
	RepoRoot      string
	EntryPoint    EntryPoint
}

// BODPDecision is the structured output from Check().
type BODPDecision struct {
	SignalA     bool
	SignalB     bool
	SignalC     bool
	Recommended Choice
	Rationale   string
	BaseBranch  string
}

// DefaultBase is the canonical base branch for fresh BODP recommendations.
const DefaultBase = "origin/main"

// gitCommand executes a git subcommand and returns combined stdout.
// Overridable in tests via the test-only setter.
var gitCommand = func(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}

// ghCommand executes a gh subcommand and returns stdout.
// Overridable in tests via the test-only setter.
var ghCommand = func(args ...string) (string, error) {
	out, err := exec.Command("gh", args...).Output()
	return string(out), err
}

// Check evaluates the 3 BODP signals and returns a recommendation.
//
// Stub implementation for the RED phase of W7-T01. The GREEN phase wires
// signal evaluation, the decision matrix, and rationale rendering.
func Check(input CheckInput) (BODPDecision, error) {
	return BODPDecision{}, nil
}
