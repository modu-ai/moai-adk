package safety

import "fmt"

// CanaryChecker verifies that a proposed change does not cause regressions.
type CanaryChecker struct{}

// NewCanaryChecker creates a new CanaryChecker.
func NewCanaryChecker() *CanaryChecker {
	return &CanaryChecker{}
}

// Check compares the proposed result against baselines.
// Returns true if the proposed score does not drop by more than the threshold.
// threshold is the maximum allowable score drop (e.g., 0.10 = 10%).
func (c *CanaryChecker) Check(baselines []Baseline, proposed float64, threshold float64) (bool, error) {
	if threshold <= 0 {
		return false, fmt.Errorf("research/safety: threshold must be positive: %f", threshold)
	}

	// No baselines means nothing to compare against, so pass
	if len(baselines) == 0 {
		return true, nil
	}

	// Check for regression against each baseline
	for _, b := range baselines {
		drop := b.Score - proposed
		if drop > threshold {
			return false, nil
		}
	}

	return true, nil
}
