package loop

import "github.com/modu-ai/moai-adk/internal/config"

// IsStagnantWithDiagnostics returns true if stagnation is detected considering
// both integer metrics and LSPDiagnostics count trends (REQ-LL-010).
//
// Stagnation is signaled when:
//   - Integer metrics are unchanged (TestsFailed, LintErrors, Coverage) AND
//     LSPDiagnostics count is the same or increasing (not decreasing).
//
// @MX:NOTE: [AUTO] Stagnation detector extended with diagnostic trend awareness (REQ-LL-010)
func IsStagnantWithDiagnostics(prev, curr *Feedback) bool {
	if prev == nil || curr == nil {
		return false
	}

	// Check integer metric stagnation.
	intStagnant := curr.TestsFailed == prev.TestsFailed &&
		curr.LintErrors == prev.LintErrors &&
		curr.Coverage == prev.Coverage

	if !intStagnant {
		return false
	}

	// REQ-LL-010: diagnostic count trending up is a stagnation signal.
	// Decreasing diagnostics = improvement = NOT stagnant.
	prevCount := len(prev.LSPDiagnostics)
	currCount := len(curr.LSPDiagnostics)
	if currCount < prevCount {
		// Diagnostic count decreased: improvement detected, not stagnant.
		return false
	}

	// Integer metrics unchanged and diagnostics are same or increasing: stagnant.
	return true
}

// IsImproved returns true if the current feedback shows improvement
// over the previous feedback in any metric: fewer test failures,
// fewer lint errors, or higher coverage.
func IsImproved(prev, curr *Feedback) bool {
	if prev == nil || curr == nil {
		return false
	}
	return curr.TestsFailed < prev.TestsFailed ||
		curr.LintErrors < prev.LintErrors ||
		curr.Coverage > prev.Coverage
}

// IsStagnant returns true if two consecutive feedbacks show no change
// in test failures, lint errors, and coverage.
func IsStagnant(prev, curr *Feedback) bool {
	if prev == nil || curr == nil {
		return false
	}
	return curr.TestsFailed == prev.TestsFailed &&
		curr.LintErrors == prev.LintErrors &&
		curr.Coverage == prev.Coverage
}

// MeetsQualityGate returns true if the feedback meets all quality gate
// criteria: zero test failures, zero lint errors, build success,
// and coverage at or above config.DefaultTestCoverageTarget (85%).
func MeetsQualityGate(fb *Feedback) bool {
	if fb == nil {
		return false
	}
	return fb.TestsFailed == 0 &&
		fb.LintErrors == 0 &&
		fb.BuildSuccess &&
		fb.Coverage >= float64(config.DefaultTestCoverageTarget)
}

// FindPreviousReviewFeedback searches the feedback history for the most
// recent review-phase feedback from an iteration before currentIteration.
func FindPreviousReviewFeedback(feedbacks []Feedback, currentIteration int) *Feedback {
	for i := len(feedbacks) - 1; i >= 0; i-- {
		fb := feedbacks[i]
		if fb.Phase == PhaseReview && fb.Iteration < currentIteration {
			return &fb
		}
	}
	return nil
}
