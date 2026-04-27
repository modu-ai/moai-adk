package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidator_ValidateTemplate_SHA요구사항 검증
func TestValidator_ValidateTemplate_SHA요구사항(t *testing.T) {
	// RED phase: 실패하는 테스트 작성
	v := NewValidator()

	// SHA가 없는 템플릿 (실패 예상)
	noSHA := `
name: Test Workflow
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "test"
`
	tmpDir := t.TempDir()
	noSHATemplatePath := filepath.Join(tmpDir, "no-sha.yml.tmpl")
	if err := os.WriteFile(noSHATemplatePath, []byte(noSHA), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	result, err := v.ValidateTemplate(noSHATemplatePath)
	if err != nil {
		t.Fatalf("ValidateTemplate failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected validation to fail for missing SHA pin, but passed")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for missing SHA pin, got none")
	}

	// SHA가 있는 템플릿 (성공 예상)
	withSHA := `
name: Test Workflow
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4  # SHA: actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f9536
      - run: echo "test"
`
	withSHATemplatePath := filepath.Join(tmpDir, "with-sha.yml.tmpl")
	if err := os.WriteFile(withSHATemplatePath, []byte(withSHA), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	result, err = v.ValidateTemplate(withSHATemplatePath)
	if err != nil {
		t.Fatalf("ValidateTemplate failed: %v", err)
	}

	if !result.IsValid {
		t.Errorf("Expected validation to pass with SHA pin, but failed: %v", result.Errors)
	}
}

// TestValidator_ValidateTemplate_CodexPrivateRepoGuard REQ-SEC-001 검증
func TestValidator_ValidateTemplate_CodexPrivateRepoGuard(t *testing.T) {
	v := NewValidator()

	// Codex private guard가 없는 템플릿 (실패 예상)
	noGuard := `
name: Codex Review
on: pull_request
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: codex review
`
	tmpDir := t.TempDir()
	noGuardPath := filepath.Join(tmpDir, "codex-no-guard.yml.tmpl")
	if err := os.WriteFile(noGuardPath, []byte(noGuard), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	result, err := v.ValidateTemplate(noGuardPath)
	if err != nil {
		t.Fatalf("ValidateTemplate failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected validation to fail for missing Codex private repo guard, but passed")
	}

	// Guard 검사: 에러 메시지 확인
	foundGuardError := false
	for _, e := range result.Errors {
		if strings.Contains(e, "REQ-SEC-001") || strings.Contains(e, "private repo") || strings.Contains(e, "repository_visibility") {
			foundGuardError = true
			break
		}
	}
	if !foundGuardError {
		t.Error("Expected error about missing private repo guard, got:", result.Errors)
	}
}

// TestValidator_ValidateTemplate_NoHardcodedCredentials SEC-003 검증
func TestValidator_ValidateTemplate_NoHardcodedCredentials(t *testing.T) {
	v := NewValidator()

	// 하드코딩된 credential이 있는 템플릿 (실패 예상)
	hardcodedCred := `
name: Test Workflow
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "API_KEY=sk-1234567890abcdef"
`
	tmpDir := t.TempDir()
	hardcodedPath := filepath.Join(tmpDir, "hardcoded.yml.tmpl")
	if err := os.WriteFile(hardcodedPath, []byte(hardcodedCred), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	result, err := v.ValidateTemplate(hardcodedPath)
	if err != nil {
		t.Fatalf("ValidateTemplate failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected validation to fail for hardcoded credentials, but passed")
	}

	// Credential 검사: 에러 메시지 확인
	foundCredError := false
	for _, e := range result.Errors {
		if strings.Contains(e, "SEC-003") || strings.Contains(e, "hardcoded") || strings.Contains(e, "credential") {
			foundCredError = true
			break
		}
	}
	if !foundCredError {
		t.Error("Expected error about hardcoded credentials, got:", result.Errors)
	}
}

// TestValidator_ValidateTemplate_ProperPermissions SEC-005 검증
func TestValidator_ValidateTemplate_ProperPermissions(t *testing.T) {
	v := NewValidator()

	// 과도한 권한이 있는 템플릿 (실패 예상)
	excessivePerms := `
name: Test Workflow
on: push
permissions: write-all  # 너무 과도한 권한
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "test"
`
	tmpDir := t.TempDir()
	excessivePath := filepath.Join(tmpDir, "excessive-perms.yml.tmpl")
	if err := os.WriteFile(excessivePath, []byte(excessivePerms), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	result, err := v.ValidateTemplate(excessivePath)
	if err != nil {
		t.Fatalf("ValidateTemplate failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected validation to fail for excessive permissions, but passed")
	}

	// 권한 검사: 에러 또는 경고 확인
	foundPermWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "SEC-005") || strings.Contains(w, "permission") {
			foundPermWarning = true
			break
		}
	}
	if !foundPermWarning {
		t.Error("Expected warning about excessive permissions, got:", result.Warnings)
	}
}

// TestValidator_ValidateAllTemplates 템플릿 디렉토리 전체 검증
func TestValidator_ValidateAllTemplates(t *testing.T) {
	v := NewValidator()

	// 테스트용 템플릿 디렉토리 생성
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// 유효한 템플릿 생성
	validTemplate := `
name: Valid Workflow
on: pull_request
permissions:
  contents: read
  pull-requests: write
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4  # SHA: actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f9536
      - run: echo "test"
`
	validPath := filepath.Join(templatesDir, "valid.yml.tmpl")
	if err := os.WriteFile(validPath, []byte(validTemplate), 0644); err != nil {
		t.Fatalf("Failed to write valid template: %v", err)
	}

	results, err := v.ValidateAllTemplates(templatesDir)
	if err != nil {
		t.Fatalf("ValidateAllTemplates failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected validation results, got none")
	}

	// 최소한 하나의 결과는 유효해야 함
	foundValid := false
	for _, r := range results {
		if r.IsValid {
			foundValid = true
			break
		}
	}
	if !foundValid {
		t.Error("Expected at least one valid template, all failed")
	}
}

// TestValidator_ValidateTemplate_YAMLSyntax YAML 문법 검증
func TestValidator_ValidateTemplate_YAMLSyntax(t *testing.T) {
	v := NewValidator()

	// 잘못된 YAML 문법 (tab 문자로 인한 indentation error)
	invalidYAML := "name: Test Workflow\non: push\njobs:\n\ttest:\n  runs-on: ubuntu-latest" // mixed tabs/spaces
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid-yaml.yml.tmpl")
	if err := os.WriteFile(invalidPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write invalid YAML: %v", err)
	}

	result, err := v.ValidateTemplate(invalidPath)
	if err != nil {
		t.Fatalf("ValidateTemplate failed: %v", err)
	}

	// YAML 파서는 tab을 허용하므로 실제로는 에러가 아닐 수 있음
	// 대신 진짜 invalid YAML: unmatched brackets
	reallyInvalidYAML := `
name: Test
on: [push
jobs:  # Unmatched bracket
`
	reallyInvalidPath := filepath.Join(tmpDir, "really-invalid-yaml.yml.tmpl")
	if err := os.WriteFile(reallyInvalidPath, []byte(reallyInvalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write really invalid YAML: %v", err)
	}

	result, err = v.ValidateTemplate(reallyInvalidPath)
	if err != nil {
		t.Fatalf("ValidateTemplate failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected validation to fail for invalid YAML syntax, but passed")
	}

	// YAML 에러 확인
	foundYAMLError := false
	for _, e := range result.Errors {
		if strings.Contains(e, "YAML") || strings.Contains(e, "syntax") {
			foundYAMLError = true
			break
		}
	}
	if !foundYAMLError {
		t.Error("Expected error about YAML syntax, got:", result.Errors)
	}
}

// TestNewValidator Validator 생성자 검증
func TestNewValidator(t *testing.T) {
	v := NewValidator()

	if v == nil {
		t.Fatal("NewValidator returned nil")
	}

	// embed.FS는 항상 zero value로 초기화 가능하므로 nil 체크 불필요
	// 실제 사용 시 os.ReadFile로 직접 파일을 읽으므로 문제 없음
}

// TestValidationResult_ResultStructure 검증 결과 구조 검증
func TestValidationResult_ResultStructure(t *testing.T) {
	result := &ValidationResult{
		TemplatePath: "/test/path.yml.tmpl",
		IsValid:     false,
		Errors:      []string{"error 1", "error 2"},
		Warnings:    []string{"warning 1"},
	}

	if result.TemplatePath != "/test/path.yml.tmpl" {
		t.Error("TemplatePath not set correctly")
	}

	if result.IsValid {
		t.Error("Expected IsValid to be false")
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}

// TestValidator_ValidateTemplate_EmptyTemplate 빈 템플릿 검증
func TestValidator_ValidateTemplate_EmptyTemplate(t *testing.T) {
	v := NewValidator()

	tmpDir := t.TempDir()
	emptyPath := filepath.Join(tmpDir, "empty.yml.tmpl")
	if err := os.WriteFile(emptyPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty template: %v", err)
	}

	result, err := v.ValidateTemplate(emptyPath)
	if err != nil {
		t.Fatalf("ValidateTemplate failed: %v", err)
	}

	if result.IsValid {
		t.Error("Expected validation to fail for empty template, but passed")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for empty template, got none")
	}
}
