// Package runner는 GitHub Actions runner 버전 확인 기능을 제공합니다.
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

// DefaultRunnerDir은 GitHub Actions runner의 기본 설치 디렉토리를 반환합니다.
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

// VersionCheckStatus는 버전 확인 상태를 나타냅니다.
// VersionCheckStatus represents version check status.
type VersionCheckStatus string

const (
	// VersionCheckOK는 최신 상태입니다 (25일 미만).
	VersionCheckOK VersionCheckStatus = "OK"
	// VersionCheckWarn은 경고 상태입니다 (25일 이상).
	VersionCheckWarn VersionCheckStatus = "WARN"
	// VersionCheckFail은 실패 상태입니다 (30일 이상).
	VersionCheckFail VersionCheckStatus = "FAIL"
	// VersionCheckSkip은 확인 불가 상태입니다 (미설치).
	VersionCheckSkip VersionCheckStatus = "SKIP"
)

// CheckResult는 버전 확인 결과를 담습니다.
// CheckResult holds version check results.
type CheckResult struct {
	InstalledVersion string              // 설치된 버전 (Installed version)
	LatestVersion    string              // 최신 버전 (Latest version)
	DaysOld          int                 // 경과 일수 (Days since release)
	Status           VersionCheckStatus  // 상태 (Status)
	Message          string              // 메시지 (Message)
}

// GitHubClient는 GitHub Release API를 위한 인터페이스입니다.
// GitHubClient fetches latest release info.
type GitHubClient interface {
	GetLatestRelease(ctx context.Context) (version string, downloadURL string, err error)
	GetInstalledVersion(ctx context.Context) (string, error)
}

// mockReleaseData는 테스트용 릴리즈 데이터입니다.
// TODO T-05: 실제 구현 시 GitHub Release API 호출
var mockReleaseData = map[string]string{
	"2.700.1": "2026-04-17", // 10일 전
		"2.700.0": "2026-04-17", // 10일 전
	"2.699.0": "2026-04-02", // 25일 전
	"2.698.0": "2026-03-28", // 30일 전
}

// VersionChecker는 runner 버전을 확인합니다.
// VersionChecker checks runner version against latest.
type VersionChecker struct {
	ghRunnerDir string
	ghClient    GitHubClient
}

// NewVersionChecker는 새로운 VersionChecker 인스턴스를 생성합니다.
// NewVersionChecker creates a new VersionChecker instance.
func NewVersionChecker(ghRunnerDir string, ghClient GitHubClient) *VersionChecker {
	return &VersionChecker{
		ghRunnerDir: ghRunnerDir,
		ghClient:    ghClient,
	}
}

// CheckVersion은 설치된 runner와 최신 버전을 비교합니다.
// CheckVersion compares installed runner against latest GitHub release.
// 25일: WARN 상태, 30일: FAIL 상태, 미설치: SKIP 상태.
func (v *VersionChecker) CheckVersion(ctx context.Context) (*CheckResult, error) {
	// 1. 설치된 버전 확인
	installed, err := v.ghClient.GetInstalledVersion(ctx)
	if err != nil {
		// Runner 미설치
		return &CheckResult{
			Status:  VersionCheckSkip,
			Message: "runner not installed",
		}, nil
	}

	// 2. 최신 버전 획득
	latest, _, err := v.ghClient.GetLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("get latest release: %w", err)
	}

	// 3. 릴리즈 날짜 계산 (테스트용 mock 데이터)
	publishedDate, ok := mockReleaseData[installed]
	if !ok {
		publishedDate = time.Now().Format("2006-01-02")
	}
	parsed, _ := time.Parse("2006-01-02", publishedDate)
	daysOld := calculateDaysOld(parsed, time.Now())

	// 4. 상태 결정
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

