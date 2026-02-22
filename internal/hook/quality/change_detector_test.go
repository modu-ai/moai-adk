package quality

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestComputeHash_SHA256 verifies SHA-256 hash computation.
func TestComputeHash_SHA256(t *testing.T) {
	t.Run("computes consistent hash for same content", func(t *testing.T) {
		// Setup: Create temporary file with known content
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")
		content := []byte("package main\n\nfunc main() {}\n")

		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Act: Create detector and compute hash
		detector := NewChangeDetector()
		hash1, err1 := detector.ComputeHash(testFile)
		hash2, err2 := detector.ComputeHash(testFile)

		// Assert: No errors and hashes match
		if err1 != nil {
			t.Errorf("first ComputeHash failed: %v", err1)
		}
		if err2 != nil {
			t.Errorf("second ComputeHash failed: %v", err2)
		}
		if len(hash1) != 32 {
			t.Errorf("expected SHA-256 hash length of 32, got %d", len(hash1))
		}
		if len(hash2) != 32 {
			t.Errorf("expected SHA-256 hash length of 32, got %d", len(hash2))
		}

		// Hashes should be identical for same content
		for i := range hash1 {
			if hash1[i] != hash2[i] {
				t.Errorf("hashes differ at byte %d: %v != %v", i, hash1[i], hash2[i])
			}
		}
	})

	t.Run("computes different hash for different content", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")

		// Write first content
		content1 := []byte("package main\n\nfunc main() {}\n")
		if err := os.WriteFile(testFile, content1, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		detector := NewChangeDetector()
		hash1, err1 := detector.ComputeHash(testFile)
		if err1 != nil {
			t.Fatalf("first ComputeHash failed: %v", err1)
		}

		// Write different content
		time.Sleep(10 * time.Millisecond) // Ensure different mtime
		content2 := []byte("package main\n\nfunc main() { print(\"hello\") }\n")
		if err := os.WriteFile(testFile, content2, 0644); err != nil {
			t.Fatalf("failed to update test file: %v", err)
		}

		hash2, err2 := detector.ComputeHash(testFile)
		if err2 != nil {
			t.Fatalf("second ComputeHash failed: %v", err2)
		}

		// Hashes should differ
		same := true
		for i := range hash1 {
			if hash1[i] != hash2[i] {
				same = false
				break
			}
		}
		if same {
			t.Error("expected different hashes for different content, got same hash")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		detector := NewChangeDetector()
		_, err := detector.ComputeHash("/nonexistent/path/file.go")

		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})
}

// TestHasChanged verifies change detection logic.
func TestHasChanged(t *testing.T) {
	t.Run("detects no change when content is identical", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")
		content := []byte("package main\n\nfunc main() {}\n")

		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		detector := NewChangeDetector()
		originalHash, err := detector.ComputeHash(testFile)
		if err != nil {
			t.Fatalf("failed to compute original hash: %v", err)
		}

		// File hasn't changed
		changed, err := detector.HasChanged(testFile, originalHash)
		if err != nil {
			t.Errorf("HasChanged failed: %v", err)
		}
		if changed {
			t.Error("expected no change, but HasChanged returned true")
		}
	})

	t.Run("detects change when content is modified", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")
		content := []byte("package main\n\nfunc main() {}\n")

		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		detector := NewChangeDetector()
		originalHash, err := detector.ComputeHash(testFile)
		if err != nil {
			t.Fatalf("failed to compute original hash: %v", err)
		}

		// Modify file
		time.Sleep(10 * time.Millisecond)
		newContent := []byte("package main\n\nfunc main() { fmt.Println(\"hi\") }\n")
		if err := os.WriteFile(testFile, newContent, 0644); err != nil {
			t.Fatalf("failed to modify test file: %v", err)
		}

		changed, err := detector.HasChanged(testFile, originalHash)
		if err != nil {
			t.Errorf("HasChanged failed: %v", err)
		}
		if !changed {
			t.Error("expected change, but HasChanged returned false")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		detector := NewChangeDetector()
		someHash := make([]byte, 32)
		_, err := detector.HasChanged("/nonexistent/path/file.go", someHash)

		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})
}

// TestGetCachedHash verifies hash caching functionality.
func TestGetCachedHash(t *testing.T) {
	t.Run("returns cached hash after computation", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")
		content := []byte("package main\n\nfunc main() {}\n")

		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		detector := NewChangeDetector()
		hash1, err := detector.ComputeHash(testFile)
		if err != nil {
			t.Fatalf("ComputeHash failed: %v", err)
		}

		// Should retrieve cached hash
		cachedHash, found := detector.GetCachedHash(testFile)
		if !found {
			t.Error("expected to find cached hash, but found was false")
		}
		if len(cachedHash) != len(hash1) {
			t.Errorf("cached hash length mismatch: got %d, want %d", len(cachedHash), len(hash1))
		}
	})

	t.Run("returns not found for uncached file", func(t *testing.T) {
		detector := NewChangeDetector()
		_, found := detector.GetCachedHash("/some/uncached/file.go")
		if found {
			t.Error("expected not found for uncached file, but found was true")
		}
	})
}

// TestCacheHash verifies explicit hash caching.
func TestCacheHash(t *testing.T) {
	t.Run("caches and retrieves hash", func(t *testing.T) {
		detector := NewChangeDetector()
		testPath := "/some/test/file.go"
		testHash := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
			11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
			21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

		detector.CacheHash(testPath, testHash)

		cachedHash, found := detector.GetCachedHash(testPath)
		if !found {
			t.Fatal("expected to find cached hash, but found was false")
		}
		if len(cachedHash) != 32 {
			t.Errorf("cached hash length mismatch: got %d, want 32", len(cachedHash))
		}
		for i := range testHash {
			if cachedHash[i] != testHash[i] {
				t.Errorf("cached hash byte mismatch at %d: got %d, want %d", i, cachedHash[i], testHash[i])
			}
		}
	})
}

// TestThreadSafety verifies concurrent access is safe.
func TestThreadSafety(t *testing.T) {
	t.Run("concurrent hash computation is safe", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.go")
		content := []byte("package main\n\nfunc main() {}\n")

		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		detector := NewChangeDetector()
		done := make(chan bool, 10)

		// Run 10 concurrent computations
		for range 10 {
			go func() {
				_, err := detector.ComputeHash(testFile)
				if err != nil {
					t.Errorf("concurrent ComputeHash failed: %v", err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for range 10 {
			<-done
		}
	})

	t.Run("concurrent cache access is safe", func(t *testing.T) {
		detector := NewChangeDetector()
		done := make(chan bool, 10)

		// Run 5 concurrent caches and 5 concurrent reads
		for i := range 5 {
			go func(n int) {
				testPath := "/test/file.go"
				testHash := make([]byte, 32)
				testHash[0] = byte(n)
				detector.CacheHash(testPath, testHash)
				done <- true
			}(i)
		}

		for range 5 {
			go func() {
				detector.GetCachedHash("/test/file.go")
				done <- true
			}()
		}

		// Wait for all goroutines
		for range 10 {
			<-done
		}
	})
}
