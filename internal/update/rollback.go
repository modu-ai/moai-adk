package update

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// rollbackImpl is the concrete implementation of Rollback.
type rollbackImpl struct {
	binaryPath string
}

// NewRollback creates a Rollback that manages backups for the given binary path.
func NewRollback(binaryPath string) Rollback {
	return &rollbackImpl{binaryPath: binaryPath}
}

// CreateBackup copies the current binary to a timestamped backup file.
// The backup preserves the original file permissions.
func (r *rollbackImpl) CreateBackup() (string, error) {
	backupPath := fmt.Sprintf("%s.backup.%d", r.binaryPath, time.Now().Unix())

	if err := copyFile(r.binaryPath, backupPath); err != nil {
		return "", fmt.Errorf("rollback: create backup: %w", err)
	}

	return backupPath, nil
}

// Restore copies the backup file back to the original binary location.
// On Windows, a two-step rename approach is used to avoid "Access is denied"
// errors that occur when writing to a running executable.
func (r *rollbackImpl) Restore(backupPath string) error {
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("%w: backup not found at %s: %v", ErrRollbackFailed, backupPath, err)
	}

	if runtime.GOOS == "windows" {
		if err := r.restoreOnWindows(backupPath); err != nil {
			return err
		}
	} else {
		if err := copyFile(backupPath, r.binaryPath); err != nil {
			return fmt.Errorf("%w: restore from %s: %v", ErrRollbackFailed, backupPath, err)
		}
	}

	// Ensure execute permission on restored binary.
	if err := os.Chmod(r.binaryPath, 0o755); err != nil {
		return fmt.Errorf("%w: chmod after restore: %v", ErrRollbackFailed, err)
	}

	return nil
}

// restoreOnWindows performs a Windows-safe rollback restoration.
// It renames the (potentially locked) current binary away before copying
// the backup into the original path.
func (r *rollbackImpl) restoreOnWindows(backupPath string) error {
	// Step 1: Rename the current (failed) binary to a temporary path.
	// Windows allows renaming a running executable.
	// Note: there is a brief TOCTOU window between steps 1 and 2 where the
	// binary path is empty. Any concurrent process attempting to exec the binary
	// would fail. Exploiting this requires write access to the binary directory.
	failedPath := fmt.Sprintf("%s.failed-%d", r.binaryPath, time.Now().UnixNano())
	if err := os.Rename(r.binaryPath, failedPath); err != nil {
		// If the current binary does not exist (e.g. it was already removed),
		// continue and attempt the copy regardless.
		if !os.IsNotExist(err) {
			return fmt.Errorf("%w: windows: rename failed binary away: %v", ErrRollbackFailed, err)
		}
	}

	// Step 2: Copy (not rename) the backup to the original path.
	// We use copyFile instead of os.Rename so the backup file is preserved
	// even if the copy fails mid-way, enabling manual recovery.
	if err := copyFile(backupPath, r.binaryPath); err != nil {
		// Best-effort: try to put the failed binary back.
		if rerr := os.Rename(failedPath, r.binaryPath); rerr != nil {
			slog.Error("binary restore and recovery both failed; binary may be missing",
				"binary_path", r.binaryPath,
				"backup_at", failedPath,
				"restore_error", err,
				"recovery_error", rerr)
		}
		return fmt.Errorf("%w: restore from %s: %v", ErrRollbackFailed, backupPath, err)
	}

	// Step 3: Opportunistically remove the failed binary.
	if err := os.Remove(failedPath); err != nil {
		slog.Debug("could not remove orphaned binary", "path", failedPath, "error", err)
	}

	return nil
}

// copyFile copies src to dst, preserving file permissions.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close destination: %w", closeErr)
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}

	return nil
}
