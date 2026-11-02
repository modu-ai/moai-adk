package quality

import "time"

// OverallBudget returns the budget for one whole gate run and, when the
// configured override had to be corrected, a one-line warning.
//
// STUB (card t2020 RED): returns the historical fixed ten minutes.
func (g *QualityGate) OverallBudget() (time.Duration, string) {
	return 10 * time.Minute, ""
}
