package cli

// factory_target_holder.go — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 M5b (card
// t1831, rescope 0.4.0): the target-holder probe of the delivery routing.
//
// Under agentic-trunk a card lands by the TARGET HOLDER (design.md section
// 1): a worktree OTHER than the primary checkout holding `main` is a
// dedicated integration worktree — the integration-window local merge (the
// M0-M5 behavior). Where the ONLY tree holding `main` is the primary
// checkout, no dedicated landing surface exists — the shared human/leader
// surface never becomes a merge window — so the delivery takes the
// pull-request edge and the LEADER merges. The probe sits beside the
// no-worktree-holds-main refusal (integration_merge.go integrationMergeWorktree):
// an empty probe is NOT primary-only, and the existing refusals keep
// answering that state; an unresolvable primary refuses to guess.

import (
	"github.com/modu-ai/moai-adk/internal/config"
)

// factoryPrimaryHoldsTarget reports whether the ONLY worktree holding
// branch is the repository's primary checkout — the primary-route condition
// of the target-holder routing (REQ-AGT-006, rescope 0.4.0). No tree
// holding the branch, and an unresolvable primary, both answer "no": the
// first keeps the existing no-worktree-holds-main refusals, the second
// refuses to guess.
func factoryPrimaryHoldsTarget(root, branch string) bool {
	holder := factoryWorktreeForBranchIn(root, branch)
	if holder == "" {
		return false
	}
	primary, _, err := identifyPrimaryCheckout(root)
	if err != nil {
		return false
	}
	return factorySameTree(holder, primary)
}

// factoryAgenticTrunkPrimaryRoute reports that the project rooted at root
// runs manual agentic-trunk AND holds its fixed integration target `main`
// nowhere but the primary checkout — the condition under which `factory
// complete` and `factory merge ready` take the pull-request edge instead
// of the integration window (REQ-AGT-006, rescope 0.4.0; design.md
// section 1, candidate B).
func factoryAgenticTrunkPrimaryRoute(root string) bool {
	cfg := config.LoadGitFlowIntegrationConfig(root)
	if !cfg.Manual || !cfg.AgenticTrunkWorkflow || cfg.IntegrationTarget == "" {
		return false
	}
	return factoryPrimaryHoldsTarget(root, cfg.IntegrationTarget)
}
