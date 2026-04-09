// Package dashboard는 리서치 실험 대시보드의 터미널 렌더링을 담당한다.
// lipgloss를 사용하여 진행률 바, 실험 통계, 기준별 분석을 렌더링한다.
package dashboard

// DashboardData는 리서치 대시보드 렌더링에 필요한 전체 데이터를 보관한다.
type DashboardData struct {
	Target         string            // 리서치 대상 이름
	Baseline       float64           // 기준 점수 (0.0~1.0)
	CurrentScore   float64           // 현재 점수 (0.0~1.0)
	TargetScore    float64           // 목표 점수 (0.0~1.0)
	Experiments    int               // 완료된 실험 수
	MaxExperiments int               // 최대 실험 수
	KeepCount      int               // 유지된 실험 수
	DiscardCount   int               // 폐기된 실험 수
	PerCriterion   []CriterionStatus // 기준별 상태 목록
}

// CriterionStatus는 단일 평가 기준의 통과율을 나타낸다.
type CriterionStatus struct {
	Name     string  // 기준 이름
	PassRate float64 // 통과율 (0.0~1.0)
	Weight   string  // "MUST" 또는 빈 문자열
}
