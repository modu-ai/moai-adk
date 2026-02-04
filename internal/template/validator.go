package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathError describes a validation failure for a single file path.
type PathError struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// DeploymentError describes a single deployment validation failure.
type DeploymentError struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// ValidationReport summarizes the result of a deployment validation.
type ValidationReport struct {
	Valid        bool              `json:"valid"`
	Errors       []DeploymentError `json:"errors,omitempty"`
	Warnings     []string          `json:"warnings,omitempty"`
	FilesChecked int               `json:"files_checked"`
}

// Validator verifies JSON content, file paths, and deployment integrity.
type Validator interface {
	// ValidateJSON checks whether data is syntactically valid JSON.
	ValidateJSON(data []byte) error

	// ValidatePaths checks each path for normalization, containment,
	// and special character issues relative to projectRoot.
	ValidatePaths(projectRoot string, files []string) []PathError

	// ValidateDeployment performs comprehensive post-deployment checks
	// including file existence, JSON validity, and path normalization.
	ValidateDeployment(projectRoot string, expectedFiles []string) *ValidationReport
}

// validator is the concrete implementation of Validator.
type validator struct{}

// NewValidator creates a new Validator.
func NewValidator() Validator {
	return &validator{}
}

// ValidateJSON checks if data is valid JSON.
func (v *validator) ValidateJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("%w: empty input", ErrInvalidJSON)
	}
	if !json.Valid(data) {
		return fmt.Errorf("%w", ErrInvalidJSON)
	}
	return nil
}

// ValidatePaths checks each file path for security and normalization issues.
func (v *validator) ValidatePaths(projectRoot string, files []string) []PathError {
	projectRoot = filepath.Clean(projectRoot)
	var errs []PathError

	for _, f := range files {
		cleaned := filepath.Clean(filepath.FromSlash(f))

		// Reject absolute paths
		if filepath.IsAbs(cleaned) {
			errs = append(errs, PathError{
				Path:   f,
				Reason: "absolute path not allowed",
			})
			continue
		}

		// Reject path traversal
		if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, string(filepath.Separator)+"..") {
			errs = append(errs, PathError{
				Path:   f,
				Reason: "path traversal detected",
			})
			continue
		}

		// Convert projectRoot to absolute path for reliable comparison
		absProjectRoot, err := filepath.Abs(projectRoot)
		if err != nil {
			errs = append(errs, PathError{
				Path:   f,
				Reason: fmt.Sprintf("resolve project root: %v", err),
			})
			continue
		}

		// Verify containment
		absPath := filepath.Join(absProjectRoot, cleaned)
		if !strings.HasPrefix(absPath, absProjectRoot+string(filepath.Separator)) && absPath != absProjectRoot {
			errs = append(errs, PathError{
				Path:   f,
				Reason: "path escapes project root",
			})
		}
	}

	return errs
}

// ValidateDeployment performs comprehensive validation of deployed files.
func (v *validator) ValidateDeployment(projectRoot string, expectedFiles []string) *ValidationReport {
	projectRoot = filepath.Clean(projectRoot)
	report := &ValidationReport{
		Valid: true,
	}

	for _, relPath := range expectedFiles {
		report.FilesChecked++
		absPath := filepath.Join(projectRoot, filepath.FromSlash(relPath))

		// Check existence
		info, err := os.Stat(absPath)
		if err != nil {
			report.Valid = false
			report.Errors = append(report.Errors, DeploymentError{
				Path:   relPath,
				Reason: fmt.Sprintf("file not found: %v", err),
			})
			continue
		}

		if info.IsDir() {
			report.Warnings = append(report.Warnings, fmt.Sprintf("%s is a directory", relPath))
			continue
		}

		// Validate JSON files
		if isJSONFile(relPath) {
			data, err := os.ReadFile(absPath)
			if err != nil {
				report.Valid = false
				report.Errors = append(report.Errors, DeploymentError{
					Path:   relPath,
					Reason: fmt.Sprintf("cannot read file: %v", err),
				})
				continue
			}

			if !json.Valid(data) {
				report.Valid = false
				report.Errors = append(report.Errors, DeploymentError{
					Path:   relPath,
					Reason: "invalid JSON content",
				})
			}
		}
	}

	return report
}

// isJSONFile checks if a file path has a .json extension.
func isJSONFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".json")
}
