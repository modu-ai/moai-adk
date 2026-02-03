package loop

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
// and coverage at or above DefaultCoverageTarget (85%).
func MeetsQualityGate(fb *Feedback) bool {
	if fb == nil {
		return false
	}
	return fb.TestsFailed == 0 &&
		fb.LintErrors == 0 &&
		fb.BuildSuccess &&
		fb.Coverage >= DefaultCoverageTarget
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
