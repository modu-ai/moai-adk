package cli

// update_migrate.go — the update-side migration of legacy (record-less)
// projects (SPEC-INIT-SHRINK-001 REQ-011..REQ-016, plan M3; OD-4 settled
// (a, amended), design §3).
//
// The trigger runs AFTER the user confirmed the update and BEFORE the step
// table: the install step (a potential network touch) never precedes the
// confirmation. Its verdict decides the run's shape:
//
//	confirmed         classify → archive modified per file → the classified
//	                  removal list rides the Clean step's target list; the
//	                  record is written plugin when the run completes.
//	not-demonstrated  nothing is removed, nothing redeploys into the dropped
//	                  roots (the run deploys thin); the record is written
//	                  local.
//	opted-out         the full local payload deploys through the normal local
//	                  path; the record is written local.
//
// The archive unit is the classified FILE, never a directory (D-17): a
// skill directory archives only its classified modified members, and any
// single archive-write failure aborts before ANY removal (OD-3 settled
// condition; the P-08 REQ-UDS-008 rule).

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// migrationOutcome is the trigger's decision for one update run. The zero
// value never constructs a plan: a record-bearing project skips the trigger
// entirely (the plan stays nil).
type migrationOutcome string

const (
	// migrateConfirmed — the probe demonstrated this-run success.
	migrateConfirmed migrationOutcome = "confirmed"
	// migrateNotDemonstrated — every other observable diff outcome.
	migrateNotDemonstrated migrationOutcome = "not-demonstrated"
	// migrateOptedOut — the opt-out was set before the step ran.
	migrateOptedOut migrationOutcome = "opted-out"
)

// migrationPlan carries the trigger's verdict and (on confirmed) the
// classified removal targets through the step table.
type migrationPlan struct {
	outcome migrationOutcome
	// removalTargets is exactly the classified set (identical plus the
	// modified set once every archive in the batch succeeded) — the list the
	// Clean step processes on a confirmed run. Nil for every other outcome.
	removalTargets []deploy.CleanTarget
	// counts are the three printed classification counts (REQ-010).
	identical, modified, foreign int
	// archivedModified counts the per-file archives written this run.
	archivedModified int
}

// updateDroppedRootTargets are the managed clean targets whose roots the
// thin deploy does not rewrite: they leave the global walk on every
// thin-mode run, because P-08's backup exemption assumes the deploy
// rewrites what it removes — a promise the thin deploy does not keep.
func updateDroppedRootTargets(targets []deploy.CleanTarget) []deploy.CleanTarget {
	kept := make([]deploy.CleanTarget, 0, len(targets))
	for _, t := range targets {
		norm := filepath.ToSlash(t.DisplayPath)
		if strings.HasPrefix(norm, ".claude/skills") || strings.HasPrefix(norm, ".claude/commands") {
			continue
		}
		kept = append(kept, t)
	}
	return kept
}

// computeRunCleanTargets is the run's Clean Managed Paths target list — the
// single computation the execution and the --dry-run preview share (card
// t1438 review finding 4): the global managed walk, minus the dropped roots
// on a thin (plugin-mode) run, plus the confirmed migration's classified
// removal list. A preview that computes this differently announces removals
// the run preserves.
func computeRunCleanTargets(projectRoot string, deployMode template.DeployMode, migration *migrationPlan) []deploy.CleanTarget {
	cleanTargets := deploy.ManagedCleanTargets(projectRoot)
	if deployMode == template.DeployModePlugin {
		cleanTargets = updateDroppedRootTargets(cleanTargets)
	}
	if migration != nil && migration.outcome == migrateConfirmed {
		cleanTargets = append(cleanTargets, migration.removalTargets...)
	}
	return cleanTargets
}

