package agentlint

// Sentinel keys for structured error identification (SPEC-V3R2-ORC-004).
//
// These keys appear in lint violation messages to enable programmatic detection
// by CI systems, pre-commit hooks, and downstream tooling.
//
// @MX:ANCHOR: [AUTO] agentlint sentinel keys — invariant contract for lint enforcement
// @MX:REASON: High fan_in: agent_lint.go (LR-05/LR-09) + workflow_lint.go + workflow_lint_test.go + agent_lint_test.go all reference these constants.
const (
	// SentinelWorktreeMissing is emitted by LR-05 when a write-heavy agent
	// lacks 'isolation: worktree' in its frontmatter.
	SentinelWorktreeMissing = "ORC_WORKTREE_MISSING"

	// SentinelWorktreeOnReadonly is emitted by LR-09 when a read-only agent
	// (permissionMode: plan) has 'isolation: worktree' set — prohibited overhead.
	SentinelWorktreeOnReadonly = "ORC_WORKTREE_ON_READONLY"

	// SentinelModelRoutingInvalid is emitted by 'moai workflow lint' when a
	// workflow.model_routing_profiles entry violates the closed sets
	// (SPEC-AGENT-TEAM-RETIRE-001 REQ-ATR-009 — replaces the retired
	// Agent Teams role_profiles isolation check ORC_WORKTREE_REQUIRED).
	SentinelModelRoutingInvalid = "MODEL_ROUTING_INVALID"
)
