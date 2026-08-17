// Package deploy contains the template-deploy, legacy-migration, and
// evolution-scaffold functions extracted from the root update.go as part of
// SPEC-CLI-TUX-V3-003 M3d-B1. These functions execute the destructive side of
// the update flow (stale-path cleanup, .moai/memory → .moai/state migration,
// .agency → .moai migration adapter, .moai/evolution scaffold) and share NO
// root-package-cli helpers, so the package is a leaf (cli → deploy one-way).
//
// Namespace-protection care: CleanMoaiManagedPaths is the stale-file removal
// step that runs BEFORE template deployment. It only removes MoAI-managed
// paths; user-owned namespace preservation is governed by the plan package's
// IsUserOwnedNamespace predicate at the orchestration layer, NOT here.
package deploy

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// preCleanBackupSubdir is the name the unmanaged-file backup tree takes
// inside the run-scoped backup directory — a sibling of the config backup's
// own layout and of legacyMemoryBackupSubdir, so a single timestamped
// directory holds everything one run preserved before it destroyed anything.
const preCleanBackupSubdir = "pre-clean"

// CleanTarget describes one path removed by CleanMoaiManagedPaths: what the
// progress messages show, what gets deleted, and whether the path is a glob.
type CleanTarget struct {
	// DisplayPath is shown in progress messages (e.g., ".claude/settings.json")
	DisplayPath string
	// FullPath is the absolute filesystem path to delete
	FullPath string
	// IsGlob indicates the target uses filepath.Glob matching
	IsGlob bool
}

// ManagedCleanTargets returns the fixed list of MoAI-managed paths that
// CleanMoaiManagedPaths removes before template deployment. It is shared with
// the read-only inventory surface (InventoryManagedPaths) so a deletion
// preview or removal accounting can never drift from the removal itself — a
// diverging copy of this list would be exactly the quiet-failure class the
// t40 observability work exists to prevent.
func ManagedCleanTargets(projectRoot string) []CleanTarget {
	return []CleanTarget{
		{
			DisplayPath: filepath.Join(defs.ClaudeDir, defs.SettingsJSON),
			FullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.SettingsJSON),
		},
		{
			DisplayPath: filepath.Join(defs.ClaudeDir, defs.CommandsMoaiSubdir),
			FullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.CommandsMoaiSubdir),
		},
		{
			DisplayPath: filepath.Join(defs.ClaudeDir, defs.AgentsMoaiSubdir),
			FullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.AgentsMoaiSubdir),
		},
		{
			DisplayPath: filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
			FullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.SkillsSubdir, "moai*"),
			IsGlob:      true,
		},
		{
			DisplayPath: filepath.Join(defs.ClaudeDir, defs.RulesMoaiSubdir),
			FullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.RulesMoaiSubdir),
		},
		{
			DisplayPath: filepath.Join(defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
			FullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.OutputStylesSubdir, "moai"),
		},
		{
			DisplayPath: filepath.Join(defs.ClaudeDir, defs.HooksMoaiSubdir),
			FullPath:    filepath.Join(projectRoot, defs.ClaudeDir, defs.HooksMoaiSubdir),
		},
	}
}

