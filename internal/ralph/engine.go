// Package ralph implements the decision engine for the Ralph Feedback Loop.
// It evaluates loop state and feedback to determine the next action:
// continue, converge, request human review, or abort.
package ralph

import (
	"context"
	"fmt"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/loop"
)

// RalphEngine implements loop.DecisionEngine using configurable heuristics.
type RalphEngine struct {
	cfg config.RalphConfig
}

// NewRalphEngine creates a new decision engine with the given configuration.
func NewRalphEngine(cfg config.RalphConfig) *RalphEngine {
	return &RalphEngine{cfg: cfg}
}

// Decide evaluates the current loop state and feedback to produce a Decision.
// Decision priority (highest to lowest):
//  1. Max iterations reached -> abort
//  2. Perfect quality gate -> converge
//  3. Stagnation detected (auto_converge) -> converge
//  4. Human review requested (human_review) -> request_review
//  5. Default -> continue
func (e *RalphEngine) Decide(_ context.Context, state *loop.LoopState, feedback *loop.Feedback) (*loop.Decision, error) {
	if state == nil {
		return nil, fmt.Errorf("ralph: cannot decide on nil state")
	}
	if feedback == nil {
		return nil, fmt.Errorf("ralph: cannot decide on nil feedback")
	}

	// 1. Max iterations check.
	if state.Iteration >= state.MaxIter {
		return &loop.Decision{
			Action:    loop.ActionAbort,
			Converged: false,
			Reason:    "max iterations reached",
		}, nil
	}

	// 2. Perfect success: all quality gates satisfied.
	if loop.MeetsQualityGate(feedback) {
		return &loop.Decision{
			Action:    loop.ActionConverge,
			NextPhase: loop.PhaseAnalyze,
			Converged: true,
			Reason:    "quality gate satisfied",
		}, nil
	}

	// 3. Stagnation detection (auto-converge enabled).
	if e.cfg.AutoConverge {
		prev := loop.FindPreviousReviewFeedback(state.Feedback, state.Iteration)
		if prev != nil && loop.IsStagnant(prev, feedback) {
			return &loop.Decision{
				Action:    loop.ActionConverge,
				Converged: true,
				Reason:    "no improvement detected (stagnant)",
			}, nil
		}
	}

	// 4. Human review breakpoint.
	if e.cfg.HumanReview && state.Phase == loop.PhaseReview {
		return &loop.Decision{
			Action:    loop.ActionRequestReview,
			NextPhase: loop.PhaseAnalyze,
			Converged: false,
			Reason:    "human review requested",
		}, nil
	}

	// 5. Default: continue to next iteration.
	return &loop.Decision{
		Action:    loop.ActionContinue,
		NextPhase: loop.PhaseAnalyze,
		Converged: false,
		Reason:    "continuing to next iteration",
	}, nil
}

// Compile-time interface compliance check.
var _ loop.DecisionEngine = (*RalphEngine)(nil)
