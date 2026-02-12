package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUpdater_Download_Success(t *testing.T) {
	t.Parallel()

	// Use platform-specific binary name
	binaryName := "moai"
	if runtime.GOOS == "windows" {
		binaryName = "moai.exe"
	}

	binaryContent := []byte("fake binary content for testing")
	archiveData := createTarGz(t, binaryName, binaryContent)
	checksum := sha256Hex(archiveData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(archiveData)
	}))
	defer ts.Close()

	dir := t.TempDir()
	u := NewUpdater(filepath.Join(dir, "moai"), http.DefaultClient)

	info := &VersionInfo{
		Version:  "v1.2.0",
		URL:      ts.URL + "/moai-darwin-arm64.tar.gz",
		Checksum: checksum,
	}

	path, err := u.Download(context.Background(), info)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}

	// Verify extracted binary content.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(data) != string(binaryContent) {
		t.Error("extracted binary content mismatch")
	}

	// Verify no archive temp files remain.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if e.Name() != filepath.Base(path) {
			name := e.Name()
			if filepath.Ext(name) == ".tmp" && name != filepath.Base(path) {
				t.Errorf("archive temp file not cleaned up: %s", name)
			}
		}
	}
}

func TestUpdater_Download_ChecksumMismatch(t *testing.T) {
	t.Parallel()

	binaryContent := []byte("real content")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(binaryContent)
	}))
	defer ts.Close()

	dir := t.TempDir()
	u := NewUpdater(filepath.Join(dir, "moai"), http.DefaultClient)

	info := &VersionInfo{
		Version:  "v1.2.0",
		URL:      ts.URL + "/binary",
		Checksum: "wrong_checksum_value",
	}

	path, err := u.Download(context.Background(), info)
	if err == nil {
		t.Fatal("expected error for checksum mismatch")
	}
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Errorf("expected ErrChecksumMismatch, got: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path on error, got %q", path)
	}
}

func TestUpdater_Download_ServerError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	dir := t.TempDir()
	u := NewUpdater(filepath.Join(dir, "moai"), http.DefaultClient)

	info := &VersionInfo{
		Version: "v1.2.0",
		URL:     ts.URL + "/binary",
	}

	_, err := u.Download(context.Background(), info)
	if err == nil {
		t.Error("expected error for server error")
	}
}

func TestUpdater_Download_ContextCancelled(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slow response.
		select {}
	}))
	defer ts.Close()

	dir := t.TempDir()
	u := NewUpdater(filepath.Join(dir, "moai"), http.DefaultClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	info := &VersionInfo{
		Version: "v1.2.0",
		URL:     ts.URL + "/binary",
	}

	_, err := u.Download(ctx, info)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestUpdater_Download_CleanupOnFailure(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("content"))
	}))
	defer ts.Close()

	dir := t.TempDir()
	u := NewUpdater(filepath.Join(dir, "moai"), http.DefaultClient)

	info := &VersionInfo{
		Version:  "v1.2.0",
		URL:      ts.URL + "/binary",
		Checksum: "wrong",
	}

	_, _ = u.Download(context.Background(), info)

	// Verify no temp files remain.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file not cleaned up: %s", e.Name())
		}
	}
}

func TestUpdater_Replace_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	// Write old binary.
	if err := os.WriteFile(binaryPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}

	// Write new binary with valid Mach-O 64-bit magic bytes.
	newPath := filepath.Join(dir, "moai-new")
	newContent := append([]byte{0xcf, 0xfa, 0xed, 0xfe}, []byte("new binary payload")...)
	if err := os.WriteFile(newPath, newContent, 0o755); err != nil {
		t.Fatalf("write new: %v", err)
	}

	u := NewUpdater(binaryPath, http.DefaultClient)
	if err := u.Replace(context.Background(), newPath); err != nil {
		t.Fatalf("Replace: %v", err)
	}

	// Verify new content.
	data, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("read replaced: %v", err)
	}
	if !bytes.Equal(data, newContent) {
		t.Errorf("content mismatch after replace")
	}

	// Verify permissions (skip on Windows).
	if runtime.GOOS != "windows" {
		info, err := os.Stat(binaryPath)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if info.Mode().Perm()&0o111 == 0 {
			t.Error("binary should have execute permission")
		}
	}
}