// CleanMoaiManagedPaths removes MoAI-managed directories and files before template
// deployment. This ensures stale files are cleaned up during version upgrades.
// The .moai/config/ directory is deleted entirely (backup was done by the Backup step).
// Paths that do not exist are silently skipped.
//
// tmplFS is the embedded template filesystem — the same FS deployment reads
// from, passed in by the caller so this package stays a leaf. Card t111:
// BEFORE each root is removed, every regular file tmplFS does not carry at
// the same relative path is copied into the run's pre-clean backup
// (.moai-backups/<timestamp>/pre-clean/<root>/...). A backup failure aborts
// before any removal — the REQ-UDS-008 rule generalized from
// MigrateLegacyMemoryDir: a failed backup must never be followed by the
// destruction it was taken to survive. Template-managed files are NOT backed
// up: deployment rewrites them moments later, so their only copy is never at
// stake, and skipping them keeps the backup to what would otherwise be lost
// (measured 2026-08-15: 12 files vanished this way, among them
// .moai/config/astgrep-rules and dev-only rules under .claude/rules/moai).
func CleanMoaiManagedPaths(projectRoot string, out io.Writer, tmplFS fs.FS) error {
	targets := ManagedCleanTargets(projectRoot)

	// One timestamp for the whole run, so every root's unmanaged files land
	// under a single pre-clean directory and are restored together.
	backupBase := filepath.Join(projectRoot, defs.BackupsDir,
		time.Now().Format(defs.BackupTimestampFormat), preCleanBackupSubdir)

	// Process standard targets (files and directories)
	// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces the legacy
	// CR-plus-format pair in this hot path (REQ-UPR-004).
	for _, t := range targets {
		pl := tui.ProgressLine(out, fmt.Sprintf("Removing %s...", t.DisplayPath), nil)

		if t.IsGlob {
			matches, err := filepath.Glob(t.FullPath)
			if err != nil {
				pl.Fail(fmt.Sprintf("Failed to glob %s: %v", t.DisplayPath, err))
				return fmt.Errorf("glob %s: %w", t.DisplayPath, err)
			}
			backedUp := 0
			for _, match := range matches {
				rel, relErr := filepath.Rel(projectRoot, match)
				if relErr != nil {
					pl.Fail(fmt.Sprintf("Failed to remove %s: %v", t.DisplayPath, relErr))
					return fmt.Errorf("rel %s: %w", match, relErr)
				}
				n, rmErr := backupThenRemove(match, rel, backupBase, tmplFS)
				if rmErr != nil {
					pl.Fail(fmt.Sprintf("Failed to remove %s: %v", t.DisplayPath, rmErr))
					return fmt.Errorf("remove %s: %w", match, rmErr)
				}
				backedUp += n
			}
			pl.Done(removedMsg(t.DisplayPath, backedUp))
			continue
		}

		if _, err := os.Stat(t.FullPath); err != nil {
			if os.IsNotExist(err) {
				// Use Done with a "skipped" marker — semantically a successful
				// no-op (target already absent). The leading symbol shifts from
				// "-" to "✓" but the message text is preserved; this aligns
				// the visual category with the success branch.
				pl.Done(fmt.Sprintf("Skipped %s (not found)", t.DisplayPath))
				continue
			}
			pl.Fail(fmt.Sprintf("Failed to stat %s: %v", t.DisplayPath, err))
			return fmt.Errorf("stat %s: %w", t.DisplayPath, err)
		}

		rel, relErr := filepath.Rel(projectRoot, t.FullPath)
		if relErr != nil {
			pl.Fail(fmt.Sprintf("Failed to remove %s: %v", t.DisplayPath, relErr))
			return fmt.Errorf("rel %s: %w", t.FullPath, relErr)
		}
		backedUp, err := backupThenRemove(t.FullPath, rel, backupBase, tmplFS)
		if err != nil {
			pl.Fail(fmt.Sprintf("Failed to remove %s: %v", t.DisplayPath, err))
			return fmt.Errorf("remove %s: %w", t.FullPath, err)
		}
		pl.Done(removedMsg(t.DisplayPath, backedUp))
	}

	// Clean .moai/config/ entirely - backup was already done by the Backup step.
	// For v1.x -> v2.x: old config is incompatible, fresh install needed.
	// For v2.x -> v2.x: backup includes sections/, restore will merge values back.
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
	configDisplayPath := filepath.Join(defs.MoAIDir, defs.ConfigSubdir)
	// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces the legacy
	// CR-plus-format pair (REQ-UPR-004).
	plConfig := tui.ProgressLine(out, fmt.Sprintf("Removing %s...", configDisplayPath), nil)

	// Card t111: files the template does not carry (e.g. a user's
	// astgrep-rules/ tree) reach the pre-clean backup here before the wipe.
	configBackedUp, err := backupThenRemove(configDir,
		filepath.Join(defs.MoAIDir, defs.ConfigSubdir), backupBase, tmplFS)
	if err != nil {
		plConfig.Fail(fmt.Sprintf("Failed to remove %s: %v", configDisplayPath, err))
		return fmt.Errorf("remove %s: %w", configDisplayPath, err)
	}
	plConfig.Done(removedMsg(configDisplayPath, configBackedUp))

	// Migrate legacy .moai/memory/ to .moai/state/.
	// Prior to v2.x, state files (checkpoints, coverage, diagnostics) lived under
	// .moai/memory/. If the old directory still exists, migrate or remove it.
	if err := MigrateLegacyMemoryDir(projectRoot, out); err != nil {
		return err
	}

	return nil
}

