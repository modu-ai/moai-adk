package spec_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// testdataDir는 픽스처 SPEC 파일들이 위치한 디렉토리이다.
const testdataDir = "testdata"

// registryPath는 테스트에서 사용하는 zone registry 경로이다.
// 테스트에서는 실제 zone registry를 찾을 수 없는 경우 nil registry를 사용한다.
func testRegistryPath() string {
	// worktree 루트에서 zone registry 찾기
	dir := "../../.claude/rules/moai/core/zone-registry.md"
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	return ""
}

// specPath는 testdata 하위의 특정 픽스처 경로를 반환한다.
func specPath(fixture string) string {
	return filepath.Join(testdataDir, fixture, "spec.md")
}

// containsCode는 findings 슬라이스에 주어진 코드가 있는지 확인한다.
func containsCode(findings []spec.Finding, code string) bool {
	for _, f := range findings {
		if f.Code == code {
			return true
		}
	}
	return false
}

// findingsForCode는 주어진 코드의 findings를 반환한다.
func findingsForCode(findings []spec.Finding, code string) []spec.Finding {
	var result []spec.Finding
	for _, f := range findings {
		if f.Code == code {
			result = append(result, f)
		}
	}
	return result
}

// TestLinter_AC01_HappyPath는 완전히 유효한 SPEC에서 findings가 없음을 검증한다.
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

// TestLinter_AC02_CoverageIncomplete는 AC에서 참조되지 않는 REQ가 있을 때
// CoverageIncomplete 오류가 보고됨을 검증한다.
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

	// REQ-TST-002-007이 명시되어야 함
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

// TestLinter_AC03_ModalityMalformed는 EARS 모달리티가 잘못된 REQ에 대해
// ModalityMalformed 오류가 보고됨을 검증한다.
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

// TestLinter_AC04_DependencyCycle는 A→B→A 사이클이 있을 때
// DependencyCycle 오류가 보고됨을 검증한다.
// AC-SPC-003-04: A depends B, B depends A → DependencyCycle.
func TestLinter_AC04_DependencyCycle(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	// cycle-a/spec.md와 cycle-b/spec.md를 함께 lint
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

// TestLinter_AC05_DuplicateREQID는 동일 SPEC 내 중복 REQ ID가 있을 때
// DuplicateREQID 오류가 보고됨을 검증한다.
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

// TestLinter_AC06_MissingExclusions는 Out of Scope 섹션이 없을 때
// MissingExclusions 오류가 보고됨을 검증한다.
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

// TestLinter_AC07_MissingDependency는 존재하지 않는 SPEC 의존성이 있을 때
// MissingDependency 오류가 보고됨을 검증한다.
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

// TestLinter_AC08_DanglingRuleReference는 존재하지 않는 CONST-V3R2-NNN 참조가 있을 때
// DanglingRuleReference 경고가 보고됨을 검증한다.
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

	// severity는 warning이어야 함
	for _, f := range found {
		if f.Severity != spec.SeverityWarning {
			t.Errorf("expected DanglingRuleReference to be warning severity, got %s", f.Severity)
		}
	}
}

// TestLinter_AC09_JSONOutput는 --json 플래그로 실행 시 JSON 배열이 출력됨을 검증한다.
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

	// 유효한 JSON이어야 함
	var findings []spec.Finding
	if err := json.Unmarshal(jsonBytes, &findings); err != nil {
		t.Errorf("invalid JSON output: %v\nraw: %s", err, jsonBytes)
	}

	// 각 finding에 필수 필드가 있어야 함
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

// TestLinter_AC10_SARIFOutput는 --sarif 플래그로 실행 시 SARIF 2.1.0 형식이 출력됨을 검증한다.
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

	// 유효한 JSON이어야 함
	var sarif map[string]interface{}
	if err := json.Unmarshal(sarifBytes, &sarif); err != nil {
		t.Errorf("invalid SARIF JSON: %v", err)
	}

	// SARIF 2.1.0 필수 필드 확인
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

