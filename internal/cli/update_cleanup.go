// Package cli — update_cleanup.go
// Implements SPEC-V3R3-UPDATE-CLEANUP-001: deprecated path detection, backup,
// user confirmation, manifest provenance classification, telemetry, symlink
// safety, and lock file management for `moai update`.

package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	// updateLockFile is the path of the lock file relative to .moai/.
	// JSON payload: {pid, started_at, hostname} (OQ4, spec.md §6).
	updateLockFile = ".update.lock"

	// BackupDirPrefix is the prefix for the dated backup directory (REQ-UPC-009).
	BackupDirPrefix = "agency-"

	// cleanupSpecID is the SPEC ID recorded in backup MANIFEST.json (REQ-UPC-011).
	cleanupSpecID = "SPEC-V3R3-UPDATE-CLEANUP-001"
)

// ErrUpdateLockHeld is returned when a concurrent moai update is already running.
var ErrUpdateLockHeld = errors.New("moai update is already running (lock held by another process)")

// ---------------------------------------------------------------------------
// M1 — Lock file management (REQ-UPC-005, NFR-UPC-S2)
// ---------------------------------------------------------------------------

// updateLockPayload is the JSON schema stored in .moai/.update.lock (OQ4).
type updateLockPayload struct {
	PID       int    `json:"pid"`
	StartedAt string `json:"started_at"`
	Hostname  string `json:"hostname"`
}

// acquireUpdateLock creates the lock file under <projectRoot>/.moai/.update.lock.
// On success it returns a release function that removes the lock file.
// If the lock is already held by a live process it returns ErrUpdateLockHeld.
func acquireUpdateLock(projectRoot string) (release func(), err error) {
	lockPath := filepath.Join(projectRoot, defs.MoAIDir, updateLockFile)

	// Attempt to detect and clean a stale lock first.
	cleanStaleLock(lockPath)

	// Try to create the lock file exclusively.
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil, ErrUpdateLockHeld
		}
		return nil, fmt.Errorf("create lock file %q: %w", lockPath, err)
	}

	hostname, _ := os.Hostname()
	payload := updateLockPayload{
		PID:       os.Getpid(),
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Hostname:  hostname,
	}
	data, _ := json.Marshal(payload)
	_, _ = f.Write(data)
	_ = f.Close()

	return func() { _ = os.Remove(lockPath) }, nil
}

// cleanStaleLock removes the lock file if it belongs to a dead process.
// Stale detection: parse PID from JSON payload and check if the process exists.
func cleanStaleLock(lockPath string) {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return // no lock file or unreadable — not stale
	}
	var payload updateLockPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return // malformed — leave it; real lock may be writing
	}
	if payload.PID <= 0 {
		return
	}
	if !isProcessAlive(payload.PID) {
		_ = os.Remove(lockPath)
	}
}

// isProcessAlive is defined in platform-specific files:
//   - update_cleanup_unix.go    (build tag: !windows) — uses syscall.Kill(pid, 0)
//   - update_cleanup_windows.go (build tag: windows)  — conservative no-op
//
// SPEC-V2.20.0-RC1 hotfix: extracted to resolve `syscall.Kill undefined` on
// Windows compilation. The previous `if runtime.GOOS == "windows" { return }`
// guard was insufficient because the Go compiler still resolves syscall.Kill
// at compile time for the Windows target before runtime branch elimination.

// ---------------------------------------------------------------------------
// M2 — Deprecated path detection (REQ-UPC-007, REQ-UPC-024, REQ-UPC-025)
// ---------------------------------------------------------------------------

