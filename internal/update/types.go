// Package update provides self-update functionality for the MoAI-ADK binary.
//
// It implements ADR-009 (Self-Update via Binary Replacement) by checking
// GitHub Releases for new versions, downloading platform-specific binaries,
// verifying checksums, and performing atomic binary replacement with rollback.
package update

import (
	"context"
	"errors"
	"time"
)

// VersionInfo holds metadata about a GitHub Release.
type VersionInfo struct {
	Version  string    `json:"version"`
	URL      string    `json:"url"`
	Checksum string    `json:"checksum"`
	Date     time.Time `json:"date"`
}

// UpdateResult summarizes the outcome of an update operation.
type UpdateResult struct {
	PreviousVersion string
	NewVersion      string
	FilesUpdated    int
	FilesMerged     int
	FilesConflicted int
	FilesSkipped    int
	RollbackPath    string
}

// Checker queries GitHub Releases API for version information.
type Checker interface {
	// CheckLatest fetches the latest release metadata.
	CheckLatest(ctx context.Context) (*VersionInfo, error)

	// IsUpdateAvailable compares current version against latest.
	// Returns (true, info, nil) if an update is available,
	// (false, nil, nil) if already up to date.
	IsUpdateAvailable(current string) (bool, *VersionInfo, error)
}

// Updater handles binary download and replacement.
type Updater interface {
	// Download fetches the platform binary to a temp file and verifies its checksum.
	// Returns the path to the downloaded file.
	Download(ctx context.Context, version *VersionInfo) (string, error)

	// Replace atomically replaces the current binary with the new one.
	Replace(ctx context.Context, newBinaryPath string) error
}

// Rollback provides backup and restore capabilities.
type Rollback interface {
	// CreateBackup copies the current binary to a timestamped backup path.
	// Returns the backup file path.
	CreateBackup() (string, error)

	// Restore copies the backup back to the original binary location.
	Restore(backupPath string) error
}

// Orchestrator coordinates the full update pipeline.
type Orchestrator interface {
	// Update executes the complete update workflow.
	Update(ctx context.Context) (*UpdateResult, error)
}

// Sentinel errors for the update package.
var (
	// ErrUpdateNotAvail indicates no newer version is available.
	ErrUpdateNotAvail = errors.New("update: no update available")

	// ErrDownloadFailed indicates the binary download failed.
	ErrDownloadFailed = errors.New("update: binary download failed")

	// ErrChecksumMismatch indicates checksum verification failed.
	ErrChecksumMismatch = errors.New("update: checksum verification failed")

	// ErrReplaceFailed indicates binary replacement failed.
	ErrReplaceFailed = errors.New("update: binary replacement failed")

	// ErrRollbackFailed indicates rollback restoration failed.
	ErrRollbackFailed = errors.New("update: rollback restoration failed")
)
