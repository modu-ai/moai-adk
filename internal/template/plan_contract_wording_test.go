package template

import (
	"io/fs"
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
