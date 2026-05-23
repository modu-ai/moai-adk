// Package cli — user-owned namespace protection for moai update.
//
// SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 implementation:
//   - backupUserOwnedNamespace creates a backup at .moai/backups/update-<ISO>/
//     before any destructive operation that could touch user-owned paths.
//   - assertNoUserOwnedNamespaceTouch is the pre-modification abort sentinel
//     (REQ-UNP-006) emitting UPDATE_USER_NAMESPACE_VIOLATION on contraband
//     deploy operations.
//   - newNamespaceBackupStamp formats an ISO-8601 UTC timestamp with hyphens
//     substituted for colons (Windows-safe filenames) per REQ-UNP-010.
//
// Three distinct backup roots coexist after this SPEC:
//   - .moai-backups/                                  (config backups; backupMoaiConfig)
//   - .moai/archive/skills/v2.16-drift-<compact>/     (archive-drift; update_archive.go)
//   - .moai/backups/update-<hyphenated-ISO>/          (this file)
//
// No consolidation; the three concerns remain separately tracked.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// userOwnedScanRoots enumerates the relative directories that
// backupUserOwnedNamespace scans for user-owned content. Each root is checked
// for existence; absent roots contribute zero entries to the backup.
//
// Order is fixed: skills first (smallest footprint typically), then agents,
// then .moai/harness. The deterministic order makes backup directory contents
// predictable for verification.
var userOwnedScanRoots = []string{
	filepath.Join(defs.ClaudeDir, "skills"),  // .claude/skills/ (filter via isUserOwnedNamespace)
	filepath.Join(defs.ClaudeDir, "agents"),  // .claude/agents/ (filter via isUserOwnedNamespace)
	filepath.Join(defs.MoAIDir, "harness"),   // .moai/harness/ (all contents user-owned per REQ-UNP-003)
}

// deployOp represents a planned destructive operation against a single path.
// Used by assertNoUserOwnedNamespaceTouch to inspect a deploy plan before
// any filesystem mutation occurs (REQ-UNP-006).
type deployOp struct {
	rel    string // project-root-relative path (e.g., ".claude/agents/harness/foo.md")
	action string // one of: "overwrite", "delete", "merge"
}

// newNamespaceBackupStamp returns the current UTC timestamp formatted per
// REQ-UNP-010: ^\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z$.
//
// The hyphen substitution (colons → hyphens) ensures Windows-safe filenames
// because ':' is reserved in NTFS path components. The 'Z' suffix and 'T'
// separator preserve ISO-8601 readability.
//
// Distinct from defs.BackupTimestampFormat ("20060102_150405") used by
// backupMoaiConfig, and from update_archive.go driftStamp format
// ("20060102T150405Z"). Three formats, three concerns.
func newNamespaceBackupStamp() string {
	return time.Now().UTC().Format("2006-01-02T15-04-05Z")
}

// resolveNamespaceBackupDir returns the absolute path of the namespace backup
// directory for the given stamp, handling NFR-UNP-004 collision avoidance via
// numeric suffix (-1, -2, ...) when a same-second directory already exists.
//
// If the destination is byte-identical to existing user-owned content, the
// function may return ("", nil) to signal a skip; that decision is made by
// the caller after enumerating contents.
//
// Returns (absolute backup dir path, error). The directory is NOT created here.
func resolveNamespaceBackupDir(projectRoot, stamp string) (string, error) {
	baseDir := filepath.Join(projectRoot, defs.MoAIDir, defs.NamespaceBackupsSubdir)
	candidate := filepath.Join(baseDir, "update-"+stamp)

	// NFR-UNP-004: if directory exists, append numeric suffix
	if _, err := os.Stat(candidate); err == nil {
		for i := 1; i < 1000; i++ {
			candidate = filepath.Join(baseDir, fmt.Sprintf("update-%s-%d", stamp, i))
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				return candidate, nil
			}
		}
		return "", fmt.Errorf("namespace backup collision: exceeded 1000 suffix attempts for stamp %s", stamp)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat namespace backup directory: %w", err)
	}

	return candidate, nil
}

// collectUserOwnedFiles walks userOwnedScanRoots and returns the
// project-root-relative paths of every file matched by isUserOwnedNamespace.
//
// Symlinks are returned as-is (their targets are not dereferenced); the
// downstream copy step copies symlink targets as files per existing copyFile
// semantics. This matches EC-UNP-004 expectation.
func collectUserOwnedFiles(projectRoot string) ([]string, error) {
	var results []string

	for _, root := range userOwnedScanRoots {
		absRoot := filepath.Join(projectRoot, root)
		if _, err := os.Stat(absRoot); err != nil {
			if os.IsNotExist(err) {
				continue // No content at this root — skip
			}
			return nil, fmt.Errorf("stat %s: %w", root, err)
		}

		walkErr := filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(projectRoot, path)
			if relErr != nil {
				return relErr
			}
			// Normalize separators for isUserOwnedNamespace match.
			relNorm := strings.ReplaceAll(rel, "\\", "/")
			if isUserOwnedNamespace(relNorm) {
				results = append(results, relNorm)
			}
			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("walk %s: %w", root, walkErr)
		}
	}

	return results, nil
}

