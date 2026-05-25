// Package defs unit tests for DeprecatedPaths enumeration.
//
// @MX:ANCHOR: DeprecatedPaths is the SSOT for v.2.x → v3 cleanup targets;
// see SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 §A.4 Canonical Derivation Table.
// @MX:REASON: External-user cleanup correctness depends on the 43-entry
// total + 9/31/3 category split; any future modification MUST update both
// this test and spec.md §A.4 atomically.
package defs

import (
	"strings"
	"testing"
)

// TestDeprecatedPathsTotalCount asserts the total slice size matches the
// canonical derivation table in SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 §A.4.
//
// After M2 of that SPEC: 9 (Category A) + 31 (Category B) + 3 (Category C)
// = 43 entries total. AC-VVCR-005 references this assertion.
func TestDeprecatedPathsTotalCount(t *testing.T) {
	const want = 43
	got := len(DeprecatedPaths)
	if got != want {
		t.Errorf("len(DeprecatedPaths) = %d, want %d (per spec.md §A.4 Canonical Derivation Table: 9 Category A + 31 Category B + 3 Category C)", got, want)
	}
}

// TestDeprecatedPathsCategorySplit asserts the per-category subtotals match
// the canonical derivation table, classified by DeprecatedSince field.
//
//   - Category A (9 entries): DeprecatedSince == "SPEC-AGENCY-ABSORB-001"
//   - Category B (31 entries): DeprecatedSince == "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"
//   - Category C (3 entries):  DeprecatedSince == "SPEC-V3R6-AGENT-FOLDER-SPLIT-001"
//
// AC-VVCR-005 explicitly requires verifying both the total count and the
// per-category subtotals.
func TestDeprecatedPathsCategorySplit(t *testing.T) {
	const (
		wantCategoryA = 9  // SPEC-AGENCY-ABSORB-001
		wantCategoryB = 31 // SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
		wantCategoryC = 3  // SPEC-V3R6-AGENT-FOLDER-SPLIT-001
	)

	counts := map[string]int{}
	for _, entry := range DeprecatedPaths {
		counts[entry.DeprecatedSince]++
	}

	if got := counts["SPEC-AGENCY-ABSORB-001"]; got != wantCategoryA {
		t.Errorf("Category A (DeprecatedSince=SPEC-AGENCY-ABSORB-001): got %d entries, want %d", got, wantCategoryA)
	}
	if got := counts["SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"]; got != wantCategoryB {
		t.Errorf("Category B (DeprecatedSince=SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001): got %d entries, want %d", got, wantCategoryB)
	}
	if got := counts["SPEC-V3R6-AGENT-FOLDER-SPLIT-001"]; got != wantCategoryC {
		t.Errorf("Category C (DeprecatedSince=SPEC-V3R6-AGENT-FOLDER-SPLIT-001): got %d entries, want %d", got, wantCategoryC)
	}
}

// TestDeprecatedPathsRequiredFields asserts every entry has all 4 required
// fields populated (Path, DeprecatedSince, DeprecatedBy, RemovalSchedule).
func TestDeprecatedPathsRequiredFields(t *testing.T) {
	for i, entry := range DeprecatedPaths {
		if entry.Path == "" {
			t.Errorf("DeprecatedPaths[%d].Path is empty", i)
		}
		if entry.DeprecatedSince == "" {
			t.Errorf("DeprecatedPaths[%d] (%q).DeprecatedSince is empty", i, entry.Path)
		}
		if entry.DeprecatedBy == "" {
			t.Errorf("DeprecatedPaths[%d] (%q).DeprecatedBy is empty", i, entry.Path)
		}
		if entry.RemovalSchedule == "" {
			t.Errorf("DeprecatedPaths[%d] (%q).RemovalSchedule is empty", i, entry.Path)
		}
	}
}