func TestUpdater_Replace_NonexistentNewBinary(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")
	if err := os.WriteFile(binaryPath, []byte("old"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	u := NewUpdater(binaryPath, http.DefaultClient)
	err := u.Replace(context.Background(), "/nonexistent/new-binary")
	if err == nil {
		t.Error("expected error for nonexistent new binary")
	}
}

func TestUpdater_Replace_RejectsGzipArchive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	// Write old binary.
	if err := os.WriteFile(binaryPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}

	// Write a gzip file as "new binary".
	newPath := filepath.Join(dir, "moai-new")
	gzipContent := []byte{0x1f, 0x8b, 0x08, 0x00} // gzip magic bytes
	if err := os.WriteFile(newPath, gzipContent, 0o755); err != nil {
		t.Fatalf("write gzip: %v", err)
	}

	u := NewUpdater(binaryPath, http.DefaultClient)
	err := u.Replace(context.Background(), newPath)
	if err == nil {
		t.Fatal("expected error when replacing with gzip archive")
	}
	if !errors.Is(err, ErrReplaceFailed) {
		t.Errorf("expected ErrReplaceFailed, got: %v", err)
	}

	// Verify original binary is unchanged.
	data, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	if string(data) != "old binary" {
		t.Error("original binary was modified")
	}
}

func TestUpdater_Replace_RejectsZipArchive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	// Write old binary.
	if err := os.WriteFile(binaryPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}

	// Write a zip file as "new binary".
	newPath := filepath.Join(dir, "moai-new")
	zipContent := []byte{0x50, 0x4b, 0x03, 0x04} // zip magic bytes (PK)
	if err := os.WriteFile(newPath, zipContent, 0o755); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	u := NewUpdater(binaryPath, http.DefaultClient)
	err := u.Replace(context.Background(), newPath)
	if err == nil {
		t.Fatal("expected error when replacing with zip archive")
	}
	if !errors.Is(err, ErrReplaceFailed) {
		t.Errorf("expected ErrReplaceFailed, got: %v", err)
	}
}

func TestUpdater_Replace_RejectsUnknownFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	// Write old binary.
	if err := os.WriteFile(binaryPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}

	// Write a file with unknown magic bytes.
	newPath := filepath.Join(dir, "moai-new")
	unknownContent := []byte{0xde, 0xad, 0xbe, 0xef}
	if err := os.WriteFile(newPath, unknownContent, 0o755); err != nil {
		t.Fatalf("write unknown: %v", err)
	}

	u := NewUpdater(binaryPath, http.DefaultClient)
	err := u.Replace(context.Background(), newPath)
	if err == nil {
		t.Fatal("expected error when replacing with unknown format")
	}
	if !errors.Is(err, ErrReplaceFailed) {
		t.Errorf("expected ErrReplaceFailed, got: %v", err)
	}
}

func TestUpdater_Replace_AcceptsELF(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	// Write old binary.
	if err := os.WriteFile(binaryPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}

	// Write a valid ELF binary.
	newPath := filepath.Join(dir, "moai-new")
	elfContent := append([]byte{0x7f, 0x45, 0x4c, 0x46}, []byte("elf payload")...)
	if err := os.WriteFile(newPath, elfContent, 0o755); err != nil {
		t.Fatalf("write elf: %v", err)
	}

	u := NewUpdater(binaryPath, http.DefaultClient)
	if err := u.Replace(context.Background(), newPath); err != nil {
		t.Fatalf("Replace should accept ELF binary: %v", err)
	}
}

func TestUpdater_Replace_AcceptsPE(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	// Write old binary.
	if err := os.WriteFile(binaryPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}

	// Write a valid PE (Windows) binary.
	newPath := filepath.Join(dir, "moai-new")
	peContent := append([]byte{0x4d, 0x5a}, []byte("pe payload")...)
	if err := os.WriteFile(newPath, peContent, 0o755); err != nil {
		t.Fatalf("write pe: %v", err)
	}

	u := NewUpdater(binaryPath, http.DefaultClient)
	if err := u.Replace(context.Background(), newPath); err != nil {
		t.Fatalf("Replace should accept PE binary: %v", err)
	}
}

func TestUpdater_Replace_AcceptsMachOUniversal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	// Write old binary.
	if err := os.WriteFile(binaryPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}

	// Write a valid Mach-O Universal Binary (Fat Binary).
	newPath := filepath.Join(dir, "moai-new")
	fatContent := append([]byte{0xca, 0xfe, 0xba, 0xbe}, []byte("fat payload")...)
	if err := os.WriteFile(newPath, fatContent, 0o755); err != nil {
		t.Fatalf("write fat: %v", err)
	}

	u := NewUpdater(binaryPath, http.DefaultClient)
	if err := u.Replace(context.Background(), newPath); err != nil {
		t.Fatalf("Replace should accept Mach-O Universal binary: %v", err)
	}
}

func TestUpdater_Replace_RejectsTooSmallFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "moai")

	// Write old binary.
	if err := os.WriteFile(binaryPath, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}

	// Write a file too small to have valid magic bytes.
	newPath := filepath.Join(dir, "moai-new")
	if err := os.WriteFile(newPath, []byte{0x00}, 0o755); err != nil {
		t.Fatalf("write tiny: %v", err)
	}

	u := NewUpdater(binaryPath, http.DefaultClient)
	err := u.Replace(context.Background(), newPath)
	if err == nil {
		t.Fatal("expected error when replacing with too-small file")
	}
	if !errors.Is(err, ErrReplaceFailed) {
		t.Errorf("expected ErrReplaceFailed, got: %v", err)
	}
}

// sha256Hex computes the hex-encoded SHA-256 hash of data.
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// createTarGz creates a .tar.gz archive containing a single file with the given name and content.
func createTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("tar write header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("tar write content: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	return buf.Bytes()
}
