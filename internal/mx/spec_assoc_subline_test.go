package mx

import (
	"strings"
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

// TestAC003_UnresolvedSpecRefFlagButKeep verifies that a captured @MX:SPEC ID
// not present in the known-SPEC set is kept in spec_associations AND produces
// an UnresolvedSpecRef warning, with no panic (AC-MX-ASSOC-003 /
// REQ-MX-ASSOC-003, flag-but-keep).
func TestAC003_UnresolvedSpecRefFlagButKeep(t *testing.T) {
	tag := Tag{
		Kind:    MXNote,
		File:    "internal/nowhere/helper.go",
		Line:    10,
		Body:    "context note",
		SpecRef: "SPEC-DOES-NOT-EXIST-001",
	}

	// specModules does NOT contain the fixture SPEC → unresolved.
	associator := NewSpecAssociator(map[string][]string{
		"SPEC-OTHER-001": {"internal/other/"},
	})
	specs, warnings := associator.AssociateWithDiagnostics(tag)

	if !containsStr(specs, "SPEC-DOES-NOT-EXIST-001") {
		t.Errorf("flag-but-keep: expected SPEC-DOES-NOT-EXIST-001 kept in %v", specs)
	}

	found := false
	for _, w := range warnings {
		if strings.Contains(w, "UnresolvedSpecRef") && strings.Contains(w, "SPEC-DOES-NOT-EXIST-001") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected UnresolvedSpecRef warning naming SPEC-DOES-NOT-EXIST-001, got %v", warnings)
	}

	// Associate (the backward-compatible entry point) must still return the ID
	// even though it discards diagnostics — the keep half of flag-but-keep.
	if !containsStr(associator.Associate(tag), "SPEC-DOES-NOT-EXIST-001") {
		t.Errorf("Associate should still keep the unresolved ID")
	}
}

// TestAC003_ResolvedSpecRefNoWarning verifies that a captured @MX:SPEC ID that
// IS in the known-SPEC set produces NO UnresolvedSpecRef warning (the validation
// is flag-but-keep, not flag-always).
func TestAC003_ResolvedSpecRefNoWarning(t *testing.T) {
	tag := Tag{
		Kind:    MXNote,
		File:    "internal/other/helper.go",
		Line:    10,
		Body:    "context note",
		SpecRef: "SPEC-OTHER-001",
	}

	associator := NewSpecAssociator(map[string][]string{
		"SPEC-OTHER-001": {"internal/other/"},
	})
	_, warnings := associator.AssociateWithDiagnostics(tag)

	for _, w := range warnings {
		if strings.Contains(w, "UnresolvedSpecRef") {
			t.Errorf("resolved SPEC should not warn, got %q", w)
		}
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
