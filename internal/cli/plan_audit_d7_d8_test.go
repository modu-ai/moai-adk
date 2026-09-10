// plan_audit_d7_d8_test.go: Fixture-based validation of the D7 and D8 audit
// verbs documented in the plan-auditor agent definition
// (`internal/template/templates/.claude/agents/moai/plan-auditor.md`, mirrored
// at `.claude/agents/moai/plan-auditor.md`).
//
// These tests do NOT invoke the plan-auditor agent itself. They EXTRACT the
// bash verbs from the agent definition — the first ```bash block under the
// Group 7 and Group 8 headings — and run them against synthetic SPEC fixtures.
// The verbs are read from the document on every run, never pasted into this
// file: a pasted copy stays green when the document changes, which is exactly
// the drift these tests exist to catch.
//
// Contract pinned here (card t623):
//   - D7: the script never emits BLOCKING. A retired/superseded/archived
//     reference prints a REVIEW line, plus a reconciliation-candidate paragraph
//     when one names the SPEC next to a reconciliation keyword; the auditor
//     decides BLOCKING by reading.
//   - D8: the build-constraint check is scoped per heading-delimited section,
//     so a //go:build in one section cannot cover a syscall in another; an
//     unreadable input is reported as GAP.
//
// Sentinel on failure: PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT — the bash verb in the
// agent definition no longer matches the documented behavior.
//
// @MX:ANCHOR: [AUTO] fan_in=2 — guards plan-auditor D7/D8 verb correctness.
// Touching this test signature affects the audit contract for cross-SPEC
// reconciliation (D7) and cross-platform discipline (D8).
// @MX:REASON: The fixtures run the verbs extracted from the agent definition,
// so reverting either verb to a weaker or stronger form turns a named cell red.
package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// planAuditorDocs are the two copies of the agent definition, relative to this
// package directory. The template copy is the one distributed to projects.
var planAuditorDocs = map[string]string{
	"template": filepath.Join("..", "template", "templates", ".claude", "agents", "moai", "plan-auditor.md"),
	"local":    filepath.Join("..", "..", ".claude", "agents", "moai", "plan-auditor.md"),
}

// auditVerb returns the first ```bash block after the line starting with
// heading in doc, with the <new-spec.md> placeholder resolved to new-spec.md in
// the command's working directory. A missing heading or an empty block fails
// the test rather than yielding a verb that silently checks nothing.
func auditVerb(t *testing.T, doc, heading string) string {
	t.Helper()

	raw, err := os.ReadFile(doc)
	if err != nil {
		t.Fatalf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: read %s: %v", doc, err)
	}
	var body []string
	inSection, inBlock := false, false
	for _, line := range strings.Split(string(raw), "\n") {
		switch {
		case strings.HasPrefix(line, heading):
			inSection = true
		case inSection && !inBlock && line == "```bash":
			inBlock = true
		case inBlock && line == "```":
			verb := strings.ReplaceAll(strings.Join(body, "\n"), "<new-spec.md>", "new-spec.md")
			if strings.TrimSpace(verb) == "" {
				t.Fatalf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: empty bash block under %q in %s", heading, doc)
			}
			return verb
		case inBlock:
			body = append(body, line)
		}
	}
	t.Fatalf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: no bash block under %q in %s", heading, doc)
	return ""
}

// runVerb writes the fixture files under a fresh directory and runs verb there.
// A nil newSpec leaves new-spec.md absent.
func runVerb(t *testing.T, verb string, newSpec *string, specs map[string]string) string {
	t.Helper()

	tmp := t.TempDir()
	for id, body := range specs {
		dir := filepath.Join(tmp, ".moai", "specs", id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", id, err)
		}
	}
	if newSpec != nil {
		if err := os.WriteFile(filepath.Join(tmp, "new-spec.md"), []byte(*newSpec), 0o644); err != nil {
			t.Fatalf("write new-spec.md: %v", err)
		}
	}

	cmd := exec.Command("bash", "-c", verb)
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: verb execution failed: %v\noutput: %s", err, out)
	}
	return string(out)
}

