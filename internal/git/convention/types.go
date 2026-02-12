// Package convention provides commit message convention validation,
// parsing, and detection for common commit formats.
package convention

import "regexp"

// ViolationType categorizes validation violations.
type ViolationType string

const (
	// ViolationPattern indicates the message does not match the expected pattern.
	ViolationPattern ViolationType = "pattern"

	// ViolationInvalidType indicates an unrecognized commit type.
	ViolationInvalidType ViolationType = "invalid_type"

	// ViolationInvalidScope indicates an unrecognized commit scope.
	ViolationInvalidScope ViolationType = "invalid_scope"

	// ViolationMaxLength indicates the header exceeds the maximum length.
	ViolationMaxLength ViolationType = "max_length"

	// ViolationRequired indicates a required field is missing.
	ViolationRequired ViolationType = "required"
)

// Violation represents a single validation error.
type Violation struct {
	Type       ViolationType
	Field      string
	Expected   string
	Actual     string
	Suggestion string
}

// ValidationResult contains the outcome of validating a commit message.
type ValidationResult struct {
	Valid      bool
	Message    string
	Violations []Violation
}

// ConventionConfig represents a convention definition loadable from YAML.
type ConventionConfig struct {
	Name      string   `yaml:"name"`
	Pattern   string   `yaml:"pattern"`
	Types     []string `yaml:"types"`
	Scopes    []string `yaml:"scopes"`
	MaxLength int      `yaml:"max_length"`
	Required  []string `yaml:"required"`
	Examples  []string `yaml:"examples"`
}

// Convention represents a compiled commit message convention ready for use.
type Convention struct {
	Name      string
	Pattern   *regexp.Regexp
	Types     []string
	Scopes    []string
	MaxLength int
	Required  []string
	Examples  []string
}

// DetectionResult contains the outcome of auto-detecting a convention
// from a set of commit messages.
type DetectionResult struct {
	Convention *Convention
	Confidence float64
	SampleSize int
	MatchCount int
}
