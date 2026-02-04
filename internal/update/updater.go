package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// updaterImpl is the concrete implementation of Updater.
type updaterImpl struct {
	binaryPath string
	client     *http.Client
}

// NewUpdater creates an Updater for the given binary path.
func NewUpdater(binaryPath string, client *http.Client) Updater {
	if client == nil {
		client = http.DefaultClient
	}
	return &updaterImpl{
		binaryPath: binaryPath,
		client:     client,
	}
}

// Download fetches the platform binary to a temp file and verifies its checksum.
// On checksum mismatch or any error, the temp file is cleaned up.
func (u *updaterImpl) Download(ctx context.Context, version *VersionInfo) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, version.URL, nil)
	if err != nil {
		return "", fmt.Errorf("%w: create request: %v", ErrDownloadFailed, err)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: unexpected status %d", ErrDownloadFailed, resp.StatusCode)
	}

	// Create temp file in the same directory as the binary for atomic rename.
	dir := filepath.Dir(u.binaryPath)
	tmpFile, err := os.CreateTemp(dir, ".moai-download-*.tmp")
	if err != nil {
		return "", fmt.Errorf("%w: create temp file: %v", ErrDownloadFailed, err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on any error path.
	success := false
	defer func() {
		if !success {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	// Download with checksum computation.
	hasher := sha256.New()
	writer := io.MultiWriter(tmpFile, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return "", fmt.Errorf("%w: copy data: %v", ErrDownloadFailed, err)
	}

	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("%w: close temp file: %v", ErrDownloadFailed, err)
	}

	// Verify checksum if provided.
	if version.Checksum != "" {
		gotChecksum := hex.EncodeToString(hasher.Sum(nil))
		if gotChecksum != version.Checksum {
			return "", fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, version.Checksum, gotChecksum)
		}
	}

	success = true
	return tmpPath, nil
}

// Replace atomically replaces the current binary with the new one.
// It sets execute permissions and uses os.Rename for atomicity.
func (u *updaterImpl) Replace(ctx context.Context, newBinaryPath string) error {
	// Verify the new binary exists.
	if _, err := os.Stat(newBinaryPath); err != nil {
		return fmt.Errorf("%w: new binary not found: %v", ErrReplaceFailed, err)
	}

	// Set execute permission on new binary.
	if err := os.Chmod(newBinaryPath, 0o755); err != nil {
		return fmt.Errorf("%w: chmod: %v", ErrReplaceFailed, err)
	}

	// Atomic rename (works when src and dst are on the same filesystem).
	if err := os.Rename(newBinaryPath, u.binaryPath); err != nil {
		return fmt.Errorf("%w: rename: %v", ErrReplaceFailed, err)
	}

	return nil
}