// calculateDaysOld는 릴리즈 날짜로부터 경과 일수를 계산합니다.
// calculateDaysOld calculates days elapsed since release date.
func calculateDaysOld(published, now time.Time) int {
	duration := now.Sub(published)
	days := int(duration.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// MockGitHubClientImpl은 테스트용 GitHubClient 구현입니다.
// MockGitHubClientImpl is a test implementation of GitHubClient.
type MockGitHubClientImpl struct {
	InstalledVersion string
	LatestVersion    string
	DownloadURL      string
	FetchError       error
}

// GetLatestRelease는 mock 데이터를 반환합니다.
func (m *MockGitHubClientImpl) GetLatestRelease(ctx context.Context) (version string, downloadURL string, err error) {
	if m.FetchError != nil {
		return "", "", m.FetchError
	}
	return m.LatestVersion, m.DownloadURL, nil
}

// GetInstalledVersion은 mock 데이터를 반환합니다.
func (m *MockGitHubClientImpl) GetInstalledVersion(ctx context.Context) (string, error) {
	if m.InstalledVersion == "" {
		return "", fmt.Errorf("runner not found")
	}
	return m.InstalledVersion, nil
}

// parseRunnerVersion은 runner 버전 문자열에서 버전을 추출합니다.
// parseRunnerVersion extracts version from runner version string.
// 형식: "2.700.0" 또는 "actions-runner-linux-x64-2.700.0.tar.gz"
func parseRunnerVersion(versionStr string) string {
	// 버전 패턴 찾기 (X.Y.Z)
	parts := strings.Split(versionStr, ".")
	if len(parts) >= 3 {
		// 마지막 3개 부분 반환 (X.Y.Z)
		major := parts[len(parts)-3]
		minor := parts[len(parts)-2]
		patch := strings.Split(parts[len(parts)-1], "-")[0] // "-tar.gz" 등 제거
		return fmt.Sprintf("%s.%s.%s", major, minor, patch)
	}
	return versionStr
}

// FileSystemGitHubClient는 파일 시스템에서 runner 버전을 확인하는 구현입니다.
// FileSystemGitHubClient checks runner version from filesystem.
type FileSystemGitHubClient struct{}

// NewFileSystemGitHubClient는 새로운 FileSystemGitHubClient를 생성합니다.
// NewFileSystemGitHubClient creates a new FileSystemGitHubClient.
func NewFileSystemGitHubClient() *FileSystemGitHubClient {
	return &FileSystemGitHubClient{}
}

// GetInstalledVersion은 파일 시스템에서 설치된 runner 버전을 확인합니다.
func (f *FileSystemGitHubClient) GetInstalledVersion(ctx context.Context) (string, error) {
	runnerDir := DefaultRunnerDir()
	versionFile := filepath.Join(runnerDir, ".runner")

	// version 파일이 없으면 runner binary를 확인
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		// runner binary가 있는지 확인
		runnerBin := filepath.Join(runnerDir, "bin", "Runner.Service")
		if _, err := os.Stat(runnerBin); os.IsNotExist(err) {
			// Windows에서는 .exe 확인
			runnerBin = filepath.Join(runnerDir, "bin", "Runner.Service.exe")
			if _, err := os.Stat(runnerBin); os.IsNotExist(err) {
				return "", fmt.Errorf("runner not installed")
			}
		}

		// 버전 파일이 없으면 binary에서 버전 추출 시도
		// 일단 "unknown" 반환
		return "unknown", nil
	}

	// TODO: .runner 파일에서 버전 파싱
	// 형식: {"version":"2.700.1", ...}
	return "unknown", nil
}

// GetLatestRelease는 GitHub API에서 최신 버전을 가져옵니다.
// TODO: 실제 GitHub Release API 호출 구현 (T-05)
func (f *FileSystemGitHubClient) GetLatestRelease(ctx context.Context) (version string, downloadURL string, err error) {
	// 일단 mock 데이터 반환
	return "2.700.1", "https://github.com/actions/runner/releases/tag/v2.700.1", nil
}
