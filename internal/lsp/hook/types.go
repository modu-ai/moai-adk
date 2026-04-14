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

// PhaseType identifies the workflow phase in which a quality gate is evaluated.
// Values match the phase identifiers used in the .moai SPEC workflow.
type PhaseType string

const (
	// PhasePlan is the plan phase — captures LSP baseline state.
	PhasePlan PhaseType = "plan"

	// PhaseRun is the run phase — enforces error/type/lint thresholds.
	PhaseRun PhaseType = "run"

	// PhaseSync is the sync phase — enforces clean LSP requirement.
	PhaseSync PhaseType = "sync"

	// PhaseAuto detects the active phase from the hook invocation context.
	// Falls back to run-phase thresholds when the phase cannot be determined.
	PhaseAuto PhaseType = "auto"
)

// SeverityCounts represents counts of diagnostics by severity.
type SeverityCounts struct {
	Errors      int `json:"errors"`
	Warnings    int `json:"warnings"`
	Information int `json:"information"`
	Hints       int `json:"hints"`
	// TypeErrors is the count of type-checking errors (maps to max_type_errors).
	TypeErrors int `json:"typeErrors"`
	// LintErrors is the count of lint errors (maps to max_lint_errors).
	LintErrors int `json:"lintErrors"`
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
	// Phase identifies the workflow phase this gate applies to (REQ-QG-001).
	Phase PhaseType `json:"phase,omitempty"`

	// MaxErrors is the maximum allowed error count.
	MaxErrors int `json:"maxErrors"`

	// MaxWarnings is the maximum allowed warning count.
	MaxWarnings int `json:"maxWarnings"`

	// BlockOnError indicates whether to block on error threshold exceeded.
	BlockOnError bool `json:"blockOnError"`

	// BlockOnWarning indicates whether to block on warning threshold exceeded.
	BlockOnWarning bool `json:"blockOnWarning"`
}

// PhaseEnforceResult holds the outcome of a phase-aware gate enforcement check.
type PhaseEnforceResult struct {
	// ShouldBlock is true if the gate determines execution should be blocked.
	ShouldBlock bool `json:"shouldBlock"`

	// Phase is the phase that was evaluated.
	Phase PhaseType `json:"phase"`

	// BaselineCaptured is true when the plan phase recorded a new baseline (REQ-QG-002).
	BaselineCaptured bool `json:"baselineCaptured"`

	// HasRegression is true when a regression was detected against the baseline (REQ-QG-007).
	HasRegression bool `json:"hasRegression"`

	// ViolatedThreshold describes which threshold was exceeded, if any.
	ViolatedThreshold string `json:"violatedThreshold,omitempty"`
}

// ExitCode converts the enforcement result to a process exit code.
// Returns ExitCodeQualityGateFailed (2) when ShouldBlock is true,
// otherwise ExitCodeSuccess (0).
func (r PhaseEnforceResult) ExitCode() int {
	if r.ShouldBlock {
		return ExitCodeQualityGateFailed
	}
	return ExitCodeSuccess
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

	// EnforcePhase evaluates the quality gate for the given phase (REQ-QG-001 through REQ-QG-004).
	// baseline is the prior diagnostic snapshot used for regression detection (REQ-QG-007).
	// Pass nil baseline to skip regression comparison.
	EnforcePhase(phase PhaseType, counts SeverityCounts, baseline *SeverityCounts) (PhaseEnforceResult, error)

	// LoadPhaseAwareConfig loads all phase-specific config fields via pkg/models (REQ-QG-005, REQ-QG-006).
	LoadPhaseAwareConfig() (LSPQualityGatesConfig, error)

	// LoadTRUST5Config loads the trust5_integration dimension mapping (REQ-QG-009).
	LoadTRUST5Config() (TRUST5Config, error)
}

// LSPQualityGatesConfig is the gate-layer view of models.LSPQualityGates.
// It mirrors models.LSPQualityGates so callers in this package have a local type
// without a hard dependency on pkg/models in every file.
type LSPQualityGatesConfig struct {
	Enabled         bool               `json:"enabled"`
	Plan            LSPPlanGateConfig  `json:"plan"`
	Run             LSPRunGateConfig   `json:"run"`
	Sync            LSPSyncGateConfig  `json:"sync"`
	CacheTTLSeconds int                `json:"cacheTTLSeconds"`
	TimeoutSeconds  int                `json:"timeoutSeconds"`
}

// LSPPlanGateConfig holds plan-phase gate settings.
type LSPPlanGateConfig struct {
	RequireBaseline bool `json:"requireBaseline"`
}

// LSPRunGateConfig holds run-phase gate settings.
type LSPRunGateConfig struct {
	MaxErrors       int  `json:"maxErrors"`
	MaxTypeErrors   int  `json:"maxTypeErrors"`
	MaxLintErrors   int  `json:"maxLintErrors"`
	AllowRegression bool `json:"allowRegression"`
}

// LSPSyncGateConfig holds sync-phase gate settings.
type LSPSyncGateConfig struct {
	MaxErrors       int  `json:"maxErrors"`
	MaxWarnings     int  `json:"maxWarnings"`
	RequireCleanLSP bool `json:"requireCleanLSP"`
}

// TRUST5Config holds the trust5_integration dimension mapping (REQ-QG-009).
type TRUST5Config struct {
	Tested         []string `json:"tested"`
	Readable       []string `json:"readable"`
	Understandable []string `json:"understandable"`
	Secured        []string `json:"secured"`
	Trackable      []string `json:"trackable"`
}

// RegressionDetectionConfig holds regression detection thresholds (REQ-QG-007).
type RegressionDetectionConfig struct {
	ErrorIncreaseThreshold     int `json:"errorIncreaseThreshold"`
	WarningIncreaseThreshold   int `json:"warningIncreaseThreshold"`
	TypeErrorIncreaseThreshold int `json:"typeErrorIncreaseThreshold"`
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

// @MX:ANCHOR: [AUTO] Error sentinel used across quality gate error handling throughout all LSP hook paths
// @MX:REASON: fan_in=44, quality gate enforcer, fallback diagnostics, and session tracker all branch on this error type; changing the message format breaks error string comparisons
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