func d7Verb(t *testing.T) string {
	return auditVerb(t, planAuditorDocs["template"], "### Group 7:")
}

func d8Verb(t *testing.T) string {
	return auditVerb(t, planAuditorDocs["template"], "### Group 8:")
}

func specWithStatus(id, status string) string {
	return "---\nid: " + id + "\nstatus: " + status + "\n---\n\n# Referenced fixture\n"
}

func strPtr(s string) *string { return &s }

// hasFinding reports whether any output line starts with the finding marker
// (for example "BLOCKING:"). Matching the marker rather than the bare word is
// load-bearing: the D7 REVIEW line itself says "before emitting BLOCKING", so a
// substring check reads a review candidate as a BLOCKING finding.
func hasFinding(out, marker string) bool {
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, marker) {
			return true
		}
	}
	return false
}

// TestPlanAuditD7_RetiredSPECConflict: an unreconciled reference to a retired
// SPEC is surfaced for review — and the script itself does not decide BLOCKING.
func TestPlanAuditD7_RetiredSPECConflict(t *testing.T) {
	t.Parallel()

	out := runVerb(t, d7Verb(t), strPtr(`---
id: SPEC-FIXTURE-CONSUMER-001
status: draft
---

# Consumer SPEC

This SPEC depends on SPEC-FIXTURE-OLD-001 for its baseline behavior.
`), map[string]string{"SPEC-FIXTURE-OLD-001": specWithStatus("SPEC-FIXTURE-OLD-001", "retired")})
	t.Logf("D7 output:\n%s", out)

	if !strings.Contains(out, "REVIEW: SPEC-FIXTURE-OLD-001 has status=retired") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 did not surface the retired reference for review; got: %s", out)
	}
	if hasFinding(out, "BLOCKING:") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 script emitted BLOCKING itself; BLOCKING is the auditor's decision; got: %s", out)
	}
	if strings.Contains(out, "reconciliation candidate") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 offered a reconciliation candidate for a body with none; got: %s", out)
	}
}

// TestPlanAuditD7_ReconciledReferenceIsNotBlocking (AC-02, the false positive):
// an explicitly reconciled superseded reference must not come out BLOCKING, and
// the reconciling paragraph must be surfaced for the auditor to read.
func TestPlanAuditD7_ReconciledReferenceIsNotBlocking(t *testing.T) {
	t.Parallel()

	out := runVerb(t, d7Verb(t), strPtr(`---
id: SPEC-FIXTURE-CONSUMER-001
status: draft
---

# Consumer SPEC

## Reconciliation

This SPEC supersedes SPEC-FIXTURE-OLD-001: its eviction requirement is absorbed here.
`), map[string]string{"SPEC-FIXTURE-OLD-001": specWithStatus("SPEC-FIXTURE-OLD-001", "superseded")})
	t.Logf("D7 output:\n%s", out)

	if hasFinding(out, "BLOCKING:") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 emitted BLOCKING for an explicitly reconciled reference; got: %s", out)
	}
	if !strings.Contains(out, "reconciliation candidate") || !strings.Contains(out, "This SPEC supersedes SPEC-FIXTURE-OLD-001") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 did not surface the reconciling paragraph; got: %s", out)
	}
}

// TestPlanAuditD7_LiveReferenceIsSilent: a reference to a live SPEC produces no
// review line — the control showing D7 does not flag every reference.
func TestPlanAuditD7_LiveReferenceIsSilent(t *testing.T) {
	t.Parallel()

	out := runVerb(t, d7Verb(t), strPtr(`---
id: SPEC-FIXTURE-CONSUMER-001
status: draft
---

# Consumer SPEC

This SPEC extends SPEC-FIXTURE-OLD-001.
`), map[string]string{"SPEC-FIXTURE-OLD-001": specWithStatus("SPEC-FIXTURE-OLD-001", "implemented")})

	if hasFinding(out, "REVIEW:") || hasFinding(out, "BLOCKING:") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 flagged a reference to a live SPEC; got: %s", out)
	}
}

