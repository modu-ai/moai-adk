// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M2 — `moai spec audit` CLI integration tests.
//
// These tests cover the CLI surface only (flag parsing, help text, JSON output
// schema, filter behavior). The backing audit engine lives in
// internal/spec/audit.go (M1) and is exercised by audit_test.go.
//
// AC coverage (M2 scope):
//   - AC-LSG-002  — era classification 5 buckets surfaced through CLI
//   - AC-LSG-007  — JSON output schema (audited_at, total_specs, grandfathered,
//                   modern_era_clean, drift_findings)
//   - AC-LSG-016  — NFR performance (verified via dedicated benchmark in M1;
//                   M2 wires the CLI surface invocation path)
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureV3R6CleanSpec writes a SPEC fixture classified as V3R6 with no drift:
// status: completed, §E.2 + §E.5 sections + valid SHAs all present.
func fixtureV3R6CleanSpec(t *testing.T, baseDir, specID string) {
	t.Helper()
	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec: %v", err)
	}

	specMD := `---
id: ` + specID + `
title: Fixture V3R6 clean
version: 0.1.0
status: completed
era: V3R6
created: 2026-01-01
updated: 2026-01-01
author: Test
priority: P1
phase: run
module: TEST
lifecycle: spec-anchored
tags: test
---
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	progressMD := `---
id: ` + specID + `
artifact: progress
status: completed
---

## §E.2 Sync-phase Audit-Ready Signal

sync_commit_sha: abc1234567890

## §E.3 Run-phase Status

status: completed

## §E.5 Mx-phase Audit-Ready Signal

mx_commit_sha: def1234567890
`
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressMD), 0o644); err != nil {
		t.Fatalf("write progress.md: %v", err)
	}
}

// fixtureV3R6DriftSpec writes a SPEC at the Y_Y_Y_Y_StatusDrift fixture state:
// §E.2 + §E.5 + SHAs all present, but spec.md status = implemented (drift).
func fixtureV3R6DriftSpec(t *testing.T, baseDir, specID string) {
	t.Helper()
	specDir := filepath.Join(baseDir, ".moai", "specs", specID)
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec: %v", err)
	}

	specMD := `---
id: ` + specID + `
title: Fixture V3R6 drift
version: 0.1.0
status: implemented
era: V3R6
created: 2026-01-01
updated: 2026-01-01
author: Test
priority: P1
phase: run
module: TEST
lifecycle: spec-anchored
tags: test
---
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	progressMD := `---
id: ` + specID + `
artifact: progress
status: in-progress
---

## §E.2 Sync-phase Audit-Ready Signal

sync_commit_sha: abc1234567890

## §E.5 Mx-phase Audit-Ready Signal

mx_commit_sha: def1234567890
`
	if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(progressMD), 0o644); err != nil {
		t.Fatalf("write progress.md: %v", err)
	}
}

// runSpecAuditCmd invokes the `spec audit` cobra command in-process with the
// given args, captures stdout/stderr, and returns (stdout, stderr, exit-error).
func runSpecAuditCmd(t *testing.T, baseDir string, args ...string) (string, string, error) {
	t.Helper()

	cmd := newSpecAuditCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	fullArgs := append([]string{"--base-dir", baseDir}, args...)
	cmd.SetArgs(fullArgs)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

// TestSpecAudit_Help — `moai spec audit --help` must list the documented flags.
func TestSpecAudit_Help(t *testing.T) {
	cmd := newSpecAuditCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should not error: %v", err)
	}

	out := stdout.String()
	for _, want := range []string{"--json", "--filter-era", "--include-grandfathered"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q\n----\n%s", want, out)
		}
	}
}

// TestSpecAudit_EmptyProject — AC-LSG-007 happy path on empty fixture.
// JSON output must parse and contain all required schema fields with sensible
// zero-values when no SPECs exist.
func TestSpecAudit_EmptyProject_JSONSchema(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, stderr, err := runSpecAuditCmd(t, tmpDir, "--json")
	if err != nil {
		t.Fatalf("audit on empty project should succeed; err=%v stderr=%q", err, stderr)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout must be valid JSON; err=%v stdout=%q", err, stdout)
	}

	// AC-LSG-007 schema fields
	for _, field := range []string{"audited_at", "total_specs", "grandfathered", "modern_era_clean", "drift_findings"} {
		if _, ok := got[field]; !ok {
			t.Errorf("JSON output missing required field %q; got keys: %v", field, mapKeys(got))
		}
	}
}

