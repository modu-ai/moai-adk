package spec_test

// lint_tier_artifacts_test.go — TierArtifactMissingRule fixture-pair suite
// (card t1121). The rule reads the per-tier artifact set from the project's
// own `.claude/rules/moai/workflow/spec-workflow.md` Tier table at lint time,
// so every fixture here is a whole t.TempDir project: a minimal rule file plus
// one or more SPEC directories under .moai/specs/.
//
// Each positive test names the mutation that must turn it red.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

const (
	tierArtifactMissingCode    = "TierArtifactMissing"
	tierArtifactUnreadableCode = "TierArtifactSetUnreadable"
)

// tierTableFixture is a minimal spec-workflow.md carrying the Tier table in
// the shipped shape, including the parenthetical prose on the S row that must
// not add a duplicate spec.md.
const tierTableFixture = `# SPEC Workflow

## SPEC Complexity Tier (S/M/L)

| Tier | Scope guidance (LOC) | Files affected | Artifact set | plan-auditor PASS threshold |
|------|----------------------|----------------|--------------|------------------------------|
| S (Simple) | < 300 LOC | < 5 files | **2 files**: spec.md + plan.md (AC inline in spec.md §3) | 0.75 |
| M (Medium) | 300 - 1000 LOC | 5 - 15 files | **3 files**: spec.md + plan.md + acceptance.md | 0.80 |
| L (Large) | > 1000 LOC or constitutional | > 15 files | **5 files**: spec.md + plan.md + acceptance.md + design.md + research.md | 0.85 |
`

