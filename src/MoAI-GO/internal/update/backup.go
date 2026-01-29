package update

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupManager handles backup creation and restoration
type BackupManager struct {
	projectDir string
}

// NewBackupManager creates a new backup manager
func NewBackupManager(projectDir string) *BackupManager {
	return &BackupManager{
		projectDir: projectDir,
	}
}

// CreateBackup creates a backup of .claude/ and .moai/ directories
func (bm *BackupManager) CreateBackup() (string, error) {
	// Create backup directory
	timestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(bm.projectDir, ".moai", "rollbacks", timestamp)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("error creating backup directory: %w", err)
	}

	// Backup .claude/
	claudeSrc := filepath.Join(bm.projectDir, ".claude")
	claudeDst := filepath.Join(backupDir, ".claude")
	if _, err := os.Stat(claudeSrc); err == nil {
		if err := copyDir(claudeSrc, claudeDst); err != nil {
			return "", fmt.Errorf("error backing up .claude/: %w", err)
		}
	}

	// Backup .moai/
	moaiSrc := filepath.Join(bm.projectDir, ".moai")
	moaiDst := filepath.Join(backupDir, ".moai")
	if _, err := os.Stat(moaiSrc); err == nil {
		if err := copyDir(moaiSrc, moaiDst); err != nil {
			return "", fmt.Errorf("error backing up .moai/: %w", err)
		}
	}

	return backupDir, nil
}

// Restore restores a backup from the specified directory
func (bm *BackupManager) Restore(backupDir string) error {
	// Restore .claude/
	claudeBackup := filepath.Join(backupDir, ".claude")
	claudeTarget := filepath.Join(bm.projectDir, ".claude")
	if _, err := os.Stat(claudeBackup); err == nil {
		// Remove existing .claude/
		_ = os.RemoveAll(claudeTarget)
		// Copy backup
		if err := copyDir(claudeBackup, claudeTarget); err != nil {
			return fmt.Errorf("error restoring .claude/: %w", err)
		}
	}

	// Restore .moai/
	moaiBackup := filepath.Join(backupDir, ".moai")
	moaiTarget := filepath.Join(bm.projectDir, ".moai")
	if _, err := os.Stat(moaiBackup); err == nil {
		// Remove existing .moai/
		_ = os.RemoveAll(moaiTarget)
		// Copy backup
		if err := copyDir(moaiBackup, moaiTarget); err != nil {
			return fmt.Errorf("error restoring .moai/: %w", err)
		}
	}

	return nil
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Construct destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
