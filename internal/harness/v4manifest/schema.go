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
	PrimitiveGoal              = "/moai goal"
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
	IsolationNone     = "none"
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

// Schedule mechanisms. "loop" is the native /loop scheduler (session-scoped);
// "cron" is the Cron tools registration (persistent across sessions).
const (
	MechanismLoop = "loop"
	MechanismCron = "cron"
)

// validMechanisms is the closed set of the 2 schedule mechanisms.
var validMechanisms = map[string]bool{
	MechanismLoop: true,
	MechanismCron: true,
}

// ScheduleModeDiscoveryOnly is the sole valid schedule mode literal.
// Scheduled harness runs are discovery-only: read-only analysis, findings
// persisted to a queue surface, no writes/commits/pushes, no run-phase entry.
const ScheduleModeDiscoveryOnly = "discovery-only"

// ─── Learning tier vocabulary (REQ-HRR-002, DERIVED from harness.Tier.String() SSOT) ───
//
// The learning.tier valid-value set is DERIVED from the learning-subsystem
// classifier Tier.String() vocabulary at internal/harness/types.go (the SSOT
// PIPE-REPAIR aligned). These constants reproduce that vocabulary VERBATIM —
// they are NOT a parallel vocabulary. A mechanical cross-check
// (TestLearningTierVocabularyMatchesHarnessSSOT in learning_test.go) imports
// the SSOT and asserts the two sets are identical, so any drift between the
// SSOT and these constants fails the test.
//
// REQ-HRR-002 / AP-1: defining a separate parallel vocabulary here (e.g.
// "recommendation"/"approval_required" — the very values PIPE-REPAIR removed)
// is FORBIDDEN. The closed set is exactly {observation, heuristic, rule,
// auto_update}.
//
// Why a test-only import rather than a production import: the v4manifest
// package is deliberately separate from the learning-subsystem internal/harness
// package (see the types.go package doc) to avoid pulling learning-subsystem
// concerns into the schema layer. The SSOT derivation is enforced at test
// time, not at compile time.
const (
	// LearningTierObservation is the lowest tier — a finding observed but not
	// yet actionable. Pre-actionable (plan.md §D-D1 / REQ-HRR-002).
	LearningTierObservation = "observation"

	// LearningTierHeuristic is a recurring-pattern tier, still pre-actionable.
	LearningTierHeuristic = "heuristic"

	// LearningTierRule is an actionable tier subject to trigger injection.
	LearningTierRule = "rule"

	// LearningTierAutoUpdate is the actionable auto-update candidate tier.
	LearningTierAutoUpdate = "auto_update"
)

// validLearningTiers is the closed set of learning.tier values derived from
// harness.Tier.String(). A non-empty learning.tier MUST be exactly one of
// these (REQ-HRR-002). The empty string is accepted as "unset" (EC-1
// partial-block policy).
//
// @MX:ANCHOR: [AUTO] learning.tier vocabulary SSOT — derived from harness.Tier.String()
// @MX:REASON: [AUTO] fan_in >= 3 candidate: Validate, TestLearningTierVocabularyMatchesHarnessSSOT, future M2 BuildHarnessRunCandidates; REQ-HRR-002 forbids a parallel vocabulary.
// @MX:SPEC: SPEC-HARNESS-EVO-RUN-REPORT-001 M1 / REQ-HRR-002
var validLearningTiers = map[string]bool{
	LearningTierObservation: true,
	LearningTierHeuristic:   true,
	LearningTierRule:        true,
	LearningTierAutoUpdate:  true,
}

// 6-pattern catalog (design §E). Patterns are selected/combined dynamically
// by the PLAN phase; the selection is recorded in manifest.patterns.
// AC-HV4-004b requires patterns[] entries to be from this catalog (no custom
// patterns).
const (
	PatternPipeline               = "Pipeline"
	PatternFanOutFanIn            = "Fan-out/Fan-in"
	PatternExpertPool             = "Expert Pool"
	PatternProducerReviewer       = "Producer-Reviewer"
	PatternSupervisor             = "Supervisor"
	PatternHierarchicalDelegation = "Hierarchical Delegation"
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
