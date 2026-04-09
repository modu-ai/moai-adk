package eval

import "testing"

// TestComputeResult ComputeResult 함수에 대한 테이블 기반 테스트.
func TestComputeResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		criteria     []EvalCriterion
		results      map[string]bool
		wantOverall  float64
		wantMustPass bool
	}{
		{
			name: "전부 통과 → Overall=1.0, MustPassOK=true",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": true, "b": true},
			wantOverall:  1.0,
			wantMustPass: true,
		},
		{
			name: "전부 실패 → Overall=0.0, MustPassOK=false",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": false, "b": false},
			wantOverall:  0.0,
			wantMustPass: false,
		},
		{
			name: "must_pass 통과, nice_to_have 실패 → MustPassOK=true, Overall < 1.0",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": true, "b": false},
			wantOverall:  0.5,
			wantMustPass: true,
		},
		{
			name: "must_pass 실패 → MustPassOK=false (overall 무관)",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
				{Name: "c", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": false, "b": true, "c": true},
			wantOverall:  2.0 / 3.0,
			wantMustPass: false,
		},
		{
			name:         "빈 결과 → Overall=0.0",
			criteria:     []EvalCriterion{},
			results:      map[string]bool{},
			wantOverall:  0.0,
			wantMustPass: true, // must_pass 기준이 없으면 true
		},
		{
			name: "must_pass만 있고 전부 통과",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: MustPass},
			},
			results:      map[string]bool{"a": true, "b": true},
			wantOverall:  1.0,
			wantMustPass: true,
		},
		{
			name: "results에 없는 기준은 실패 처리",
			criteria: []EvalCriterion{
				{Name: "a", Weight: MustPass},
				{Name: "b", Weight: NiceToHave},
			},
			results:      map[string]bool{"a": true},
			wantOverall:  0.5,
			wantMustPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ComputeResult(tt.criteria, tt.results)

			if result == nil {
				t.Fatal("ComputeResult() returned nil")
				return // staticcheck SA5011 방어
			}

			// float64 비교 (허용 오차 0.001)
			if diff := result.Overall - tt.wantOverall; diff > 0.001 || diff < -0.001 {
				t.Errorf("Overall = %f, want %f", result.Overall, tt.wantOverall)
			}

			if result.MustPassOK != tt.wantMustPass {
				t.Errorf("MustPassOK = %v, want %v", result.MustPassOK, tt.wantMustPass)
			}

			// PerCriterion 맵 크기 검증
			if len(result.PerCriterion) != len(tt.criteria) {
				t.Errorf("PerCriterion 크기 = %d, want %d", len(result.PerCriterion), len(tt.criteria))
			}

			// Timestamp가 제로 값이 아닌지 검증
			if result.Timestamp.IsZero() {
				t.Error("Timestamp이 제로 값")
			}
		})
	}
}