// scanDeprecatedPaths returns the slash-relative paths (relative to projectRoot)
// of deprecated files that exist in the project. Paths under .moai/backup/ and
// .moai/logs/ are excluded (REQ-UPC-024). Absent files are silently skipped
// (REQ-UPC-025).
func scanDeprecatedPaths(projectRoot string) ([]string, error) {
	var found []string
	for _, entry := range defs.DeprecatedPaths {
		rel := filepath.ToSlash(entry.Path)
		// Exclude self-referential backup and logs paths (REQ-UPC-024)
		if strings.HasPrefix(rel, ".moai/backup/") ||
			strings.HasPrefix(rel, ".moai/logs/") {
			continue
		}
		abs := filepath.Join(projectRoot, filepath.FromSlash(rel))
		// Use Lstat so we detect symlinks without following them.
		if _, err := os.Lstat(abs); err != nil {
			if os.IsNotExist(err) {
				continue // silent skip (REQ-UPC-025)
			}
			return nil, fmt.Errorf("stat deprecated path %q: %w", rel, err)
		}
		found = append(found, rel)
	}
	return found, nil
}

// ---------------------------------------------------------------------------
// M2 — Skip marker (REQ-UPC-018)
// ---------------------------------------------------------------------------

// filterSkipMarkerPaths partitions paths into (filtered, skipped).
// A path is skipped if its containing directory has a .moai-skip-cleanup marker.
// Skipped paths are logged to out with an [INFO] line.
func filterSkipMarkerPaths(projectRoot string, paths []string, out io.Writer) (filtered, skipped []string) {
	checkedDirs := make(map[string]bool)
	for _, rel := range paths {
		dir := filepath.Dir(filepath.FromSlash(rel))
		skipDir, ok := checkedDirs[dir]
		if !ok {
			markerPath := filepath.Join(projectRoot, dir, ".moai-skip-cleanup")
			_, err := os.Stat(markerPath)
			skipDir = err == nil
			checkedDirs[dir] = skipDir
			if skipDir {
				_, _ = fmt.Fprintf(out, "[INFO] cleanup skipped due to .moai-skip-cleanup marker: %s\n",
					filepath.ToSlash(dir))
			}
		}
		if skipDir {
			skipped = append(skipped, rel)
		} else {
			filtered = append(filtered, rel)
		}
	}
	return filtered, skipped
}

// ---------------------------------------------------------------------------
// M2 — Backup (REQ-UPC-009, REQ-UPC-010, REQ-UPC-011)
// ---------------------------------------------------------------------------

// BackupManifestFile is the file written to the backup directory.
type BackupManifestFile struct {
	SpecID    string            `json:"spec_id"`
	DeletedAt time.Time         `json:"deleted_at"`
	Files     []BackupFileEntry `json:"files"`
}

// BackupFileEntry records metadata for one backed-up file.
type BackupFileEntry struct {
	Path               string `json:"path"`
	Hash               string `json:"hash"`
	Classification     string `json:"classification"`
	SymlinkTarget      string `json:"symlink_target,omitempty"`
	SymlinkTargetStatus string `json:"symlink_target_status,omitempty"`
}

// backupDeprecatedPaths copies each path from the project into a dated backup
// directory and writes a MANIFEST.json. Returns the backup directory path.
// If any copy fails the entire operation is aborted (REQ-UPC-010).
func backupDeprecatedPaths(projectRoot string, paths []string, mgr manifest.Manager) (string, error) {
	if len(paths) == 0 {
		return "", nil
	}

	ts := time.Now().UTC().Format("2006-01-02T15-04-05Z")
	backupDir := filepath.Join(projectRoot, defs.MoAIDir, "backup",
		BackupDirPrefix+ts)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup dir %q: %w", backupDir, err)
	}

	manifest := BackupManifestFile{
		SpecID:    cleanupSpecID,
		DeletedAt: time.Now().UTC(),
	}

	for _, rel := range paths {
		srcPath := filepath.Join(projectRoot, filepath.FromSlash(rel))
		info := inspectDeprecatedPath(srcPath)
		class := classifyDeprecatedFile(projectRoot, rel, mgr)

		entry := BackupFileEntry{
			Path:           rel,
			Classification: string(class),
		}

		if info.IsSymlink {
			entry.SymlinkTarget = info.SymlinkTarget
			entry.SymlinkTargetStatus = info.SymlinkTargetStatus
			// For symlinks we record metadata only — do not copy the target.
		} else {
			// Copy file contents preserving directory structure
			destPath := filepath.Join(backupDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
				return "", fmt.Errorf("backup mkdir %q: %w", filepath.Dir(destPath), err)
			}
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return "", fmt.Errorf("backup read %q: %w", rel, err)
			}
			if err := os.WriteFile(destPath, data, 0o644); err != nil {
				return "", fmt.Errorf("backup write %q: %w", rel, err)
			}
			entry.Hash = manifest2hashBytes(data)
		}

		manifest.Files = append(manifest.Files, entry)
	}

	// Write MANIFEST.json
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal MANIFEST.json: %w", err)
	}
	manifestPath := filepath.Join(backupDir, "MANIFEST.json")
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return "", fmt.Errorf("write MANIFEST.json: %w", err)
	}

	return backupDir, nil
}