// runUpdateMigrationTrigger executes the migration trigger (design §3 step
// 1-3) for a projectRoot whose record is absent. It returns the plan for
// the run and, on a confirmed outcome, the classified removal list ready
// for the Clean step.
func runUpdateMigrationTrigger(projectRoot string, noPlugin bool, run pluginCommandRunner, out, errOut io.Writer) (*migrationPlan, error) {
	plan := &migrationPlan{}

	// Opted-out: decided before the step runs; no probing, full local path.
	if noPlugin {
		plan.outcome = migrateOptedOut
		return plan, nil
	}

	// The install step runs fail-open under the opt-out, its outcome read
	// through the post-install list-surface probe (design §2.4 — the
	// migration surface of the arm mapping).
	wiring := agentWiring(config.ReadHarness(projectRoot))
	if wiring != agentWiringGPT && wiring != agentWiringBoth {
		wiring = agentWiringClaude
	}
	opts := newPluginInstallOptions(pluginToolsForHarness(wiring), projectRoot, false)
	pre := snapshotPluginListSurfaces(opts.Tools, opts.ProjectRoot, run)
	_ = runPluginInstallStep(errOut, opts)
	switch diffPluginListSurfaces(pre, opts.Tools, opts.ProjectRoot, run) {
	case probeOutcomeConfirmed:
		plan.outcome = migrateConfirmed
	default:
		// not-demonstrated: nothing is deduped, nothing removed, and no
		// path records plugin (OD-4 settled (a, amended)).
		plan.outcome = migrateNotDemonstrated
		return plan, nil
	}

	// Classification (REQ-010) against the production render.
	mgr := manifest.NewManager()
	if _, err := mgr.Load(projectRoot); err != nil {
		return nil, fmt.Errorf("migration classify: load manifest: %w", err)
	}
	embedded, err := template.EmbeddedTemplates()
	if err != nil {
		return nil, fmt.Errorf("migration classify: load embedded templates: %w", err)
	}
	plan2, err := update.ClassifyMigration(projectRoot, templateRenderCarriage(embedded, nil, nil), mgr.Manifest())
	if err != nil {
		return nil, fmt.Errorf("migration classify: %w", err)
	}
	plan.identical = plan2.CountIdentical()
	plan.modified = plan2.CountModified()
	plan.foreign = plan2.CountForeign()

	// Archive (REQ-012) BEFORE any removal: every modified member archives
	// per file, and any archive-write failure aborts the whole migration.
	for _, f := range plan2.Modified {
		if err := archiveMigrationFile(projectRoot, f.RelPath); err != nil {
			return nil, fmt.Errorf("migration archive %s: %w", f.RelPath, err)
		}
		plan.archivedModified++
	}

	// Design §3 mirror paragraph (card t1438 review finding 2): the KEPT
	// mirror entries are re-homed to real directory copies rendered from the
	// embedded tree BEFORE the dropped-root removal runs — a kept symlink
	// would dangle into the removed .claude/skills. The re-home converts
	// existing links only (never provisions, REQ-019) and is best-effort
	// like every other mirror producer: a per-entry failure warns and the
	// migration continues (rehomeOneSkill's own convention).
	rehomedCount := 0
	for _, e := range template.RehomeExistingMirrorEntries(projectRoot, migrationTemplateContext(projectRoot)) {
		if e.Mode == template.MirrorModeFailed {
			_, _ = fmt.Fprintf(errOut, "warning: migration mirror re-home: %s\n", e.Warning)
			continue
		}
		if e.Mode == template.MirrorModeCopy {
			rehomedCount++
		}
	}
	if rehomedCount > 0 {
		_, _ = fmt.Fprintf(out, "migration: re-homed %d mirror %s to real copies\n",
			rehomedCount, pluralMirrorEntries(rehomedCount))
	}

	// The classified removal list (design §3 step 4): identical files, plus
	// the modified set now that every archive in the batch succeeded.
	for _, f := range plan2.Identical {
		plan.removalTargets = append(plan.removalTargets, classifiedCleanTarget(projectRoot, f.RelPath))
	}
	for _, f := range plan2.Modified {
		plan.removalTargets = append(plan.removalTargets, classifiedCleanTarget(projectRoot, f.RelPath))
	}

	_, _ = fmt.Fprintf(out, "migration: classified %d identical, %d modified, %d foreign; archived %d modified file(s)\n",
		plan.identical, plan.modified, plan.foreign, plan.archivedModified)
	return plan, nil
}

// classifiedCleanTarget builds one classified removal target.
func classifiedCleanTarget(projectRoot, relSlash string) deploy.CleanTarget {
	return deploy.CleanTarget{
		DisplayPath: relSlash,
		FullPath:    filepath.Join(projectRoot, filepath.FromSlash(relSlash)),
	}
}

