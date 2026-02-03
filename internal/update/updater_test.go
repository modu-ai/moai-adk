package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdater_Download_Success(t *testing.T) {
	t.Parallel()

	binaryContent := []byte("fake binary content for testing")
	checksum := sha256Hex(binaryContent)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(binaryContent)
	}))
	defer ts.Close()

	dir := t.TempDir()
	u := NewUpdater(filepath.Join(dir, "moai"), http.DefaultClient)

	info := &VersionInfo{
		Version:  "v1.2.0",
		URL:      ts.URL + "/moai-darwin-arm64",
		Checksum: checksum,
	}

	path, err := u.Download(context.Background(), info)
	if err != nil {
		t.Fatalf("Download: %v", err)
	}

	// Verify file exists.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(data) != string(binaryContent) {
		t.Error("downloaded content mismatch")
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

	// Write new binary.
	newPath := filepath.Join(dir, "moai-new")
	if err := os.WriteFile(newPath, []byte("new binary"), 0o755); err != nil {
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
	if string(data) != "new binary" {
		t.Errorf("content = %q, want %q", string(data), "new binary")
	}

	// Verify permissions.
	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Error("binary should have execute permission")
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

// sha256Hex computes the hex-encoded SHA-256 hash of data.
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
