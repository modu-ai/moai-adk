// Package ralph implements the decision engine for the Ralph Feedback Loop.
// It evaluates loop state and feedback to determine the next action:
// continue, converge, request human review, or abort.
package ralph

import (
	"context"
	"fmt"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/loop"
	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
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

// ErrorLevel represents the severity classification for feedback errors.
// Higher levels require more human involvement.
type ErrorLevel int

const (
	// ErrorLevelSkip is a sentinel for diagnostics that should be ignored
	// (Severity=Information or Severity=Hint per REQ-LL-005).
	ErrorLevelSkip ErrorLevel = 0

	// ErrorLevelAutoFix indicates errors that can be fixed automatically
	// (e.g., formatting, simple lint issues, unused imports).
	ErrorLevelAutoFix ErrorLevel = 1

	// ErrorLevelLogOnly indicates errors that should be logged but
	// don't block progress (e.g., minor warnings, style suggestions).
	ErrorLevelLogOnly ErrorLevel = 2

	// ErrorLevelApproval indicates errors that need human review before
	// proceeding (e.g., security warnings, breaking API changes).
	ErrorLevelApproval ErrorLevel = 3

	// ErrorLevelManual indicates errors that require manual intervention
	// (e.g., build failures, test infrastructure issues, env problems).
	ErrorLevelManual ErrorLevel = 4

	// ErrorLevelBlocker indicates compiler errors that must be resolved
	// before any other action (REQ-LL-005: Severity=Error + Source=compiler).
	// @MX:ANCHOR: [AUTO] ErrorLevelBlocker — highest-priority classification for compiler errors
	// @MX:REASON: fan_in >= 3 — ClassifyFeedback, ClassifyFeedbackWithConfig, and tests all reference this constant; classification logic must remain consistent
	ErrorLevelBlocker ErrorLevel = 5
)

// ClassifiedError pairs an error description with its severity level.
type ClassifiedError struct {
	Level       ErrorLevel `json:"level"`
	Description string     `json:"description"`
	Count       int        `json:"count"`
}

// ClassifyFeedback analyzes feedback metrics and classifies the issues
// into error levels for the decision engine.
// When fb.LSPDiagnostics is non-empty, source-aware classification (REQ-LL-004/005)
// is applied. When empty, integer-based classification is used as fallback (REQ-LL-008).
//
// @MX:ANCHOR: [AUTO] ClassifyFeedback — central classification entry point
// @MX:REASON: fan_in >= 3 — LoopController, tests, RalphEngine, and external callers all invoke ClassifyFeedback; classification rules must remain stable
func ClassifyFeedback(fb *loop.Feedback) []ClassifiedError {
	return ClassifyFeedbackWithConfig(fb, config.RalphConfig{
		LintAsInstruction: false,
		WarnAsInstruction: false,
	})
}

// ClassifyFeedbackWithConfig analyzes feedback metrics and classifies issues
// using the provided ralph configuration flags (REQ-LL-006, REQ-LL-007).
// When LintAsInstruction or WarnAsInstruction is true, warning-severity
// diagnostics are treated as instruction-level input (no gate block).
//
// @MX:NOTE: [AUTO] LintAsInstruction/WarnAsInstruction influence classification severity
func ClassifyFeedbackWithConfig(fb *loop.Feedback, cfg config.RalphConfig) []ClassifiedError {
	if fb == nil {
		return nil
	}

	var classified []ClassifiedError

	// REQ-LL-004/005: If LSPDiagnostics is populated, use source-aware classification.
	// REQ-LL-008: If empty, fall through to integer-based classification (backwards compat).
	if len(fb.LSPDiagnostics) > 0 {
		classified = append(classified, classifyLSPDiagnostics(fb.LSPDiagnostics, cfg)...)
		return classified
	}

	// Backwards compatibility path (REQ-LL-008): integer-based classification.
	// Lint errors are typically auto-fixable (Level 1).
	if fb.LintErrors > 0 {
		classified = append(classified, ClassifiedError{
			Level:       ErrorLevelAutoFix,
			Description: "lint errors detected",
			Count:       fb.LintErrors,
		})
	}

	// Test failures need investigation (Level 2 for few, Level 3 for many).
	if fb.TestsFailed > 0 {
		level := ErrorLevelLogOnly
		if fb.TestsFailed > 5 {
			level = ErrorLevelApproval
		}
		classified = append(classified, ClassifiedError{
			Level:       level,
			Description: "test failures",
			Count:       fb.TestsFailed,
		})
	}

	// Build failure is a manual-level issue (Level 4).
	if !fb.BuildSuccess {
		classified = append(classified, ClassifiedError{
			Level:       ErrorLevelManual,
			Description: "build failure",
			Count:       1,
		})
	}

	// GOPLS-BRIDGE-001: legacy gopls.Diagnostic classification (backwards compat).
	// Kept for callers that still populate fb.Diagnostics directly.
	classified = append(classified, classifyDiagnostics(fb.Diagnostics)...)

	return classified
}

// classifyLSPDiagnostics applies source-aware classification rules per REQ-LL-005:
//
//	Rule 1: Severity=Error  + Source=compiler         → ErrorLevelBlocker
//	Rule 2: Severity=Error  + Source=staticcheck SA*   → ErrorLevelApproval
//	Rule 3: Severity=Warning + Source=staticcheck      → ErrorLevelAutoFix
//	Rule 4: Severity=Information                       → ErrorLevelSkip (omitted)
//	Rule 5: Severity=Hint                              → ErrorLevelSkip (omitted)
//
// When cfg.LintAsInstruction or cfg.WarnAsInstruction is true, warning-severity
// diagnostics are demoted to instruction-level (REQ-LL-006, REQ-LL-007).
//
// @MX:NOTE: [AUTO] 5-rule severity+source classification — rule order matters; compiler errors rank highest
func classifyLSPDiagnostics(diags []lsp.Diagnostic, cfg config.RalphConfig) []ClassifiedError {
	if len(diags) == 0 {
		return nil
	}

	var blockerCount, approvalCount, autoFixCount, logOnlyCount int

	for _, d := range diags {
		switch d.Severity {
		case lsp.SeverityError:
			if d.Source == "compiler" {
				// Rule 1: compiler errors are the highest priority blocker.
				blockerCount++
			} else if d.Source == "staticcheck" && strings.HasPrefix(d.Code, "SA") {
				// Rule 2: staticcheck SA-class errors need approval.
				approvalCount++
			} else {
				// Other error-level diagnostics need manual intervention.
				approvalCount++
			}
		case lsp.SeverityWarning:
			// REQ-LL-006: LintAsInstruction downgrades warnings to auto-fix instruction.
			// REQ-LL-007: WarnAsInstruction also downgrades warnings.
			if cfg.LintAsInstruction || cfg.WarnAsInstruction {
				// Warning treated as instruction: auto-fix level (no gate block).
				autoFixCount++
			} else if d.Source == "staticcheck" {
				// Rule 3: staticcheck warnings are auto-fixable.
				autoFixCount++
			} else {
				// Other warnings are log-only.
				logOnlyCount++
			}
		case lsp.SeverityInfo, lsp.SeverityHint:
			// Rules 4 & 5: skip information and hints entirely.
			continue
		}
	}

	var result []ClassifiedError
	if blockerCount > 0 {
		result = append(result, ClassifiedError{
			Level:       ErrorLevelBlocker,
			Description: "compiler errors (LSP blocker)",
			Count:       blockerCount,
		})
	}
	if approvalCount > 0 {
		result = append(result, ClassifiedError{
			Level:       ErrorLevelApproval,
			Description: "staticcheck errors requiring approval (LSP)",
			Count:       approvalCount,
		})
	}
	if autoFixCount > 0 {
		result = append(result, ClassifiedError{
			Level:       ErrorLevelAutoFix,
			Description: "staticcheck warnings (LSP auto-fix)",
			Count:       autoFixCount,
		})
	}
	if logOnlyCount > 0 {
		result = append(result, ClassifiedError{
			Level:       ErrorLevelLogOnly,
			Description: "diagnostic warnings (LSP)",
			Count:       logOnlyCount,
		})
	}
	return result
}

// classifyDiagnostics converts legacy gopls.Diagnostic entries into ClassifiedError.
// Error severity maps to ErrorLevelManual, Warning to ErrorLevelLogOnly.
// This function is kept for backwards compatibility with callers that populate
// fb.Diagnostics directly (pre-REQ-LL-001 callers).
func classifyDiagnostics(diags []gopls.Diagnostic) []ClassifiedError {
	if len(diags) == 0 {
		return nil
	}

	var errorCount, warningCount int
	for _, d := range diags {
		switch d.Severity {
		case gopls.SeverityError:
			errorCount++
		case gopls.SeverityWarning:
			warningCount++
		}
	}

	result := make([]ClassifiedError, 0, 2)
	if errorCount > 0 {
		result = append(result, ClassifiedError{
			Level:       ErrorLevelManual,
			Description: "diagnostic errors (LSP)",
			Count:       errorCount,
		})
	}
	if warningCount > 0 {
		result = append(result, ClassifiedError{
			Level:       ErrorLevelLogOnly,
			Description: "diagnostic warnings (LSP)",
			Count:       warningCount,
		})
	}
	return result
}

// MaxErrorLevel returns the highest error level from classified errors.
// Returns 0 if no errors are present.
func MaxErrorLevel(errors []ClassifiedError) ErrorLevel {
	var maxLevel ErrorLevel
	for _, e := range errors {
		if e.Level > maxLevel {
			maxLevel = e.Level
		}
	}
	return maxLevel
}

// Compile-time interface compliance check.
var _ loop.DecisionEngine = (*RalphEngine)(nil)
