package eval

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// EvalSuite defines the complete evaluation configuration loaded from YAML.
type EvalSuite struct {
	Target   TargetSpec      `yaml:"target"`
	Inputs   []TestInput     `yaml:"test_inputs"`
	Criteria []EvalCriterion `yaml:"evals"`
	Settings EvalSettings    `yaml:"settings"`
}

// LoadSuite reads and parses an evaluation suite from a YAML file.
// Returns an error if the file does not exist, the YAML is invalid, or validation fails.
func LoadSuite(path string) (*EvalSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read eval suite file: %w", err)
	}

	var suite EvalSuite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse eval suite YAML: %w", err)
	}

	if err := suite.Validate(); err != nil {
		return nil, fmt.Errorf("eval suite validation failed: %w", err)
	}

	return &suite, nil
}

// Validate checks that the suite is correctly configured.
// Requires at least 1 criterion, valid weights, and at least 1 test input.
func (s *EvalSuite) Validate() error {
	if len(s.Criteria) == 0 {
		return fmt.Errorf("at least one evaluation criterion (evals) is required")
	}

	if len(s.Inputs) == 0 {
		return fmt.Errorf("at least one test input (test_inputs) is required")
	}

	for i, c := range s.Criteria {
		if !c.Weight.IsValid() {
			return fmt.Errorf("criterion[%d] %q has invalid weight %q (only must_pass or nice_to_have are allowed)", i, c.Name, c.Weight)
		}
	}

	return nil
}