// archiveMigrationFile archives ONE classified file (never a directory)
// into the migration archive layout (REQ-012; the archiveSkill contract's
// idempotency, drift-check, and symlink refusal applied at file scope):
//
//	.claude/skills/<skill>/<rest>  → .moai/archive/skills/<tag>/<skill>/<rest>
//	everything else                → .moai/archive/files/<tag>/<original path>
//
// Idempotent: an archive already holding identical bytes is a no-op; an
// archive holding different bytes is ARCHIVE_DRIFT (the user archived
// something of their own there — the migration never overwrites it).
func archiveMigrationFile(projectRoot, relSlash string) error {
	src := filepath.Join(projectRoot, filepath.FromSlash(relSlash))

	// REQ-013 / REQ-SEC-003: refuse a symlink source — the link itself is
	// never dereferenced into the archive.
	if info, err := os.Lstat(src); err != nil {
		if os.IsNotExist(err) {
			return nil // already gone: idempotent no-op
		}
		return fmt.Errorf("stat: %w", err)
	} else if info.Mode()&os.ModeSymlink != 0 {
		return &MigrateError{
			Code:    "ARCHIVE_SYMLINK",
			Message: fmt.Sprintf("refusing to archive symlink %s", relSlash),
		}
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	var dst string
	if rest, ok := strings.CutPrefix(relSlash, ".claude/skills/"); ok {
		dst = filepath.Join(projectRoot, update.ArchiveSkillFileRoot(), filepath.FromSlash(rest))
	} else {
		dst = filepath.Join(projectRoot, update.ArchiveFilesRoot(), filepath.FromSlash(relSlash))
	}

	// Card t1438 review finding 1: the write below must never travel THROUGH
	// a symlink — a planted link at any component of the archive path (the
	// destination itself, or a parent the MkdirAll below would otherwise
	// treat as an existing directory) would redirect the archive bytes
	// outside the archive. Any symlink on the way aborts this archive, which
	// aborts the whole migration before any removal (OD-3).
	if err := rejectSymlinkedArchivePath(projectRoot, dst); err != nil {
		return err
	}

	if existing, err := os.ReadFile(dst); err == nil {
		if string(existing) == string(data) {
			return nil // archive already holds the file: idempotent success
		}
		return &MigrateError{
			Code:    "ARCHIVE_DRIFT",
			Message: fmt.Sprintf("archive %s already exists with different content (the migration never overwrites an archive)", dst),
		}
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create archive directory: %w", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return fmt.Errorf("write archive: %w", err)
	}
	return nil
}

// rejectSymlinkedArchivePath refuses an archive destination whose path — the
// destination itself, or any parent component between it and the project
// root — is a symlink (card t1438 review finding 1; the REQ-SEC-003
// no-dereference rule applied to the archive's write side). The error names
// the symlinked component.
func rejectSymlinkedArchivePath(projectRoot, dst string) error {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("resolve project root: %w", err)
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("resolve archive path: %w", err)
	}
	cur := absDst
	for {
		if info, statErr := os.Lstat(cur); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			name := cur
			if rel, relErr := filepath.Rel(absRoot, cur); relErr == nil && !strings.HasPrefix(rel, "..") {
				name = filepath.ToSlash(rel)
			}
			return &MigrateError{
				Code:    "ARCHIVE_SYMLINK",
				Message: fmt.Sprintf("refusing to write through symlinked archive path %s", name),
			}
		}
		if cur == absRoot {
			return nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return nil // reached the filesystem root without meeting absRoot
		}
		cur = parent
	}
}

// migrationTemplateContext builds the render context the migration's mirror
// re-home writes its copies with — the same shape the update flow's deploy
// steps construct, so a re-homed .tmpl source renders identically to a
// deployed one.
func migrationTemplateContext(projectRoot string) *template.TemplateContext {
	homeDir, _ := userHomeDirFn()
	return template.NewTemplateContext(
		template.WithGoBinPath(detectGoBinPathForUpdate(homeDir)),
		template.WithResolvedMoaiPath(resolveMoaiExecutable()),
		template.WithHomeDir(homeDir),
		template.WithSmartPATH(template.BuildSmartPATH()),
		template.WithPlatform(runtime.GOOS),
		template.WithVersion(version.GetVersion()),
		template.WithHookOptIn(readHookOptInEnabled(projectRoot)),
		template.WithGitMode(config.LoadGitMode(projectRoot)),
		loadUpdateUserValues(projectRoot),
	)
}
