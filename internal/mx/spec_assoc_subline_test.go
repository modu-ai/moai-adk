package mx

import (
	"testing"
)

// TestAC001_SubLineDrivesAssociationWithNoPathOrBodyMatch verifies that a
// captured @MX:SPEC sub-line drives a spec_associations entry even when neither
// the path-based nor body-based sources would match (AC-MX-ASSOC-001 /
// REQ-MX-ASSOC-001 + REQ-MX-ASSOC-002).
func TestAC001_SubLineDrivesAssociationWithNoPathOrBodyMatch(t *testing.T) {
	// Body contains no SPEC-* token; file path matches no SPEC module path; the
	// sole association source is the captured SpecRef field.
	tag := Tag{
		Kind:    MXNote,
		File:    "internal/nowhere/helper.go",
		Line:    10,
		Body:    "a context note with no SPEC token",
		SpecRef: "SPEC-FIXTURE-001",
	}

	associator := NewSpecAssociator(map[string][]string{
		"SPEC-FIXTURE-001": {"some/other/path/"},
	})
	specs := associator.Associate(tag)

	if !containsStr(specs, "SPEC-FIXTURE-001") {
		t.Errorf("Associate: expected SPEC-FIXTURE-001 in %v", specs)
	}
}

// TestAC007_SubLineAndBodyDeDup verifies that a SPEC ID present in BOTH the
// body and the @MX:SPEC sub-line appears exactly once in spec_associations
// (AC-MX-ASSOC-007, SHOULD).
func TestAC007_SubLineAndBodyDeDup(t *testing.T) {
	tag := Tag{
		Kind:    MXNote,
		File:    "internal/nowhere/helper.go",
		Line:    10,
		Body:    "see SPEC-DUP-001 for context",
		SpecRef: "SPEC-DUP-001",
	}

	associator := NewSpecAssociator(map[string][]string{})
	specs := associator.Associate(tag)

	count := 0
	for _, s := range specs {
		if s == "SPEC-DUP-001" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("de-dup: expected SPEC-DUP-001 exactly once, got %d (%v)", count, specs)
	}
}

// containsStr reports whether s is present in the slice.
func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
