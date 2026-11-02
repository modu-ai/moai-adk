package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Card t2007 guard: the shipped plan workflow must not make the plan-phase
// audit-ready signal (progress.md §E.1) depend on user approval (approval is
// the separate plan->run Kickoff gate), and must name the single owner of the
// issue_number frontmatter write (manager-spec) so the optional field, the
// "no issue_number for new SPECs" sentence, and Step 2.5.2 do not contradict.
//
// Lexical checks on the embedded copy only; they prove the wording, not that an
// agent obeys it.

// between returns the text from start up to the next occurrence of end after it.
func between(doc, start, end string) (string, bool) {
	i := strings.Index(doc, start)
	if i < 0 {
		return "", false
	}
	rest := doc[i+len(start):]
	if j := strings.Index(rest, end); j >= 0 {
		rest = rest[:j]
	}
	return rest, true
}

// TestPlanSignalDocsNameAuditReady guards the two rule docs that describe the
// plan-phase completion signal: both must name the `audit_ready: true` line the
// Kickoff gate actually reads, in the embedded copy and in the local .claude
// tree. Lexical: it proves the wording, not that an agent writes the line.
func TestPlanSignalDocsNameAuditReady(t *testing.T) {
	t.Parallel()

	files := []struct{ rel, marker string }{
		{"rules/moai/workflow/spec-workflow.md", "Pre-condition: plan.md records"},
		{"rules/moai/development/spec-frontmatter-schema.md", "## §E.1 Plan-phase Audit-Ready Signal"},
	}
	roots := []struct{ label, dir string }{
		{"local", filepath.Join("..", "..", ".claude")},
		{"template", filepath.Join("templates", ".claude")},
	}
	for _, root := range roots {
		for _, f := range files {
			path := filepath.Join(root.dir, f.rel)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("%s: ReadFile(%q): %v", root.label, path, err)
				continue
			}
			var line string
			for _, l := range strings.Split(string(raw), "\n") {
				if strings.Contains(l, f.marker) {
					line = l
					break
				}
			}
			if line == "" {
				t.Errorf("%s %s: line containing %q not found", root.label, f.rel, f.marker)
				continue
			}
			if !strings.Contains(line, "audit_ready: true") {
				t.Errorf("%s %s: signal description lacks `audit_ready: true`: %.160s", root.label, f.rel, line)
			}
			if idx, p := strings.Index(line, "audit_ready: true"), strings.Index(line, "plan_status"); idx >= 0 && p >= 0 && p < idx {
				t.Errorf("%s %s: `plan_status` is named before `audit_ready: true` (audit_ready must be the first-named signal)", root.label, f.rel)
			}
		}
	}
}