// TestPlanAuditD7_MissingReferencedSPEC: a reference to a SPEC absent from
// .moai/specs is reported at SHOULD severity and never as BLOCKING or REVIEW.
func TestPlanAuditD7_MissingReferencedSPEC(t *testing.T) {
	t.Parallel()

	out := runVerb(t, d7Verb(t), strPtr(`# Consumer SPEC

This SPEC references SPEC-FIXTURE-GHOST-001 which does not exist.
`), nil)

	if !strings.Contains(out, "SHOULD: referenced SPEC SPEC-FIXTURE-GHOST-001 not found") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 did not report the missing SPEC at SHOULD severity; got: %s", out)
	}
	if hasFinding(out, "BLOCKING:") || hasFinding(out, "REVIEW:") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 escalated a missing SPEC reference; got: %s", out)
	}
}

// TestPlanAuditD8_MissingBuildTag: a syscall with no build constraint anywhere
// is BLOCKING.
func TestPlanAuditD8_MissingBuildTag(t *testing.T) {
	t.Parallel()

	out := runVerb(t, d8Verb(t), strPtr(`# D8 Fixture — Missing Build Tag

This SPEC introduces a Go file that imports syscall.Flock to manage file locks.
`), nil)

	if !hasFinding(out, "BLOCKING:") || !strings.Contains(out, "//go:build") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D8 did not emit BLOCKING naming //go:build for a syscall without a constraint; got: %s", out)
	}
}

// TestPlanAuditD8_WithBuildTagPasses: a constraint in the same section as the
// syscall satisfies D8.
func TestPlanAuditD8_WithBuildTagPasses(t *testing.T) {
	t.Parallel()

	out := runVerb(t, d8Verb(t), strPtr("# D8 Fixture — With Build Tag\n\n"+
		"This SPEC introduces a Go file that imports syscall.Flock on POSIX.\n"+
		"The file is gated by `//go:build !windows` and a sibling stub uses `//go:build windows`.\n"), nil)

	if hasFinding(out, "BLOCKING:") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D8 emitted BLOCKING for a syscall whose own section carries //go:build; got: %s", out)
	}
}

// TestPlanAuditD8_BuildTagInOtherSectionDoesNotCover (AC-03, the false
// negative): a //go:build in one section must not satisfy D8 for a syscall in
// another section.
func TestPlanAuditD8_BuildTagInOtherSectionDoesNotCover(t *testing.T) {
	t.Parallel()

	out := runVerb(t, d8Verb(t), strPtr(`# File watcher

## 1. Existing lock helper

The existing helper file carries a `+"`//go:build !windows`"+` constraint and is unchanged.

## 2. New file watcher

The watcher calls syscall.Kqueue to register file events.
`), nil)

	if !strings.Contains(out, `BLOCKING: section "## 2. New file watcher"`) {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: a //go:build in another section masked the unconstrained syscall section; got: %s", out)
	}
	if strings.Contains(out, "## 1. Existing lock helper") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D8 flagged the section that carries no syscall; got: %s", out)
	}
}

// TestPlanAuditD8_UnreadableSpecIsGap: an input D8 cannot read is reported as a
// gap rather than passing silently.
func TestPlanAuditD8_UnreadableSpecIsGap(t *testing.T) {
	t.Parallel()

	out := runVerb(t, d8Verb(t), nil, nil)

	if !strings.Contains(out, "GAP:") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D8 did not report an unreadable input as GAP; got: %s", out)
	}
	if hasFinding(out, "BLOCKING:") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D8 emitted BLOCKING for an unreadable input; got: %s", out)
	}
}

// TestPlanAuditD7D8_VerbsIdenticalAcrossCopies: the local mirror carries the
// same verbs as the distributed template copy.
func TestPlanAuditD7D8_VerbsIdenticalAcrossCopies(t *testing.T) {
	t.Parallel()

	for _, heading := range []string{"### Group 7:", "### Group 8:"} {
		tmpl := auditVerb(t, planAuditorDocs["template"], heading)
		local := auditVerb(t, planAuditorDocs["local"], heading)
		if tmpl != local {
			t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: %s verb differs between the template and local copies", heading)
		}
	}
}
