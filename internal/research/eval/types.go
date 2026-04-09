// Package eval implements the Binary Eval Engine for the Self-Research System.
// Each evaluation criterion is processed as a binary pass/fail judgment to
// aggregate experiment results.
package eval

import "time"

// CriterionWeight classifies the importance of an evaluation criterion.
type CriterionWeight string

const (
	// MustPass is a mandatory criterion that must be passed.
	MustPass CriterionWeight = "must_pass"
	// NiceToHave is a criterion that is good to pass but not required.
	NiceToHave CriterionWeight = "nice_to_have"
)

// validWeights is the set of allowed CriterionWeight values.
var validWeights = map[CriterionWeight]bool{
	MustPass:   true,
	NiceToHave: true,
}

// IsValid checks whether the CriterionWeight value is valid.
func (w CriterionWeight) IsValid() bool {
	return validWeights[w]
}

// TargetSpec identifies the research target.
type TargetSpec struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"` // skill | agent | rule | config
}

// TestInput defines a test scenario for evaluation.
type TestInput struct {
	Name   string `yaml:"name"`
	Prompt string `yaml:"prompt"`
}

// EvalCriterion defines a binary pass/fail evaluation criterion.
type EvalCriterion struct {
	Name     string          `yaml:"name"`
	Question string          `yaml:"question"`
	Pass     string          `yaml:"pass"`
	Fail     string          `yaml:"fail"`
	Weight   CriterionWeight `yaml:"weight"`
}

// EvalSettings configures the experiment execution settings.
type EvalSettings struct {
	RunsPerExperiment int     `yaml:"runs_per_experiment"`
	MaxExperiments    int     `yaml:"max_experiments"`
	PassThreshold     float64 `yaml:"pass_threshold"`
	TargetScore       float64 `yaml:"target_score"`
	BudgetCapTokens   int     `yaml:"budget_cap_tokens"`
}

// CriterionResult holds the evaluation result for a single criterion.
type CriterionResult struct {
	Name   string
	Passed bool
	Weight CriterionWeight
}

// EvalResult holds the aggregated evaluation result.
type EvalResult struct {
	Overall      float64
	PerCriterion map[string]CriterionResult
	MustPassOK   bool
	Timestamp    time.Time
}
