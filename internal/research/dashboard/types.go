// Package dashboard handles terminal rendering for the research experiment dashboard.
// It uses lipgloss to render progress bars, experiment statistics, and per-criterion breakdowns.
package dashboard

// DashboardData holds all data required to render the research dashboard.
type DashboardData struct {
	Target         string            // Research target name
	Baseline       float64           // Baseline score (0.0~1.0)
	CurrentScore   float64           // Current score (0.0~1.0)
	TargetScore    float64           // Target score (0.0~1.0)
	Experiments    int               // Number of completed experiments
	MaxExperiments int               // Maximum number of experiments
	KeepCount      int               // Number of kept experiments
	DiscardCount   int               // Number of discarded experiments
	PerCriterion   []CriterionStatus // Per-criterion status list
}

// CriterionStatus represents the pass rate of a single evaluation criterion.
type CriterionStatus struct {
	Name     string  // Criterion name
	PassRate float64 // Pass rate (0.0~1.0)
	Weight   string  // "MUST" or empty string
}
