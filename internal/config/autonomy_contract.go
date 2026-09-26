package config

import "errors"

// autonomy_contract.go — the single read point for workflow.autonomy.*
// (SPEC-AUTONOMY-CONTRACT-001 REQ-CONTRACT-015). Distinct from autonomy.go,
// which reads the MOAI_AUTONOMY_TIER environment key; the two are unrelated.

// Closed value sets for workflow.autonomy.*.
const (
	AutonomyModeGuided   = "guided"
	AutonomyModeContract = "contract"

	AutonomySecondReviewRequired = "required"
	AutonomySecondReviewAdvisory = "advisory"
	AutonomySecondReviewOff      = "off"

	AutonomyDeciderHuman  = "human"
	AutonomyDeciderLLM    = "llm"
	AutonomyDeciderLLMJev = "llm+jev"
	// AutonomyDeciderJev is not a member of the decider set. It is named so
	// the resolver can report it as a configuration error rather than as an
	// ordinary out-of-set value.
	AutonomyDeciderJev = "jev"
)

// ErrKickoffDeciderJevSole is reported (never returned as a load failure) when
// workflow.autonomy.kickoff.decider is "jev".
var ErrKickoffDeciderJevSole = errors.New("workflow.autonomy.kickoff.decider: jev is never a sole decider; use llm+jev")

// AutonomyBudget is the effective escalation budget default.
type AutonomyBudget struct {
	Turns        int
	Operations   int
	AuditRetries int
}

// AutonomySettings is the effective workflow.autonomy configuration.
type AutonomySettings struct {
	Mode             string
	BatchSign        bool
	SecondReview     string
	PushDevelop      bool
	Decider          string
	DeciderDerived   bool
	JevMinConfidence float64
	BudgetDefault    AutonomyBudget
	JevEnabled       bool
	Warnings         []string
	DeciderError     error
}

// ResolveAutonomy resolves workflow.autonomy.* to its effective values.
func ResolveAutonomy(wf WorkflowConfig) AutonomySettings {
	return AutonomySettings{}
}
