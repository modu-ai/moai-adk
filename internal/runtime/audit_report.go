// Package runtime provides core runtime utilities for MoAI workflow operations.
// Source: SPEC-WF-AUDIT-GATE-001 REQ-WAG-004
package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AuditReporter writes audit run records to the daily plan-audit report file.
//
// Daily report path: .moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md
// Multiple runs on the same day append to the same file (AC-WAG-04).
//
// @MX:ANCHOR: [AUTO] AuditReporter is the primary interface for persisting audit verdicts
// @MX:REASON: fan_in >= 3 (GateConfig.Invoke, test mocks, daily report validation)
type AuditReporter interface {
	// AppendRun appends a single audit run record to the daily report file.
	//
	// If the report directory does not exist, it is created automatically (AC-WAG-10).
	// If the file exists, the run is appended; otherwise the file is created.
	// Reports persist to .moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md.
	//
	// REQ-WAG-004
	AppendRun(specID string, result *AuditResult) error
}

// FileAuditReporter implements AuditReporter by writing to the filesystem.
type FileAuditReporter struct {
	// ReportDir is the directory where reports are written.
	// Typically .moai/reports/plan-audit/ relative to the project root.
	ReportDir string

	// ProjectDir is the project root for path safety validation.
	ProjectDir string

	// Clock provides the current time for date-stamping report files.
	Clock Clock
}

// NewFileAuditReporter creates a FileAuditReporter with a SystemClock.
func NewFileAuditReporter(projectDir, reportDir string) *FileAuditReporter {
	return &FileAuditReporter{
		ReportDir:  reportDir,
		ProjectDir: projectDir,
		Clock:      SystemClock{},
	}
}

// reportFilePath returns the safe, date-stamped report file path for a SPEC.
// REQ-WAG-004: path must be within .moai/reports/plan-audit/.
func (r *FileAuditReporter) reportFilePath(specID string, date time.Time) (string, error) {
	dateStr := date.UTC().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", specID, dateStr)
	fullPath := filepath.Clean(filepath.Join(r.ReportDir, filename))

	// Path traversal prevention: report must be inside ReportDir.
	cleanDir := filepath.Clean(r.ReportDir)
	if !strings.HasPrefix(fullPath, cleanDir+string(filepath.Separator)) &&
		fullPath != cleanDir {
		return "", fmt.Errorf("computed report path %q is outside report dir %q (path traversal prevented)",
			fullPath, cleanDir)
	}

	return fullPath, nil
}

