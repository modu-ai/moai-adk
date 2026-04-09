package eval

import "time"

// ComputeResult aggregates individual criterion results to produce an EvalResult.
// Overall = number of passing criteria / total criteria (0.0 if no criteria).
// MustPassOK = whether all must_pass criteria passed (true if no must_pass criteria).
// Criteria absent from the results map are treated as failed.
func ComputeResult(criteria []EvalCriterion, results map[string]bool) *EvalResult {
	total := len(criteria)
	if total == 0 {
		return &EvalResult{
			Overall:      0.0,
			PerCriterion: make(map[string]CriterionResult),
			MustPassOK:   true,
			Timestamp:    time.Now(),
		}
	}

	passCount := 0
	mustPassOK := true
	perCriterion := make(map[string]CriterionResult, total)

	for _, c := range criteria {
		passed := results[c.Name] // false if absent from map (treated as failed)

		perCriterion[c.Name] = CriterionResult{
			Name:   c.Name,
			Passed: passed,
			Weight: c.Weight,
		}

		if passed {
			passCount++
		}

		if c.Weight == MustPass && !passed {
			mustPassOK = false
		}
	}

	return &EvalResult{
		Overall:      float64(passCount) / float64(total),
		PerCriterion: perCriterion,
		MustPassOK:   mustPassOK,
		Timestamp:    time.Now(),
	}
}
