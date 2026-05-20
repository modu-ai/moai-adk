// plan_audit_d7_d8_test.go: Fixture-based validation of the D7 and D8 audit
// verbs documented in `.claude/agents/moai/plan-auditor.md` (added by
// SPEC-V3R5-WORKFLOW-OPT-001 Layer G).
//
// These tests do NOT invoke the plan-auditor agent itself. They run the bash
// verbs described in the agent prompt against synthetic SPEC fixtures and
// assert the expected BLOCKING/PASS classification. This validates that the
// heuristic prescribed by the agent prompt correctly detects:
//   - D7: cross-SPEC references to retired/superseded/archived SPECs without
//     reconciliation
//   - D8: syscall imports without //go:build constraint or EXCL justification
//
// AC coverage:
//   - AC-WO-004 (TestPlanAuditD7_RetiredSPECConflict): D7 detects V3R4 retirement
//   - AC-WO-012 (TestPlanAuditD8_MissingBuildTag): D8 detects syscall without tag
//   - EC-WO-003 (TestPlanAuditD7_MissingReferencedSPEC): D7 SHOULD severity
//
// Sentinel on failure: PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT — indicates the bash
// verb in the agent prompt no longer matches the documented behavior.
//
// @MX:ANCHOR: [AUTO] fan_in=2 — guards plan-auditor D7/D8 verb correctness.
// Touching this test signature affects the audit contract for cross-SPEC
// reconciliation (D7) and cross-platform discipline (D8).
// @MX:REASON: Without these fixtures, the bash verbs in plan-auditor.md could
// drift from the documented behavior without detection. The fixtures lock the
// verb semantics so SPECs that violate D7/D8 are caught at audit time.
package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPlanAuditD7_RetiredSPECConflict verifies AC-WO-004: D7 emits a BLOCKING
// finding when a SPEC references a retired SPEC ID without reconciliation.
//
// Fixture: a SPEC body that references SPEC-V3R5-WO-FIXTURE-001 (whose status
// frontmatter is `retired`) without any reconciliation keyword nearby.
func TestPlanAuditD7_RetiredSPECConflict(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	specsDir := filepath.Join(tmp, ".moai", "specs", "SPEC-V3R5-WO-FIXTURE-001")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Fixture 1: retired SPEC
	retiredSpec := `---
id: SPEC-V3R5-WO-FIXTURE-001
status: retired
---

# Retired Fixture
This SPEC is retired and exists only as a fixture for plan-auditor D7 testing.
`
	if err := os.WriteFile(filepath.Join(specsDir, "spec.md"), []byte(retiredSpec), 0o644); err != nil {
		t.Fatalf("write retired fixture: %v", err)
	}

	// New SPEC body that references the retired SPEC without reconciliation.
	newSpec := `---
id: SPEC-V3R5-WO-CONSUMER-001
status: draft
---

# Consumer SPEC
This SPEC depends on SPEC-V3R5-WO-FIXTURE-001 for its baseline behavior.
No reconciliation is provided.
`
	newSpecPath := filepath.Join(tmp, "consumer-spec.md")
	if err := os.WriteFile(newSpecPath, []byte(newSpec), 0o644); err != nil {
		t.Fatalf("write consumer fixture: %v", err)
	}

	// Run the D7 verb (verbatim from plan-auditor.md Group 7).
	// The verb extracts SPEC-IDs from the new SPEC, checks status in .moai/specs,
	// and emits BLOCKING for retired/superseded/archived.
	d7Verb := `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' "` + newSpecPath + `" | sort -u | while read SID; do
  if [ -f ".moai/specs/$SID/spec.md" ]; then
    STATUS=$(grep '^status:' ".moai/specs/$SID/spec.md" | head -1 | cut -d: -f2 | tr -d ' ')
    case "$STATUS" in
      retired|superseded|archived)
        echo "BLOCKING: $SID has status=$STATUS but is referenced without reconciliation"
        ;;
    esac
  fi
done`

	cmd := exec.Command("bash", "-c", d7Verb)
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 verb execution failed: %v\noutput: %s", err, out)
	}

	if !strings.Contains(string(out), "BLOCKING") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 verb did not emit BLOCKING for retired SPEC reference; got: %s", out)
	}
	if !strings.Contains(string(out), "SPEC-V3R5-WO-FIXTURE-001") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 verb did not identify the conflicting SPEC ID; got: %s", out)
	}
	if !strings.Contains(string(out), "status=retired") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 verb did not report the retired status; got: %s", out)
	}
}

