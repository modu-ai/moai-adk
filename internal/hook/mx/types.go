// Package mx provides MX tag auto-validation for Go source files.
// It detects missing @MX annotations (ANCHOR, WARN, NOTE, TODO) based on
// code structure analysis and classifies violations by priority (P1-P4).
//
// This package is READ-ONLY: it never modifies source files or tags.
package mx

import (
	"context"
	"time"
)

// Priority represents the severity of an MX tag violation.
type Priority int

const (
	// P1 is blocking: exported function with fan_in >= 3 missing @MX:ANCHOR.
	P1 Priority = 1
	// P2 is blocking: goroutine pattern missing @MX:WARN.
	P2 Priority = 2
	// P3 is advisory: exported function >= 100 lines missing @MX:NOTE.
	P3 Priority = 3
	// P4 is advisory: untested public function missing @MX:TODO.
	P4 Priority = 4
)

// String returns the string representation of a Priority.
func (p Priority) String() string {
	switch p {
	case P1:
		return "P1"
	case P2:
		return "P2"
	case P3:
		return "P3"
	case P4:
		return "P4"
	default:
		return "unknown"
	}
}

// IsBlocking returns true if this priority level blocks sync/CI.
func (p Priority) IsBlocking() bool {
	return p == P1 || p == P2
}

// Violation represents a single missing @MX tag violation.
type Violation struct {
	// FuncName is the exported function name.
	FuncName string
	// FilePath is the absolute path to the file.
	FilePath string
	// Line is the 1-indexed line number of the function declaration.
	Line int
	// Priority is the violation severity (P1-P4).
	Priority Priority
	// FanIn is the number of callers (used for P1 violations).
	FanIn int
	// MissingTag is the tag that should be present (e.g., "@MX:ANCHOR").
	MissingTag string
	// Reason describes why this violation was detected.
	Reason string
	// Blocking indicates whether this violation blocks sync/CI.
	Blocking bool
}

// FileReport holds validation results for a single file.
type FileReport struct {
	// FilePath is the absolute path to the validated file.
	FilePath string
	// Violations contains all detected violations.
	Violations []Violation
	// Fallback indicates the validation used Grep fallback (ast-grep unavailable).
	Fallback bool
	// Duration is the time taken to validate this file.
	Duration time.Duration
	// Error is set if validation failed for this file.
	Error error
	// TimedOut indicates validation was cancelled due to timeout.
	TimedOut bool
}

// P1Count returns the number of P1 violations in this file report.
func (r *FileReport) P1Count() int { return r.countByPriority(P1) }

// P2Count returns the number of P2 violations in this file report.
func (r *FileReport) P2Count() int { return r.countByPriority(P2) }

// P3Count returns the number of P3 violations in this file report.
func (r *FileReport) P3Count() int { return r.countByPriority(P3) }

// P4Count returns the number of P4 violations in this file report.
func (r *FileReport) P4Count() int { return r.countByPriority(P4) }

func (r *FileReport) countByPriority(p Priority) int {
	count := 0
	for _, v := range r.Violations {
		if v.Priority == p {
			count++
		}
	}
	return count
}

// ValidationReport holds aggregate validation results across multiple files.
type ValidationReport struct {
	// FileReports contains per-file validation results (completed files only).
	FileReports []*FileReport
	// TimedOutFiles contains paths of files that timed out during validation.
	TimedOutFiles []string
	// Duration is the total time taken for the batch validation.
	Duration time.Duration
	// Fallback indicates at least one file used Grep fallback.
	Fallback bool
}

// P1Count returns the total number of P1 violations across all files.
func (r *ValidationReport) P1Count() int { return r.countByPriority(P1) }

// P2Count returns the total number of P2 violations across all files.
func (r *ValidationReport) P2Count() int { return r.countByPriority(P2) }

// P3Count returns the total number of P3 violations across all files.
func (r *ValidationReport) P3Count() int { return r.countByPriority(P3) }

// P4Count returns the total number of P4 violations across all files.
func (r *ValidationReport) P4Count() int { return r.countByPriority(P4) }

// TotalViolations returns the total number of violations across all files.
func (r *ValidationReport) TotalViolations() int {
	total := 0
	for _, fr := range r.FileReports {
		total += len(fr.Violations)
	}
	return total
}

// HasBlockingViolations returns true if there are any P1 or P2 violations.
func (r *ValidationReport) HasBlockingViolations() bool {
	return r.P1Count() > 0 || r.P2Count() > 0
}

func (r *ValidationReport) countByPriority(p Priority) int {
	count := 0
	for _, fr := range r.FileReports {
		count += fr.countByPriority(p)
	}
	return count
}

// Validator is the interface for MX tag validation.
// Implementations must be thread-safe: concurrent ValidateFile calls must work.
type Validator interface {
	// ValidateFile validates a single Go source file for missing @MX tags.
	// Returns a FileReport describing any violations found.
	// Context is used for timeout control.
	ValidateFile(ctx context.Context, filePath string) (*FileReport, error)

	// ValidateFiles validates multiple Go source files in parallel.
	// Returns partial results if context is cancelled (completed files only).
	// Never returns an error for timeout: partial results are returned instead.
	ValidateFiles(ctx context.Context, filePaths []string) (*ValidationReport, error)
}

// FormatReport formats a ValidationReport as a human-readable string.
// Format includes summary and per-priority violation sections.
func FormatReport(report *ValidationReport) string {
	return formatReport(report)
}