// TestLinter_AC11_StrictMode는 --strict 플래그로 실행 시 경고도 오류로 처리됨을 검증한다.
// AC-SPC-003-11: --strict + warnings only → non-zero exit.
func TestLinter_AC11_StrictMode(t *testing.T) {
	// dangling-rule SPEC은 DanglingRuleReference warning만 발생시킴
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
		Strict:       true,
	})

	report, err := linter.Lint([]string{specPath("dangling-rule")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// strict 모드에서는 경고가 오류로 승격되어야 함
	if !report.HasErrors() {
		t.Error("expected HasErrors()=true in strict mode with warnings, got false")
	}
}

// TestLinter_AC12_DuplicateSPECID는 두 SPEC이 같은 id를 선언할 때
// DuplicateSPECID 오류가 보고됨을 검증한다.
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

// TestLinter_AC13_LintSkip는 lint.skip으로 지정된 코드가 억제됨을 검증한다.
// AC-SPC-003-13: lint.skip: [DanglingRuleReference] → no DanglingRuleReference finding.
func TestLinter_AC13_LintSkip(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	// lint-skip/spec.md는 CONST-V3R2-999 dangling 참조가 있지만
	// lint.skip: [DanglingRuleReference]로 억제됨
	report, err := linter.Lint([]string{specPath("lint-skip")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := findingsForCode(report.Findings, "DanglingRuleReference")
	if len(found) > 0 {
		t.Errorf("expected DanglingRuleReference to be suppressed by lint.skip, but got: %v", found)
	}
}

// TestLinter_AC14_BreakingChangeMissingID는 breaking:true + bc_id:[] 일 때
// BreakingChangeMissingID 오류가 보고됨을 검증한다.
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

// TestLinter_AC15_ParseFailure는 파싱 실패한 SPEC 파일에서 ParseFailure 오류가 보고되고
// linter가 다른 파일은 계속 처리함을 검증한다.
// AC-SPC-003-15: malformed YAML → ParseFailure + continue with other files.
func TestLinter_AC15_ParseFailure(t *testing.T) {
	linter := spec.NewLinter(spec.LinterOptions{
		RegistryPath: testRegistryPath(),
		BaseDir:      testdataDir,
	})

	// malformed-yaml과 valid를 함께 처리
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

	// linter가 valid SPEC도 처리했는지 확인
	// valid SPEC에 대한 findings는 없어야 함 (ParseFailure 외)
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

// TestLinter_AC16_HierarchicalACCoverage는 계층 AC에서 부모 레벨의
// REQ 참조도 커버리지로 인정됨을 검증한다.
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

	// 계층 AC에서 모든 REQ가 커버되어야 함 → CoverageIncomplete 없어야 함
	found := findingsForCode(report.Findings, "CoverageIncomplete")
	if len(found) > 0 {
		t.Errorf("expected no CoverageIncomplete for hierarchical AC, got: %v", found)
	}
}

// --- helper functions ---

// filterBySeverity는 주어진 severity의 findings만 반환한다.
func filterBySeverity(findings []spec.Finding, sev spec.Severity) []spec.Finding {
	var result []spec.Finding
	for _, f := range findings {
		if f.Severity == sev {
			result = append(result, f)
		}
	}
	return result
}

// TestReport_HasErrors는 Report.HasErrors가 올바르게 동작함을 검증한다.
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

// TestLinter_NoArgs_DiscoversSPECs는 경로 인수 없이 실행 시
// BaseDir 아래 spec.md 파일들을 자동 탐색함을 검증한다.
func TestLinter_NoArgs_DiscoversSPECs(t *testing.T) {
	// 단일 valid SPEC만 있는 임시 디렉토리 생성
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "SPEC-TST-999")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}

	// valid spec.md를 임시 디렉토리에 복사
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

	// 인수 없이 호출 → BaseDir에서 자동 탐색
	report, err := linter.Lint(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errors := filterBySeverity(report.Findings, spec.SeverityError)
	if len(errors) != 0 {
		t.Errorf("expected no errors for valid SPEC, got: %v", errors)
	}
}
