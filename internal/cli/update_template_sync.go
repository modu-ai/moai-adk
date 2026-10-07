package cli

// Template-sync half of the update command. Extracted verbatim from update.go
// by SPEC-CLIFIX-HYGIENE-001 M5 (mechanical move, no logic change) to bring
// update.go under the 1,200-line ceiling.

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	updatemerge "github.com/modu-ai/moai-adk/internal/cli/update/merge"
	"github.com/modu-ai/moai-adk/internal/cli/update/plan"
	"github.com/modu-ai/moai-adk/internal/cli/update/report"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/merge"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/internal/tui"
	"github.com/modu-ai/moai-adk/pkg/version"
	"github.com/spf13/cobra"
)

// newTemplateSyncDeployer builds the deployer the template-sync path uses.
//
// It is a package-level variable so a test can substitute a double and drive
// the mirror-notice wiring — the real fallback needs a symlink syscall to fail,
// and the seams for that are unexported in internal/template.
//
// The default MUST stay the production deployer AND MUST satisfy
// template.ResultDeployer. A default that reports no result would leave the
// mirror notice permanently unreachable in production while every
// injected-double test stayed green, which is the one failure this seam could
// introduce; TestSeamDefaultIsTheProductionDeployer and
// TestSeamDefaultSatisfiesResultDeployer guard both halves.
//
// SPEC-INIT-HARNESS-001 (REQ-IH-010): the seam reads llm.harness from the
// live config — a codex-only project re-deploys through the codex-only
// deployer (force-update semantics preserved) instead of resurrecting the
// Claude surfaces. The other profiles use their matching deployers. A
// catalog/construction error aborts update instead of changing profiles.
//
// SPEC-INIT-SHRINK-001 (REQ-016): the deploy mode rides through as an
// option — a plugin-mode project deploys the thin set, a local-mode project
// today's full set. The caller resolves the mode from the record (and the
// migration's shape for a record-less project). On plugin mode the mirror
// policy is None: REQ-019 holds the .agents/skills entries STABLE on update
// runs — the re-home belongs to the fresh plugin deploy (init), and the
// update never adds, restores, or rewrites mirror entries.
var newTemplateSyncDeployer = func(embedded fs.FS, _ template.DeployMode) (template.Deployer, error) {
	renderer := template.NewRenderer(embedded)
	cat, catErr := template.LoadEmbeddedCatalog()
	if catErr != nil {
		return nil, fmt.Errorf("load harness catalog: %w", catErr)
	}
	// SPEC-USER-ASSET-INSTALL-001 (M7): no deploy-mode options — the
	// deployer carries a single project payload shape.
	switch config.ReadHarness(".") {
	case "gpt":
		return template.NewCodexOnlyDeployerWithRendererAndForceUpdate(cat, renderer)
	case "both":
		return template.NewDualHarnessDeployerWithRendererAndForceUpdate(cat, renderer)
	default:
		return template.NewClaudeHarnessDeployerWithRendererAndForceUpdate(cat, renderer)
	}
}

// resolveUpdateDeployMode resolves the run's deploy mode (REQ-016, design
// §3 step 4's last paragraph): a recorded plugin project deploys thin, a
// recorded local project today's full set, and a record-less project (the
// migration path) deploys thin unless the opt-out is set — both migration
// deploy arms (confirmed and not-demonstrated) leave the dropped roots
// untouched, and the opted-out arm deploys the full local payload.
func resolveUpdateDeployMode(projectRoot string, noPlugin bool) template.DeployMode {
	switch config.ReadDeployMode(projectRoot) {
	case "local":
		return template.DeployModeLocal
	}
	// SPEC-USER-ASSET-INSTALL-001 (M7): every arm resolves LOCAL — the
	// plugin payload is retired with its carrier.
	return template.DeployModeLocal
}

// mergeableFilePaths is the fixed mergeable-file set — the files merged with
// the 3-way merge engine after the deploy (NOT handled by restoreMoaiConfig,
// which owns .moai/config/sections/*.yaml). Package-level so the
// reconciliation merge-phase exclusion (update_reconcile.go) shares the one
// definition (card t1547 review finding 2; the 429-interrupted refactor
// moved it out of the runTemplateSyncWithReporter closure).
//
// .mcp.json IS shipped (internal/template/templates/.mcp.json, deployed by
// deployer.go's no-dotfile-skip WalkDir) and IS a 3-way merge target so a
// user's own MCP entries survive `moai update` instead of being clobbered
// by the template deploy (SPEC-MCP-DEFAULT-ON-001 REQ-A-4 / AC-A-012).
func mergeableFilePaths() []string {
	return []string{
		".claude/settings.json",
		".moai/status_line.sh",
		".mcp.json",
	}
}

// runTemplateSync synchronizes embedded templates with the project directory.
// It performs a quick version comparison first - if the project's template version
// matches the package version, the sync is skipped for performance (70-80% faster).
//
// Template deployment uses a 3-way merge strategy to preserve local modifications.
// Users are prompted to confirm the merge before proceeding.
func runTemplateSync(cmd *cobra.Command) error {
	return runTemplateSyncWithReporter(cmd, nil, false, nil)
}

// managedRedeployCount derives the outcome-summary accounting from the
// template list the deployer reports: the rendered-target set (each entry
// stripped of its .tmpl suffix — the deployed path) for the removal
// accounting, and the count of MoAI-managed files the deploy writes.
//
// Each rendered deployment target counts ONCE: a `.sh`/`.sh.tmpl` deployment
// pair (both list entries converging on the same stripped target — the 4
// hook-wrapper pairs) is one deployed file, not two
// (SPEC-INIT-UPDATE-CONSISTENCY-001 REQ-ICU-003). The ListTemplates
// stripped-target contract itself is untouched — the dedupe lives in the
// counting source only (plan §G).
func managedRedeployCount(templateFiles []string) (managedRedeployed int, restoredSet map[string]bool) {
	restoredSet = make(map[string]bool, len(templateFiles))
	seenTargets := make(map[string]bool, len(templateFiles))
	for _, tmpl := range templateFiles {
		if before, ok := strings.CutSuffix(tmpl, ".tmpl"); ok {
			tmpl = before
		}
		target := filepath.ToSlash(tmpl)
		restoredSet[target] = true
		if seenTargets[target] {
			// A .tmpl entry whose stripped target was already counted via
			// its rendered sibling is the same deployed file.
			continue
		}
		seenTargets[target] = true
		if plan.IsMoaiManaged(target) {
			managedRedeployed++
		}
	}
	return managedRedeployed, restoredSet
}

