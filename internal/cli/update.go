package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/shell"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	// maxConfigSize is the maximum allowed size for config.yaml to prevent DoS
	maxConfigSize = 10 * 1024 * 1024 // 10MB
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Sync MoAI-ADK project templates to the latest version",
	Long:  "Synchronize embedded templates with the project. Binary updates happen automatically at session start.",
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Bool("check", false, "Check if a newer binary version is available (informational)")
	updateCmd.Flags().Bool("shell-env", false, "Configure shell environment variables for Claude Code")
	updateCmd.Flags().BoolP("config", "c", false, "Edit project configuration (same as init wizard)")
	updateCmd.Flags().Bool("force", false, "Skip backup and force the update")
	updateCmd.Flags().Bool("yes", false, "Auto-confirm all prompts (CI/CD mode)")
}

// runUpdate synchronizes embedded templates with the project directory.
// Binary self-updates now happen automatically at session start.
//
// Flags:
//
//	-c, --config: Edit project configuration (same as init wizard)
//	--check: Check if a newer binary version is available (informational)
//	--force: Skip backup and force the update
//	--shell-env: Configure shell environment variables
//	--yes: Auto-confirm all prompts (CI/CD mode)
func runUpdate(cmd *cobra.Command, _ []string) error {
	checkOnly := getBoolFlag(cmd, "check")
	shellEnv := getBoolFlag(cmd, "shell-env")
	editConfig := getBoolFlag(cmd, "config")
	out := cmd.OutOrStdout()

	// Handle --config / -c mode (edit configuration only, no template updates)
	// This takes priority over all other flags
	if editConfig {
		return runInitWizard(cmd, true) // true = reconfigure mode
	}

	currentVersion := version.GetVersion()
	_, _ = fmt.Fprintf(out, "Current version: moai-adk %s\n", currentVersion)

	// Handle shell-env mode
	if shellEnv {
		return runShellEnvConfig(cmd)
	}

	// Handle --check mode (informational: check if newer binary exists)
	if checkOnly {
		// Lazily initialize update dependencies
		if deps != nil {
			if err := deps.EnsureUpdate(); err != nil {
				deps.Logger.Debug("failed to initialize update system", "error", err)
			}
		}

		if deps == nil || deps.UpdateChecker == nil {
			_, _ = fmt.Fprintln(out, "Update checker not available. Using current version.")
			return nil
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		info, err := deps.UpdateChecker.CheckLatest(ctx)
		if err != nil {
			return fmt.Errorf("check latest version: %w", err)
		}
		_, _ = fmt.Fprintf(out, "Latest version:  %s\n", info.Version)
		_, _ = fmt.Fprintln(out, "\nNote: Binary updates happen automatically at session start.")
		return nil
	}

	// Default: template sync only
	return runTemplateSyncWithProgress(cmd)
}

// runTemplateSync synchronizes embedded templates with the project directory.
// It performs a quick version comparison first - if the project's template version
// matches the package version, the sync is skipped for performance (70-80% faster).
//
// Template deployment uses a 3-way merge strategy to preserve local modifications.
// Users are prompted to confirm the merge before proceeding.
func runTemplateSync(cmd *cobra.Command) error {
	return runTemplateSyncWithReporter(cmd, nil, false)
}

// runTemplateSyncWithReporter synchronizes templates with progress reporting.
func runTemplateSyncWithReporter(cmd *cobra.Command, reporter project.ProgressReporter, skipConfirm bool) error {
	out := cmd.OutOrStdout()
	ctx := cmd.Context()

	// Get flags for template sync
	forceBackup := getBoolFlag(cmd, "force")
	autoConfirm := getBoolFlag(cmd, "yes")

	// Use current directory as project root
	projectRoot := "."

	currentVersion := version.GetVersion()
	_, _ = fmt.Fprintf(out, "Current version: moai-adk %s\n", currentVersion)
	_, _ = fmt.Fprintln(out, "Syncing templates from embedded filesystem...")

	if reporter != nil {
		reporter.StepStart("Version Check", "Checking template version...")
	}

	// Stage 2: Config Version Comparison (before template sync)
	// Compare package template_version with project config template_version
	// If versions match, skip sync for performance (70-80% faster)
	packageVersion := version.GetVersion()
	projectVersion, err := getProjectConfigVersion(projectRoot)
	if err == nil && packageVersion == projectVersion {
		if reporter != nil {
			reporter.StepComplete("Already up-to-date")
		}
		_, _ = fmt.Fprintln(out, "\n✓ Template version up-to-date. Skipping sync.")
		return nil
	}

	if reporter != nil {
		reporter.StepComplete("Version check complete")
	}

	if reporter != nil {
		reporter.StepStart("Loading Templates", "Reading embedded templates...")
	}

	// Load embedded templates
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		if reporter != nil {
			reporter.StepError(err)
		}
		return fmt.Errorf("load embedded templates: %w", err)
	}

	if reporter != nil {
		reporter.StepComplete("Templates loaded")
	}

	// Generate template settings.json for merge using SettingsGenerator
	// Note: settings.json is NOT in embedded templates (runtime-generated per ADR-011)
	settingsGen := template.NewSettingsGenerator()
	templateSettingsData, err := settingsGen.Generate(nil, runtime.GOOS)
	if err != nil {
		if reporter != nil {
			reporter.StepError(err)
		}
		return fmt.Errorf("generate template settings.json: %w", err)
	}

	if reporter != nil {
		reporter.StepStart("Loading Manifest", "Reading project manifest...")
	}

	// Initialize manifest manager
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		if reporter != nil {
			reporter.StepError(err)
		}
		return fmt.Errorf("load manifest: %w", err)
	}

	if reporter != nil {
		reporter.StepComplete("Manifest loaded")
	}

	// Create renderer for template variable substitution
	renderer := template.NewRenderer(embedded)

	// Create deployer with renderer and force update enabled for template sync
	// This ensures template files are rendered (.tmpl -> actual file) and updated even if they exist
	deployer := template.NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)

	// Analyze merge and get user confirmation
	analysis := analyzeMergeChanges(deployer, projectRoot)

	_, _ = fmt.Fprintln(out, "\nAnalyzing merge changes...")

	if reporter != nil {
		reporter.StepUpdate("Found " + fmt.Sprintf("%d files to sync", len(analysis.Files)))
	}

	// Skip confirmation if --yes flag is provided (CI/CD mode) or pre-confirmed
	var proceed bool
	if skipConfirm {
		proceed = true
	} else if autoConfirm {
		proceed = true
		_, _ = fmt.Fprintln(out, "Auto-confirming merge (CI/CD mode)...")
	} else {
		var err error
		proceed, err = merge.ConfirmMerge(analysis)
		if err != nil {
			if reporter != nil {
				reporter.StepError(err)
			}
			return fmt.Errorf("confirm merge for %d files (risk: %s): %w",
				len(analysis.Files), analysis.RiskLevel, err)
		}
	}

	if !proceed {
		_, _ = fmt.Fprintln(out, "\nMerge cancelled by user.")
		if reporter != nil {
			reporter.StepError(errors.New("cancelled by user"))
		}
		return nil
	}

	// Deploy templates
	_, _ = fmt.Fprintln(out, "\nProceeding with template deployment...")
	_, _ = fmt.Fprintln(out)

	// Define deployment steps
	steps := []struct {
		name    string
		message string
		execute func() error
	}{
		{
			name:    "Backup",
			message: "Backing up configuration",
			execute: func() error {
				if forceBackup {
					_, _ = fmt.Fprintf(out, "  ○ Skipping backup (--force mode)...\n")
					return nil
				}

				_, _ = fmt.Fprintf(out, "  ○ Backing up .moai/config...")
				configBackupPath, backupErr := backupMoaiConfig(projectRoot)
				if backupErr != nil {
					_, _ = fmt.Fprintf(out, "\r  ✗ Backup failed: %v\n", backupErr)
					return backupErr
				}
				if configBackupPath != "" {
					_, _ = fmt.Fprintf(out, "\r  ✓ .moai/config backed up\n")
				} else {
					_, _ = fmt.Fprintln(out, "\r  - No config to backup")
				}
				return nil
			},
		},
		{
			name:    "Merge Settings",
			message: "Merging settings.json",
			execute: func() error {
				settingsPath := filepath.Join(projectRoot, defs.ClaudeDir, defs.SettingsJSON)
				if len(templateSettingsData) == 0 {
					return nil
				}
				if _, err := os.Stat(settingsPath); err != nil {
					return nil // No existing settings
				}

				_, _ = fmt.Fprintf(out, "  ○ Merging settings.json...")
				tmpFile, tmpErr := os.CreateTemp("", "settings-template-*.json")
				if tmpErr != nil {
					_, _ = fmt.Fprintf(out, "\r  ✗ Failed to create temp file: %v\n", tmpErr)
					return tmpErr
				}
				tmpPath := tmpFile.Name()
				defer func() { _ = os.Remove(tmpPath) }()
				if _, writeErr := tmpFile.Write(templateSettingsData); writeErr != nil {
					_, _ = fmt.Fprintf(out, "\r  ✗ Failed to write temp file: %v\n", writeErr)
					_ = tmpFile.Close()
					return writeErr
				}
				_ = tmpFile.Close()
				if mergeErr := mergeSettingsJSON(tmpPath, settingsPath); mergeErr != nil {
					_, _ = fmt.Fprintf(out, "\r  ✗ Merge failed: %v\n", mergeErr)
					return mergeErr
				}
				_, _ = fmt.Fprintln(out, "\r  ✓ settings.json merged")
				return nil
			},
		},
		{
			name:    "Deploy Templates",
			message: "Deploying template files",
			execute: func() error {
				_, _ = fmt.Fprintf(out, "  ○ Deploying templates...")

				// Build TemplateContext with detected paths for template rendering
				homeDir, _ := os.UserHomeDir()
				goBinPath := detectGoBinPathForUpdate(homeDir)
				tmplCtx := template.NewTemplateContext(
					template.WithGoBinPath(goBinPath),
					template.WithHomeDir(homeDir),
				)

				if deployErr := deployer.Deploy(ctx, projectRoot, mgr, tmplCtx); deployErr != nil {
					_, _ = fmt.Fprintf(out, "\r  ✗ Deployment failed: %v\n", deployErr)
					return fmt.Errorf("deploy templates: %w", deployErr)
				}
				_, _ = fmt.Fprintln(out, "\r  ✓ Templates deployed")
				return nil
			},
		},
		{
			name:    "Restore Settings",
			message: "Restoring user settings",
			execute: func() error {
				// This step's status is tracked via configBackupPath variable
				// We'll handle this in the main flow
				return nil
			},
		},
	}

	// Track config backup path for restore step
	var configBackupPath string

	// Execute each step with progress reporting
	for i, step := range steps {
		if reporter != nil {
			reporter.StepStart(step.name, step.message)
		}

		// Special handling for backup step to capture backup path
		if step.name == "Backup" && !forceBackup {
			_, _ = fmt.Fprintf(out, "  ○ Backing up .moai/config...")
			var backupErr error
			configBackupPath, backupErr = backupMoaiConfig(projectRoot)
			if backupErr != nil {
				_, _ = fmt.Fprintf(out, "\r  ✗ Backup failed: %v\n", backupErr)
				if reporter != nil {
					reporter.StepError(backupErr)
				}
				return backupErr
			}
			if configBackupPath != "" {
				_, _ = fmt.Fprintf(out, "\r  ✓ .moai/config backed up\n")
			} else {
				_, _ = fmt.Fprintln(out, "\r  - No config to backup")
			}
			if reporter != nil {
				reporter.StepComplete("Configuration backed up")
			}
		} else if step.name == "Restore Settings" {
			// Handle restore step with captured backup path
			if configBackupPath != "" {
				if reporter != nil {
					reporter.StepStart("Restore Settings", "Restoring user settings")
				}
				_, _ = fmt.Fprintf(out, "  ○ Restoring user settings...")
				if restoreErr := restoreMoaiConfig(projectRoot, configBackupPath); restoreErr != nil {
					_, _ = fmt.Fprintf(out, "\r  ✗ Restore failed: %v\n", restoreErr)
					if reporter != nil {
						reporter.StepError(restoreErr)
					}
					return restoreErr
				}
				_, _ = fmt.Fprintln(out, "\r  ✓ User settings restored")
				deletedCount := cleanup_old_backups(projectRoot, 5)
				if deletedCount > 0 {
					_, _ = fmt.Fprintf(out, "  ✓ Cleaned up %d old backup(s)\n", deletedCount)
				}
				if reporter != nil {
					reporter.StepComplete("Settings restored")
				}
			}
		} else {
			// Execute normal step
			if err := step.execute(); err != nil {
				if reporter != nil {
					reporter.StepError(err)
				}
				return err
			}
			if reporter != nil {
				reporter.StepComplete(fmt.Sprintf("%s complete", step.name))
			}
		}

		// Update progress for remaining steps
		if reporter != nil && i < len(steps)-1 {
			reporter.StepUpdate(fmt.Sprintf("%d/%d steps complete", i+1, len(steps)))
		}
	}

	_, _ = fmt.Fprintln(out, "\n✓ Template sync complete.")

	// Ensure global settings.json has required env variables
	if err := ensureGlobalSettingsEnv(); err != nil {
		_, _ = fmt.Fprintf(out, "Warning: Failed to update global settings env: %v\n", err)
	}

	return nil
}

