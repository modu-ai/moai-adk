package observe

import (
	"encoding/json"
	"testing"
	"time"
)

// TestObservationType_Constants verifies ObservationType constant values.
func TestObservationType_Constants(t *testing.T) {
	tests := []struct {
		name string
		got  ObservationType
		want string
	}{
		{"correction", ObsCorrection, "correction"},
		{"failure", ObsFailure, "failure"},
		{"success", ObsSuccess, "success"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("ObservationType = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestPatternClassification_Constants verifies PatternClassification constant values.
func TestPatternClassification_Constants(t *testing.T) {
	tests := []struct {
		name string
		got  PatternClassification
		want string
	}{
		{"observation", ClassObservation, "observation"},
		{"heuristic", ClassHeuristic, "heuristic"},
		{"rule", ClassRule, "rule"},
		{"high_confidence", ClassHighConfidence, "high_confidence"},
		{"anti_pattern", ClassAntiPattern, "anti_pattern"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("PatternClassification = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestObservation_JSONRoundTrip verifies JSON marshal/unmarshal round-trip for Observation.
func TestObservation_JSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	obs := &Observation{
		Type:      ObsCorrection,
		Agent:     "expert-backend",
		Target:    "error-handling",
		Detail:    "fmt.Errorf wrapping missing",
		Timestamp: now,
	}

	// Serialize
	data, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Deserialize
	var got Observation
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got.Type != obs.Type {
		t.Errorf("Type = %q, want %q", got.Type, obs.Type)
	}
	if got.Agent != obs.Agent {
		t.Errorf("Agent = %q, want %q", got.Agent, obs.Agent)
	}
	if got.Target != obs.Target {
		t.Errorf("Target = %q, want %q", got.Target, obs.Target)
	}
	if got.Detail != obs.Detail {
		t.Errorf("Detail = %q, want %q", got.Detail, obs.Detail)
	}
	if !got.Timestamp.Equal(obs.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, obs.Timestamp)
	}
}

// TestPattern_JSONRoundTrip verifies JSON marshal/unmarshal round-trip for Pattern.
func TestPattern_JSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	obs := &Observation{
		Type:      ObsFailure,
		Agent:     "expert-testing",
		Target:    "coverage",
		Detail:    "coverage below 85%",
		Timestamp: now,
	}
	pattern := &Pattern{
		Key:            "expert-testing:coverage",
		Classification: ClassHeuristic,
		Count:          3,
		Observations:   []*Observation{obs},
		FirstSeen:      now.Add(-time.Hour),
		LastSeen:       now,
	}

	data, err := json.Marshal(pattern)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var got Pattern
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if got.Key != pattern.Key {
		t.Errorf("Key = %q, want %q", got.Key, pattern.Key)
	}
	if got.Classification != pattern.Classification {
		t.Errorf("Classification = %q, want %q", got.Classification, pattern.Classification)
	}
	if got.Count != pattern.Count {
		t.Errorf("Count = %d, want %d", got.Count, pattern.Count)
	}
	if len(got.Observations) != 1 {
		t.Fatalf("len(Observations) = %d, want 1", len(got.Observations))
	}
	if !got.FirstSeen.Equal(pattern.FirstSeen) {
		t.Errorf("FirstSeen = %v, want %v", got.FirstSeen, pattern.FirstSeen)
	}
	if !got.LastSeen.Equal(pattern.LastSeen) {
		t.Errorf("LastSeen = %v, want %v", got.LastSeen, pattern.LastSeen)
	}
}
