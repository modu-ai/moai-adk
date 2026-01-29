// Package cli provides the self-update command for MoAI-ADK
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	PublishedAt string `json:"published_at"`
}

// NewSelfUpdateCommand creates the self-update command
func NewSelfUpdateCommand() *cobra.Command {
	var checkOnly bool
	var version string

	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Update moai to the latest version",
		Long: `Check for updates and download the latest version of moai.

This command queries GitHub Releases for the latest version and downloads
the appropriate binary for your platform. If no update is available, it
will report that you're already on the latest version.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSelfUpdate(checkOnly, version)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check-only", false, "only check for updates, don't download")
	cmd.Flags().StringVar(&version, "version", "", "specific version to update to (default: latest)")

	return cmd
}

// runSelfUpdate performs the self-update process
func runSelfUpdate(checkOnly bool, targetVersion string) error {
	fmt.Println("Checking for updates...")

	// Get current version
	currentVersion := getCurrentVersion()
	fmt.Printf("Current version: %s\n", currentVersion)

	// Fetch latest release info
	release, err := fetchLatestRelease(targetVersion)
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	fmt.Printf("Latest version: %s\n", latestVersion)

	// Check if update is needed
	if currentVersion == latestVersion {
		fmt.Println("Already up to date!")
		return nil
	}

	// Simple version comparison (dev is always newer than release versions)
	if currentVersion != "dev" && latestVersion != "dev" {
		// Both are version strings, we can compare
		if compareVersions(currentVersion, latestVersion) > 0 {
			fmt.Printf("Current version (%s) is newer than latest release (%s)\n", currentVersion, latestVersion)
			return nil
		}
	}

	fmt.Printf("Update available: %s -> %s\n", currentVersion, latestVersion)

	if checkOnly {
		return nil
	}

	// Download update
	return downloadUpdate(release.TagName)
}

// getCurrentVersion returns the current version
func getCurrentVersion() string {
	// Try to read from version info embedded in binary
	// For now, return a placeholder
	if version := os.Getenv("MOAI_VERSION"); version != "" {
		return version
	}
	return "dev"
}

// fetchLatestRelease fetches the latest release from GitHub
func fetchLatestRelease(version string) (*GitHubRelease, error) {
	var url string
	if version == "" || version == "latest" {
		url = "https://api.github.com/repos/anthropics/moai-go/releases/latest"
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/anthropics/moai-go/releases/tags/%s", version)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "moai-self-update")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadUpdate downloads and installs the update
func downloadUpdate(tagName string) error {
	fmt.Printf("Downloading version %s...\n", tagName)

	// Detect platform
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	binaryName := fmt.Sprintf("moai-%s", platform)

	// Windows needs .exe extension
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// Download URL
	downloadURL := fmt.Sprintf("https://github.com/anthropics/moai-go/releases/download/%s/%s", tagName, binaryName)

	fmt.Printf("Downloading from: %s\n", downloadURL)

	// Download to temp file
	tempFile := filepath.Join(os.TempDir(), binaryName)
	if err := downloadFile(downloadURL, tempFile); err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer func() { _ = os.Remove(tempFile) }()

	// Make executable
	if err := os.Chmod(tempFile, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Get current binary path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current binary path: %w", err)
	}

	// Backup current binary
	backupPath := execPath + ".backup"
	if err := copyFile(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Replace binary
	if err := os.Rename(tempFile, execPath); err != nil {
		// Rollback on failure
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Remove backup on success
	_ = os.Remove(backupPath)

	fmt.Printf("Successfully updated to %s\n", tagName)
	return nil
}

// downloadFile downloads a file from URL to destination path
func downloadFile(url, dest string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "moai-self-update")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	// Copy downloaded content
	_, err = io.Copy(out, resp.Body)
	return err
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// compareVersions compares two version strings
// Returns 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	// Simple semantic version comparison
	// Split by "."
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1Num, v2Num int

		if i < len(v1Parts) {
			v1Num, _ = strconv.Atoi(v1Parts[i])
		}
		if i < len(v2Parts) {
			v2Num, _ = strconv.Atoi(v2Parts[i])
		}

		if v1Num > v2Num {
			return 1
		}
		if v1Num < v2Num {
			return -1
		}
	}

	return 0
}
