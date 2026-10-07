package cli

// doctor_git_strategy_workflow.go — SPEC-GITSTRAT-WORKFLOW-READER-001 M3
// (card t656).
//
// The production consumer of the git-strategy workflow interpretation table
// (REQ-GWS-009): it reports the configured flow's validity, its
// standing-branch interpretation, and the resolved integration target (or
// the empty-key fallback). Read-only — repair is left to the user (edit the
// YAML); nothing here rewrites git-strategy.yaml.
//
// Precedent: doctor_worktree_base.go (SPEC-WORKTREE-BASEREF-001 M4).

import (
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
)

// workflowStandingBranches is the per-flow standing-branch interpretation,
// the phrase the doctor item reports for a valid flow. It describes how the
// flow uses its branches — this SPEC interprets values, it does not automate
// the flows (spec.md §F).
func workflowStandingBranches(flow string) string {
	switch flow {
	case config.WorkflowGitHubFlow:
		return "single standing branch main; topic branches are short-lived"
	case config.WorkflowGitFlow:
		return "standing branches develop + main; release branches cut for each release"
	case config.WorkflowGitLabFlow:
		return "one long-lived branch per deploy environment (environment key)"
	case config.WorkflowReleaseFlow:
		return "release/* branches cut per release from the integration branch"
	}
	// Unreachable through checkGitStrategyWorkflow: the disposition gate
	// routes every non-allowed value to the invalid state before this runs.
	return ""
}

// knownEnvironmentLabels are the SHIPPED DEFAULT values of the environment
// key (defaults.go): environment LABELS, not branch names. A gitlab-flow
// adopter who leaves the default gets a non-branch string as the interpreted
// target — a configuration-quality note, not a load failure (spec.md §G R2).
var knownEnvironmentLabels = map[string]bool{"local": true, "github": true}

// checkGitStrategyWorkflow reports the configured
// git_strategy.<mode>.workflow through the validated reader. Every state
// names what was read — no silent pass (plan.md §F M3).
func checkGitStrategyWorkflow(projectRoot string, verbose bool) DiagnosticCheck {
	check := DiagnosticCheck{Name: "Git Strategy Workflow"}

	gitFlow := config.LoadGitFlowIntegrationConfig(projectRoot)

	// Unknown: the file is unreadable or the mode has no active profile.
	// Distinguished from "value invalid" per acceptance.md §D.3.
	if gitFlow.Disposition == "" {
		check.Status = uikit.CheckOK
		check.Message = "git_strategy workflow unreadable or no active profile — nothing to interpret"
		if verbose {
			check.Detail = "set git_strategy.mode and git_strategy.<mode>.workflow in .moai/config/sections/git-strategy.yaml"
		}
		return check
	}

	if gitFlow.Disposition == config.DispositionInvalid {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf(
			"git_strategy workflow = %q is not one of the allowed flows (%s) — repair by editing .moai/config/sections/git-strategy.yaml",
			gitFlow.Workflow, strings.Join(config.AllowedWorkflows(), ", "))
		if verbose {
			check.Detail = "the value is treated as non-git-flow by consumers; `trunk-based` is deliberately unsupported"
		}
		return check
	}

	target := gitFlow.IntegrationTarget

	// An empty target is no longer a caller fallback: worktree done / sweep,
	// goal approve, factory merge ready / complete and the card diff base all
	// refuse it (card t1453 M1). Warn, naming the key to set.
	if target == "" {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf(
			"git_strategy workflow = %s; integration target unset — the base-branch consumers refuse it: %s",
			gitFlow.Workflow, gitFlow.EmptyTargetGuidance(projectRoot, ""))
		if verbose {
			check.Detail = "resolved through the validated reader (LoadGitFlowIntegrationConfig); no consumer substitutes a branch"
		}
		return check
	}

	// A target that differs from worktree_base_branch is the signature of a
	// primary-config revert (the flow and the base branch disagree). Only when
	// both are set — an unset base branch is a deliberate "follow the default".
	// Since SPEC-GITHUB-FLOW-CI-RESIDUE-001 the LANDING checks resolve their
	// base from the worktree_base_branch chain, not the target, so this stays
	// an informational mismatch warning — naming both values, no repair
	// imperative (design D-1.3).
	if wtBase := config.LoadWorktreeBaseBranch(projectRoot); wtBase != "" && wtBase != target {
		check.Status = uikit.CheckWarn
		check.Message = fmt.Sprintf(
			"git_strategy workflow = %s resolves integration target %s but git_strategy.worktree_base_branch is %s — landing checks follow the worktree_base_branch chain (else refs/remotes/origin/HEAD, else origin/main), not the target; informational (a reverted config reads this way)",
			gitFlow.Workflow, target, wtBase)
		if verbose {
			check.Detail = fmt.Sprintf("landing checks resolve their base from worktree_base_branch first (here %s) while the %s flow names the target %s; card worktrees are cut from %s", wtBase, gitFlow.Workflow, target, wtBase)
		}
		return check
	}

	check.Status = uikit.CheckOK
	check.Message = fmt.Sprintf(
		"git_strategy workflow = %s; %s; integration target: %s",
		gitFlow.Workflow, workflowStandingBranches(gitFlow.Workflow), target)

	if verbose {
		check.Detail = "resolved through the validated reader (LoadGitFlowIntegrationConfig)"
		if gitFlow.Workflow == config.WorkflowGitLabFlow && knownEnvironmentLabels[target] {
			check.Detail = fmt.Sprintf("environment = %q is a shipped default label, not a branch name — set a real branch in git-strategy.yaml", target)
		}
	}
	return check
}