// AppendRun appends an audit run record to the daily report file.
//
// The run is formatted as a Markdown section matching the format defined in run.md Phase 0.5 Step 5.
func (r *FileAuditReporter) AppendRun(specID string, result *AuditResult) error {
	if err := ValidateReportDir(r.ProjectDir, r.ReportDir); err != nil {
		return fmt.Errorf("validate report dir: %w", err)
	}

	now := r.Clock.Now().UTC()
	filePath, err := r.reportFilePath(specID, now)
	if err != nil {
		return err
	}

	// Count existing runs to number this one correctly.
	runNum, err := countExistingRuns(filePath)
	if err != nil {
		return fmt.Errorf("count runs in %q: %w", filePath, err)
	}
	runNum++

	// Build the run trigger label.
	trigger := runTriggerLabel(result.Verdict)

	// Sanitize bypass reason if present.
	bypassReason := ""
	if result.Verdict == VerdictBypassed {
		bypassReason = SanitizeMarkdown(result.BypassReason)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "\n## Audit Run %d\n\n", runNum)
	fmt.Fprintf(&sb, "- verdict: %s\n", result.Verdict)

	if result.ReportPath != "" {
		fmt.Fprintf(&sb, "- report_path: %s\n", result.ReportPath)
	}

	fmt.Fprintf(&sb, "- audit_at: %s\n", result.AuditAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(&sb, "- run_trigger: %s\n", trigger)

	if result.AuditorVersion != "" {
		fmt.Fprintf(&sb, "- auditor_version: %s\n", result.AuditorVersion)
	}

	if result.CacheHit {
		fmt.Fprintf(&sb, "- cache_hit: true\n")
		fmt.Fprintf(&sb, "- cached_audit_at: %s\n", result.CachedAuditAt.UTC().Format(time.RFC3339))
	}

	if result.Verdict == VerdictBypassed {
		fmt.Fprintf(&sb, "- bypass_user: %s\n", result.BypassUser)
		fmt.Fprintf(&sb, "- bypass_reason: %q\n", bypassReason)
	}

	if result.InconclusiveAcknowledgedBy != "" {
		fmt.Fprintf(&sb, "- inconclusive_acknowledged_by: %s\n", result.InconclusiveAcknowledgedBy)
	}

	if result.GraceWindowActive {
		sb.WriteString("- grace_window_active: true\n")
		fmt.Fprintf(&sb, "- grace_window_remaining_days: %d\n", result.GraceWindowRemainingDays)
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open report file %q: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(sb.String()); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	return nil
}

// runTriggerLabel returns the run_trigger label for the given verdict.
func runTriggerLabel(v Verdict) string {
	switch v {
	case VerdictBypassed:
		return "bypassed"
	case VerdictInconclusive:
		return "inconclusive"
	default:
		return "automatic"
	}
}

// countExistingRuns counts how many "## Audit Run" sections exist in the file.
// Returns 0 if the file does not exist.
func countExistingRuns(filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	count := strings.Count(string(data), "## Audit Run ")
	return count, nil
}

// ProgressEntry holds the 4 required fields to append to progress.md. REQ-WAG-003.
type ProgressEntry struct {
	// Verdict is the audit verdict (PASS, FAIL, FAIL_WARNED, BYPASSED, INCONCLUSIVE).
	Verdict Verdict

	// ReportPath is the path to the iteration-scoped audit report.
	ReportPath string

	// AuditAt is the UTC timestamp of the audit.
	AuditAt time.Time

	// AuditorVersion is the plan-auditor identifier.
	AuditorVersion string

	// CacheHit indicates a 24h cache was used.
	CacheHit bool

	// CachedAuditAt is the original cache entry timestamp.
	CachedAuditAt time.Time

	// InconclusiveAcknowledgedBy is populated when user proceeds on INCONCLUSIVE.
	InconclusiveAcknowledgedBy string
}

// AppendToProgress appends the audit verdict fields to the SPEC's progress.md file.
//
// If the file does not exist, it is created. The 4 required fields per REQ-WAG-003
// are always written: audit_verdict, audit_report, audit_at, auditor_version.
// Optional fields (audit_cache_hit, inconclusive_acknowledged_by) are included when non-zero.
//
// REQ-WAG-003, AC-WAG-03
func AppendToProgress(progressPath string, result *AuditResult) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "- audit_verdict: %s\n", result.Verdict)
	fmt.Fprintf(&sb, "- audit_report: %s\n", result.ReportPath)
	fmt.Fprintf(&sb, "- audit_at: %s\n", result.AuditAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(&sb, "- auditor_version: %s\n", result.AuditorVersion)

	if result.CacheHit {
		sb.WriteString("- audit_cache_hit: true\n")
		fmt.Fprintf(&sb, "- cached_audit_at: %s\n", result.CachedAuditAt.UTC().Format(time.RFC3339))
	}

	if result.InconclusiveAcknowledgedBy != "" {
		fmt.Fprintf(&sb, "- inconclusive_acknowledged_by: %s\n", result.InconclusiveAcknowledgedBy)
	}

	f, err := os.OpenFile(progressPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open progress.md %q: %w", progressPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(sb.String()); err != nil {
		return fmt.Errorf("write progress.md: %w", err)
	}

	return nil
}