// runTemplateSyncWithProgress runs template sync with simple console output.
func runTemplateSyncWithProgress(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	projectRoot := "."
	autoConfirm := getBoolFlag(cmd, "yes")

	// Use simple console output for progress reporting
	consoleReporter := project.NewConsoleReporter()

	// Check for version match before proceeding
	packageVersion := version.GetVersion()
	projectVersion, err := getProjectConfigVersion(projectRoot)
	if err == nil && packageVersion == projectVersion {
		_, _ = fmt.Fprintln(out, "\n✓ Template version up-to-date. Skipping sync.")
		return nil
	}

	// Confirm merge before proceeding (unless auto-confirm is set)
	if !autoConfirm {
		embedded, err := template.EmbeddedTemplates()
		if err != nil {
			return fmt.Errorf("load embedded templates: %w", err)
		}

		deployer := template.NewDeployerWithForceUpdate(embedded, true)
		analysis := analyzeMergeChanges(deployer, projectRoot)

		_, _ = fmt.Fprintln(out, "\nAnalyzing merge changes...")
		proceed, err := merge.ConfirmMerge(analysis)
		if err != nil {
			return fmt.Errorf("confirm merge for %d files (risk: %s): %w",
				len(analysis.Files), analysis.RiskLevel, err)
		}
		if !proceed {
			_, _ = fmt.Fprintln(out, "\nMerge cancelled by user.")
			return nil
		}
	}

	return runTemplateSyncWithReporter(cmd, consoleReporter, false)
}

