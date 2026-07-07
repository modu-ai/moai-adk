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
	filepath.Join(defs.ClaudeDir, "skills"), // .claude/skills/ (filter via isUserOwnedNamespace)
	filepath.Join(defs.ClaudeDir, "agents"), // .claude/agents/ (filter via isUserOwnedNamespace)
	filepath.Join(defs.MoAIDir, "harness"),  // .moai/harness/ (all contents user-owned per REQ-UNP-003)
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

// collectUserOwnedFilesWith walks userOwnedScanRoots and returns the
// project-root-relative paths of every file matched by classify.
//
// Symlinks are ALWAYS skipped (REQ-SEC-003): a symlink inside a user-owned
// namespace must not have its dereferenced target content recorded into the
// backup. The guard reuses isSymlinkEntry (update.go) — the same Lstat-based
// pattern proven in SPEC-SEC-HARDEN-003 — so no new pattern is invented.
func collectUserOwnedFilesWith(projectRoot string, classify func(string) bool) ([]string, error) {
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
			// REQ-SEC-003 / AC-SEC-003b: skip symlinks so copyFile never
			// records the link's dereferenced target content into the backup.
			if isSymlinkEntry(path) {
				return nil
			}
			rel, relErr := filepath.Rel(projectRoot, path)
			if relErr != nil {
				return relErr
			}
			// Normalize separators for classifier match.
			relNorm := strings.ReplaceAll(rel, "\\", "/")
			if classify(relNorm) {
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

// collectUserOwnedFiles walks userOwnedScanRoots and returns the
// project-root-relative paths of every file matched by the STRICT
// isUserOwnedNamespace classifier. Symlinks are skipped (REQ-SEC-003).
//
// Used by buildPreserveInventory (update_preserve_inventory.go) which needs the
// strict authoritative classification — clean-reinstall preserve semantics must
// NOT be broadened by the conservative backup expansion (REQ-SEC-005 is a
// backup-only concern).
func collectUserOwnedFiles(projectRoot string) ([]string, error) {
	return collectUserOwnedFilesWith(projectRoot, isUserOwnedNamespace)
}

// collectUserOwnedFilesConservative walks userOwnedScanRoots and returns the
// project-root-relative paths of every file matched by the CONSERVATIVE
// isUserOwnedNamespaceConservative classifier. Symlinks are skipped
// (REQ-SEC-003).
//
// Used by backupUserOwnedNamespace (REQ-SEC-005 conservative backup
// expansion): reserved-prefix names that are ambiguous (could be user-authored,
// e.g. moai-my-notes / expert-mydomain.md) are included in the backup pass so
// they are never overwritten/deleted without a backup.
func collectUserOwnedFilesConservative(projectRoot string) ([]string, error) {
	return collectUserOwnedFilesWith(projectRoot, isUserOwnedNamespaceConservative)
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
//
// REQ-SEC-005: uses the CONSERVATIVE collector so reserved-prefix names that
// are ambiguous (could be user-authored) are also backed up rather than risk
// overwrite-without-backup.
func backupUserOwnedNamespace(projectRoot string) (string, error) {
	files, err := collectUserOwnedFilesConservative(projectRoot)
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
// @MX:REASON: [AUTO] REQ-UNP-006 sentinel — must run before any deploy/delete/merge to user-owned path; wired as a real pre-modification abort gate via verifyNamespaceBackupCoverage (SPEC-INTERNAL-SECURITY-001 REQ-SEC-006)
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

// verifyNamespaceBackupCoverage is the wired pre-modification abort gate for
// SPEC-INTERNAL-SECURITY-001 REQ-SEC-006 (AC-SEC-006a). It runs AFTER the
// namespace backup completes and BEFORE any destructive deploy step, confirming
// that every user-owned namespace file on disk was captured in the backup.
//
// It collects the conservative user-owned set, builds a deploy plan of those
// NOT present in backupDir, and delegates to assertNoUserOwnedNamespaceTouch
// — which emits the grep-able UPDATE_USER_NAMESPACE_VIOLATION sentinel on the
// first user-owned file that would be overwritten without a backup.
//
// Behavior preservation (NFR-SEC-003):
//   - No user-owned content → collected set empty → empty plan → nil (pass).
//   - All user-owned content backed up → every file present in backupDir →
//     empty plan → nil (pass). Normal updates still succeed.
//   - A user-owned file on disk missing from backupDir → non-empty plan →
//     UPDATE_USER_NAMESPACE_VIOLATION (abort before destructive ops).
//
// @MX:NOTE: [AUTO] SPEC-INTERNAL-SECURITY-001 REQ-SEC-006 — wires the previously-dead
// assertNoUserOwnedNamespaceTouch as a real backup-coverage gate on the moai update deploy path.
func verifyNamespaceBackupCoverage(projectRoot, backupDir string) error {
	collected, err := collectUserOwnedFilesConservative(projectRoot)
	if err != nil {
		return fmt.Errorf("verify namespace backup coverage: %w", err)
	}

	var unprotected []deployOp
	for _, rel := range collected {
		backedUp := false
		if backupDir != "" {
			if _, statErr := os.Stat(filepath.Join(backupDir, rel)); statErr == nil {
				backedUp = true
			}
		}
		if !backedUp {
			unprotected = append(unprotected, deployOp{rel: rel, action: "overwrite"})
		}
	}

	return assertNoUserOwnedNamespaceTouch(unprotected)
}

// isUserOwnedNamespaceConservative is a superset of isUserOwnedNamespace used
// by the backup pass (collectUserOwnedFilesConservative) for
// SPEC-INTERNAL-SECURITY-001 REQ-SEC-005 conservative backup expansion.
//
// Reserved-prefix names whose provenance is ambiguous (could be MoAI-managed OR
// user-authored) are treated as user-owned so they are included in the backup
// pass — back up rather than risk overwrite-without-backup (R4 low-risk path).
// The accepted tradeoff is increased backup size: real MoAI-managed assets with
// reserved prefixes (e.g. moai-foundation-cc) are also backed up.
//
// Unambiguous MoAI system directories (.claude/agents/{core,expert,meta}) and
// non-namespace config paths remain excluded.
//
// @MX:NOTE: [AUTO] SPEC-INTERNAL-SECURITY-001 REQ-SEC-005 — conservative backup superset of isUserOwnedNamespace.
func isUserOwnedNamespaceConservative(rel string) bool {
	// Strict superset: anything the authoritative check classifies as user-owned.
	if isUserOwnedNamespace(rel) {
		return true
	}

	norm := strings.ReplaceAll(rel, "\\", "/")

	// REQ-SEC-005: .claude/skills/ with reserved prefix moai/moai- is ambiguous
	// (user could have authored moai-my-notes). Back it up conservatively.
	if strings.HasPrefix(norm, ".claude/skills/") {
		rest := strings.TrimPrefix(norm, ".claude/skills/")
		seg := strings.SplitN(rest, "/", 2)[0]
		if seg == "moai" || strings.HasPrefix(seg, "moai-") {
			return true
		}
	}

	// REQ-SEC-005: .claude/agents/ top-level files with reserved prefixes are
	// ambiguous. System agent directories (core/expert/meta) stay excluded —
	// they are unambiguously MoAI-managed, not user-authored.
	if strings.HasPrefix(norm, ".claude/agents/") {
		rest := strings.TrimPrefix(norm, ".claude/agents/")
		seg := strings.SplitN(rest, "/", 2)[0]
		switch seg {
		case "core", "expert", "meta":
			return false // unambiguously MoAI system agents
		}
		base := seg
		if dot := strings.LastIndex(base, "."); dot > 0 {
			base = base[:dot]
		}
		if strings.HasPrefix(base, "moai-") || strings.HasPrefix(base, "moai") ||
			strings.HasPrefix(base, "manager-") || strings.HasPrefix(base, "expert-") ||
			strings.HasPrefix(base, "builder-") || strings.HasPrefix(base, "evaluator-") {
			return true
		}
	}

	return false
}
