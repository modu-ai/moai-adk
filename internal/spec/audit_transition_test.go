// audit_transition_test.go — TDD coverage for the status-transition-audit.log
// consumer (SPEC-OBSERVE-HYGIENE-001 M1, REQ-OBH-001, AC-OBH-001).
//
// The log at <baseDir>/.moai/logs/status-transition-audit.log is produced by
// .claude/hooks/moai/status-transition-ownership.sh. Each line records a
// Write/Edit/MultiEdit on a SPEC artifact together with the status value
// captured at write time. moai spec audit consumes the log as a cross-check
// input: a status value that is not in the canonical 8-value enum (and not the
// recognized "<file absent — Write creating new>" sentinel) is a write-site
// anomaly the git-history lint (OwnershipTransitionRule) structurally cannot
// see. The finding is INFO severity (non-breaking).
//
// Graceful contract (EC-1): absent, corrupt, or unparseable log degrades to
// zero findings and never an error.
package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTransitionLog writes the given log body to
// <tempDir>/.moai/logs/status-transition-audit.log and returns tempDir.
func writeTransitionLog(t *testing.T, body string) string {
	t.Helper()
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, ".moai", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(logDir, "status-transition-audit.log")
	if err := os.WriteFile(logPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return tempDir
}

// logLine is a convenience builder producing one canonical-format log line.
// Format: "<ISO-ts> [status-transition-ownership] <Tool> <Path> status=<status>"
func logLine(ts, tool, path, status string) string {
	return ts + " [status-transition-ownership] " + tool + " " + path + " status=" + status + "\n"
}

// --- parseStatusTransitionLog unit tests ---

func TestParseStatusTransitionLog_Nominal(t *testing.T) {
	t.Parallel()
	body := logLine("2026-07-09T08:00:00Z", "Edit", "/p/.moai/specs/SPEC-FOO-001/spec.md", "draft") +
		logLine("2026-07-09T08:01:00Z", "Write", "/p/.moai/specs/SPEC-BAR-002/plan.md", "completed")
	dir := writeTransitionLog(t, body)

	entries, stats, err := parseStatusTransitionLog(filepath.Join(dir, ".moai", "logs", "status-transition-audit.log"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Tool != "Edit" || entries[0].RawStatus != "draft" {
		t.Errorf("entry[0] = %+v", entries[0])
	}
	if stats.Parsed != 2 || stats.SkippedUnparseable != 0 {
		t.Errorf("stats = %+v, want Parsed=2 SkippedUnparseable=0", stats)
	}
}

// EC-1 — format drift: unknown line shapes are skipped + counted, never error.
func TestParseStatusTransitionLog_FormatDrift(t *testing.T) {
	t.Parallel()
	body := logLine("2026-07-09T08:00:00Z", "Edit", "/p/.moai/specs/SPEC-FOO-001/spec.md", "draft") +
		"this line is garbage and does not match the format\n" +
		"2026-07-09T08 another malformed line\n" +
		logLine("2026-07-09T08:01:00Z", "Write", "/p/.moai/specs/SPEC-BAR-002/spec.md", "completed")
	dir := writeTransitionLog(t, body)

	entries, stats, err := parseStatusTransitionLog(filepath.Join(dir, ".moai", "logs", "status-transition-audit.log"))
	if err != nil {
		t.Fatalf("format drift must not error, got %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 parsed entries (2 garbage skipped), got %d", len(entries))
	}
	if stats.SkippedUnparseable != 2 {
		t.Errorf("SkippedUnparseable = %d, want 2", stats.SkippedUnparseable)
	}
}

// Absent log → graceful no-op (zero entries, nil error).
func TestParseStatusTransitionLog_Absent(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, ".moai", "logs", "status-transition-audit.log")

	entries, stats, err := parseStatusTransitionLog(logPath)
	if err != nil {
		t.Fatalf("absent log must not error, got %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("absent log → 0 entries, got %d", len(entries))
	}
	if !stats.LogAbsent {
		t.Errorf("LogAbsent = false, want true")
	}
}

// --- crossCheckTransitionLog correlation tests ---

// AC-OBH-001 — anomalous (non-canonical) status yields an INFO finding.
func TestCrossCheckTransitionLog_AnomalousStatus(t *testing.T) {
	t.Parallel()
	body := logLine("2026-07-09T08:00:00Z", "Edit",
		"/proj/.moai/specs/SPEC-FOO-001/spec.md", "invalid-garbage-state")
	dir := writeTransitionLog(t, body)

	findings := crossCheckTransitionLog(dir, AuditOptions{})
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Severity != "INFO" {
		t.Errorf("Severity = %q, want INFO", f.Severity)
	}
	if f.FindingType != FindingStatusTransitionLogAnomaly {
		t.Errorf("FindingType = %q, want %q", f.FindingType, FindingStatusTransitionLogAnomaly)
	}
	if f.SpecID != "SPEC-FOO-001" {
		t.Errorf("SpecID = %q, want SPEC-FOO-001", f.SpecID)
	}
	if raw, _ := f.Details["raw_status"].(string); raw != "invalid-garbage-state" {
		t.Errorf("raw_status detail = %q", raw)
	}
}

// Canonical status → no finding.
func TestCrossCheckTransitionLog_CanonicalStatus(t *testing.T) {
	t.Parallel()
	for _, st := range []string{"draft", "in-progress", "implemented", "completed", "superseded"} {
		body := logLine("2026-07-09T08:00:00Z", "Edit",
			"/proj/.moai/specs/SPEC-FOO-001/spec.md", st)
		dir := writeTransitionLog(t, body)
		findings := crossCheckTransitionLog(dir, AuditOptions{})
		if len(findings) != 0 {
			t.Errorf("canonical status %q yielded %d findings (want 0)", st, len(findings))
		}
	}
}

// Leading-whitespace canonical status (observed format drift: "status= implemented")
// trims cleanly → no finding.
func TestCrossCheckTransitionLog_LeadingSpaceCanonical(t *testing.T) {
	t.Parallel()
	body := logLine("2026-07-09T08:00:00Z", "Edit",
		"/proj/.moai/specs/SPEC-FOO-001/spec.md", " implemented")
	dir := writeTransitionLog(t, body)
	findings := crossCheckTransitionLog(dir, AuditOptions{})
	if len(findings) != 0 {
		t.Errorf("leading-space canonical status yielded %d findings (want 0)", len(findings))
	}
}

// The "<file absent — Write creating new>" sentinel (new-file Write) → no finding.
func TestCrossCheckTransitionLog_SentinelNoFinding(t *testing.T) {
	t.Parallel()
	body := logLine("2026-07-09T08:00:00Z", "Write",
		"/proj/.moai/specs/SPEC-FOO-001/spec.md", "<file absent — Write creating new>")
	dir := writeTransitionLog(t, body)
	findings := crossCheckTransitionLog(dir, AuditOptions{})
	if len(findings) != 0 {
		t.Errorf("sentinel yielded %d findings (want 0)", len(findings))
	}
}

// A path with no extractable SPEC-ID → skipped (not a SPEC artifact).
func TestCrossCheckTransitionLog_NoSpecID(t *testing.T) {
	t.Parallel()
	body := logLine("2026-07-09T08:00:00Z", "Edit",
		"/some/random/path/spec.md", "garbage")
	dir := writeTransitionLog(t, body)
	findings := crossCheckTransitionLog(dir, AuditOptions{})
	if len(findings) != 0 {
		t.Errorf("non-SPEC path yielded %d findings (want 0)", len(findings))
	}
}

// Absent log → zero findings.
func TestCrossCheckTransitionLog_AbsentLog(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	findings := crossCheckTransitionLog(tempDir, AuditOptions{})
	if len(findings) != 0 {
		t.Fatalf("absent log → 0 findings, got %d", len(findings))
	}
}

// FilterSpec gates which SPEC's findings surface.
func TestCrossCheckTransitionLog_FilterSpec(t *testing.T) {
	t.Parallel()
	body := logLine("2026-07-09T08:00:00Z", "Edit",
		"/proj/.moai/specs/SPEC-AAA-001/spec.md", "garbage-a") +
		logLine("2026-07-09T08:01:00Z", "Edit",
			"/proj/.moai/specs/SPEC-BBB-002/spec.md", "garbage-b")
	dir := writeTransitionLog(t, body)

	findings := crossCheckTransitionLog(dir, AuditOptions{FilterSpec: "SPEC-AAA-001"})
	if len(findings) != 1 {
		t.Fatalf("FilterSpec=SPEC-AAA-001 → 1 finding, got %d", len(findings))
	}
	if findings[0].SpecID != "SPEC-AAA-001" {
		t.Errorf("SpecID = %q", findings[0].SpecID)
	}
}

// --- Audit() integration test ---

// AC-OBH-001 end-to-end: Audit() with a fixture log carrying an anomalous
// status yields an INFO finding; absent log yields zero findings + exit 0.
func TestAudit_TransitionLogConsumerIntegration(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	// Create one V3R6 SPEC fixture (so TotalSpecs > 0 and the audit is real).
	specDir := filepath.Join(tempDir, ".moai", "specs", "SPEC-V3R6-DEMO-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	specMD := makeSpecMD("SPEC-V3R6-DEMO-001", "in-progress", "V3R6", "2026-07-09")
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatal(err)
	}
	progressMD := "## §E.2 Run-phase\nrun evidence here\n## §E.3\n## §E.4 Sync\nsync_commit_sha:\n"
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressMD), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write a transition log with one anomalous entry referencing this SPEC.
	logDir := filepath.Join(tempDir, ".moai", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	logBody := logLine("2026-07-09T08:00:00Z", "Edit",
		tempDir+"/.moai/specs/SPEC-V3R6-DEMO-001/spec.md", "bogus-status")
	if err := os.WriteFile(filepath.Join(logDir, "status-transition-audit.log"), []byte(logBody), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Audit(AuditOptions{BaseDir: tempDir})
	if err != nil {
		t.Fatalf("Audit error: %v", err)
	}
	found := false
	for _, f := range result.DriftFindings {
		if f.FindingType == FindingStatusTransitionLogAnomaly && f.SpecID == "SPEC-V3R6-DEMO-001" {
			if f.Severity != "INFO" {
				t.Errorf("anomaly finding Severity = %q, want INFO", f.Severity)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("expected a StatusTransitionLogAnomaly finding; got findings: %+v", result.DriftFindings)
	}
}

// AC-OBH-001 graceful path: absent log → Audit() completes with no transition
// findings (the per-SPEC loop findings are unaffected).
func TestAudit_TransitionLogAbsentNoFindings(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	specDir := filepath.Join(tempDir, ".moai", "specs", "SPEC-V3R6-CLEAN-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	specMD := makeSpecMD("SPEC-V3R6-CLEAN-001", "in-progress", "V3R6", "2026-07-09")
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte("## §E.2\n## §E.4\nsync_commit_sha:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// No .moai/logs/ at all.
	result, err := Audit(AuditOptions{BaseDir: tempDir})
	if err != nil {
		t.Fatalf("Audit error: %v", err)
	}
	for _, f := range result.DriftFindings {
		if f.FindingType == FindingStatusTransitionLogAnomaly {
			t.Errorf("absent log should yield no anomaly finding; got %+v", f)
		}
	}
}

// Ensure the finding type constant is stable (JSON consumers depend on it).
func TestStatusTransitionLogFindingTypeStable(t *testing.T) {
	t.Parallel()
	if !strings.HasPrefix(FindingStatusTransitionLogAnomaly, "StatusTransitionLog") {
		t.Errorf("FindingStatusTransitionLogAnomaly = %q", FindingStatusTransitionLogAnomaly)
	}
}