// classifyFileRisk determines the risk level for a file modification.
// Returns "high" for core config files (CLAUDE.md, settings.json, config.yaml),
// "low" for new files, and "medium" for existing file updates.
func classifyFileRisk(filename string, exists bool) string {
	base := filepath.Base(filename)

	// High risk files
	highRiskFiles := []string{"CLAUDE.md", "settings.json", "config.yaml"}
	for _, high := range highRiskFiles {
		if base == high {
			return "high"
		}
	}

	// New files are low risk
	if !exists {
		return "low"
	}

	// Existing files are medium risk
	return "medium"
}

// determineStrategy selects the appropriate merge strategy based on file type.
// Returns SectionMerge for CLAUDE.md, EntryMerge for .gitignore, JSONMerge for .json,
// YAMLDeep for .yaml/.yml, and LineMerge as default.
func determineStrategy(filename string) merge.MergeStrategy {
	base := filepath.Base(filename)
	ext := filepath.Ext(filename)

	switch {
	case base == "CLAUDE.md":
		return merge.SectionMerge
	case base == ".gitignore":
		return merge.EntryMerge
	case ext == ".json":
		return merge.JSONMerge
	case ext == ".yaml" || ext == ".yml":
		return merge.YAMLDeep
	default:
		return merge.LineMerge
	}
}

// determineChangeType returns a user-friendly description of the change type.
// Returns "update existing" if the file exists, otherwise "new file".
func determineChangeType(exists bool) string {
	if exists {
		return "update existing"
	}
	return "new file"
}

// analyzeFiles examines each template file and returns detailed analysis results.
// For each template, it checks if the file exists, classifies its risk level,
// determines the appropriate merge strategy, and identifies the change type.
//
// Filters out moai* skills from the analysis since they are managed by MoAI-ADK
// and users typically don't need to see them in the merge confirmation UI.
//
// For .tmpl files, displays the rendered target path (without .tmpl extension)
// since that's what users will see in their project.
func analyzeFiles(templates []string, projectRoot string) []merge.FileAnalysis {
	var files []merge.FileAnalysis
	for _, tmpl := range templates {
		// Filter out MoAI-managed files - they are automatically installed
		if isMoaiManaged(tmpl) {
			continue
		}

		// Handle .tmpl files: use rendered target path for display and existence check
		displayPath := tmpl
		targetPath := filepath.Join(projectRoot, tmpl)
		if strings.HasSuffix(tmpl, ".tmpl") {
			// .tmpl files are rendered to paths without the .tmpl extension
			displayPath = strings.TrimSuffix(tmpl, ".tmpl")
			targetPath = filepath.Join(projectRoot, displayPath)
		}

		_, err := os.Stat(targetPath)
		exists := err == nil

		// Classify risk and determine strategy (use displayPath for classification)
		risk := classifyFileRisk(displayPath, exists)
		strategy := determineStrategy(displayPath)
		changeType := determineChangeType(exists)

		files = append(files, merge.FileAnalysis{
			Path:      displayPath,
			Changes:   changeType,
			Strategy:  strategy,
			RiskLevel: risk,
			Note:      "",
		})
	}
	return files
}

// isMoaiManaged returns true if the path is managed by MoAI-ADK and should be excluded from merge confirmation.
// MoAI-managed paths include:
//   - .claude/skills/moai-* and .claude/skills/moai/
//   - .claude/rules/moai/
//   - .claude/agents/moai/
//   - .claude/commands/moai/
//   - .claude/output-styles/moai/
//   - .moai/config/ (entire directory)
//
// These paths are automatically deleted and reinstalled without user confirmation.
func isMoaiManaged(path string) bool {
	// Check .moai/config/ paths first
	if strings.HasPrefix(path, ".moai/config/") || strings.HasPrefix(path, ".moai\\config\\") {
		return true
	}

	// Check if path is in .claude directory
	if !strings.Contains(path, ".claude") {
		return false
	}

	// Split by both '/' and '\' for cross-platform compatibility.
	// Template manifests always use '/' but filepath.Separator is '\' on Windows.
	parts := strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	for i, part := range parts {
		switch part {
		case "skills", "rules", "agents", "commands", "output-styles":
			// Check if the next directory starts with "moai-"
			if i+1 < len(parts) {
				itemName := parts[i+1]
				return strings.HasPrefix(itemName, "moai-") || strings.HasPrefix(itemName, "moai")
			}
		}
	}

	return false
}

