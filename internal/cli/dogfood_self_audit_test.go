//go:build integration

// Package cli provides the dogfood self-audit integration test for SPEC-WF-AUDIT-GATE-001.
//
// AC-WAG-11: the SPEC must pass all plan-auditor must-pass criteria when audited
// by a mock harness that verifies structural completeness.
//
// Run with: go test -tags=integration -race ./internal/cli/... -run TestSelfAuditPassesOnOwnSpec
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SelfAuditSpec represents the structural must-pass criteria evaluated by plan-auditor.
// This is a lightweight mock of plan-auditor's must-pass checks.
type SelfAuditSpec struct {
	// EARSRequirementsExist: spec.md must contain at least 1 EARS-format requirement.
	// Pattern: "REQ-WAG-" in spec.md
	EARSRequirementsExist bool

	// AcceptanceCriteriaExist: acceptance.md must contain at least 1 AC.
	// Pattern: "AC-WAG-" in acceptance.md
	AcceptanceCriteriaExist bool

	// ExclusionSectionPresent: spec.md must have an Out of Scope or exclusion section.
	// Pattern: "Out of Scope" or "4.2" in spec.md
	ExclusionSectionPresent bool

	// FrontmatterSchemaValid: spec.md must have all 9 required frontmatter fields.
	// Fields: id, version, status, created_at, updated_at, author, priority, labels, issue_number
	FrontmatterSchemaValid bool
}

// auditSpec performs a structural audit of the SPEC directory.
// Returns a SelfAuditSpec with boolean results for each must-pass criterion.
func auditSpec(specDir string) (SelfAuditSpec, error) {
	var result SelfAuditSpec

	// Read spec.md.
	specPath := filepath.Join(specDir, "spec.md")
	specContent, err := os.ReadFile(specPath)
	if err != nil {
		return result, err
	}
	spec := string(specContent)

	// Read acceptance.md.
	acPath := filepath.Join(specDir, "acceptance.md")
	acContent, _ := os.ReadFile(acPath)
	ac := string(acContent)

	// Must-pass 1: EARS requirements exist.
	result.EARSRequirementsExist = strings.Contains(spec, "REQ-WAG-")

	// Must-pass 2: acceptance criteria exist.
	result.AcceptanceCriteriaExist = strings.Contains(ac, "AC-WAG-")

	// Must-pass 3: exclusion section present.
	result.ExclusionSectionPresent = strings.Contains(spec, "Out of Scope") ||
		strings.Contains(spec, "4.2") ||
		strings.Contains(spec, "비변경")

	// Must-pass 4: frontmatter schema valid (check all 9 required fields).
	requiredFields := []string{
		"id:", "version:", "status:", "created_at:", "updated_at:",
		"author:", "priority:", "labels:", "issue_number:",
	}
	allFieldsPresent := true
	for _, field := range requiredFields {
		if !strings.Contains(spec, field) {
			allFieldsPresent = false
			break
		}
	}
	result.FrontmatterSchemaValid = allFieldsPresent

	return result, nil
}

// TestSelfAuditPassesOnOwnSpec verifies AC-WAG-11:
// SPEC-WF-AUDIT-GATE-001 itself passes all 4 must-pass criteria.
//
// Given: .moai/specs/SPEC-WF-AUDIT-GATE-001/ with all 4 artifact files
// When: plan-auditor mock harness audits the SPEC
// Then: all 4 must-pass criteria PASS, verdict=PASS
//
// AC: AC-WAG-11
func TestSelfAuditPassesOnOwnSpec(t *testing.T) {
	// Find the SPEC directory relative to the test file.
	// We walk up from the package directory to find the repo root.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	// Walk up to find .moai/specs/SPEC-WF-AUDIT-GATE-001/
	specDir := ""
	dir := wd
	for i := 0; i < 5; i++ {
		candidate := filepath.Join(dir, ".moai", "specs", "SPEC-WF-AUDIT-GATE-001")
		if _, err := os.Stat(candidate); err == nil {
			specDir = candidate
			break
		}
		dir = filepath.Dir(dir)
	}

	if specDir == "" {
		t.Skip("SPEC-WF-AUDIT-GATE-001 directory not found — skipping dogfood test")
	}

	t.Logf("Auditing SPEC at: %s", specDir)

	result, err := auditSpec(specDir)
	if err != nil {
		t.Fatalf("auditSpec: %v", err)
	}

	// Evaluate all 4 must-pass criteria.
	if !result.EARSRequirementsExist {
		t.Error("Must-pass FAIL: spec.md has no EARS requirements (REQ-WAG-* pattern) — AC-WAG-11")
	}
	if !result.AcceptanceCriteriaExist {
		t.Error("Must-pass FAIL: acceptance.md has no acceptance criteria (AC-WAG-* pattern) — AC-WAG-11")
	}
	if !result.ExclusionSectionPresent {
		t.Error("Must-pass FAIL: spec.md has no Out of Scope / exclusion section — AC-WAG-11")
	}
	if !result.FrontmatterSchemaValid {
		t.Error("Must-pass FAIL: spec.md frontmatter is missing required fields — AC-WAG-11")
	}

	// All must-pass criteria passed.
	if t.Failed() {
		t.Fatalf("Dogfood verdict: FAIL — SPEC-WF-AUDIT-GATE-001 does not meet its own must-pass criteria")
	}

	t.Log("Dogfood verdict: PASS — all 4 must-pass criteria satisfied (AC-WAG-11)")
}
