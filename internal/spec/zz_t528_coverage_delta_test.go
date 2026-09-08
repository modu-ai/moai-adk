// zz_t528_coverage_delta_test.go — t528 / SPEC-AC-COLLECTOR-ANCHOR-001,
// REQ-ACA-001-009 / AC-ACA-001-010 / M8.
//
// Counts CoverageRule ("CoverageIncomplete") findings across the live corpus,
// and counts the parse errors ParseAcceptanceCriteria actually RETURNS.
//
// The second count exists because lint cannot supply it. internal/spec/lint.go
// discards the error slice at its call site — `criteria, _ :=
// ParseAcceptanceCriteria(body, false)` — so DuplicateAcceptanceID and depth
// violations read as 0 through `moai spec lint` whether or not they fired.
// Measuring them there would be an unfalsifiable instrument, so this reads the
// second return value directly (acceptance.md §D.10, audit finding D3).
//
// This test MEASURES; it does not gate. A hard-coded expected finding count
// would fail the moment the corpus changed, for reasons having nothing to do
// with this card. The delta lives in progress.md §E.2, taken from this output.
package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestT528CoverageAndParseErrorCensus(t *testing.T) {
	root := "../../.moai/specs"

	var files, withACSection int
	byErrorKind := map[string]int{}
	var coverageFindings int
	var dupOrDepth []string

	rule := &CoverageRule{}

	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() != "spec.md" {
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return nil
		}
		files++
		body := string(b)

		criteria, parseErrs := ParseAcceptanceCriteria(body, false)
		for _, pe := range parseErrs {
			switch pe.(type) {
			case *DuplicateAcceptanceID:
				byErrorKind["DuplicateAcceptanceID"]++
				dupOrDepth = append(dupOrDepth, p+"  "+pe.Error())
			case *DanglingRequirementReference:
				byErrorKind["DanglingRequirementReference"]++
			case *MissingRequirementMapping:
				byErrorKind["MissingRequirementMapping"]++
			default:
				msg := pe.Error()
				if strings.Contains(msg, "acceptance criteria section not found") {
					byErrorKind["section-not-found"]++
				} else {
					byErrorKind["other: "+msg]++
					dupOrDepth = append(dupOrDepth, p+"  "+msg)
				}
			}
		}
		if len(criteria) > 0 {
			withACSection++
		}

		// REQs must be populated the way lint.go:633 populates it. CoverageRule
		// early-returns on len(doc.REQs)==0, so a doc built without them yields a
		// confident corpus-wide 0 that means only "the rule never ran" — the first
		// version of this test did exactly that.
		doc := &SPECDoc{Path: p, Criteria: criteria, REQs: parseREQsWithProvenance(body)}
		coverageFindings += len(rule.Check(doc, nil))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// NON-VACUITY: a walk that read nothing would report 0 findings and 0
	// errors, which is the shape this card has been fooled by before.
	if files < 800 {
		t.Fatalf("walked only %d spec.md files; the corpus denominator is 807. "+
			"A short walk makes every count below meaningless.", files)
	}

	// POSITIVE CONTROL for the CoverageRule count. A corpus total of 0 is
	// consistent both with "every REQ is covered" and with "the rule never
	// fired", and those are different facts. A document with one REQ and no AC
	// referencing it MUST produce exactly one finding.
	control := &SPECDoc{
		Path: "control.md",
		REQs: []REQEntry{{ID: "REQ-CTRL-001-001", Line: 1}},
	}
	if n := len(rule.Check(control, nil)); n != 1 {
		t.Fatalf("CoverageRule positive control produced %d findings, want 1 — "+
			"the corpus count below would be a rule that never fires, not a covered corpus", n)
	}
	t.Logf("positive control: uncovered REQ produces 1 CoverageIncomplete finding (rule fires)")

	t.Logf("spec.md walked                       = %d", files)
	t.Logf("files yielding >=1 acceptance criterion = %d", withACSection)
	t.Logf("CoverageRule (CoverageIncomplete) findings = %d", coverageFindings)
	for k, v := range byErrorKind {
		t.Logf("  parse error kind %-32s %d", k, v)
	}
	t.Logf("DuplicateAcceptanceID / depth / other-fatal occurrences = %d", len(dupOrDepth))
	for _, d := range dupOrDepth {
		t.Logf("    %s", d)
	}
}
