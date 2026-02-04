// Package security provides AST-Grep based security scanning for Claude Code hooks.
// It implements SPEC-HOOK-003: Security & Scanning.
package security

import (
	"context"
	"time"
)

// Severity represents the severity level of a security finding.
type Severity string

const (
	// SeverityError indicates a critical security issue that should be addressed.
	SeverityError Severity = "error"

	// SeverityWarning indicates a potential security issue.
	SeverityWarning Severity = "warning"

	// SeverityInfo indicates an informational security note.
	SeverityInfo Severity = "info"

	// SeverityHint indicates a minor suggestion.
	SeverityHint Severity = "hint"
)

// Finding represents a single security finding from AST-Grep scan.
type Finding struct {
	RuleID    string   `json:"ruleId"`
	Severity  Severity `json:"severity"`
	Message   string   `json:"message"`
	File      string   `json:"file"`
	Line      int      `json:"line"`
	Column    int      `json:"column,omitempty"`
	EndLine   int      `json:"endLine,omitempty"`
	EndColumn int      `json:"endColumn,omitempty"`
	Code      string   `json:"code,omitempty"`
}

// ScanResult represents the result of a security scan.
type ScanResult struct {
	Scanned      bool          `json:"scanned"`
	ErrorCount   int           `json:"errorCount"`
	WarningCount int           `json:"warningCount"`
	InfoCount    int           `json:"infoCount"`
	Findings     []Finding     `json:"findings"`
	Error        string        `json:"error,omitempty"`
	Duration     time.Duration `json:"duration"`
}

// CountBySeverity returns counts of findings by severity.
func (r *ScanResult) CountBySeverity() (errors, warnings, infos int) {
	for _, f := range r.Findings {
		switch f.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		case SeverityInfo, SeverityHint:
			infos++
		}
	}
	return errors, warnings, infos
}

// HasErrors returns true if the result contains error-severity findings.
func (r *ScanResult) HasErrors() bool {
	return r.ErrorCount > 0
}

// ASTGrepScanner handles AST-Grep security scanning.
// Implements REQ-HOOK-100 through REQ-HOOK-103.
type ASTGrepScanner interface {
	// IsAvailable checks if ast-grep (sg) binary is available.
	// Implements REQ-HOOK-100.
	IsAvailable() bool

	// GetVersion returns the ast-grep version string.
	GetVersion() string

	// Scan runs ast-grep scan on a single file.
	// Implements REQ-HOOK-101, REQ-HOOK-102.
	Scan(ctx context.Context, filePath string, configPath string) (*ScanResult, error)

	// ScanMultiple runs ast-grep scan on multiple files.
	// Implements REQ-HOOK-123 (optional parallel scanning).
	ScanMultiple(ctx context.Context, filePaths []string, configPath string) ([]*ScanResult, error)
}

// RuleManager manages security rule configuration.
// Implements REQ-HOOK-110 through REQ-HOOK-112.
type RuleManager interface {
	// FindRulesConfig finds the rules configuration file in project directory.
	// Implements REQ-HOOK-110.
	FindRulesConfig(projectDir string) string

	// LoadRules loads rules from a configuration file.
	// Implements REQ-HOOK-110, REQ-HOOK-112.
	LoadRules(configPath string) ([]string, error)

	// GetDefaultRules returns the default OWASP security rules.
	// Implements REQ-HOOK-111.
	GetDefaultRules() []string

	// GetEffectiveRules returns rules to use for a project.
	// If project has a config, use that; otherwise use defaults.
	// Implements REQ-HOOK-111, REQ-HOOK-112.
	GetEffectiveRules(projectDir string) []string
}

// FindingReporter formats scan results for Claude Code output.
// Implements REQ-HOOK-130 through REQ-HOOK-132.
type FindingReporter interface {
	// FormatResult formats a single scan result for output.
	// Implements REQ-HOOK-130.
	FormatResult(result *ScanResult, filePath string) string

	// FormatMultiple formats multiple scan results.
	FormatMultiple(results []*ScanResult) string

	// ShouldExitWithError returns true if findings warrant exit code 2.
	// Implements REQ-HOOK-131.
	ShouldExitWithError(result *ScanResult) bool
}

// SupportedLanguage represents a programming language supported by AST-Grep.
type SupportedLanguage struct {
	Name       string
	Extensions []string
	ASTGrepID  string // ast-grep language identifier
}

// DefaultScanTimeout is the default timeout for security scans (30 seconds).
// Per REQ-HOOK-120.
const DefaultScanTimeout = 30 * time.Second

// MaxFindingsToReport is the maximum number of findings to show in output.
// Per REQ-HOOK-132.
const MaxFindingsToReport = 10
