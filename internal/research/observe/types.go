// Package observe는 수동적 관찰 수집 및 패턴 탐지를 제공한다.
// 에이전트 실행 중 교정, 실패, 성공 이벤트를 기록하고
// 반복되는 패턴을 자동으로 분류한다.
package observe

import "time"

// ObservationType은 관찰 유형을 나타낸다.
type ObservationType string

const (
	// ObsCorrection은 사용자 교정 관찰을 나타낸다.
	ObsCorrection ObservationType = "correction"
	// ObsFailure는 에이전트 실패 관찰을 나타낸다.
	ObsFailure ObservationType = "failure"
	// ObsSuccess는 에이전트 성공 관찰을 나타낸다.
	ObsSuccess ObservationType = "success"
)

// Observation은 단일 관찰 레코드를 나타낸다.
type Observation struct {
	Type      ObservationType `json:"type"`
	Agent     string          `json:"agent"`
	Target    string          `json:"target"`
	Detail    string          `json:"detail"`
	Timestamp time.Time       `json:"timestamp"`
}

// PatternClassification은 패턴의 신뢰도 분류를 나타낸다.
type PatternClassification string

const (
	// ClassObservation은 단순 관찰 단계 (반복 횟수 부족).
	ClassObservation PatternClassification = "observation"
	// ClassHeuristic은 휴리스틱 단계 (일정 횟수 이상 반복 확인).
	ClassHeuristic PatternClassification = "heuristic"
	// ClassRule은 규칙 단계 (높은 반복 횟수).
	ClassRule PatternClassification = "rule"
	// ClassHighConfidence는 높은 신뢰도 단계 (매우 높은 반복 횟수).
	ClassHighConfidence PatternClassification = "high_confidence"
	// ClassAntiPattern은 안티패턴 (즉시 플래그 처리 대상).
	ClassAntiPattern PatternClassification = "anti_pattern"
)

// Pattern은 그룹화된 관찰들에서 탐지된 패턴을 나타낸다.
type Pattern struct {
	Key            string                `json:"key"`
	Classification PatternClassification `json:"classification"`
	Count          int                   `json:"count"`
	Observations   []*Observation        `json:"observations"`
	FirstSeen      time.Time             `json:"first_seen"`
	LastSeen       time.Time             `json:"last_seen"`
}