// InventoryManagedPaths lists, read-only, every file currently on disk under
// the MoAI-managed clean targets (the same roots CleanMoaiManagedPaths
// removes). Callers use the snapshot to preview deletions before a run and to
// account for removals after one; the walk is advisory, so unreadable entries
// are skipped rather than failed. Paths are returned relative to projectRoot
// with forward separators so they compare directly against template paths.
func InventoryManagedPaths(projectRoot string) []string {
	var files []string
	for _, t := range ManagedCleanTargets(projectRoot) {
		paths := []string{t.FullPath}
		if t.IsGlob {
			matches, err := filepath.Glob(t.FullPath)
			if err != nil {
				continue
			}
			paths = matches
		}
		for _, p := range paths {
			_ = filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				rel, relErr := filepath.Rel(projectRoot, path)
				if relErr != nil {
					return nil
				}
				files = append(files, filepath.ToSlash(rel))
				return nil
			})
		}
	}
	return files
}

// MigrateLegacyMemoryDir handles the .moai/memory/ → .moai/state/ migration.
// If only the old directory exists, it is renamed. If both exist, the old one
// is removed (the new directory takes precedence). If neither exists, this is
// a no-op because template deployment will create .moai/state/.
func MigrateLegacyMemoryDir(projectRoot string, out io.Writer) error {
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

	// SPEC-V3R6-UPDATE-PROGRESS-001 M1: tui.ProgressLine replaces the legacy
	// CR-plus-format pair (REQ-UPR-004).
	plLegacy := tui.ProgressLine(out, fmt.Sprintf("Migrating %s...", legacyDisplayPath), nil)

	stateExists := false
	if _, err := os.Stat(stateDir); err == nil {
		stateExists = true
	}

	if !stateExists {
		// Rename .moai/memory/ → .moai/state/ (fast atomic move).
		if err := os.Rename(legacyDir, stateDir); err != nil {
			plLegacy.Fail(fmt.Sprintf("Failed to migrate %s: %v", legacyDisplayPath, err))
			return fmt.Errorf("migrate %s to %s: %w", legacyDisplayPath, defs.StateSubdir, err)
		}
		plLegacy.Done(fmt.Sprintf("Migrated %s → %s", legacyDisplayPath, filepath.Join(defs.MoAIDir, defs.StateSubdir)))
	} else {
		// Both exist — state directory takes precedence; remove legacy.
		//
		// SPEC-UPDATE-DATA-SURVIVAL-001 REQ-UDS-008: the legacy directory holds
		// user-authored data that nothing else in the update subsystem backs up
		// (it is in none of preserveInventoryRoots, BackupMoaiConfig's
		// .moai/config scope, or userOwnedScanRoots), so removing it outright is
		// unrecoverable. Copy it out first, and abort the removal when the copy
		// fails — a failed backup must never be followed by the destruction it
		// was taken to survive.
		backupDir, err := backupLegacyMemoryDir(projectRoot, legacyDir)
		if err != nil {
			plLegacy.Fail(fmt.Sprintf("Failed to back up %s: %v", legacyDisplayPath, err))
			return fmt.Errorf("back up legacy %s: %w", legacyDisplayPath, err)
		}

		if err := os.RemoveAll(legacyDir); err != nil {
			plLegacy.Fail(fmt.Sprintf("Failed to remove %s: %v", legacyDisplayPath, err))
			return fmt.Errorf("remove legacy %s: %w", legacyDisplayPath, err)
		}
		plLegacy.Done(fmt.Sprintf("Removed legacy %s (backed up to %s)", legacyDisplayPath, backupDir))
	}

	return nil
}

