package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashFile(t *testing.T) {
	t.Run("known_hash_hello_world", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "hello.txt")
		content := []byte("hello world")
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		got, err := HashFile(path)
		if err != nil {
			t.Fatalf("HashFile error: %v", err)
		}

		// Known SHA-256 of "hello world"
		h := sha256.Sum256(content)
		want := hashPrefix + hex.EncodeToString(h[:])

		if got != want {
			t.Errorf("HashFile = %q, want %q", got, want)
		}
	})

	t.Run("sha256_prefix", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		got, err := HashFile(path)
		if err != nil {
			t.Fatalf("HashFile error: %v", err)
		}

		if !strings.HasPrefix(got, "sha256:") {
			t.Errorf("HashFile result %q does not start with 'sha256:'", got)
		}
	})

	t.Run("empty_file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")
		if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		got, err := HashFile(path)
		if err != nil {
			t.Fatalf("HashFile error: %v", err)
		}

		// SHA-256 of empty data
		h := sha256.Sum256([]byte{})
		want := hashPrefix + hex.EncodeToString(h[:])

		if got != want {
			t.Errorf("HashFile(empty) = %q, want %q", got, want)
		}
	})

	t.Run("large_file_10MB", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "large.bin")

		// Create a 10MB file
		data := make([]byte, 10*1024*1024)
		for i := range data {
			data[i] = byte(i % 256)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		got, err := HashFile(path)
		if err != nil {
			t.Fatalf("HashFile error: %v", err)
		}

		// Compute expected hash
		h := sha256.Sum256(data)
		want := hashPrefix + hex.EncodeToString(h[:])

		if got != want {
			t.Errorf("HashFile(large) = %q, want %q", got, want)
		}
	})

	t.Run("nonexistent_file", func(t *testing.T) {
		got, err := HashFile("/nonexistent/path/file.txt")
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
		if got != "" {
			t.Errorf("expected empty string for nonexistent file, got %q", got)
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("expected os.ErrNotExist in error chain, got: %v", err)
		}
	})

	t.Run("deterministic_same_content", func(t *testing.T) {
		dir := t.TempDir()
		path1 := filepath.Join(dir, "file1.txt")
		path2 := filepath.Join(dir, "file2.txt")
		content := []byte("same content")

		if err := os.WriteFile(path1, content, 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}
		if err := os.WriteFile(path2, content, 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		hash1, err := HashFile(path1)
		if err != nil {
			t.Fatalf("HashFile(1) error: %v", err)
		}
		hash2, err := HashFile(path2)
		if err != nil {
			t.Fatalf("HashFile(2) error: %v", err)
		}

		if hash1 != hash2 {
			t.Errorf("same content produced different hashes: %q vs %q", hash1, hash2)
		}
	})

	t.Run("different_content_different_hash", func(t *testing.T) {
		dir := t.TempDir()
		path1 := filepath.Join(dir, "file1.txt")
		path2 := filepath.Join(dir, "file2.txt")

		if err := os.WriteFile(path1, []byte("content A"), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}
		if err := os.WriteFile(path2, []byte("content B"), 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		hash1, err := HashFile(path1)
		if err != nil {
			t.Fatalf("HashFile(1) error: %v", err)
		}
		hash2, err := HashFile(path2)
		if err != nil {
			t.Fatalf("HashFile(2) error: %v", err)
		}

		if hash1 == hash2 {
			t.Error("different content produced same hash")
		}
	})
}

func TestHashBytes(t *testing.T) {
	t.Run("known_hash_hello_world", func(t *testing.T) {
		data := []byte("hello world")
		got := HashBytes(data)

		h := sha256.Sum256(data)
		want := hashPrefix + hex.EncodeToString(h[:])

		if got != want {
			t.Errorf("HashBytes = %q, want %q", got, want)
		}
	})

	t.Run("sha256_prefix", func(t *testing.T) {
		got := HashBytes([]byte("test"))
		if !strings.HasPrefix(got, "sha256:") {
			t.Errorf("HashBytes result %q does not start with 'sha256:'", got)
		}
	})

	t.Run("empty_bytes", func(t *testing.T) {
		got := HashBytes([]byte{})
		h := sha256.Sum256([]byte{})
		want := hashPrefix + hex.EncodeToString(h[:])

		if got != want {
			t.Errorf("HashBytes(empty) = %q, want %q", got, want)
		}
	})

	t.Run("consistency_with_HashFile", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		content := []byte("hello world")

		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("WriteFile error: %v", err)
		}

		fileHash, err := HashFile(path)
		if err != nil {
			t.Fatalf("HashFile error: %v", err)
		}

		bytesHash := HashBytes(content)

		if fileHash != bytesHash {
			t.Errorf("HashFile = %q, HashBytes = %q; want equal", fileHash, bytesHash)
		}
	})

	t.Run("nil_bytes", func(t *testing.T) {
		got := HashBytes(nil)
		// SHA-256 of nil is same as SHA-256 of empty
		h := sha256.Sum256(nil)
		want := hashPrefix + hex.EncodeToString(h[:])

		if got != want {
			t.Errorf("HashBytes(nil) = %q, want %q", got, want)
		}
	})
}