// manifest2hashBytes is a helper that avoids importing the manifest package's
// internal hash directly; uses the same SHA-256 via manifest.HashBytes facade.
func manifest2hashBytes(data []byte) string {
	return manifest.HashBytes(data)
}

// ---------------------------------------------------------------------------
// M3 — Path inspection (symlink detection, REQ-UPC-023)
// ---------------------------------------------------------------------------

// DeprecatedPathInfo describes metadata about a deprecated path on the filesystem.
type DeprecatedPathInfo struct {
	IsSymlink           bool
	SymlinkTarget       string
	SymlinkTargetStatus string // "ok" | "broken"
}

// inspectDeprecatedPath uses os.Lstat to determine whether a path is a symlink
// and, if so, whether the target exists (REQ-UPC-023).
func inspectDeprecatedPath(abs string) DeprecatedPathInfo {
	linfo, err := os.Lstat(abs)
	if err != nil {
		return DeprecatedPathInfo{}
	}
	if linfo.Mode()&os.ModeSymlink == 0 {
		return DeprecatedPathInfo{IsSymlink: false}
	}
	target, err := os.Readlink(abs)
	if err != nil {
		return DeprecatedPathInfo{IsSymlink: true, SymlinkTargetStatus: "broken"}
	}
	// Check if target exists
	status := "ok"
	if _, err := os.Stat(abs); err != nil { // follows the link
		status = "broken"
	}
	return DeprecatedPathInfo{
		IsSymlink:           true,
		SymlinkTarget:       target,
		SymlinkTargetStatus: status,
	}
}

// removeDeprecatedFile removes a deprecated file or symlink from the project.
// For symlinks, only the link itself is removed (target is preserved).
func removeDeprecatedFile(abs string) error {
	// Use Lstat to avoid following symlinks
	info, err := os.Lstat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // already gone
		}
		return fmt.Errorf("lstat %q: %w", abs, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		// Regular file or symlink: use os.Remove (does not follow symlinks)
		return os.Remove(abs)
	}
	// Directory: use RemoveAll
	return os.RemoveAll(abs)
}

// ---------------------------------------------------------------------------
// M3 — Provenance classification (REQ-UPC-015a/b/c)
// ---------------------------------------------------------------------------

// DeprecatedClassification describes how a deprecated file was classified.
type DeprecatedClassification string

const (
	DeprecatedPristine     DeprecatedClassification = "PristineDeprecated"
	DeprecatedUserModified DeprecatedClassification = "UserModifiedDeprecated"
	DeprecatedUnverified   DeprecatedClassification = "UnverifiedDeprecated"
)

// classifyDeprecatedFile compares the current file content hash against the
// manifest DeployedHash and returns the appropriate classification.
func classifyDeprecatedFile(projectRoot, relPath string, mgr manifest.Manager) DeprecatedClassification {
	entry, found := mgr.GetEntry(relPath)
	if !found {
		return DeprecatedUnverified
	}
	abs := filepath.Join(projectRoot, filepath.FromSlash(relPath))
	data, err := os.ReadFile(abs)
	if err != nil {
		return DeprecatedUnverified
	}
	currentHash := manifest.HashBytes(data)
	if currentHash == entry.DeployedHash {
		return DeprecatedPristine
	}
	return DeprecatedUserModified
}

