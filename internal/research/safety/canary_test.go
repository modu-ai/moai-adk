package safety

import (
	"testing"
	"time"
)

// TestCanaryChecker_Check는 카나리 회귀 검사 로직을 검증한다.
func TestCanaryChecker_Check(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		baselines []Baseline
		proposed  float64
		threshold float64
		want      bool
		wantErr   bool
	}{
		{
			name:      "베이스라인 없음 → true (비교 대상 없음)",
			baselines: nil,
			proposed:  0.50,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "빈 베이스라인 슬라이스 → true",
			baselines: []Baseline{},
			proposed:  0.50,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name: "임계값 이내: 0.85 → 0.80, threshold=0.10 → true",
			baselines: []Baseline{
				{Target: "test", Score: 0.85, Timestamp: now},
			},
			proposed:  0.80,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name: "임계값 초과: 0.85 → 0.70, threshold=0.10 → false",
			baselines: []Baseline{
				{Target: "test", Score: 0.85, Timestamp: now},
			},
			proposed:  0.70,
			threshold: 0.10,
			want:      false,
			wantErr:   false,
		},
		{
			name: "정확히 임계값과 동일: 0.85 → 0.75, threshold=0.10 → true (경계값)",
			baselines: []Baseline{
				{Target: "test", Score: 0.85, Timestamp: now},
			},
			proposed:  0.75,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name: "여러 베이스라인, 하나가 회귀 → false",
			baselines: []Baseline{
				{Target: "a", Score: 0.80, Timestamp: now},
				{Target: "b", Score: 0.90, Timestamp: now},
			},
			proposed:  0.75,
			threshold: 0.10,
			want:      false, // b: 0.90 - 0.75 = 0.15 > 0.10
			wantErr:   false,
		},
		{
			name: "여러 베이스라인, 모두 임계값 이내 → true",
			baselines: []Baseline{
				{Target: "a", Score: 0.80, Timestamp: now},
				{Target: "b", Score: 0.82, Timestamp: now},
			},
			proposed:  0.75,
			threshold: 0.10,
			want:      true, // a: 0.05, b: 0.07 — 모두 <= 0.10
			wantErr:   false,
		},
		{
			name: "제안 점수가 베이스라인보다 높음 → true (향상)",
			baselines: []Baseline{
				{Target: "test", Score: 0.70, Timestamp: now},
			},
			proposed:  0.85,
			threshold: 0.10,
			want:      true,
			wantErr:   false,
		},
		{
			name:      "임계값 0 → 에러",
			baselines: []Baseline{{Target: "t", Score: 0.80, Timestamp: now}},
			proposed:  0.75,
			threshold: 0.0,
			want:      false,
			wantErr:   true,
		},
		{
			name:      "음수 임계값 → 에러",
			baselines: []Baseline{{Target: "t", Score: 0.80, Timestamp: now}},
			proposed:  0.75,
			threshold: -0.05,
			want:      false,
			wantErr:   true,
		},
	}

	checker := NewCanaryChecker()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.Check(tt.baselines, tt.proposed, tt.threshold)
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}
