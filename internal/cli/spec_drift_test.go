package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// assertNoDrift는 보고서가 drift를 보고하지 않음을 검증한다 (Count==0 + 모든 record Drifted==false).
//
// SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 M1 (mechanism ④ + D3 record-emission contract):
// progress.md가 없는 fixture SPEC은 H-1 → V2.x → EraFinal()==true로 grandfather-exempt가
// 되어 이제 DriftRecord{Drifted:false}로 PRESERVE된다 (이전: getGitImpliedStatus 오류 →
// drop). 따라서 "records 0개" 대신 "Count 0 + 모든 record Drifted==false"가 no-drift의
// 올바른 검증 기준이다.
func assertNoDrift(t *testing.T, report *spec.DriftReport) {
	t.Helper()
	if report.Count != 0 {
		t.Errorf("Expected Count=0 (no drift), got %d", report.Count)
	}
	for _, r := range report.Records {
		if r.Drifted {
			t.Errorf("Expected no drifted records, but %s is Drifted=true", r.SPECID)
		}
	}
}

// TestSpecDriftCmd_NoDrift tests the drift command with no drift
func TestSpecDriftCmd_NoDrift(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test SPEC with matching status
	createTestSPEC(t, tmpDir, "SPEC-TEST-001", "implemented")

	report, err := spec.DetectDrift(tmpDir)
	if err != nil {
		t.Fatalf("DetectDrift failed: %v", err)
	}

	// No progress.md → V2.x grandfather-exempt → record PRESERVED with Drifted=false (D3).
	assertNoDrift(t, report)
}

// TestSpecDriftCmd_WithDrift tests the drift command with drift detected
func TestSpecDriftCmd_WithDrift(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test SPEC with mismatched status
	createTestSPEC(t, tmpDir, "SPEC-TEST-002", "draft")

	report, err := spec.DetectDrift(tmpDir)
	if err != nil {
		t.Fatalf("DetectDrift failed: %v", err)
	}

	// No progress.md → V2.x grandfather-exempt → no drift reported even with mismatched status (D3).
	assertNoDrift(t, report)
}

// TestSpecDriftCmd_MultipleSPECs tests drift detection across multiple SPECs
func TestSpecDriftCmd_MultipleSPECs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple SPECs
	createTestSPEC(t, tmpDir, "SPEC-TEST-003", "completed")
	createTestSPEC(t, tmpDir, "SPEC-TEST-004", "in-progress")
	createTestSPEC(t, tmpDir, "SPEC-TEST-005", "planned")

	report, err := spec.DetectDrift(tmpDir)
	if err != nil {
		t.Fatalf("DetectDrift failed: %v", err)
	}

	// All three lack progress.md → V2.x grandfather-exempt → records PRESERVED, Drifted=false (D3).
	assertNoDrift(t, report)
}

// TestSpecDriftCmd_NoSpecsDirectory tests behavior when specs directory doesn't exist
func TestSpecDriftCmd_NoSpecsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create .moai/specs directory

	report, err := spec.DetectDrift(tmpDir)
	if err != nil {
		t.Fatalf("DetectDrift failed: %v", err)
	}

	if len(report.Records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(report.Records))
	}

	if report.Count != 0 {
		t.Errorf("Expected Count=0, got %d", report.Count)
	}
}

// Helper function to create a test SPEC file
func createTestSPEC(t *testing.T, baseDir, specID, status string) {
	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("Failed to create spec directory: %v", err)
	}

	specContent := `---
id: ` + specID + `
title: Test SPEC
version: 0.1.0
status: ` + status + `
created: 2026-01-01
updated: 2026-01-01
author: Test
priority: P1
phase: run
module: TEST
lifecycle: active
tags: test
---

# Test SPEC

This is a test SPEC.
`

	specPath := filepath.Join(specDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("Failed to write spec.md: %v", err)
	}
}