// buildMergeAnalysis creates a summary from individual file analysis results.
// It counts high/medium/low risk files, determines overall risk level,
// identifies conflicts, and generates a human-readable summary.
func buildMergeAnalysis(files []merge.FileAnalysis) merge.MergeAnalysis {
	var highRisk, medRisk, lowRisk int
	for _, f := range files {
		switch f.RiskLevel {
		case "high":
			highRisk++
		case "medium":
			medRisk++
		case "low":
			lowRisk++
		}
	}

	overallRisk := "low"
	hasConflicts := false
	if highRisk > 0 {
		overallRisk = "high"
		hasConflicts = true
	} else if medRisk > 0 {
		overallRisk = "medium"
	}

	summary := fmt.Sprintf("Found %d files to sync", len(files))
	if highRisk > 0 {
		summary += fmt.Sprintf(" (%d high-risk files)", highRisk)
	}

	return merge.MergeAnalysis{
		Files:        files,
		HasConflicts: hasConflicts,
		SafeToMerge:  highRisk == 0,
		Summary:      summary,
		RiskLevel:    overallRisk,
	}
}

// analyzeMergeChanges performs a quick analysis of template files that will be modified.
// It evaluates risk levels based on file types and existing content:
//   - High risk: CLAUDE.md, settings.json, config.yaml (core config files)
//   - Medium risk: Existing files being updated
//   - Low risk: New files being created
//
// Returns a MergeAnalysis with file-by-file risk assessment and recommended strategies.
func analyzeMergeChanges(deployer template.Deployer, projectRoot string) merge.MergeAnalysis {
	templates := deployer.ListTemplates()
	files := analyzeFiles(templates, projectRoot)
	return buildMergeAnalysis(files)
}

// runShellEnvConfig configures shell environment variables for Claude Code.
func runShellEnvConfig(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintln(out, "Configuring shell environment for Claude Code...")

	// Get recommendation first
	configurator := shell.NewEnvConfigurator(nil)
	rec := configurator.GetRecommendation()

	_, _ = fmt.Fprintf(out, "  Shell: %s\n", rec.Shell)
	_, _ = fmt.Fprintf(out, "  Config file: %s\n", rec.ConfigFile)
	_, _ = fmt.Fprintf(out, "  Explanation: %s\n", rec.Explanation)
	_, _ = fmt.Fprintln(out, "  Changes to add:")
	for _, change := range rec.Changes {
		_, _ = fmt.Fprintf(out, "    - %s\n", change)
	}
	_, _ = fmt.Fprintln(out)

	// Execute configuration
	result, err := configurator.Configure(shell.ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
		PreferLoginShell:        true,
	})
	if err != nil {
		return fmt.Errorf("configure shell environment: %w", err)
	}

	if result.Skipped {
		_, _ = fmt.Fprintf(out, "Shell environment already configured in %s\n", result.ConfigFile)
	} else {
		_, _ = fmt.Fprintf(out, "Shell environment configured in %s\n", result.ConfigFile)
		_, _ = fmt.Fprintln(out, "Please restart your terminal or run:")
		_, _ = fmt.Fprintf(out, "  source %s\n", result.ConfigFile)
	}

	return nil
}

