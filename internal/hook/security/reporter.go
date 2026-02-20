package security

import (
	"fmt"
	"strings"
)

// findingReporter implements FindingReporter interface.
type findingReporter struct{}

// NewFindingReporter creates a new FindingReporter instance.
func NewFindingReporter() FindingReporter {
	return &findingReporter{}
}

// FormatResult formats a single scan result for output.
// Implements REQ-HOOK-130.
func (r *findingReporter) FormatResult(result *ScanResult, filePath string) string {
	if result == nil {
		return ""
	}

	// Handle unscanned result
	if !result.Scanned {
		if result.Error != "" {
			return fmt.Sprintf("AST-Grep scan skipped for %s: %s", filePath, result.Error)
		}
		return fmt.Sprintf("AST-Grep scan skipped for %s", filePath)
	}

	// Handle no findings
	if len(result.Findings) == 0 {
		return fmt.Sprintf("AST-Grep: No security issues found in %s", filePath)
	}

	// Build output with findings
	var sb strings.Builder

	// Summary line per SPEC 4.4
	fmt.Fprintf(&sb, "AST-Grep found %d error(s), %d warning(s) in %s\n",
		result.ErrorCount, result.WarningCount, filePath)

	// List findings (limited to MaxFindingsToReport per REQ-HOOK-132)
	displayCount := min(len(result.Findings), MaxFindingsToReport)

	for i := range displayCount {
		f := result.Findings[i]
		sb.WriteString(formatFinding(&f))
		sb.WriteString("\n")
	}

	// Add "... and N more" if truncated
	if len(result.Findings) > MaxFindingsToReport {
		remaining := len(result.Findings) - MaxFindingsToReport
		fmt.Fprintf(&sb, "  ... and %d more\n", remaining)
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// FormatMultiple formats multiple scan results.
func (r *findingReporter) FormatMultiple(results []*ScanResult) string {
	if len(results) == 0 {
		return ""
	}

	var sb strings.Builder
	totalErrors := 0
	totalWarnings := 0
	totalInfos := 0
	allFindings := []Finding{}

	// Aggregate findings from all results
	for _, result := range results {
		if result == nil || !result.Scanned {
			continue
		}
		totalErrors += result.ErrorCount
		totalWarnings += result.WarningCount
		totalInfos += result.InfoCount
		allFindings = append(allFindings, result.Findings...)
	}

	if len(allFindings) == 0 {
		return "AST-Grep: No security issues found"
	}

	// Summary line
	fmt.Fprintf(&sb, "AST-Grep found %d error(s), %d warning(s), %d info(s) across %d files\n",
		totalErrors, totalWarnings, totalInfos, len(results))

	// List findings (limited to MaxFindingsToReport)
	displayCount := min(len(allFindings), MaxFindingsToReport)

	for i := range displayCount {
		f := allFindings[i]
		sb.WriteString(formatFindingWithFile(&f))
		sb.WriteString("\n")
	}

	// Add "... and N more" if truncated
	if len(allFindings) > MaxFindingsToReport {
		remaining := len(allFindings) - MaxFindingsToReport
		fmt.Fprintf(&sb, "  ... and %d more\n", remaining)
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// ShouldExitWithError returns true if findings warrant exit code 2.
// Implements REQ-HOOK-131.
func (r *findingReporter) ShouldExitWithError(result *ScanResult) bool {
	if result == nil || !result.Scanned {
		return false
	}
	return result.ErrorCount > 0
}

// formatFinding formats a single finding for output.
func formatFinding(f *Finding) string {
	severityTag := getSeverityTag(f.Severity)
	return fmt.Sprintf("  - [%s] %s: %s (line %d)", severityTag, f.RuleID, f.Message, f.Line)
}

// formatFindingWithFile formats a single finding with file path for multi-file output.
func formatFindingWithFile(f *Finding) string {
	severityTag := getSeverityTag(f.Severity)
	return fmt.Sprintf("  - [%s] %s:%d %s: %s", severityTag, f.File, f.Line, f.RuleID, f.Message)
}

// getSeverityTag returns the display tag for a severity level.
func getSeverityTag(s Severity) string {
	switch s {
	case SeverityError:
		return "ERROR"
	case SeverityWarning:
		return "WARNING"
	case SeverityInfo:
		return "INFO"
	case SeverityHint:
		return "HINT"
	default:
		return "INFO"
	}
}