// presentArchiveDriftRoots lists the archive-drift backup roots that exist
// under .moai/archive/skills (the v<ver>-drift-<stamp> directories
// archiveLegacySkills creates under --force), as project-root-relative slash
// paths. Read-only; used to diff which roots a run created
// (SPEC-INIT-UPDATE-CONSISTENCY-001 REQ-ICU-004).
func presentArchiveDriftRoots(projectRoot string) map[string]bool {
	pattern := filepath.ToSlash(filepath.Join(projectRoot, ".moai", "archive", "skills")) +
		"/" + archiveVersion + "-drift-*"
	matches, _ := filepath.Glob(pattern)
	set := make(map[string]bool, len(matches))
	for _, m := range matches {
		set[filepath.ToSlash(m)] = true
	}
	return set
}

// newArchiveDriftRoots returns the project-root-relative paths of the drift
// roots present now that were absent from before — the roots this run created.
func newArchiveDriftRoots(projectRoot string, before map[string]bool) []string {
	var created []string
	for path := range presentArchiveDriftRoots(projectRoot) {
		if !before[path] {
			display := path
			if rel, relErr := filepath.Rel(filepath.ToSlash(projectRoot), path); relErr == nil {
				display = rel
			}
			created = append(created, display)
		}
	}
	return created
}

