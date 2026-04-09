package safety

import "time"

// Baseline holds the evaluation reference value for canary checks.
type Baseline struct {
	Target    string    `json:"target"`
	Score     float64   `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

// RateLimitConfig defines the rate limit parameters.
type RateLimitConfig struct {
	MaxExperimentsPerSession int `yaml:"max_experiments_per_session"`
	MaxAcceptedPerSession    int `yaml:"max_accepted_per_session"`
	MaxAutoResearchPerWeek   int `yaml:"max_auto_research_per_week"`
}

// ActionRecord is a single action record for rate limit tracking.
type ActionRecord struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
