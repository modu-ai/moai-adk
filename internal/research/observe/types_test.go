package observe

import (
	"encoding/json"
	"testing"
	"time"
)

// ObservationType 상수 값 검증
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

// PatternClassification 상수 값 검증
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

// Observation JSON 라운드트립 검증
func TestObservation_JSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	obs := &Observation{
		Type:      ObsCorrection,
		Agent:     "expert-backend",
		Target:    "error-handling",
		Detail:    "fmt.Errorf 래핑 누락",
		Timestamp: now,
	}

	// 직렬화
	data, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("json.Marshal 실패: %v", err)
	}

	// 역직렬화
	var got Observation
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal 실패: %v", err)
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

// Pattern JSON 라운드트립 검증
func TestPattern_JSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	obs := &Observation{
		Type:      ObsFailure,
		Agent:     "expert-testing",
		Target:    "coverage",
		Detail:    "커버리지 85% 미달",
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
		t.Fatalf("json.Marshal 실패: %v", err)
	}

	var got Pattern
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal 실패: %v", err)
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
