package spec_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// testdataDir is the directory containing fixture SPEC files.
const testdataDir = "testdata"

// registryPath is the zone registry path used by tests.
// Tests use a nil registry when the actual zone registry cannot be found.
func testRegistryPath() string {
	// Find the zone registry from the worktree root
	dir := "../../.claude/rules/moai/core/zone-registry.md"
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	return ""
}

// specPath returns the path to a specific fixture under testdata.
func specPath(fixture string) string {
	return filepath.Join(testdataDir, fixture, "spec.md")
}

// containsCode checks whether the findings slice contains the given code.
func containsCode(findings []spec.Finding, code string) bool {
	for _, f := range findings {
		if f.Code == code {
			return true
		}
	}
	return false
}

// findingsForCode returns the findings for the given code.
func findingsForCode(findings []spec.Finding, code string) []spec.Finding {
	var result []spec.Finding
	for _, f := range findings {
		if f.Code == code {
			result = append(result, f)
		}
	}
	return result
}

// TestLinter_AC01_HappyPath verifies that a fully valid SPEC produces no findings.
// AC-SPC-003-01: Given a valid SPEC with all REQs covered, When lint runs, Then exit 0 no findings.
func TestLinter_AC01_HappyPath(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("valid")})
	if err != nil {
		t.Fatalf("Lint returned unexpected error: %v", err)
	}

	errors := filterBySeverity(report.Findings, spec.SeverityError)
	if len(errors) != 0 {
		t.Errorf("expected 0 errors for valid SPEC, got %d: %v", len(errors), errors)
	}
}

// TestLinter_AC02_CoverageIncomplete verifies that CoverageIncomplete is reported
// when a REQ is not referenced from any AC.
// AC-SPC-003-02: Given SPEC with uncovered REQ-X-001-007, When lint, Then CoverageIncomplete.
func TestLinter_AC02_CoverageIncomplete(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("missing-coverage")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsCode(report.Findings, "CoverageIncomplete") {
		t.Error("expected CoverageIncomplete finding, got none")
	}

	// REQ-TST-002-007 must be named
	found := findingsForCode(report.Findings, "CoverageIncomplete")
	var hasUncovered bool
	for _, f := range found {
		if strings.Contains(f.Message, "REQ-TST-002-007") {
			hasUncovered = true
			break
		}
	}
	if !hasUncovered {
		t.Errorf("expected CoverageIncomplete to name REQ-TST-002-007, got: %v", found)
	}
}

// TestLinter_AC03_ModalityMalformed verifies that ModalityMalformed is reported
// for REQs with a malformed EARS modality.
// AC-SPC-003-03: WHEN...without SHALL → ModalityMalformed.
func TestLinter_AC03_ModalityMalformed(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("modality-malformed")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsCode(report.Findings, "ModalityMalformed") {
		t.Errorf("expected ModalityMalformed finding, findings were: %v", report.Findings)
	}
}