// getProjectConfigVersion reads the template_version from .moai/config/config.yaml.
// Returns "0.0.0" if the file doesn't exist or parsing fails, which triggers a sync.
// This enables the version comparison optimization in runTemplateSync.
func getProjectConfigVersion(projectRoot string) (string, error) {
	configPath := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir, defs.ConfigYAML)

	// Check file size before reading to prevent DoS
	info, err := os.Stat(configPath)
	if err != nil {
		// If config file doesn't exist, return "0.0.0" to force update
		if os.IsNotExist(err) {
			return "0.0.0", nil
		}
		return "", fmt.Errorf("stat config file: %w", err)
	}

	// Reject files larger than 10MB
	if info.Size() > maxConfigSize {
		return "", fmt.Errorf("config file too large: %d bytes (max: %d)", info.Size(), maxConfigSize)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("read config file: %w", err)
	}

	// Parse YAML to extract project.template_version
	var config struct {
		Project struct {
			TemplateVersion string `yaml:"template_version"`
		} `yaml:"project"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("parse config YAML: %w", err)
	}

	// If template_version is not set, return "0.0.0" to force update
	if config.Project.TemplateVersion == "" {
		return "0.0.0", nil
	}

	return config.Project.TemplateVersion, nil
}

// backupMoaiConfig creates a backup of .moai/config/ directory.
// Creates a timestamped backup under .moai-backups/YYYYMMDD_HHMMSS/ and
// excludes config/sections/ (user settings) from backup.
// Returns the backup directory path, or empty string if directory doesn't exist.
func backupMoaiConfig(projectRoot string) (string, error) {
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)

	// Check if config directory exists
	info, err := os.Stat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No config to backup
		}
		return "", fmt.Errorf("stat config directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("config path is not a directory")
	}

	timestamp := time.Now().Format(defs.BackupTimestampFormat)
	backupDir := filepath.Join(projectRoot, defs.BackupsDir, timestamp)

	// Create backup directory
	if err := os.MkdirAll(backupDir, defs.DirPerm); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}

	// Paths excluded from backups (protect user settings)
	// Note: relPath from configDir will be "sections", not "config/sections"
	excludedDirs := []string{"sections"}

	// Track backed up items and excluded items for metadata
	backedUpItems := []string{}
	excludedItems := []string{}

	// Copy all files from config to backup, excluding sections directory
	err = filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(configDir, path)
		if err != nil {
			return err
		}

		// Check for exclusion first - both directory and file level
		for _, excludedDir := range excludedDirs {
			if relPath == excludedDir || strings.HasPrefix(relPath, excludedDir+string(filepath.Separator)) {
				// Track excluded item
				excludedItems = append(excludedItems, relPath)
				// Skip this file or directory
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip directories that are not excluded
		if info.IsDir() {
			return nil
		}

		// Get relative path from backup directory
		// Use forward slashes for consistent metadata across platforms
		backupRelPath := filepath.ToSlash(filepath.Join(defs.MoAIDir, defs.ConfigSubdir, relPath))
		backedUpItems = append(backedUpItems, backupRelPath)

		backupPath := filepath.Join(backupDir, relPath)
		if err := os.MkdirAll(filepath.Dir(backupPath), defs.DirPerm); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(backupPath, data, defs.FilePerm)
	})

	if err != nil {
		_ = os.RemoveAll(backupDir)
		return "", fmt.Errorf("copy config files: %w", err)
	}

	// Create backup metadata file
	metadata := BackupMetadata{
		Timestamp:     timestamp,
		Description:   "config_backup",
		BackedUpItems: backedUpItems,
		ExcludedItems: excludedItems,
		ExcludedDirs:  excludedDirs,
		ProjectRoot:   projectRoot,
		BackupType:    "config",
	}

	metadataPath := filepath.Join(backupDir, "backup_metadata.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		_ = os.RemoveAll(backupDir)
		return "", fmt.Errorf("marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, defs.FilePerm); err != nil {
		_ = os.RemoveAll(backupDir)
		return "", fmt.Errorf("write metadata: %w", err)
	}

	return backupDir, nil
}

// BackupMetadata represents the structure of backup_metadata.json
type BackupMetadata struct {
	Timestamp     string   `json:"timestamp"`
	Description   string   `json:"description"`
	BackedUpItems []string `json:"backed_up_items"`
	ExcludedItems []string `json:"excluded_items"`
	ExcludedDirs  []string `json:"excluded_dirs"`
	ProjectRoot   string   `json:"project_root"`
	BackupType    string   `json:"backup_type"`
}

// cleanup_old_backups maintains a maximum of 'keepCount' backups, deleting the oldest ones.
// Returns the number of backups deleted.
func cleanup_old_backups(projectRoot string, keepCount int) int {
	backupDir := filepath.Join(projectRoot, defs.BackupsDir)

	// Check if backup directory exists
	info, err := os.Stat(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0 // No backups to clean up
		}
		// Return 0 on stat errors (ignore for cleanup)
		return 0
	}
	if !info.IsDir() {
		return 0
	}

	// Get all subdirectories in backup directory
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return 0
	}

	// Filter for directories matching YYYYMMDD_HHMMSS pattern
	// Pattern: 8 digits + underscore + 6 digits = 15 characters
	var backups []string
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) == 15 {
			// Check if it matches the timestamp pattern (digits + underscore + digits)
			parts := strings.SplitN(entry.Name(), "_", 2)
			if len(parts) == 2 {
				if len(parts[0]) == 8 && len(parts[1]) == 6 {
					backups = append(backups, entry.Name())
				}
			}
		}
	}

	// If we have fewer backups than keepCount, no cleanup needed
	if len(backups) <= keepCount {
		return 0
	}

	// Sort backups by name (timestamp) ascending (oldest first)
	sort.Strings(backups)

	// Delete backups exceeding the keep limit
	deletedCount := 0
	for _, backupName := range backups[keepCount:] {
		backupPath := filepath.Join(backupDir, backupName)
		if err := os.RemoveAll(backupPath); err != nil {
			// Log error but continue with other backups
			fmt.Fprintf(os.Stderr, "Warning: failed to delete backup %s: %v\n", backupName, err)
		} else {
			deletedCount++
		}
	}

	return deletedCount
}

// restoreMoaiConfig restores user settings from backup to new config files.
// It performs a deep YAML merge to preserve user settings while adopting new structure.
func restoreMoaiConfig(projectRoot, backupDir string) error {
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)

	// Walk through backup files
	err := filepath.Walk(backupDir, func(backupPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(backupDir, backupPath)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(configDir, relPath)

		// Read backup data
		backupData, err := os.ReadFile(backupPath)
		if err != nil {
			return err
		}

		// Check if target file exists
		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				// Target doesn't exist, just copy backup
				return os.WriteFile(targetPath, backupData, defs.FilePerm)
			}
			return err
		}

		// Both files exist, merge them
		targetData, err := os.ReadFile(targetPath)
		if err != nil {
			return err
		}

		// Perform YAML deep merge
		merged, err := mergeYAMLDeep(targetData, backupData)
		if err != nil {
			// If merge fails, backup the new file and restore old one
			_, _ = fmt.Fprintf(os.Stderr, "Warning: merge failed for %s, restoring backup\n", relPath)
			return os.WriteFile(targetPath, backupData, defs.FilePerm)
		}

		return os.WriteFile(targetPath, merged, defs.FilePerm)
	})

	return err
}

// mergeYAMLDeep performs a deep merge of two YAML documents.
// The newData takes precedence for structure, but values from oldData are preserved
// when the key exists in both.
func mergeYAMLDeep(newData, oldData []byte) ([]byte, error) {
	var newMap, oldMap map[string]interface{}

	if err := yaml.Unmarshal(newData, &newMap); err != nil {
		return nil, fmt.Errorf("unmarshal new YAML: %w", err)
	}
	if err := yaml.Unmarshal(oldData, &oldMap); err != nil {
		return nil, fmt.Errorf("unmarshal old YAML: %w", err)
	}

	// Deep merge old values into new structure
	merged := deepMergeMaps(newMap, oldMap)

	return yaml.Marshal(merged)
}

// deepMergeMaps recursively merges oldMap into newMap, preserving old values.
// System fields (like template_version) always use new values.
func deepMergeMaps(newMap, oldMap map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// System fields that should always use new values (not preserved from old config)
	systemFields := map[string]bool{
		"template_version": true,
	}

	// Copy all new values
	for k, v := range newMap {
		result[k] = v
	}

	// Merge old values, preserving when they exist
	for k, v := range oldMap {
		// Skip system fields - always use new value
		if systemFields[k] {
			continue
		}

		if newV, exists := newMap[k]; exists {
			// Both exist, check if they are maps
			newMapVal, newIsMap := newV.(map[string]interface{})
			oldMapVal, oldIsMap := v.(map[string]interface{})

			if newIsMap && oldIsMap {
				// Recursively merge nested maps
				result[k] = deepMergeMaps(newMapVal, oldMapVal)
			} else {
				// Keep old value (preserve user setting)
				result[k] = v
			}
		} else {
			// Only exists in old, add it
			result[k] = v
		}
	}

	return result
}

// mergeSettingsJSON performs smart merge for .claude/settings.json.
// Rules:
// - env: shallow merge (user variables preserved)
// - permissions.allow: array merge (deduplicated)
// - permissions.deny: template priority (security)
// - permissions.ask: template priority + user additions
// - hooks: template priority
// - outputStyle, spinnerTipsEnabled: user preserved
func mergeSettingsJSON(templatePath, existingPath string) error {
	// Load template
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read template settings.json: %w", err)
	}

	var tmplSettings map[string]interface{}
	if err := json.Unmarshal(templateData, &tmplSettings); err != nil {
		return fmt.Errorf("parse template settings.json: %w", err)
	}

	// Load existing for user settings
	userData := make(map[string]interface{})
	if existingData, err := os.ReadFile(existingPath); err == nil {
		if err := json.Unmarshal(existingData, &userData); err != nil {
			return fmt.Errorf("parse existing settings.json: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read existing settings.json: %w", err)
	}

	// Merge env (template priority for known keys, preserve user-added custom keys)
	templateEnv := getMap(tmplSettings, "env")
	userEnv := getMap(userData, "env")
	mergedEnv := make(map[string]interface{})

	// Copy template env first
	for k, v := range templateEnv {
		mergedEnv[k] = v
	}

	// Always refresh PATH from current terminal (not stale from template or previous init)
	mergedEnv["PATH"] = template.BuildSmartPATH()

	// Add user custom env keys not in template
	for k, v := range userEnv {
		if _, exists := templateEnv[k]; !exists {
			// User added a custom env key, preserve it
			mergedEnv[k] = v
		}
	}

	// Merge permissions.allow (deduplicated array merge)
	templatePerms := getMap(tmplSettings, "permissions")
	userPerms := getMap(userData, "permissions")

	templateAllow := getSlice(templatePerms, "allow")
	userAllow := getSlice(userPerms, "allow")
	mergedAllow := mergeStringSlices(templateAllow, userAllow)

	// permissions.deny: template priority (security)
	mergedDeny := getSlice(templatePerms, "deny")

	// permissions.ask: template priority + user additions
	templateAsk := getSlice(templatePerms, "ask")
	userAsk := getSlice(userPerms, "ask")
	mergedAsk := mergeStringSlices(templateAsk, userAsk)

	// Start with full template (include all fields from template)
	// Note: statusLine comes from template only (not user-preserved)
	merged := deepCopyMap(tmplSettings)

	// Override with merged values
	merged["env"] = mergedEnv
	merged["permissions"] = map[string]interface{}{
		"defaultMode": getValue(templatePerms, "defaultMode", "default"),
		"allow":       mergedAllow,
		"ask":         mergedAsk,
		"deny":        mergedDeny,
	}

	// Preserve user customizations for specific fields
	preserveFields := []string{"outputStyle", "spinnerTipsEnabled"}
	for _, field := range preserveFields {
		if val, exists := userData[field]; exists {
			merged[field] = val
		}
	}

	// Write merged settings
	jsonContent, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal merged settings: %w", err)
	}

	if err := os.WriteFile(existingPath, append(jsonContent, '\n'), defs.FilePerm); err != nil {
		return fmt.Errorf("write merged settings: %w", err)
	}

	return nil
}

// getMap safely gets a map value from a map
func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, exists := m[key]; exists {
		if mapVal, ok := val.(map[string]interface{}); ok {
			return mapVal
		}
	}
	return make(map[string]interface{})
}

// getSlice safely gets a string slice from a map
func getSlice(m map[string]interface{}, key string) []string {
	if val, exists := m[key]; exists {
		if sliceVal, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(sliceVal))
			for _, v := range sliceVal {
				if strVal, ok := v.(string); ok {
					result = append(result, strVal)
				}
			}
			return result
		}
	}
	return []string{}
}

// getValue safely gets a value from a map with default
func getValue(m map[string]interface{}, key, defaultVal string) string {
	if val, exists := m[key]; exists {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultVal
}

// mergeStringSlices merges two string slices, deduplicating
func mergeStringSlices(a, b []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(a)+len(b))

	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	sort.Strings(result)
	return result
}

// deepCopyMap creates a deep copy of a map
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = deepCopyMap(val)
		case []interface{}:
			copy := make([]interface{}, len(val))
			for i, item := range val {
				if subMap, ok := item.(map[string]interface{}); ok {
					copy[i] = deepCopyMap(subMap)
				} else {
					copy[i] = item
				}
			}
			result[k] = copy
		default:
			result[k] = v
		}
	}
	return result
}

// runInitWizard runs the configuration wizard for reconfiguring an existing project.
// Used by 'moai update -c/--config' to edit project settings.
func runInitWizard(cmd *cobra.Command, reconfigure bool) error {
	out := cmd.OutOrStdout()

	// Verify the project is initialized
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if _, err := os.Stat(filepath.Join(cwd, defs.MoAIDir)); os.IsNotExist(err) {
		_, _ = fmt.Fprintln(out, "Project not initialized. Run 'moai init' first.")
		return fmt.Errorf("project not initialized")
	}

	// Print banner and welcome message
	PrintBanner(version.GetVersion())
	if reconfigure {
		_, _ = fmt.Fprintln(out, "🔧 Project Reconfiguration Wizard")
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, "This wizard will help you update your project configuration.")
	} else {
		PrintWelcomeMessage()
	}

	// Run wizard with current directory as project root
	result, err := wizard.RunWithDefaults(cwd)
	if err != nil {
		if errors.Is(err, wizard.ErrCancelled) {
			_, _ = fmt.Fprintln(out, "Configuration cancelled.")
			return nil
		}
		return fmt.Errorf("wizard failed: %w", err)
	}

	// Apply configuration updates to .moai/config/sections/
	// This updates the YAML configuration files based on wizard results
	if err := applyWizardConfig(cwd, result); err != nil {
		return fmt.Errorf("apply configuration: %w", err)
	}

	_, _ = fmt.Fprintln(out, "✓ Configuration updated successfully.")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "  Language: %s\n", result.Locale)
	_, _ = fmt.Fprintf(out, "  Development mode: %s\n", result.DevelopmentMode)

	return nil
}

// applyWizardConfig applies wizard results to the project configuration files.
func applyWizardConfig(projectRoot string, result *wizard.WizardResult) error {
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)

	// Update language.yaml
	langPath := filepath.Join(sectionsDir, defs.LanguageYAML)
	langContent := fmt.Sprintf("language:\n  conversation_language: %s\n  conversation_language_name: %s\n",
		result.Locale, result.Locale)
	if err := os.WriteFile(langPath, []byte(langContent), defs.FilePerm); err != nil {
		return fmt.Errorf("write language.yaml: %w", err)
	}

	// Update quality.yaml if development mode changed
	if result.DevelopmentMode != "" {
		qualityPath := filepath.Join(sectionsDir, defs.QualityYAML)
		// Read existing content
		qualityData, err := os.ReadFile(qualityPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read quality.yaml: %w", err)
		}

		// Parse YAML
		var quality map[string]interface{}
		if len(qualityData) > 0 {
			if err := yaml.Unmarshal(qualityData, &quality); err != nil {
				return fmt.Errorf("parse quality.yaml: %w", err)
			}
		} else {
			quality = make(map[string]interface{})
		}

		// Update development_mode
		if constitution, ok := quality["constitution"].(map[string]interface{}); ok {
			constitution["development_mode"] = result.DevelopmentMode
		} else {
			quality["constitution"] = map[string]interface{}{
				"development_mode": result.DevelopmentMode,
			}
		}

		// Write back
		updatedData, err := yaml.Marshal(quality)
		if err != nil {
			return fmt.Errorf("marshal quality.yaml: %w", err)
		}
		if err := os.WriteFile(qualityPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write quality.yaml: %w", err)
		}
	}

	// Update workflow.yaml with Agent Teams settings
	if result.AgentTeamsMode != "" {
		workflowPath := filepath.Join(sectionsDir, defs.WorkflowYAML)
		// Read existing content
		workflowData, err := os.ReadFile(workflowPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read workflow.yaml: %w", err)
		}

		// Parse YAML
		var workflow map[string]interface{}
		if len(workflowData) > 0 {
			if err := yaml.Unmarshal(workflowData, &workflow); err != nil {
				return fmt.Errorf("parse workflow.yaml: %w", err)
			}
		} else {
			workflow = make(map[string]interface{})
		}

		// Ensure workflow and workflow.team exist
		workflowVal, ok := workflow["workflow"].(map[string]interface{})
		if !ok {
			workflowVal = make(map[string]interface{})
			workflow["workflow"] = workflowVal
		}

		// Set execution_mode
		workflowVal["execution_mode"] = result.AgentTeamsMode

		// Handle team configuration
		var teamConfig map[string]interface{}
		if existingTeam, ok := workflowVal["team"].(map[string]interface{}); ok {
			teamConfig = existingTeam
		} else {
			teamConfig = make(map[string]interface{})
		}

		// Set enabled flag based on AgentTeamsMode
		teamConfig["enabled"] = (result.AgentTeamsMode == "team")

		// Set max_teammates if provided (valid values: 2-5)
		if result.MaxTeammates != "" {
			// Validate max_teammates is between 2 and 5
			if val, err := strconv.Atoi(result.MaxTeammates); err == nil && val >= 2 && val <= 5 {
				teamConfig["max_teammates"] = val
			}
		}

		// Set default_model if provided
		if result.DefaultModel != "" {
			// Validate default_model is one of: haiku, sonnet, opus
			if result.DefaultModel == "haiku" || result.DefaultModel == "sonnet" || result.DefaultModel == "opus" {
				teamConfig["default_model"] = result.DefaultModel
			}
		}

		workflowVal["team"] = teamConfig
		workflow["workflow"] = workflowVal

		// Write back
		updatedData, err := yaml.Marshal(workflow)
		if err != nil {
			return fmt.Errorf("marshal workflow.yaml: %w", err)
		}
		if err := os.WriteFile(workflowPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write workflow.yaml: %w", err)
		}
	}

	// Update user.yaml with GitHub username and token
	if result.GitHubUsername != "" || result.GitHubToken != "" {
		userPath := filepath.Join(sectionsDir, defs.UserYAML)
		// Read existing content
		userData, err := os.ReadFile(userPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read user.yaml: %w", err)
		}

		// Parse YAML
		var user map[string]interface{}
		if len(userData) > 0 {
			if err := yaml.Unmarshal(userData, &user); err != nil {
				return fmt.Errorf("parse user.yaml: %w", err)
			}
		} else {
			user = make(map[string]interface{})
		}

		// Ensure user.user exists
		var userConfig map[string]interface{}
		if existingUser, ok := user["user"].(map[string]interface{}); ok {
			userConfig = existingUser
		} else {
			userConfig = make(map[string]interface{})
		}

		// Set github_username if provided
		if result.GitHubUsername != "" {
			userConfig["github_username"] = result.GitHubUsername
		}

		// Set github_token if provided
		if result.GitHubToken != "" {
			userConfig["github_token"] = result.GitHubToken
		}

		user["user"] = userConfig

		// Write back
		updatedData, err := yaml.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user.yaml: %w", err)
		}
		if err := os.WriteFile(userPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write user.yaml: %w", err)
		}
	}

	return nil
}

// ensureGlobalSettingsEnv ensures only the SessionEnd hook for moai-rank is set in ~/.claude/settings.json.
// All other settings (env, permissions, teammateMode) are managed at the project level in .claude/settings.json.
// This minimizes modifications to the user's global settings file.
func ensureGlobalSettingsEnv() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	globalSettingsPath := filepath.Join(homeDir, defs.ClaudeDir, defs.SettingsJSON)

	// SessionEnd hook for moai-rank submission (the only global setting we manage)
	sessionEndHookCommand := buildSessionEndHookCommand()

	// Read existing global settings
	var existingSettings map[string]interface{}
	if data, err := os.ReadFile(globalSettingsPath); err == nil {
		if err := json.Unmarshal(data, &existingSettings); err != nil {
			return fmt.Errorf("parse existing global settings: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read global settings: %w", err)
	} else {
		existingSettings = make(map[string]interface{})
	}

	// Check if SessionEnd hook needs to be added/updated
	needsUpdate := ensureSessionEndHook(existingSettings, sessionEndHookCommand)

	// Clean up moai-managed settings that have been migrated to project level.
	// Preserve any user-added custom env keys but remove moai-specific ones.
	if envVal, exists := existingSettings["env"]; exists {
		if envMap, ok := envVal.(map[string]interface{}); ok {
			moaiKeys := []string{"PATH", "ENABLE_TOOL_SEARCH", "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"}
			for _, key := range moaiKeys {
				if _, exists := envMap[key]; exists {
					delete(envMap, key)
					needsUpdate = true
				}
			}
			// If env is now empty, remove it entirely
			if len(envMap) == 0 {
				delete(existingSettings, "env")
			}
		}
	}

	// Clean up moai-managed permissions if they only contain Task:*
	if permVal, exists := existingSettings["permissions"]; exists {
		if permMap, ok := permVal.(map[string]interface{}); ok {
			if allowVal, exists := permMap["allow"]; exists {
				if allowArr, ok := allowVal.([]interface{}); ok {
					if len(allowArr) == 1 && allowArr[0] == "Task:*" {
						delete(existingSettings, "permissions")
						needsUpdate = true
					}
				}
			}
		}
	}

	// Clean up moai-managed teammateMode
	if mode, exists := existingSettings["teammateMode"]; exists {
		if mode == "auto" {
			delete(existingSettings, "teammateMode")
			needsUpdate = true
		}
	}

	if !needsUpdate {
		return nil
	}

	// Add SessionEnd hook
	addSessionEndHook(existingSettings, sessionEndHookCommand)

	// Write back
	jsonContent, err := json.MarshalIndent(existingSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal global settings: %w", err)
	}

	if err := os.WriteFile(globalSettingsPath, append(jsonContent, '\n'), defs.FilePerm); err != nil {
		return fmt.Errorf("write global settings: %w", err)
	}

	return nil
}

// buildSessionEndHookCommand builds the SessionEnd hook command for moai-rank submission.
// Uses shell script wrapper with $HOME environment variable for global hook installation.
func buildSessionEndHookCommand() string {
	// The hook wrapper script is installed at ~/.claude/hooks/moai/handle-session-end.sh
	// This wrapper calls: moai hook session-end
	// Note: For global settings, use $HOME instead of $CLAUDE_PROJECT_DIR
	return "\"$HOME/.claude/hooks/moai/handle-session-end.sh\""
}

// ensureSessionEndHook checks if the SessionEnd hook needs to be added or updated.
// Returns true if an update is needed.
func ensureSessionEndHook(settings map[string]interface{}, hookCommand string) bool {
	existingHooks, hasHooks := settings["hooks"]
	if !hasHooks {
		return true // Need to add hooks
	}

	hooksMap, ok := existingHooks.(map[string]interface{})
	if !ok {
		return true // Invalid hooks structure, need to fix
	}

	sessionEndHooks, hasSessionEnd := hooksMap["SessionEnd"]
	if !hasSessionEnd {
		return true // Need to add SessionEnd hook
	}

	// Check if our hook already exists
	sessionEndList, ok := sessionEndHooks.([]interface{})
	if !ok {
		return true // Invalid structure, need to fix
	}

	for _, hookGroup := range sessionEndList {
		groupMap, ok := hookGroup.(map[string]interface{})
		if !ok {
			continue
		}

		hooksList, ok := groupMap["hooks"].([]interface{})
		if !ok {
			continue
		}

		for _, hookEntry := range hooksList {
			entryMap, ok := hookEntry.(map[string]interface{})
			if !ok {
				continue
			}

			if command, ok := entryMap["command"].(string); ok {
				// Check if this is our moai-rank hook
				if strings.Contains(command, "handle-session-end.sh") || strings.Contains(command, "session_end__rank_submit") {
					// Hook exists, check if it needs updating to use shell script
					if strings.Contains(command, "python") || strings.Contains(command, "uv run") {
						return true // Needs update from Python to shell
					}
					return false // Hook exists and is up to date
				}
			}
		}
	}

	return true // Hook not found, need to add
}

// addSessionEndHook adds the SessionEnd hook to the settings map.
func addSessionEndHook(settings map[string]interface{}, hookCommand string) {
	// Get or create hooks map
	var hooksMap map[string]interface{}
	if existingHooks, hasHooks := settings["hooks"]; hasHooks {
		if hm, ok := existingHooks.(map[string]interface{}); ok {
			hooksMap = hm
		} else {
			hooksMap = make(map[string]interface{})
		}
	} else {
		hooksMap = make(map[string]interface{})
	}

	// Remove any existing Python-based moai-rank hooks
	cleanSessionEndHooks(hooksMap)

	// Add new shell-based SessionEnd hook with timeout
	hooksMap["SessionEnd"] = []interface{}{
		map[string]interface{}{
			"hooks": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": hookCommand,
					"timeout": 5,
				},
			},
		},
	}

	settings["hooks"] = hooksMap
}

// cleanSessionEndHooks removes any existing Python-based moai-rank hooks from SessionEnd.
func cleanSessionEndHooks(hooksMap map[string]interface{}) {
	sessionEndHooks, hasSessionEnd := hooksMap["SessionEnd"]
	if !hasSessionEnd {
		return
	}

	sessionEndList, ok := sessionEndHooks.([]interface{})
	if !ok {
		return
	}

	var cleanedHooks []interface{}
	for _, hookGroup := range sessionEndList {
		groupMap, ok := hookGroup.(map[string]interface{})
		if !ok {
			cleanedHooks = append(cleanedHooks, hookGroup)
			continue
		}

		hooksList, ok := groupMap["hooks"].([]interface{})
		if !ok {
			cleanedHooks = append(cleanedHooks, hookGroup)
			continue
		}

		var cleanedGroupHooks []interface{}
		for _, hookEntry := range hooksList {
			entryMap, ok := hookEntry.(map[string]interface{})
			if !ok {
				cleanedGroupHooks = append(cleanedGroupHooks, hookEntry)
				continue
			}

			command, ok := entryMap["command"].(string)
			if !ok {
				cleanedGroupHooks = append(cleanedGroupHooks, hookEntry)
				continue
			}

			// Remove Python-based moai-rank hooks
			if strings.Contains(command, "session_end__rank_submit") ||
				(strings.Contains(command, "moai-rank") && strings.Contains(command, "python")) ||
				strings.Contains(command, "uv run python") && strings.Contains(command, "rank_submit") {
				// Skip this hook (remove it)
				continue
			}

			cleanedGroupHooks = append(cleanedGroupHooks, hookEntry)
		}

		if len(cleanedGroupHooks) > 0 {
			groupMap["hooks"] = cleanedGroupHooks
			cleanedHooks = append(cleanedHooks, groupMap)
		}
	}

	if len(cleanedHooks) > 0 {
		hooksMap["SessionEnd"] = cleanedHooks
	} else {
		delete(hooksMap, "SessionEnd")
	}
}

// execCommand executes a command and returns its output.
func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// detectGoBinPathForUpdate detects the Go binary installation path for template rendering.
// Returns the path where Go binaries are installed (e.g., "/Users/goos/go/bin").
func detectGoBinPathForUpdate(homeDir string) string {
	// Try GOBIN first (explicit override)
	if output, err := execCommand("go", "env", "GOBIN"); err == nil {
		if goBin := strings.TrimSpace(output); goBin != "" {
			return goBin
		}
	}

	// Try GOPATH/bin (user's Go workspace)
	if output, err := execCommand("go", "env", "GOPATH"); err == nil {
		if goPath := strings.TrimSpace(output); goPath != "" {
			return filepath.Join(goPath, "bin")
		}
	}

	// Fallback to default ~/go/bin
	if homeDir != "" {
		return filepath.Join(homeDir, "go", "bin")
	}

	// Last resort: common Go install location
	return "/usr/local/go/bin"
}
