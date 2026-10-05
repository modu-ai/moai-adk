package spec

import (
	"strings"
	"testing"
)

// SPEC-COVERAGE-RULE-SCOPE-001 M2 — reqIDPattern / extraction-pattern alignment.
//
// AC-CRS-001-003 (maps REQ-CRS-001-002).
//
// The two tests below are a PAIR and neither is sufficient alone:
//
//   - TestReqIDPattern_AcceptsObservedCorpusShapes asserts the validation
//     pattern accepts what the widened extraction actually collects. Alone, it
//     is satisfied by the vacuous mutant `^REQ-.*$`.
//   - TestReqIDPattern_RejectsShapesTheExtractionAccepts is the negative
//     control. It names IDs the WIDE extraction pattern collects and the
//     validation pattern must still reject, so `InvalidREQIDRule` retains a
//     non-empty rejection class and cannot pass vacuously. It is the mutant
//     probe required by `.claude/rules/moai/development/verification-completeness.md`
//     §2: a validation pattern aligned exactly to the extraction satisfies the
//     first test while violating its requirement, and this test catches it.
//
// TestReqIDPattern_ExtractionRejectionClassIsReachable closes the loop by
// showing the rejection class is reachable through the REAL rule, not only
// through the regexp: a fixture whose REQ line the extraction collects and the
// validation rejects must produce an InvalidREQID finding.

func TestReqIDPattern_AcceptsObservedCorpusShapes(t *testing.T) {
	// Every shape below is present in the live corpus and is collected by
	// reqLineWidePattern. Measured by TestCorpusRejectedREQIDDecomposition:
	// 825 of 1,085 wide extractions fail the pre-M2 reqIDPattern, in exactly
	// three shape classes (3-segment alpha 519, 5-segment alnum 200,
	// 3-segment alnum 106).
	cases := []struct {
		name string
		id   string
	}{
		{"four-segment canonical (pre-M2 shape)", "REQ-CRS-001-001"},
		{"three-segment", "REQ-HOOK-001"},
		{"digits inside the domain segment", "REQ-WF001-001"},
		{"five-segment, two-part alnum domain", "REQ-VNRN-RT-001-001"},
		{"two-part alpha domain", "REQ-HRN-FND-001"},
		{"domain ending in a digit", "REQ-TUX1-001"},
		{"alphanumeric domain", "REQ-WC01-001"},
		{"five-segment with alnum first domain part", "REQ-V3R2-RT-001-001"},
		{"three-segment alnum", "REQ-GD1-001"},
		{"long alpha domain beyond the old 5-char cap", "REQ-COVERAGE-001"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !reqIDPattern.MatchString(tc.id) {
				t.Errorf("reqIDPattern rejected %q; the widened extraction collects it, so InvalidREQID would fire corpus-wide", tc.id)
			}
		})
	}
}

func TestReqIDPattern_RejectsShapesTheExtractionAccepts(t *testing.T) {
	// Each ID below is collected by reqLineWidePattern (verified in-test) and
	// MUST still be rejected by reqIDPattern. If this test goes green-by-
	// acceptance, the validation pattern has been aligned to the extraction and
	// InvalidREQIDRule has become vacuous.
	cases := []struct {
		name string
		id   string
		why  string
	}{
		{"purely numeric domain", "REQ-123-001", "a domain segment must start with a letter"},
		{"domain starting with a digit", "REQ-256K-001", "a domain segment must start with a letter"},
		{"three-part domain", "REQ-AAA-BBB-CCC-001", "at most two domain segments"},
		{"one-digit tail", "REQ-ABC-1", "the numeric tail is groups of exactly three digits"},
		{"four-digit tail", "REQ-ABC-0001", "the numeric tail is groups of exactly three digits"},
		{"three-group tail", "REQ-ABC-001-001-001", "at most two numeric tail groups"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			line := "- " + tc.id + ": The system shall do something."
			if got := parseREQsWide(line); len(got) != 1 || got[0].ID != tc.id {
				t.Fatalf("precondition failed: the WIDE extraction must collect %q for this to be a negative control; got %#v", tc.id, got)
			}
			if reqIDPattern.MatchString(tc.id) {
				t.Errorf("reqIDPattern accepted %q (%s); validation is no longer narrower than extraction, so InvalidREQIDRule is vacuous", tc.id, tc.why)
			}
		})
	}
}

func TestReqIDPattern_ExtractionRejectionClassIsReachable(t *testing.T) {
	// The rejection class must be reachable through the REAL rule, not only the
	// regexp: doc.REQs is populated by the extraction, so a rule whose reject
	// branch no extraction can reach is a check whose non-execution is
	// indistinguishable from its success.
	doc := &SPECDoc{
		Path: "SPEC-X/spec.md",
		REQs: parseREQsWide(strings.Join([]string{
			"- REQ-ABC-001: The system shall be valid.",
			"- REQ-123-001: The system shall be rejected.",
		}, "\n")),
	}
	if len(doc.REQs) != 2 {
		t.Fatalf("precondition: expected the wide extraction to collect 2 REQs, got %d", len(doc.REQs))
	}

	findings := (&REQIDUniquenessRule{}).Check(doc, nil)
	var invalid []Finding
	for _, f := range findings {
		if f.Code == "InvalidREQID" {
			invalid = append(invalid, f)
		}
	}
	if len(invalid) != 1 {
		t.Fatalf("expected exactly 1 InvalidREQID finding (for REQ-123-001), got %d: %#v", len(invalid), findings)
	}
	if !strings.Contains(invalid[0].Message, "REQ-123-001") {
		t.Errorf("finding does not name the offending ID: %q", invalid[0].Message)
	}
}

func TestDuplicateREQIDRule_StillFiresOnWidenedShapes(t *testing.T) {
	// DuplicateREQIDRule shares REQIDUniquenessRule's loop and is NOT vacuous.
	// Widening the validation pattern moves previously-rejected IDs past the
	// `continue`, so duplicate detection must now reach them.
	doc := &SPECDoc{
		Path: "SPEC-X/spec.md",
		REQs: parseREQsWide(strings.Join([]string{
			"- REQ-HOOK-001: The system shall do a thing.",
			"- REQ-HOOK-002: The system shall do another thing.",
			"- REQ-HOOK-001: The system shall duplicate the first.",
		}, "\n")),
	}
	if len(doc.REQs) != 3 {
		t.Fatalf("precondition: expected 3 collected REQs, got %d", len(doc.REQs))
	}

	var dups []Finding
	for _, f := range (&REQIDUniquenessRule{}).Check(doc, nil) {
		if f.Code == "DuplicateREQID" {
			dups = append(dups, f)
		}
	}
	if len(dups) != 1 {
		t.Fatalf("expected exactly 1 DuplicateREQID finding for the widened shape REQ-HOOK-001, got %d", len(dups))
	}
	if !strings.Contains(dups[0].Message, "REQ-HOOK-001") {
		t.Errorf("finding does not name the duplicated ID: %q", dups[0].Message)
	}
}
