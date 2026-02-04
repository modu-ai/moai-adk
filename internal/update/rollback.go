package update

import (
	"fmt"
	"io"
	"os"
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
func (r *rollbackImpl) Restore(backupPath string) error {
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("%w: backup not found at %s: %v", ErrRollbackFailed, backupPath, err)
	}

	if err := copyFile(backupPath, r.binaryPath); err != nil {
		return fmt.Errorf("%w: restore from %s: %v", ErrRollbackFailed, backupPath, err)
	}

	// Ensure execute permission on restored binary.
	if err := os.Chmod(r.binaryPath, 0o755); err != nil {
		return fmt.Errorf("%w: chmod after restore: %v", ErrRollbackFailed, err)
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