// legacyMemoryBackupSubdir is the name the legacy .moai/memory/ tree takes
// inside the run-scoped backup directory. It is distinct from the config
// backup's own layout, so a run whose config backup shares the same timestamp
// directory cannot collide with it.
const legacyMemoryBackupSubdir = "legacy-memory"

// backupLegacyMemoryDir copies legacyDir into a run-scoped backup directory
// under the project's backup root and returns that directory
// (SPEC-UPDATE-DATA-SURVIVAL-001 REQ-UDS-008).
//
// The timestamp format matches backup.BackupMoaiConfig, so a memory backup and
// a config backup taken in the same run land under one timestamped directory
// and are restored together.
func backupLegacyMemoryDir(projectRoot, legacyDir string) (string, error) {
	timestamp := time.Now().Format(defs.BackupTimestampFormat)
	dest := filepath.Join(projectRoot, defs.BackupsDir, timestamp, legacyMemoryBackupSubdir)

	if err := copyTree(legacyDir, dest); err != nil {
		return "", fmt.Errorf("copy %s to %s: %w", legacyDir, dest, err)
	}
	return dest, nil
}

// copyTree recursively copies the regular files under src into dst, recreating
// the directory structure. Symlinks and other irregular entries are skipped:
// this is a data-preservation copy, and following a link would write outside
// the backup directory.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, defs.DirPerm)
		}
		if !d.Type().IsRegular() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), defs.DirPerm); err != nil {
			return err
		}
		return os.WriteFile(target, data, defs.FilePerm)
	})
}

// removedMsg renders the per-target removal result, naming the unmanaged
// file count only when the backup actually carried something — a count of
// zero would be noise on every clean upgrade.
func removedMsg(displayPath string, backedUp int) string {
	if backedUp == 0 {
		return fmt.Sprintf("Removed %s", displayPath)
	}
	return fmt.Sprintf("Removed %s (backed up %d unmanaged file(s))", displayPath, backedUp)
}

// backupThenRemove backs up every regular file under diskPath that tmplFS
// does not carry at the same relative path, then removes diskPath, and
// returns how many files the backup carried.
//
// The abort ordering is the contract: backup FIRST, removal SECOND, and a
// backup failure returns before os.RemoveAll runs. Paths that do not exist
// are a successful no-op (the caller's not-found skip and the config root's
// historic IsNotExist tolerance both land here).
//
// A FILE target (settings.json) is backed up unless the template carries the
// exact path; a DIRECTORY target is walked and compared against the template
// set collected under the same relative root. Directory entries that are not
// regular files (symlinks, fifos) are skipped by the copy — the same rule
// copyTree applies: a preservation copy must not follow a link out of the
// backup tree.
func backupThenRemove(diskPath, relTarget, backupBase string, tmplFS fs.FS) (int, error) {
	info, err := os.Stat(diskPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("stat %s: %w", relTarget, err)
	}

	if !info.IsDir() {
		if templateCarries(tmplFS, relTarget) {
			return 0, os.RemoveAll(diskPath)
		}
		if err := copyRegularFile(diskPath, filepath.Join(backupBase, relTarget)); err != nil {
			return 0, fmt.Errorf("back up %s: %w", relTarget, err)
		}
		return 1, os.RemoveAll(diskPath)
	}

	managed, err := templateManagedPaths(tmplFS, relTarget)
	if err != nil {
		return 0, fmt.Errorf("collect template paths under %s: %w", relTarget, err)
	}
	backedUp, err := backupUnmanagedTree(diskPath, relTarget, managed, backupBase)
	if err != nil {
		return 0, fmt.Errorf("back up %s: %w", relTarget, err)
	}
	return backedUp, os.RemoveAll(diskPath)
}

