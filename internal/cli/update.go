package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/profile"
	"github.com/modu-ai/moai-adk/internal/shell"
	"github.com/modu-ai/moai-adk/internal/statusline"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	// maxConfigSize is the maximum allowed size for a config YAML file to prevent DoS
	maxConfigSize = 10 * 1024 * 1024 // 10MB
)

// CLI output styles for consistent MoAI-themed terminal output.
var (
	cliSuccess = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"})
	cliWarn    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#F59E0B"})
	cliError   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"})
	cliMuted   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"})
	cliPrimary = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C45A3C", Dark: "#DA7756"})
	cliBorder  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"})
)

func symSuccess() string  { return cliSuccess.Render("\u2713") }
func symError() string    { return cliError.Render("\u2717") }
func symWarning() string  { return cliWarn.Render("!") }
func symProgress() string { return cliMuted.Render("\u25CB") }

// fileBackup holds a file path and its backed-up content for merging.
type fileBackup struct {
	path string
	data []byte
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Sync MoAI-ADK project templates to the latest version",
	GroupID: "project",
	Long:    "Check for binary updates, install if available, then synchronize embedded templates with the project.",
	RunE:    runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Bool("check", false, "Check if a newer binary version is available (informational)")
	updateCmd.Flags().Bool("shell-env", false, "Configure shell environment variables for Claude Code")
	updateCmd.Flags().BoolP("config", "c", false, "Edit project configuration (same as init wizard)")
	updateCmd.Flags().Bool("force", false, "Force update even if version matches (still performs backup and merge)")
	updateCmd.Flags().Bool("yes", false, "Auto-confirm all prompts (CI/CD mode)")
	updateCmd.Flags().Bool("templates-only", false, "Skip binary update, sync templates only")
	updateCmd.Flags().Bool("binary", false, "Update binary only, skip template sync")
}

