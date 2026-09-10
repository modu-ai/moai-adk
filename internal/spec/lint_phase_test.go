package spec

import (
	"os"
	"strings"
	"testing"
)

// Selector caveats for anyone running these tests from an acceptance command:
//
//  1. `go test -run '^TestPhaseValueShape' -list '.*'` does NOT honour the -run
//     selector. Go's -list overrides -run and prints every test in the package,
//     so the `grep -c '^TestPhaseValueShape'` that follows such a command is what
//     actually filters. Reading the -run argument as proof the selector matched
//     is a mistake — verify via the grep count, not the command's shape.
//  2. A `-run 'Leak|Neutral'` style selector over ./internal/template does NOT
//     reach TestManifestHashFormat, which is what catches a catalog.yaml hash
//     invalidated by editing a catalogued template file. Editing a template
//     requires the full package run, not the filtered one.
//
// Both are selector-scope defects, not test defects: a passing filtered run is
// not evidence that the package is green.

// phaseTestFrontmatter builds a complete, otherwise-valid 12-field frontmatter so
// that the only variable under test is the phase value. Every other field is
// populated to keep the required-field branch of FrontmatterSchemaRule silent.
func phaseTestFrontmatter(phase string) SPECFrontmatter {
	return SPECFrontmatter{
		ID:        "SPEC-PFV-001",
		Title:     "Phase value shape",
		Version:   "0.1.0",
		Status:    "draft",
		Created:   "2026-08-02",
		Updated:   "2026-08-02",
		Author:    "Author Name",
		Priority:  "P2",
		Phase:     phase,
		Module:    "internal/spec",
		Lifecycle: "spec-anchored",
		Tags:      "spec-lint, frontmatter, phase",
	}
}

// phaseFindingsOf filters findings down to a single code.
func phaseFindingsOf(findings []Finding, code string) []Finding {
	var out []Finding
	for _, f := range findings {
		if f.Code == code {
			out = append(out, f)
		}
	}
	return out
}

// TestPhaseValueShape_WorkflowStageTokenRejected verifies that a phase value which
// is exactly a workflow-stage token emits one FrontmatterPhaseInvalid finding at
// error severity. The phase field is schema-defined as a release target, not a
// lifecycle-stage field.
func TestPhaseValueShape_WorkflowStageTokenRejected(t *testing.T) {
	cases := []struct {
		name  string
		phase string
	}{
		{"plan", "plan"},
		{"run", "run"},
		{"sync", "sync"},
		{"mx", "mx"},
		{"uppercase", "PLAN"},
		{"mixed case", "Sync"},
		{"surrounding whitespace", "  run  "},
	}

	rule := &FrontmatterSchemaRule{}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			doc := &SPECDoc{Path: "spec.md", Frontmatter: phaseTestFrontmatter(tt.phase)}

			phaseFindings := phaseFindingsOf(rule.Check(doc, nil), "FrontmatterPhaseInvalid")
			if len(phaseFindings) != 1 {
				t.Fatalf("phase %q: expected exactly 1 FrontmatterPhaseInvalid finding, got %d: %v",
					tt.phase, len(phaseFindings), phaseFindings)
			}
			if phaseFindings[0].Severity != SeverityError {
				t.Errorf("phase %q: expected severity %q, got %q",
					tt.phase, SeverityError, phaseFindings[0].Severity)
			}
			if phaseFindings[0].Advisory {
				t.Errorf("phase %q: finding must not be advisory at emission time", tt.phase)
			}
			if !strings.Contains(phaseFindings[0].Message, tt.phase) {
				t.Errorf("phase %q: message must quote the offending value, got %q",
					tt.phase, phaseFindings[0].Message)
			}
		})
	}
}

// TestPhaseValueShape_LegitimateValuesAccepted verifies the exact-match predicate does
// not degrade into substring containment. The legacy corpus carries release-target
// values that contain "Run" inside "Runtime"; those must not be flagged.
func TestPhaseValueShape_LegitimateValuesAccepted(t *testing.T) {
	cases := []string{
		"v3.0.2",
		"v3.0.0",
		"v3.0.0 — Phase 2 — Runtime Hardening",
		"v3.0.0 — Phase 7 — Agent Runtime Robustness",
		"v3.0.0 R3 — Phase A — Runtime Safety Net",
		"v3.0.0 R2 — Runtime Protocol Migration",
	}

	rule := &FrontmatterSchemaRule{}
	for _, phase := range cases {
		t.Run(phase, func(t *testing.T) {
			doc := &SPECDoc{Path: "spec.md", Frontmatter: phaseTestFrontmatter(phase)}

			phaseFindings := phaseFindingsOf(rule.Check(doc, nil), "FrontmatterPhaseInvalid")
			if len(phaseFindings) != 0 {
				t.Fatalf("phase %q is a legitimate release target: expected 0 FrontmatterPhaseInvalid findings, got %d: %v",
					phase, len(phaseFindings), phaseFindings)
			}
		})
	}
}

// TestPhaseValueShape_EmptyPhaseEmitsOnlyRequiredFieldFinding verifies the value-shape
// check runs after the required-field emptiness check, so an empty phase produces the
// existing required-field finding once and no duplicate value-shape finding.
func TestPhaseValueShape_EmptyPhaseEmitsOnlyRequiredFieldFinding(t *testing.T) {
	rule := &FrontmatterSchemaRule{}
	for _, phase := range []string{"", "   "} {
		doc := &SPECDoc{Path: "spec.md", Frontmatter: phaseTestFrontmatter(phase)}
		findings := rule.Check(doc, nil)

		if got := len(phaseFindingsOf(findings, "FrontmatterPhaseInvalid")); got != 0 {
			t.Errorf("empty phase %q: expected 0 FrontmatterPhaseInvalid findings, got %d", phase, got)
		}

		required := phaseFindingsOf(findings, "FrontmatterInvalid")
		if len(required) != 1 {
			t.Fatalf("empty phase %q: expected exactly 1 FrontmatterInvalid finding, got %d: %v",
				phase, len(required), required)
		}
		if !strings.Contains(required[0].Message, "phase") {
			t.Errorf("empty phase %q: required-field finding should name phase, got %q",
				phase, required[0].Message)
		}
	}
}

