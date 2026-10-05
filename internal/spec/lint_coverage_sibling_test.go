package spec_test

// lint_coverage_sibling_test.go — CoverageRule sibling-acceptance.md reading
// (SPEC-COVERAGE-RULE-SCOPE-001 M4, REQ-CRS-001-006..008).
//
// AC-CRS-001-006a and -006b are a REGRESSION PAIR and are written as such: the
// first alone cannot distinguish "the rule reads the sibling correctly" from
// "the rule was switched off", because switching CoverageRule off satisfies it
// too. Only the two results diverging shows the repair.
//
// Judgment is by FINDING EMISSION, not by exit code. CoverageRule emits
// `warning` with `Advisory: true` at the emission site (M3, plan.md §D option
// A), and lint.go escalates a warning under --strict only when it is NOT
// advisory — so both halves of the pair exit 0, in plain and in --strict alike.
// An exit-code assertion here would be testing a criterion no correct
// implementation can satisfy; acceptance.md records the amendment.
//
// Fixtures live under testdata/coveragesibling/<case>/ and are linted ONE AT A
// TIME (an explicit spec.md path per call). Their acceptance.md files carry NO
// frontmatter, which keeps ArtifactStatusFieldForbiddenRule out of the picture.
//
// The sibling acceptance.md fixtures use the shape the corpus actually
// writes — a bold AC id with the mapping in an emphasis span, under a heading
// that says nothing about "acceptance" — which is precisely the shape the
// inline AC parser cannot read: findACSectionStart needs an `##` heading
// containing "acceptance", and parseSingleACLine needs `AC-…:` with a colon.
// A fixture written in the inline-parseable shape would pass through a much
// narrower path and would not decide the practice case.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

const coverageCode = "CoverageIncomplete"

// coverageSiblingFindings lints one fixture case and returns only
// CoverageIncomplete findings.
func coverageSiblingFindings(t *testing.T, c string) []spec.Finding {
	t.Helper()
	base := filepath.Join(testdataDir, "coveragesibling", c)
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      base,
	})
	report, err := linter.Lint([]string{filepath.Join(base, "spec.md")})
	if err != nil {
		t.Fatalf("Lint(%s) returned unexpected error: %v", c, err)
	}
	if report.HasErrors() {
		// Not the criterion (see the file header), but a fixture that trips an
		// error-severity rule would make the pair measure the wrong thing.
		t.Fatalf("fixture %s produced error-severity findings: %+v", c, report.Findings)
	}
	return findingsForCode(report.Findings, coverageCode)
}

// TestCoverageSibling_CoveredByAcceptanceMD decides AC-CRS-001-006a
// (maps REQ-CRS-001-006, REQ-CRS-001-008) — the front half of the regression
// pair. A REQ defined in spec.md whose only AC lives in the sibling
// acceptance.md is COVERED.
//
// This is the practice shape: the Tier M/L convention puts AC in acceptance.md
// and manager-develop-prompt-template.md names that file the AC SSOT, so a
// spec.md with no inline AC block is convention-following, not incomplete.
//
// MUTATION: make CoverageRule ignore the sibling (drop the merge in Check) —
// findings rise from 0 to 1. That mutation is the pair's discriminator: it
// leaves TestCoverageSibling_UncoveredStillFires green.
func TestCoverageSibling_CoveredByAcceptanceMD(t *testing.T) {
	if got := coverageSiblingFindings(t, "sibling-covered"); len(got) != 0 {
		t.Fatalf("CoverageIncomplete findings = %d, want 0: %+v", len(got), got)
	}
}

// TestCoverageSibling_UncoveredStillFires decides AC-CRS-001-006b
// (maps REQ-CRS-001-007) — the back half. A REQ mapped by no AC in EITHER file
// still reports.
//
// The fixture puts both REQs in one SPEC and maps only the first in the
// sibling, so a single lint run separates them. That is stronger than two
// fixtures: "the sibling is not read" and "the rule is switched off" each
// produce a count this test rejects (2 and 0 respectively), and only the
// correct behavior produces 1.
//
// MUTATION: return nil from CoverageRule.Check — findings drop from 1 to 0.
func TestCoverageSibling_UncoveredStillFires(t *testing.T) {
	got := coverageSiblingFindings(t, "sibling-uncovered")
	if len(got) != 1 {
		t.Fatalf("CoverageIncomplete findings = %d, want 1: %+v", len(got), got)
	}
	if !strings.Contains(got[0].Message, "REQ-CSB-002-002") {
		t.Errorf("finding does not name the uncovered REQ: %q", got[0].Message)
	}
	if got[0].Severity != spec.SeverityWarning || !got[0].Advisory {
		t.Errorf("severity/advisory = %q/%v, want warning/true (M3 plan.md §D option A)",
			got[0].Severity, got[0].Advisory)
	}
}

// TestCoverageSibling_NoAcceptanceArtifact decides AC-CRS-001-007
// (maps REQ-CRS-001-006): a Tier S SPEC has no acceptance.md, which is the
// dominant corpus shape and not a defect. The absent file must produce no
// error and no panic, and coverage must be judged from the inline AC section
// alone.
//
// The fixture carries one covered and one uncovered REQ, so a run that silently
// stopped judging coverage (0 findings) fails just as loudly as one that
// mis-read the absent file.
//
// MUTATION: treat a read error on the sibling as a hard failure instead of a
// skip — the test fails on the Lint error path.
func TestCoverageSibling_NoAcceptanceArtifact(t *testing.T) {
	got := coverageSiblingFindings(t, "no-sibling")
	if len(got) != 1 {
		t.Fatalf("CoverageIncomplete findings = %d, want 1: %+v", len(got), got)
	}
	if !strings.Contains(got[0].Message, "REQ-CSB-003-002") {
		t.Errorf("finding does not name the uncovered REQ: %q", got[0].Message)
	}
}
