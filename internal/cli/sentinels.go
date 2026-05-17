package cli

// Sentinel keys for structured error identification (SPEC-V3R2-ORC-004).
//
// These keys appear in lint violation messages to enable programmatic detection
// by CI systems, pre-commit hooks, and downstream tooling.
//
// @MX:ANCHOR: [AUTO] SentinelWorktreeRequired — invariant contract for workflow.yaml enforcement
// @MX:REASON: High fan_in: agent_lint.go (LR-05/LR-09) + workflow_lint.go + workflow_lint_test.go + agent_lint_test.go all reference these constants.
const (
	// SentinelWorktreeMissing is emitted by LR-05 when a write-heavy agent
	// lacks 'isolation: worktree' in its frontmatter.
	SentinelWorktreeMissing = "ORC_WORKTREE_MISSING"

	// SentinelWorktreeOnReadonly is emitted by LR-09 when a read-only agent
	// (permissionMode: plan) has 'isolation: worktree' set — prohibited overhead.
	SentinelWorktreeOnReadonly = "ORC_WORKTREE_ON_READONLY"

	// SentinelWorktreeRequired is emitted by 'moai workflow lint' when
	// role_profiles implementer/tester/designer have incorrect isolation value.
	SentinelWorktreeRequired = "ORC_WORKTREE_REQUIRED"
)