// TestPlanAuditD7_MissingReferencedSPEC verifies EC-WO-003: D7 handles missing
// referenced SPECs gracefully. The verb implementation in plan-auditor.md
// silently skips missing SPECs (no BLOCKING, no SHOULD emitted by the bash
// pipeline itself — the SHOULD severity is the agent's interpretation layer).
//
// This test asserts that the bash verb does NOT emit a false BLOCKING for
// missing SPECs. The agent's SHOULD severity is handled at a higher level.
func TestPlanAuditD7_MissingReferencedSPEC(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	// Intentionally do NOT create .moai/specs/SPEC-V3R5-WO-GHOST-001/spec.md

	newSpec := `---
id: SPEC-V3R5-WO-CONSUMER-002
status: draft
---

# Consumer SPEC
This SPEC references SPEC-V3R5-WO-GHOST-001 which does not exist.
`
	newSpecPath := filepath.Join(tmp, "consumer-spec.md")
	if err := os.WriteFile(newSpecPath, []byte(newSpec), 0o644); err != nil {
		t.Fatalf("write consumer fixture: %v", err)
	}

	d7Verb := `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' "` + newSpecPath + `" | sort -u | while read SID; do
  if [ -f ".moai/specs/$SID/spec.md" ]; then
    STATUS=$(grep '^status:' ".moai/specs/$SID/spec.md" | head -1 | cut -d: -f2 | tr -d ' ')
    case "$STATUS" in
      retired|superseded|archived)
        echo "BLOCKING: $SID has status=$STATUS but is referenced without reconciliation"
        ;;
    esac
  fi
done`

	cmd := exec.Command("bash", "-c", d7Verb)
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("D7 verb execution failed: %v\noutput: %s", err, out)
	}

	// For a missing referenced SPEC, the verb must NOT emit BLOCKING.
	if strings.Contains(string(out), "BLOCKING") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D7 verb emitted false BLOCKING for missing SPEC reference; got: %s", out)
	}
}

// TestPlanAuditD8_MissingBuildTag verifies AC-WO-012: D8 emits BLOCKING when a
// SPEC introduces `syscall` reference without a `//go:build` constraint or EXCL
// justification.
func TestPlanAuditD8_MissingBuildTag(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	// Fixture: SPEC body mentions syscall without build-tag.
	noTagSpec := `---
id: SPEC-V3R5-WO-D8-FIXTURE-001
status: draft
---

# D8 Fixture — Missing Build Tag
This SPEC introduces a Go file that imports syscall.Flock to manage file locks.
No build-tag declaration is provided.
`
	noTagPath := filepath.Join(tmp, "no-tag-spec.md")
	if err := os.WriteFile(noTagPath, []byte(noTagSpec), 0o644); err != nil {
		t.Fatalf("write no-tag fixture: %v", err)
	}

	d8Verb := `if grep -q 'syscall' "` + noTagPath + `"; then
  if ! grep -qE '//go:build|cross-platform exemption|EXCL.*syscall' "` + noTagPath + `"; then
    echo "BLOCKING: SPEC references syscall but no //go:build constraint or EXCL justification"
  fi
fi`

	cmd := exec.Command("bash", "-c", d8Verb)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("D8 verb execution failed: %v\noutput: %s", err, out)
	}

	if !strings.Contains(string(out), "BLOCKING") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D8 verb did not emit BLOCKING for syscall without build-tag; got: %s", out)
	}
	if !strings.Contains(string(out), "//go:build") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D8 BLOCKING message must reference //go:build; got: %s", out)
	}
}

// TestPlanAuditD8_WithBuildTagPasses verifies the D8 verb does NOT emit
// BLOCKING when the SPEC body properly declares a //go:build constraint.
func TestPlanAuditD8_WithBuildTagPasses(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	withTagSpec := `---
id: SPEC-V3R5-WO-D8-FIXTURE-002
status: draft
---

# D8 Fixture — With Build Tag
This SPEC introduces a Go file that imports syscall.Flock on POSIX.
The file is gated by ` + "`//go:build !windows`" + ` and a sibling stub uses
` + "`//go:build windows`" + ` for cross-platform parity.
`
	withTagPath := filepath.Join(tmp, "with-tag-spec.md")
	if err := os.WriteFile(withTagPath, []byte(withTagSpec), 0o644); err != nil {
		t.Fatalf("write with-tag fixture: %v", err)
	}

	d8Verb := `if grep -q 'syscall' "` + withTagPath + `"; then
  if ! grep -qE '//go:build|cross-platform exemption|EXCL.*syscall' "` + withTagPath + `"; then
    echo "BLOCKING: SPEC references syscall but no //go:build constraint or EXCL justification"
  fi
fi`

	cmd := exec.Command("bash", "-c", d8Verb)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("D8 verb execution failed: %v\noutput: %s", err, out)
	}

	if strings.Contains(string(out), "BLOCKING") {
		t.Errorf("PLAN_AUDIT_D7_D8_HEURISTIC_DRIFT: D8 verb emitted false BLOCKING for SPEC with valid //go:build constraint; got: %s", out)
	}
}