// @MX:ANCHOR: [AUTO] runUpdate orchestrates binary update and template synchronization
// @MX:REASON: [AUTO] fan_in=3, called from update.go init(), coverage_test.go, remaining_coverage_test.go
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

	// Auto-prompt profile setup if no profile exists yet
	nonInteractive := getBoolFlag(cmd, "yes")
	if !nonInteractive && isatty.IsTerminal(os.Stdin.Fd()) {
		profileName := profile.GetCurrentName()
		if !profile.IsSetup(profileName) {
			var wantSetup bool
			confirm := huh.NewConfirm().
				Title("No profile found. Set up profile preferences now?").
				Description("Configure your name, language, and model preferences.").
				Value(&wantSetup)
			if err := confirm.Run(); err == nil && wantSetup {
				if err := runProfileSetup(cmd, nil); err != nil {
					_, _ = fmt.Fprintf(out, "Warning: profile setup failed: %v\n", err)
				}
			}
		}
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
	if err := runTemplateSyncWithProgress(cmd); err != nil {
		return err
	}

	// Sync profile preferences to project config (after template deployment)
	profileName := profile.GetCurrentName()
	prefs, err := profile.ReadPreferences(profileName)
	if err != nil {
		_, _ = fmt.Fprintf(out, "Warning: failed to read profile preferences: %v\n", err)
	} else {
		if err := profile.SyncToProjectConfig(".", prefs); err != nil {
			_, _ = fmt.Fprintf(out, "Warning: failed to sync profile to project config: %v\n", err)
		}
	}

	return nil
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
	if err == nil && packageVersion == projectVersion && !forceBackup {
		if reporter != nil {
			reporter.StepComplete("Already up-to-date")
		}
		_, _ = fmt.Fprintf(out, "\n%s Template version up-to-date. Skipping sync.\n", symSuccess())
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
				// Always backup before update (even with --force)
				// --force only skips version check, not backup/merge
				_, _ = fmt.Fprintf(out, "  %s Backing up .moai/config...", symProgress())
				configBackupPath, backupErr := backupMoaiConfig(projectRoot)
				if backupErr != nil {
					_, _ = fmt.Fprintf(out, "\r  %s Backup failed: %v\n", symError(), backupErr)
					return backupErr
				}
				if configBackupPath != "" {
					_, _ = fmt.Fprintf(out, "\r  %s .moai/config backed up\n", symSuccess())
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
				_, _ = fmt.Fprintf(out, "  %s Deploying templates...", symProgress())

				// Build TemplateContext with detected paths for template rendering
				homeDir, _ := userHomeDir()
				goBinPath := detectGoBinPathForUpdate(homeDir)
				tmplCtx := template.NewTemplateContext(
					template.WithGoBinPath(goBinPath),
					template.WithHomeDir(homeDir),
					template.WithSmartPATH(template.BuildSmartPATH()),
					template.WithPlatform(runtime.GOOS),
					template.WithVersion(version.GetVersion()),
				)

				if deployErr := deployer.Deploy(ctx, projectRoot, mgr, tmplCtx); deployErr != nil {
					_, _ = fmt.Fprintf(out, "\r  %s Deployment failed: %v\n", symError(), deployErr)
					return fmt.Errorf("deploy templates: %w", deployErr)
				}
				_, _ = fmt.Fprintf(out, "\r  %s Templates deployed\n", symSuccess())
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
	// Backup of user's .gitignore content for EntryMerge after deploy
	var gitignoreBackup []byte
	// Backups of mergeable files for 3-way merge after deploy
	var mergeableBackups []fileBackup

	// collectMergeableFiles returns a list of files that should be merged
	// using the 3-way merge engine during update.
	// Note: .moai/config/sections/*.yaml files are already handled by
	// restoreMoaiConfig with 3-way merge, so they are excluded here.
	collectMergeableFiles := func(projectRoot string) []string {
		// Fixed mergeable files at project root that are NOT handled by restoreMoaiConfig
		return []string{
			".mcp.json",
			".claude/settings.json",
			".moai/status_line.sh",
		}
	}

	// Execute each step with progress reporting
	for i, step := range steps {
		if reporter != nil {
			reporter.StepStart(step.name, step.message)
		}

		// Special handling for backup/restore steps; default executes normally
		switch step.name {
		case "Backup":
			_, _ = fmt.Fprintf(out, "  %s Backing up .moai/config...", symProgress())
			var backupErr error
			configBackupPath, backupErr = backupMoaiConfig(projectRoot)
			if backupErr != nil {
				_, _ = fmt.Fprintf(out, "\r  %s Backup failed: %v\n", symError(), backupErr)
				if reporter != nil {
					reporter.StepError(backupErr)
				}
				return backupErr
			}
			if configBackupPath != "" {
				_, _ = fmt.Fprintf(out, "\r  %s .moai/config backed up\n", symSuccess())
			} else {
				_, _ = fmt.Fprintln(out, "\r  - No config to backup")
			}
			// Also backup .gitignore for EntryMerge after deploy
			gitignorePath := filepath.Join(projectRoot, ".gitignore")
			if data, readErr := os.ReadFile(gitignorePath); readErr == nil {
				gitignoreBackup = data
			}
			// Backup mergeable files for 3-way merge after deploy
			mergeableFiles := collectMergeableFiles(projectRoot)
			for _, mf := range mergeableFiles {
				mfPath := filepath.Join(projectRoot, mf)
				if data, readErr := os.ReadFile(mfPath); readErr == nil {
					mergeableBackups = append(mergeableBackups, fileBackup{path: mf, data: data})
				}
			}
			if reporter != nil {
				reporter.StepComplete("Configuration backed up")
			}
		case "Restore Settings":
			// Handle restore step with captured backup path
			if configBackupPath != "" {
				if reporter != nil {
					reporter.StepStart("Restore Settings", "Restoring user settings")
				}
				_, _ = fmt.Fprintf(out, "  %s Restoring user settings...", symProgress())
				if restoreErr := restoreMoaiConfig(projectRoot, configBackupPath); restoreErr != nil {
					_, _ = fmt.Fprintf(out, "\r  %s Restore failed: %v\n", symError(), restoreErr)
					if reporter != nil {
						reporter.StepError(restoreErr)
					}
					return restoreErr
				}
				_, _ = fmt.Fprintf(out, "\r  %s User settings restored\n", symSuccess())
				deletedCount := cleanup_old_backups(projectRoot, 5)
				if deletedCount > 0 {
					_, _ = fmt.Fprintf(out, "  %s Cleaned up %d old backup(s)\n", symSuccess(), deletedCount)
				}
				if reporter != nil {
					reporter.StepComplete("Settings restored")
				}
			}
			// Merge .gitignore: preserve user-added patterns via EntryMerge
			if len(gitignoreBackup) > 0 {
				gitignorePath := filepath.Join(projectRoot, ".gitignore")
				if mergeErr := mergeGitignoreFile(gitignorePath, gitignoreBackup); mergeErr != nil {
					_, _ = fmt.Fprintf(out, "  %s .gitignore merge warning: %v\n", symWarning(), mergeErr)
				} else {
					_, _ = fmt.Fprintf(out, "  %s .gitignore user patterns preserved\n", symSuccess())
				}
			}
			// Merge user-customized files using 3-way merge engine
			if len(mergeableBackups) > 0 {
				if err := mergeUserFiles(projectRoot, mergeableBackups, out); err != nil {
					_, _ = fmt.Fprintf(out, "  %s File merge warning: %v\n", symWarning(), err)
				}
			}
		default:
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

	_, _ = fmt.Fprintf(out, "\n%s Template sync complete.\n", symSuccess())

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "To reconfigure project settings (development mode, git, model policy), run:")
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
	forceUpdate := getBoolFlag(cmd, "force")

	// Use simple console output for progress reporting
	consoleReporter := project.NewConsoleReporter()

	// Check for version match before proceeding
	packageVersion := version.GetVersion()
	projectVersion, err := getProjectConfigVersion(projectRoot)
	if err == nil && packageVersion == projectVersion && !forceUpdate {
		_, _ = fmt.Fprintf(out, "\n%s Template version up-to-date. Skipping sync.\n", symSuccess())
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
// Returns "high" for core config files (CLAUDE.md, settings.json),
// "low" for new files, and "medium" for existing file updates.
func classifyFileRisk(filename string, exists bool) string {
	base := filepath.Base(filename)

	// High risk files
	highRiskFiles := []string{"CLAUDE.md", "settings.json"}
	if slices.Contains(highRiskFiles, base) {
		return "high"
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

// mergeGitignoreFile reads the newly deployed .gitignore template and merges
// user-specific patterns from the backup. Template content is kept as-is;
// user-added lines (not present in the template) are appended under a
// "User Custom Patterns" header.
func mergeGitignoreFile(gitignorePath string, userBackup []byte) error {
	templateContent, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("read new .gitignore: %w", err)
	}

	// Build a set of non-blank, non-comment lines from the template
	templateLines := strings.Split(string(templateContent), "\n")
	templateSet := make(map[string]bool, len(templateLines))
	for _, line := range templateLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			templateSet[trimmed] = true
		}
	}

	// Collect user-specific lines that are not in the new template
	userLines := strings.Split(string(userBackup), "\n")
	var userAdditions []string
	for _, line := range userLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !templateSet[trimmed] {
			userAdditions = append(userAdditions, line)
		}
	}

	if len(userAdditions) == 0 {
		return nil // No user-specific patterns to preserve
	}

	// Append user additions to the template content
	result := string(templateContent)
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	result += "\n# User Custom Patterns (preserved by moai update)\n"
	for _, line := range userAdditions {
		result += line + "\n"
	}

	return os.WriteFile(gitignorePath, []byte(result), defs.FilePerm)
}

// mergeUserFiles performs 3-way merge for user-customized files after template deployment.
// It uses the manifest's TemplateHash as the base, user's backed-up content as current,
// and the newly deployed template as updated. This preserves user customizations while
// incorporating template changes.
func mergeUserFiles(projectRoot string, backups []fileBackup, out io.Writer) error {
	// Load embedded templates to get original template content for base version
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}

	// Load manifest to get template hashes for base version
	mgr := manifest.NewManager()
	if _, loadErr := mgr.Load(projectRoot); loadErr != nil {
		return fmt.Errorf("load manifest: %w", loadErr)
	}

	// Create merge engine
	engine := merge.NewEngine()

	var mergedCount int
	for _, fb := range backups {
		destPath := filepath.Join(projectRoot, fb.path)

		// Read newly deployed file (updated version)
		updatedContent, err := os.ReadFile(destPath)
		if err != nil {
			// File might not exist in new template version - keep user's version
			if writeErr := os.WriteFile(destPath, fb.data, defs.FilePerm); writeErr != nil {
				return fmt.Errorf("restore removed file %s: %w", fb.path, writeErr)
			}
			_, _ = fmt.Fprintf(out, "  %s %s preserved (removed in new template)\n", symSuccess(), fb.path)
			mergedCount++
			continue
		}

		// Get original template content from embedded filesystem for base version
		// Try both with and without leading dot
		possiblePaths := []string{fb.path, strings.TrimPrefix(fb.path, ".")}
		var baseContent []byte
		for _, p := range possiblePaths {
			if data, readErr := fs.ReadFile(embedded, p); readErr == nil {
				baseContent = data
				break
			}
		}

		// Perform 3-way merge: base (original template), current (user's backup), updated (new template)
		// If base is not available, treat as new file - preserve user content
		if baseContent == nil {
			// No base available - this might be a user-created file
			// Prefer user's content but merge if compatible
			if string(fb.data) == string(updatedContent) {
				continue // No change needed
			}
			// Keep user's version as-is
			if err := os.WriteFile(destPath, fb.data, defs.FilePerm); err != nil {
				return fmt.Errorf("restore user file %s: %w", fb.path, err)
			}
			_, _ = fmt.Fprintf(out, "  %s %s user content preserved\n", symSuccess(), fb.path)
			mergedCount++
			continue
		}

		// Use merge engine for proper 3-way merge
		result, mergeErr := engine.MergeFile(context.Background(), fb.path, baseContent, fb.data, updatedContent)
		if mergeErr != nil {
			// Merge failed - preserve user's version
			_, _ = fmt.Fprintf(out, "  %s %s merge failed, preserving user version: %v\n", symWarning(), fb.path, mergeErr)
			if err := os.WriteFile(destPath, fb.data, defs.FilePerm); err != nil {
				return fmt.Errorf("preserve user file %s: %w", fb.path, err)
			}
			mergedCount++
			continue
		}

		// Write merged result
		if err := os.WriteFile(destPath, result.Content, defs.FilePerm); err != nil {
			return fmt.Errorf("write merged file %s: %w", fb.path, err)
		}

		// Report merge status
		if result.HasConflict {
			_, _ = fmt.Fprintf(out, "  %s %s merged with conflicts (user version preferred)\n", symWarning(), fb.path)
		} else {
			_, _ = fmt.Fprintf(out, "  %s %s user customizations preserved\n", symSuccess(), fb.path)
		}
		mergedCount++
	}

	if mergedCount > 0 {
		_, _ = fmt.Fprintf(out, "  %s Merged %d file(s) with 3-way merge engine\n", symSuccess(), mergedCount)
	}

	return nil
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
		if before, ok := strings.CutSuffix(tmpl, ".tmpl"); ok {
			displayPath = before
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
//   - High risk: CLAUDE.md, settings.json (core config files)
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

// getProjectConfigVersion reads the template_version from .moai/config/sections/system.yaml.
// Returns "0.0.0" if the file doesn't exist or parsing fails, which triggers a sync.
// This enables the version comparison optimization in runTemplateSync.
func getProjectConfigVersion(projectRoot string) (string, error) {
	configPath := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, defs.SystemYAML)

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

	// Parse YAML to extract moai.template_version
	var config struct {
		Moai struct {
			TemplateVersion string `yaml:"template_version"`
		} `yaml:"moai"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("parse config YAML: %w", err)
	}

	// If template_version is not set, return "0.0.0" to force update
	if config.Moai.TemplateVersion == "" {
		return "0.0.0", nil
	}

	return config.Moai.TemplateVersion, nil
}

// backupMoaiConfig creates a backup of .moai/config/ directory.
// Creates a timestamped backup under .moai-backups/YYYYMMDD_HHMMSS/ including
// all files (sections/*.yaml, etc.) for full restore capability.
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

	// Save template defaults from embedded FS for 3-way merge.
	// This allows the restore step to distinguish user-modified values
	// from unchanged template defaults.
	templateDefaultsDir := filepath.Join(backupDir, ".template-defaults")
	if err := saveTemplateDefaults(templateDefaultsDir); err != nil {
		// Non-fatal: if template defaults can't be saved, restore falls back to 2-way merge
		_, _ = fmt.Fprintf(os.Stderr, "Warning: could not save template defaults: %v\n", err)
	}

	// Create backup metadata file
	metadata := BackupMetadata{
		Timestamp:           timestamp,
		Description:         "config_backup",
		BackedUpItems:       backedUpItems,
		ExcludedItems:       excludedItems,
		ExcludedDirs:        excludedDirs,
		ProjectRoot:         projectRoot,
		BackupType:          "config",
		TemplateDefaultsDir: ".template-defaults",
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

// saveTemplateDefaults extracts config section files from embedded templates
// and saves them to the given directory as baseline for 3-way merge.
func saveTemplateDefaults(destDir string) error {
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("load embedded templates: %w", err)
	}

	// Walk embedded FS to find config section files
	prefix := ".moai/config/sections/"
	entries, err := fs.ReadDir(embedded, ".moai/config/sections")
	if err != nil {
		return fmt.Errorf("read embedded config sections: %w", err)
	}

	sectionsDestDir := filepath.Join(destDir, "sections")
	if err := os.MkdirAll(sectionsDestDir, defs.DirPerm); err != nil {
		return fmt.Errorf("create template defaults directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Read the raw file from embedded FS
		data, err := fs.ReadFile(embedded, prefix+name)
		if err != nil {
			continue // Skip files that can't be read
		}

		// For .tmpl files, save the raw template (not rendered) - the keys
		// and structure are what matter for 3-way comparison, template vars
		// will have placeholder values like {{.Version}} which won't match
		// user values, so they'll be treated as "user changed" = correct behavior.
		// Strip .tmpl extension for the output filename.
		outputName := strings.TrimSuffix(name, ".tmpl")
		if err := os.WriteFile(filepath.Join(sectionsDestDir, outputName), data, defs.FilePerm); err != nil {
			continue
		}
	}

	return nil
}

// BackupMetadata represents the structure of backup_metadata.json
type BackupMetadata struct {
	Timestamp           string   `json:"timestamp"`
	Description         string   `json:"description"`
	BackedUpItems       []string `json:"backed_up_items"`
	ExcludedItems       []string `json:"excluded_items"`
	ExcludedDirs        []string `json:"excluded_dirs"`
	ProjectRoot         string   `json:"project_root"`
	BackupType          string   `json:"backup_type"`
	TemplateDefaultsDir string   `json:"template_defaults_dir,omitempty"`
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
		_, _ = fmt.Fprintf(out, "  %s Removing %s...", symProgress(), t.displayPath)

		if t.isGlob {
			matches, err := filepath.Glob(t.fullPath)
			if err != nil {
				_, _ = fmt.Fprintf(out, "\r  %s Failed to glob %s: %v\n", symError(), t.displayPath, err)
				return fmt.Errorf("glob %s: %w", t.displayPath, err)
			}
			for _, match := range matches {
				if err := os.RemoveAll(match); err != nil {
					_, _ = fmt.Fprintf(out, "\r  %s Failed to remove %s: %v\n", symError(), t.displayPath, err)
					return fmt.Errorf("remove %s: %w", match, err)
				}
			}
			_, _ = fmt.Fprintf(out, "\r  %s Removed %s\n", symSuccess(), t.displayPath)
			continue
		}

		if _, err := os.Stat(t.fullPath); err != nil {
			if os.IsNotExist(err) {
				_, _ = fmt.Fprintf(out, "\r  - Skipped %s (not found)\n", t.displayPath)
				continue
			}
			_, _ = fmt.Fprintf(out, "\r  %s Failed to stat %s: %v\n", symError(), t.displayPath, err)
			return fmt.Errorf("stat %s: %w", t.displayPath, err)
		}

		if err := os.RemoveAll(t.fullPath); err != nil {
			_, _ = fmt.Fprintf(out, "\r  %s Failed to remove %s: %v\n", symError(), t.displayPath, err)
			return fmt.Errorf("remove %s: %w", t.displayPath, err)
		}
		_, _ = fmt.Fprintf(out, "\r  %s Removed %s\n", symSuccess(), t.displayPath)
	}

	// Clean .moai/config/ entirely - backup was already done by the Backup step.
	// For v1.x -> v2.x: old config is incompatible, fresh install needed.
	// For v2.x -> v2.x: backup includes sections/, restore will merge values back.
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
	configDisplayPath := filepath.Join(defs.MoAIDir, defs.ConfigSubdir)
	_, _ = fmt.Fprintf(out, "  %s Removing %s...", symProgress(), configDisplayPath)

	if err := os.RemoveAll(configDir); err != nil {
		if !os.IsNotExist(err) {
			_, _ = fmt.Fprintf(out, "\r  %s Failed to remove %s: %v\n", symError(), configDisplayPath, err)
			return fmt.Errorf("remove %s: %w", configDisplayPath, err)
		}
	}
	_, _ = fmt.Fprintf(out, "\r  %s Removed %s\n", symSuccess(), configDisplayPath)

	// Migrate legacy .moai/memory/ to .moai/state/.
	// Prior to v2.x, state files (checkpoints, coverage, diagnostics) lived under
	// .moai/memory/. If the old directory still exists, migrate or remove it.
	if err := migrateLegacyMemoryDir(projectRoot, out); err != nil {
		return err
	}

	return nil
}

// migrateLegacyMemoryDir handles the .moai/memory/ → .moai/state/ migration.
// If only the old directory exists, it is renamed. If both exist, the old one
// is removed (the new directory takes precedence). If neither exists, this is
// a no-op because template deployment will create .moai/state/.
func migrateLegacyMemoryDir(projectRoot string, out io.Writer) error {
	legacyDir := filepath.Join(projectRoot, defs.MoAIDir, "memory")
	stateDir := filepath.Join(projectRoot, defs.MoAIDir, defs.StateSubdir)

	legacyDisplayPath := filepath.Join(defs.MoAIDir, "memory")

	legacyExists := false
	if _, err := os.Stat(legacyDir); err == nil {
		legacyExists = true
	}

	if !legacyExists {
		return nil
	}

	_, _ = fmt.Fprintf(out, "  %s Migrating %s...", symProgress(), legacyDisplayPath)

	stateExists := false
	if _, err := os.Stat(stateDir); err == nil {
		stateExists = true
	}

	if !stateExists {
		// Rename .moai/memory/ → .moai/state/ (fast atomic move).
		if err := os.Rename(legacyDir, stateDir); err != nil {
			_, _ = fmt.Fprintf(out, "\r  %s Failed to migrate %s: %v\n", symError(), legacyDisplayPath, err)
			return fmt.Errorf("migrate %s to %s: %w", legacyDisplayPath, defs.StateSubdir, err)
		}
		_, _ = fmt.Fprintf(out, "\r  %s Migrated %s → %s\n", symSuccess(), legacyDisplayPath, filepath.Join(defs.MoAIDir, defs.StateSubdir))
	} else {
		// Both exist — state directory takes precedence; remove legacy.
		if err := os.RemoveAll(legacyDir); err != nil {
			_, _ = fmt.Fprintf(out, "\r  %s Failed to remove %s: %v\n", symError(), legacyDisplayPath, err)
			return fmt.Errorf("remove legacy %s: %w", legacyDisplayPath, err)
		}
		_, _ = fmt.Fprintf(out, "\r  %s Removed legacy %s\n", symSuccess(), legacyDisplayPath)
	}

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
// It performs a 3-way YAML merge using old template defaults as the base,
// allowing it to distinguish user-modified values from unchanged defaults.
// Falls back to 2-way merge when template defaults are not available.
func restoreMoaiConfig(projectRoot, backupDir string) error {
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
	templateDefaultsDir := filepath.Join(backupDir, ".template-defaults")

	// Check if template defaults are available for 3-way merge
	has3Way := false
	if info, err := os.Stat(templateDefaultsDir); err == nil && info.IsDir() {
		has3Way = true
	}

	// Walk through backup files (only sections/*.yaml)
	sectionsBackupDir := filepath.Join(backupDir, "sections")
	if info, err := os.Stat(sectionsBackupDir); err != nil || !info.IsDir() {
		// No sections in backup, try walking from backup root
		return restoreMoaiConfigLegacy(projectRoot, backupDir, configDir)
	}

	return filepath.Walk(sectionsBackupDir, func(backupPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Skip non-YAML files (e.g., backup_metadata.json)
		if filepath.Ext(backupPath) != ".yaml" && filepath.Ext(backupPath) != ".yml" {
			return nil
		}

		relPath, err := filepath.Rel(sectionsBackupDir, backupPath)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(configDir, "sections", relPath)

		// Read backup (old user) data
		oldData, err := os.ReadFile(backupPath)
		if err != nil {
			return err
		}

		// Check if target file exists (new template)
		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				// User's custom config section not in new template - restore as-is
				destDir := filepath.Dir(targetPath)
				if mkErr := os.MkdirAll(destDir, defs.DirPerm); mkErr != nil {
					return mkErr
				}
				return os.WriteFile(targetPath, oldData, defs.FilePerm)
			}
			return err
		}

		// Read new template data
		newData, err := os.ReadFile(targetPath)
		if err != nil {
			return err
		}

		// Try 3-way merge if template defaults are available
		if has3Way {
			basePath := filepath.Join(templateDefaultsDir, "sections", relPath)
			baseData, err := os.ReadFile(basePath)
			if err == nil {
				merged, mergeErr := mergeYAML3Way(newData, oldData, baseData)
				if mergeErr == nil {
					return os.WriteFile(targetPath, merged, defs.FilePerm)
				}
				// 3-way merge failed, fall through to 2-way
				_, _ = fmt.Fprintf(os.Stderr, "Warning: 3-way merge failed for %s, falling back to 2-way\n", relPath)
			}
		}

		// Fallback: 2-way merge (old behavior)
		merged, err := mergeYAMLDeep(newData, oldData)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: merge failed for %s, restoring backup\n", relPath)
			return os.WriteFile(targetPath, oldData, defs.FilePerm)
		}

		return os.WriteFile(targetPath, merged, defs.FilePerm)
	})
}

// restoreMoaiConfigLegacy handles restore from legacy backup format
// (pre-3-way merge) where files might be at the backup root level.
func restoreMoaiConfigLegacy(projectRoot, backupDir, configDir string) error {
	return filepath.Walk(backupDir, func(backupPath string, info os.FileInfo, err error) error {
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

		// Skip metadata and template defaults
		if filepath.Base(relPath) == "backup_metadata.json" ||
			strings.HasPrefix(relPath, ".template-defaults") {
			return nil
		}

		targetPath := filepath.Join(configDir, relPath)

		backupData, err := os.ReadFile(backupPath)
		if err != nil {
			return err
		}

		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(targetPath), defs.DirPerm); err != nil {
					return fmt.Errorf("create parent directory for %s: %w", relPath, err)
				}
				return os.WriteFile(targetPath, backupData, defs.FilePerm)
			}
			return err
		}

		targetData, err := os.ReadFile(targetPath)
		if err != nil {
			return err
		}

		merged, err := mergeYAMLDeep(targetData, backupData)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: merge failed for %s, restoring backup\n", relPath)
			return os.WriteFile(targetPath, backupData, defs.FilePerm)
		}

		return os.WriteFile(targetPath, merged, defs.FilePerm)
	})
}

// mergeYAML3Way performs a 3-way merge of YAML documents.
// It uses baseData (old template defaults) to detect user changes:
//   - If user value == base value: user didn't change it → use new template value
//   - If user value != base value: user customized it → preserve user value
//
// System fields (like template_version) always use new values regardless.
func mergeYAML3Way(newData, oldData, baseData []byte) ([]byte, error) {
	var newMap, oldMap, baseMap map[string]any

	if err := yaml.Unmarshal(newData, &newMap); err != nil {
		return nil, fmt.Errorf("unmarshal new YAML: %w", err)
	}
	if err := yaml.Unmarshal(oldData, &oldMap); err != nil {
		return nil, fmt.Errorf("unmarshal old YAML: %w", err)
	}
	if err := yaml.Unmarshal(baseData, &baseMap); err != nil {
		return nil, fmt.Errorf("unmarshal base YAML: %w", err)
	}

	merged := deepMerge3Way(newMap, oldMap, baseMap)
	return yaml.Marshal(merged)
}

// deepMerge3Way recursively performs 3-way merge of maps.
// Decision logic for each key:
//   - old == base → user didn't change → use new value
//   - old != base → user changed → preserve old value
//   - key only in new → new field added by template → use new value
//   - key only in old → removed from template → drop it
func deepMerge3Way(newMap, oldMap, baseMap map[string]any) map[string]any {
	result := make(map[string]any)

	// System fields that always use new values
	systemFields := map[string]bool{
		"template_version": true,
		"version":          true,
	}

	// Start with all new values as the base result
	for k, newV := range newMap {
		// System fields always use new value
		if systemFields[k] {
			result[k] = newV
			continue
		}

		oldV, oldExists := oldMap[k]
		baseV, baseExists := baseMap[k]

		if !oldExists {
			// Key only in new template → add it (new field)
			result[k] = newV
			continue
		}

		// Both new and old exist
		newMapVal, newIsMap := newV.(map[string]any)
		oldMapVal, oldIsMap := oldV.(map[string]any)

		if newIsMap && oldIsMap {
			// Both are maps → recurse
			baseMapVal, baseIsMap := baseV.(map[string]any)
			if !baseIsMap {
				baseMapVal = make(map[string]any)
			}
			result[k] = deepMerge3Way(newMapVal, oldMapVal, baseMapVal)
		} else {
			// Scalar or list values
			if !baseExists {
				// No base value → user added this; preserve user value
				result[k] = oldV
			} else if valuesEqual(oldV, baseV) {
				// User didn't change from template default → use new template value
				result[k] = newV
			} else {
				// User changed from template default → preserve user value
				result[k] = oldV
			}
		}
	}

	// Keys only in old (not in new template) are dropped:
	// they were removed from the template, so we don't carry them forward.

	return result
}

// valuesEqual compares two interface{} values for equality.
// Handles string, int, float, bool, and nil comparisons.
func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// mergeYAMLDeep performs a deep merge of two YAML documents (2-way fallback).
// The newData takes precedence for structure, but values from oldData are preserved
// when the key exists in both.
func mergeYAMLDeep(newData, oldData []byte) ([]byte, error) {
	var newMap, oldMap map[string]any

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
func deepMergeMaps(newMap, oldMap map[string]any) map[string]any {
	result := make(map[string]any)

	// System fields that should always use new values (not preserved from old config)
	systemFields := map[string]bool{
		"template_version": true,
	}

	// Copy all new values
	maps.Copy(result, newMap)

	// Merge old values, preserving when they exist
	for k, v := range oldMap {
		// Skip system fields - always use new value
		if systemFields[k] {
			continue
		}

		if newV, exists := newMap[k]; exists {
			// Both exist, check if they are maps
			newMapVal, newIsMap := newV.(map[string]any)
			oldMapVal, oldIsMap := v.(map[string]any)

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

	// REQ-1: Read locale from language.yaml
	locale := wizard.ReadLocaleFromProject(cwd)

	// REQ-2: Read existing username from config (used as default value)
	existingGitHubUsername := wizard.ReadGitHubUsernameFromConfig(cwd)
	existingGitLabUsername := wizard.ReadGitLabUsernameFromConfig(cwd)

	// REQ-3: Check whether gh CLI is authenticated
	ghAuthenticated := wizard.IsGhAuthenticated()

	// Generate default questions and set defaults from existing values
	questions := wizard.DefaultQuestions(cwd)
	if existingGitHubUsername != "" {
		if q := wizard.QuestionByID(questions, "github_username"); q != nil {
			q.Default = existingGitHubUsername
		}
	}
	if existingGitLabUsername != "" {
		if q := wizard.QuestionByID(questions, "gitlab_username"); q != nil {
			q.Default = existingGitLabUsername
		}
	}
	// REQ-3: Skip github_token question when gh auth is authenticated
	if ghAuthenticated {
		if q := wizard.QuestionByID(questions, "github_token"); q != nil {
			q.Condition = func(_ *wizard.WizardResult) bool { return false }
		}
	}

	// Run wizard with locale and custom questions
	result, err := wizard.RunWithLocale(questions, nil, locale)
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

	_, _ = fmt.Fprintf(out, "%s Configuration updated successfully.\n", symSuccess())

	return nil
}

// applyWizardConfig applies wizard results to the project configuration files.
func applyWizardConfig(projectRoot string, result *wizard.WizardResult) error {
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)

	// user.yaml: Save GitHub/GitLab username and token (REQ-4, REQ-5)
	hasUserFields := result.GitHubUsername != "" || result.GitHubToken != "" ||
		result.GitLabUsername != "" || result.GitLabToken != ""
	if hasUserFields {
		userPath := filepath.Join(sectionsDir, defs.UserYAML)
		// Read existing file
		userData, err := os.ReadFile(userPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read user.yaml: %w", err)
		}

		// Parse YAML
		var user map[string]any
		if len(userData) > 0 {
			if err := yaml.Unmarshal(userData, &user); err != nil {
				return fmt.Errorf("parse user.yaml: %w", err)
			}
		} else {
			user = make(map[string]any)
		}

		// Ensure user.user section exists
		var userConfig map[string]any
		if existingUser, ok := user["user"].(map[string]any); ok {
			userConfig = existingUser
		} else {
			userConfig = make(map[string]any)
		}

		// Save GitHub credentials
		if result.GitHubUsername != "" {
			userConfig["github_username"] = result.GitHubUsername
		}
		if result.GitHubToken != "" {
			userConfig["github_token"] = result.GitHubToken
		}

		// Save GitLab credentials (REQ-5)
		if result.GitLabUsername != "" {
			userConfig["gitlab_username"] = result.GitLabUsername
		}
		if result.GitLabToken != "" {
			userConfig["gitlab_token"] = result.GitLabToken
		}

		user["user"] = userConfig

		// Save to file
		updatedData, err := yaml.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user.yaml: %w", err)
		}
		if err := os.WriteFile(userPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write user.yaml: %w", err)
		}
	}

	// git-strategy.yaml: Save git mode and provider (REQ-4)
	if result.GitMode != "" || result.GitProvider != "" {
		gitStratPath := filepath.Join(sectionsDir, defs.GitStrategyYAML)
		gsData, err := os.ReadFile(gitStratPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read git-strategy.yaml: %w", err)
		}

		var gs map[string]any
		if len(gsData) > 0 {
			if err := yaml.Unmarshal(gsData, &gs); err != nil {
				return fmt.Errorf("parse git-strategy.yaml: %w", err)
			}
		} else {
			gs = make(map[string]any)
		}

		// Ensure git_strategy section exists
		var gitStrategy map[string]any
		if existing, ok := gs["git_strategy"].(map[string]any); ok {
			gitStrategy = existing
		} else {
			gitStrategy = make(map[string]any)
		}

		if result.GitMode != "" {
			gitStrategy["mode"] = result.GitMode
		}
		if result.GitProvider != "" {
			gitStrategy["provider"] = result.GitProvider
		}

		// Save GitLab instance URL (REQ-5)
		if result.GitLabInstanceURL != "" {
			var gitlabSection map[string]any
			if existing, ok := gitStrategy["gitlab"].(map[string]any); ok {
				gitlabSection = existing
			} else {
				gitlabSection = make(map[string]any)
			}
			gitlabSection["instance_url"] = result.GitLabInstanceURL
			gitStrategy["gitlab"] = gitlabSection
		}

		gs["git_strategy"] = gitStrategy

		updatedData, err := yaml.Marshal(gs)
		if err != nil {
			return fmt.Errorf("marshal git-strategy.yaml: %w", err)
		}
		if err := os.WriteFile(gitStratPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write git-strategy.yaml: %w", err)
		}
	}

	// Apply model policy to agent definition files (project-level, not profile-level)
	if result.ModelPolicy != "" {
		policy := template.ModelPolicy(result.ModelPolicy)
		if template.IsValidModelPolicy(string(policy)) {
			mgr := manifest.NewManager()
			if _, err := mgr.Load(projectRoot); err == nil {
				if err := template.ApplyModelPolicy(projectRoot, policy, mgr); err != nil {
					return fmt.Errorf("apply model policy: %w", err)
				}
			}
			// Persist model_policy to system.yaml so it survives future updates
			systemPath := filepath.Join(sectionsDir, defs.SystemYAML)
			systemData, _ := os.ReadFile(systemPath)
			var sys map[string]any
			if len(systemData) > 0 {
				_ = yaml.Unmarshal(systemData, &sys)
			}
			if sys == nil {
				sys = make(map[string]any)
			}
			moaiSection, _ := sys["moai"].(map[string]any)
			if moaiSection == nil {
				moaiSection = make(map[string]any)
			}
			moaiSection["model_policy"] = string(policy)
			sys["moai"] = moaiSection
			if updatedData, err := yaml.Marshal(sys); err == nil {
				_ = os.WriteFile(systemPath, updatedData, defs.FilePerm)
			}
		}
	}

	// quality.yaml: Save development mode (REQ-4)
	if result.DevelopmentMode != "" {
		qualityPath := filepath.Join(sectionsDir, defs.QualityYAML)
		qualityData, err := os.ReadFile(qualityPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read quality.yaml: %w", err)
		}

		var quality map[string]any
		if len(qualityData) > 0 {
			if err := yaml.Unmarshal(qualityData, &quality); err != nil {
				return fmt.Errorf("parse quality.yaml: %w", err)
			}
		} else {
			quality = make(map[string]any)
		}

		// Ensure constitution section exists
		var constitution map[string]any
		if existing, ok := quality["constitution"].(map[string]any); ok {
			constitution = existing
		} else {
			constitution = make(map[string]any)
		}

		constitution["development_mode"] = result.DevelopmentMode
		quality["constitution"] = constitution

		updatedData, err := yaml.Marshal(quality)
		if err != nil {
			return fmt.Errorf("marshal quality.yaml: %w", err)
		}
		if err := os.WriteFile(qualityPath, updatedData, defs.FilePerm); err != nil {
			return fmt.Errorf("write quality.yaml: %w", err)
		}
	}

	return nil
}

// allStatuslineSegments lists all supported statusline segment names in display order.
var allStatuslineSegments = []string{
	statusline.SegmentModel, statusline.SegmentContext, statusline.SegmentOutputStyle, statusline.SegmentDirectory,
	statusline.SegmentGitStatus, statusline.SegmentClaudeVersion, statusline.SegmentMoaiVersion, statusline.SegmentGitBranch,
}

// presetToSegments converts a statusline preset name and optional custom segment map
// into a full segment-to-enabled map. Unknown presets fall back to "full" (all enabled).
func presetToSegments(preset string, custom map[string]bool) map[string]bool {
	segments := make(map[string]bool, len(allStatuslineSegments))

	switch preset {
	case "compact":
		compactEnabled := map[string]bool{
			statusline.SegmentModel: true, statusline.SegmentContext: true, statusline.SegmentGitStatus: true, statusline.SegmentGitBranch: true,
		}
		for _, seg := range allStatuslineSegments {
			segments[seg] = compactEnabled[seg]
		}
	case "minimal":
		minimalEnabled := map[string]bool{
			statusline.SegmentModel: true, statusline.SegmentContext: true,
		}
		for _, seg := range allStatuslineSegments {
			segments[seg] = minimalEnabled[seg]
		}
	case "custom":
		if custom == nil {
			// No custom selections provided, default all to true
			for _, seg := range allStatuslineSegments {
				segments[seg] = true
			}
		} else {
			for _, seg := range allStatuslineSegments {
				val, exists := custom[seg]
				if exists {
					segments[seg] = val
				} else {
					segments[seg] = true // Default missing segments to enabled
				}
			}
		}
	default:
		// "full" and any unknown preset: all segments enabled
		for _, seg := range allStatuslineSegments {
			segments[seg] = true
		}
	}

	return segments
}

// settingsLocalEnv represents the structure of .claude/settings.local.json.
type settingsLocalEnv struct {
	Env map[string]string `json:"env,omitempty"`
}

// updateSettingsLocalEnv updates a single environment variable in settings.local.json.
// If the file doesn't exist, it creates a new one. If the env map doesn't exist, it creates it.
func updateSettingsLocalEnv(settingsPath, key, value string) error {
	var settings settingsLocalEnv

	// Read existing settings if file exists
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings.local.json: %w", err)
		}
	}

	// Initialize env map if nil
	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	// Set the environment variable
	settings.Env[key] = value

	// Marshal back to JSON
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings.local.json: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, defs.FilePerm); err != nil {
		return fmt.Errorf("write settings.local.json: %w", err)
	}

	return nil
}

// ensureGlobalSettingsEnv cleans up moai-managed settings from ~/.claude/settings.json.
// All settings (env, permissions, teammateMode, hooks) are managed at the project level.
// The global hooks directory (~/.claude/hooks/moai/) is also removed since hooks
// are only deployed to project-level directories via moai init.
func ensureGlobalSettingsEnv() error {
	homeDir, err := userHomeDir()
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
	var existingSettings map[string]any
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
	// Note: CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS is kept as a default value (see below).
	if envVal, exists := existingSettings["env"]; exists {
		if envMap, ok := envVal.(map[string]any); ok {
			moaiKeys := []string{"PATH", "ENABLE_TOOL_SEARCH"}
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

	// Ensure default global settings are present.
	// CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 enables Agent Teams mode by default.
	defaultEnvKeys := map[string]string{
		"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1",
	}
	for key, value := range defaultEnvKeys {
		if envVal, exists := existingSettings["env"]; exists {
			if envMap, ok := envVal.(map[string]any); ok {
				if _, exists := envMap[key]; !exists {
					envMap[key] = value
					needsUpdate = true
				}
			}
		} else {
			// No env section yet, create it with defaults
			existingSettings["env"] = map[string]any{
				key: value,
			}
			needsUpdate = true
		}
	}

	// Clean up moai-managed permissions if they only contain Task:*
	if permVal, exists := existingSettings["permissions"]; exists {
		if permMap, ok := permVal.(map[string]any); ok {
			if allowVal, exists := permMap["allow"]; exists {
				if allowArr, ok := allowVal.([]any); ok {
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
func cleanLegacyHooks(settings map[string]any) bool {
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
		"post_tool__code_formatter.py",
		"post_tool__linter.py",
		"post_tool__ast_grep_scan.py",
	}

	hooksMap, ok := settings["hooks"].(map[string]any)
	if !ok {
		return false
	}

	modified := false
	for hookType, hookListInterface := range hooksMap {
		hookList, ok := hookListInterface.([]any)
		if !ok {
			continue
		}

		var cleanedHooks []any
		for _, hookGroup := range hookList {
			groupMap, ok := hookGroup.(map[string]any)
			if !ok {
				cleanedHooks = append(cleanedHooks, hookGroup)
				continue
			}

			hooksList, ok := groupMap["hooks"].([]any)
			if !ok {
				cleanedHooks = append(cleanedHooks, hookGroup)
				continue
			}

			var cleanedGroupHooks []any
			for _, hookEntry := range hooksList {
				entryMap, ok := hookEntry.(map[string]any)
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