// writePhaseFixture builds an era fixture and rewrites its phase to the given
// value. withProgress=false → no progress.md → grandfather era (H-1, V2.x);
// withProgress=true → the V3R6 markers → modern era.
func writePhaseFixture(t *testing.T, withProgress bool, phase, status string) string {
	t.Helper()
	specPath := writeEraFixture(t, withProgress)

	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(content), `phase: "v2.0.0"`, "phase: "+phase, 1)
	if updated == string(content) {
		t.Fatal("phase line not found in fixture — the fixture's phase literal changed")
	}
	if status != "" {
		updated = replaceStatus(updated, status)
	}
	if err := os.WriteFile(specPath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
	return specPath
}

// assertPhaseFindingSurvivesDemotion asserts the contrast that makes the design
// decision observable: in ONE report over ONE SPEC, the era-demotable structural
// code is downgraded to an advisory warning while FrontmatterPhaseInvalid stays a
// non-advisory error and still gates the exit code. Asserting only the phase
// finding would not prove the demotion path ran at all.
func assertPhaseFindingSurvivesDemotion(t *testing.T, specPath, branch string) {
	t.Helper()

	linter := NewLinter(LinterOptions{Strict: true})
	report, err := linter.Lint([]string{specPath})
	if err != nil {
		t.Fatalf("%s: Lint failed: %v", branch, err)
	}

	demoted := findByCode(report.Findings, "MissingExclusions")
	if len(demoted) == 0 {
		t.Fatalf("%s: expected a MissingExclusions finding as the demotion control, got none", branch)
	}
	for _, f := range demoted {
		if f.Severity != SeverityWarning || !f.Advisory {
			t.Fatalf("%s: control finding MissingExclusions severity=%q advisory=%v, want warning/advisory — the demotion path did not run, so this fixture cannot test demotion bypass",
				branch, f.Severity, f.Advisory)
		}
	}

	phaseFindings := findByCode(report.Findings, "FrontmatterPhaseInvalid")
	if len(phaseFindings) != 1 {
		t.Fatalf("%s: expected exactly 1 FrontmatterPhaseInvalid finding, got %d: %v",
			branch, len(phaseFindings), phaseFindings)
	}
	if phaseFindings[0].Severity != SeverityError {
		t.Errorf("%s: FrontmatterPhaseInvalid severity = %q, want %q — it must bypass demotion",
			branch, phaseFindings[0].Severity, SeverityError)
	}
	if phaseFindings[0].Advisory {
		t.Errorf("%s: FrontmatterPhaseInvalid Advisory = true, want false — it must not be marked advisory",
			branch)
	}
	if !report.HasErrors() {
		t.Errorf("%s: HasErrors() = false, want true — the surviving error must gate the lint exit code", branch)
	}
}

// TestPhaseValueShape_GrandfatheredEraBypassesDemotion pins the era branch: a
// grandfather-era SPEC still reports a workflow-stage phase as a hard error.
func TestPhaseValueShape_GrandfatheredEraBypassesDemotion(t *testing.T) {
	t.Parallel()
	specPath := writePhaseFixture(t, false, "plan", "")
	assertPhaseFindingSurvivesDemotion(t, specPath, "grandfather era")
}

// TestPhaseValueShape_TerminalStatusBypassesDemotion pins the terminal branch: a
// modern-era SPEC in a closed lifecycle status also still reports a hard error.
func TestPhaseValueShape_TerminalStatusBypassesDemotion(t *testing.T) {
	t.Parallel()
	specPath := writePhaseFixture(t, true, "sync", "completed")
	assertPhaseFindingSurvivesDemotion(t, specPath, "terminal status")
}

// TestPhaseValueShape_NotEraDemotable verifies the new code is absent from the
// era-demotion set, so the finding survives as an error on grandfather-era SPECs.
// This is the design decision the SPEC exists for: the guard must fire at authoring
// time, when almost every in-flight SPEC classifies as grandfathered.
func TestPhaseValueShape_NotEraDemotable(t *testing.T) {
	if eraDemotableCodes["FrontmatterPhaseInvalid"] {
		t.Error("FrontmatterPhaseInvalid must NOT be registered in eraDemotableCodes")
	}
	if !eraDemotableCodes["FrontmatterInvalid"] {
		t.Error("FrontmatterInvalid must remain registered in eraDemotableCodes")
	}

	in := []Finding{
		{Code: "FrontmatterPhaseInvalid", Severity: SeverityError, Message: "m"},
		{Code: "FrontmatterInvalid", Severity: SeverityError, Message: "m"},
	}
	out := applyEraDemotion(in, demotionCause{GrandfatheredEra: true})

	if out[0].Severity != SeverityError || out[0].Advisory {
		t.Errorf("FrontmatterPhaseInvalid must survive era demotion as a non-advisory error, got severity=%q advisory=%v",
			out[0].Severity, out[0].Advisory)
	}
	if out[1].Severity != SeverityWarning || !out[1].Advisory {
		t.Errorf("FrontmatterInvalid must still demote to an advisory warning, got severity=%q advisory=%v",
			out[1].Severity, out[1].Advisory)
	}
}
