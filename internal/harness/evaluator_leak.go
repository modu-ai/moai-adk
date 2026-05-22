// Package harness provides validation utilities for the GAN loop harness.
// REQ-HRN-002-017: prior-judgment leak detection for evaluator spawn prompts.
package harness

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrPriorJudgmentLeak is the sentinel error returned when a trace of a prior
// iteration's judgment is detected in the evaluator spawn prompt.
// Defined by design-constitution §11.4.1 and REQ-HRN-002-017.
var ErrPriorJudgmentLeak = errors.New("HRN_EVAL_PRIOR_JUDGMENT_LEAK: prior iteration judgment detected in evaluator spawn prompt (violates design-constitution §11.4.1)")

// forbiddenSubstrings is the list of forbidden substrings that indicate a trace
// of a prior iteration's judgment in the evaluator spawn prompt.
var forbiddenSubstrings = []string{
	"Score:",
	"Feedback:",
	"Verdict:",
}

// iterationPattern is the regular expression that detects numeric iteration references.
// Detects "Iteration 3", "iteration 2", etc.
var iterationPattern = regexp.MustCompile(`(?i)\bIteration\s+\d+`)

// priorEvaluatorPattern is the regular expression that detects mentions of a prior evaluator.
// Detects "previous evaluator", "prior evaluator", etc.
var priorEvaluatorPattern = regexp.MustCompile(`(?i)\b(previous|prior)\s+evaluator`)

// DetectPriorJudgmentLeak detects traces of a prior iteration's judgment
// (score, feedback, verdict, iteration-number references, etc.) in the evaluator spawn prompt.
// Returns ErrPriorJudgmentLeak when a trace is detected.
//
// @MX:WARN reason="prior-judgment leak detection per REQ-HRN-002-017"
// @MX:REASON: design-constitution §11.4.1 — prior iteration judgment rationales, scoring internals,
// or reflection traces MUST NOT appear in the evaluator's context window.
func DetectPriorJudgmentLeak(spawnPrompt string) error {
	// Detect forbidden substrings.
	for _, sub := range forbiddenSubstrings {
		if strings.Contains(spawnPrompt, sub) {
			return fmt.Errorf("%w: forbidden substring %q found in spawn prompt", ErrPriorJudgmentLeak, sub)
		}
	}

	// Detect numeric iteration-reference patterns.
	if loc := iterationPattern.FindStringIndex(spawnPrompt); loc != nil {
		matched := spawnPrompt[loc[0]:loc[1]]
		return fmt.Errorf("%w: forbidden iteration reference %q found in spawn prompt", ErrPriorJudgmentLeak, matched)
	}

	// Detect prior-evaluator mention patterns.
	if loc := priorEvaluatorPattern.FindStringIndex(spawnPrompt); loc != nil {
		matched := spawnPrompt[loc[0]:loc[1]]
		return fmt.Errorf("%w: forbidden prior-evaluator reference %q found in spawn prompt", ErrPriorJudgmentLeak, matched)
	}

	return nil
}
