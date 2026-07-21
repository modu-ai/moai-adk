package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// specMDMissingOutOfScope is a parseable spec.md whose 'Out of Scope'
// subsection is absent (triggers MissingExclusions) and whose frontmatter
// omits the 'module' field (triggers FrontmatterInvalid).
const specMDMissingOutOfScope = `---
id: SPEC-GF-001
title: "Grandfathered era fixture"
version: "0.1.0"
status: draft
created: 2026-01-15
updated: 2026-01-15
author: Test Author
priority: P2 Medium
phase: "v2.0.0"
dependencies: []
bc_id: []
lifecycle: spec-anchored
tags: "test"
breaking: false
related_rule: []
---

# SPEC-GF-001: Grandfathered era fixture

## 2. Scope

### 2.1 In Scope

- Everything is in scope

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-GFX-001-001: The system SHALL do thing A.

## 6. Acceptance Criteria

- AC-GFX-001-01: Given A, When something, Then result. (maps REQ-GFX-001-001)
`

// modernProgressMD carries the V3R6 era markers (H-4: §E.2 + §E.4 +
// sync_commit_sha) so ClassifyEra returns EraV3R6 (modern, not grandfathered).
const modernProgressMD = "# Progress\n\n## §E.2 Run-phase Evidence\n\n## §E.4 Sync-phase Audit-Ready Signal\n\nsync_commit_sha: abc1234\n"

func writeEraFixture(t *testing.T, withProgress bool) string {
	t.Helper()
	dir := t.TempDir()
	specDir := filepath.Join(dir, "SPEC-GF-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMDMissingOutOfScope), 0o644); err != nil {
		t.Fatal(err)
	}
	if withProgress {
		if err := os.WriteFile(filepath.Join(specDir, "progress.md"), []byte(modernProgressMD), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(specDir, "spec.md")
}

func findByCode(findings []Finding, code string) []Finding {
	var out []Finding
	for _, f := range findings {
		if f.Code == code {
			out = append(out, f)
		}
	}
	return out
}

// A grandfather-era SPEC (no progress.md → V2.x per H-1) must have its
// structural rules (MissingExclusions, FrontmatterInvalid) downgraded from
// ERROR to advisory WARNING, and --strict must not escalate them back.
func TestLint_GrandfatheredSPEC_StructuralRulesDowngraded(t *testing.T) {
	t.Parallel()

	specPath := writeEraFixture(t, false)
	linter := NewLinter(LinterOptions{Strict: true})
	report, err := linter.Lint([]string{specPath})
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	for _, code := range []string{"MissingExclusions", "FrontmatterInvalid"} {
		got := findByCode(report.Findings, code)
		if len(got) == 0 {
			t.Fatalf("expected %s finding to remain visible, got none", code)
		}
		for _, f := range got {
			if f.Severity != SeverityWarning {
				t.Errorf("%s severity = %q, want %q (grandfathered downgrade)", code, f.Severity, SeverityWarning)
			}
			if !f.Advisory {
				t.Errorf("%s Advisory = false, want true (must not escalate under --strict)", code)
			}
		}
	}

	if report.HasErrors() {
		t.Error("HasErrors() = true under --strict, want false: grandfathered structural findings must not fail the lint")
	}
}

// A modern-era SPEC (V3R6 per H-4) keeps full ERROR enforcement.
func TestLint_ModernSPEC_StructuralRulesRemainError(t *testing.T) {
	t.Parallel()

	specPath := writeEraFixture(t, true)
	linter := NewLinter(LinterOptions{Strict: true})
	report, err := linter.Lint([]string{specPath})
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	got := findByCode(report.Findings, "MissingExclusions")
	if len(got) == 0 {
		t.Fatal("expected MissingExclusions finding, got none")
	}
	for _, f := range got {
		if f.Severity != SeverityError {
			t.Errorf("MissingExclusions severity = %q, want %q (modern era keeps enforcement)", f.Severity, SeverityError)
		}
	}
	if !report.HasErrors() {
		t.Error("HasErrors() = false, want true for modern-era structural error")
	}
}

// A modern-era SPEC in a terminal lifecycle status (completed) is closed
// history: structural rules are demoted to advisory warnings just like the
// grandfather clause.
func TestLint_TerminalStatusSPEC_StructuralRulesDowngraded(t *testing.T) {
	t.Parallel()

	specPath := writeEraFixture(t, true)
	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	updated := []byte(replaceStatus(string(content), "completed"))
	if err := os.WriteFile(specPath, updated, 0o644); err != nil {
		t.Fatal(err)
	}

	linter := NewLinter(LinterOptions{Strict: true})
	report, err := linter.Lint([]string{specPath})
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}
	got := findByCode(report.Findings, "MissingExclusions")
	if len(got) == 0 {
		t.Fatal("expected MissingExclusions finding, got none")
	}
	for _, f := range got {
		if f.Severity != SeverityWarning || !f.Advisory {
			t.Errorf("terminal-status SPEC: severity=%q advisory=%v, want warning/advisory", f.Severity, f.Advisory)
		}
	}
	if report.HasErrors() {
		t.Error("HasErrors() = true, want false for terminal-status SPEC under --strict")
	}
}

func replaceStatus(content, status string) string {
	return strings.Replace(content, "status: draft", "status: "+status, 1)
}

// Advisory warnings never escalate under --strict; plain warnings still do.
func TestReport_HasErrors_AdvisoryWarningNotEscalated(t *testing.T) {
	t.Parallel()

	advisory := &Report{
		Strict:   true,
		Findings: []Finding{{Severity: SeverityWarning, Code: "StatusGitConsistency", Advisory: true}},
	}
	if advisory.HasErrors() {
		t.Error("advisory warning escalated under strict; want no escalation")
	}

	plain := &Report{
		Strict:   true,
		Findings: []Finding{{Severity: SeverityWarning, Code: "SomeRule"}},
	}
	if !plain.HasErrors() {
		t.Error("non-advisory warning not escalated under strict; want escalation")
	}
}
