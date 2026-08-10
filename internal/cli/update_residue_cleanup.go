// Package cli — residue cleanup for v3-confirmed projects.
//
// SPEC-UPDATE-REINSTALL-LOOP-002 M3 (REQ-RIL2-018..023).
//
// Defect 3: probeVersionSignal's v3-version override short-circuits the v2
// residue signals unconditionally, so a tree carrying full v2 residue but a v3
// version string resolves IsV2=false and is never cleaned. M1's normalization
// made V3VersionConfirmed true for the entire released population, which turns
// that override from a rare edge into the default path — converting issue
// #1243's "loops forever" into "residue is never cleaned", including the
// half-migrated freeze of spec.md §A Defect 3.
//
// The fix narrows the override's CONSEQUENCE rather than the override itself:
// a v3-confirmed project still must not enter the destructive full reinstall
// (REQ-RIL2-018), but one carrying residue is routed here — back up, remove,
// and let the existing .agency/ migration pre-step run. No PRESERVE snapshot,
// no tree wipe, no forced template redeployment (REQ-RIL2-020).
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// v3ResidueCleanupResult reports what the residue-cleanup path did.
type v3ResidueCleanupResult struct {
	// Removed holds the slash-relative deprecated paths whose absence was
	// verified after removal. Derived from the filesystem, never from the
	// planned-list length (the REQ-RIL2-017 discipline).
	Removed []string

	// BackupDir is the absolute path of the backup written before removal.
	// Empty when nothing needed backing up.
	BackupDir string

	// Skipped reports that the project-marker gate closed (REQ-RIL2-023): the
	// directory is not a genuine MoAI project, so nothing was inspected or
	// removed.
	Skipped bool
}

// runV3ResidueCleanup performs the residue-cleanup path for a v3-confirmed
// project. It is deliberately NOT a second clean-reinstall: it takes no template
// deployer, builds no PRESERVE inventory, and writes nothing under
// .moai/backups/v2-to-v3-* (REQ-RIL2-020).
//
// Order of operations, and why it matters:
//
//  1. The project marker is checked HERE rather than only at the call site, so
//     the gate holds for every caller (REQ-RIL2-023).
//  2. The deprecated-path residue is scanned BEFORE the .agency/ migration runs.
//     The migration copies .agency/ to .agency.archived/ and leaves both in
//     place; .agency.archived is itself a defs.DeprecatedPaths entry, so a sweep
//     that scanned afterwards would delete the archive the migration had just
//     written. Sweeping the pre-migration snapshot keeps that archive intact
//     (REQ-RIL2-021 / AC-RIL2-019) while still clearing the legacy .agency/.
//  3. The migration pre-step fires under its existing conditions, unchanged.
//  4. Backup precedes removal, all-or-nothing (REQ-RIL2-019). A backup failure
//     aborts with the grep-able DEPRECATED_BACKUP_FAILED sentinel before any
//     path is deleted — the same contract clean-reinstall Step 4 uses.
func runV3ResidueCleanup(projectRoot string, dryRun, force bool, out io.Writer) (v3ResidueCleanupResult, error) {
	var result v3ResidueCleanupResult
	if out == nil {
		out = os.Stderr
	}

	// (1) REQ-RIL2-023 — genuine MoAI project only.
	if !isMoAIProject(projectRoot) {
		result.Skipped = true
		return result, nil
	}

	// (2) Pre-migration residue snapshot.
	preScan, err := scanDeprecatedPaths(projectRoot)
	if err != nil {
		return result, fmt.Errorf("residue cleanup: scan deprecated paths: %w", err)
	}

	// (3) REQ-RIL2-021 — the independent .agency/ migration pre-step, unchanged.
	if _, agencyStatErr := os.Stat(filepath.Join(projectRoot, ".agency")); agencyStatErr == nil {
		if migrateErr := runAgencyMigrationAdapter(projectRoot, dryRun, force, out); migrateErr != nil {
			return result, fmt.Errorf("pre-step agency migration: %w", migrateErr)
		}
	}

	// Re-filter the snapshot for existence: the migration may have consumed a
	// path (and a concurrent actor may have removed one), and backing up a path
	// that no longer exists would abort the whole sweep.
	sweep := make([]string, 0, len(preScan))
	for _, rel := range preScan {
		if _, statErr := os.Lstat(filepath.Join(projectRoot, filepath.FromSlash(rel))); statErr == nil {
			sweep = append(sweep, rel)
		}
	}
	if len(sweep) == 0 {
		return result, nil
	}

	if dryRun {
		_, _ = fmt.Fprintf(out, "[residue-cleanup] Would remove %d deprecated paths\n", len(sweep))
		return result, nil
	}

	// (4) REQ-RIL2-019 — back up before deleting.
	//
	// expandDeprecatedBackupTargets is required, not optional: scanDeprecatedPaths
	// returns DIRECTORY entries (.agency, .moai/db) while
	// backupDeprecatedPaths copies FILES and errors on a directory. Wiring the two
	// together without the expansion aborts on every tree carrying .agency/.
	backupTargets, expandErr := expandDeprecatedBackupTargets(projectRoot, sweep)
	if expandErr != nil {
		return result, fmt.Errorf("residue cleanup: DEPRECATED_BACKUP_FAILED: enumerate backup targets: %w", expandErr)
	}
	if len(backupTargets) > 0 {
		mgr := manifest.NewManager()
		if _, loadErr := mgr.Load(projectRoot); loadErr != nil {
			return result, fmt.Errorf("residue cleanup: load manifest: %w", loadErr)
		}
		backupDir, bkErr := backupDeprecatedPaths(projectRoot, backupTargets, mgr)
		if bkErr != nil {
			return result, fmt.Errorf("residue cleanup: DEPRECATED_BACKUP_FAILED: %s (and %d other path(s)) not backed up — aborting before REMOVE: %w",
				backupTargets[0], len(backupTargets)-1, bkErr)
		}
		result.BackupDir = backupDir
		_, _ = fmt.Fprintf(out, "[residue-cleanup] Deprecated paths backed up at %s (%d files)\n",
			backupDir, len(backupTargets))
	}

	for _, rel := range sweep {
		abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
		if rmErr := os.RemoveAll(abs); rmErr != nil {
			return result, fmt.Errorf("residue cleanup: remove %s: %w", rel, rmErr)
		}
		// Confirm from the filesystem rather than from the planned list.
		if _, statErr := os.Lstat(abs); os.IsNotExist(statErr) {
			result.Removed = append(result.Removed, rel)
		}
	}

	if len(result.Removed) > 0 {
		_, _ = fmt.Fprintf(out, "[residue-cleanup] Removed %d deprecated paths\n", len(result.Removed))
	}
	return result, nil
}