// tierProject creates a temp project root. When ruleBody is non-empty it is
// written as the project's spec-workflow.md; an empty ruleBody leaves the rule
// file absent.
func tierProject(t *testing.T, ruleBody string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if ruleBody != "" {
		ruleDir := filepath.Join(root, ".claude", "rules", "moai", "workflow")
		if err := os.MkdirAll(ruleDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(ruleDir, "spec-workflow.md"), []byte(ruleBody), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// addTierSPEC writes a SPEC directory whose spec.md declares tierLine (the
// literal frontmatter line, or "" for no tier field) plus the extra sibling
// artifacts named.
func addTierSPEC(t *testing.T, root, id, tierLine string, siblings ...string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fm := "---\nid: " + id + "\ntitle: Tier fixture\nversion: \"0.1.0\"\nstatus: draft\ncreated: 2026-09-24\nupdated: 2026-09-24\nauthor: tester\npriority: P2\nphase: v3.0.0\nmodule: internal/spec\ndependencies: []\ntags: fixture\n"
	if tierLine != "" {
		fm += tierLine + "\n"
	}
	fm += "---\n\n# " + id + "\n"
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(fm), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, s := range siblings {
		if err := os.WriteFile(filepath.Join(dir, s), []byte("# "+s+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// lintTierProject lints the whole corpus the way the CLI does (BaseDir is the
// .moai/specs directory, discovery path) and returns the report findings.
func lintTierProject(t *testing.T, root string) []spec.Finding {
	t.Helper()
	linter := spec.NewLinter(spec.LinterOptions{BaseDir: filepath.Join(root, ".moai", "specs")})
	report, err := linter.Lint(nil)
	if err != nil {
		t.Fatalf("Lint returned unexpected error: %v", err)
	}
	return report.Findings
}

// TestTierArtifact_LMissingDesignResearch is the RED anchor: a Tier L SPEC
// carrying spec/plan/acceptance but no design.md or research.md yields exactly
// one warning naming both missing files.
//
// MUTATION: make TierArtifactMissingRule.Check return nil — findings drop to 0.
func TestTierArtifact_LMissingDesignResearch(t *testing.T) {
	root := tierProject(t, tierTableFixture)
	addTierSPEC(t, root, "SPEC-TIERL-001", "tier: L", "plan.md", "acceptance.md")

	got := findingsForCode(lintTierProject(t, root), tierArtifactMissingCode)
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1: %+v", len(got), got)
	}
	f := got[0]
	if f.Severity != spec.SeverityWarning {
		t.Errorf("Severity = %q, want warning", f.Severity)
	}
	for _, want := range []string{"SPEC-TIERL-001", "tier L", "design.md", "research.md", "spec-workflow.md"} {
		if !strings.Contains(f.Message, want) {
			t.Errorf("Message missing %q: %q", want, f.Message)
		}
	}
	for _, notWant := range []string{"plan.md", "acceptance.md"} {
		if strings.Contains(f.Message, notWant+",") || strings.HasSuffix(f.Message, notWant) {
			t.Errorf("Message names present file %q as missing: %q", notWant, f.Message)
		}
	}
}

// TestTierArtifact_AllTiersChecked proves S and M sets are enforced too, not
// only L: Tier S without plan.md and Tier M without acceptance.md each fire.
func TestTierArtifact_AllTiersChecked(t *testing.T) {
	root := tierProject(t, tierTableFixture)
	addTierSPEC(t, root, "SPEC-TIERS-001", "tier: S")
	addTierSPEC(t, root, "SPEC-TIERM-001", "tier: M", "plan.md")

	got := findingsForCode(lintTierProject(t, root), tierArtifactMissingCode)
	if len(got) != 2 {
		t.Fatalf("findings = %d, want 2: %+v", len(got), got)
	}
	joined := got[0].Message + "\n" + got[1].Message
	if !strings.Contains(joined, "SPEC-TIERS-001") || !strings.Contains(joined, "plan.md") {
		t.Errorf("Tier S plan.md gap not reported: %q", joined)
	}
	if !strings.Contains(joined, "SPEC-TIERM-001") || !strings.Contains(joined, "acceptance.md") {
		t.Errorf("Tier M acceptance.md gap not reported: %q", joined)
	}
}

// TestTierArtifact_NegativeControls: complete sets per tier, an absent tier,
// and a non-S/M/L tier value all produce no finding.
func TestTierArtifact_NegativeControls(t *testing.T) {
	root := tierProject(t, tierTableFixture)
	addTierSPEC(t, root, "SPEC-NEGS-001", "tier: S", "plan.md")
	addTierSPEC(t, root, "SPEC-NEGM-001", "tier: M", "plan.md", "acceptance.md")
	addTierSPEC(t, root, "SPEC-NEGL-001", "tier: L", "plan.md", "acceptance.md", "design.md", "research.md")
	addTierSPEC(t, root, "SPEC-NEGQ-001", `tier: "l"`, "plan.md", "acceptance.md", "design.md", "research.md")
	addTierSPEC(t, root, "SPEC-NEGA-001", "")
	addTierSPEC(t, root, "SPEC-NEGN-001", "tier: 2")

	findings := lintTierProject(t, root)
	if got := findingsForCode(findings, tierArtifactMissingCode); len(got) != 0 {
		t.Fatalf("negative controls produced findings: %+v", got)
	}
	if got := findingsForCode(findings, tierArtifactUnreadableCode); len(got) != 0 {
		t.Fatalf("readable table produced an unreadable warning: %+v", got)
	}
}

// TestTierArtifact_FailOpenWithoutRuleFile: a project that does not carry the
// rule file gets zero findings from either code and no lint error, even with
// a Tier L SPEC missing four artifacts.
func TestTierArtifact_FailOpenWithoutRuleFile(t *testing.T) {
	root := tierProject(t, "")
	addTierSPEC(t, root, "SPEC-OPEN-001", "tier: L")

	linter := spec.NewLinter(spec.LinterOptions{BaseDir: filepath.Join(root, ".moai", "specs")})
	report, err := linter.Lint(nil)
	if err != nil {
		t.Fatalf("Lint returned error without the rule file: %v", err)
	}
	if got := findingsForCode(report.Findings, tierArtifactMissingCode); len(got) != 0 {
		t.Errorf("absent rule file produced TierArtifactMissing: %+v", got)
	}
	if got := findingsForCode(report.Findings, tierArtifactUnreadableCode); len(got) != 0 {
		t.Errorf("absent rule file produced TierArtifactSetUnreadable: %+v", got)
	}
	if report.HasErrors() {
		t.Errorf("absent rule file changed the lint exit: %+v", report.Findings)
	}
}

// TestTierArtifact_UnreadableTableWarnsOnce: a rule file present but without a
// parseable Tier table yields exactly ONE corpus warning (not one per SPEC)
// and no per-SPEC findings.
//
// MUTATION: make the unreadable branch emit nothing — the count drops to 0.
func TestTierArtifact_UnreadableTableWarnsOnce(t *testing.T) {
	root := tierProject(t, "# SPEC Workflow\n\nThe tier table moved elsewhere.\n")
	addTierSPEC(t, root, "SPEC-BRKA-001", "tier: L")
	addTierSPEC(t, root, "SPEC-BRKB-001", "tier: M")

	findings := lintTierProject(t, root)
	got := findingsForCode(findings, tierArtifactUnreadableCode)
	if len(got) != 1 {
		t.Fatalf("unreadable-table warnings = %d, want exactly 1: %+v", len(got), got)
	}
	if got[0].Severity != spec.SeverityWarning {
		t.Errorf("Severity = %q, want warning", got[0].Severity)
	}
	if !strings.Contains(got[0].Message, "spec-workflow.md") {
		t.Errorf("Message does not name the SSOT file: %q", got[0].Message)
	}
	if miss := findingsForCode(findings, tierArtifactMissingCode); len(miss) != 0 {
		t.Errorf("unreadable table must not drive per-SPEC checks: %+v", miss)
	}
}

// TestTierArtifact_ClosedSPECIsAdvisory: a completed SPEC with the same gap
// still reports the finding, but era/terminal-status demotion marks it
// advisory, so closed history never gates --strict.
func TestTierArtifact_ClosedSPECIsAdvisory(t *testing.T) {
	root := tierProject(t, tierTableFixture)
	addTierSPEC(t, root, "SPEC-DONE-001", "tier: L", "plan.md", "acceptance.md")
	specPath := filepath.Join(root, ".moai", "specs", "SPEC-DONE-001", "spec.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(specPath, []byte(strings.Replace(string(data), "status: draft", "status: completed", 1)), 0o644); err != nil {
		t.Fatal(err)
	}

	got := findingsForCode(lintTierProject(t, root), tierArtifactMissingCode)
	if len(got) != 1 {
		t.Fatalf("findings = %d, want 1: %+v", len(got), got)
	}
	if !got[0].Advisory {
		t.Errorf("completed SPEC finding is not advisory: %+v", got[0])
	}
}

// TestTierArtifact_LintSkipApplies: lint.skip on the SPEC silences the code,
// matching every other per-SPEC rule.
func TestTierArtifact_LintSkipApplies(t *testing.T) {
	root := tierProject(t, tierTableFixture)
	addTierSPEC(t, root, "SPEC-SKIP-001", "tier: L\nlint:\n  skip: [TierArtifactMissing]")

	if got := findingsForCode(lintTierProject(t, root), tierArtifactMissingCode); len(got) != 0 {
		t.Errorf("lint.skip did not silence the rule: %+v", got)
	}
}