// TestSpecAudit_JSONSchema_DriftFindings — AC-LSG-007 + AC-LSG-009. Fixture
// with a Y_Y_Y_Y_StatusDrift SPEC must surface a drift_findings entry with the
// documented schema.
func TestSpecAudit_JSONSchema_DriftFindings(t *testing.T) {
	tmpDir := t.TempDir()
	specID := "SPEC-V3R6-AUDIT-DRIFT-001"
	fixtureV3R6DriftSpec(t, tmpDir, specID)

	stdout, stderr, err := runSpecAuditCmd(t, tmpDir, "--json")
	if err != nil {
		t.Fatalf("audit should succeed; err=%v stderr=%q", err, stderr)
	}

	var got struct {
		TotalSpecs    int `json:"total_specs"`
		DriftFindings []struct {
			SpecID      string `json:"spec_id"`
			Era         string `json:"era"`
			FindingType string `json:"finding_type"`
			Severity    string `json:"severity"`
			Remediation string `json:"remediation"`
		} `json:"drift_findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("invalid JSON: %v\nstdout=%q", err, stdout)
	}
	if got.TotalSpecs != 1 {
		t.Errorf("total_specs = %d, want 1", got.TotalSpecs)
	}
	if len(got.DriftFindings) == 0 {
		t.Fatalf("expected at least 1 drift finding, got 0; stdout=%q", stdout)
	}

	var found bool
	for _, f := range got.DriftFindings {
		if f.FindingType == "Y_Y_Y_Y_StatusDrift" && f.Severity == "MUST-FIX" {
			found = true
			if !strings.Contains(f.Remediation, "moai spec close") {
				t.Errorf("remediation should reference `moai spec close`; got %q", f.Remediation)
			}
		}
	}
	if !found {
		t.Errorf("expected drift_findings to contain Y_Y_Y_Y_StatusDrift MUST-FIX entry; got %+v", got.DriftFindings)
	}
}

// TestSpecAudit_FilterEra — filter restricts drift_findings to the named era.
func TestSpecAudit_FilterEra(t *testing.T) {
	tmpDir := t.TempDir()
	fixtureV3R6DriftSpec(t, tmpDir, "SPEC-V3R6-AUDIT-FILTER-001")
	fixtureV3R6CleanSpec(t, tmpDir, "SPEC-V3R6-AUDIT-CLEAN-002")

	stdout, stderr, err := runSpecAuditCmd(t, tmpDir, "--json", "--filter-era=V3R6")
	if err != nil {
		t.Fatalf("audit should succeed; err=%v stderr=%q", err, stderr)
	}

	var got struct {
		DriftFindings []struct {
			SpecID string `json:"spec_id"`
			Era    string `json:"era"`
		} `json:"drift_findings"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	for _, f := range got.DriftFindings {
		if f.Era != "" && f.Era != "V3R6" {
			t.Errorf("filter-era=V3R6 should drop non-V3R6 entries; got era=%q for %s",
				f.Era, f.SpecID)
		}
	}
}

// TestSpecAudit_HumanReadableDefault — without --json the CLI must produce
// human-readable text output (not raw JSON).
func TestSpecAudit_HumanReadableDefault(t *testing.T) {
	tmpDir := t.TempDir()
	fixtureV3R6DriftSpec(t, tmpDir, "SPEC-V3R6-AUDIT-HUMAN-001")

	stdout, _, err := runSpecAuditCmd(t, tmpDir)
	if err != nil {
		t.Fatalf("audit should succeed; err=%v", err)
	}

	// Output should NOT be JSON (no leading `{`)
	trimmed := strings.TrimSpace(stdout)
	if strings.HasPrefix(trimmed, "{") {
		t.Errorf("default output should be human-readable, not JSON; got: %q", stdout)
	}
}

// TestSpecAudit_StrictMode_ExitsNonZeroOnDrift — --strict + MUST-FIX drift → error.
// Verifies the renderAuditResult strict-mode escalation path. Default (non-strict)
// invocations exit 0 even with drift; --strict surfaces MUST-FIX findings as
// a non-zero exit, used by CI gates.
func TestSpecAudit_StrictMode_ExitsNonZeroOnDrift(t *testing.T) {
	tmpDir := t.TempDir()
	specID := "SPEC-V3R6-AUDIT-STRICT-001"
	fixtureV3R6DriftSpec(t, tmpDir, specID)

	_, _, err := runSpecAuditCmd(t, tmpDir, "--strict")
	if err == nil {
		t.Errorf("--strict with MUST-FIX drift should return error; got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "strict") {
		t.Errorf("strict-mode error should mention 'strict'; got: %v", err)
	}
}

// TestSpecAudit_StrictMode_CleanFixture_ExitsZero — --strict on a clean fixture
// must exit 0 (no MUST-FIX findings → strict path is a no-op).
func TestSpecAudit_StrictMode_CleanFixture_ExitsZero(t *testing.T) {
	tmpDir := t.TempDir()
	fixtureV3R6CleanSpec(t, tmpDir, "SPEC-V3R6-AUDIT-CLEAN-001")

	_, _, err := runSpecAuditCmd(t, tmpDir, "--strict", "--json")
	if err != nil {
		t.Errorf("--strict on clean fixture should exit 0; got err=%v", err)
	}
}

// TestSpecAudit_IncludeGrandfathered — when --include-grandfathered is passed,
// the audit emits INFO findings for pre-V3R6 SPECs (otherwise excluded).
func TestSpecAudit_IncludeGrandfathered(t *testing.T) {
	tmpDir := t.TempDir()
	// Use V3R6 drift spec — when --include-grandfathered is enabled, the era
	// classification still distinguishes V3R6 from grandfathered eras; the
	// flag is a no-op for V3R6 fixtures but exercises the boolean wiring.
	fixtureV3R6CleanSpec(t, tmpDir, "SPEC-V3R6-AUDIT-INCL-GFD-001")

	_, _, err := runSpecAuditCmd(t, tmpDir, "--include-grandfathered", "--json")
	if err != nil {
		t.Errorf("--include-grandfathered should not error; got: %v", err)
	}
}

// TestSpecAudit_NoAskUserQuestion — REQ-PGN-012 / C-HRA-008 static guard.
// Uses the same line-comment-stripping pattern as TestSpecClose_NoAskUserQuestion
// so that doctrine documentation referring to AskUserQuestion by name does not
// produce a false positive.
func TestSpecAudit_NoAskUserQuestion(t *testing.T) {
	src, err := os.ReadFile("spec_audit.go")
	if err != nil {
		t.Skipf("spec_audit.go not present (RED phase): %v", err)
	}
	if hasNonCommentRefToAskUser(string(src)) {
		t.Errorf("spec_audit.go must not reference AskUserQuestion or mcp__askuser outside comments")
	}
}

// mapKeys returns the keys of map[string]any as a string slice. Test helper.
func mapKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