func TestPlanContractWording(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	read := func(p string) string {
		raw, err := fs.ReadFile(fsys, p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		return string(raw)
	}

	plan := read(".claude/skills/moai/workflows/plan.md")
	sig, ok := between(plan, "\n## Audit-Ready Signal\n", "\n## ")
	if !ok {
		t.Fatal("plan.md: section '## Audit-Ready Signal' not found")
	}
	if strings.Contains(sig, "user approved") {
		t.Error("plan.md Audit-Ready Signal still makes the signal depend on 'user approved'")
	}
	if !strings.Contains(sig, "Kickoff") {
		t.Error("plan.md Audit-Ready Signal does not name the Kickoff gate as the separate approval")
	}

	if !strings.Contains(sig, "audit_ready: true") {
		t.Error("plan.md Audit-Ready Signal block lacks the `audit_ready: true` line the kickoff gate reads")
	}

	sa := read(".claude/skills/moai/workflows/plan/spec-assembly.md")
	step, ok := between(sa, "#### Step 2.5.2", "#### Step 2.5.3")
	if !ok {
		t.Fatal("spec-assembly.md: Step 2.5.2 not found")
	}
	if !strings.Contains(step, "manager-spec") {
		t.Error("spec-assembly.md Step 2.5.2 does not name manager-spec as owner of the issue_number write")
	}
	if crit, ok := between(sa, "**Audit-ready signal**", "\n---"); !ok || !strings.Contains(crit, "audit_ready: true") {
		t.Error("spec-assembly.md completion-criteria signal block lacks `audit_ready: true`")
	}
	para, ok := between(sa, "[HARD] Per the late-branch opt-in policy", "\n")
	if !ok {
		t.Fatal("spec-assembly.md: late-branch opt-in paragraph not found")
	}
	if strings.Contains(para, "MUST NOT carry an `issue_number` field for new SPECs") {
		t.Error("spec-assembly.md still says unconditionally that new SPEC frontmatter MUST NOT carry issue_number")
	}
	if !strings.Contains(para, "Step 2.5.2") {
		t.Error("spec-assembly.md opt-in paragraph does not point at Step 2.5.2 as the only writer of issue_number")
	}
}

// TestPlanSpecAssemblyIssueAndArtifactWording guards three in-doc contradictions
// in spec-assembly.md: the Phase 12 skip-condition header versus its warning
// bullet, the default-flow example expecting an Issue without --issue, and the
// audit-ready signal naming tasks.md as a plan-phase artifact (no tier produces
// it at plan time). Lexical checks on the embedded copy only.
func TestPlanSpecAssemblyIssueAndArtifactWording(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	raw, err := fs.ReadFile(fsys, ".claude/skills/moai/workflows/plan/spec-assembly.md")
	if err != nil {
		t.Fatalf("read spec-assembly.md: %v", err)
	}
	sa := string(raw)

	skip, ok := between(sa, "Skip conditions (", "#### Step 2.5.1")
	if !ok {
		t.Fatal("spec-assembly.md: Phase 12 'Skip conditions (' block not found")
	}
	if strings.Contains(skip, "silent skip with no warning") {
		t.Error("Phase 12 skip header still says 'silent skip with no warning' while a bullet logs a warning")
	}
	if !strings.Contains(skip, "exactly one warning line") {
		t.Error("Phase 12 skip block does not state the single-warning-line rule for --issue with missing gh or remote")
	}

	flow, ok := between(sa, "### Normal Flow", "### Existing Assets Flow")
	if !ok {
		t.Fatal("spec-assembly.md: Normal Flow example not found")
	}
	// Unconditional: the fixed example itself names --issue (to say an Issue is
	// created only with it), so gating these checks on its absence made them dead.
	if strings.Contains(flow, "GitHub Issue created") {
		t.Error("Normal Flow example (no --issue) still expects a GitHub Issue to be created")
	}
	if !strings.Contains(flow, "no GitHub Issue") {
		t.Error("Normal Flow example does not state that no GitHub Issue is created without --issue")
	}

	sig, ok := between(sa, "**Audit-ready signal**", "\n---")
	if !ok {
		t.Fatal("spec-assembly.md: audit-ready signal block not found")
	}
	if strings.Contains(sig, "tasks.md") {
		t.Error("audit-ready signal still lists tasks.md as a plan-phase artifact")
	}
	if !strings.Contains(sig, "tier") {
		t.Error("audit-ready signal does not name the tier-appropriate artifact set")
	}
}

// TestPlanStaleContractPhrases guards two remaining stale phrases from the plan
// contract reconciliation (PLN-08 and PLN-10). Lexical checks on the embedded
// copy only: they prove the wording, not that a validator or agent obeys it.
//   - PLN-08: the Step 2.3.4 retry loop must not restate a fixed three-iteration
//     budget; the tier-resolved ceiling (S1/M2/L3) owns every review attempt.
//   - PLN-10: the validator-authoring docs must not forbid mixed patterns in one
//     requirement while the canonical GEARS clause chains Where/While/When.
func TestPlanStaleContractPhrases(t *testing.T) {
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	read := func(p string) string {
		raw, err := fs.ReadFile(fsys, p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		return string(raw)
	}

	sa := read(".claude/skills/moai/workflows/plan/spec-assembly.md")
	for _, stale := range []string{"Repeat until PASS or 3 iterations exhausted", "Retry Loop (max 3 iterations)", "review iteration {N}/3"} {
		if strings.Contains(sa, stale) {
			t.Errorf("spec-assembly.md still carries the fixed-budget phrase %q (the tier-resolved ceiling owns the attempts)", stale)
		}
	}

	stale := map[string][]string{
		".claude/skills/moai-workflow-spec/modules/advanced-patterns.md": {"exactly one EARS pattern", "Detect mixed patterns"},
		".claude/skills/moai-workflow-spec/references/reference.md":      {"No mixing of patterns in single requirement"},
	}
	for path, phrases := range stale {
		doc := read(path)
		for _, p := range phrases {
			if strings.Contains(doc, p) {
				t.Errorf("%s still says %q, which contradicts the GEARS compound clause", path, p)
			}
		}
	}
}
