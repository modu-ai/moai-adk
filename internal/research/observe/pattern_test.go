package observe

import (
	"testing"
	"time"
)

// 기본 임계값 검증
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

// 빈 관찰 목록 → 빈 패턴 목록
func TestPatternDetector_Detect_Empty(t *testing.T) {
	d := NewPatternDetector(DefaultThresholds())
	patterns := d.Detect(nil)
	if len(patterns) != 0 {
		t.Errorf("len = %d, want 0", len(patterns))
	}
}

// 분류 임계값에 따른 패턴 분류 검증
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
		{"1개 관찰 → observation", 1, ClassObservation},
		{"2개 관찰 → observation", 2, ClassObservation},
		{"3개 관찰 → heuristic", 3, ClassHeuristic},
		{"4개 관찰 → heuristic", 4, ClassHeuristic},
		{"5개 관찰 → rule", 5, ClassRule},
		{"9개 관찰 → rule", 9, ClassRule},
		{"10개 관찰 → high_confidence", 10, ClassHighConfidence},
		{"15개 관찰 → high_confidence", 15, ClassHighConfidence},
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
					Detail:    "상세",
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

// 서로 다른 키 → 별도 패턴 생성
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

// count 내림차순 정렬 확인
func TestPatternDetector_Detect_SortedByCountDesc(t *testing.T) {
	base := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	var observations []*Observation

	// agent-a:target → 1개
	observations = append(observations, &Observation{
		Agent: "agent-a", Target: "target", Timestamp: base,
	})

	// agent-b:target → 3개
	for i := 0; i < 3; i++ {
		observations = append(observations, &Observation{
			Agent: "agent-b", Target: "target", Timestamp: base.Add(time.Duration(i) * time.Minute),
		})
	}

	// agent-c:target → 2개
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
	// 내림차순: 3, 2, 1
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

// FirstSeen/LastSeen 정확성 검증
func TestPatternDetector_Detect_FirstSeenLastSeen(t *testing.T) {
	t1 := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 4, 9, 14, 0, 0, 0, time.UTC)

	// 의도적으로 순서를 섞어서 입력
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

// 패턴 키 형식 검증 (Agent:Target)
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
