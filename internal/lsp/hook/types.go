// Package hook provides LSP diagnostics integration for MoAI-ADK hooks.
// It implements REQ-HOOK-150 through REQ-HOOK-191 from SPEC-HOOK-004.
package hook

import (
	"context"
	"time"
)

// DiagnosticSeverity represents the severity of a diagnostic.
// Values are string constants for JSON serialization compatibility.
type DiagnosticSeverity string

const (
	// SeverityError reports an error.
	SeverityError DiagnosticSeverity = "error"

	// SeverityWarning reports a warning.
	SeverityWarning DiagnosticSeverity = "warning"

	// SeverityInformation reports an information message.
	SeverityInformation DiagnosticSeverity = "information"

	// SeverityHint reports a hint.
	SeverityHint DiagnosticSeverity = "hint"
)

// Position represents a zero-based position in a text document.
type Position struct {
	// Line is the zero-based line number.
	Line int `json:"line"`

	// Character is the zero-based character offset on the line.
	Character int `json:"character"`
}

// Range represents a range in a text document defined by start and end positions.
type Range struct {
	// Start is the range's start position (inclusive).
	Start Position `json:"start"`

	// End is the range's end position (exclusive).
	End Position `json:"end"`
}

// Diagnostic represents a single diagnostic (error, warning, etc.).
type Diagnostic struct {
	// Range is the range at which the message applies.
	Range Range `json:"range"`

	// Severity is the diagnostic's severity level.
	Severity DiagnosticSeverity `json:"severity"`

	// Code is the diagnostic's code (e.g., "E0001"). May be empty.
	Code string `json:"code,omitempty"`

	// Source identifies the tool that produced this diagnostic (e.g., "gopls").
	Source string `json:"source,omitempty"`

	// Message is the diagnostic's human-readable message.
	Message string `json:"message"`
}

// SeverityCounts represents counts of diagnostics by severity.
type SeverityCounts struct {
	Errors      int `json:"errors"`
	Warnings    int `json:"warnings"`
	Information int `json:"information"`
	Hints       int `json:"hints"`
}

// Total returns the total count of all diagnostics.
func (s SeverityCounts) Total() int {
	return s.Errors + s.Warnings + s.Information + s.Hints
}

// RegressionReport compares current diagnostics with baseline.
type RegressionReport struct {
	// HasRegression is true if new errors were introduced.
	HasRegression bool `json:"hasRegression"`

	// HasImprovement is true if errors were fixed.
	HasImprovement bool `json:"hasImprovement"`

	// NewErrors is the count of new errors since baseline.
	NewErrors int `json:"newErrors"`

	// FixedErrors is the count of fixed errors since baseline.
	FixedErrors int `json:"fixedErrors"`

	// NewWarnings is the count of new warnings since baseline.
	NewWarnings int `json:"newWarnings"`

	// FixedWarnings is the count of fixed warnings since baseline.
	FixedWarnings int `json:"fixedWarnings"`
}

// QualityGate defines quality gate thresholds.
type QualityGate struct {
	// MaxErrors is the maximum allowed error count.
	MaxErrors int `json:"maxErrors"`

	// MaxWarnings is the maximum allowed warning count.
	MaxWarnings int `json:"maxWarnings"`

	// BlockOnError indicates whether to block on error threshold exceeded.
	BlockOnError bool `json:"blockOnError"`

	// BlockOnWarning indicates whether to block on warning threshold exceeded.
	BlockOnWarning bool `json:"blockOnWarning"`
}

// FileBaseline represents the diagnostic baseline for a single file.
type FileBaseline struct {
	// Path is the file path.
	Path string `json:"path"`

	// Hash is the content hash of the file when baseline was captured.
	Hash string `json:"hash"`

	// Diagnostics are the diagnostics at baseline time.
	Diagnostics []Diagnostic `json:"diagnostics"`

	// UpdatedAt is when this baseline was last updated.
	UpdatedAt time.Time `json:"updatedAt"`
}

// DiagnosticsBaseline represents the complete baseline state.
type DiagnosticsBaseline struct {
	// Version is the baseline format version.
	Version string `json:"version"`

	// UpdatedAt is when this baseline was last updated.
	UpdatedAt time.Time `json:"updatedAt"`

	// Files maps file paths to their baselines.
	Files map[string]FileBaseline `json:"files"`
}

