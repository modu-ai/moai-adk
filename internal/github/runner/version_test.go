// Package runner는 GitHub Actions runner 버전 확인 기능을 테스트합니다.
// Package runner provides tests for GitHub Actions runner version checking.
package runner

import (
	"context"
	"testing"
	"time"
)

// fixedTestNow는 시간 종속적 버전 검증 테스트의 결정적 기준 시각입니다.
// fixedTestNow is the deterministic reference time for time-sensitive version tests.
// Choice rationale: 2026-04-27 makes mockReleaseData entries hit threshold boundaries:
//   - 2.700.0 (2026-04-17) -> 10 days (OK)
//   - 2.699.0 (2026-04-02) -> 25 days (WARN)
//   - 2.698.0 (2026-03-28) -> 30 days (FAIL)
var fixedTestNow = time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)

// fakeClock는 테스트용 결정적 Clock 구현입니다.
// fakeClock is a deterministic Clock implementation for tests.
type fakeClock struct{ now time.Time }

// Now는 고정된 시각을 반환합니다.
func (f fakeClock) Now() time.Time { return f.now }

// MockGitHubClient는 테스트용 GitHubClient 인터페이스 구현입니다.
// MockGitHubClient is a test implementation of GitHubClient interface.
type MockGitHubClient struct {
	// GetLatestReleaseFunc는 GetLatestRelease 호출 시 실행됩니다.
	GetLatestReleaseFunc func(ctx context.Context) (version string, downloadURL string, err error)

	// GetInstalledVersionFunc는 GetInstalledVersion 호출 시 실행됩니다.
	GetInstalledVersionFunc func(ctx context.Context) (string, error)
}

// GetLatestRelease는 GetLatestReleaseFunc를 실행합니다.
func (m *MockGitHubClient) GetLatestRelease(ctx context.Context) (version string, downloadURL string, err error) {
	if m.GetLatestReleaseFunc == nil {
		return "", "", nil
	}
	return m.GetLatestReleaseFunc(ctx)
}

// GetInstalledVersion은 GetInstalledVersionFunc를 실행합니다.
func (m *MockGitHubClient) GetInstalledVersion(ctx context.Context) (string, error) {
	if m.GetInstalledVersionFunc == nil {
		return "", nil
	}
	return m.GetInstalledVersionFunc(ctx)
}

// TestVersionChecker_CheckVersion_OK는 최신 버전 상태를 테스트합니다.
func TestVersionChecker_CheckVersion_OK(t *testing.T) {
	t.Parallel()

	mockClient := &MockGitHubClient{
		GetInstalledVersionFunc: func(ctx context.Context) (string, error) {
			return "2.700.0", nil
		},
		GetLatestReleaseFunc: func(ctx context.Context) (string, string, error) {
			return "2.700.1", "https://github.com/actions/runner/releases/tag/v2.700.1", nil
		},
	}

	checker := NewVersionCheckerWithClock("/tmp/actions-runner", mockClient, fakeClock{now: fixedTestNow})

	result, err := checker.CheckVersion(context.Background())
	if err != nil {
		t.Fatalf("CheckVersion failed: %v", err)
	}

	if result.Status != VersionCheckOK {
		t.Errorf("Expected VersionCheckOK, got %s", result.Status)
	}

	// fixedTestNow(2026-04-27) - mockReleaseData["2.700.0"](2026-04-17) = 10 days
	wantDays := calculateDaysOld(parseDate("2026-04-17"), fixedTestNow)
	if result.DaysOld != wantDays {
		t.Errorf("Expected %d days old, got %d", wantDays, result.DaysOld)
	}
}

// TestVersionChecker_CheckVersion_Warn25는 25일 경고 상태를 테스트합니다.
func TestVersionChecker_CheckVersion_Warn25(t *testing.T) {
	t.Parallel()

	mockClient := &MockGitHubClient{
		GetInstalledVersionFunc: func(ctx context.Context) (string, error) {
			return "2.699.0", nil
		},
		GetLatestReleaseFunc: func(ctx context.Context) (string, string, error) {
			return "2.700.1", "https://github.com/actions/runner/releases/tag/v2.700.1", nil
		},
	}

	checker := NewVersionCheckerWithClock("/tmp/actions-runner", mockClient, fakeClock{now: fixedTestNow})

	result, err := checker.CheckVersion(context.Background())
	if err != nil {
		t.Fatalf("CheckVersion failed: %v", err)
	}

	if result.Status != VersionCheckWarn {
		t.Errorf("Expected VersionCheckWarn (25 days), got %s", result.Status)
	}

	// fixedTestNow(2026-04-27) - mockReleaseData["2.699.0"](2026-04-02) = 25 days
	wantDays := calculateDaysOld(parseDate("2026-04-02"), fixedTestNow)
	if result.DaysOld != wantDays {
		t.Errorf("Expected %d days old, got %d", wantDays, result.DaysOld)
	}
}

