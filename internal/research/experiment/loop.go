package experiment

import (
	"github.com/modu-ai/moai-adk/internal/research/eval"
)

// LoopConfig는 실험 루프의 실행 조건을 설정한다.
type LoopConfig struct {
	MaxExperiments      int     // 최대 실험 횟수
	TargetScore         float64 // 목표 점수 (연속 3회 달성 시 종료)
	StagnationThreshold float64 // 정체 판단 기준 (개선 폭이 이 값 미만이면 정체)
	StagnationPatience  int     // 정체 허용 횟수 (초과 시 종료)
}

// Loop는 실험 루프 상태 머신이다.
// 베이스라인 설정 후, 실험 결과를 기록하며 지속 여부를 판단한다.
type Loop struct {
	config          LoopConfig
	state           ExperimentState
	baseline        *eval.EvalResult
	bestScore       float64
	history         []*Experiment
	stagnationCount int
	targetHitCount  int
}

// NewLoop는 주어진 설정으로 새 실험 루프를 생성한다.
// 초기 상태는 StateIdle이다.
func NewLoop(config LoopConfig) *Loop {
	return &Loop{
		config:  config,
		state:   StateIdle,
		history: make([]*Experiment, 0),
	}
}

// SetBaseline은 베이스라인 평가 결과를 설정하고 상태를 StateBaseline으로 전이한다.
func (l *Loop) SetBaseline(result *eval.EvalResult) {
	l.baseline = result
	l.bestScore = result.Overall
	l.state = StateBaseline
}

// ShouldContinue는 실험 루프를 계속 진행해야 하는지 판단한다.
// false를 반환하는 조건:
//  1. 실험 수가 MaxExperiments에 도달
//  2. 목표 점수를 연속 3회 달성
//  3. 정체 횟수가 StagnationPatience 초과
//  4. 상태가 StateComplete
func (l *Loop) ShouldContinue() bool {
	if l.state == StateComplete {
		return false
	}
	if len(l.history) >= l.config.MaxExperiments {
		return false
	}
	if l.targetHitCount >= 3 {
		return false
	}
	if l.stagnationCount >= l.config.StagnationPatience {
		return false
	}
	return true
}

// RecordExperiment는 실험 결과를 기록하고 채택/기각 결정을 반환한다.
// 결정 로직:
//  1. Result가 nil → DecisionDiscard
//  2. MustPassOK가 false → DecisionDiscard
//  3. Overall이 bestScore보다 높으면 → DecisionKeep (bestScore 갱신)
//  4. 그 외 → DecisionDiscard
//
// 부수 효과로 정체 카운터와 목표 달성 카운터를 갱신한다.
func (l *Loop) RecordExperiment(exp *Experiment) Decision {
	l.history = append(l.history, exp)

	// nil 결과 처리
	if exp.Result == nil {
		exp.Decision = DecisionDiscard
		l.stagnationCount++
		l.targetHitCount = 0
		return DecisionDiscard
	}

	// must_pass 실패 처리
	if !exp.Result.MustPassOK {
		exp.Decision = DecisionDiscard
		l.stagnationCount++
		l.targetHitCount = 0
		return DecisionDiscard
	}

	// 개선 폭 계산
	improvement := exp.Result.Overall - l.bestScore

	// 채택/기각 결정
	var decision Decision
	if exp.Result.Overall > l.bestScore {
		decision = DecisionKeep
		l.bestScore = exp.Result.Overall
	} else {
		decision = DecisionDiscard
	}
	exp.Decision = decision

	// 정체 추적: 개선 폭이 임계값 미만이면 정체 카운터 증가
	if improvement < l.config.StagnationThreshold {
		l.stagnationCount++
	} else {
		l.stagnationCount = 0
	}

	// 목표 달성 추적: 목표 점수 이상이면 연속 카운터 증가
	if exp.Result.Overall >= l.config.TargetScore {
		l.targetHitCount++
	} else {
		l.targetHitCount = 0
	}

	return decision
}

// BestScore는 현재까지의 최고 점수를 반환한다.
func (l *Loop) BestScore() float64 {
	return l.bestScore
}

// ExperimentCount는 기록된 실험의 총 수를 반환한다.
func (l *Loop) ExperimentCount() int {
	return len(l.history)
}

// State는 실험 루프의 현재 상태를 반환한다.
func (l *Loop) State() ExperimentState {
	return l.state
}
