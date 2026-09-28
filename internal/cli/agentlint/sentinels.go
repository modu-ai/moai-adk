package agentlint

// Sentinel keys for structured error identification (SPEC-V3R2-ORC-004).
//
// These keys appear in lint violation messages to enable programmatic detection
// by CI systems, pre-commit hooks, and downstream tooling.
//
// @MX:ANCHOR: [AUTO] agentlint sentinel keys — invariant contract for lint enforcement
// @MX:REASON: High fan_in: agent_lint.go (LR-05/LR-09) + agent_lint_test.go reference these constants.
const (
	// SentinelWorktreeMissing is emitted by LR-05 when a write-heavy agent
	// lacks 'isolation: worktree' in its frontmatter.
	SentinelWorktreeMissing = "ORC_WORKTREE_MISSING"

	// SentinelWorktreeOnReadonly is emitted by LR-09 when a read-only agent
	// (permissionMode: plan) has 'isolation: worktree' set — prohibited overhead.
	SentinelWorktreeOnReadonly = "ORC_WORKTREE_ON_READONLY"

	// The workflow-lint sentinel MODEL_ROUTING_INVALID was retired with the
	// model_routing_profiles check (SPEC-AGENT-MODEL-INHERIT-001 D12).
)
