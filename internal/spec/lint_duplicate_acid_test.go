package spec_test

// lint_duplicate_acid_test.go — `moai spec lint` reports a duplicate inline AC
// id (card t564).
//
// Before the repair lint discarded every parse error at its call site, so a
// duplicate id was invisible to it, while `moai spec view` failed hard on the
// same document. The duplicate is now a DuplicateAcceptanceID finding, and the
// REQ mapping of every line carrying the id counts as coverage.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

const duplicateACIDCode = "DuplicateAcceptanceID"

func lintDuplicateACIDFixture(t *testing.T, c string) (string, []spec.Finding) {
	t.Helper()
	base := filepath.Join(testdataDir, "duplicateacid", c)
	path := filepath.Join(base, "spec.md")
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      base,
	})
	report, err := linter.Lint([]string{path})
	if err != nil {
		t.Fatalf("Lint(%s) returned unexpected error: %v", c, err)
	}
	if report.HasErrors() {
		t.Fatalf("fixture %s produced error-severity findings: %+v", c, report.Findings)
	}
	return path, report.Findings
}

// TestLintDuplicateACID_ReportsDuplicateAndKeepsCoverage: a citing bullet sits
// above the real declaration. Exactly one DuplicateAcceptanceID finding must
// point at the second occurrence, and the declaration's REQ must count as
// covered.
func TestLintDuplicateACID_ReportsDuplicateAndKeepsCoverage(t *testing.T) {
	path, findings := lintDuplicateACIDFixture(t, "citing-above")

	dups := findingsForCode(findings, duplicateACIDCode)
	if len(dups) != 1 {
		t.Fatalf("%s findings = %+v, want exactly 1", duplicateACIDCode, dups)
	}
	if dups[0].Severity != spec.SeverityWarning {
		t.Errorf("severity = %q, want %q", dups[0].Severity, spec.SeverityWarning)
	}
	if !strings.Contains(dups[0].Message, "AC-DAI-001-01") {
		t.Errorf("message = %q, want it to name AC-DAI-001-01", dups[0].Message)
	}

	// The line is body-relative, like every other per-SPEC finding in the
	// package; it must land on the SECOND line carrying the id.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, body, err := spec.ExtractFrontmatter(string(data))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(body, "\n")
	last := -1
	for i, l := range lines {
		if strings.Contains(l, "AC-DAI-001-01") {
			last = i + 1
		}
	}
	if dups[0].Line != last {
		t.Errorf("line = %d, want %d (second occurrence)", dups[0].Line, last)
	}

	if cov := findingsForCode(findings, "CoverageIncomplete"); len(cov) != 0 {
		t.Errorf("CoverageIncomplete findings = %+v, want none", cov)
	}
}

// TestLintDuplicateACID_NoDuplicateNoFinding is the control: the same document
// with a single declaration yields no duplicate finding.
func TestLintDuplicateACID_NoDuplicateNoFinding(t *testing.T) {
	_, findings := lintDuplicateACIDFixture(t, "single-declaration")

	if dups := findingsForCode(findings, duplicateACIDCode); len(dups) != 0 {
		t.Errorf("%s findings = %+v, want none", duplicateACIDCode, dups)
	}
	if cov := findingsForCode(findings, "CoverageIncomplete"); len(cov) != 0 {
		t.Errorf("CoverageIncomplete findings = %+v, want none", cov)
	}
}
