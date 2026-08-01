package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// SPEC-UPDATE-DATA-SURVIVAL-001 M3 — on-disk backup before the first
// destructive step.
//
// Three files used to be protected only by an in-process []byte captured into
// updatemerge.FileBackup (and, for .gitignore, a bare slice). Between the
// removal and the merge-back, their contents existed nowhere but the process
// heap: a crash, an ENOSPC, or a Ctrl+C in that window lost them permanently.
//
// This file adds an on-disk copy written BEFORE the first destructive step on
// both execution paths. It is additive insurance for the crash window, NOT a
// replacement for the merge-back — the in-memory path is retained unchanged
// (REQ-UDS-004).

// diskBackupSubdir is the sub-directory of the run-scoped backup directory that
// holds the crash-window copies. It is namespaced so it cannot collide with the
// .moai/config tree BackupMoaiConfig writes alongside it.
const diskBackupSubdir = "in-memory-backups"

// backupWriteFailedSentinel is the grep-able marker a failed on-disk backup
// write surfaces (REQ-UDS-003). It is always followed by the project-root-
// relative path of the file that could not be backed up.
const backupWriteFailedSentinel = "backup-write-failed:"

// diskBackupWriteFile is the seam through which the on-disk copies are written.
// Tests replace it to reach the write-failure branch deterministically rather
// than by racing a real ENOSPC (NFR-UDS-001).
var diskBackupWriteFile = os.WriteFile

// inMemoryOnlyBackupTargets returns the project-root-relative files whose only
// backup, before this SPEC, lived in process memory.
//
// This is the single source of truth for both execution paths (REQ-UDS-005):
// the template-sync path and the clean-reinstall path build their in-memory
// slices independently, so a per-path list here would let one path silently
// become the unprotected one.
func inMemoryOnlyBackupTargets() []string {
	return []string{
		".claude/settings.json",
		".moai/status_line.sh",
		".gitignore",
	}
}

// ensureRunBackupDir returns the run-scoped backup directory to write into.
//
// When the caller already has one — the path BackupMoaiConfig or the
// clean-reinstall Step 3 snapshot produced — it is reused, so a run has exactly
// one backup directory (REQ-UDS-002). When the caller has none (a project with
// no .moai/config gives BackupMoaiConfig nothing to back up), a fresh
// run-scoped directory is created under the same backup root.
func ensureRunBackupDir(projectRoot, backupDir string) (string, error) {
	if backupDir != "" {
		if err := os.MkdirAll(backupDir, defs.DirPerm); err != nil {
			return "", fmt.Errorf("create run backup directory: %w", err)
		}
		return backupDir, nil
	}

	dir := filepath.Join(projectRoot, defs.BackupsDir, time.Now().Format(defs.BackupTimestampFormat))
	if err := os.MkdirAll(dir, defs.DirPerm); err != nil {
		return "", fmt.Errorf("create run backup directory: %w", err)
	}
	return dir, nil
}

// backupInMemoryOnlyFiles copies every existing in-memory-only target into
// backupDir/<diskBackupSubdir>/ preserving its project-root-relative layout.
//
// It returns the resolved run-scoped backup directory (so the caller can report
// it) and the relative paths actually written. A target that does not exist in
// the project is skipped silently — its absence is not a failure.
//
// On the first write failure it returns an error carrying
// backupWriteFailedSentinel plus the offending relative path, so the caller can
// abort before destroying anything (REQ-UDS-003).
func backupInMemoryOnlyFiles(projectRoot, backupDir string) (string, []string, error) {
	resolvedDir, err := ensureRunBackupDir(projectRoot, backupDir)
	if err != nil {
		return "", nil, err
	}

	var written []string
	for _, rel := range inMemoryOnlyBackupTargets() {
		src := filepath.Join(projectRoot, filepath.FromSlash(rel))
		data, readErr := os.ReadFile(src)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			return resolvedDir, written, fmt.Errorf("%s %s: read: %w", backupWriteFailedSentinel, rel, readErr)
		}

		dst := filepath.Join(resolvedDir, diskBackupSubdir, filepath.FromSlash(rel))
		if mkErr := os.MkdirAll(filepath.Dir(dst), defs.DirPerm); mkErr != nil {
			return resolvedDir, written, fmt.Errorf("%s %s: create directory: %w", backupWriteFailedSentinel, rel, mkErr)
		}
		if wErr := diskBackupWriteFile(dst, data, fs.FileMode(defs.FilePerm)); wErr != nil {
			return resolvedDir, written, fmt.Errorf("%s %s: write: %w", backupWriteFailedSentinel, rel, wErr)
		}
		written = append(written, rel)
	}

	return resolvedDir, written, nil
}

// guardFirstDestructiveStep writes the on-disk crash-window copies and only
// then runs destructive.
//
// A backup-write failure aborts BEFORE destructive is invoked (REQ-UDS-003):
// the destructive function is never called, so nothing is removed while its
// only copy is still in memory. Both execution paths route their first
// destructive step through this function, which is what makes their coverage
// structurally identical rather than coincidental (REQ-UDS-005).
func guardFirstDestructiveStep(projectRoot, backupDir string, destructive func() error) error {
	if _, _, err := backupInMemoryOnlyFiles(projectRoot, backupDir); err != nil {
		return err
	}
	return destructive()
}
