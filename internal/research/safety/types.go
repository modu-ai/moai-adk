package safety

import "time"

// Baseline은 카나리 체크를 위한 평가 기준값을 나타낸다.
type Baseline struct {
	Target    string    `json:"target"`
	Score     float64   `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

// RateLimitConfig는 속도 제한 파라미터를 정의한다.
type RateLimitConfig struct {
	MaxExperimentsPerSession int `yaml:"max_experiments_per_session"`
	MaxAcceptedPerSession    int `yaml:"max_accepted_per_session"`
	MaxAutoResearchPerWeek   int `yaml:"max_auto_research_per_week"`
}

// ActionRecord는 속도 제한 추적을 위한 단일 액션 기록이다.
type ActionRecord struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
