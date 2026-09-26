package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"gopkg.in/yaml.v3"
)

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

// @MX:ANCHOR: [AUTO] Project structure validation factory. Shared entry point across many commands including init, update, and sync.
// @MX:REASON: fan_in=8, responsible for creating validator instances throughout the project lifecycle
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
	"state",
	"logs",
}

// configCacheArtifactName is the fixed cache file name the config loader
// writes under .moai/state/ (internal/config cache.go cacheFileName). It is
// duplicated here as a local constant because the canonical one is unexported
// and pulling the whole config package in would couple this validator to its
// loader. Keep in sync with internal/config/cache.go.
const configCacheArtifactName = "config-cache.json"

// moaiDirIsOnlyCacheArtifact reports whether dir contains nothing but the
// config cache artifact (.moai/state/config-cache.json) — the exact litter a
// pre-fix moai command left in an uninitialized directory (issue #1568), and
// therefore not evidence of an initialized project. Any other content
// (config/, manifest.json, specs/, ...) marks a real project and MUST keep
// failing validation as before.
func moaiDirIsOnlyCacheArtifact(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.Name() != "state" {
			return false
		}
	}

	stateEntries, err := os.ReadDir(filepath.Join(dir, "state"))
	if err != nil {
		// .moai/ without a readable state/ is not this hazard's shape.
		return false
	}
	for _, entry := range stateEntries {
		if entry.Name() != configCacheArtifactName {
			return false
		}
	}
	return true
}

// requiredClaudeDirs lists the directories that must exist under .claude/.
// Post SPEC-V3R6-AGENT-FOLDER-SPLIT-001: agents are split into 4 domain subfolders.
var requiredClaudeDirs = []string{
	"agents/core",
	"agents/expert",
	"agents/meta",
	"agents/harness",
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
		if moaiDirIsOnlyCacheArtifact(moaiDir) {
			// The directory holds nothing but the config cache artifact a
			// pre-fix moai command left behind (issue #1568): tool litter,
			// not an initialized project. Treat the project as
			// uninitialized so init succeeds without manual cleanup.
			result.Warnings = append(result.Warnings, ".moai/ contains only the config cache artifact; treating the directory as uninitialized.")
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, "project already initialized: .moai/ directory exists. Use --force to reinitialize.")
		}
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

	backupsDir := filepath.Join(root, defs.BackupsDir)
	if err := os.MkdirAll(backupsDir, defs.DirPerm); err != nil {
		return "", fmt.Errorf("create backups directory: %w", err)
	}

	timestamp := time.Now().Format(defs.BackupTimestampFormat)
	backupDir := filepath.Join(backupsDir, timestamp)

	if err := os.Rename(moaiDir, backupDir); err != nil {
		return "", fmt.Errorf("backup existing project: %w", err)
	}

	return backupDir, nil
}

