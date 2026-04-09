// Package experiment는 Self-Research System의 실험 루프 상태 머신을 구현한다.
// 베이스라인 측정, 변이 적용, 평가, 점수 산정의 전체 실험 사이클을 관리한다.
package experiment

import (
	"time"

	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// ExperimentState는 실험 루프의 현재 상태를 나타낸다.
type ExperimentState string

const (
	// StateIdle은 실험 루프가 초기화되었지만 아직 시작되지 않은 상태이다.
	StateIdle ExperimentState = "idle"
	// StateBaseline은 베이스라인이 설정된 상태이다.
	StateBaseline ExperimentState = "baseline"
	// StateMutating은 변이를 적용 중인 상태이다.
	StateMutating ExperimentState = "mutating"
	// StateEvaluating은 변이 결과를 평가 중인 상태이다.
	StateEvaluating ExperimentState = "evaluating"
	// StateScoring은 평가 결과를 점수화 중인 상태이다.
	StateScoring ExperimentState = "scoring"
	// StateComplete는 실험 루프가 완료된 상태이다.
	StateComplete ExperimentState = "complete"
)

// Decision은 실험 결과에 대한 채택/기각 결정을 나타낸다.
type Decision string

const (
	// DecisionKeep은 실험 변경을 채택한다.
	DecisionKeep Decision = "keep"
	// DecisionDiscard는 실험 변경을 기각한다.
	DecisionDiscard Decision = "discard"
	// DecisionPending은 아직 결정이 내려지지 않은 상태이다.
	DecisionPending Decision = "pending"
)

// ChangeRecord는 실험에서 적용된 변경 내용을 기록한다.
type ChangeRecord struct {
	Type    string `json:"type"`    // addition | modification | deletion
	Section string `json:"section"` // 변경된 섹션 이름
	Diff    string `json:"diff"`    // 변경 내용의 diff 텍스트
}

// Experiment는 단일 실험의 전체 정보를 보관한다.
type Experiment struct {
	ID         string           `json:"id"`
	Target     string           `json:"target"`
	Hypothesis string           `json:"hypothesis"`
	Change     ChangeRecord     `json:"change"`
	Result     *eval.EvalResult `json:"result"`
	Decision   Decision         `json:"decision"`
	Timestamp  time.Time        `json:"timestamp"`
}

// ChangelogEntry는 실험 결과의 변경 이력 항목이다.
type ChangelogEntry struct {
	ExperimentID string   `json:"experiment_id"`
	Score        float64  `json:"score"`
	Change       string   `json:"change"`
	Reasoning    string   `json:"reasoning"`
	Decision     Decision `json:"decision"`
}
