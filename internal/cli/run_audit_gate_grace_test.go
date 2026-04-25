//go:build integration

// Package cli provides integration tests for the Plan Audit Gate grace window behavior.
//
// AC-WAG-08: FAIL during grace window emits warning only; FAIL after grace blocks.
// Run with: go test -tags=integration -race ./internal/cli/... -run TestGraceWindow
package cli

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/runtime"
)

// TestGraceWindowWarnOnlyMode verifies AC-WAG-08:
// FAIL during grace window returns FAIL_WARNED (not blocking).
//
// Given: merge timestamp T0, current time T0+3days (within 7-day grace window)
// When: plan-auditor returns FAIL
// Then: result.Verdict == FAIL_WARNED, GraceWindowActive == true, D-4 remaining
//
// AC: AC-WAG-08
// REQ: REQ-WAG-002 (variant)
func TestGraceWindowWarnOnlyMode(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-FAIL-001")

	t0 := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	nowInGrace := t0.Add(3 * 24 * time.Hour) // T0+3days, D-4 remaining

	gate := &runtime.GateConfig{
		SpecID:     "SPEC-DUMMY-FAIL-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    &deterministicAuditor{verdict: runtime.VerdictFail},
		Cache:      runtime.NewInMemoryCache(),
		Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:      runtime.FakeClock{FixedTime: nowInGrace},
		UserName:   "test-user",
		T0:         t0,
	}

	result, _ := gate.Invoke(context.Background())

	if result.Verdict != runtime.VerdictFailWarned {
		t.Errorf("Verdict = %q, want FAIL_WARNED during grace window (AC-WAG-08)", result.Verdict)
	}
	if !result.GraceWindowActive {
		t.Error("GraceWindowActive = false, want true (AC-WAG-08)")
	}
	if result.GraceWindowRemainingDays != 4 {
		t.Errorf("GraceWindowRemainingDays = %d, want 4 (AC-WAG-08)", result.GraceWindowRemainingDays)
	}
}

// TestGraceWindowExpiryRevertsToBlockingMode verifies AC-WAG-08:
// FAIL after grace window returns FAIL (blocking mode).
//
// Given: merge timestamp T0, current time T0+8days (past 7-day grace window)
// When: plan-auditor returns FAIL
// Then: result.Verdict == FAIL, GraceWindowActive == false
//
// AC: AC-WAG-08
// REQ: REQ-WAG-002
func TestGraceWindowExpiryRevertsToBlockingMode(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	specDir := fixtureDir(t, "SPEC-DUMMY-FAIL-001")

	t0 := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	nowAfterGrace := t0.Add(8 * 24 * time.Hour) // T0+8days, grace expired

	gate := &runtime.GateConfig{
		SpecID:     "SPEC-DUMMY-FAIL-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    &deterministicAuditor{verdict: runtime.VerdictFail},
		Cache:      runtime.NewInMemoryCache(),
		Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
		Clock:      runtime.FakeClock{FixedTime: nowAfterGrace},
		UserName:   "test-user",
		T0:         t0,
	}

	result, _ := gate.Invoke(context.Background())

	if result.Verdict != runtime.VerdictFail {
		t.Errorf("Verdict = %q, want FAIL after grace expiry (AC-WAG-08)", result.Verdict)
	}
	if result.GraceWindowActive {
		t.Error("GraceWindowActive = true, want false after grace window expiry (AC-WAG-08)")
	}
}

// TestGraceWindowCountdownAccurate verifies D-N countdown is accurate.
// AC: AC-WAG-08 (precision of countdown)
func TestGraceWindowCountdownAccurate(t *testing.T) {
	t.Parallel()

	t0 := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		daysSinceT0  int
		expectedDays int
	}{
		{0, 7},
		{1, 6},
		{3, 4},
		{6, 1},
	}

	for _, tc := range testCases {
		projectDir := t.TempDir()
		specDir := fixtureDir(t, "SPEC-DUMMY-FAIL-001")

		gate := &runtime.GateConfig{
			SpecID:     "SPEC-DUMMY-FAIL-001",
			SpecDir:    specDir,
			ProjectDir: projectDir,
			Auditor:    &deterministicAuditor{verdict: runtime.VerdictFail},
			Cache:      runtime.NewInMemoryCache(),
			Reporter:   runtime.NewFileAuditReporter(projectDir, filepath.Join(projectDir, ".moai", "reports", "plan-audit")),
			Clock:      runtime.FakeClock{FixedTime: t0.Add(time.Duration(tc.daysSinceT0) * 24 * time.Hour)},
			UserName:   "test-user",
			T0:         t0,
		}

		result, _ := gate.Invoke(context.Background())

		if result.GraceWindowRemainingDays != tc.expectedDays {
			t.Errorf("at T0+%dd: GraceWindowRemainingDays = %d, want %d",
				tc.daysSinceT0, result.GraceWindowRemainingDays, tc.expectedDays)
		}
	}
}