// templateManagedPaths returns the set of slash-separated file paths tmplFS
// carries under prefix — the paths this update is about to redeploy at the
// same relative locations. A prefix the template does not carry at all yields
// an empty set: every file under the disk root is then unmanaged and every
// one of them is backed up.
func templateManagedPaths(tmplFS fs.FS, prefix string) (map[string]bool, error) {
	managed := make(map[string]bool)
	err := fs.WalkDir(tmplFS, filepath.ToSlash(prefix), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				// The template carries nothing under this root.
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		managed[path] = true
		return nil
	})
	return managed, err
}

// templateCarries reports whether tmplFS holds a regular file at rel.
func templateCarries(tmplFS fs.FS, rel string) bool {
	info, err := fs.Stat(tmplFS, filepath.ToSlash(rel))
	return err == nil && !info.IsDir()
}

// backupUnmanagedTree copies every regular file under diskRoot whose
// slash-relative path (joined onto relRoot) is absent from managed into
// backupBase/relRoot/..., preserving the directory layout, and returns the
// number of files copied.
func backupUnmanagedTree(diskRoot, relRoot string, managed map[string]bool, backupBase string) (int, error) {
	count := 0
	err := filepath.WalkDir(diskRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(diskRoot, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(filepath.Join(relRoot, rel))
		if managed[key] {
			return nil
		}
		if err := copyRegularFile(path, filepath.Join(backupBase, relRoot, rel)); err != nil {
			return fmt.Errorf("copy %s: %w", path, err)
		}
		count++
		return nil
	})
	return count, err
}

// copyRegularFile copies one regular file into dst, creating parent
// directories as needed. Same read-into-memory shape as copyTree's file
// branch — these are configuration-sized files, not bulk data.
func copyRegularFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), defs.DirPerm); err != nil {
		return err
	}
	return os.WriteFile(dst, data, defs.FilePerm)
}

// ScaffoldEvolutionDir ensures the .moai/evolution/ directory tree exists.
// Called during both init and update so that existing projects that predate
// the evolution infrastructure receive the directory structure automatically.
// Non-destructive: only creates missing files; never overwrites existing ones.
func ScaffoldEvolutionDir(projectRoot string) error {
	evolutionDir := filepath.Join(projectRoot, ".moai", "evolution")

	// Sub-directories to create.
	subdirs := []string{
		"telemetry",
		"learnings",
		"new-skills",
	}

	for _, sub := range subdirs {
		dir := filepath.Join(evolutionDir, sub)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}

		// Ensure .gitkeep exists so the directory is tracked by git.
		gitkeep := filepath.Join(dir, ".gitkeep")
		if _, err := os.Stat(gitkeep); os.IsNotExist(err) {
			if err := os.WriteFile(gitkeep, []byte{}, 0o644); err != nil {
				return fmt.Errorf("create %s: %w", gitkeep, err)
			}
		}
	}

	// Create default manifest.yaml if missing.
	manifestPath := filepath.Join(evolutionDir, "manifest.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		const defaultManifest = `schema_version: 1
evolved_skills: []
new_skills: []
learnings_count: 0
last_evolution_date: ""
rate_limit:
  week_start: ""
  proposals_this_week: 0
  last_proposal_time: ""
`
		if err := os.WriteFile(manifestPath, []byte(defaultManifest), 0o644); err != nil {
			return fmt.Errorf("create %s: %w", manifestPath, err)
		}
	}

	// Create default changelog.md if missing.
	changelogPath := filepath.Join(evolutionDir, "changelog.md")
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		const defaultChangelog = `# MoAI Evolution Changelog

All notable skill evolutions and learning graduations will be documented here.

## Format

Each entry: date, learning ID, skill affected, change summary.
`
		if err := os.WriteFile(changelogPath, []byte(defaultChangelog), 0o644); err != nil {
			return fmt.Errorf("create %s: %w", changelogPath, err)
		}
	}

	return nil
}
