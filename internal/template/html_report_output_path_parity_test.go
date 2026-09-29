// html_report_output_path_parity_test.go
// Pair-drift guard for the html-report skill output-path convention
// (SPEC-REPORTS-LIFECYCLE-001 REQ-RLC-003 / REQ-RLC-009 / AC-RLC-005).
//
// The html-report skill exists in two mirrors: the locally-executed copy
// (.claude/skills/moai-domain-html-report/SKILL.md) and the deployed
// template (internal/template/templates/.claude/skills/...). The output
// default moved from <cwd>/reports/ to .moai/reports/ — an edit that lands
// in one mirror only regresses the other on the next `moai update`, which
// is the same pair-drift hazard the hook wrapper parity guard pins for
// .sh.tmpl pairs.
//
// This guard pins both mirrors to the convention: the default output_path
// and Output section name .moai/reports/, zero <cwd>/reports/ remnants,
// and the create-before-write sentence (REQ-RLC-004). Its red is observed
// on a known failing input: a mutant copy restoring the legacy default
// must fail the assertion (TestHtmlReportOutputPathParity_MutationObservedRed).

package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const htmlReportSkillRel = ".claude/skills/moai-domain-html-report/SKILL.md"

func readHtmlReportSkill(t *testing.T, base string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(base, htmlReportSkillRel))
	if err != nil {
		t.Fatalf("read %s: %v", htmlReportSkillRel, err)
	}
	return string(data)
}

// htmlReportConventionFailures returns one description per violated
// convention point. Kept t-free so the mutation cell can require a non-empty
// failure set without a subtest's failure tainting its parent.
func htmlReportConventionFailures(content string, where string) []string {
	var failures []string
	if n := strings.Count(content, "<cwd>/reports/"); n != 0 {
		failures = append(failures, fmt.Sprintf("%s: %d legacy <cwd>/reports/ occurrence(s) remain", where, n))
	}
	if !strings.Contains(content, ".moai/reports/<slug>-<YYYYMMDD>.html") {
		failures = append(failures, where+": default output_path does not name .moai/reports/<slug>-<YYYYMMDD>.html")
	}
	if !strings.Contains(content, ".moai/reports/<slug>-<YYYYMMDD>.{html,md}") {
		failures = append(failures, where+": Output section does not name .moai/reports/<slug>-<YYYYMMDD>.{html,md}")
	}
	if !strings.Contains(content, "create it before writing either file") {
		failures = append(failures, where+": create-before-write sentence (REQ-RLC-004) missing")
	}
	return failures
}

// assertHtmlReportOutputConvention fails t unless content carries the
// .moai/reports/ output convention.
func assertHtmlReportOutputConvention(t *testing.T, content string, where string) {
	t.Helper()
	for _, f := range htmlReportConventionFailures(content, where) {
		t.Error(f)
	}
}

// TestHtmlReportOutputPathParity pins the .moai/reports/ output convention
// on BOTH mirrors and asserts the two copies stay byte-identical.
func TestHtmlReportOutputPathParity(t *testing.T) {
	root := hocProjectRoot(t)
	local := readHtmlReportSkill(t, root)
	mirror := readHtmlReportSkill(t, filepath.Join(root, "internal", "template", "templates"))

	assertHtmlReportOutputConvention(t, local, "local skill")
	assertHtmlReportOutputConvention(t, mirror, "template mirror")

	if local != mirror {
		t.Error("local and template-mirror SKILL.md have drifted apart (byte identity expected)")
	}
}

// TestHtmlReportOutputPathParity_MutationObservedRed is the completion cell:
// a mutant restoring the legacy default must fail the convention assertion.
// If the mutant passes, the guard cannot detect a regression and is
// unfinished.
func TestHtmlReportOutputPathParity_MutationObservedRed(t *testing.T) {
	root := hocProjectRoot(t)
	local := readHtmlReportSkill(t, root)

	mutant := strings.ReplaceAll(local, ".moai/reports/<slug>-<YYYYMMDD>", "<cwd>/reports/<slug>-<YYYYMMDD>")
	if mutant == local {
		t.Fatal("mutant is identical to source — nothing was mutated")
	}

	// The mutant MUST produce a non-empty failure set — that non-empty set
	// is the guard's observed red.
	if failures := htmlReportConventionFailures(mutant, "mutant"); len(failures) == 0 {
		t.Error("mutant with legacy default did NOT fail the convention guard — guard cannot detect the regression")
	}
}