// TestLinter_AC04_DependencyCycle verifies that DependencyCycle is reported
// when an A->B->A cycle exists.
// AC-SPC-003-04: A depends B, B depends A → DependencyCycle.
func TestLinter_AC04_DependencyCycle(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	// Lint cycle-a/spec.md and cycle-b/spec.md together
	report, err := linter.Lint([]string{
		specPath("cycle-a"),
		specPath("cycle-b"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsCode(report.Findings, "DependencyCycle") {
		t.Errorf("expected DependencyCycle finding, got: %v", report.Findings)
	}
}

// TestLinter_AC05_DuplicateREQID verifies that DuplicateREQID is reported
// when a SPEC contains duplicate REQ IDs.
// AC-SPC-003-05: duplicate REQ-X-001-005 twice → DuplicateREQID.
func TestLinter_AC05_DuplicateREQID(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("dup-req-id")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsCode(report.Findings, "DuplicateREQID") {
		t.Errorf("expected DuplicateREQID finding, got: %v", report.Findings)
	}
}

// TestLinter_AC06_MissingExclusions verifies that MissingExclusions is reported
// when the Out of Scope section is missing.
// AC-SPC-003-06: missing Out of Scope → MissingExclusions.
func TestLinter_AC06_MissingExclusions(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("missing-exclusions")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsCode(report.Findings, "MissingExclusions") {
		t.Errorf("expected MissingExclusions finding, got: %v", report.Findings)
	}
}

// TestLinter_AC07_MissingDependency verifies that MissingDependency is reported
// when a SPEC depends on a non-existent SPEC.
// AC-SPC-003-07: dependencies: [SPEC-NONEXISTENT] → MissingDependency.
func TestLinter_AC07_MissingDependency(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("missing-dep")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsCode(report.Findings, "MissingDependency") {
		t.Errorf("expected MissingDependency finding, got: %v", report.Findings)
	}

	found := findingsForCode(report.Findings, "MissingDependency")
	var hasMissing bool
	for _, f := range found {
		if strings.Contains(f.Message, "SPEC-NONEXISTENT") {
			hasMissing = true
			break
		}
	}
	if !hasMissing {
		t.Errorf("expected MissingDependency to name SPEC-NONEXISTENT, got: %v", found)
	}
}

// TestLinter_AC08_DanglingRuleReference verifies that a DanglingRuleReference warning
// is reported when a non-existent CONST-V3R2-NNN reference exists.
// AC-SPC-003-08: related_rule: [CONST-V3R2-999] not in registry → DanglingRuleReference warning.
func TestLinter_AC08_DanglingRuleReference(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("dangling-rule")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := findingsForCode(report.Findings, "DanglingRuleReference")
	if len(found) == 0 {
		t.Errorf("expected DanglingRuleReference finding, got: %v", report.Findings)
		return
	}

	// severity must be warning
	for _, f := range found {
		if f.Severity != spec.SeverityWarning {
			t.Errorf("expected DanglingRuleReference to be warning severity, got %s", f.Severity)
		}
	}
}

// TestLinter_AC09_JSONOutput verifies that a JSON array is emitted when run with --json.
// AC-SPC-003-09: --json → valid JSON array of finding objects.
func TestLinter_AC09_JSONOutput(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("missing-coverage")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jsonBytes, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Must be valid JSON
	var findings []spec.Finding
	if err := json.Unmarshal(jsonBytes, &findings); err != nil {
		t.Errorf("invalid JSON output: %v\nraw: %s", err, jsonBytes)
	}

	// Each finding must have the required fields
	for i, f := range findings {
		if f.Code == "" {
			t.Errorf("finding[%d] missing Code field", i)
		}
		if f.Severity == "" {
			t.Errorf("finding[%d] missing Severity field", i)
		}
		if f.Message == "" {
			t.Errorf("finding[%d] missing Message field", i)
		}
	}
}

// TestLinter_AC10_SARIFOutput verifies that SARIF 2.1.0 format is emitted when run with --sarif.
// AC-SPC-003-10: --sarif → SARIF 2.1.0-conformant JSON.
func TestLinter_AC10_SARIFOutput(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("missing-coverage")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sarifBytes, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF failed: %v", err)
	}

	// Must be valid JSON
	var sarif map[string]interface{}
	if err := json.Unmarshal(sarifBytes, &sarif); err != nil {
		t.Errorf("invalid SARIF JSON: %v", err)
	}

	// Verify SARIF 2.1.0 required fields
	if v, ok := sarif["version"]; !ok || v != "2.1.0" {
		t.Errorf("expected SARIF version 2.1.0, got %v", v)
	}
	if _, ok := sarif["$schema"]; !ok {
		t.Error("SARIF missing $schema field")
	}
	if runs, ok := sarif["runs"]; !ok {
		t.Error("SARIF missing runs field")
	} else {
		runsSlice, ok := runs.([]interface{})
		if !ok || len(runsSlice) == 0 {
			t.Error("SARIF runs must be a non-empty array")
		}
	}
}

// TestLinter_AC11_StrictMode verifies that warnings are treated as errors when run with --strict.
// AC-SPC-003-11: --strict + warnings only → non-zero exit.
func TestLinter_AC11_StrictMode(t *testing.T) {
	// The dangling-rule SPEC only produces a DanglingRuleReference warning
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
		Strict:       true,
	})

	report, err := linter.Lint([]string{specPath("dangling-rule")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In strict mode, warnings must be promoted to errors
	if !report.HasErrors() {
		t.Error("expected HasErrors()=true in strict mode with warnings, got false")
	}
}

// TestLinter_AC12_DuplicateSPECID verifies that DuplicateSPECID is reported
// when two SPECs declare the same id.
// AC-SPC-003-12: two SPECs with same id → DuplicateSPECID.
func TestLinter_AC12_DuplicateSPECID(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{
		specPath("dup-spec-id-a"),
		specPath("dup-spec-id-b"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsCode(report.Findings, "DuplicateSPECID") {
		t.Errorf("expected DuplicateSPECID finding, got: %v", report.Findings)
	}
}

// TestLinter_AC13_LintSkip verifies that codes listed in lint.skip are suppressed.
// AC-SPC-003-13: lint.skip: [DanglingRuleReference] → no DanglingRuleReference finding.
func TestLinter_AC13_LintSkip(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	// lint-skip/spec.md has a CONST-V3R2-999 dangling reference but
	// it is suppressed by lint.skip: [DanglingRuleReference]
	report, err := linter.Lint([]string{specPath("lint-skip")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := findingsForCode(report.Findings, "DanglingRuleReference")
	if len(found) > 0 {
		t.Errorf("expected DanglingRuleReference to be suppressed by lint.skip, but got: %v", found)
	}
}

// TestLinter_AC14_BreakingChangeMissingID verifies that BreakingChangeMissingID is reported
// when breaking:true and bc_id:[].
// AC-SPC-003-14: breaking:true + bc_id:[] → BreakingChangeMissingID.
func TestLinter_AC14_BreakingChangeMissingID(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("breaking-no-bcid")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !containsCode(report.Findings, "BreakingChangeMissingID") {
		t.Errorf("expected BreakingChangeMissingID finding, got: %v", report.Findings)
	}
}

// TestLinter_AC15_ParseFailure verifies that ParseFailure is reported for a SPEC file
// that fails to parse, and that the linter continues processing other files.
// AC-SPC-003-15: malformed YAML → ParseFailure + continue with other files.
func TestLinter_AC15_ParseFailure(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	// Process malformed-yaml and valid together
	report, err := linter.Lint([]string{
		specPath("malformed-yaml"),
		specPath("valid"),
	})
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}

	if !containsCode(report.Findings, "ParseFailure") {
		t.Errorf("expected ParseFailure finding for malformed YAML, got: %v", report.Findings)
	}

	// Verify that the linter also processed the valid SPEC
	// The valid SPEC must have no findings (other than ParseFailure)
	nonParseFail := func() []spec.Finding {
		var result []spec.Finding
		for _, f := range report.Findings {
			if f.Code != "ParseFailure" && f.File != specPath("malformed-yaml") {
				result = append(result, f)
			}
		}
		return result
	}()

	errors := filterBySeverity(nonParseFail, spec.SeverityError)
	if len(errors) != 0 {
		t.Errorf("expected valid SPEC to have no errors, got: %v", errors)
	}
}

// TestLinter_AC16_HierarchicalACCoverage verifies that parent-level REQ references
// in hierarchical ACs are also counted as coverage.
// AC-SPC-003-16: hierarchical AC — leaf children cover parent REQ refs.
func TestLinter_AC16_HierarchicalACCoverage(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	report, err := linter.Lint([]string{specPath("hierarchical-ac")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All REQs in a hierarchical AC must be covered -> no CoverageIncomplete
	found := findingsForCode(report.Findings, "CoverageIncomplete")
	if len(found) > 0 {
		t.Errorf("expected no CoverageIncomplete for hierarchical AC, got: %v", found)
	}
}

// --- helper functions ---

// filterBySeverity returns only the findings of the given severity.
func filterBySeverity(findings []spec.Finding, sev spec.Severity) []spec.Finding {
	var result []spec.Finding
	for _, f := range findings {
		if f.Severity == sev {
			result = append(result, f)
		}
	}
	return result
}

// TestReport_HasErrors verifies that Report.HasErrors behaves correctly.
func TestReport_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		findings []spec.Finding
		strict   bool
		want     bool
	}{
		{
			name:     "no findings",
			findings: nil,
			want:     false,
		},
		{
			name: "only info",
			findings: []spec.Finding{
				{Severity: spec.SeverityInfo, Code: "SomeInfo"},
			},
			want: false,
		},
		{
			name: "has warning without strict",
			findings: []spec.Finding{
				{Severity: spec.SeverityWarning, Code: "SomeWarning"},
			},
			strict: false,
			want:   false,
		},
		{
			name: "has warning with strict",
			findings: []spec.Finding{
				{Severity: spec.SeverityWarning, Code: "SomeWarning"},
			},
			strict: true,
			want:   true,
		},
		{
			name: "has error",
			findings: []spec.Finding{
				{Severity: spec.SeverityError, Code: "SomeError"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &spec.Report{
				Findings: tt.findings,
				Strict:   tt.strict,
			}
			got := report.HasErrors()
			if got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLinter_NoArgs_DiscoversSPECs verifies that running without path arguments
// auto-discovers spec.md files under BaseDir.
func TestLinter_NoArgs_DiscoversSPECs(t *testing.T) {
	// Create a temp directory containing only a single valid SPEC
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "SPEC-TST-999")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Copy the valid spec.md into the temp directory
	content, err := os.ReadFile(specPath("valid"))
	if err != nil {
		t.Fatalf("failed to read valid spec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), content, 0644); err != nil {
		t.Fatal(err)
	}

	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      tmpDir,
	})

	// Call without arguments -> auto-discover from BaseDir
	report, err := linter.Lint(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errors := filterBySeverity(report.Findings, spec.SeverityError)
	if len(errors) != 0 {
		t.Errorf("expected no errors for valid SPEC, got: %v", errors)
	}
}

// TestStatusValueEnumRule_Valid verifies that a valid status value produces no findings.
func TestStatusValueEnumRule_Valid(t *testing.T) {
	doc := &spec.SPECDoc{
		Frontmatter: spec.SPECFrontmatter{
			Status: "planned",
		},
	}

	rule := &spec.StatusValueEnumRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 0 {
		t.Errorf("expected no findings for valid status, got: %v", findings)
	}
}

// TestStatusValueEnumRule_Invalid verifies that an invalid status value reports an error.
func TestStatusValueEnumRule_Invalid(t *testing.T) {
	doc := &spec.SPECDoc{
		Frontmatter: spec.SPECFrontmatter{
			Status: "Planned", // uppercase, not in enum
		},
	}

	rule := &spec.StatusValueEnumRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %v", len(findings), findings)
	}

	if findings[0].Code != "StatusValueInvalid" {
		t.Errorf("expected code StatusValueInvalid, got %s", findings[0].Code)
	}
}

// TestStatusValueEnumRule_Empty verifies that an empty status value produces no findings.
func TestStatusValueEnumRule_Empty(t *testing.T) {
	doc := &spec.SPECDoc{
		Frontmatter: spec.SPECFrontmatter{
			Status: "",
		},
	}

	rule := &spec.StatusValueEnumRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 0 {
		t.Errorf("expected no findings for empty status (handled by FrontmatterSchemaRule), got: %v", findings)
	}
}

// TestStatusCaseNormalizationRule_Lowercase verifies that lowercase status produces no findings.
func TestStatusCaseNormalizationRule_Lowercase(t *testing.T) {
	doc := &spec.SPECDoc{
		Frontmatter: spec.SPECFrontmatter{
			Status: "planned",
		},
	}

	rule := &spec.StatusCaseNormalizationRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 0 {
		t.Errorf("expected no findings for lowercase status, got: %v", findings)
	}
}

// TestStatusCaseNormalizationRule_Uppercase verifies that uppercase status reports an error.
func TestStatusCaseNormalizationRule_Uppercase(t *testing.T) {
	doc := &spec.SPECDoc{
		Frontmatter: spec.SPECFrontmatter{
			Status: "COMPLETED",
		},
	}

	rule := &spec.StatusCaseNormalizationRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %v", len(findings), findings)
	}

	if findings[0].Code != "StatusCaseInvalid" {
		t.Errorf("expected code StatusCaseInvalid, got %s", findings[0].Code)
	}

	expectedMsg := `status "COMPLETED" contains uppercase; use lowercase "completed" instead`
	if findings[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, findings[0].Message)
	}
}

// TestFrontmatterSchemaRule_Valid12Field verifies that the 12-field canonical fixture
// produces 0 FrontmatterInvalid findings.
// AC-SDBT-002-002 (valid case): canonical 12 fields → 0 FrontmatterInvalid findings.
func TestFrontmatterSchemaRule_Valid12Field(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	path := filepath.Join(testdataDir, "frontmatter-schema", "valid-12-field", "spec.md")
	report, err := linter.Lint([]string{path})
	if err != nil {
		t.Fatalf("Lint returned unexpected error: %v", err)
	}

	// Extract only the FrontmatterInvalid findings
	schemaFindings := findingsForCode(report.Findings, "FrontmatterInvalid")
	if len(schemaFindings) != 0 {
		t.Errorf("valid 12-field fixture에서 FrontmatterInvalid finding 0건 기대, 실제 %d건: %v",
			len(schemaFindings), schemaFindings)
	}
}

// TestFrontmatterSchemaRule_SnakeCaseRejected verifies that a snake_case-only alias fixture
// produces exactly 3 FrontmatterInvalid findings (missing created, updated, tags).
// AC-SDBT-002-002: snake_case aliases only (created_at/updated_at/labels) → exactly 3 FrontmatterInvalid findings.
func TestFrontmatterSchemaRule_SnakeCaseRejected(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	path := filepath.Join(testdataDir, "frontmatter-schema", "invalid-snake-case-only", "spec.md")
	report, err := linter.Lint([]string{path})
	if err != nil {
		t.Fatalf("Lint returned unexpected error: %v", err)
	}

	// Extract only the FrontmatterInvalid findings
	schemaFindings := findingsForCode(report.Findings, "FrontmatterInvalid")

	// Exactly 3 (missing created, updated, tags)
	if len(schemaFindings) != 3 {
		t.Fatalf("snake_case-only fixture에서 FrontmatterInvalid finding 정확히 3건 기대, 실제 %d건: %v",
			len(schemaFindings), schemaFindings)
	}

	// Verify that each finding mentions the expected field
	expectedFields := map[string]bool{
		"created": false,
		"updated": false,
		"tags":    false,
	}
	for _, f := range schemaFindings {
		for field := range expectedFields {
			if strings.Contains(f.Message, "Frontmatter required field missing: "+field) {
				expectedFields[field] = true
			}
		}
	}
	for field, found := range expectedFields {
		if !found {
			t.Errorf("FrontmatterInvalid finding에서 field %q 언급 없음. findings: %v", field, schemaFindings)
		}
	}
}

// TestStatusCaseNormalizationRule_MixedCase verifies that a mixed-case status reports an error.
func TestStatusCaseNormalizationRule_MixedCase(t *testing.T) {
	doc := &spec.SPECDoc{
		Frontmatter: spec.SPECFrontmatter{
			Status: "In-Progress",
		},
	}

	rule := &spec.StatusCaseNormalizationRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %v", len(findings), findings)
	}

	if findings[0].Code != "StatusCaseInvalid" {
		t.Errorf("expected code StatusCaseInvalid, got %s", findings[0].Code)
	}

	expectedMsg := `status "In-Progress" contains uppercase; use lowercase "in-progress" instead`
	if findings[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, findings[0].Message)
	}
}

// =============================================================================
// SPEC-V3R6-GEARS-MIGRATION-001 — M2 LegacyEARSKeyword tests (4 cases)
// =============================================================================
//
// This SPEC marks IF/THEN patterns as deprecated and emits a LegacyEARSKeyword warning.
// WHEN/WHILE/WHERE/Ubiquitous remain GEARS-compatible.
//
// SSOT: .moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/spec.md REQ-GM-002, REQ-GM-006, REQ-GM-008, REQ-GM-009

// TestEARSModalityRule_LegacyEARSKeyword_IFThen verifies that an IF/THEN REQ
// emits exactly one LegacyEARSKeyword warning.
// AC-GM-002 binary criteria.
func TestEARSModalityRule_LegacyEARSKeyword_IFThen(t *testing.T) {
	doc := &spec.SPECDoc{
		Path: "test.md",
		REQs: []spec.REQEntry{
			{ID: "REQ-LEG-001-005", Text: "IF a deprecated keyword is detected THEN the system SHALL emit a migration warning.", Line: 42},
		},
	}

	rule := &spec.EARSModalityRule{}
	findings := rule.Check(doc, nil)

	legacy := findingsForCode(findings, "LegacyEARSKeyword")
	if len(legacy) != 1 {
		t.Fatalf("expected exactly 1 LegacyEARSKeyword finding, got %d: %v", len(legacy), legacy)
	}

	if legacy[0].Severity != spec.SeverityWarning {
		t.Errorf("expected severity warning, got %s", legacy[0].Severity)
	}

	if legacy[0].Line != 42 {
		t.Errorf("expected line 42, got %d", legacy[0].Line)
	}

	// AC-GM-002: ModalityMalformed MUST NOT additionally fire when SHALL is present.
	malformed := findingsForCode(findings, "ModalityMalformed")
	if len(malformed) != 0 {
		t.Errorf("expected 0 ModalityMalformed (SHALL is present), got %d: %v", len(malformed), malformed)
	}
}

// TestEARSModalityRule_GEARSWellFormed verifies that canonical GEARS REQs
// (WHEN/WHILE/WHERE/Ubiquitous) produce 0 findings.
// AC-GM-003 binary criteria.
func TestEARSModalityRule_GEARSWellFormed(t *testing.T) {
	doc := &spec.SPECDoc{
		Path: "test.md",
		REQs: []spec.REQEntry{
			{ID: "REQ-GRS-001-001", Text: "The system SHALL always preserve EARS modality compliance.", Line: 10},
			{ID: "REQ-GRS-001-002", Text: "WHEN a new SPEC is added, the system SHALL detect it during discovery.", Line: 20},
			{ID: "REQ-GRS-001-003", Text: "WHILE the linter holds the rule registry, the system SHALL apply rules in declaration order.", Line: 30},
			{ID: "REQ-GRS-001-004", Text: "WHERE strict mode is enabled, the system SHALL escalate warnings to errors via Report.HasErrors().", Line: 40},
		},
	}

	rule := &spec.EARSModalityRule{}
	findings := rule.Check(doc, nil)

	if len(findings) != 0 {
		t.Errorf("expected 0 findings on GEARS-well-formed REQs, got %d: %v", len(findings), findings)
	}
}

// TestEARSModalityRule_LegacyEARSKeyword_StrictExitCode verifies that in --strict mode
// a LegacyEARSKeyword warning escalates to exit-1 via Report.HasErrors().
// AC-GM-008 binary criteria.
func TestEARSModalityRule_LegacyEARSKeyword_StrictExitCode(t *testing.T) {
	// Synthesize a report with one LegacyEARSKeyword warning + strict=true.
	report := &spec.Report{
		Findings: []spec.Finding{
			{
				File:     "test.md",
				Line:     42,
				Severity: spec.SeverityWarning,
				Code:     "LegacyEARSKeyword",
				Message:  "REQ REQ-LEG-001-005: GEARS migration: ...",
			},
		},
		Strict: true,
	}

	if !report.HasErrors() {
		t.Error("expected HasErrors()=true in strict mode with LegacyEARSKeyword warning, got false")
	}

	// Severity field is unchanged — only HasErrors() escalates.
	if report.Findings[0].Severity != spec.SeverityWarning {
		t.Errorf("expected severity field unchanged (warning), got %s", report.Findings[0].Severity)
	}

	// Non-strict mode: same finding must NOT trigger HasErrors().
	reportNonStrict := &spec.Report{
		Findings: report.Findings,
		Strict:   false,
	}
	if reportNonStrict.HasErrors() {
		t.Error("expected HasErrors()=false in non-strict mode with only warnings, got true")
	}
}

// TestEARSModalityRule_MessageContainsDocsURL verifies that the LegacyEARSKeyword
// finding's Message contains the substrings "GEARS migration" and "adk.mo.ai.kr".
// AC-GM-002 binary criteria (docs URL linkage) + REQ-GM-006.
func TestEARSModalityRule_MessageContainsDocsURL(t *testing.T) {
	doc := &spec.SPECDoc{
		Path: "test.md",
		REQs: []spec.REQEntry{
			{ID: "REQ-LEG-001-005", Text: "IF a deprecated keyword is detected THEN the system SHALL emit a migration warning.", Line: 42},
		},
	}

	rule := &spec.EARSModalityRule{}
	findings := rule.Check(doc, nil)

	legacy := findingsForCode(findings, "LegacyEARSKeyword")
	if len(legacy) != 1 {
		t.Fatalf("expected 1 LegacyEARSKeyword finding, got %d", len(legacy))
	}

	msg := legacy[0].Message

	if !strings.Contains(msg, "GEARS migration") {
		t.Errorf("expected Message to contain %q, got %q", "GEARS migration", msg)
	}

	if !strings.Contains(msg, "adk.mo.ai.kr") {
		t.Errorf("expected Message to contain docs URL substring %q, got %q", "adk.mo.ai.kr", msg)
	}
}