// @MX:NOTE: [AUTO] runTemplateSyncWithReporter — M4-S4d-2 DDD migration. Top-level header/section/
// final-outcome lines are converted to tui.KV / tui.Section / tui.CheckLine / tui.Pill. Sub-step
// micro messages (\r-prefixed sym* helpers) are preserved because they drive the progress display.
// Design source: screens.jsx ScreenUpdate.
//
// runTemplateSyncWithReporter synchronizes templates with progress reporting.
// runTemplateSyncWithReporter drives the update step table.
// migrationRemoved carries the paths the per-file user-asset migration
// removed before this call (gate round 12) — the outcome's deletion tally
// counts them with their own content-safe disposition, since the
// reconciliation can no longer see an already-removed file.
func runTemplateSyncWithReporter(cmd *cobra.Command, reporter project.ProgressReporter, skipConfirm bool, migrationRemoved []string) error {
	out := cmd.OutOrStdout()
	// The mirror notice is a warning, so it gets its own stderr writer rather
	// than riding `out` — which is stdout here (internal/cli/CLAUDE.md:14).
	errOut := cmd.ErrOrStderr()
	th := resolveTheme()
	ctx := cmd.Context()

	// Get flags for template sync
	forceBackup := getBoolFlag(cmd, "force")
	autoConfirm := getBoolFlag(cmd, "yes")
	// SPEC-INIT-SHRINK-001: the deploy mode and the opt-out resolve before
	// the deployer is constructed; the migration trigger (below) reads the
	// same opt-out.
	updateNoPlugin := getBoolFlag(cmd, "no-plugin") || updatePluginOptedOut()
	deployMode := resolveUpdateDeployMode(".", updateNoPlugin)

	// REQ-018: a recorded project is mode-aware, and update never flips the
	// record — the switch surface is the init re-entry, named here.
	if record := config.ReadDeployMode("."); record != "" {
		_, _ = fmt.Fprintf(out, "deploy mode: %s (to switch, re-run moai init — with --no-plugin for local, without for plugin)\n", record)
	}

	// Use current directory as project root
	projectRoot := "."

	currentVersion := version.GetVersion()
	// Identity header band (REQ-TUXIU-015): "◆ MoAI-ADK <version> <go-runtime>
	// · claude" with the version as a solid brand pill. Card t1527 D1: the band
	// is THE single version surface of a sync — the former duplicate
	// "Current version" KV beside it is gone.
	_, _ = fmt.Fprintln(out, renderIdentityBand(currentVersion, th))
	_, _ = fmt.Fprintln(out, tui.CheckLine("run", "Syncing templates", "from embedded filesystem", "", &th))

	// Card t1527 D2: the three pre-deploy phases render on the SAME stdout
	// ✓-line model as the deploy steps (tui.ProgressLine). They previously rode
	// the printerReporter — a second (stderr) surface whose ○/✓ pairs
	// interleaved with the stdout progress the operator was already reading.
	// The reporter parameter stays for error surfacing below; no production
	// caller drives step pairs through it any more.
	plVersion := tui.ProgressLine(out, "Checking template version...", nil)

	// Stage 2: Config Version Comparison (before template sync)
	// Compare package template_version with project config template_version
	// If versions match, skip sync for performance (70-80% faster)
	packageVersion := version.GetVersion()
	projectVersion, err := plan.GetProjectConfigVersion(projectRoot)
	if err == nil && packageVersion == projectVersion && !forceBackup {
		plVersion.Done("Already up-to-date")
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, ""), Theme: &th}))
		return nil
	}

	plVersion.Done("Version check complete")

	plLoad := tui.ProgressLine(out, "Loading embedded templates...", nil)

	// Load embedded templates
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		plLoad.Fail(fmt.Sprintf("Template load failed: %v", err))
		if reporter != nil {
			reporter.StepError(err)
		}
		return fmt.Errorf("load embedded templates: %w", err)
	}

	plLoad.Done("Templates loaded")

	plManifest := tui.ProgressLine(out, "Loading project manifest...", nil)

	// Initialize manifest manager
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		plManifest.Fail(fmt.Sprintf("Manifest load failed: %v", err))
		if reporter != nil {
			reporter.StepError(err)
		}
		return fmt.Errorf("load manifest: %w", err)
	}

	plManifest.Done("Manifest loaded")

	// Create deployer with renderer and force update enabled for template sync
	// This ensures template files are rendered (.tmpl -> actual file) and updated even if they exist
	deployer, err := newTemplateSyncDeployer(embedded, deployMode)
	if err != nil {
		return fmt.Errorf("construct harness deployer: %w", err)
	}

	// t40 defect 2: AnalyzeFiles skips IsMoaiManaged paths, so analysis.Files
	// carries only the merged/added files. Count the managed re-deployments
	// the deploy writes anyway (and keep the rendered-target set for the
	// removal accounting) so the outcome summary reports the real total.
	templateFiles := deployer.ListTemplates()
	managedRedeployed, restoredSet := managedRedeployCount(templateFiles)

	// Analyze merge changes. The "Analyzing merge changes" header + the
	// classification card are shown only when this function owns the
	// confirmation flow (skipConfirm=false). When skipConfirm=true the caller
	// (runTemplateSyncWithProgress) already printed both before the user's
	// y/n confirmation — reprinting them here duplicates the visible output.
	analysis := updatemerge.AnalyzeMergeChanges(deployer, projectRoot)

	// Card t1527 D3: derive the class counts ONCE for the whole flow — the
	// pre-confirm card and the end-of-run outcome breakdown read the same
	// numbers, so a --yes run (which skips the card) still carries the
	// add/update/conflict summary to the outcome.
	addCount, updateCount, conflictCount := classifyUpdateCounts(analysis.Files)

	if !skipConfirm {
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, tui.Section("Analyzing merge changes", tui.SectionOpts{Theme: &th}))
		// Card-style classification summary (REQ-TUXIU-010/011): accent box with
		// up to three count pills; zero-count pills omitted; suppressed entirely
		// when the run is clean (all counts zero).
		if card := renderClassificationSummary(addCount, updateCount, conflictCount, th); card != "" {
			_, _ = fmt.Fprintln(out, card)
		}
	}

	// Card t1527 D2: the found-count rides the same stdout model (the former
	// reporter.StepUpdate printed it on the interleaved stderr surface).
	_, _ = fmt.Fprintf(out, "%s Found %d files to sync\n",
		paintToken(tui.StatusIcon("info"), th.Faint, false), len(analysis.Files))

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

	// SPEC-INIT-SHRINK-001 (REQ-015, OD-4 settled (a, amended)): on a
	// record-less project the migration trigger runs AFTER the user
	// confirmed (the install step may touch the network) and BEFORE the step
	// table. Its plan decides what the Clean step may remove and what the
	// record reads at the end of the run.
	var migration *migrationPlan
	if config.ReadDeployMode(projectRoot) == "" {
		plan, migErr := runUpdateMigrationTrigger(projectRoot, updateNoPlugin, out, errOut)
		if migErr != nil {
			// The abort-before-removal contract: nothing was removed, the
			// record is unwritten, the next update re-triggers.
			return fmt.Errorf("migration: %w", migErr)
		}
		migration = plan
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

	// SPEC-INIT-UPDATE-CONSISTENCY-001 REQ-ICU-004: the namespace backup root
	// and the archive-drift roots created this run, so the outcome summary can
	// name every backup root (record-only roots — no consolidation).
	var nsBackupPath string
	var nsBackupDisplay string
	var archiveDriftRootsCreated []string

	// t40 defect 2: read-only snapshot of the managed roots, taken inside the
	// Clean step immediately before the removal, so the outcome summary can
	// account for what was deleted (and what the templates do not restore).
	var preCleanFiles []string

	// SPEC-UPDATE-MIGRATION-001 (card t1547): the preservation pipeline's
	// state, produced by the Clean step's reconcile phase and consumed by the
	// Restore step's merge phase. Empty on the legacy fresh-install branch.
	var reconSummary update.ReconciliationSummary
	var reconPending []update.PendingMerge

	// The Clean Managed Paths step removes .moai/config before Deploy Templates
	// renders, so the project's git mode is read here, while the file exists.
	// Without it every render falls back to the template default (manual).
	gitMode := config.LoadGitMode(projectRoot)
	// Cards t1139 / t1147: the same holds for the user-owned values the other
	// section files render (names, languages, development mode, git provider).
	// A value the render cannot carry verbatim falls back to the default (see
	// loadUpdateUserValues for when the merge then keeps it).
	userValues := loadUpdateUserValues(projectRoot)

	// SPEC-UPDATE-MIGRATION-001 (card t1547, REQ-UPM-040): computed once —
	// the Clean step branches on it and the outcome accounting differs
	// between the legacy wholesale branch and the default reconcile path.
	legacyClean := legacyFreshConfigInstall(projectRoot)

	// collectMergeableFiles returns a list of files that should be merged
	// using the 3-way merge engine during update.
	// Note: .moai/config/sections/*.yaml files are already handled by
	// restoreMoaiConfig with 3-way merge, so they are excluded here.
	// (Moved above the step table: the Clean step's reconcile closure reads
	// it for the merge-phase exclusion set — card t1547 review finding 2.)
	collectMergeableFiles := func(projectRoot string) []string {
		return mergeableFilePaths()
	}

	// Define deployment steps
	steps := []struct {
		name    string
		message string
		execute func() error
	}{
		{
			// Card t1527 D2: the loop intercepts "Backup" by name before any
			// execute runs (the interception owns config backup, namespace
			// backup, and the pre-merge snapshots) — this closure was dead
			// in-loop and is now gone. The message keeps the stage readable if
			// it is ever listed.
			name:    "Backup",
			message: "Backing up configuration",
			execute: nil,
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
					template.WithGitMode(gitMode),
					userValues,
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
				// Archive legacy skills (BC-V3R3-007) BEFORE the removal below:
				// they live under .claude/skills/moai*, which this step deletes,
				// so archiving afterwards finds no source. --force is propagated
				// so drift routes through the overwrite + backup path
				// (SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001 REQ-UAC-002). An archive
				// failure warns and does not stop the update.
				legacyBefore := presentLegacySkillIDs(projectRoot)
				// REQ-ICU-004: snapshot the archive-drift roots before the
				// archive step so only roots THIS run created reach the
				// outcome summary.
				driftBefore := presentArchiveDriftRoots(projectRoot)
				archived, archiveErr := archiveLegacySkills(projectRoot, out, forceBackup)
				if archiveErr != nil {
					// Card t1527 D5 + repair round: ONE surface — unarchived
					// skills are removed by this step's cleanup, so the
					// terminal ACTION REQUIRED row is the failure's home.
					updateLedger.requiref(sevWarn, "legacy skill archive failed: %v", archiveErr)
				}
				archiveDriftRootsCreated = newArchiveDriftRoots(projectRoot, driftBefore)

				// SPEC-INIT-SHRINK-001 (REQ-011/REQ-013/REQ-016, design §3
				// step 4): the removal scope is the run's target list —
				//  a thin-mode run (plugin deployer, migration included)
				//  excludes the dropped roots from the global walk, because
				//  P-08's backup exemption assumes the deploy rewrites what
				//  Clean removes, and the thin deploy does not rewrite them;
				//  a confirmed migration APPENDS the classified removal list
				//  (identical + archived modified), the only removal the
				//  dropped roots ever see;
				//  a local-mode run keeps today's global walk (the documented
				//  REQ-017 boundary — preservation is one migration run, not
				//  a byte-for-byte promise for every future update).
				// The list comes from computeRunCleanTargets — the exact
				// computation the --dry-run preview shares (card t1438
				// review finding 4), so the preview can never announce a
				// removal the run does not make.
				cleanTargets := computeRunCleanTargets(projectRoot, deployMode, migration)

				switch {
				case legacyClean:
					// t40 defect 2: snapshot what exists under THIS run's
					// wholesale scope BEFORE the removal (read-only; the
					// accounting matches the removal scope).
					preCleanFiles = deploy.InventoryManagedPathsWithTargets(projectRoot, cleanTargets)
					// A skill present now but not archived is deleted by the
					// wholesale removal below, so the shortfall is reported as
					// a loss — this branch only.
					reportArchiveShortfall(legacyBefore, archived, out)
					plLegacy := tui.ProgressLine(out, "Checking config compatibility...", nil)
					plLegacy.Done("Legacy config detected: fresh config install required (full backup already taken)")
					// SPEC-UPDATE-MIGRATION-001 (card t1547, REQ-UPM-040): the
					// legacy v1→v2 fresh-install branch keeps the wholesale
					// walk behind the REQ-UPM-015 guard.
					if err := guardFirstDestructiveStep(projectRoot, configBackupPath, func() error {
						tmplFS, tmplErr := template.EmbeddedTemplates()
						if tmplErr != nil {
							// Without the template FS the removal cannot tell
							// managed from unmanaged files, so it cannot back
							// anything up — abort rather than delete blind.
							return fmt.Errorf("load embedded templates: %w", tmplErr)
						}
						// REQ-UPM-015: even the fresh-install walk refuses
						// user-owned namespace and unresolved user-modified
						// files.
						protect := update.ProtectFuncFor(projectRoot, mgr.Manifest())
						return deploy.CleanMoaiManagedPathsWithTargetsGuarded(projectRoot, out, tmplFS, cleanTargets, protect)
					}); err != nil {
						return err
					}

				case migration != nil:
					// SPEC-INIT-SHRINK-001 (REQ-011, design §3 step 4): the
					// migration's WHOLESALE removal is exactly its classified
					// dropped-root list (identical + already-archived
					// modified) — the unguarded form is correct because that
					// list IS classified by the migration classifier. The
					// MANAGED roots no longer ride the wholesale walk (gate
					// round 10, finding 1): a record-less --no-plugin run
					// must not delete the local files the dry-run preview
					// promised to preserve — they reconcile like every
					// default run.
					preCleanFiles = deploy.InventoryManagedPathsWithTargets(projectRoot, migration.removalTargets)
					reportArchiveShortfall(legacyBefore, archived, out)
					if err := guardFirstDestructiveStep(projectRoot, configBackupPath, func() error {
						tmplFS, tmplErr := template.EmbeddedTemplates()
						if tmplErr != nil {
							return fmt.Errorf("load embedded templates: %w", tmplErr)
						}
						if len(migration.removalTargets) > 0 {
							if err := deploy.CleanMoaiManagedPathsWithTargets(projectRoot, out, tmplFS, migration.removalTargets); err != nil {
								return err
							}
						}
						return reconcileManagedRoots(projectRoot, out, tmplFS, mgr, deployMode, &reconSummary, &reconPending)
					}); err != nil {
						return err
					}

				default:
					// t40 defect 2: snapshot what exists under THIS run's
					// target list BEFORE the removal (read-only; the
					// accounting matches the removal scope).
					preCleanFiles = deploy.InventoryManagedPathsWithTargets(projectRoot, cleanTargets)
					// SPEC-UPDATE-DATA-SURVIVAL-001 REQ-UDS-001/003/005: the
					// three in-memory-only files reach disk before this step
					// removes anything. A backup-write failure aborts here,
					// so the removal never runs while a file's only copy is
					// in the heap. Card t111: the embedded FS rides along so
					// the removal can back up every file the template does
					// not carry before deleting the root it lives under.
					if err := guardFirstDestructiveStep(projectRoot, configBackupPath, func() error {
						tmplFS, tmplErr := template.EmbeddedTemplates()
						if tmplErr != nil {
							return fmt.Errorf("load embedded templates: %w", tmplErr)
						}
						return reconcileManagedRoots(projectRoot, out, tmplFS, mgr, deployMode, &reconSummary, &reconPending)
					}); err != nil {
						return err
					}
				}
				return nil
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
					template.WithGitMode(gitMode),
					userValues,
				)

				if deployErr := deployWithMirrorNotice(ctx, deployer, projectRoot, mgr, tmplCtx, errOut); deployErr != nil {
					pl.Fail(fmt.Sprintf("Deployment failed: %v", deployErr))
					return fmt.Errorf("deploy templates: %w", deployErr)
				}
				// SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001 (REQ-USB-001): stage the
				// settings.json render this deploy wrote, before the Restore
				// Settings merge rewrites the file. Best-effort: it only warns.
				backup.StageDeployedSettingsSnapshot(projectRoot, mgr, errOut)
				// Card t1029: the same staging for .mcp.json, which this flow
				// merges (see collectMergeableFiles below). Without it the merge
				// derives its base from the new render, and a template value
				// change to a server the user never touched stays invisible
				// while a newly added server still arrives — an update that
				// looks successful and is half-applied.
				backup.StageDeployedMCPSnapshot(projectRoot, mgr, errOut)
				// Card t1139: record the section render this deploy wrote as the
				// next update's merge BASE — here, before Restore Settings writes
				// the user's values over it. A snapshot taken after the restore
				// records the user's own values as BASE, and the next merge reads
				// every carried customization as "unchanged" and drops it.
				// This run's BASE was already copied into the backup, so the
				// write cannot affect the merge below. Best-effort non-blocking.
				writeTemplateSnapshotBestEffort(projectRoot, errOut)
				// card t1275: Deploy tracks every written file in the in-memory
				// manifest, but nothing in the update flow persisted it (init's
				// initializer calls Save; update never did) — so the on-disk
				// manifest kept pre-update hashes and the next `init --force`
				// reclassified every content-changed file user_modified. Persist
				// the deploy's tracking here, before the merge/restore steps
				// rewrite their files (those retrack separately below).
				if saveErr := mgr.Save(); saveErr != nil {
					_, _ = fmt.Fprintf(errOut, "  manifest save after deploy: %v\n", saveErr)
				}
				pl.Done("Templates deployed")
				return nil
			},
		},
		{
			// Card t1527 D2: like "Backup", the loop intercepts "Restore
			// Settings" by name — the interception owns the restore + merge
			// + retrack work, so the no-op closure is gone.
			name:    "Restore Settings",
			message: "Restoring user settings",
			execute: nil,
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

	// Execute each step with progress reporting. The per-step in-flight line is
	// rendered by the inline tui.ProgressLine primitive inside each step body;
	// the coarse reporter Step wrapper is intentionally NOT driven here — doing
	// so double-rendered every step on the stderr channel (the printer.Step
	// in-flight line) alongside the stdout ProgressLine, producing the two-part
	// "○…○" spinner residue on a TTY (REQ-TUXIU-020/021). Errors still surface
	// via reporter.StepError (orphan-safe) below.
	for _, step := range steps {
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
			var nsBackupErr error
			nsBackupPath, nsBackupErr = backupUserOwnedNamespace(projectRoot)
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
				nsBackupDisplay = displayNs
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
				// same stdout channel as the progress line below. The skip
				// filter excludes the reconciliation's archived-removed
				// section files (gate round 10, finding 3): the pipeline
				// archived and removed them, and a restore that re-created
				// them from the backup would contradict the reported removal.
				archivedSet := make(map[string]bool, len(reconSummary.ArchivedRemoved))
				for _, a := range reconSummary.ArchivedRemoved {
					if rel, ok := strings.CutPrefix(a, ".moai/config/sections/"); ok {
						archivedSet[rel] = true
					}
				}
				// Gate round 23, finding 2 (P2): preserved section files are
				// excluded too — a preserved local-only file was reported
				// untouched, and the restore's merge rewrote it with normalized
				// YAML bytes (4-space indent collapsed), the same contradiction
				// as the archived arm, on the preserve side (REQ-UPM-002).
				preservedSet := make(map[string]bool, len(reconSummary.Preserved))
				for _, p := range reconSummary.Preserved {
					if rel, ok := strings.CutPrefix(p, ".moai/config/sections/"); ok {
						preservedSet[rel] = true
					}
				}
				retainedKeys, restoreErr := backup.RestoreMoaiConfigRetained(projectRoot, configBackupPath, func(pr, relPath string, success bool, errOut io.Writer) {
					// Bridge to the noise-suppression ledger (recordMergeFallback +
					// updateVerboseMode), which stays in package cli. The closure
					// captures updateVerboseMode so the backup subpackage does not
					// need a cross-package mutable-state seam.
					recordMergeFallback(pr, relPath, success, updateVerboseMode, errOut)
				}, func(relPath string) bool { return archivedSet[relPath] || preservedSet[relPath] })
				if restoreErr != nil {
					plRestore.Fail(fmt.Sprintf("Restore failed: %v", restoreErr))
					if reporter != nil {
						reporter.StepError(restoreErr)
					}
					return recovery.fail(step.name, restoreErr)
				}
				plRestore.Done("User settings restored")
				// Strip the retired per-agent model/effort keys the merge just
				// retained (old-only keys are kept), after the backup holds their
				// originals. Stripped keys leave the retained list so each is
				// reported once, as removed. A strip failure warns; the restore
				// itself already succeeded.
				removedModelKeys, stripErr := stripRetiredModelConfig(out, projectRoot, configBackupPath)
				if stripErr != nil {
					// Card t1527 repair round: one surface (terminal row).
					updateLedger.requiref(sevWarn, "retired model key removal failed: %v", stripErr)
				}
				retainedKeys = withoutStrippedKeys(retainedKeys, removedModelKeys)
				// t63: one summary line by default; the key list expands only
				// under --verbose (the same verbose ledger recordMergeFallback
				// reads), never interleaving with the progress redraw.
				renderRetainedKeyAdvisory(out, retainedKeys, updateVerboseMode, th)
				// SPEC-INIT-HARNESS-001 (REQ-IH-002/010): re-assert the harness
				// value the PRE-UPDATE config carried. The deploy just rewrote
				// llm.yaml with the template default (claude), and whatever the
				// merge decided, the resolved selection must survive explicitly —
				// doctor and the next update read this key, not an inference.
				// Best-effort: a read failure degrades to claude the same way an
				// absent key does.
				if harness := config.ReadHarnessFrom(filepath.Join(configBackupPath, "sections")); harness != "" {
					if err := template.ApplyHarness(projectRoot, harness); err != nil {
						// Card t1527 repair round: one surface (terminal row).
						updateLedger.requiref(sevWarn, "llm.harness re-assert failed: %v", err)
					}
				}
				// SPEC-INIT-SHRINK-001 (REQ-009, OD-5 settled condition): the
				// same re-assert for the deploy-mode record — the Clean step's
				// .moai/config wipe must not cost the key, and the deploy just
				// rewrote llm.yaml from the template (which carries no record).
				if recorded := config.ReadDeployModeFrom(filepath.Join(configBackupPath, "sections")); recorded != "" {
					if err := template.ApplyDeployMode(projectRoot, recorded); err != nil {
						_, _ = fmt.Fprintf(out, "  %s deployment_mode re-assert warning: %v\n", uikit.SymWarning(), err)
					}
				}
				// card t1275: RestoreMoaiConfigRetained + ApplyHarness just
				// rewrote .moai/config/sections/*.yaml on top of the deployed
				// render — re-record those hashes so the manifest matches what
				// this update actually left on disk. Card t1527 repair round 3:
				// the failure now returns; the ! line keeps the immediate
				// surface the internal print used to own.
				if retrackErr := retrackSectionFiles(projectRoot, errOut); retrackErr != nil {
					emitSeverityLine(errOut, sevWarn, resolveTheme(), "manifest retrack (config sections) failed: %v", retrackErr)
				}
				deletedCount := backup.CleanupOldBackups(projectRoot, 5)
				if deletedCount > 0 {
					_, _ = fmt.Fprintf(out, "  %s Cleaned up %d old backup(s)\n", uikit.SymSuccess(), deletedCount)
				}
			}
			// Merge .gitignore: preserve user-added patterns via EntryMerge
			if len(gitignoreBackup) > 0 {
				gitignorePath := filepath.Join(projectRoot, ".gitignore")
				if mergeErr := updatemerge.MergeGitignoreFile(gitignorePath, gitignoreBackup); mergeErr != nil {
					// Card t1527 repair round: one surface (terminal row).
					updateLedger.requiref(sevWarn, ".gitignore merge failed: %v", mergeErr)
				} else {
					_, _ = fmt.Fprintf(out, "  %s .gitignore user patterns preserved\n", uikit.SymSuccess())
					// card t1276 F1 (leader-approved option A): the EntryMerge
					// rewrote .gitignore as template + user entries AFTER the
					// deploy tracked its render — re-record the merged output
					// so the manifest matches the tree. Must go through the
					// SAME in-memory mgr the later mergeable retrack saves:
					// a fresh manager would save the new hash here and then
					// have it overwritten by the stale entry that mgr still
					// carries. The merge itself is what preserves the user's
					// entries, and a user who edits .gitignore outside any
					// update is not in this path, so their drift still reads
					// user_modified (two-way).
					if retrackErr := retrackManifestFiles(projectRoot, mgr, errOut, []string{".gitignore"}); retrackErr != nil {
						_, _ = fmt.Fprintf(errOut, "  manifest retrack (.gitignore): %v\n", retrackErr)
					}
				}
			}
			// Merge user-customized files using 3-way merge engine, then settle
			// the staged settings.json render (SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001
			// REQ-USB-005). Deliberately outside the configBackupPath block and
			// run even with no backups: the promotion decision belongs to every
			// flow that deployed (plan.md D4 ③, M-07d).
			if err := mergeUserFilesSettlingSnapshot(projectRoot, mergeableBackups, out, errOut); err != nil {
				// Card t1527 repair round: one surface (terminal row).
				updateLedger.requiref(sevWarn, "mergeable-file merge failed: %v (a pre-update copy is in the run backup)", err)
			}
			// card t1275: the 3-way merge rewrote the mergeable set
			// (.claude/settings.json, .moai/status_line.sh, .mcp.json, ...)
			// after the deploy already tracked the fresh render — re-record
			// the merged result for exactly those paths. user-owned files are
			// filtered out inside retrackManifestFiles, so a file the user
			// alone changed keeps its drift and still reads user_modified.
			if retrackErr := retrackManifestFiles(projectRoot, mgr, errOut, collectMergeableFiles(projectRoot)); retrackErr != nil {
				_, _ = fmt.Fprintf(errOut, "  manifest retrack (mergeable set): %v\n", retrackErr)
			}
			// SPEC-UPDATE-MIGRATION-001 (card t1547): the reconciliation merge
			// phase — the deploy rewrote every template-carried path, so each
			// user-modified file captured by the reconcile phase is now merged
			// (clean) or conflicted (operator bytes restored + .moai-new
			// sidecar) against the fresh render, with the deploy-time snapshot
			// as the base where one exists. Runs after the config restore so
			// already-converged section files cost nothing. Failures here are
			// partial-update failures and ride the recovery guard like the
			// config restore above (REQ-UDS-019).
			if len(reconPending) > 0 {
				plRecon := tui.ProgressLine(out, "Reconciling local modifications...", nil)
				var reconErr error
				reconSummary, reconErr = update.ReconcileMerges(projectRoot, out,
					nil, mgr.Manifest(),
					update.ReconcileMergeOptions{Base: update.SnapshotBaseSource(projectRoot)},
					reconPending, reconSummary)
				if reconErr != nil {
					plRecon.Fail(fmt.Sprintf("Reconciliation failed: %v", reconErr))
					if reporter != nil {
						reporter.StepError(reconErr)
					}
					return recovery.fail(step.name, reconErr)
				}
				plRecon.Done(fmt.Sprintf("%d merged, %d conflict(s), %d preserved",
					len(reconSummary.Merged), len(reconSummary.Conflicts), len(reconSummary.Preserved)))
				// The merges rewrote files the deploy just tracked — re-record
				// the MERGED set (the t1275 pattern). Conflict-preserved paths
				// are deliberately EXCLUDED (gate round 9, finding 1): their
				// on-disk bytes are the OPERATOR's restored content, not this
				// run's intended state — re-tracking one would register the
				// user's bytes as a healthy template_managed record, and the
				// next update would classify the file template-owned and
				// overwrite it with the render (the R-2 recurrence path). An
				// untouched record keeps the divergent hash, so the next run
				// classifies the file user-modified and reports the conflict
				// again instead of silently overwriting.
				if len(reconSummary.Merged) > 0 {
					if retrackErr := retrackManifestFiles(projectRoot, mgr, errOut, reconSummary.Merged); retrackErr != nil {
						_, _ = fmt.Fprintf(errOut, "  manifest retrack (reconciled set): %v\n", retrackErr)
					}
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
	}

	// Card t1527 D2: the block progress bar (REQ-TUXIU-014) renders ONCE at
	// completion. The former per-step render printed five intermediate bars
	// ("1/5"…"5/5") that read as the accumulated-snapshot breakage t694 already
	// fixed inside the bar glyph — one completed bar is the whole story.
	_, _ = fmt.Fprintln(out, renderDeployProgress(len(steps), len(steps), th))

	_, _ = fmt.Fprintln(out)
	// Outcome banner (REQ-TUXIU-016): solid success pill + dim detail note.
	// t40 defect 2: the pill total includes the managed re-deployments and
	// the note carries the removal accounting (local-only losses named by
	// count).
	detail := updateOutcomeDetail{
		ManagedRedeployed:   managedRedeployed,
		NamespaceBackupPath: nsBackupDisplay,
		ArchiveDriftRoots:   archiveDriftRootsCreated,
		AddFiles:            addCount,
		UpdatedFiles:        updateCount,
		ConflictFiles:       conflictCount,
	}
	if legacyClean || migration != nil {
		// Wholesale branches (legacy fresh-install, init-shrink migration):
		// the walk removed what its target list covered — the t40 accounting
		// applies unchanged.
		for _, f := range preCleanFiles {
			detail.RemovedManaged++
			if !restoredSet[f] {
				detail.RemovedLocalOnly++
			}
		}
		if migration != nil {
			// Gate round 8 (card t1547 repair round): the record-less
			// migration arm carries NO removal targets of its own — the
			// reconcile the same branch drives is what archives-then-removes
			// the stale managed set, and that removal never reached this
			// tally (the wholesale inventory above is empty on that arm).
			// REQ-UPM-030/031: a run that removed N managed files says so,
			// whatever branch removed them.
			detail.RemovedManaged += len(reconSummary.ArchivedRemoved)
			detail.ArchivedForRecovery += len(reconSummary.ArchivedRemoved)
		}
	} else {
		// SPEC-UPDATE-MIGRATION-001 (card t1547, review finding 6): the
		// default path removed NOTHING wholesale — its only removals are the
		// reconciler's archive-then-remove stale set, and every one of those
		// carries a recovery copy. Counting the pre-clean snapshot as
		// deletions reported preserved files as removed; the accounting now
		// reads the reconciliation's actual dispositions (REQ-UPM-031:
		// deletions reported, never a deletion-free claim over removals).
		detail.RemovedManaged = len(reconSummary.ArchivedRemoved)
		// Gate round 21: the recovery copies are named as exactly that —
		// the breakdown must not claim "all re-deployed" over files that are
		// gone from place and only recoverable from the archive.
		detail.ArchivedForRecovery = len(reconSummary.ArchivedRemoved)
		// RemovedLocalOnly stays 0: an archived removal is recoverable by
		// definition, and preserved files were never removed.
	}
	// Gate round 12 (card t1547): the asset migration's confirmed-counterpart
	// removals are deletions too — the tally carries them with their own
	// content-safe disposition (the user copy holds the bytes), on every arm:
	// the migration runs after the confirm gate regardless of branch.
	detail.MigrationRemoved = len(migrationRemoved)
	detail.RemovedManaged += len(migrationRemoved)
	renderUpdateOutcome(out, len(analysis.Files), detail, configBackupPath, th)
	// SPEC-UPDATE-MIGRATION-001 (card t1547, REQ-UPM-030/031): the
	// reconciliation outcome rides the existing report structures — plain
	// counts and per-path lists, every deletion named.
	renderReconciliationOutcome(out, reconSummary, th)

	// Card t1527 D5, repaired by gate round 4 (card t1547): the ACTION
	// REQUIRED row keys on the ACTUAL unresolved conflicts — the
	// reconciliation's conflict list (files preserved byte-for-byte with a
	// .moai-new sidecar) — never on the pre-merge risk classification:
	// conflictCount counts RiskLevel=="high" files, and a fresh project's
	// high-risk AGENTS.md/settings.json merge clean, so the risk-derived
	// count cried wolf on every such run.
	if unresolved := len(reconSummary.Conflicts); unresolved > 0 {
		updateLedger.requiref(sevErr, "%d conflicting file(s) flagged by the 3-way merge — resolve them before the next template sync (see the backup at %s)", unresolved, configBackupPath)
	}

	// REQ-DHR-007: a .codex/ template the target harness profile (or this
	// version) no longer ships is reported and left in place, never deleted.
	reportUndeployedCodexTemplates(errOut, projectRoot, mgr.Manifest().Files, restoredSet)

	// Card t1527 D5: the tail's mid-noise emitters move into the terminal
	// block — the hook-restart step is ACTION REQUIRED, the -c hint and the
	// worktree advisory are Reference one-liners. The block renders at the end
	// of runUpdate, after the post-sync steps.
	updateLedger.requiref(sevWarn, "%s", hooksReviewGuidanceMsg())

	updateLedger.referencef("Reconfigure project settings (development mode, git, model policy): moai update -c")

	// Ensure global settings.json has required env variables
	if err := ensureGlobalSettingsEnv(); err != nil {
		// Card t1527 repair round: one surface (terminal row).
		updateLedger.requiref(sevWarn, "global settings env update failed: %v", err)
	}

	// Install pre-push hook (REQ-CIAUT-002). Non-fatal; --no-hooks opts out.
	// Card t1527 D5: a failed install is escalated into the terminal block.
	if pushErr := installPrePushHookOptional(projectRoot, getBoolFlag(cmd, "no-hooks"), out, errOut); pushErr != nil {
		updateLedger.requiref(sevErr, "pre-push hook installation failed: %v", pushErr)
	}

	// Install pre-commit hook (REQ-PC-001). Fast-subset commit tier; --no-hooks opts out.
	if commitErr := installPreCommitHookOptional(projectRoot, getBoolFlag(cmd, "no-hooks"), out, errOut); commitErr != nil {
		updateLedger.requiref(sevErr, "pre-commit hook installation failed: %v", commitErr)
	}

	// SPEC-WORKTREE-BRANCH-GUARD-001 (REQ-WBG-009): one-line worktree advisory on
	// update completion. The primary checkout is shared; branch-changing work
	// belongs in a worktree. Card t1527 D5: routed into the Reference section
	// (the AC-WBG-009 wording travels with it verbatim).
	updateLedger.referencef("%s", worktreeAdvisoryText(projectRoot))

	// SPEC-INIT-SHRINK-001 (REQ-015, design §3 step 5): a completed
	// migration run writes the record — plugin after a confirmed install,
	// local under every other arm. A record-bearing project is unchanged:
	// update never flips the record (REQ-018). A failed run lands here never
	// (the error paths above return first), so the record is written only
	// over a completed sync.
	if migration != nil {
		switch migration.outcome {
		case migrateConfirmed:
			if err := template.ApplyDeployMode(projectRoot, "plugin"); err != nil {
				_, _ = fmt.Fprintf(errOut, "  %s deployment_mode write warning: %v\n", uikit.SymWarning(), err)
			}
		case migrateNotDemonstrated, migrateOptedOut:
			if err := template.ApplyDeployMode(projectRoot, "local"); err != nil {
				_, _ = fmt.Fprintf(errOut, "  %s deployment_mode write warning: %v\n", uikit.SymWarning(), err)
			}
		}
	}

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
// runTemplateSyncWithProgress is `moai update`'s sync entry: the
// version-match skip, the interactive confirmation (unless --yes), then the
// reporter-driven flow. userAssetsInstalled rides in from the user-asset
// phase (runUpdate) as the project migration's run-level removal gate — the
// migration itself runs BELOW, after this function's confirmation gate.
func runTemplateSyncWithProgress(cmd *cobra.Command, userAssetsInstalled bool) (skipped bool, err error) {
	out := cmd.OutOrStdout()
	th := resolveTheme()
	projectRoot := "."
	autoConfirm := getBoolFlag(cmd, "yes")
	forceUpdate := getBoolFlag(cmd, "force")

	// Card t1527 D2: pass nil — the sync renders its pre-deploy phases on the
	// same stdout ✓-line model as the deploy steps, and the former
	// printer-backed stderr phase pairs no longer interleave with it. (The
	// reporter parameter itself stays: it surfaces step errors for any future
	// spinner-backed caller, exactly as runInit does.)

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
		proceed, cerr := confirmViaPreviewFn(analysis, projectRoot)
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

	// Item 5 (fix round 3), relocated by the repair round (gate r5 finding):
	// the per-file project migration runs HERE — after the confirmation gate
	// above, whose cancel return already exited — so a cancelled update
	// leaves the project's managed assets byte-intact. The user-side install
	// (runUpdate, before this prompt) still precedes the removal, and
	// userAssetsInstalled gates every removal: a failed or skipped install
	// leaves the project-side assets in place (leader scope addition #5 —
	// the stale-counterpart deletion hazard). The removed paths ride into
	// the reporter so the outcome's deletion tally carries them (gate round
	// 12 — a removal that only bumped a counter vanished from every list).
	var migrationRemoved []string
	if homeDir, homeErr := userHomeDirFn(); homeErr == nil {
		removed, err := migrateProjectCommonAssets(projectRoot, homeDir, userAssetsInstalled, nil, func(format string, args ...interface{}) {
			_, _ = fmt.Fprintf(out, format+"\n", args...)
		})
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai: migration warning: %v\n", err)
		}
		migrationRemoved = removed
	}

	return false, runTemplateSyncWithReporter(cmd, nil, true, migrationRemoved)
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
//
// runTemplateSyncWithProgress calls it through confirmViaPreviewFn, a seam a
// test replaces to answer "cancel" without a TTY.
func confirmViaPreview(analysis merge.MergeAnalysis, projectRoot string) (bool, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return false, fmt.Errorf("merge confirmation UI requires an interactive terminal; " +
			"rerun with --yes to auto-confirm in a non-TTY environment")
	}
	return update.PreviewClassification(toPreviewInputs(analysis, projectRoot), plan.IsUserOwnedNamespace, update.PreviewOptions{Interactive: true})
}
