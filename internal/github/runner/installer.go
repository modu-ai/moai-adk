// Package runner는 GitHub Actions self-hosted runner 설치 및 관리 기능을 제공합니다.
// Package runner provides installation and management for GitHub Actions self-hosted runners.
package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// HTTPClient는 HTTP GET 요청을 위한 인터페이스입니다.
// HTTPClient is an interface for HTTP GET requests.
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// Installer는 GitHub Actions runner 다운로드 및 설치를 담당합니다.
// Installer handles downloading and installing the GitHub Actions runner.
type Installer struct {
	ghRunnerDir string          // Runner 설치 디렉토리 (Runner installation directory)
	httpClient  HTTPClient      // HTTP 클라이언트 (HTTP client interface for testing)
}

// NewInstaller는 새로운 Installer 인스턴스를 생성합니다.
// NewInstaller creates a new Installer instance.
func NewInstaller(ghRunnerDir string, httpClient HTTPClient) *Installer {
	return &Installer{
		ghRunnerDir: ghRunnerDir,
		httpClient:  httpClient,
	}
}

// DownloadRunner는 지정된 OS 및 아키텍처용 runner를 다운로드하고 압축을 해제합니다.
// 네트워크 실패 시 최대 3회 재시도합니다 (REQ-CI-003.1).
// DownloadRunner downloads and extracts the runner for the specified OS and architecture.
// Retries up to 3 times on network failure (REQ-CI-003.1).
func (i *Installer) DownloadRunner(ctx context.Context, goos, arch string) error {
	// 기존 설치 확인 (REQ-CI-003.2)
	runnerDir := filepath.Join(i.ghRunnerDir, "actions-runner")
	if _, err := os.Stat(runnerDir); err == nil {

		// 이미 설치됨 - 현재는 에러 반환
		// 추후 사용자 프롬프트 인터페이스 추가 필요
		return fmt.Errorf("runner already installed in %s", runnerDir)
	}

	// 다운로드 URL 구성
	downloadURL := fmt.Sprintf(
		"https://github.com/actions/runner/releases/latest/download/actions-runner-%s-%s.tar.gz",
		goos, arch,
	)

	// 다운로드 파일 경로
	downloadPath := filepath.Join(i.ghRunnerDir, filepath.Base(downloadURL))

	// 재시도 로직 (최대 3회)
	var lastErr error
	for attempt := range 3 {
		if attempt > 0 {
			// Exponential backoff
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}

		lastErr = i.downloadWithRetry(ctx, downloadURL, downloadPath)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return fmt.Errorf("download failed after 3 attempts: %w", lastErr)
	}

	return nil
}

// downloadWithRetry는 단일 다운로드 시도를 수행합니다.
// downloadWithRetry performs a single download attempt.
func (i *Installer) downloadWithRetry(ctx context.Context, url, destPath string) error {
	// HTTP 클라이언트가 없으면 기본 http.Client 사용
	client := &http.Client{}
	if i.httpClient != nil {
		// Mock HTTPClient 인터페이스 사용
		resp, err := i.httpClient.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		// 파일 생성
		outFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer outFile.Close()

		// 다운로드 복사
		if _, err := io.Copy(outFile, resp.Body); err != nil {
			return fmt.Errorf("download error: %w", err)
		}

		return nil
	}

	// 기본 http.Client 사용
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// 파일 생성
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()

	// 다운로드 복사
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("download error: %w", err)
	}

	return nil
}

// VerifyChecksum은 다운로드된 파일의 SHA256 해시를 검증합니다.
// 불일치 시 파일을 삭제하고 에러를 반환합니다 (REQ-CI-003.3).
// VerifyChecksum verifies the SHA256 checksum of the downloaded file.
// Deletes the file and returns error on mismatch (REQ-CI-003.3).
func (i *Installer) VerifyChecksum(ctx context.Context, filePath, expectedHash string) error {
	// 파일 읽기
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// SHA256 계산
	hash := sha256.Sum256(data)
	actualHash := hex.EncodeToString(hash[:])

	// 검증
	if actualHash != expectedHash {
		// 불일치 시 파일 삭제
		os.Remove(filePath)
		return fmt.Errorf("checksum mismatch: expected %s, got %s (file deleted)", expectedHash, actualHash)
	}

	return nil
}
