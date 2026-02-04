package update

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalConfig holds configuration for local file-based updates.
type LocalConfig struct {
	// ReleasesDir is the directory containing local release files.
	// Defaults to ~/.moai/releases/
	ReleasesDir string

	// CurrentVersion is the version string of the running binary.
	CurrentVersion string
}

// localChecker implements Checker for local file-based releases.
type localChecker struct {
	config LocalConfig
}

// NewLocalChecker creates a Checker that reads releases from a local directory.
//
// The expected directory structure:
//
//	~/.moai/releases/
//	├── moai-2.0.0-darwin-arm64      # Platform binary
//	├── moai-2.0.0-darwin-arm64.sha256 # Checksum file
//	├── version.json                  # Version metadata
//	└── LATEST                        # Symlink to latest version dir (optional)
//
// The version.json format:
//
//	{
//	  "version": "2.0.0",
//	  "date": "2026-02-04T10:00:00Z",
//	  "platform": "darwin-arm64",
//	  "binary": "moai-2.0.0-darwin-arm64"
//	}
func NewLocalChecker(config LocalConfig) Checker {
	if config.ReleasesDir == "" {
		homeDir, err := os.UserHomeDir()
		// If home dir cannot be determined, use current directory
		if err != nil {
			homeDir = "."
		}
		config.ReleasesDir = filepath.Join(homeDir, ".moai", "releases")
	}
	return &localChecker{config: config}
}

// CheckLatest reads the latest version from the local releases directory.
func (c *localChecker) CheckLatest(ctx context.Context) (*VersionInfo, error) {
	// Check for version.json in releases directory
	versionFile := filepath.Join(c.config.ReleasesDir, "version.json")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local checker: no release found at %s", versionFile)
		}
		return nil, fmt.Errorf("local checker: read version file: %w", err)
	}

	var localVersion localVersionInfo
	if err := json.Unmarshal(data, &localVersion); err != nil {
		return nil, fmt.Errorf("local checker: parse version file: %w", err)
	}

	// Check if the platform binary exists
	binaryPath := filepath.Join(c.config.ReleasesDir, localVersion.Binary)
	if _, err := os.Stat(binaryPath); err != nil {
		return nil, fmt.Errorf("local checker: binary not found: %s", binaryPath)
	}

	// Parse date
	publishDate, err := time.Parse(time.RFC3339, localVersion.Date)
	if err != nil {
		publishDate = time.Now()
	}

	info := &VersionInfo{
		Version:  localVersion.Version,
		URL:      "file://" + binaryPath,
		Date:     publishDate,
		Checksum: "", // Local releases may not have checksums
	}

	// Check for checksum file
	checksumPath := binaryPath + ".sha256"
	if data, err := os.ReadFile(checksumPath); err == nil {
		info.Checksum = strings.TrimSpace(string(data))
	}

	return info, nil
}

// IsUpdateAvailable compares the current version against the local release.
func (c *localChecker) IsUpdateAvailable(current string) (bool, *VersionInfo, error) {
	info, err := c.CheckLatest(context.Background())
	if err != nil {
		return false, nil, err
	}

	// For dev versions, compare file modification times
	// since version strings may be identical (e.g., commit-hash-dirty)
	if c.isDevVersion(current) {
		// Get current binary path and modification time
		currentBinary, err := os.Executable()
		if err == nil {
			currentInfo, statErr := os.Stat(currentBinary)
			releaseBinaryPath := strings.TrimPrefix(info.URL, "file://")
			releaseInfo, releaseStatErr := os.Stat(releaseBinaryPath)

			// If release binary is newer (by mtime), an update is available
			if statErr == nil && releaseStatErr == nil {
				if releaseInfo.ModTime().After(currentInfo.ModTime()) {
					return true, info, nil
				}
				// Same version and release is older or same age - no update needed
				return false, nil, nil
			}
		}
		// If we can't compare mtimes for dev versions with same version string,
		// assume no update needed (avoid unnecessary updates)
		if info.Version == current {
			return false, nil, nil
		}
		// Different version strings - proceed to version comparison
	}

	cmp := compareSemver(info.Version, current)
	if cmp <= 0 {
		return false, nil, nil
	}

	return true, info, nil
}

// isDevVersion checks if the version string indicates a dev build.
func (c *localChecker) isDevVersion(v string) bool {
	return strings.Contains(v, "dirty") ||
		strings.Contains(v, "dev") ||
		strings.Contains(v, "none") ||
		!strings.HasPrefix(v, "v") && !strings.Contains(v, ".")
}

// localVersionInfo represents the contents of version.json.
type localVersionInfo struct {
	Version  string `json:"version"`
	Date     string `json:"date"`
	Platform string `json:"platform"`
	Binary   string `json:"binary"`
}

// localUpdater implements Updater for local file-based releases.
type localUpdater struct {
	releasesDir string
	binaryPath  string
}

// NewLocalUpdater creates an Updater for local file-based releases.
func NewLocalUpdater(releasesDir, binaryPath string) Updater {
	return &localUpdater{
		releasesDir: releasesDir,
		binaryPath:  binaryPath,
	}
}

// Download copies the local binary to a temp location.
func (u *localUpdater) Download(ctx context.Context, version *VersionInfo) (string, error) {
	// Extract local path from file:// URL
	binaryPath := strings.TrimPrefix(version.URL, "file://")

	// For local updates, just return the source path directly
	// The Replace step will handle the copy
	return binaryPath, nil
}

// Replace copies the local binary to the target location.
func (u *localUpdater) Replace(ctx context.Context, newBinaryPath string) error {
	// Read source binary
	data, err := os.ReadFile(newBinaryPath)
	if err != nil {
		return fmt.Errorf("local updater: read source binary: %w", err)
	}

	// Write to target location
	if err := os.WriteFile(u.binaryPath, data, 0755); err != nil {
		return fmt.Errorf("local updater: write target binary: %w", err)
	}

	return nil
}
