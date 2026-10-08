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
	"strings"

	"github.com/modu-ai/moai-adk/internal/cli/update"
	"github.com/modu-ai/moai-adk/internal/cli/update/deploy"
	"github.com/modu-ai/moai-adk/internal/template"
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
	// archivedModified counts the per-file archives written this run.
}

// updateDroppedRootTargets are the managed clean targets whose roots the
// thin deploy does not rewrite: they leave the global walk on every
// thin-mode run, because P-08's backup exemption assumes the deploy
// rewrites what it removes — a promise the thin deploy does not keep.

// computeRunCleanTargets is the run's Clean Managed Paths target list — the
// single computation the execution and the --dry-run preview share (card
// t1438 review finding 4): the global managed walk, minus the dropped roots
// on a thin (plugin-mode) run, plus the confirmed migration's classified
// removal list. A preview that computes this differently announces removals
// the run preserves.
func computeRunCleanTargets(projectRoot string, deployMode template.DeployMode, migration *migrationPlan) []deploy.CleanTarget {
	all := deploy.ManagedCleanTargets(projectRoot)
	// B1 (review-fix round 2 addendum): the WHOLESALE common-root deletion
	// is removed entirely — on a first update the preserve list is empty,
	// so the managed glob (.claude/skills/moai*, the agent dirs) deleted
	// user-modified copies and failed user-side installs before the
	// migration could protect them. The per-file REQ-020 migration
	// (migrateProjectCommonAssets) is the ONLY common-asset removal: it
	// classifies provenance, compares current bytes, and confirms the user
	// counterpart per file. Non-common-root managed targets (settings,
	// rules, hooks, output-styles, command wrappers) keep the managed walk.
	cleanTargets := make([]deploy.CleanTarget, 0, len(all))
	for _, ct := range all {
		if isCommonAssetCleanTarget(ct.DisplayPath) {
			continue
		}
		cleanTargets = append(cleanTargets, ct)
	}
	if migration != nil && migration.outcome == migrateConfirmed {
		cleanTargets = append(cleanTargets, migration.removalTargets...)
	}
	return cleanTargets
}

// isCommonAssetCleanTarget reports whether a managed-clean display path
// lives under a common-asset root (whose removal is now exclusively the
// migration's per-file job). The display path is normalized to slash form
// FIRST (repair round, Windows data-loss fix): ManagedCleanTargets builds
// DisplayPath with filepath.Join, so on Windows it carries backslashes while
// projectCommonAssetRels is slash-rooted — a raw prefix comparison missed,
// the exclusion silently opened the clean scope over preserved assets, and
// the cleanup deleted them. The normalization is host-independent (not
// filepath.ToSlash, which only rewrites the HOST separator): a Windows-shaped
// display path must classify as common-asset on every platform, and the
// rare Unix filename containing a literal backslash can only ever be
// SPARED by the rewrite (exclusion is the safe direction).
func isCommonAssetCleanTarget(displayPath string) bool {
	norm := strings.ReplaceAll(displayPath, "\\", "/")
	for _, root := range projectCommonAssetRels {
		if strings.HasPrefix(norm, root) || strings.HasPrefix(norm, strings.TrimSuffix(root, "/")) {
			return true
		}
	}
	return false
}

// migrationPreservedProjectFiles holds the absolute paths the REQ-020
// migration preserved (user-modified bytes or unverified counterpart) in
// THIS update run — the cleanup list drops them (review fix F2). Set by
// runUpdate's migration call, read by computeRunCleanTargets inside the
// template sync the same run drives; the two stages are sequential in one
// goroutine, so no lock is needed.
var migrationPreservedProjectFiles = map[string]bool{}

// runUpdateMigrationTrigger executes the migration trigger (design §3 step
// 1-3) for a projectRoot whose record is absent. It returns the plan for
// the run and, on a confirmed outcome, the classified removal list ready
// for the Clean step.
func runUpdateMigrationTrigger(projectRoot string, noPlugin bool, out, errOut io.Writer) (*migrationPlan, error) {
	plan := &migrationPlan{}

	// SPEC-USER-ASSET-INSTALL-001 (M6, REQ-017): the plugin install probe is
	// retired with its carrier — update invokes no marketplace or plugin
	// install command. A record-less project takes the local deployment (the
	// asset migration itself rides migrateProjectCommonAssets, driven by the
	// user-asset phase's confirmed install).
	plan.outcome = migrateOptedOut
	return plan, nil
}

// classifiedCleanTarget builds one classified removal target.

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

// pluralMirrorEntries renders the singular/plural noun for a mirror-entry count.