// TestVersionChecker_CheckVersion_Fail30은 30일 실패 상태를 테스트합니다.
func TestVersionChecker_CheckVersion_Fail30(t *testing.T) {
	t.Parallel()

	mockClient := &MockGitHubClient{
		GetInstalledVersionFunc: func(ctx context.Context) (string, error) {
			return "2.698.0", nil
		},
		GetLatestReleaseFunc: func(ctx context.Context) (string, string, error) {
			return "2.700.1", "https://github.com/actions/runner/releases/tag/v2.700.1", nil
		},
	}

	checker := NewVersionCheckerWithClock("/tmp/actions-runner", mockClient, fakeClock{now: fixedTestNow})

	result, err := checker.CheckVersion(context.Background())
	if err != nil {
		t.Fatalf("CheckVersion failed: %v", err)
	}

	if result.Status != VersionCheckFail {
		t.Errorf("Expected VersionCheckFail (30 days), got %s", result.Status)
	}

	// fixedTestNow(2026-04-27) - mockReleaseData["2.698.0"](2026-03-28) = 30 days
	wantDays := calculateDaysOld(parseDate("2026-03-28"), fixedTestNow)
	if result.DaysOld != wantDays {
		t.Errorf("Expected %d days old, got %d", wantDays, result.DaysOld)
	}
}

// TestVersionChecker_CheckVersion_NotInstalled는 미설치 상태를 테스트합니다.
func TestVersionChecker_CheckVersion_NotInstalled(t *testing.T) {
	t.Parallel()

	mockClient := &MockGitHubClient{
		GetInstalledVersionFunc: func(ctx context.Context) (string, error) {
			return "", context.Canceled
		},
	}

	checker := NewVersionChecker("/tmp/actions-runner", mockClient)

	result, err := checker.CheckVersion(context.Background())
	if err != nil {
		t.Fatalf("CheckVersion failed: %v", err)
	}

	if result.Status != VersionCheckSkip {
		t.Errorf("Expected VersionCheckSkip (not installed), got %s", result.Status)
	}
}

// TestVersionChecker_CheckVersion_FetchError는 버전 획득 실패를 테스트합니다.
func TestVersionChecker_CheckVersion_FetchError(t *testing.T) {
	t.Parallel()

	mockClient := &MockGitHubClient{
		GetInstalledVersionFunc: func(ctx context.Context) (string, error) {
			return "2.700.0", nil
		},
		GetLatestReleaseFunc: func(ctx context.Context) (string, string, error) {
			return "", "", context.DeadlineExceeded
		},
	}

	checker := NewVersionChecker("/tmp/actions-runner", mockClient)

	_, err := checker.CheckVersion(context.Background())
	if err == nil {
		t.Fatal("Expected error for fetch failure, got nil")
	}
}

// TestNewVersionChecker는 생성자를 테스트합니다.
func TestNewVersionChecker(t *testing.T) {
	t.Parallel()

	mockClient := &MockGitHubClient{}
	checker := NewVersionChecker("/test/dir", mockClient)

	if checker == nil {
		t.Fatal("NewVersionChecker returned nil")
	}

	if checker.ghRunnerDir != "/test/dir" {
		t.Errorf("Expected ghRunnerDir=/test/dir, got %s", checker.ghRunnerDir)
	}

	if checker.ghClient != mockClient {
		t.Error("GitHubClient not set correctly")
	}
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

// TestVersionChecker_CalculateDaysOld는 날짜 계산을 테스트합니다.
func TestVersionChecker_CalculateDaysOld(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		published string
		wantDays  int
	}{
		{"today", "2026-04-27", 0},
		{"yesterday", "2026-04-26", 1},
		{"10 days", "2026-04-17", 10},
		{"25 days", "2026-04-02", 25},
		{"30 days", "2026-03-28", 30},
		{"100 days", "2026-01-17", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, _ := time.Parse("2006-01-02", tt.published)
			got := calculateDaysOld(parsed, now)
			if got != tt.wantDays {
				t.Errorf("calculateDaysOld() = %d, want %d", got, tt.wantDays)
			}
		})
	}
}
