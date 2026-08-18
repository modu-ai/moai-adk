package cli

// Template-sync half of the update command. Extracted verbatim from update.go
// by SPEC-CLIFIX-HYGIENE-001 M5 (mechanical move, no logic change) to bring
// update.go under the 1,200-line ceiling.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	updatemerge "github.com/modu-ai/moai-adk/internal/cli/update/merge"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
)

// runTemplateSync synchronizes embedded templates with the project directory.
// It performs a quick version comparison first - if the project's template version
// matches the package version, the sync is skipped for performance (70-80% faster).
//
// Template deployment uses a 3-way merge strategy to preserve local modifications.
// Users are prompted to confirm the merge before proceeding.
func runTemplateSync(cmd *cobra.Command) error {
	return runTemplateSyncWithReporter(cmd, nil, false)
}

// @MX:NOTE: [AUTO] runTemplateSyncWithReporter — M4-S4d-2 DDD migration. Top-level header/section/
// final-outcome lines are converted to tui.KV / tui.Section / tui.CheckLine / tui.Pill. Sub-step
// micro messages (\r-prefixed sym* helpers) are preserved because they drive the progress display.
// Design source: screens.jsx ScreenUpdate.
//
// runTemplateSyncWithReporter synchronizes templates with progress reporting.
func runTemplateSyncWithReporter(cmd *cobra.Command, reporter project.ProgressReporter, skipConfirm bool) error {
	out := cmd.OutOrStdout()
	th := resolveTheme()
	ctx := cmd.Context()

	// Get flags for template sync
	forceBackup := getBoolFlag(cmd, "force")
	autoConfirm := getBoolFlag(cmd, "yes")

	// Use current directory as project root
	projectRoot := "."

	currentVersion := version.GetVersion()
	// Identity header band (REQ-TUXIU-015): "◆ MoAI-ADK <version> <go-runtime>
	// · claude" with the version as a solid brand pill.
	_, _ = fmt.Fprintln(out, renderIdentityBand(currentVersion, th))
	_, _ = fmt.Fprintln(out, tui.KV("Current version", "moai-adk "+currentVersion, tui.KVOpts{Theme: &th, KeyWidth: 16}))
	_, _ = fmt.Fprintln(out, tui.CheckLine("run", "Syncing templates", "from embedded filesystem", "", &th))

	if reporter != nil {
		reporter.StepStart("Version Check", "Checking template version...")
	}

	// Stage 2: Config Version Comparison (before template sync)
	// Compare package template_version with project config template_version
	// If versions match, skip sync for performance (70-80% faster)
	packageVersion := version.GetVersion()
	projectVersion, err := plan.GetProjectConfigVersion(projectRoot)
	if err == nil && packageVersion == projectVersion && !forceBackup {
		if reporter != nil {
			reporter.StepComplete("Already up-to-date")
		}
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, ""), Theme: &th}))
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

	// t40 defect 2: AnalyzeFiles skips IsMoaiManaged paths, so analysis.Files
	// carries only the merged/added files. Count the managed re-deployments
	// the deploy writes anyway (and keep the rendered-target set for the
	// removal accounting) so the outcome summary reports the real total.
	templateFiles := deployer.ListTemplates()
	managedRedeployed := 0
	restoredSet := make(map[string]bool, len(templateFiles))
	for _, tmpl := range templateFiles {
		if before, ok := strings.CutSuffix(tmpl, ".tmpl"); ok {
			tmpl = before
		}
		restoredSet[filepath.ToSlash(tmpl)] = true
		if plan.IsMoaiManaged(tmpl) {
			managedRedeployed++
		}
	}

	// Analyze merge changes. The "Analyzing merge changes" header + the
	// classification card are shown only when this function owns the
	// confirmation flow (skipConfirm=false). When skipConfirm=true the caller
	// (runTemplateSyncWithProgress) already printed both before the user's
	// y/n confirmation — reprinting them here duplicates the visible output.
	analysis := updatemerge.AnalyzeMergeChanges(deployer, projectRoot)

	if !skipConfirm {
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Section("Analyzing merge changes", tui.SectionOpts{Theme: &th}))
		// Card-style classification summary (REQ-TUXIU-010/011): accent box with
		// up to three count pills; zero-count pills omitted; suppressed entirely
		// when the run is clean (all counts zero).
		addCount, updateCount, conflictCount := classifyUpdateCounts(analysis.Files)
		if card := renderClassificationSummary(addCount, updateCount, conflictCount, th); card != "" {
			_, _ = fmt.Fprintln(out, card)
		}
	}

	if reporter != nil {
		reporter.StepUpdate("Found " + fmt.Sprintf("%d files to sync", len(analysis.Files)))
	}

	// Skip confirmation if --yes flag is provided (CI/CD mode) or pre-confirmed
	var proceed bool
	if skipConfirm {
		proceed = true
	} else if autoConfirm {
		proceed = true
		_, _ = fmt.Fprintln(out, tui.CheckLine("info", "Auto-confirm", "CI/CD mode", "", &th))
	} else {
		var err error
		proceed, err = confirmViaPreview(analysis, projectRoot)
		if err != nil {
			if reporter != nil {
				reporter.StepError(err)
			}
			return fmt.Errorf("confirm merge for %d files (risk: %s): %w",
				len(analysis.Files), analysis.RiskLevel, err)
		}
	}

	if !proceed {
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillNeutral, Solid: false, Label: "Merge cancelled by user", Theme: &th}))
		if reporter != nil {
			reporter.StepError(errors.New("cancelled by user"))
		}
		return nil
	}

	// Deploy templates
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, tui.Section("Proceeding with template deployment", tui.SectionOpts{Theme: &th}))
	_, _ = fmt.Fprintln(out)

	// Track config backup path for restore step.
	//
	// Declared before the step table because the Clean Managed Paths step needs
	// it: SPEC-UPDATE-DATA-SURVIVAL-001 M3 routes that step through
	// guardFirstDestructiveStep, which writes the crash-window copies into this
	// run-scoped directory before anything is removed.
	var configBackupPath string

	// t40 defect 2: read-only snapshot of the managed roots, taken inside the
	// Clean step immediately before the removal, so the outcome summary can
	// account for what was deleted (and what the templates do not restore).
	var preCleanFiles []string

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
				// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
				// the legacy CR-plus-format pair (REQ-UPR-004).
				pl := tui.ProgressLine(out, "Backing up .moai/config...", nil)
				configBackupPath, backupErr := backup.BackupMoaiConfig(projectRoot)
				if backupErr != nil {
					pl.Fail(fmt.Sprintf("Backup failed: %v", backupErr))
					return backupErr
				}
				if configBackupPath != "" {
					pl.Done(".moai/config backed up")
				} else {
					pl.Done("No config to backup")
				}
				return nil
			},
		},
		{
			name:    "Validate Templates",
			message: "Validating all templates before deployment",
			execute: func() error {
				// Seam-routed per REQ-UGE-004 (test isolation).
				homeDir, _ := userHomeDirFn()
				goBinPath := detectGoBinPathForUpdate(homeDir)
				tmplCtx := template.NewTemplateContext(
					template.WithGoBinPath(goBinPath),
					template.WithResolvedMoaiPath(resolveMoaiExecutable()),
					template.WithHomeDir(homeDir),
					template.WithSmartPATH(template.BuildSmartPATH()),
					template.WithPlatform(runtime.GOOS),
					template.WithVersion(version.GetVersion()),
					template.WithHookOptIn(readHookOptInEnabled(projectRoot)),
				)

				// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
				// the legacy CR-plus-format pair (REQ-UPR-004).
				pl := tui.ProgressLine(out, "Validating templates...", nil)
				if validateErr := deployer.ValidateAll(ctx, tmplCtx); validateErr != nil {
					pl.Fail(fmt.Sprintf("Template validation failed: %v", validateErr))
					return fmt.Errorf("template validation: %w", validateErr)
				}
				pl.Done("All templates validated")
				return nil
			},
		},
		{
			name:    cleanManagedPathsStage,
			message: "Removing old MoAI-managed files",
			execute: func() error {
				// t40 defect 2: snapshot what exists under the managed roots
				// BEFORE the removal (read-only; the deletion below is
				// unchanged).
				preCleanFiles = deploy.InventoryManagedPaths(projectRoot)
				// SPEC-UPDATE-DATA-SURVIVAL-001 REQ-UDS-001/003/005: the three
				// in-memory-only files reach disk before this step removes
				// anything. A backup-write failure aborts here, so the removal
				// never runs while a file's only copy is in the heap.
				// Card t111: the embedded FS rides along so the removal can
				// back up every file the template does not carry before
				// deleting the root it lives under.
				return guardFirstDestructiveStep(projectRoot, configBackupPath, func() error {
					tmplFS, tmplErr := template.EmbeddedTemplates()
					if tmplErr != nil {
						// Without the template FS the removal cannot tell
						// managed from unmanaged files, so it cannot back
						// anything up — abort rather than delete blind.
						return fmt.Errorf("load embedded templates: %w", tmplErr)
					}
					return deploy.CleanMoaiManagedPaths(projectRoot, out, tmplFS)
				})
			},
		},
		{
			name:    "Deploy Templates",
			message: "Deploying template files",
			execute: func() error {
				// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
				// the legacy CR-plus-format pair (REQ-UPR-004).
				pl := tui.ProgressLine(out, "Deploying templates...", nil)

				// Build TemplateContext with detected paths for template rendering.
				// Seam-routed per REQ-UGE-004 (test isolation).
				homeDir, _ := userHomeDirFn()
				goBinPath := detectGoBinPathForUpdate(homeDir)
				tmplCtx := template.NewTemplateContext(
					template.WithGoBinPath(goBinPath),
					template.WithResolvedMoaiPath(resolveMoaiExecutable()),
					template.WithHomeDir(homeDir),
					template.WithSmartPATH(template.BuildSmartPATH()),
					template.WithPlatform(runtime.GOOS),
					template.WithVersion(version.GetVersion()),
					template.WithHookOptIn(readHookOptInEnabled(projectRoot)),
				)

				if deployErr := deployer.Deploy(ctx, projectRoot, mgr, tmplCtx); deployErr != nil {
					pl.Fail(fmt.Sprintf("Deployment failed: %v", deployErr))
					return fmt.Errorf("deploy templates: %w", deployErr)
				}
				pl.Done("Templates deployed")
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

	// SPEC-UPDATE-DATA-SURVIVAL-001 REQ-UDS-019/020: once the Clean Managed
	// Paths step completes, the tree is irreversibly changed. Every later step
	// failure is a partial-update failure and must leave the operator a
	// recovery manifest naming the backup and the restore command. No automatic
	// rollback is attempted (plan.md §B).
	recovery := newRecoveryGuard(projectRoot, "", out)
	// Backup of user's .gitignore content for EntryMerge after deploy
	var gitignoreBackup []byte
	// Backups of mergeable files for 3-way merge after deploy
	var mergeableBackups []updatemerge.FileBackup

	// collectMergeableFiles returns a list of files that should be merged
	// using the 3-way merge engine during update.
	// Note: .moai/config/sections/*.yaml files are already handled by
	// restoreMoaiConfig with 3-way merge, so they are excluded here.
	collectMergeableFiles := func(projectRoot string) []string {
		// Fixed mergeable files at project root that are NOT handled by restoreMoaiConfig.
		// .mcp.json IS shipped (internal/template/templates/.mcp.json, deployed by
		// deployer.go's no-dotfile-skip WalkDir) and IS a 3-way merge target so a
		// user's own MCP entries survive `moai update` instead of being clobbered
		// by the template deploy (SPEC-MCP-DEFAULT-ON-001 REQ-A-4 / AC-A-012).
		return []string{
			".claude/settings.json",
			".moai/status_line.sh",
			".mcp.json",
		}
	}

	// Execute each step with progress reporting. The per-step in-flight line is
	// rendered by the inline tui.ProgressLine primitive inside each step body;
	// the coarse reporter Step wrapper is intentionally NOT driven here — doing
	// so double-rendered every step on the stderr channel (the printer.Step
	// in-flight line) alongside the stdout ProgressLine, producing the two-part
	// "○…○" spinner residue on a TTY (REQ-TUXIU-020/021). Errors still surface
	// via reporter.StepError (orphan-safe) below.
	for i, step := range steps {
		// Special handling for backup/restore steps; default executes normally
		switch step.name {
		case "Backup":
			// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
			// the legacy CR-plus-format pair (REQ-UPR-004).
			plBackup := tui.ProgressLine(out, "Backing up .moai/config...", nil)
			var backupErr error
			configBackupPath, backupErr = backup.BackupMoaiConfig(projectRoot)
			if backupErr != nil {
				plBackup.Fail(fmt.Sprintf("Backup failed: %v", backupErr))
				if reporter != nil {
					reporter.StepError(backupErr)
				}
				return backupErr
			}
			if configBackupPath != "" {
				plBackup.Done(".moai/config backed up")
			} else {
				plBackup.Done("No config to backup")
			}
			// The run-scoped backup directory is only known here; it hosts the
			// recovery manifest a later failure writes (REQ-UDS-019).
			recovery.backupDir = configBackupPath

			// SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 M3: user-owned namespace backup
			// (REQ-UNP-004). Sequential after .moai/config backup. Skips silently
			// when no user-owned content exists (EC-UNP-001).
			plNsBackup := tui.ProgressLine(out, "Backing up user-owned namespace...", nil)
			nsBackupPath, nsBackupErr := backupUserOwnedNamespace(projectRoot)
			if nsBackupErr != nil {
				plNsBackup.Fail(fmt.Sprintf("Namespace backup failed: %v", nsBackupErr))
				if reporter != nil {
					reporter.StepError(nsBackupErr)
				}
				return nsBackupErr
			}
			if nsBackupPath != "" {
				// Derive display-friendly project-root-relative path
				displayNs := nsBackupPath
				if rel, relErr := filepath.Rel(projectRoot, nsBackupPath); relErr == nil {
					displayNs = rel
				}
				plNsBackup.Done(fmt.Sprintf("User-owned namespace backed up: %s", displayNs))
			} else {
				plNsBackup.Done("No user-owned namespace to back up")
			}

			// SPEC-INTERNAL-SECURITY-001 REQ-SEC-006 (AC-SEC-006a): wire the
			// pre-modification abort sentinel as a REAL gate on the moai update
			// deploy path. verifyNamespaceBackupCoverage runs AFTER the backup
			// completes and BEFORE any destructive deploy step; it aborts with
			// UPDATE_USER_NAMESPACE_VIOLATION when a user-owned namespace file
			// on disk would be overwritten without a backup. Normal updates
			// (no user-owned content, or fully backed up) pass through
			// unchanged (NFR-SEC-003).
			if covErr := verifyNamespaceBackupCoverage(projectRoot, nsBackupPath); covErr != nil {
				plNsBackup.Fail(fmt.Sprintf("Namespace safety check failed: %v", covErr))
				if reporter != nil {
					reporter.StepError(covErr)
				}
				return covErr
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
					mergeableBackups = append(mergeableBackups, updatemerge.FileBackup{Path: mf, Data: data})
				}
			}
		case "Restore Settings":
			// Handle restore step with captured backup path
			if configBackupPath != "" {
				// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces
				// the legacy CR-plus-format pair (REQ-UPR-004).
				plRestore := tui.ProgressLine(out, "Restoring user settings...", nil)
				// t63: RestoreMoaiConfigRetained collects the retained-key refs
				// instead of letting the merge append raw per-key "advisory:"
				// lines to stderr mid-redraw; the advisory renders through the
				// same stdout channel as the progress line below.
				retainedKeys, restoreErr := backup.RestoreMoaiConfigRetained(projectRoot, configBackupPath, func(pr, relPath string, success bool, errOut io.Writer) {
					// Bridge to the noise-suppression ledger (recordMergeFallback +
					// updateVerboseMode), which stays in package cli. The closure
					// captures updateVerboseMode so the backup subpackage does not
					// need a cross-package mutable-state seam.
					recordMergeFallback(pr, relPath, success, updateVerboseMode, errOut)
				})
				if restoreErr != nil {
					plRestore.Fail(fmt.Sprintf("Restore failed: %v", restoreErr))
					if reporter != nil {
						reporter.StepError(restoreErr)
					}
					return recovery.fail(step.name, restoreErr)
				}
				plRestore.Done("User settings restored")
				// t63: one summary line by default; the key list expands only
				// under --verbose (the same verbose ledger recordMergeFallback
				// reads), never interleaving with the progress redraw.
				renderRetainedKeyAdvisory(out, retainedKeys, updateVerboseMode, th)
				deletedCount := backup.CleanupOldBackups(projectRoot, 5)
				if deletedCount > 0 {
					_, _ = fmt.Fprintf(out, "  %s Cleaned up %d old backup(s)\n", uikit.SymSuccess(), deletedCount)
				}
				// SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001 (REQ-TBS-002, Decision
				// D4 trigger #2): capture the post-restore on-disk config so the
				// next update has a rendered BASE. Best-effort non-blocking.
				writeTemplateSnapshotBestEffort(projectRoot, out)
			}
			// Merge .gitignore: preserve user-added patterns via EntryMerge
			if len(gitignoreBackup) > 0 {
				gitignorePath := filepath.Join(projectRoot, ".gitignore")
				if mergeErr := updatemerge.MergeGitignoreFile(gitignorePath, gitignoreBackup); mergeErr != nil {
					_, _ = fmt.Fprintf(out, "  %s .gitignore merge warning: %v\n", uikit.SymWarning(), mergeErr)
				} else {
					_, _ = fmt.Fprintf(out, "  %s .gitignore user patterns preserved\n", uikit.SymSuccess())
				}
			}
			// Merge user-customized files using 3-way merge engine
			if len(mergeableBackups) > 0 {
				if err := updatemerge.MergeUserFiles(projectRoot, mergeableBackups, out); err != nil {
					_, _ = fmt.Fprintf(out, "  %s File merge warning: %v\n", uikit.SymWarning(), err)
				}
			}
		default:
			// Execute normal step under the recovery guard: a failure after the
			// destructive Clean Managed Paths step writes and prints the
			// recovery manifest (REQ-UDS-019) before the error propagates.
			if err := runUpdateStages(recovery, []updateStage{{
				name:        step.name,
				run:         step.execute,
				destructive: step.name == cleanManagedPathsStage,
			}}); err != nil {
				if reporter != nil {
					reporter.StepError(err)
				}
				return err
			}
		}

		// Block progress bar reflecting completed/total deploy steps
		// (REQ-TUXIU-014). Replaces the legacy "N/M steps complete" reporter
		// text; the bar rides the same stdout channel as the step ProgressLines.
		_, _ = fmt.Fprintln(out, renderDeployProgress(i+1, len(steps), th))
	}

	_, _ = fmt.Fprintln(out)
	// Outcome banner (REQ-TUXIU-016): solid success pill + dim detail note.
	// t40 defect 2: the pill total includes the managed re-deployments and
	// the note carries the removal accounting (local-only losses named by
	// count).
	detail := updateOutcomeDetail{ManagedRedeployed: managedRedeployed}
	for _, f := range preCleanFiles {
		detail.RemovedManaged++
		if !restoredSet[f] {
			detail.RemovedLocalOnly++
		}
	}
	renderUpdateOutcome(out, len(analysis.Files), detail, configBackupPath, th)
	report.EmitHooksReviewGuidance(out)

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "To reconfigure project settings (development mode, git, model policy), run:")
	_, _ = fmt.Fprintln(out, "   moai update -c")

	// Ensure global settings.json has required env variables
	if err := ensureGlobalSettingsEnv(); err != nil {
		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Global settings env", "update failed", err.Error(), &th))
	}

	// Install pre-push hook (REQ-CIAUT-002). Non-fatal; --no-hooks opts out.
	installPrePushHookOptional(projectRoot, getBoolFlag(cmd, "no-hooks"), out)

	// Install pre-commit hook (REQ-PC-001). Fast-subset commit tier; --no-hooks opts out.
	installPreCommitHookOptional(projectRoot, getBoolFlag(cmd, "no-hooks"), out)

	// SPEC-WORKTREE-BRANCH-GUARD-001 (REQ-WBG-009): one-line worktree advisory on
	// update completion. The primary checkout is shared; branch-changing work
	// belongs in a worktree.
	emitWorktreeAdvisory(out, projectRoot)

	return nil
}