// CarryManifestForward restores the manifest a --force backup moved aside, so
// the files the previous deployment recorded outside .moai/ keep their
// provenance: without it every one of them reads as untracked and is recorded
// user_created and skipped. A template_managed file whose content no longer
// matches its recorded hash was edited by the user and is reclassified
// user_modified, so the redeploy leaves it alone; so is one that is not a
// regular file or cannot be read, since it cannot be shown unedited. Keys
// outside the project root are carried unread. A missing or unloadable backup
// manifest carries nothing, so --force still runs as it did without one.
func CarryManifestForward(root, backupDir string) error {
	if backupDir == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(backupDir, defs.ManifestJSON))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read backed-up manifest: %w", err)
	}
	if !json.Valid(data) {
		return nil
	}
	moaiDir := filepath.Join(filepath.Clean(root), defs.MoAIDir)
	if err := os.MkdirAll(moaiDir, defs.DirPerm); err != nil {
		return fmt.Errorf("create %s: %w", defs.MoAIDir, err)
	}
	if err := os.WriteFile(filepath.Join(moaiDir, defs.ManifestJSON), data, defs.FilePerm); err != nil {
		return fmt.Errorf("carry manifest forward: %w", err)
	}

	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		// Valid JSON that is not a manifest: carry nothing, and remove what
		// Load and the copy left behind so the deploy starts from none.
		_ = os.Remove(filepath.Join(moaiDir, defs.ManifestJSON))
		_ = os.Remove(filepath.Join(moaiDir, defs.ManifestJSON+".corrupt"))
		return nil
	}
	files := mgr.Manifest().Files
	for rel, entry := range files {
		if entry.Provenance != manifest.TemplateManaged {
			continue
		}
		clean := filepath.Clean(filepath.FromSlash(rel))
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			continue
		}
		path := filepath.Join(root, clean)
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue // deleted: the redeploy restores it
		}
		current := ""
		if err == nil && info.Mode().IsRegular() {
			current, _ = manifest.HashFile(path)
		}
		if current == entry.CurrentHash {
			continue
		}
		entry.Provenance = manifest.UserModified
		if current != "" {
			entry.CurrentHash = current
		}
		files[rel] = entry
	}
	return mgr.Save()
}

// HealManifestFromBackups repairs a manifest an `init --force` that predates
// CarryManifestForward already damaged: that run recorded every file it found
// outside .moai/ as user_created, and moved the manifest that knew better to
// .moai-backups/<ts>/manifest.json. A user_created entry is restored from the
// newest backup that recorded the path template_managed — template_managed
// when the file still hashes to the backed-up value, user_modified otherwise.
// Backups are read newest first; a backup that also reads user_created is
// skipped, because a later damaged run leaves exactly that behind. Files on
// disk are never touched, and missing or unreadable backups heal nothing.
// It returns the number of entries it restored.
func HealManifestFromBackups(root string) (int, error) {
	root = filepath.Clean(root)
	// A corrupt live manifest is the update's own concern; Load would move
	// it aside, so leave it for the code that already handles it.
	if data, err := os.ReadFile(filepath.Join(root, defs.MoAIDir, defs.ManifestJSON)); err != nil || !json.Valid(data) {
		return 0, nil
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		return 0, nil
	}
	files := mgr.Manifest().Files
	pending := map[string]bool{}
	for rel, entry := range files {
		if entry.Provenance == manifest.UserCreated {
			pending[rel] = true
		}
	}
	if len(pending) == 0 {
		return 0, nil
	}
	backups, err := os.ReadDir(filepath.Join(root, defs.BackupsDir))
	if err != nil {
		return 0, nil
	}
	healed := 0
	for i := len(backups) - 1; i >= 0 && len(pending) > 0; i-- {
		if !backups[i].IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, defs.BackupsDir, backups[i].Name(), defs.ManifestJSON))
		if err != nil {
			continue
		}
		var backup manifest.Manifest
		if json.Unmarshal(data, &backup) != nil {
			continue
		}
		for rel := range pending {
			old, ok := backup.Files[rel]
			if !ok || old.Provenance != manifest.TemplateManaged {
				continue
			}
			delete(pending, rel)
			clean := filepath.Clean(filepath.FromSlash(rel))
			if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
				continue
			}
			current := ""
			if info, err := os.Lstat(filepath.Join(root, clean)); err == nil && info.Mode().IsRegular() {
				current, _ = manifest.HashFile(filepath.Join(root, clean))
			}
			if current != "" && current == old.CurrentHash {
				files[rel] = old
			} else {
				entry := files[rel]
				entry.Provenance = manifest.UserModified
				if current != "" {
					entry.CurrentHash = current
				}
				files[rel] = entry
			}
			healed++
		}
	}
	if healed == 0 {
		return 0, nil
	}
	if err := mgr.Save(); err != nil {
		return 0, fmt.Errorf("save healed manifest: %w", err)
	}
	return healed, nil
}
