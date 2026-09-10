package spec

// parser_duplicate_union_test.go — duplicate AC ids in the inline parser
// (card t564).
//
// The inline AC grammar cannot tell a bullet that CITES an id from the bullet
// that DECLARES it: acIDPattern requires only a separator after the id. So a
// rule that keeps one line and drops the other loses the dropped line's REQ
// mapping silently. The resolution pinned here: keep the first line's text,
// collect the REQ mappings of EVERY line carrying the id, and still report the
// duplicate so the author sees it.

import (
	"slices"
	"testing"
)

func duplicateIDErrors(errs []error) []*DuplicateAcceptanceID {
	var dups []*DuplicateAcceptanceID
	for _, err := range errs {
		if d, ok := err.(*DuplicateAcceptanceID); ok {
			dups = append(dups, d)
		}
	}
	return dups
}

// TestParser_DuplicateID_CitingAboveKeepsDeclarationMapping: a citing bullet
// above the real declaration. Before the repair the first line wins and the
// declaration's `maps REQ-DUP-001-001` is lost.
func TestParser_DuplicateID_CitingAboveKeepsDeclarationMapping(t *testing.T) {
	markdown := `
## Acceptance Criteria

- AC-DUP-001-01: cited here only as a cross-reference note
- AC-DUP-001-01: Given a SPEC, When it is parsed, Then the declaration is kept (maps REQ-DUP-001-001)
`
	criteria, errs := ParseAcceptanceCriteria(markdown, false)

	if len(criteria) != 1 || criteria[0].ID != "AC-DUP-001-01" {
		t.Fatalf("criteria = %+v, want exactly one root AC-DUP-001-01", criteria)
	}
	if !collectAllREQIDs(criteria)["REQ-DUP-001-001"] {
		t.Errorf("REQ-DUP-001-001 not collected from the duplicate declaration; covered = %v", collectAllREQIDs(criteria))
	}
	// The single root is auto-wrapped; the leaf `.a` is what `spec view`
	// renders, so the merged mapping must reach it, not only the wrapper.
	if len(criteria[0].Children) != 1 || !slices.Contains(criteria[0].Children[0].RequirementIDs, "DUP-001-001") {
		t.Errorf("leaf = %+v, want one child carrying DUP-001-001", criteria[0].Children)
	}
	dups := duplicateIDErrors(errs)
	if len(dups) != 1 || dups[0].ID != "AC-DUP-001-01" {
		t.Errorf("duplicate errors = %+v, want exactly one for AC-DUP-001-01", dups)
	}
}

// TestParser_DuplicateID_UnionsMappingsOfEveryLine: each line maps a
// different REQ; both must be collected, whatever the order.
func TestParser_DuplicateID_UnionsMappingsOfEveryLine(t *testing.T) {
	markdown := `
## Acceptance Criteria

- AC-DUP-002-01: cross-reference note (maps REQ-DUP-002-002)
- AC-DUP-002-01: Given a SPEC, When it is parsed, Then it is kept (maps REQ-DUP-002-001)
`
	criteria, errs := ParseAcceptanceCriteria(markdown, false)

	covered := collectAllREQIDs(criteria)
	for _, id := range []string{"REQ-DUP-002-001", "REQ-DUP-002-002"} {
		if !covered[id] {
			t.Errorf("%s not collected; covered = %v", id, covered)
		}
	}
	if len(duplicateIDErrors(errs)) != 1 {
		t.Errorf("duplicate errors = %v, want exactly one", errs)
	}
}

// TestParser_DuplicateID_DeclarationAboveStillReported: the order that already
// resolves correctly today must keep its mapping and keep reporting the
// duplicate.
func TestParser_DuplicateID_DeclarationAboveStillReported(t *testing.T) {
	markdown := `
## Acceptance Criteria

- AC-DUP-003-01: Given a SPEC, When it is parsed, Then it is kept (maps REQ-DUP-003-001)
- AC-DUP-003-01: cited here only as a cross-reference note
`
	criteria, errs := ParseAcceptanceCriteria(markdown, false)

	if !collectAllREQIDs(criteria)["REQ-DUP-003-001"] {
		t.Errorf("REQ-DUP-003-001 not collected; covered = %v", collectAllREQIDs(criteria))
	}
	if len(duplicateIDErrors(errs)) != 1 {
		t.Errorf("duplicate errors = %v, want exactly one", errs)
	}
}
