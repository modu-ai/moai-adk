package statusline

// workflow_provider.go — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001 M3 (card
// t1831).
//
// The provider behind the branch segment's workflow label (REQ-AGT-013). It
// reads git-strategy.yaml through the validated reader at the project root
// the render already anchors its state to, and answers the workflow value
// only for a valid non-github-flow value — github-flow, an absent or
// unreadable file, and an invalid value answer "" so those renders stay
// byte-identical to the pre-label bar. The read is one small yaml read: the
// provider spawns no git process (REQ-AGT-013's no-invocation clause).

import "github.com/modu-ai/moai-adk/internal/config"

// resolveWorkflowLabel reports the workflow value the active git-strategy
// profile reads at root, for the states the label is allowed to name. Every
// other state — github-flow, unreadable file, no active profile, an invalid
// value — answers "", the withheld label.
func resolveWorkflowLabel(root string) string {
	if root == "" {
		return ""
	}
	c := config.LoadGitFlowIntegrationConfig(root)
	switch c.Disposition {
	case config.DispositionGitFlow, config.DispositionValidNonGitFlow:
		if c.Workflow == config.WorkflowGitHubFlow {
			return ""
		}
		return c.Workflow
	default:
		// The zero disposition (unknown: unreadable file / no active
		// profile) and DispositionInvalid both withhold the label.
		return ""
	}
}