// ---------------------------------------------------------------------------
// M3 — Telemetry (REQ-UPC-022)
// ---------------------------------------------------------------------------

// CleanupEvent is the structured telemetry record emitted at cleanup phase end.
type CleanupEvent struct {
	AtomicWriteUsed       bool     `json:"atomic_write_used"`
	PreUpdateSuffix2Files []string `json:"pre_update_suffix2_files"`
	BackupOutcome         string   `json:"backup_outcome"`  // "success" | "skipped" | "failed"
	BackupPath            string   `json:"backup_path"`
	CleanupOutcome        string   `json:"cleanup_outcome"` // "completed" | "deferred" | "aborted"
	UserOptOutPaths       []string `json:"user_opt_out_paths"`
}

// ---------------------------------------------------------------------------
// M4 — Case-insensitive filesystem probe (REQ-UPC-026)
// ---------------------------------------------------------------------------

// probeCaseSensitiveFS probes the filesystem rooted at projectRoot by creating a
// temporary file then attempting to stat it with an upper-case variant.
//
//   - Returns (true, nil) when the FS is case-sensitive (Linux ext4 default).
//   - Returns (false, nil) when the FS is case-insensitive (macOS APFS default).
//   - Returns (true, nil) on probe failure (fallback to safe case-sensitive mode)
//     and logs a [INFO] message to stderr (REQ-UPC-026 probe-failure fallback).
//
// The probe file is always cleaned up before returning (REQ-UPC-026).
func probeCaseSensitiveFS(projectRoot string) (caseSensitive bool, _ error) {
	probeFile := filepath.Join(projectRoot, ".moai-fscase-probe")

	// Write lower-case probe file.
	if err := os.WriteFile(probeFile, []byte("probe"), 0o600); err != nil {
		// Cannot write → fallback to case-sensitive (safe default).
		// Not an error from caller perspective; probe just couldn't run.
		return true, nil
	}
	defer os.Remove(probeFile) //nolint:errcheck

	// Attempt to stat the UPPER-CASE variant of the probe file.
	upperProbe := filepath.Join(projectRoot, ".MOAI-FSCASE-PROBE")
	_, err := os.Stat(upperProbe)
	if err == nil {
		// Upper-case path found → filesystem is case-insensitive.
		return false, nil
	}
	if os.IsNotExist(err) {
		// Upper-case path not found → filesystem is case-sensitive.
		return true, nil
	}
	// Unexpected error → fallback to case-sensitive.
	return true, nil
}

// emitCleanupTelemetry writes the CleanupEvent as a JSON line to:
//  1. A dated file in <projectRoot>/.moai/logs/update-cleanup-{ts}.jsonl
//  2. stderr (mirror for observability)
//
// If the log file cannot be written (e.g., chmod 0500 on parent), it emits a
// [WARN] to stderr and continues without returning an error (REQ-UPC-022 D-02-03).
func emitCleanupTelemetry(projectRoot string, event CleanupEvent, stderr io.Writer) error {
	// Ensure slices are non-nil for consistent JSON output
	if event.PreUpdateSuffix2Files == nil {
		event.PreUpdateSuffix2Files = []string{}
	}
	if event.UserOptOutPaths == nil {
		event.UserOptOutPaths = []string{}
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal telemetry event: %w", err)
	}
	line := string(data) + "\n"

	// Mirror to stderr always
	_, _ = fmt.Fprint(stderr, line)

	// Persist to file
	logsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.LogsSubdir)
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		_, _ = fmt.Fprintf(stderr, "[WARN] telemetry persistence skipped: cannot create %s: %v\n",
			logsDir, err)
		return nil // not a fatal error (REQ-UPC-022)
	}

	ts := time.Now().UTC().Format("2006-01-02T15-04-05Z")
	logPath := filepath.Join(logsDir, "update-cleanup-"+ts+".jsonl")
	if err := os.WriteFile(logPath, []byte(line), 0o644); err != nil {
		_, _ = fmt.Fprintf(stderr, "[WARN] telemetry persistence skipped: %v\n", err)
		return nil // not a fatal error
	}
	return nil
}
