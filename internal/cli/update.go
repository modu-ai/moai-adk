package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
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
	Long:  "Check for binary updates, install if available, then synchronize embedded templates with the project.",
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Bool("check", false, "Check if a newer binary version is available (informational)")
	updateCmd.Flags().Bool("shell-env", false, "Configure shell environment variables for Claude Code")
	updateCmd.Flags().BoolP("config", "c", false, "Edit project configuration (same as init wizard)")
	updateCmd.Flags().Bool("force", false, "Skip backup and force the update")
	updateCmd.Flags().Bool("yes", false, "Auto-confirm all prompts (CI/CD mode)")
	updateCmd.Flags().Bool("templates-only", false, "Skip binary update, sync templates only")
	updateCmd.Flags().Bool("binary", false, "Update binary only, skip template sync")
}

// runUpdate checks for binary updates first, then synchronizes embedded
// templates with the project directory. If a newer binary is installed,
// the process re-execs itself so the latest templates are used.
//
// Flags:
//
//	-c, --config: Edit project configuration (same as init wizard)
//	--check: Check if a newer binary version is available (informational)
//	--force: Skip backup and force the update
//	--shell-env: Configure shell environment variables
//	--yes: Auto-confirm all prompts (CI/CD mode)
//	--templates-only: Skip binary update, sync templates only
//	--binary: Update binary only, skip template sync
func runUpdate(cmd *cobra.Command, _ []string) error {
	checkOnly := getBoolFlag(cmd, "check")
	shellEnv := getBoolFlag(cmd, "shell-env")
	editConfig := getBoolFlag(cmd, "config")
	binaryOnly := getBoolFlag(cmd, "binary")
	templatesOnly := getBoolFlag(cmd, "templates-only")
	out := cmd.OutOrStdout()

	// Validate mutually exclusive flags
	if binaryOnly && templatesOnly {
		return fmt.Errorf("--binary and --templates-only are mutually exclusive")
	}

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

	// Step 1: Binary update (unless skipped)
	if !shouldSkipBinaryUpdate(cmd) {
		updated, err := runBinaryUpdateStep(cmd)
		if err != nil {
			// Binary update failure is never fatal; warn and continue
			_, _ = fmt.Fprintf(out, "Warning: binary update check failed: %v\n", err)
		}
		if updated {
			if binaryOnly {
				// --binary mode: skip re-exec and template sync
				_, _ = fmt.Fprintln(out, "Binary updated successfully (template sync skipped).")
				return nil
			}
			// New binary installed; re-exec so the latest templates are used
			if err := reexecNewBinary(); err != nil {
				_, _ = fmt.Fprintf(out, "Warning: failed to re-exec new binary: %v\n", err)
				// Fall through to template sync with the current binary
			}
			// reexecNewBinary replaces the process on success, so we only
			// reach here if it failed.
		} else if binaryOnly {
			_, _ = fmt.Fprintln(out, "Already up to date (no newer binary available).")
			return nil
		}
	}

	// Step 2: Template sync (skipped when --binary is set)
	if binaryOnly {
		_, _ = fmt.Fprintln(out, "Binary update skipped (dev build). Template sync skipped (--binary).")
		return nil
	}
	return runTemplateSyncWithProgress(cmd)
}

// shouldSkipBinaryUpdate returns true when the binary update step should
// be skipped. This happens in three cases:
//  1. The --templates-only flag is set (update command only).
//  2. The MOAI_SKIP_BINARY_UPDATE=1 environment variable is set (used by
//     reexecNewBinary to prevent infinite re-exec loops).
//  3. The current binary is a dev build (version contains "dirty", "dev",
//     or "none"), where self-update is meaningless.
func shouldSkipBinaryUpdate(cmd *cobra.Command) bool {
	// Flag check (only the update command registers this flag)
	if f := cmd.Flags().Lookup("templates-only"); f != nil && f.Value.String() == "true" {
		return true
	}

	// Environment variable guard (set by reexecNewBinary)
	if os.Getenv("MOAI_SKIP_BINARY_UPDATE") == "1" {
		return true
	}

	// Dev build detection (reuse pattern from buildAutoUpdateFunc in deps.go)
	v := version.GetVersion()
	if strings.Contains(v, "dirty") || v == "dev" || strings.Contains(v, "none") {
		return true
	}

	return false
}

