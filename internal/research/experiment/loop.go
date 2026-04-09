package experiment

import (
	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// LoopConfig configures the execution conditions of the experiment loop.
type LoopConfig struct {
	MaxExperiments      int     // Maximum number of experiments
	TargetScore         float64 // Target score (loop ends after 3 consecutive hits)
	StagnationThreshold float64 // Threshold for stagnation detection (improvement below this value is considered stagnation)
	StagnationPatience  int     // Number of allowed stagnation occurrences before terminating
}

// Loop is the experiment loop state machine.
// After baseline is set, it records experiment results and determines whether to continue.
type Loop struct {
	config          LoopConfig
	state           ExperimentState
	baseline        *eval.EvalResult
	bestScore       float64
	history         []*Experiment
	stagnationCount int
	targetHitCount  int
}

// NewLoop creates a new experiment loop with the given configuration.
// The initial state is StateIdle.
func NewLoop(config LoopConfig) *Loop {
	return &Loop{
		config:  config,
		state:   StateIdle,
		history: make([]*Experiment, 0),
	}
}

// SetBaseline sets the baseline evaluation result and transitions state to StateBaseline.
func (l *Loop) SetBaseline(result *eval.EvalResult) {
	l.baseline = result
	l.bestScore = result.Overall
	l.state = StateBaseline
}

// ShouldContinue determines whether the experiment loop should continue.
// Returns false when:
//  1. The number of experiments reaches MaxExperiments
//  2. The target score is achieved 3 consecutive times
//  3. The stagnation count exceeds StagnationPatience
//  4. The state is StateComplete
func (l *Loop) ShouldContinue() bool {
	if l.state == StateComplete {
		return false
	}
	if len(l.history) >= l.config.MaxExperiments {
		l.state = StateComplete
		return false
	}
	if l.targetHitCount >= 3 {
		l.state = StateComplete
		return false
	}
	if l.stagnationCount >= l.config.StagnationPatience {
		l.state = StateComplete
		return false
	}
	return true
}

// RecordExperiment records an experiment result and returns the keep/discard decision.
// Decision logic:
//  1. Result is nil → DecisionDiscard
//  2. MustPassOK is false → DecisionDiscard
//  3. Overall is higher than bestScore → DecisionKeep (updates bestScore)
//  4. Otherwise → DecisionDiscard
//
// As a side effect, updates the stagnation counter and target hit counter.
func (l *Loop) RecordExperiment(exp *Experiment) Decision {
	l.history = append(l.history, exp)

	// Handle nil result
	if exp.Result == nil {
		exp.Decision = DecisionDiscard
		l.stagnationCount++
		l.targetHitCount = 0
		return DecisionDiscard
	}

	// Handle must_pass failure
	if !exp.Result.MustPassOK {
		exp.Decision = DecisionDiscard
		l.stagnationCount++
		l.targetHitCount = 0
		return DecisionDiscard
	}

	// Calculate improvement
	improvement := exp.Result.Overall - l.bestScore

	// Keep/discard decision
	var decision Decision
	if exp.Result.Overall > l.bestScore {
		decision = DecisionKeep
		l.bestScore = exp.Result.Overall
	} else {
		decision = DecisionDiscard
	}
	exp.Decision = decision

	// Stagnation tracking: increment stagnation counter if improvement is below threshold
	if improvement < l.config.StagnationThreshold {
		l.stagnationCount++
	} else {
		l.stagnationCount = 0
	}

	// Target hit tracking: increment consecutive counter if target score is reached
	if exp.Result.Overall >= l.config.TargetScore {
		l.targetHitCount++
	} else {
		l.targetHitCount = 0
	}

	return decision
}

// BestScore returns the highest score achieved so far.
func (l *Loop) BestScore() float64 {
	return l.bestScore
}

// ExperimentCount returns the total number of recorded experiments.
func (l *Loop) ExperimentCount() int {
	return len(l.history)
}

// State returns the current state of the experiment loop.
func (l *Loop) State() ExperimentState {
	return l.state
}
