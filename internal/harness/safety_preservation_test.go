// Package harness — 5-Layer Safety Architecture preservation tests (T-C5).
//
// This file holds the architectural assertion tests for SPEC-V3R4-HARNESS-002 Wave C.
// The two tests verify architectural invariants rather than code:
//  1. constitution.md §5 Safety Architecture must contain exactly 5 layer names.
//  2. The frozenPrefixes slice must contain exactly 4 canonical entries.
//
// REQ-HRN-FND-006: guarantee immutability of the FROZEN path protection list.
// REQ-HRN-OBS-002: guarantee immutability of the 5-Layer Safety Architecture.
package harness

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// findProjectRoot detects the project root from the test file location.
// Go tests run from the package directory, so the root is N levels up.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	// Detect the project root from the test file's runtime path
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 실패")
	}
	// internal/harness/safety_preservation_test.go → ../../ = project root
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("filepath.Abs 실패: %v", err)
	}
	return abs
}

// TestSafetyArchitecture_LayerCount verifies that constitution.md §5 contains
// exactly the 5 canonical Safety Architecture layer names.
//
// Verified layers (Layer N: Name pattern):
//   - Layer 1: Frozen Guard
//   - Layer 2: Canary Check
//   - Layer 3: Contradiction Detector
//   - Layer 4: Rate Limiter
//   - Layer 5: Human Oversight
//
// This test verifies via string search that constitution.md has not been altered.
func TestSafetyArchitecture_LayerCount(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRoot(t)
	constitutionPath := filepath.Join(projectRoot, ".claude", "rules", "moai", "design", "constitution.md")

	data, err := os.ReadFile(constitutionPath)
	if err != nil {
		t.Fatalf("constitution.md 읽기 실패 (%s): %v", constitutionPath, err)
	}
	body := string(data)

	// Canonical layer names (per constitution.md §5 Safety Architecture)
	wantLayers := []string{
		"Frozen Guard",
		"Canary Check",
		"Contradiction Detector",
		"Rate Limiter",
		"Human Oversight",
	}

	for _, layerName := range wantLayers {
		if !strings.Contains(body, layerName) {
			t.Errorf("constitution.md §5 missing layer name %q — possible 5-Layer structure damage", layerName)
		}
	}

	// Total layer count: number of "### Layer N:" patterns
	layerCount := strings.Count(body, "### Layer ")
	if layerCount != 5 {
		t.Errorf("constitution.md §5 '### Layer ' count: got=%d, want=5", layerCount)
	}
}

// TestSafetyArchitecture_FrozenZoneUnchanged verifies the frozenPrefixes slice
// contains exactly the 4 canonical entries.
// REQ-HRN-FND-006: immutability of the FROZEN path protection list.
//
// Note: tasks.md T-C5 lists `.moai/project/brand/` as the 4th entry, but the
// current frozen_guard.go implementation includes `.claude/skills/moai/`
// instead. This test verifies actual code state; the discrepancy with
// tasks.md is recorded as an SPEC-V3R4-HARNESS-002 implementation note.
// Future changes must update this test together with the code.
func TestSafetyArchitecture_FrozenZoneUnchanged(t *testing.T) {
	t.Parallel()

	// Actual entries currently present in frozen_guard.go (order included)
	wantPrefixes := []string{
		".claude/agents/moai/",
		".claude/skills/moai-",
		".claude/skills/moai/",
		".claude/rules/moai/",
	}

	// Verify entry count
	if len(frozenPrefixes) != len(wantPrefixes) {
		t.Errorf("frozenPrefixes entry count: got=%d, want=%d", len(frozenPrefixes), len(wantPrefixes))
		t.Logf("actual entries: %v", frozenPrefixes)
		return
	}

	// Verify each entry order and value
	for i, want := range wantPrefixes {
		if frozenPrefixes[i] != want {
			t.Errorf("frozenPrefixes[%d]: got=%q, want=%q", i, frozenPrefixes[i], want)
		}
	}

	// Integration check: FROZEN paths must actually be blocked.
	// The constitution.md path must be blocked by the .claude/rules/moai/ prefix.
	constitutionPath := ".claude/rules/moai/design/constitution.md"
	_, err := IsAllowedPath(constitutionPath)
	if err == nil {
		t.Errorf("IsAllowedPath(%q) must return FrozenViolationError", constitutionPath)
	} else {
		var frozenErr *FrozenViolationError
		if !isFrozenViolationError(err, &frozenErr) {
			t.Errorf("IsAllowedPath(%q) error type: got=%T, want=*FrozenViolationError", constitutionPath, err)
		}
	}
}

// isFrozenViolationError checks whether the error is of type *FrozenViolationError.
func isFrozenViolationError(err error, out **FrozenViolationError) bool {
	if fve, ok := err.(*FrozenViolationError); ok {
		if out != nil {
			*out = fve
		}
		return true
	}
	return false
}
