package config

// workflow_accessors.go — Option (c) backward-compat accessors for WorkflowConfig.
//
// These accessor methods on *Config provide the recommended migration path for
// downstream consumers that previously read the deprecated FLAT WorkflowConfig
// fields. Each accessor returns the value from the canonical nested location so
// callers can drop their dependence on the legacy flat identifiers.
//
// See SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 REQ-WSE-005.

// WorkflowAutoClearEnabled returns whether context-window auto-clear is enabled,
// reading from the canonical nested location Workflow.AutoClear.Enabled.
func (c *Config) WorkflowAutoClearEnabled() bool {
	return c.Workflow.AutoClear.Enabled
}

// WorkflowPlanTokens returns the plan-phase token budget from Workflow.TokenBudget.Plan.
func (c *Config) WorkflowPlanTokens() int {
	return c.Workflow.TokenBudget.Plan
}

// WorkflowRunTokens returns the run-phase token budget from Workflow.TokenBudget.Run.
func (c *Config) WorkflowRunTokens() int {
	return c.Workflow.TokenBudget.Run
}

// WorkflowSyncTokens returns the sync-phase token budget from Workflow.TokenBudget.Sync.
func (c *Config) WorkflowSyncTokens() int {
	return c.Workflow.TokenBudget.Sync
}

// WorkflowTeamAutoSelection returns the team auto-selection thresholds from
// Workflow.Team.AutoSelection.
func (c *Config) WorkflowTeamAutoSelection() TeamAutoSelectionConfig {
	return c.Workflow.Team.AutoSelection
}