// @MX:NOTE: [AUTO] runTemplateSyncWithProgress — M4-S4d-2 DDD migration. Console reporter
// wrapper. Only the key outputs are converted: tui.Pill (skip), tui.Section (analyzing), tui.Pill (cancel).
//
// @MX:NOTE: [AUTO] SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 — return shape extended
// to (skipped bool, err error) so that runUpdate can short-circuit the
// downstream legacy-skill archive block when sync is skipped (version-match
// without --force). skipped == true means "no template files were written";
// callers MUST NOT trigger archive checks in that case.
//
// runTemplateSyncWithProgress runs template sync with simple console output.
// Return values:
//   - skipped: true when version matches and --force is absent (sync was a no-op).
//     When skipped == true, callers should not trigger downstream archive checks.
//   - err: any non-skip error encountered.
//
// A skipped sync is not an error; callers receive (true, nil).
// A user-cancelled merge is also (true, nil) — it is a no-op for downstream
// purposes (no files written), so archive does not need to run.
func runTemplateSyncWithProgress(cmd *cobra.Command) (skipped bool, err error) {
	out := cmd.OutOrStdout()
	th := resolveTheme()
	projectRoot := "."
	autoConfirm := getBoolFlag(cmd, "yes")
	forceUpdate := getBoolFlag(cmd, "force")

	// Use printer-backed console output for progress reporting
	// (REQ-CTX-015: routes step events through the Printer to stderr).
	consoleReporter := newPrinterReporter(
		printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr())))

	// Check for version match before proceeding
	packageVersion := version.GetVersion()
	projectVersion, verr := plan.GetProjectConfigVersion(projectRoot)
	if verr == nil && packageVersion == projectVersion && !forceUpdate {
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, ""), Theme: &th}))
		return true, nil
	}

	// Confirm merge before proceeding (unless auto-confirm is set)
	if !autoConfirm {
		embedded, eerr := template.EmbeddedTemplates()
		if eerr != nil {
			return false, fmt.Errorf("load embedded templates: %w", eerr)
		}

		deployer := template.NewDeployerWithForceUpdate(embedded, true)
		analysis := updatemerge.AnalyzeMergeChanges(deployer, projectRoot)

		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Section("Analyzing merge changes", tui.SectionOpts{Theme: &th}))
		proceed, cerr := confirmViaPreview(analysis, projectRoot)
		if cerr != nil {
			return false, fmt.Errorf("confirm merge for %d files (risk: %s): %w",
				len(analysis.Files), analysis.RiskLevel, cerr)
		}
		if !proceed {
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillNeutral, Solid: false, Label: "Merge cancelled by user", Theme: &th}))
			return true, nil
		}
	}

	return false, runTemplateSyncWithReporter(cmd, consoleReporter, true)
}

