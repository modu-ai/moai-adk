package eval

import "fmt"

// EvalEngine is the main interface for the evaluation system.
type EvalEngine interface {
	// LoadSuite loads an evaluation suite from a YAML file.
	LoadSuite(path string) (*EvalSuite, error)

	// Evaluate takes a suite and a results map and returns an aggregated evaluation result.
	Evaluate(suite *EvalSuite, results map[string]bool) (*EvalResult, error)
}

// engine is the default EvalEngine implementation.
type engine struct{}

// NewEngine creates a new EvalEngine.
func NewEngine() EvalEngine {
	return &engine{}
}

// LoadSuite loads an evaluation suite from a YAML file.
func (e *engine) LoadSuite(path string) (*EvalSuite, error) {
	return LoadSuite(path)
}

// Evaluate takes a suite and a results map and returns an aggregated evaluation result.
// Returns an error if suite is nil.
func (e *engine) Evaluate(suite *EvalSuite, results map[string]bool) (*EvalResult, error) {
	if suite == nil {
		return nil, fmt.Errorf("eval suite is nil")
	}

	return ComputeResult(suite.Criteria, results), nil
}
