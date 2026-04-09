// Package observe provides passive observation collection and pattern detection.
// It records correction, failure, and success events during agent execution
// and automatically classifies repeating patterns.
package observe

import "time"

// ObservationType represents the type of observation.
type ObservationType string

const (
	// ObsCorrection represents a user correction observation.
	ObsCorrection ObservationType = "correction"
	// ObsFailure represents an agent failure observation.
	ObsFailure ObservationType = "failure"
	// ObsSuccess represents an agent success observation.
	ObsSuccess ObservationType = "success"
)

// Observation represents a single observation record.
type Observation struct {
	Type      ObservationType `json:"type"`
	Agent     string          `json:"agent"`
	Target    string          `json:"target"`
	Detail    string          `json:"detail"`
	Timestamp time.Time       `json:"timestamp"`
}

// PatternClassification represents the confidence classification of a pattern.
type PatternClassification string

const (
	// ClassObservation is the simple observation stage (insufficient repetitions).
	ClassObservation PatternClassification = "observation"
	// ClassHeuristic is the heuristic stage (confirmed above a certain repetition count).
	ClassHeuristic PatternClassification = "heuristic"
	// ClassRule is the rule stage (high repetition count).
	ClassRule PatternClassification = "rule"
	// ClassHighConfidence is the high-confidence stage (very high repetition count).
	ClassHighConfidence PatternClassification = "high_confidence"
	// ClassAntiPattern is an anti-pattern (subject to immediate flagging).
	ClassAntiPattern PatternClassification = "anti_pattern"
)

// Pattern represents a pattern detected from grouped observations.
type Pattern struct {
	Key            string                `json:"key"`
	Classification PatternClassification `json:"classification"`
	Count          int                   `json:"count"`
	Observations   []*Observation        `json:"observations"`
	FirstSeen      time.Time             `json:"first_seen"`
	LastSeen       time.Time             `json:"last_seen"`
}