// runBinaryUpdateStep checks whether a newer moai binary is available and,
// if so, downloads and installs it. The caller should re-exec the process
// when updated is true.
//
// Errors are non-fatal by design: the caller should log the error and
// continue with the original operation (template sync or init).
func runBinaryUpdateStep(cmd *cobra.Command) (updated bool, err error) {
	out := cmd.OutOrStdout()

	// Lazily initialise update dependencies
	if deps != nil {
		if initErr := deps.EnsureUpdate(); initErr != nil {
			return false, fmt.Errorf("initialize update system: %w", initErr)
		}
	}

	if deps == nil || deps.UpdateChecker == nil {
		return false, nil
	}

	currentVersion := version.GetVersion()

	// Check for available update
	available, info, err := deps.UpdateChecker.IsUpdateAvailable(currentVersion)
	if err != nil {
		return false, fmt.Errorf("check for update: %w", err)
	}
	if !available {
		return false, nil
	}

	_, _ = fmt.Fprintf(out, "New version available: %s (current: %s)\n", info.Version, currentVersion)
	_, _ = fmt.Fprintln(out, "Installing update...")

	if deps.UpdateOrch == nil {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 120*time.Second)
	defer cancel()

	result, err := deps.UpdateOrch.Update(ctx)
	if err != nil {
		return false, fmt.Errorf("install update: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Updated: %s -> %s\n", result.PreviousVersion, result.NewVersion)
	return true, nil
}

// reexecNewBinary replaces the current process with the newly installed
// moai binary, preserving the original command-line arguments. It sets
// MOAI_SKIP_BINARY_UPDATE=1 to prevent the re-execed process from
// attempting another binary update.
//
// On Unix this uses syscall.Exec (the process is replaced in-place).
// On Windows syscall.Exec is not available, so we spawn a child process
// and exit the parent.
func reexecNewBinary() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	// Prevent re-exec loop
	if err := os.Setenv("MOAI_SKIP_BINARY_UPDATE", "1"); err != nil {
		return fmt.Errorf("set MOAI_SKIP_BINARY_UPDATE: %w", err)
	}

	if runtime.GOOS == "windows" {
		// Windows: spawn child and exit parent
		child := exec.Command(exe, os.Args[1:]...)
		child.Stdin = os.Stdin
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if err := child.Run(); err != nil {
			return fmt.Errorf("re-exec on windows: %w", err)
		}
		os.Exit(0)
	}

	// Unix: replace process via execve(2)
	return syscall.Exec(exe, os.Args, os.Environ())
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
			name:    "Clean Managed Paths",
			message: "Removing old MoAI-managed files",
			execute: func() error {
				return cleanMoaiManagedPaths(projectRoot, out)
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
					template.WithSmartPATH(template.BuildSmartPATH()),
					template.WithPlatform(runtime.GOOS),
					template.WithVersion(version.GetVersion()),
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

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "💡 To reconfigure your project settings, run:")
	_, _ = fmt.Fprintln(out, "   moai update -c")

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

	return runTemplateSyncWithReporter(cmd, consoleReporter, true)
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
		// Strip .tmpl suffix first - display and filter using rendered target path
		displayPath := tmpl
		if strings.HasSuffix(tmpl, ".tmpl") {
			displayPath = strings.TrimSuffix(tmpl, ".tmpl")
		}

		// Filter out MoAI-managed files - they are automatically installed
		if isMoaiManaged(displayPath) {
			continue
		}

		// Use rendered target path for existence check
		targetPath := filepath.Join(projectRoot, displayPath)

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
		case "skills", "rules", "agents", "commands", "output-styles", "hooks":
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
// Creates a timestamped backup under .moai-backups/YYYYMMDD_HHMMSS/ including
// all files (config.yaml, sections/*.yaml, etc.) for full restore capability.
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

	// All config files are backed up (including sections/) for full restore.
	// The clean step will delete everything, and the restore step will
	// merge backed-up values back into the freshly deployed templates.
	excludedDirs := []string{}

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

// cleanMoaiManagedPaths removes MoAI-managed directories and files before template
// deployment. This ensures stale files are cleaned up during version upgrades.
// The .moai/config/ directory is deleted entirely (backup was done by the Backup step).
// Paths that do not exist are silently skipped.
func cleanMoaiManagedPaths(projectRoot string, out io.Writer) error {
	type cleanTarget struct {
		// displayPath is shown in progress messages (e.g., ".claude/settings.json")
		displayPath string
		// fullPath is the absolute filesystem path to delete
		fullPath string
		// isGlob indicates the target uses filepath.Glob matching
		isGlob bool
	}

	targets := []cleanTarget{
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.SettingsJSON),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.SettingsJSON),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.CommandsMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.CommandsMoaiSubdir),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.AgentsMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.AgentsMoaiSubdir),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
			isGlob:      true,
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.RulesMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.RulesMoaiSubdir),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
		},
		{
			displayPath: filepath.Join(defs.ClaudeDir, defs.HooksMoaiSubdir),
			fullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.HooksMoaiSubdir),
		},
	}

	// Process standard targets (files and directories)
	for _, t := range targets {
		_, _ = fmt.Fprintf(out, "  ○ Removing %s...", t.displayPath)

		if t.isGlob {
			matches, err := filepath.Glob(t.fullPath)
			if err != nil {
				_, _ = fmt.Fprintf(out, "\r  ✗ Failed to glob %s: %v\n", t.displayPath, err)
				return fmt.Errorf("glob %s: %w", t.displayPath, err)
			}
			for _, match := range matches {
				if err := os.RemoveAll(match); err != nil {
					_, _ = fmt.Fprintf(out, "\r  ✗ Failed to remove %s: %v\n", t.displayPath, err)
					return fmt.Errorf("remove %s: %w", match, err)
				}
			}
			_, _ = fmt.Fprintf(out, "\r  ✓ Removed %s\n", t.displayPath)
			continue
		}

		if _, err := os.Stat(t.fullPath); err != nil {
			if os.IsNotExist(err) {
				_, _ = fmt.Fprintf(out, "\r  - Skipped %s (not found)\n", t.displayPath)
				continue
			}
			_, _ = fmt.Fprintf(out, "\r  ✗ Failed to stat %s: %v\n", t.displayPath, err)
			return fmt.Errorf("stat %s: %w", t.displayPath, err)
		}

		if err := os.RemoveAll(t.fullPath); err != nil {
			_, _ = fmt.Fprintf(out, "\r  ✗ Failed to remove %s: %v\n", t.displayPath, err)
			return fmt.Errorf("remove %s: %w", t.displayPath, err)
		}
		_, _ = fmt.Fprintf(out, "\r  ✓ Removed %s\n", t.displayPath)
	}

	// Clean .moai/config/ entirely - backup was already done by the Backup step.
	// For v1.x -> v2.x: old config is incompatible, fresh install needed.
	// For v2.x -> v2.x: backup includes sections/, restore will merge values back.
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
	configDisplayPath := filepath.Join(defs.MoAIDir, defs.ConfigSubdir)
	_, _ = fmt.Fprintf(out, "  ○ Removing %s...", configDisplayPath)

	if err := os.RemoveAll(configDir); err != nil {
		if !os.IsNotExist(err) {
			_, _ = fmt.Fprintf(out, "\r  ✗ Failed to remove %s: %v\n", configDisplayPath, err)
			return fmt.Errorf("remove %s: %w", configDisplayPath, err)
		}
	}
	_, _ = fmt.Fprintf(out, "\r  ✓ Removed %s\n", configDisplayPath)

	return nil
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
				// Target doesn't exist, ensure parent directory exists
				if err := os.MkdirAll(filepath.Dir(targetPath), defs.DirPerm); err != nil {
					return fmt.Errorf("create parent directory for %s: %w", relPath, err)
				}
				// Copy backup to target
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

	// Development mode is no longer configured via wizard.
	// It defaults to "hybrid" and is auto-configured by /moai project workflow.

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

		// Set max_teammates if provided (valid values: 2-10)
		if result.MaxTeammates != "" {
			// Validate max_teammates is between 2 and 10
			if val, err := strconv.Atoi(result.MaxTeammates); err == nil && val >= 2 && val <= 10 {
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

// ensureGlobalSettingsEnv cleans up moai-managed settings from ~/.claude/settings.json.
// All settings (env, permissions, teammateMode, hooks) are managed at the project level.
// The global hooks directory (~/.claude/hooks/moai/) is also removed since hooks
// are only deployed to project-level directories via moai init.
func ensureGlobalSettingsEnv() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	// Remove global hooks/moai directory if it exists.
	// Hooks are project-level only; the global directory causes "No such file or directory"
	// errors in non-initialized projects that reference $CLAUDE_PROJECT_DIR paths.
	globalHooksDir := filepath.Join(homeDir, defs.ClaudeDir, "hooks", "moai")
	if _, err := os.Stat(globalHooksDir); err == nil {
		_ = os.RemoveAll(globalHooksDir)
	}

	globalSettingsPath := filepath.Join(homeDir, defs.ClaudeDir, defs.SettingsJSON)

	// Read existing global settings
	var existingSettings map[string]interface{}
	if data, err := os.ReadFile(globalSettingsPath); err == nil {
		if err := json.Unmarshal(data, &existingSettings); err != nil {
			return fmt.Errorf("parse existing global settings: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read global settings: %w", err)
	} else {
		// No global settings file, nothing to clean up
		return nil
	}

	needsUpdate := false

	// Clean up legacy hooks including orphaned scripts and deprecated Python hooks
	needsUpdate = cleanLegacyHooks(existingSettings) || needsUpdate

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

// cleanLegacyHooks removes legacy hook patterns from global settings.
// This includes orphaned scripts that were never deployed and deprecated Python-based hooks.
// Returns true if any cleanup was performed.
func cleanLegacyHooks(settings map[string]interface{}) bool {
	// List of legacy hook patterns to remove.
	// All moai handle-*.sh hooks belong in project-level settings, not global.
	legacyPatterns := []string{
		"handle-session-end.sh",
		"handle-session-start.sh",
		"handle-stop.sh",
		"handle-pre-tool.sh",
		"handle-post-tool.sh",
		"handle-agent-hook.sh",
		"handle-compact.sh",
		"session_end__rank_submit",
		"post_tool__code_formatter.py",
		"post_tool__linter.py",
		"post_tool__ast_grep_scan.py",
	}

	hooksMap, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		return false
	}

	modified := false
	for hookType, hookListInterface := range hooksMap {
		hookList, ok := hookListInterface.([]interface{})
		if !ok {
			continue
		}

		var cleanedHooks []interface{}
		for _, hookGroup := range hookList {
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

				// Check if command contains any legacy pattern
				shouldRemove := false
				for _, pattern := range legacyPatterns {
					if strings.Contains(command, pattern) {
						shouldRemove = true
						break
					}
				}

				if shouldRemove {
					modified = true
				} else {
					cleanedGroupHooks = append(cleanedGroupHooks, hookEntry)
				}
			}

			if len(cleanedGroupHooks) > 0 {
				groupMap["hooks"] = cleanedGroupHooks
				cleanedHooks = append(cleanedHooks, groupMap)
			} else {
				modified = true
			}
		}

		if len(cleanedHooks) > 0 {
			hooksMap[hookType] = cleanedHooks
		} else {
			delete(hooksMap, hookType)
			modified = true
		}
	}

	if modified && len(hooksMap) == 0 {
		delete(settings, "hooks")
	}

	return modified
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
// Returns the path where Go binaries are installed (e.g., "/home/user/go/bin").
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