// toPreviewInputs maps a merge.MergeAnalysis into the neutral
// update.FilePreviewInput slice consumed by the preview entry point (M3c
// convergence, SPEC-CLI-TUX-V3-003 plan Known Issue #7). The Exists/Conflict
// derivation mirrors the same signals the deploy stage enforces — Exists from
// os.Stat (mirrors analyzeFiles), Conflict from RiskLevel=="high" (mirrors
// buildMergeAnalysis hasConflicts). The classification itself (which file is
// preserved/added/updated/conflict) is derived downstream by update.Classify
// via the injected isUserOwnedNamespace predicate (REQ-TUX3-001/002 single
// source of truth — this mapping introduces NO parallel heuristic).
func toPreviewInputs(analysis merge.MergeAnalysis, projectRoot string) []update.FilePreviewInput {
	inputs := make([]update.FilePreviewInput, 0, len(analysis.Files))
	for _, fa := range analysis.Files {
		_, statErr := os.Stat(filepath.Join(projectRoot, fa.Path))
		inputs = append(inputs, update.FilePreviewInput{
			RelPath:  fa.Path,
			Exists:   statErr == nil,
			Conflict: fa.RiskLevel == "high",
			Diff:     fa.Changes,
		})
	}
	return inputs
}

// confirmViaPreview is the single convergence surface for BOTH merge.ConfirmMerge
// call sites in this file (plan Known Issue #7). It launches the Bubble Tea v2
// TUI (AC-TUX3-008/009) when stdin is a terminal, classifying through
// update.Classify with the shared isUserOwnedNamespace predicate
// (REQ-TUX3-001/002/014 — no parallel heuristic, `preserved (user-owned)` label
// derived from the shared predicate).
//
// Behavior preservation relative to the prior merge.ConfirmMerge calls: both
// call sites reach this helper ONLY on the interactive (non-skipped, non-auto)
// path — the skip-confirm and --yes (autoConfirm) branches short-circuit
// upstream and set proceed=true directly, byte-identical to before. When stdin
// is NOT a terminal, the user cannot be asked interactively, so this helper
// returns an error directing them to --yes — mirroring both the explicit
// Windows guard that merge.ConfirmMerge carried (confirm.go REQ-CFS-007/008)
// and the de-facto tea.Run() non-TTY failure on darwin/linux that the v1
// ConfirmMerge relied on. It MUST NOT silently proceed in the non-TTY case:
// that would bypass the confirmation gate the caller explicitly entered (the
// caller did not pass --yes), changing which files get deployed. The
// preview-fallback's proceed=true semantics belong to the --yes abstraction,
// which never reaches this helper.
func confirmViaPreview(analysis merge.MergeAnalysis, projectRoot string) (bool, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return false, fmt.Errorf("merge confirmation UI requires an interactive terminal; " +
			"rerun with --yes to auto-confirm in a non-TTY environment")
	}
	return update.PreviewClassification(toPreviewInputs(analysis, projectRoot), plan.IsUserOwnedNamespace, update.PreviewOptions{Interactive: true})
}
