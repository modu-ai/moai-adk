package template

// lane_watchdog_renew_test.go — SPEC-FACTORY-LEASE-RENEW-001 (card t1663, M3).
//
// The lane stall watchdog renews its own card lease at the start of each pass,
// before it measures lane progress (REQ-FLR-010), and a refused renewal is
// recorded in the snapshot field renew_result while the same pass still measures
// (REQ-FLR-011). These tests read the shipped (embedded) watchdog procedure. They
// check procedure text, not a live watchdog turn: a text-only change passes them,
// and no criterion observes a written snapshot file (spec.md Residual risk).

import (
	"io/fs"
	"strings"
	"testing"
)

const watchdogSkillPath = ".claude/skills/moai-lane-watchdog/SKILL.md"

// watchdogSkillDoc returns the embedded watchdog SKILL.md split into its
// frontmatter and its body.
func watchdogSkillDoc(t *testing.T) (frontmatter, body string) {
	t.Helper()
	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates(): %v", err)
	}
	raw, err := fs.ReadFile(fsys, watchdogSkillPath)
	if err != nil {
		t.Fatalf("read %s: %v", watchdogSkillPath, err)
	}
	doc := string(raw)
	const fence = "---\n"
	if !strings.HasPrefix(doc, fence) {
		t.Fatalf("%s: no frontmatter fence at the top", watchdogSkillPath)
	}
	rest := doc[len(fence):]
	end := strings.Index(rest, "\n"+fence)
	if end < 0 {
		t.Fatalf("%s: frontmatter is not closed", watchdogSkillPath)
	}
	return rest[:end], rest[end+len("\n"+fence):]
}

// procedureHeadings lists the level-2 headings of the body in document order.
func procedureHeadings(body string) []string {
	var heads []string
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "## ") {
			heads = append(heads, line)
		}
	}
	return heads
}

// firstProcedureStep returns the heading that follows the boundary rule
// "## 0. Never prompt". That rule is a constraint, not a step, so the step after
// it is the first procedural step of a pass.
func firstProcedureStep(heads []string) (string, bool) {
	for i, h := range heads {
		if strings.HasPrefix(h, "## 0. Never prompt") && i+1 < len(heads) {
			return heads[i+1], true
		}
	}
	return "", false
}

// headingIndex returns the index of the first heading that contains the text,
// or -1 when none does.
func headingIndex(heads []string, contains string) int {
	for i, h := range heads {
		if strings.Contains(h, contains) {
			return i
		}
	}
	return -1
}

// TestLaneWatchdogRenewsFirst is AC-FLR-010 (REQ-FLR-010): the first procedural
// step renews the lane's own card lease through the factory renew verb, the verb
// is an allowed tool, and the measurement step follows the renewal.
func TestLaneWatchdogRenewsFirst(t *testing.T) {
	fm, body := watchdogSkillDoc(t)
	if !strings.Contains(fm, "Bash(moai factory renew:*)") {
		t.Errorf("allowed-tools does not list Bash(moai factory renew:*)")
	}
	heads := procedureHeadings(body)
	first, ok := firstProcedureStep(heads)
	if !ok {
		t.Fatalf("no procedural step follows the boundary rule %q", "## 0. Never prompt")
	}
	if !strings.Contains(first, "Renew") {
		t.Errorf("the first procedural step is %q; it must renew the card lease before any measurement", first)
	}
	stepText, ok := sectionOf(body, first)
	if !ok {
		t.Fatalf("section %q not found in the watchdog body", first)
	}
	if !strings.Contains(stepText, "moai factory renew") {
		t.Errorf("the first procedural step does not invoke the factory renew verb")
	}
	renewIdx := headingIndex(heads, "Renew")
	measureIdx := headingIndex(heads, "Measure progress")
	if measureIdx < 0 {
		t.Fatalf("no measurement step (Measure progress) in the watchdog body")
	}
	if renewIdx < 0 || renewIdx >= measureIdx {
		t.Errorf("the renewal step (heading index %d) does not precede the measurement step (heading index %d)", renewIdx, measureIdx)
	}
}

// TestLaneWatchdogRecordsRefusalAndMeasures is the procedure half of AC-FLR-011
// (REQ-FLR-011): the renewal step copies the verb's result into renew_result with
// the closed set of classes, a refused renewal does not end the pass, and the
// measurement step runs in the same pass and writes the snapshot after the verdict.
func TestLaneWatchdogRecordsRefusalAndMeasures(t *testing.T) {
	_, body := watchdogSkillDoc(t)
	heads := procedureHeadings(body)
	renew, ok := firstProcedureStep(heads)
	if !ok {
		t.Fatalf("no procedural step follows the boundary rule %q", "## 0. Never prompt")
	}
	renewText, ok := sectionOf(body, renew)
	if !ok {
		t.Fatalf("section %q not found in the watchdog body", renew)
	}
	for _, want := range []string{"--json", "result", "renew_result", "renewed", "refused-expired", "refused-not-holder", "error", "does not end the pass"} {
		if !strings.Contains(renewText, want) {
			t.Errorf("the renewal step lacks %q (refusal record, REQ-FLR-011)", want)
		}
	}
	measure := headingIndex(heads, "Measure progress")
	if measure < 0 {
		t.Fatalf("no measurement step (Measure progress) in the watchdog body")
	}
	if measure <= headingIndex(heads, "Renew") {
		t.Errorf("the measurement step does not run after the renewal step in the same pass")
	}
	measureText, ok := sectionOf(body, heads[measure])
	if !ok {
		t.Fatalf("section %q not found in the watchdog body", heads[measure])
	}
	for _, want := range []string{"renew_result", "after the verdict", "observed_at", "head_sha", "evidence_mtime_max", "window_state"} {
		if !strings.Contains(measureText, want) {
			t.Errorf("the measurement step lacks %q (snapshot record, REQ-FLR-011)", want)
		}
	}
}
