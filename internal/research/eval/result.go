package eval

import "time"

// ComputeResult는 개별 기준 결과를 집계하여 EvalResult를 생성한다.
// Overall = 통과 기준 수 / 전체 기준 수 (기준이 없으면 0.0).
// MustPassOK = 모든 must_pass 기준이 통과했는지 여부 (must_pass가 없으면 true).
// results 맵에 없는 기준은 실패로 처리한다.
func ComputeResult(criteria []EvalCriterion, results map[string]bool) *EvalResult {
	total := len(criteria)
	if total == 0 {
		return &EvalResult{
			Overall:      0.0,
			PerCriterion: make(map[string]CriterionResult),
			MustPassOK:   true,
			Timestamp:    time.Now(),
		}
	}

	passCount := 0
	mustPassOK := true
	perCriterion := make(map[string]CriterionResult, total)

	for _, c := range criteria {
		passed := results[c.Name] // 맵에 없으면 false (실패 처리)

		perCriterion[c.Name] = CriterionResult{
			Name:   c.Name,
			Passed: passed,
			Weight: c.Weight,
		}

		if passed {
			passCount++
		}

		if c.Weight == MustPass && !passed {
			mustPassOK = false
		}
	}

	return &EvalResult{
		Overall:      float64(passCount) / float64(total),
		PerCriterion: perCriterion,
		MustPassOK:   mustPassOK,
		Timestamp:    time.Now(),
	}
}
