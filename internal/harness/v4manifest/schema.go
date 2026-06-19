// Package v4manifest — schema constants (design §C.2 + §E).
//
// The 5 execution primitives, 2 isolation modes, 5 effort levels, 4 model
// tiers, and 6 pattern-catalog entries are reproduced verbatim from design.md
// §C.2 (schema validation rules) and §E (6-pattern catalog). These sets are
// closed: validation rejects any value not in the set.
package v4manifest

// Execution primitives (design §C.2). The Runner dispatches verbatim per
// specialist.primitive — no heuristic re-derivation (REQ-HV4-005 / AC-HV4-005b).
const (
	PrimitiveSubAgent          = "sub-agent"
	PrimitiveDynamicWorkflow   = "dynamic-workflow"
	PrimitiveWorktree          = "worktree"
	PrimitiveGoal              = "/goal"
	PrimitiveAdversarialFanOut = "adversarial-fan-out"
)

// validPrimitives is the closed set of the 5 execution primitives. A
// specialist.primitive MUST be exactly one of these (design §C.2).
var validPrimitives = map[string]bool{
	PrimitiveSubAgent:          true,
	PrimitiveDynamicWorkflow:   true,
	PrimitiveWorktree:          true,
	PrimitiveGoal:              true,
	PrimitiveAdversarialFanOut: true,
}

// Isolation modes (design §C.2). Per-specialist, conditional (REQ-HV4-007).
const (
	IsolationNone    = "none"
	IsolationWorktree = "worktree"
)

// validIsolations is the closed set of the 2 isolation modes.
var validIsolations = map[string]bool{
	IsolationNone:     true,
	IsolationWorktree: true,
}

// Effort levels (design §C.1). Per SPEC-V3R6-WORKFLOW-EFFORT-MAP-001
// purpose-driven taxonomy. low/medium/high/xhigh/max.
const (
	EffortLow    = "low"
	EffortMedium = "medium"
	EffortHigh   = "high"
	EffortXhigh  = "xhigh"
	EffortMax    = "max"
)

// validEfforts is the closed set of the 5 effort levels.
var validEfforts = map[string]bool{
	EffortLow:    true,
	EffortMedium: true,
	EffortHigh:   true,
	EffortXhigh:  true,
	EffortMax:    true,
}

// Model tiers (design §C.1). inherit/haiku/sonnet/opus.
const (
	ModelInherit = "inherit"
	ModelHaiku   = "haiku"
	ModelSonnet  = "sonnet"
	ModelOpus    = "opus"
)

// validModels is the closed set of the 4 model tiers.
var validModels = map[string]bool{
	ModelInherit: true,
	ModelHaiku:   true,
	ModelSonnet:  true,
	ModelOpus:    true,
}

// 6-pattern catalog (design §E). Patterns are selected/combined dynamically
// by the PLAN phase; the selection is recorded in manifest.patterns.
// AC-HV4-004b requires patterns[] entries to be from this catalog (no custom
// patterns).
const (
	PatternPipeline                 = "Pipeline"
	PatternFanOutFanIn              = "Fan-out/Fan-in"
	PatternExpertPool               = "Expert Pool"
	PatternProducerReviewer         = "Producer-Reviewer"
	PatternSupervisor               = "Supervisor"
	PatternHierarchicalDelegation   = "Hierarchical Delegation"
)

// validPatterns is the closed set of the 6-pattern catalog.
var validPatterns = map[string]bool{
	PatternPipeline:               true,
	PatternFanOutFanIn:            true,
	PatternExpertPool:             true,
	PatternProducerReviewer:       true,
	PatternSupervisor:             true,
	PatternHierarchicalDelegation: true,
}
