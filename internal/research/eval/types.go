// Package eval은 Self-Research System의 Binary Eval Engine을 구현한다.
// 각 평가 기준을 pass/fail 이진 판정으로 처리하여
// 실험 결과를 집계한다.
package eval

import "time"

// CriterionWeight는 평가 기준의 중요도를 분류한다.
type CriterionWeight string

const (
	// MustPass는 반드시 통과해야 하는 필수 기준이다.
	MustPass CriterionWeight = "must_pass"
	// NiceToHave는 통과하면 좋지만 필수는 아닌 기준이다.
	NiceToHave CriterionWeight = "nice_to_have"
)

// validWeights는 허용되는 CriterionWeight 값 집합이다.
var validWeights = map[CriterionWeight]bool{
	MustPass:   true,
	NiceToHave: true,
}

// IsValid는 CriterionWeight 값이 유효한지 확인한다.
func (w CriterionWeight) IsValid() bool {
	return validWeights[w]
}

// TargetSpec은 리서치 대상을 식별한다.
type TargetSpec struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"` // skill | agent | rule | config
}

// TestInput은 평가용 테스트 시나리오를 정의한다.
type TestInput struct {
	Name   string `yaml:"name"`
	Prompt string `yaml:"prompt"`
}

// EvalCriterion은 이진 pass/fail 평가 기준을 정의한다.
type EvalCriterion struct {
	Name     string          `yaml:"name"`
	Question string          `yaml:"question"`
	Pass     string          `yaml:"pass"`
	Fail     string          `yaml:"fail"`
	Weight   CriterionWeight `yaml:"weight"`
}

// EvalSettings는 실험 실행 설정을 구성한다.
type EvalSettings struct {
	RunsPerExperiment int     `yaml:"runs_per_experiment"`
	MaxExperiments    int     `yaml:"max_experiments"`
	PassThreshold     float64 `yaml:"pass_threshold"`
	TargetScore       float64 `yaml:"target_score"`
	BudgetCapTokens   int     `yaml:"budget_cap_tokens"`
}

// CriterionResult는 단일 기준 평가 결과를 보관한다.
type CriterionResult struct {
	Name   string
	Passed bool
	Weight CriterionWeight
}

// EvalResult는 집계된 평가 결과를 보관한다.
type EvalResult struct {
	Overall      float64
	PerCriterion map[string]CriterionResult
	MustPassOK   bool
	Timestamp    time.Time
}
