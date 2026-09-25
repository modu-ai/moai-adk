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
// artifacts named. The fixture pins `era: V3R6` (the current, non-grandfathered
// era) so an active SPEC's findings are NOT advisory; without it the missing
// progress.md classifies the fixture as V2.x and every finding is demoted,
// hiding any regression toward advisory.
func addTierSPEC(t *testing.T, root, id, tierLine string, siblings ...string) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fm := "---\nid: " + id + "\ntitle: Tier fixture\nversion: \"0.1.0\"\nstatus: draft\ncreated: 2026-09-24\nupdated: 2026-09-24\nauthor: tester\npriority: P2\nphase: v3.0.0\nmodule: internal/spec\ndependencies: []\nlifecycle: spec-anchored\ntags: fixture\nera: V3R6\n"
	if tierLine != "" {
		fm += tierLine + "\n"
	}
	// The Out of Scope section keeps a current-era fixture clean of the
	// MissingExclusions error, so the fail-open test can assert the lint exit.
	fm += "---\n\n# " + id + "\n\n### Out of Scope — fixture\n\n- nothing beyond the tier check\n"
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
	// Active (draft) SPEC in the current era: the finding must gate --strict.
	// MUTATION: set Advisory: true at the emission site — this turns red.
	if f.Advisory {
		t.Errorf("active current-era SPEC finding is advisory: %+v", f)
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

// TestTierArtifact_ClosedSPECIsAdvisory: two current-era SPECs with the same
// gap, one draft and one completed. Both report the finding; only the
// completed one is advisory (terminal-status demotion), so closed history
// never gates --strict while active work still does.
//
// MUTATIONS: dropping the completed-status flip makes both non-advisory;
// forcing Advisory: true makes both advisory — either turns this red.
func TestTierArtifact_ClosedSPECIsAdvisory(t *testing.T) {
	root := tierProject(t, tierTableFixture)
	addTierSPEC(t, root, "SPEC-LIVE-001", "tier: L", "plan.md", "acceptance.md")
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
	if len(got) != 2 {
		t.Fatalf("findings = %d, want 2: %+v", len(got), got)
	}
	advisory := map[string]bool{}
	for _, f := range got {
		advisory[filepath.Base(filepath.Dir(f.File))] = f.Advisory
	}
	if advisory["SPEC-LIVE-001"] {
		t.Errorf("active SPEC finding is advisory: %+v", got)
	}
	if !advisory["SPEC-DONE-001"] {
		t.Errorf("completed SPEC finding is not advisory: %+v", got)
	}
}

// TestTierArtifact_SubdirectoryBaseDir reproduces the CLI run from a project
// subdirectory: `moai spec lint SPEC-X` from <root>/internal/spec passes
// BaseDir = that subdirectory (it has no .moai/specs), while the lint target
// is the SPEC's absolute spec.md. The root must come from the SPEC's own path.
//
// MUTATION: resolve the root from BaseDir only — findings drop to 0.
func TestTierArtifact_SubdirectoryBaseDir(t *testing.T) {
	root := tierProject(t, tierTableFixture)
	addTierSPEC(t, root, "SPEC-SUBD-001", "tier: L", "plan.md", "acceptance.md")
	sub := filepath.Join(root, "internal", "spec")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	linter := spec.NewLinter(spec.LinterOptions{BaseDir: sub})
	report, err := linter.Lint([]string{filepath.Join(root, ".moai", "specs", "SPEC-SUBD-001", "spec.md")})
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if got := findingsForCode(report.Findings, tierArtifactMissingCode); len(got) != 1 {
		t.Fatalf("findings from subdirectory BaseDir = %d, want 1: %+v", len(got), got)
	}
}

// TestTierArtifact_RootBaseDirFallback covers BaseDir = <root> (the CLI shape
// when a project has no .moai/specs): a SPEC directory directly under the root
// cannot be mapped through .moai/specs, so the root falls back to BaseDir.
func TestTierArtifact_RootBaseDirFallback(t *testing.T) {
	root := tierProject(t, tierTableFixture)
	addTierSPEC(t, root, "SPEC-FLAT-001", "tier: L", "plan.md", "acceptance.md")
	flat := filepath.Join(root, "SPEC-FLAT-001")
	if err := os.Rename(filepath.Join(root, ".moai", "specs", "SPEC-FLAT-001"), flat); err != nil {
		t.Fatal(err)
	}

	linter := spec.NewLinter(spec.LinterOptions{BaseDir: root})
	report, err := linter.Lint([]string{filepath.Join(flat, "spec.md")})
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	if got := findingsForCode(report.Findings, tierArtifactMissingCode); len(got) != 1 {
		t.Fatalf("findings with BaseDir=<root> = %d, want 1: %+v", len(got), got)
	}
}

// TestTierArtifact_PerRootTables lints SPECs under three project roots in ONE
// Linter run: root A carries the standard table, root B a reduced table whose
// L row is only spec.md + plan.md, root C no rule file at all. Each SPEC must
// be judged by its own root's table.
//
// MUTATION: cache the table under a single fixed key (ignore the root) — B and
// C inherit A's table and report findings, turning this red.
func TestTierArtifact_PerRootTables(t *testing.T) {
	reduced := strings.Replace(tierTableFixture,
		"**5 files**: spec.md + plan.md + acceptance.md + design.md + research.md",
		"**2 files**: spec.md + plan.md", 1)
	rootA := tierProject(t, tierTableFixture)
	rootB := tierProject(t, reduced)
	rootC := tierProject(t, "")
	addTierSPEC(t, rootA, "SPEC-ROOTA-001", "tier: L", "plan.md", "acceptance.md")
	addTierSPEC(t, rootB, "SPEC-ROOTB-001", "tier: L", "plan.md")
	addTierSPEC(t, rootC, "SPEC-ROOTC-001", "tier: L")

	linter := spec.NewLinter(spec.LinterOptions{BaseDir: t.TempDir()})
	report, err := linter.Lint([]string{
		filepath.Join(rootA, ".moai", "specs", "SPEC-ROOTA-001", "spec.md"),
		filepath.Join(rootB, ".moai", "specs", "SPEC-ROOTB-001", "spec.md"),
		filepath.Join(rootC, ".moai", "specs", "SPEC-ROOTC-001", "spec.md"),
	})
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	got := findingsForCode(report.Findings, tierArtifactMissingCode)
	if len(got) != 1 || !strings.Contains(got[0].Message, "SPEC-ROOTA-001") {
		t.Fatalf("want exactly one finding, for SPEC-ROOTA-001; got %d: %+v", len(got), got)
	}
	if u := findingsForCode(report.Findings, tierArtifactUnreadableCode); len(u) != 0 {
		t.Errorf("no root has an unreadable table, got: %+v", u)
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
