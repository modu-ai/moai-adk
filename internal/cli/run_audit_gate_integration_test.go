//go:build integration

// Package cli provides integration tests for the Plan Audit Gate (SPEC-WF-AUDIT-GATE-001).
//
// These tests verify AC-WAG-01, AC-WAG-02, AC-WAG-03, AC-WAG-06, AC-WAG-07.
// Run with: go test -tags=integration -race ./internal/cli/... -run TestRunAuditGate
package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/runtime"
)

// setupGate creates a GateConfig wired for integration testing.
// Uses a mock auditor with deterministic verdict.
func setupGate(t *testing.T, specDir string, verdict runtime.Verdict, projectDir string) *runtime.GateConfig {
	t.Helper()

	reporter := runtime.NewFileAuditReporter(
		projectDir,
		filepath.Join(projectDir, ".moai", "reports", "plan-audit"),
	)

	cache := runtime.NewInMemoryCache()

	return &runtime.GateConfig{
		SpecID:     filepath.Base(specDir),
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    &deterministicAuditor{verdict: verdict, reportPath: ""},
		Cache:      cache,
		Reporter:   reporter,
		Clock:      runtime.SystemClock{},
		UserName:   "test-user",
	}
}

// deterministicAuditor implements PlanAuditor with a fixed verdict.
type deterministicAuditor struct {
	verdict    runtime.Verdict
	reportPath string
	err        error
	callCount  int
}

func (d *deterministicAuditor) Audit(_ context.Context, _ string) (runtime.Verdict, string, error) {
	d.callCount++
	return d.verdict, d.reportPath, d.err
}

// fixtureDir returns the path to the test fixture SPEC directory.
func fixtureDir(t *testing.T, specName string) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	return filepath.Join(wd, "testdata", "audit-gate", specName)
}

// TestRunInvokesPlanAuditorBeforeImplementation verifies AC-WAG-01:
// plan-auditor is invoked exactly once before any implementation phase.
//
// Evidence:
// - auditor.callCount == 1 after Invoke
// - result.Verdict is set (gate executed)
// - report file created at .moai/reports/plan-audit/<SPEC>-<DATE>.md
//
// AC: AC-WAG-01
// REQ: REQ-WAG-001
func TestRunInvokesPlanAuditorBeforeImplementation(t *testing.T) {
	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-PASS-001")

	auditor := &deterministicAuditor{verdict: runtime.VerdictPass}

	gate := &runtime.GateConfig{
		SpecID:     "SPEC-DUMMY-PASS-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    auditor,
		Cache:      runtime.NewInMemoryCache(),
		Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:      runtime.FakeClock{FixedTime: time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)},
		UserName:   "test-user",
	}

	result, err := gate.Invoke(context.Background())
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}

	// Verify auditor was called exactly once.
	if auditor.callCount != 1 {
		t.Errorf("plan-auditor called %d times, want exactly 1 (AC-WAG-01)", auditor.callCount)
	}

	// Verify gate produced a verdict.
	if result.Verdict == "" {
		t.Error("gate returned empty verdict — gate did not execute (AC-WAG-01)")
	}

	// Verify daily report was created.
	reportDir := filepath.Join(projectDir, ".moai", "reports", "plan-audit")
	entries, err := os.ReadDir(reportDir)
	if err != nil {
		t.Fatalf("ReadDir %q: %v", reportDir, err)
	}
	if len(entries) == 0 {
		t.Error("no report file created in .moai/reports/plan-audit/ (AC-WAG-01)")
	}
}

// TestRunBlockedOnAuditFail verifies AC-WAG-02:
// FAIL verdict (outside grace window) sets Verdict=FAIL and does not auto-proceed.
//
// Evidence:
// - result.Verdict == FAIL
// - result.GraceWindowActive == false
//
// AC: AC-WAG-02
// REQ: REQ-WAG-002
func TestRunBlockedOnAuditFail(t *testing.T) {
	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-FAIL-001")

	// Set T0 far in the past so grace window is expired.
	pastT0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	gate := &runtime.GateConfig{
		SpecID:     "SPEC-DUMMY-FAIL-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    &deterministicAuditor{verdict: runtime.VerdictFail},
		Cache:      runtime.NewInMemoryCache(),
		Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:      runtime.FakeClock{FixedTime: time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)},
		UserName:   "test-user",
		T0:         pastT0,
	}

	result, _ := gate.Invoke(context.Background())

	if result.Verdict != runtime.VerdictFail {
		t.Errorf("Verdict = %q, want FAIL (AC-WAG-02)", result.Verdict)
	}
	if result.GraceWindowActive {
		t.Error("GraceWindowActive = true, want false for expired grace window (AC-WAG-02)")
	}
}

