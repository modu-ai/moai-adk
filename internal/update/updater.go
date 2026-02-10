package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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

	// Extract binary from archive.
	binaryPath, err := u.extractBinary(tmpPath)
	if err != nil {
		return "", fmt.Errorf("%w: extract binary: %v", ErrDownloadFailed, err)
	}

	// Clean up the archive file; the extracted binary is all we need.
	_ = os.Remove(tmpPath)

	success = true
	return binaryPath, nil
}

// Replace atomically replaces the current binary with the new one.
// It validates binary format, sets execute permissions and uses os.Rename for atomicity.
func (u *updaterImpl) Replace(ctx context.Context, newBinaryPath string) error {
	// Verify the new binary exists.
	if _, err := os.Stat(newBinaryPath); err != nil {
		return fmt.Errorf("%w: new binary not found: %v", ErrReplaceFailed, err)
	}

	// Validate binary format before replacing.
	if err := validateBinaryFormat(newBinaryPath); err != nil {
		return err
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

// extractBinary detects the archive format and extracts the moai binary.
// It returns the path to a temp file containing the extracted binary.
func (u *updaterImpl) extractBinary(archivePath string) (string, error) {
	binaryName := "moai"
	if runtime.GOOS == "windows" {
		binaryName = "moai.exe"
	}

	// Detect format via magic bytes.
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("open archive: %w", err)
	}

	var magic [2]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("read magic bytes: %w", err)
	}
	_ = f.Close()

	switch {
	case magic[0] == 0x1f && magic[1] == 0x8b: // gzip
		return u.extractFromTarGz(archivePath, binaryName)
	case magic[0] == 0x50 && magic[1] == 0x4b: // zip (PK)
		return u.extractFromZip(archivePath, binaryName)
	default:
		return "", fmt.Errorf("unsupported archive format (magic: %x %x)", magic[0], magic[1])
	}
}

// extractFromTarGz extracts the named binary from a .tar.gz archive.
func (u *updaterImpl) extractFromTarGz(archivePath, binaryName string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("open tar.gz: %w", err)
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar next: %w", err)
		}

		// Match the binary by base name (archive may have directory prefixes).
		if filepath.Base(hdr.Name) == binaryName && hdr.Typeflag == tar.TypeReg {
			return u.writeExtractedBinary(tr)
		}
	}

	return "", fmt.Errorf("binary %q not found in tar.gz archive", binaryName)
}

// extractFromZip extracts the named binary from a .zip archive.
func (u *updaterImpl) extractFromZip(archivePath, binaryName string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	for _, zf := range zr.File {
		if filepath.Base(zf.Name) == binaryName && !zf.FileInfo().IsDir() {
			rc, err := zf.Open()
			if err != nil {
				return "", fmt.Errorf("open zip entry %q: %w", zf.Name, err)
			}
			defer func() { _ = rc.Close() }()

			return u.writeExtractedBinary(rc)
		}
	}

	return "", fmt.Errorf("binary %q not found in zip archive", binaryName)
}

// validateBinaryFormat checks that the file at path is a valid executable binary
// and not a compressed archive. This is a safety net to prevent corrupted binaries
// from being installed.
func validateBinaryFormat(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file for validation: %w", err)
	}
	defer func() { _ = f.Close() }()

	var magic [4]byte
	n, err := io.ReadFull(f, magic[:])
	if err != nil || n < 2 {
		return fmt.Errorf("%w: file too small or unreadable", ErrReplaceFailed)
	}

	// Reject known archive formats (the most common mistake).
	switch {
	case magic[0] == 0x1f && magic[1] == 0x8b:
		return fmt.Errorf("%w: file is a gzip archive, not an executable. Run: curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash", ErrReplaceFailed)
	case magic[0] == 0x50 && magic[1] == 0x4b:
		return fmt.Errorf("%w: file is a zip archive, not an executable. Run: curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash", ErrReplaceFailed)
	}

	// Accept known executable formats.
	switch {
	case magic[0] == 0x7f && magic[1] == 0x45 && magic[2] == 0x4c && magic[3] == 0x46:
		// ELF (Linux)
		return nil
	case magic[0] == 0xcf && magic[1] == 0xfa && magic[2] == 0xed && magic[3] == 0xfe:
		// Mach-O 64-bit (macOS arm64/amd64)
		return nil
	case magic[0] == 0xce && magic[1] == 0xfa && magic[2] == 0xed && magic[3] == 0xfe:
		// Mach-O 32-bit (macOS legacy)
		return nil
	case magic[0] == 0xfe && magic[1] == 0xed && magic[2] == 0xfa && magic[3] == 0xcf:
		// Mach-O 64-bit big-endian
		return nil
	case magic[0] == 0xfe && magic[1] == 0xed && magic[2] == 0xfa && magic[3] == 0xce:
		// Mach-O 32-bit big-endian
		return nil
	case magic[0] == 0xca && magic[1] == 0xfe && magic[2] == 0xba && magic[3] == 0xbe:
		// Mach-O Universal Binary (Fat Binary)
		return nil
	case magic[0] == 0x4d && magic[1] == 0x5a:
		// PE (Windows .exe)
		return nil
	}

	return fmt.Errorf("%w: unrecognized binary format (magic: %x). Run: curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash", ErrReplaceFailed, magic[:n])
}

// writeExtractedBinary writes the binary content from r to a temp file
// in the same directory as the target binary.
func (u *updaterImpl) writeExtractedBinary(r io.Reader) (string, error) {
	dir := filepath.Dir(u.binaryPath)
	tmp, err := os.CreateTemp(dir, ".moai-extracted-*.tmp")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, r); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("write binary: %w", err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("close temp file: %w", err)
	}

	return tmpPath, nil
}