// TestDeprecatedPathsCategoryBExpectedEntries asserts the 31 Category B entries
// match the exact enumeration in spec.md §A.4.
//
// This test catches accidental additions/removals that would drift away from
// the canonical derivation table without simultaneous spec.md update.
func TestDeprecatedPathsCategoryBExpectedEntries(t *testing.T) {
	wantCategoryB := []string{
		// v2 directories
		".agency",
		".agency.archived",
		// agency agents (FLAT)
		".claude/agents/moai/planner.md",
		".claude/agents/moai/designer.md",
		".claude/agents/moai/builder.md",
		".claude/agents/moai/evaluator.md",
		// retired manager agents (FLAT)
		".claude/agents/moai/manager-strategy.md",
		".claude/agents/moai/manager-quality.md",
		".claude/agents/moai/manager-brain.md",
		".claude/agents/moai/manager-project.md",
		// retired meta agents (FLAT)
		".claude/agents/moai/claude-code-guide.md",
		".claude/agents/moai/researcher.md",
		// retired expert agents (FLAT)
		".claude/agents/moai/expert-backend.md",
		".claude/agents/moai/expert-frontend.md",
		".claude/agents/moai/expert-security.md",
		".claude/agents/moai/expert-devops.md",
		".claude/agents/moai/expert-performance.md",
		".claude/agents/moai/expert-refactoring.md",
		// deprecated config yaml files
		".moai/config/sections/design.yaml",
		".moai/config/sections/db.yaml",
		".moai/config/sections/gate.yaml",
		".moai/config/sections/github-actions.yaml",
		".moai/config/sections/memo.yaml",
		// design skill directories
		".claude/skills/moai-domain-brand-design",
		".claude/skills/moai-domain-copywriting",
		".claude/skills/moai-domain-design-handoff",
		// design workflow skill directories
		".claude/skills/moai-workflow-design",
		".claude/skills/moai-workflow-gan-loop",
		// design rule directory
		".claude/rules/moai/design",
		// brand + db directories
		".moai/project/brand",
		".moai/db",
	}

	got := map[string]bool{}
	for _, entry := range DeprecatedPaths {
		if entry.DeprecatedSince == "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001" {
			got[entry.Path] = true
		}
	}

	if len(got) != len(wantCategoryB) {
		t.Errorf("Category B path count: got %d, want %d", len(got), len(wantCategoryB))
	}

	for _, wantPath := range wantCategoryB {
		if !got[wantPath] {
			t.Errorf("Category B missing entry: %q", wantPath)
		}
	}
}

// TestDeprecatedPathsCategoryCExpectedEntries asserts the 3 Category C entries
// (rc1-stage staging artifacts) match the spec.md §A.4 enumeration.
//
// These directories never existed in v.2.x; they were introduced by commit
// 1bd083725 (SPEC-V3R6-AGENT-FOLDER-SPLIT-001) and are removed by the M2a
// FLAT layout restoration. They are included in DeprecatedPaths to cover
// rc1-stage early adopters who cloned between 1bd083725 and v3.0.0-rc2.
func TestDeprecatedPathsCategoryCExpectedEntries(t *testing.T) {
	wantCategoryC := []string{
		".claude/agents/core",
		".claude/agents/expert",
		".claude/agents/meta",
	}

	got := map[string]bool{}
	for _, entry := range DeprecatedPaths {
		if entry.DeprecatedSince == "SPEC-V3R6-AGENT-FOLDER-SPLIT-001" {
			got[entry.Path] = true
		}
	}

	if len(got) != len(wantCategoryC) {
		t.Errorf("Category C path count: got %d, want %d", len(got), len(wantCategoryC))
	}

	for _, wantPath := range wantCategoryC {
		if !got[wantPath] {
			t.Errorf("Category C missing entry: %q", wantPath)
		}
	}
}

// TestDeprecatedPathsDeprecatedByConsistency asserts the DeprecatedBy field
// follows the SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 ownership convention:
//
//   - Category A pre-existing entries: DeprecatedBy unchanged (was
//     "SPEC-V3R3-UPDATE-CLEANUP-001" per the initial population). This SPEC
//     does NOT modify Category A.
//   - Category B v.2.x-era NEW entries: DeprecatedBy ==
//     "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"
//   - Category C rc1-stage staging artifacts: DeprecatedBy ==
//     "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"
func TestDeprecatedPathsDeprecatedByConsistency(t *testing.T) {
	for _, entry := range DeprecatedPaths {
		switch entry.DeprecatedSince {
		case "SPEC-AGENCY-ABSORB-001":
			// Pre-existing; this SPEC does not modify these.
			if entry.DeprecatedBy == "" {
				t.Errorf("Category A entry %q has empty DeprecatedBy", entry.Path)
			}
		case "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
			"SPEC-V3R6-AGENT-FOLDER-SPLIT-001":
			// Category B (v.2.x-era) and Category C (rc1-stage) both
			// carry DeprecatedBy = this SPEC.
			if entry.DeprecatedBy != "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001" {
				t.Errorf("entry %q DeprecatedBy = %q, want %q", entry.Path, entry.DeprecatedBy, "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001")
			}
		default:
			t.Errorf("entry %q has unexpected DeprecatedSince value %q", entry.Path, entry.DeprecatedSince)
		}
	}
}

// TestDeprecatedPathsNoDuplicatePaths asserts the Path field is unique across
// the slice — duplicate enumeration would silently mask classification bugs.
func TestDeprecatedPathsNoDuplicatePaths(t *testing.T) {
	seen := map[string]int{}
	for i, entry := range DeprecatedPaths {
		if prev, ok := seen[entry.Path]; ok {
			t.Errorf("duplicate Path %q at index %d (previous at index %d)", entry.Path, i, prev)
		}
		seen[entry.Path] = i
	}
}

// TestDeprecatedPathsSlashSeparated asserts all paths use forward slashes —
// the slice is consumed by cross-platform code that depends on slash-separated
// path normalization (per REQ-VVCR-026 cross-platform compatibility).
func TestDeprecatedPathsSlashSeparated(t *testing.T) {
	for _, entry := range DeprecatedPaths {
		if strings.Contains(entry.Path, "\\") {
			t.Errorf("entry %q contains backslash; use forward slashes only", entry.Path)
		}
	}
}
