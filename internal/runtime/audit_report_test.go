package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestReportFilePathFormat verifies the date-stamped report file path format. AC-WAG-04
func TestReportFilePathFormat(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	reportDir := filepath.Join(projectDir, ".moai", "reports", "plan-audit")
	reporter := &FileAuditReporter{
		ReportDir:  reportDir,
		ProjectDir: projectDir,
		Clock:      FakeClock{FixedTime: time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)},
	}

	path, err := reporter.reportFilePath("SPEC-WAG-001", reporter.Clock.Now())
	if err != nil {
		t.Fatalf("reportFilePath: %v", err)
	}

	wantSuffix := "SPEC-WAG-001-2026-04-25.md"
	if !strings.HasSuffix(path, wantSuffix) {
		t.Errorf("reportFilePath = %q, want suffix %q", path, wantSuffix)
	}

	// Verify path is inside reportDir.
	cleanDir := filepath.Clean(reportDir)
	if !strings.HasPrefix(filepath.Clean(path), cleanDir) {
		t.Errorf("reportFilePath %q is outside reportDir %q", path, reportDir)
	}
}

// TestAppendMultipleAuditRunsSameDay verifies multiple runs append correctly. AC-WAG-04
func TestAppendMultipleAuditRunsSameDay(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	reportDir := filepath.Join(projectDir, ".moai", "reports", "plan-audit")

	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	reporter := &FileAuditReporter{
		ReportDir:  reportDir,
		ProjectDir: projectDir,
		Clock:      FakeClock{FixedTime: now},
	}

	specID := "SPEC-MULTI-001"
	verdicts := []Verdict{VerdictPass, VerdictFail, VerdictBypassed}

	for i, v := range verdicts {
		result := &AuditResult{
			Verdict:        v,
			AuditAt:        now.Add(time.Duration(i) * time.Hour),
			AuditorVersion: "plan-auditor/v1",
			BypassReason:   "test bypass",
			BypassUser:     "test-user",
		}
		if err := reporter.AppendRun(specID, result); err != nil {
			t.Fatalf("AppendRun %d: %v", i+1, err)
		}
	}

	// Verify the report file has 3 run sections.
	filePath := filepath.Join(reportDir, "SPEC-MULTI-001-2026-04-25.md")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	runCount := strings.Count(content, "## Audit Run ")
	if runCount != 3 {
		t.Errorf("report has %d audit run sections, want 3", runCount)
	}

	// Verify all 3 verdicts are present.
	for _, v := range verdicts {
		if !strings.Contains(content, string(v)) {
			t.Errorf("report missing verdict %q", v)
		}
	}
}

// TestReportSecurityFilepathClean verifies path traversal prevention. Secured gate.
func TestReportSecurityFilepathClean(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	reportDir := filepath.Join(projectDir, ".moai", "reports", "plan-audit")

	reporter := &FileAuditReporter{
		ReportDir:  reportDir,
		ProjectDir: projectDir,
		Clock:      FakeClock{FixedTime: time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC)},
	}

	// Attempt path traversal via SPEC-ID containing ".."
	maliciousSpecID := "../../etc/passwd"
	_, err := reporter.reportFilePath(maliciousSpecID, reporter.Clock.Now())
	if err == nil {
		t.Error("expected path traversal error for malicious SPEC-ID, got nil")
	}
}

// TestAppendToProgressWritesFourFields verifies 4 required progress.md fields. AC-WAG-03
func TestAppendToProgressWritesFourFields(t *testing.T) {
	t.Parallel()

	progressPath := filepath.Join(t.TempDir(), "progress.md")
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

	result := &AuditResult{
		Verdict:        VerdictPass,
		ReportPath:     ".moai/reports/plan-audit/SPEC-TEST-001-2026-04-25.md",
		AuditAt:        now,
		AuditorVersion: "plan-auditor/v1",
	}

	if err := AppendToProgress(progressPath, result); err != nil {
		t.Fatalf("AppendToProgress: %v", err)
	}

	data, err := os.ReadFile(progressPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)
	requiredFields := []string{
		"audit_verdict: PASS",
		"audit_report:",
		"audit_at:",
		"auditor_version:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(content, field) {
			t.Errorf("progress.md missing required field: %q\ncontent:\n%s", field, content)
		}
	}
}

// TestAppendToProgressIncludesCacheFields verifies cache hit fields. AC-WAG-09
func TestAppendToProgressIncludesCacheFields(t *testing.T) {
	t.Parallel()

	progressPath := filepath.Join(t.TempDir(), "progress.md")
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	cacheTime := now.Add(-1 * time.Hour)

	result := &AuditResult{
		Verdict:        VerdictPass,
		AuditAt:        now,
		CacheHit:       true,
		CachedAuditAt:  cacheTime,
		AuditorVersion: "plan-auditor/v1",
	}

	if err := AppendToProgress(progressPath, result); err != nil {
		t.Fatalf("AppendToProgress: %v", err)
	}

	data, _ := os.ReadFile(progressPath)
	content := string(data)

	if !strings.Contains(content, "audit_cache_hit: true") {
		t.Error("progress.md missing audit_cache_hit: true")
	}
	if !strings.Contains(content, "cached_audit_at:") {
		t.Error("progress.md missing cached_audit_at field")
	}
}
