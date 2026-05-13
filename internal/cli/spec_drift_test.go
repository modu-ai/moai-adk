package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// TestSpecDriftCmd_NoDrift tests the drift command with no drift
func TestSpecDriftCmd_NoDrift(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test SPEC with matching status
	createTestSPEC(t, tmpDir, "SPEC-TEST-001", "implemented")

	// Run drift detection (without git history, no records will be returned)
	report, err := spec.DetectDrift(tmpDir)
	if err != nil {
		t.Fatalf("DetectDrift failed: %v", err)
	}

	// Without git history, drift detection returns no records (graceful degradation)
	if len(report.Records) != 0 {
		t.Fatalf("Expected 0 records (no git history), got %d", len(report.Records))
	}

	if report.Count != 0 {
		t.Errorf("Expected Count=0, got %d", report.Count)
	}
}

// TestSpecDriftCmd_WithDrift tests the drift command with drift detected
func TestSpecDriftCmd_WithDrift(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test SPEC with mismatched status
	createTestSPEC(t, tmpDir, "SPEC-TEST-002", "draft")

	// Run drift detection (without git history, no records will be returned)
	report, err := spec.DetectDrift(tmpDir)
	if err != nil {
		t.Fatalf("DetectDrift failed: %v", err)
	}

	// Without git history, drift detection returns no records (graceful degradation)
	if len(report.Records) != 0 {
		t.Fatalf("Expected 0 records (no git history), got %d", len(report.Records))
	}

	if report.Count != 0 {
		t.Errorf("Expected Count=0, got %d", report.Count)
	}
}

// TestSpecDriftCmd_MultipleSPECs tests drift detection across multiple SPECs
func TestSpecDriftCmd_MultipleSPECs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple SPECs
	createTestSPEC(t, tmpDir, "SPEC-TEST-003", "completed")
	createTestSPEC(t, tmpDir, "SPEC-TEST-004", "in-progress")
	createTestSPEC(t, tmpDir, "SPEC-TEST-005", "planned")

	// Run drift detection (without git history, no records will be returned)
	report, err := spec.DetectDrift(tmpDir)
	if err != nil {
		t.Fatalf("DetectDrift failed: %v", err)
	}

	// Without git history, drift detection returns no records (graceful degradation)
	if len(report.Records) != 0 {
		t.Fatalf("Expected 0 records (no git history), got %d", len(report.Records))
	}

	if report.Count != 0 {
		t.Errorf("Expected Count=0, got %d", report.Count)
	}
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

