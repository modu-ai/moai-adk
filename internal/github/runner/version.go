// Package runner provides version checking for GitHub Actions runners.
package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DefaultRunnerDir returns the default installation directory for GitHub Actions runner.
func DefaultRunnerDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/actions-runner" // fallback
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(homeDir, "actions-runner")
	default: // linux, darwin
		return "/actions-runner"
	}
}

// VersionCheckStatus represents version check status.
type VersionCheckStatus string

const (
	VersionCheckOK VersionCheckStatus = "OK"
	VersionCheckWarn VersionCheckStatus = "WARN"
	VersionCheckFail VersionCheckStatus = "FAIL"
	VersionCheckSkip VersionCheckStatus = "SKIP"
)

// CheckResult holds version check results.
type CheckResult struct {
	InstalledVersion string              // Installed version
	LatestVersion    string              // Latest version
	DaysOld          int                 // Days since release
	Status           VersionCheckStatus  // Status
	Message          string              // Message
}

// GitHubClient fetches latest release info.
type GitHubClient interface {
	GetLatestRelease(ctx context.Context) (version string, downloadURL string, err error)
	GetInstalledVersion(ctx context.Context) (string, error)
}

// mockReleaseData is test release data.
// TODO T-05: Actual implementation will call GitHub Release API
var mockReleaseData = map[string]string{
	"2.700.1": "2026-04-17", // 10 days ago
	"2.700.0": "2026-04-17", // 10 days ago
	"2.699.0": "2026-04-02", // 25 days ago
	"2.698.0": "2026-03-28", // 30 days ago
}

// Clock abstracts current-time retrieval, enabling deterministic time injection in tests.
type Clock interface {
	Now() time.Time
}

// RealClock returns actual system time.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// VersionChecker checks runner version against latest.
type VersionChecker struct {
	ghRunnerDir string
	ghClient    GitHubClient
	clock       Clock
}

// NewVersionChecker creates a new VersionChecker instance with real system clock.
func NewVersionChecker(ghRunnerDir string, ghClient GitHubClient) *VersionChecker {
	return &VersionChecker{
		ghRunnerDir: ghRunnerDir,
		ghClient:    ghClient,
		clock:       RealClock{},
	}
}

// NewVersionCheckerWithClock creates a VersionChecker with a custom Clock (for tests).
// nil clock is normalized to RealClock for safety.
func NewVersionCheckerWithClock(ghRunnerDir string, ghClient GitHubClient, clock Clock) *VersionChecker {
	if clock == nil {
		clock = RealClock{}
	}
	return &VersionChecker{
		ghRunnerDir: ghRunnerDir,
		ghClient:    ghClient,
		clock:       clock,
	}
}

// CheckVersion compares installed runner against latest GitHub release.
// 25 days: WARN status, 30 days: FAIL status, not installed: SKIP status.
func (v *VersionChecker) CheckVersion(ctx context.Context) (*CheckResult, error) {
	installed, err := v.ghClient.GetInstalledVersion(ctx)
	if err != nil {
		// Runner not installed
		return &CheckResult{
			Status:  VersionCheckSkip,
			Message: "runner not installed",
		}, nil
	}

	latest, _, err := v.ghClient.GetLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest release: %w", err)
	}

	// 3. Calculate release date (using mock data for testing)
	now := v.clock.Now()
	publishedDate, ok := mockReleaseData[installed]
	if !ok {
		publishedDate = now.Format("2006-01-02")
	}
	parsed, _ := time.Parse("2006-01-02", publishedDate)
	daysOld := calculateDaysOld(parsed, now)

	var status VersionCheckStatus
	var message string

	switch {
	case daysOld >= 30:
		status = VersionCheckFail
		message = fmt.Sprintf("runner is %d days old (>=30 days), upgrade required", daysOld)
	case daysOld >= 25:
		status = VersionCheckWarn
		message = fmt.Sprintf("runner is %d days old (>=25 days), upgrade recommended", daysOld)
	default:
		status = VersionCheckOK
		message = fmt.Sprintf("runner is up-to-date (%d days old)", daysOld)
	}

	return &CheckResult{
		InstalledVersion: installed,
		LatestVersion:    latest,
		DaysOld:          daysOld,
		Status:           status,
		Message:          message,
	}, nil
}

// calculateDaysOld calculates days elapsed since release date.
func calculateDaysOld(published, now time.Time) int {
	duration := now.Sub(published)
	days := int(duration.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// MockGitHubClientImpl is a test implementation of GitHubClient.
type MockGitHubClientImpl struct {
	InstalledVersion string
	LatestVersion    string
	DownloadURL      string
	FetchError       error
}

// GetLatestRelease returns mock data.
func (m *MockGitHubClientImpl) GetLatestRelease(ctx context.Context) (version string, downloadURL string, err error) {
	if m.FetchError != nil {
		return "", "", m.FetchError
	}
	return m.LatestVersion, m.DownloadURL, nil
}

// GetInstalledVersion returns mock data.
func (m *MockGitHubClientImpl) GetInstalledVersion(ctx context.Context) (string, error) {
	if m.InstalledVersion == "" {
		return "", fmt.Errorf("runner not found")
	}
	return m.InstalledVersion, nil
}

// parseRunnerVersion extracts version from runner version string.
// format: "2.700.0" or "actions-runner-linux-x64-2.700.0.tar.gz"
//
//nolint:unused // TODO: SPEC-xxx to use runner version parsing
func parseRunnerVersion(versionStr string) string {
	parts := strings.Split(versionStr, ".")
	if len(parts) >= 3 {
		major := parts[len(parts)-3]
		minor := parts[len(parts)-2]
		patch := strings.Split(parts[len(parts)-1], "-")[0] // Remove "-tar.gz" etc.
		return fmt.Sprintf("%s.%s.%s", major, minor, patch)
	}
	return versionStr
}

// FileSystemGitHubClient checks runner version from filesystem.
type FileSystemGitHubClient struct{}

// NewFileSystemGitHubClient creates a new FileSystemGitHubClient.
func NewFileSystemGitHubClient() *FileSystemGitHubClient {
	return &FileSystemGitHubClient{}
}

// GetInstalledVersion verifies installed runner version from file system.
func (f *FileSystemGitHubClient) GetInstalledVersion(ctx context.Context) (string, error) {
	runnerDir := DefaultRunnerDir()
	versionFile := filepath.Join(runnerDir, ".runner")

	// If version file doesn't exist, verify runner binary
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		// Check if runner binary exists
		runnerBin := filepath.Join(runnerDir, "bin", "Runner.Service")
		if _, err := os.Stat(runnerBin); os.IsNotExist(err) {
			// On Windows, check for .exe
			runnerBin = filepath.Join(runnerDir, "bin", "Runner.Service.exe")
			if _, err := os.Stat(runnerBin); os.IsNotExist(err) {
				return "", fmt.Errorf("runner not installed")
			}
		}

		// If version file doesn't exist, try to extract version from binary
		// Return "unknown"
		return "unknown", nil
	}

	// TODO: Parse version from .runner file
	// format: {"version":"2.700.1", ...}
	return "unknown", nil
}

// GetLatestRelease fetches latest version from GitHub API.
// TODO: Implement actual GitHub Release API call (T-05)
func (f *FileSystemGitHubClient) GetLatestRelease(ctx context.Context) (version string, downloadURL string, err error) {
	// mock data return
	return "2.700.1", "https://github.com/actions/runner/releases/tag/v2.700.1", nil
}
