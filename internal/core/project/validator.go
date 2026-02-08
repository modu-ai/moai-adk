package project

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// BackupTimestampFormat is the Go time layout for backup directory names (YYYYMMDD_HHMMSS).
// Deprecated: Use defs.BackupTimestampFormat directly.
const BackupTimestampFormat = defs.BackupTimestampFormat

// BackupsDir is the directory name where project backups are stored.
// Deprecated: Use defs.BackupsDir directly.
const BackupsDir = defs.BackupsDir

// ProjectValidator checks project structure integrity.
type ProjectValidator interface {
	// Validate checks the overall project structure.
	Validate(root string) (*ValidationResult, error)

	// ValidateMoAI checks MoAI-specific configuration and file integrity.
	ValidateMoAI(root string) (*ValidationResult, error)
}

// ValidationResult holds project validation outcomes.
type ValidationResult struct {
	Valid    bool     // True if no errors found.
	Errors   []string // Critical issues that prevent operation.
	Warnings []string // Non-critical issues.
}

// projectValidator is the concrete implementation of ProjectValidator.
type projectValidator struct {
	logger *slog.Logger
}

// NewValidator creates a new ProjectValidator.
func NewValidator(logger *slog.Logger) ProjectValidator {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &projectValidator{logger: logger}
}

// requiredMoAIDirs lists the directories that must exist under .moai/.
var requiredMoAIDirs = []string{
	"config/sections",
	"specs",
	"reports",
	"memory",
	"logs",
}

// requiredClaudeDirs lists the directories that must exist under .claude/.
var requiredClaudeDirs = []string{
	"agents/moai",
	"skills",
	"commands/moai",
	"rules/moai",
}

// Validate checks the overall project structure for MoAI initialization.
func (v *projectValidator) Validate(root string) (*ValidationResult, error) {
	root = filepath.Clean(root)
	if err := validateRoot(root); err != nil {
		return nil, err
	}

	v.logger.Debug("validating project structure", "root", root)

	result := &ValidationResult{Valid: true}

	// Check if .moai/ already exists
	moaiDir := filepath.Join(root, defs.MoAIDir)
	if dirExists(moaiDir) {
		result.Valid = false
		result.Errors = append(result.Errors, "project already initialized: .moai/ directory exists. Use --force to reinitialize.")
	}

	// Check if .claude/ already exists
	claudeDir := filepath.Join(root, defs.ClaudeDir)
	if dirExists(claudeDir) {
		result.Warnings = append(result.Warnings, ".claude/ directory already exists; templates may be updated.")
	}

	// Check if CLAUDE.md already exists
	claudeMD := filepath.Join(root, defs.ClaudeMD)
	if fileExists(claudeMD) {
		result.Warnings = append(result.Warnings, "CLAUDE.md already exists; it will be updated.")
	}

	// Check Git repository
	gitDir := filepath.Join(root, ".git")
	if !dirExists(gitDir) {
		result.Warnings = append(result.Warnings, "Git repository not detected. Some features may be limited.")
	}

	return result, nil
}

// ValidateMoAI checks MoAI-specific configuration and file integrity.
func (v *projectValidator) ValidateMoAI(root string) (*ValidationResult, error) {
	root = filepath.Clean(root)
	if err := validateRoot(root); err != nil {
		return nil, err
	}

	v.logger.Debug("validating MoAI structure", "root", root)

	result := &ValidationResult{Valid: true}

	moaiDir := filepath.Join(root, defs.MoAIDir)
	if !dirExists(moaiDir) {
		result.Valid = false
		result.Errors = append(result.Errors, ".moai/ directory not found. Run 'moai init' first.")
		return result, nil
	}

	// Check required .moai/ subdirectories
	for _, subdir := range requiredMoAIDirs {
		dirPath := filepath.Join(moaiDir, subdir)
		if !dirExists(dirPath) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("missing required directory: .moai/%s", subdir))
		}
	}

	// Check YAML config files are parseable
	sectionsDir := filepath.Join(moaiDir, defs.SectionsSubdir)
	if dirExists(sectionsDir) {
		v.validateYAMLFiles(sectionsDir, result)
	}

	// Check manifest.json is valid JSON
	manifestPath := filepath.Join(moaiDir, defs.ManifestJSON)
	if fileExists(manifestPath) {
		v.validateJSONFile(manifestPath, result)
	} else {
		result.Warnings = append(result.Warnings, "manifest.json not found.")
	}

	// Check .claude/ directories
	claudeDir := filepath.Join(root, defs.ClaudeDir)
	if dirExists(claudeDir) {
		for _, subdir := range requiredClaudeDirs {
			dirPath := filepath.Join(claudeDir, subdir)
			if !dirExists(dirPath) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("missing directory: .claude/%s", subdir))
			}
		}
	} else {
		result.Warnings = append(result.Warnings, ".claude/ directory not found.")
	}

	// Check CLAUDE.md exists
	claudeMD := filepath.Join(root, defs.ClaudeMD)
	if !fileExists(claudeMD) {
		result.Warnings = append(result.Warnings, "CLAUDE.md not found.")
	}

	return result, nil
}

// validateYAMLFiles checks that all .yaml files in a directory are parseable.
func (v *projectValidator) validateYAMLFiles(dir string, result *ValidationResult) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("cannot read config directory: %s", err))
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("cannot read %s: %s", entry.Name(), err))
			continue
		}

		var raw any
		if err := yaml.Unmarshal(data, &raw); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("invalid YAML in %s: %s", entry.Name(), err))
		}
	}
}

// validateJSONFile checks that a JSON file is valid.
func (v *projectValidator) validateJSONFile(path string, result *ValidationResult) {
	data, err := os.ReadFile(path)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("cannot read %s: %s", filepath.Base(path), err))
		return
	}

	if !json.Valid(data) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid JSON in %s", filepath.Base(path)))
	}
}

// BackupExistingProject moves .moai/ to .moai-backups/{timestamp}/.
// Returns the backup path or an error.
func BackupExistingProject(root string) (string, error) {
	root = filepath.Clean(root)
	moaiDir := filepath.Join(root, defs.MoAIDir)

	if !dirExists(moaiDir) {
		return "", nil // nothing to backup
	}

	backupsDir := filepath.Join(root, BackupsDir)
	if err := os.MkdirAll(backupsDir, defs.DirPerm); err != nil {
		return "", fmt.Errorf("create backups directory: %w", err)
	}

	timestamp := time.Now().Format(BackupTimestampFormat)
	backupDir := filepath.Join(backupsDir, timestamp)

	if err := os.Rename(moaiDir, backupDir); err != nil {
		return "", fmt.Errorf("backup existing project: %w", err)
	}

	return backupDir, nil
}
