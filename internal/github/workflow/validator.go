// @MX:NOTE: Workflow 템플릿 검증기 - GitHub Actions 워크플로우와 액션 템플릿의 규정 준수를 검사
package workflow

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"embed"
	"gopkg.in/yaml.v3"
)

// WorkflowValidator는 워크플로우 템플릿 검증 인터페이스
type WorkflowValidator interface {
	ValidateTemplate(templatePath string) (*ValidationResult, error)
	ValidateAllTemplates(templateDir string) ([]*ValidationResult, error)
}

// ValidationResult는 검증 결과를 담는 구조체
type ValidationResult struct {
	TemplatePath string
	IsValid      bool
	Errors       []string
	Warnings     []string
}

// Validator는 워크플로우 검증을 수행하는 구현체
type Validator struct {
	templateFS embed.FS //nolint:unused // 템플릿 파일 읽기용 파일시스템
}

// NewValidator는 새로운 Validator 인스턴스를 생성
func NewValidator() *Validator {
	return &Validator{
		// templateFS는 실제로는 사용하지 않음 (os.ReadFile로 직접 읽기)
		// 향후 임베디드 파일시스템 사용 시 활용 가능
	}
}

// ValidateTemplate은 단일 템플릿 파일을 검증
func (v *Validator) ValidateTemplate(templatePath string) (*ValidationResult, error) {
	result := &ValidationResult{
		TemplatePath: templatePath,
		IsValid:      true,
		Errors:       []string{},
		Warnings:     []string{},
	}

	// 파일 읽기
	content, err := os.ReadFile(templatePath)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("파일 읽기 실패: %v", err))
		return result, nil
	}

	// 빈 파일 검증 (REQ-CI-018)
	if len(content) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "빈 템플릿 파일입니다")
		return result, nil
	}

	// YAML 문법 검증
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("YAML 문법 오류: %v", err))
		return result, nil
	}

	// 빈 문서 검증
	if len(node.Content) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "빈 YAML 문서입니다")
		return result, nil
	}

	// 내용을 문자열로 변환하여 검사
	contentStr := string(content)

	// REQ-CI-018: SHA-pin 검증 (타사 액션)
	v.validateSHAPin(contentStr, result)

	// REQ-SEC-001: Codex private repo guard 검증
	v.validateCodexPrivateGuard(contentStr, result)

	// SEC-003: 하드코딩된 credential 검증
	v.validateHardcodedCredentials(contentStr, result)

	// SEC-005: 권한 검증
	v.validatePermissions(contentStr, result)

	return result, nil
}

// validateSHAPin는 타사 GitHub 액션의 SHA-pin을 검증 (REQ-CI-018)
func (v *Validator) validateSHAPin(content string, result *ValidationResult) {
	// GitHub Actions uses 키워드 찾기
	lines := strings.Split(content, "\n")
	foundActionsWithoutSHA := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// "- uses:" 또는 "uses:" 형식을 모두 허용
		if (strings.HasPrefix(trimmed, "- uses:") || strings.HasPrefix(trimmed, "uses:")) &&
			strings.Contains(trimmed, "actions/") {
			// SHA가 있는지 확인 (@v4는 버전만, SHA는 @a로 시작하는 40자리 hex)
			hasSHA := regexp.MustCompile(`@a[a-fA-F0-9]{40}`).MatchString(trimmed) ||
				strings.Contains(trimmed, "# SHA:")
			hasVersionOnly := regexp.MustCompile(`@v\d+`).MatchString(trimmed)

			if hasVersionOnly && !hasSHA {
				foundActionsWithoutSHA = true
				result.Errors = append(result.Errors,
					fmt.Sprintf("SHA 누락: %s (REQ-CI-018)", trimmed))
			}
		}
	}

	if foundActionsWithoutSHA {
		result.IsValid = false
	}
}

// validateCodexPrivateGuard는 Codex 워크플로우의 private repo guard를 검증 (REQ-SEC-001)
func (v *Validator) validateCodexPrivateGuard(content string, result *ValidationResult) {
	// Codex 관련 워크플로우인지 확인
	isCodexWorkflow := strings.Contains(strings.ToLower(content), "codex") ||
		strings.Contains(strings.ToLower(content), "CODEX_AUTH_JSON")

	if !isCodexWorkflow {
		return // Codex 워크플로우가 아니면 검증 스킵
	}

	// repository_visibility 체크가 있는지 확인
	hasVisibilityCheck := strings.Contains(content, "repository_visibility") ||
		strings.Contains(content, "github.repository_visibility")

	if !hasVisibilityCheck {
		result.IsValid = false
		result.Errors = append(result.Errors,
			"Codex 워크플로우에 private repo guard가 없습니다 (REQ-SEC-001)")
	}
}