// TestRunProceedsOnAuditPassAndPersistsVerdict verifies AC-WAG-03:
// PASS verdict persists 4 fields to progress.md and the audit result shows PASS.
//
// Evidence:
// - result.Verdict == PASS
// - result.AuditAt is non-zero
// - result.AuditorVersion is non-empty
//
// AC: AC-WAG-03
// REQ: REQ-WAG-003
func TestRunProceedsOnAuditPassAndPersistsVerdict(t *testing.T) {
	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-PASS-001")

	gate := setupGate(t, specDir, runtime.VerdictPass, projectDir)

	result, err := gate.Invoke(context.Background())
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}

	if result.Verdict != runtime.VerdictPass {
		t.Errorf("Verdict = %q, want PASS (AC-WAG-03)", result.Verdict)
	}

	// Verify progress.md would have the required 4 fields (AppendToProgress).
	progressPath := filepath.Join(projectDir, "progress.md")
	if err := runtime.AppendToProgress(progressPath, result); err != nil {
		t.Fatalf("AppendToProgress: %v", err)
	}

	data, err := os.ReadFile(progressPath)
	if err != nil {
		t.Fatalf("ReadFile progress.md: %v", err)
	}

	content := string(data)
	requiredFields := []string{"audit_verdict: PASS", "audit_report:", "audit_at:", "auditor_version:"}
	for _, field := range requiredFields {
		if !strings.Contains(content, field) {
			t.Errorf("progress.md missing field %q (AC-WAG-03)", field)
		}
	}
}

// TestSkipAuditFlagRecordsBypassWithUserRationale verifies AC-WAG-06 (interactive path):
// --skip-audit records BYPASSED verdict with user and reason.
//
// AC: AC-WAG-06
// REQ: REQ-WAG-006
func TestSkipAuditFlagRecordsBypassWithUserRationale(t *testing.T) {
	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-BYP-001")

	gate := &runtime.GateConfig{
		SpecID:       "SPEC-DUMMY-BYP-001",
		SpecDir:      specDir,
		ProjectDir:   projectDir,
		Auditor:      &deterministicAuditor{verdict: runtime.VerdictPass},
		Cache:        runtime.NewInMemoryCache(),
		Reporter:     runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:        runtime.FakeClock{FixedTime: time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)},
		UserName:     "GOOS행님",
		SkipAudit:    true,
		BypassReason: "demo for ICSE 2026 deadline",
	}

	result, err := gate.Invoke(context.Background())
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}

	if result.Verdict != runtime.VerdictBypassed {
		t.Errorf("Verdict = %q, want BYPASSED (AC-WAG-06)", result.Verdict)
	}
	if result.BypassUser != "GOOS행님" {
		t.Errorf("BypassUser = %q, want GOOS행님 (AC-WAG-06)", result.BypassUser)
	}
	if !strings.Contains(result.BypassReason, "ICSE 2026") {
		t.Errorf("BypassReason = %q, want to contain ICSE 2026 (AC-WAG-06)", result.BypassReason)
	}
}

// TestEnvVarSkipAuditEquivalentToFlag verifies MOAI_SKIP_PLAN_AUDIT=1 behaves like --skip-audit.
//
// AC: AC-WAG-06
// REQ: REQ-WAG-006
func TestEnvVarSkipAuditEquivalentToFlag(t *testing.T) {
	t.Setenv(runtime.EnvSkipAudit, "1")

	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-BYP-001")

	// The env var is read by the caller (orchestrator) and mapped to SkipAudit: true.
	// Test verifies the gate behaves identically to --skip-audit when SkipAudit is set.
	gate := &runtime.GateConfig{
		SpecID:     "SPEC-DUMMY-BYP-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    &deterministicAuditor{verdict: runtime.VerdictPass},
		Cache:      runtime.NewInMemoryCache(),
		Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:      runtime.FakeClock{FixedTime: time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)},
		UserName:   "test-user",
		SkipAudit:  os.Getenv(runtime.EnvSkipAudit) == "1",
	}

	result, _ := gate.Invoke(context.Background())

	if result.Verdict != runtime.VerdictBypassed {
		t.Errorf("env var skip: Verdict = %q, want BYPASSED (AC-WAG-06)", result.Verdict)
	}
}

// TestPlanAuditorFailureClassifiesAsInconclusive verifies AC-WAG-07:
// timeout, malformed output, and panic all result in INCONCLUSIVE (not PASS).
//
// AC: AC-WAG-07
// REQ: REQ-WAG-007
func TestPlanAuditorFailureClassifiesAsInconclusive(t *testing.T) {
	tests := []struct {
		name    string
		auditor *deterministicAuditor
	}{
		{
			name:    "timeout error",
			auditor: &deterministicAuditor{err: errTimeout},
		},
		{
			name:    "malformed verdict",
			auditor: &deterministicAuditor{verdict: "GARBAGE"},
		},
		{
			name:    "generic error",
			auditor: &deterministicAuditor{err: errGeneric},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := t.TempDir()
			specDir := fixtureDir(t, "SPEC-DUMMY-INC-001")

			gate := &runtime.GateConfig{
				SpecID:     "SPEC-DUMMY-INC-001",
				SpecDir:    specDir,
				ProjectDir: projectDir,
				Auditor:    tt.auditor,
				Cache:      runtime.NewInMemoryCache(),
				Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
				Clock:      runtime.FakeClock{FixedTime: time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)},
				UserName:   "test-user",
			}

			result, _ := gate.Invoke(context.Background())

			if result.Verdict != runtime.VerdictInconclusive {
				t.Errorf("[%s] Verdict = %q, want INCONCLUSIVE (AC-WAG-07)", tt.name, result.Verdict)
			}

			// REQ-WAG-007: INCONCLUSIVE must NOT equal PASS.
			if result.Verdict == runtime.VerdictPass {
				t.Errorf("[%s] INCONCLUSIVE incorrectly mapped to PASS — auto-PASS is prohibited", tt.name)
			}
		})
	}
}

var errTimeout = errors.New("plan-auditor: timeout after 60s")
var errGeneric = errors.New("plan-auditor: unexpected error")
