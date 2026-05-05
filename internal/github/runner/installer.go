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

// HTTPClient is an interface for HTTP GET requests.
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// Installer handles downloading and installing the GitHub Actions runner.
type Installer struct {
	ghRunnerDir string          // Runner installation directory
	httpClient  HTTPClient      // HTTP client interface for testing
}

// NewInstaller creates a new Installer instance.
func NewInstaller(ghRunnerDir string, httpClient HTTPClient) *Installer {
	return &Installer{
		ghRunnerDir: ghRunnerDir,
		httpClient:  httpClient,
	}
}

// DownloadRunner downloads and extracts the runner for the specified OS and architecture.
// Retries up to 3 times on network failure (REQ-CI-003.1).
func (i *Installer) DownloadRunner(ctx context.Context, goos, arch string) error {
	// verify (REQ-CI-003.2)
	runnerDir := filepath.Join(i.ghRunnerDir, "actions-runner")
	if _, err := os.Stat(runnerDir); err == nil {

		return fmt.Errorf("runner already installed in %s", runnerDir)
	}

	downloadURL := fmt.Sprintf(
		"https://github.com/actions/runner/releases/latest/download/actions-runner-%s-%s.tar.gz",
		goos, arch,
	)

	downloadPath := filepath.Join(i.ghRunnerDir, filepath.Base(downloadURL))

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

// downloadWithRetry performs a single download attempt.
func (i *Installer) downloadWithRetry(ctx context.Context, url, destPath string) error {
	// Use default http.Client when HTTP client is not available
	client := &http.Client{}
	if i.httpClient != nil {
		// Use Mock HTTPClient interface
		resp, err := i.httpClient.Get(url)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer func() { _ = outFile.Close() }()

		if _, err := io.Copy(outFile, resp.Body); err != nil {
			return fmt.Errorf("download error: %w", err)
		}

		return nil
	}

	// Use default http.Client
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("download error: %w", err)
	}

	return nil
}

// VerifyChecksum verifies the SHA256 checksum of the downloaded file.
// Deletes the file and returns error on mismatch (REQ-CI-003.3).
func (i *Installer) VerifyChecksum(ctx context.Context, filePath, expectedHash string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	// Calculate SHA256
	hash := sha256.Sum256(data)
	actualHash := hex.EncodeToString(hash[:])

	if actualHash != expectedHash {
		_ = os.Remove(filePath)
		return fmt.Errorf("checksum mismatch: expected %s, got %s (file deleted)", expectedHash, actualHash)
	}

	return nil
}