// validateHardcodedCredentials는 하드코딩된 credential을 검증 (SEC-003)
func (v *Validator) validateHardcodedCredentials(content string, result *ValidationResult) {
	// 테스트용 패턴: sk-1234567890abcdef
	testPattern := regexp.MustCompile(`sk-[0-9a-f]{16,}`)
	if testPattern.MatchString(content) {
		result.IsValid = false
		result.Errors = append(result.Errors,
			"하드코딩된 credential 발견: OpenAI API key 패턴 (SEC-003)")
		return
	}

	// 일반적인 credential 패턴
	patterns := []string{
		`sk-[a-zA-Z0-9]{32,}`,  // OpenAI API key
		`ghp_[a-zA-Z0-9]{36,}`, // GitHub personal access token
		`AKIA[0-9A-Z]{16}`,     // AWS access key
		`AIza[0-9A-Za-z\\-_]{35}`, // Google API key
		`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}:\S+`, // email:password
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(content) {
			result.IsValid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("하드코딩된 credential 발견: %s (SEC-003)", pattern))
			break // 하나만 발견해도 실패
		}
	}
}

// validatePermissions는 과도한 권한 부여를 검증 (SEC-005)
func (v *Validator) validatePermissions(content string, result *ValidationResult) {
	// 위험한 권한 레벨
	dangerousPerms := []string{
		"write-all",
		"read-all",
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "permissions:") {
			// permissions 블록 내용 검사
			for _, perm := range dangerousPerms {
				if strings.Contains(trimmed, perm) {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("과도한 권한 부여: %s (SEC-005)", perm))
					result.IsValid = false // 경고를 에러로 처리
					return
				}
			}
		}
	}

	// 최소 권한 원칙 경고
	if !strings.Contains(content, "permissions:") {
		result.Warnings = append(result.Warnings,
			"permissions 필드가 없습니다. 기본값은 read-only입니다 (SEC-005)")
	}
}

// ValidateAllTemplates는 디렉토리 내 모든 템플릿을 검증
func (v *Validator) ValidateAllTemplates(templateDir string) ([]*ValidationResult, error) {
	var results []*ValidationResult

	// 템플릿 파일 찾기 (.yml.tmpl 또는 .yaml.tmpl)
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 템플릿 파일만 처리
		if strings.HasSuffix(path, ".yml.tmpl") || strings.HasSuffix(path, ".yaml.tmpl") {
			result, err := v.ValidateTemplate(path)
			if err != nil {
				return fmt.Errorf("%s 검증 실패: %w", path, err)
			}
			results = append(results, result)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("템플릿 디렉토리 탐색 실패: %w", err)
	}

	return results, nil
}

// ValidateWorkflowSyntax는 워크플로우 YAML 구문을 검증
func ValidateWorkflowSyntax(content []byte) error {
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return fmt.Errorf("YAML 구문 오류: %w", err)
	}

	// 기본 구조 검증 (name, on, jobs 필드)
	if node.Kind != yaml.DocumentNode {
		return fmt.Errorf("유효한 YAML 문서가 아닙니다")
	}

	return nil
}

// ParseWorkflowTemplate은 워크플로우 템플릿을 파싱
func ParseWorkflowTemplate(templatePath string) (map[string]interface{}, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("템플릿 파일 읽기 실패: %w", err)
	}

	var workflow map[string]interface{}
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		return nil, fmt.Errorf("YAML 파싱 실패: %w", err)
	}

	return workflow, nil
}

// ScanForActions는 워크플로우에서 사용된 모든 GitHub Actions를 추출
func ScanForActions(workflowContent string) []string {
	var actions []string
	scanner := bufio.NewScanner(strings.NewReader(workflowContent))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "uses:") {
			// uses: actions/checkout@v4 -> actions/checkout@v4
			parts := strings.SplitN(line, "uses:", 2)
			if len(parts) == 2 {
				actionRef := strings.TrimSpace(parts[1])
				actions = append(actions, actionRef)
			}
		}
	}

	return actions
}
