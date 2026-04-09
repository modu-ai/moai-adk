package safety

import "fmt"

// CanaryChecker는 제안된 변경이 회귀를 일으키지 않는지 검증한다.
type CanaryChecker struct{}

// NewCanaryChecker는 새로운 CanaryChecker를 생성한다.
func NewCanaryChecker() *CanaryChecker {
	return &CanaryChecker{}
}

// Check는 제안된 결과를 베이스라인과 비교한다.
// 제안 점수가 임계값 이상 하락하지 않으면 true를 반환한다.
// threshold는 허용 가능한 최대 점수 하락폭이다 (예: 0.10 = 10%).
func (c *CanaryChecker) Check(baselines []Baseline, proposed float64, threshold float64) (bool, error) {
	if threshold <= 0 {
		return false, fmt.Errorf("research/safety: 임계값은 양수여야 함: %f", threshold)
	}

	// 베이스라인이 없으면 비교 대상이 없으므로 통과
	if len(baselines) == 0 {
		return true, nil
	}

	// 각 베이스라인에 대해 회귀 여부 확인
	for _, b := range baselines {
		drop := b.Score - proposed
		if drop > threshold {
			return false, nil
		}
	}

	return true, nil
}