// backupUserOwnedNamespace creates a backup of all user-owned namespace files
// at .moai/backups/update-<stamp>/ before any destructive moai update
// operation runs. Implements REQ-UNP-004 + REQ-UNP-007 + REQ-UNP-010.
//
// Returns:
//   - backupDir (string): the absolute path of the created backup directory
//     (empty string when no user-owned content exists — EC-UNP-001).
//   - err (error): non-nil on filesystem errors. On error mid-copy the
//     partially-written backup directory is removed (EC-UNP-007 defensive
//     cleanup, mirroring update.go:1415 / update_archive.go:91 pattern).
//
// Atomicity (REQ-UNP-007): a `.complete` marker file is written at the backup
// directory root after all copies succeed. Absence of `.complete` indicates a
// partial/aborted backup; consumers should treat such directories as suspect.
//
// Idempotency (NFR-UNP-004): handled by resolveNamespaceBackupDir via numeric
// suffix on same-second collision.
//
// Stderr emission: this function returns the backup directory path; the
// caller (cmdUpdate Backup step) is responsible for emitting the user-facing
// success/skip message via tui.ProgressLine.
func backupUserOwnedNamespace(projectRoot string) (string, error) {
	files, err := collectUserOwnedFiles(projectRoot)
	if err != nil {
		return "", fmt.Errorf("collect user-owned files: %w", err)
	}

	// EC-UNP-001: no user-owned content → no backup directory created
	if len(files) == 0 {
		return "", nil
	}

	stamp := newNamespaceBackupStamp()
	backupDir, err := resolveNamespaceBackupDir(projectRoot, stamp)
	if err != nil {
		return "", err
	}

	// Create the backup root directory
	if err := os.MkdirAll(backupDir, defs.DirPerm); err != nil {
		return "", fmt.Errorf("create namespace backup directory: %w", err)
	}

	// Copy each user-owned file, preserving directory hierarchy.
	for _, rel := range files {
		srcPath := filepath.Join(projectRoot, rel)
		dstPath := filepath.Join(backupDir, rel)

		// Ensure parent directory exists
		if mkErr := os.MkdirAll(filepath.Dir(dstPath), defs.DirPerm); mkErr != nil {
			_ = os.RemoveAll(backupDir) // EC-UNP-007 defensive cleanup
			return "", fmt.Errorf("create backup parent for %s: %w", rel, mkErr)
		}

		// copyFile lives in update_archive.go (same package). Preserves source
		// permission bits up to OS limits; matches existing pattern at
		// update_archive.go:331.
		if copyErr := copyFile(srcPath, dstPath); copyErr != nil {
			_ = os.RemoveAll(backupDir) // EC-UNP-007 defensive cleanup
			return "", fmt.Errorf("copy %s → backup: %w", rel, copyErr)
		}
	}

	// REQ-UNP-007 atomicity marker — written LAST, after all copies succeed.
	markerPath := filepath.Join(backupDir, ".complete")
	markerContent := fmt.Sprintf("stamp=%s\nfiles=%d\ntimestamp=%s\n",
		stamp, len(files), time.Now().UTC().Format(time.RFC3339))
	if markerErr := os.WriteFile(markerPath, []byte(markerContent), defs.FilePerm); markerErr != nil {
		_ = os.RemoveAll(backupDir) // EC-UNP-007 defensive cleanup
		return "", fmt.Errorf("write .complete marker: %w", markerErr)
	}

	return backupDir, nil
}

// assertNoUserOwnedNamespaceTouch is the pre-modification sentinel (REQ-UNP-006).
//
// Iterates the planned deploy operation list and returns an error containing
// the literal string "UPDATE_USER_NAMESPACE_VIOLATION" on the first hit. The
// caller (cmdUpdate) must invoke this BEFORE any filesystem mutation occurs;
// returning a non-nil error aborts the update with a non-zero exit code.
//
// The sentinel string is grep-able by acceptance.md AC-UNP-005. No
// localization (per language.yaml error_messages: en).
//
// @MX:ANCHOR: [AUTO] assertNoUserOwnedNamespaceTouch is the namespace violation gate before destructive ops
// @MX:REASON: [AUTO] REQ-UNP-006 sentinel — must run before any deploy/delete/merge to user-owned path
func assertNoUserOwnedNamespaceTouch(plan []deployOp) error {
	for _, op := range plan {
		// Normalize separators for cross-platform match
		relNorm := strings.ReplaceAll(op.rel, "\\", "/")
		if isUserOwnedNamespace(relNorm) {
			return fmt.Errorf("UPDATE_USER_NAMESPACE_VIOLATION: %s would touch user-owned path: %s",
				op.action, op.rel)
		}
	}
	return nil
}

