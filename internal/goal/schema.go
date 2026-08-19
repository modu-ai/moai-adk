// Package goal implements the MoAI goal engine: a per-session, condition-declared
// agentic loop evaluator. The Stop hook (moai hook stop-goal) loads a session's
// goal state, evaluates its conditions each turn, and emits a block decision until
// the conditions hold or a turn ceiling is reached.
//
// The package owns: the goal-state JSON schema (schema.go), atomic per-session
// state persistence (state.go), orphan pruning (prune.go), and the 2-tier
// evaluator (evaluate.go). The hook verb in internal/cli wires stdin/stdout.
package goal

// ConditionType enumerates the two condition tiers.
// Tier 1 (mechanical) conditions run a shell command and compare its exit code.
// Tier 2 (model) conditions are surfaced in the block reason for orchestrator
// self-evaluation (settled Option B — stop-goal itself runs no model call).
type ConditionType string

const (
	// ConditionMechanical is a Tier-1 condition: a shell command whose exit
	// code is compared against ExpectExit.
	ConditionMechanical ConditionType = "mechanical"
	// ConditionModel is a Tier-2 condition: a transcript-measurable claim the
	// orchestrator evaluates against conversation-surfaced evidence.
	ConditionModel ConditionType = "model"
)

// Condition is a single declared completion condition. A mechanical condition
// carries Cmd + ExpectExit; a model condition carries Claim.
type Condition struct {
	Type       ConditionType `json:"type"`
	Cmd        string        `json:"cmd,omitempty"`
	ExpectExit int           `json:"expect_exit,omitempty"`
	Claim      string        `json:"claim,omitempty"`
}

// Ceiling bounds the goal loop. MaxTurns defaults to DefaultMaxTurns (30);
// MaxTurns == 0 is the infinite entry point (the evaluate.go `> 0` guard
// disables the turn ceiling check). MaxDuration (wall-clock seconds since
// CreatedAt) is the M-default PRIMARY real bound when MaxTurns == 0
// (SPEC-INFINITE-GOAL-001 REQ-4 / OQ-2). CostCap records a cost bound
// (max invocations); its enforcement is a documented follow-up because the
// evaluator carries no invocation/token accounting today — the field is
// recorded here so the arm-time bound is captured verbatim, per REQ-4
// ("recorded in Ceiling alongside MaxTurns").
type Ceiling struct {
	MaxTurns    int `json:"max_turns"`
	MaxDuration int `json:"max_duration,omitempty"`
	CostCap     int `json:"cost_cap,omitempty"`
}

// ProgressEntry is one append-only turn record in the goal's progress log.
type ProgressEntry struct {
	Turn int    `json:"turn"`
	Note string `json:"note"`
	// Fingerprint carries the per-turn mechanical-condition fingerprint
	// (SPEC-INFINITE-GOAL-001 REQ-4 strengthened stagnation guard). Empty for
	// legacy/ceiling/satisfied entries; the stagnation guard falls back to Note
	// comparison when Fingerprint is empty (backward compat with pre-M4 entries).
	Fingerprint string `json:"fingerprint,omitempty"`
}

// Status is the lifecycle state of a goal.
type Status string

const (
	// StatusArmed: the goal is active and the evaluator blocks until satisfied.
	StatusArmed Status = "armed"
	// StatusSatisfied: all conditions hold; the loop ends.
	StatusSatisfied Status = "satisfied"
	// StatusCeilingExit: the turn ceiling was reached before satisfaction; the
	// evaluator emits a 5-section verdict and stops blocking.
	StatusCeilingExit Status = "ceiling-exit"
	// StatusCleared: the goal was explicitly cleared (moai goal clear).
	StatusCleared Status = "cleared"
)

// ProgressionMode selects post-approval progression behavior. It is chosen at
// the Implementation Kickoff Approval gate (a DISTINCT axis from approve/decline)
// and persisted in goal state. The gate stays mandatory in both modes.
type ProgressionMode string

const (
	// ProgressionAutonomous is the default: the evaluator blocks each turn
	// until conditions hold or the ceiling is reached, with NO per-turn user
	// prompt (existing D3 behavior).
	ProgressionAutonomous ProgressionMode = "autonomous"
	// ProgressionSemiAutonomous makes stop-goal emit a checkpoint-signal block
	// JSON each turn for orchestrator-side AskUserQuestion confirmation. The
	// hook itself never calls AskUserQuestion (REQ-GLE-014).
	ProgressionSemiAutonomous ProgressionMode = "semi-autonomous"
)

// DefaultMaxTurns is the default turn ceiling for a goal loop.
const DefaultMaxTurns = 30

// DefaultProgressionMode is the default progression mode for an armed goal.
const DefaultProgressionMode ProgressionMode = ProgressionAutonomous

// Goal is the per-session goal-state document persisted to
// .moai/state/goal/<session-id>.json. One file per session (never a single
// shared file — multi-session write-race avoidance per REQ-GLE-004).
type Goal struct {
	SessionID       string          `json:"session_id"`
	Goal            string          `json:"goal"`
	Conditions      []Condition     `json:"conditions"`
	Ceiling         Ceiling         `json:"ceiling"`
	TurnsUsed       int             `json:"turns_used"`
	Progress        []ProgressEntry `json:"progress"`
	ProgressionMode ProgressionMode `json:"progression_mode"`
	CreatedAt       string          `json:"created_at"`
	Status          Status          `json:"status"`
}

// NewGoal constructs a goal with sensible defaults: ceiling 30, autonomous
// progression, armed status, and the given session id + condition text.
func NewGoal(sessionID, goalText string, conditions []Condition) *Goal {
	return &Goal{
		SessionID:       sessionID,
		Goal:            goalText,
		Conditions:      conditions,
		Ceiling:         Ceiling{MaxTurns: DefaultMaxTurns},
		TurnsUsed:       0,
		Progress:        nil,
		ProgressionMode: DefaultProgressionMode,
		CreatedAt:       "",
		Status:          StatusArmed,
	}
}

// HasModelConditions reports whether the goal declares any Tier-2 (model)
// condition. Used by the evaluator to gate Tier-2.
func (g *Goal) HasModelConditions() bool {
	for _, c := range g.Conditions {
		if c.Type == ConditionModel {
			return true
		}
	}
	return false
}

// IsUnrecoverableStopFailure reports whether a StopFailure `error_type` names a
// condition a retry cannot clear, and therefore one that leaves an armed goal
// with nothing running to advance it.
//
// The predicate lives here rather than beside the hook handler because the
// question it answers is a goal-lifecycle one: not "was this turn an error"
// but "can the work this goal is waiting on still resume". Three types say no
// — a revoked or org-refused credential, and an exhausted balance. Everything
// else says yes, INCLUDING the unknown and empty cases: an unrecognized type is
// not evidence of unrecoverability, and treating it as such would destroy live
// state on a condition nobody has classified.
//
// The recoverable side is the load-bearing half. `rate_limit`, `overloaded`,
// and `server_error` resolve on their own, and the goal is exactly the state
// that must survive to see it; `max_output_tokens` is classified
// withheld-recoverable by runtime-recovery-doctrine.md §1, which obliges the
// agent to recover rather than treat the turn as failed. Disarming on any of
// these is the opposite failure from the one this predicate exists to prevent,
// and a quieter one — the goal is simply gone, with no idle turns to notice.
func IsUnrecoverableStopFailure(errorType string) bool {
	switch errorType {
	case "authentication_failed", "oauth_org_not_allowed", "billing_error":
		return true
	default:
		return false
	}
}
