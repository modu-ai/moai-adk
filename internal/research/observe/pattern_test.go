package observe

import (
	"testing"
	"time"
)

// TestDefaultThresholds verifies default threshold values.
func TestDefaultThresholds(t *testing.T) {
	d := DefaultThresholds()
	if d.Heuristic != 3 {
		t.Errorf("Heuristic = %d, want 3", d.Heuristic)
	}
	if d.Rule != 5 {
		t.Errorf("Rule = %d, want 5", d.Rule)
	}
	if d.HighConfidence != 10 {
		t.Errorf("HighConfidence = %d, want 10", d.HighConfidence)
	}
}

// TestPatternDetector_Detect_Empty verifies that an empty observation list produces an empty pattern list.
func TestPatternDetector_Detect_Empty(t *testing.T) {
	d := NewPatternDetector(DefaultThresholds())
	patterns := d.Detect(nil)
	if len(patterns) != 0 {
		t.Errorf("len = %d, want 0", len(patterns))
	}
}

// TestPatternDetector_Detect_Classification verifies pattern classification based on count thresholds.
func TestPatternDetector_Detect_Classification(t *testing.T) {
	thresholds := PatternThresholds{
		Heuristic:      3,
		Rule:           5,
		HighConfidence: 10,
	}

	tests := []struct {
		name  string
		count int
		want  PatternClassification
	}{
		{"1 observation → observation", 1, ClassObservation},
		{"2 observations → observation", 2, ClassObservation},
		{"3 observations → heuristic", 3, ClassHeuristic},
		{"4 observations → heuristic", 4, ClassHeuristic},
		{"5 observations → rule", 5, ClassRule},
		{"9 observations → rule", 9, ClassRule},
		{"10 observations → high_confidence", 10, ClassHighConfidence},
		{"15 observations → high_confidence", 15, ClassHighConfidence},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
			var observations []*Observation
			for i := 0; i < tt.count; i++ {
				observations = append(observations, &Observation{
					Type:      ObsCorrection,
					Agent:     "agent-a",
					Target:    "target-x",
					Detail:    "detail",
					Timestamp: base.Add(time.Duration(i) * time.Minute),
				})
			}

			d := NewPatternDetector(thresholds)
			patterns := d.Detect(observations)

			if len(patterns) != 1 {
				t.Fatalf("len = %d, want 1", len(patterns))
			}
			if patterns[0].Classification != tt.want {
				t.Errorf("Classification = %q, want %q", patterns[0].Classification, tt.want)
			}
			if patterns[0].Count != tt.count {
				t.Errorf("Count = %d, want %d", patterns[0].Count, tt.count)
			}
		})
	}
}

// TestPatternDetector_Detect_MixedKeys verifies that different keys produce separate patterns.
func TestPatternDetector_Detect_MixedKeys(t *testing.T) {
	base := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	observations := []*Observation{
		{Agent: "agent-a", Target: "target-1", Timestamp: base},
		{Agent: "agent-a", Target: "target-1", Timestamp: base.Add(time.Minute)},
		{Agent: "agent-b", Target: "target-2", Timestamp: base.Add(2 * time.Minute)},
	}

	d := NewPatternDetector(DefaultThresholds())
	patterns := d.Detect(observations)

	if len(patterns) != 2 {
		t.Fatalf("len = %d, want 2", len(patterns))
	}
}

// TestPatternDetector_Detect_SortedByCountDesc verifies that patterns are sorted in descending order by count.
func TestPatternDetector_Detect_SortedByCountDesc(t *testing.T) {
	base := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	var observations []*Observation

	// agent-a:target → 1 observation
	observations = append(observations, &Observation{
		Agent: "agent-a", Target: "target", Timestamp: base,
	})

	// agent-b:target → 3 observations
	for i := 0; i < 3; i++ {
		observations = append(observations, &Observation{
			Agent: "agent-b", Target: "target", Timestamp: base.Add(time.Duration(i) * time.Minute),
		})
	}

	// agent-c:target → 2 observations
	for i := 0; i < 2; i++ {
		observations = append(observations, &Observation{
			Agent: "agent-c", Target: "target", Timestamp: base.Add(time.Duration(i) * time.Minute),
		})
	}

	d := NewPatternDetector(DefaultThresholds())
	patterns := d.Detect(observations)

	if len(patterns) != 3 {
		t.Fatalf("len = %d, want 3", len(patterns))
	}
	// Descending order: 3, 2, 1
	if patterns[0].Count != 3 {
		t.Errorf("patterns[0].Count = %d, want 3", patterns[0].Count)
	}
	if patterns[1].Count != 2 {
		t.Errorf("patterns[1].Count = %d, want 2", patterns[1].Count)
	}
	if patterns[2].Count != 1 {
		t.Errorf("patterns[2].Count = %d, want 1", patterns[2].Count)
	}
}

// TestPatternDetector_Detect_FirstSeenLastSeen verifies that FirstSeen and LastSeen are set correctly.
func TestPatternDetector_Detect_FirstSeenLastSeen(t *testing.T) {
	t1 := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 4, 9, 14, 0, 0, 0, time.UTC)

	// Intentionally provide observations out of order
	observations := []*Observation{
		{Agent: "agent", Target: "target", Timestamp: t2},
		{Agent: "agent", Target: "target", Timestamp: t1},
		{Agent: "agent", Target: "target", Timestamp: t3},
	}

	d := NewPatternDetector(DefaultThresholds())
	patterns := d.Detect(observations)

	if len(patterns) != 1 {
		t.Fatalf("len = %d, want 1", len(patterns))
	}
	if !patterns[0].FirstSeen.Equal(t1) {
		t.Errorf("FirstSeen = %v, want %v", patterns[0].FirstSeen, t1)
	}
	if !patterns[0].LastSeen.Equal(t3) {
		t.Errorf("LastSeen = %v, want %v", patterns[0].LastSeen, t3)
	}
}

// TestPatternDetector_Detect_KeyFormat verifies the pattern key format (Agent:Target).
func TestPatternDetector_Detect_KeyFormat(t *testing.T) {
	observations := []*Observation{
		{
			Agent:     "expert-backend",
			Target:    "error-handling",
			Timestamp: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		},
	}

	d := NewPatternDetector(DefaultThresholds())
	patterns := d.Detect(observations)

	if len(patterns) != 1 {
		t.Fatalf("len = %d, want 1", len(patterns))
	}
	want := "expert-backend:error-handling"
	if patterns[0].Key != want {
		t.Errorf("Key = %q, want %q", patterns[0].Key, want)
	}
}