// SessionStats tracks cumulative diagnostic statistics for a session.
type SessionStats struct {
	// TotalErrors is the cumulative error count.
	TotalErrors int `json:"totalErrors"`

	// TotalWarnings is the cumulative warning count.
	TotalWarnings int `json:"totalWarnings"`

	// TotalInformation is the cumulative information count.
	TotalInformation int `json:"totalInformation"`

	// TotalHints is the cumulative hint count.
	TotalHints int `json:"totalHints"`

	// FilesAnalyzed is the count of unique files analyzed.
	FilesAnalyzed int `json:"filesAnalyzed"`

	// StartedAt is when the session started.
	StartedAt time.Time `json:"startedAt"`
}

// FileStats tracks per-file diagnostic history.
type FileStats struct {
	// Path is the file path.
	Path string `json:"path"`

	// DiagnosticHistory is the history of diagnostic counts.
	DiagnosticHistory []SeverityCounts `json:"diagnosticHistory"`

	// LastAnalyzed is when the file was last analyzed.
	LastAnalyzed time.Time `json:"lastAnalyzed"`
}

// LSPDiagnosticsCollector collects LSP diagnostics.
// Implementations must be thread-safe.
type LSPDiagnosticsCollector interface {
	// GetDiagnostics retrieves diagnostics for the given file path.
	// Returns an empty slice (not nil) if no diagnostics are found.
	GetDiagnostics(ctx context.Context, filePath string) ([]Diagnostic, error)

	// GetSeverityCounts calculates severity counts from diagnostics.
	GetSeverityCounts(diagnostics []Diagnostic) SeverityCounts
}

// FallbackDiagnostics uses CLI tools when LSP is unavailable.
// Implementations must be thread-safe.
type FallbackDiagnostics interface {
	// RunFallback executes fallback CLI tool for the given file.
	// Returns "diagnostics unavailable" error if no tool is available.
	RunFallback(ctx context.Context, filePath string) ([]Diagnostic, error)

	// IsAvailable checks if a fallback tool is available for the language.
	IsAvailable(language string) bool

	// GetLanguage returns the detected language for a file path.
	GetLanguage(filePath string) string
}

// RegressionTracker tracks diagnostic baselines and detects regressions.
// Implementations must be thread-safe.
type RegressionTracker interface {
	// SaveBaseline saves the current diagnostics as baseline for a file.
	SaveBaseline(filePath string, diagnostics []Diagnostic) error

	// CompareWithBaseline compares current diagnostics against baseline.
	// Returns empty report if no baseline exists.
	CompareWithBaseline(filePath string, diagnostics []Diagnostic) (RegressionReport, error)

	// GetBaseline retrieves the baseline for a file.
	GetBaseline(filePath string) (*FileBaseline, error)

	// ClearBaseline removes the baseline for a file.
	ClearBaseline(filePath string) error
}

// QualityGateEnforcer enforces quality gate rules.
type QualityGateEnforcer interface {
	// ShouldBlock determines if the counts exceed gate thresholds.
	// Returns true if execution should be blocked.
	ShouldBlock(counts SeverityCounts, gate QualityGate) bool

	// LoadConfig loads quality gate configuration from YAML.
	LoadConfig() (QualityGate, error)

	// CheckWithConfig loads config and checks if should block.
	CheckWithConfig(counts SeverityCounts) (shouldBlock bool, gate QualityGate, err error)
}

// SessionTracker tracks diagnostic statistics for a session.
// Implementations must be thread-safe.
type SessionTracker interface {
	// StartSession initializes a new session.
	StartSession() error

	// RecordDiagnostics records diagnostics for a file.
	RecordDiagnostics(filePath string, diagnostics []Diagnostic)

	// GetSessionStats returns the current session statistics.
	GetSessionStats() SessionStats

	// GetFileStats returns statistics for a specific file.
	GetFileStats(filePath string) (*FileStats, error)

	// EndSession finalizes the session and returns summary.
	EndSession() (SessionStats, error)
}

// ErrDiagnosticsUnavailable is returned when no diagnostic tool is available.
type ErrDiagnosticsUnavailable struct {
	Language string
	Reason   string
}

func (e *ErrDiagnosticsUnavailable) Error() string {
	if e.Reason != "" {
		return "diagnostics unavailable for " + e.Language + ": " + e.Reason
	}
	return "diagnostics unavailable for " + e.Language
}

// ErrBaselineNotFound is returned when no baseline exists for a file.
type ErrBaselineNotFound struct {
	FilePath string
}

func (e *ErrBaselineNotFound) Error() string {
	return "baseline not found for " + e.FilePath
}
