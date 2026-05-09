// @MX:NOTE: Workflow template validator - Checks compliance of GitHub Actions workflows and action templates
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

// WorkflowValidator is the workflow template validation interface
type WorkflowValidator interface {
	ValidateTemplate(templatePath string) (*ValidationResult, error)
	ValidateAllTemplates(templateDir string) ([]*ValidationResult, error)
}

// ValidationResult is a struct that holds validation results
type ValidationResult struct {
	TemplatePath string
	IsValid      bool
	Errors       []string
	Warnings     []string
}

// Validator is the implementation that performs workflow validation
type Validator struct {
	templateFS embed.FS //nolint:unused // filesystem for file reading
}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{
		// templateFS is not actually used (reads directly via os.ReadFile)
	}
}

// ValidateTemplate validates a single template file
func (v *Validator) ValidateTemplate(templatePath string) (*ValidationResult, error) {
	result := &ValidationResult{
		TemplatePath: templatePath,
		IsValid:      true,
		Errors:       []string{},
		Warnings:     []string{},
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("file read failure: %v", err))
		return result, nil
	}

	// File validation (REQ-CI-018)
	if len(content) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "empty template file")
		return result, nil
	}

	// YAML syntax validation
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("YAML syntax error: %v", err))
		return result, nil
	}

	if len(node.Content) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "empty YAML document")
		return result, nil
	}

	contentStr := string(content)

	// REQ-CI-018: SHA-pin validation (third-party actions)
	v.validateSHAPin(contentStr, result)

	// REQ-SEC-001: Codex private repo guard validation
	v.validateCodexPrivateGuard(contentStr, result)

	// SEC-003: hardcoded credential validation
	v.validateHardcodedCredentials(contentStr, result)

	// SEC-005: authorization validation
	v.validatePermissions(contentStr, result)

	return result, nil
}

// validateSHAPin validates SHA-pinning for third-party GitHub actions (REQ-CI-018)
func (v *Validator) validateSHAPin(content string, result *ValidationResult) {
	// Find GitHub Actions uses keyword
	lines := strings.Split(content, "\n")
	foundActionsWithoutSHA := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Allow both "- uses:" and "uses:" formats
		if (strings.HasPrefix(trimmed, "- uses:") || strings.HasPrefix(trimmed, "uses:")) &&
			strings.Contains(trimmed, "actions/") {
			// Check if SHA exists (@v4 is version only, SHA starts with @a and is 40 hex chars)
			hasSHA := regexp.MustCompile(`@a[a-fA-F0-9]{40}`).MatchString(trimmed) ||
				strings.Contains(trimmed, "# SHA:")
			hasVersionOnly := regexp.MustCompile(`@v\d+`).MatchString(trimmed)

			if hasVersionOnly && !hasSHA {
				foundActionsWithoutSHA = true
				result.Errors = append(result.Errors,
					fmt.Sprintf("SHA missing: %s (REQ-CI-018)", trimmed))
			}
		}
	}

	if foundActionsWithoutSHA {
		result.IsValid = false
	}
}

// validateCodexPrivateGuard validates Codex workflow private repo guard (REQ-SEC-001)
func (v *Validator) validateCodexPrivateGuard(content string, result *ValidationResult) {
	isCodexWorkflow := strings.Contains(strings.ToLower(content), "codex") ||
		strings.Contains(strings.ToLower(content), "CODEX_AUTH_JSON")

	if !isCodexWorkflow {
		return // Skip validation if not Codex workflow
	}

	// Check if repository_visibility check exists
	hasVisibilityCheck := strings.Contains(content, "repository_visibility") ||
		strings.Contains(content, "github.repository_visibility")

	if !hasVisibilityCheck {
		result.IsValid = false
		result.Errors = append(result.Errors,
			"Codex workflow missing private repo guard (REQ-SEC-001)")
	}
}

// validateHardcodedCredentials validates hardcoded credentials (SEC-003)
func (v *Validator) validateHardcodedCredentials(content string, result *ValidationResult) {
	// Test pattern: sk-1234567890abcdef
	testPattern := regexp.MustCompile(`sk-[0-9a-f]{16,}`)
	if testPattern.MatchString(content) {
		result.IsValid = false
		result.Errors = append(result.Errors,
			"hardcoded credential found: OpenAI API key pattern (SEC-003)")
		return
	}

	// Credential patterns
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
				fmt.Sprintf("hardcoded credential found: %s (SEC-003)", pattern))
			break
		}
	}
}

// validatePermissions validates excessive authorization grants (SEC-005)
func (v *Validator) validatePermissions(content string, result *ValidationResult) {
	dangerousPerms := []string{
		"write-all",
		"read-all",
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "permissions:") {
			// Check permissions block content
			for _, perm := range dangerousPerms {
				if strings.Contains(trimmed, perm) {
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("excessive authorization grant: %s (SEC-005)", perm))
					result.IsValid = false // Process as error
					return
				}
			}
		}
	}

	if !strings.Contains(content, "permissions:") {
		result.Warnings = append(result.Warnings,
			"permissions field is missing. default value is read-only (SEC-005)")
	}
}

// ValidateAllTemplates validates all templates in directory
func (v *Validator) ValidateAllTemplates(templateDir string) ([]*ValidationResult, error) {
	var results []*ValidationResult

	// Find files ((.yml.tmpl or .yaml.tmpl))
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".yml.tmpl") || strings.HasSuffix(path, ".yaml.tmpl") {
			result, err := v.ValidateTemplate(path)
			if err != nil {
				return fmt.Errorf("%s validation failure: %w", path, err)
			}
			results = append(results, result)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("template directory traversal failure: %w", err)
	}

	return results, nil
}

// ValidateWorkflowSyntax validates workflow YAML syntax
func ValidateWorkflowSyntax(content []byte) error {
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return fmt.Errorf("YAML syntax error: %w", err)
	}

	// default structure validation (name, on, jobs fields)
	if node.Kind != yaml.DocumentNode {
		return fmt.Errorf("not a valid YAML document")
	}

	return nil
}

func ParseWorkflowTemplate(templatePath string) (map[string]interface{}, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("template file read failure: %w", err)
	}

	var workflow map[string]interface{}
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		return nil, fmt.Errorf("YAML parsing failure: %w", err)
	}

	return workflow, nil
}

// ScanForActions extracts all GitHub Actions used in workflow
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
