package quality

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChangeDetector_ComputeHash(t *testing.T) {
	t.Parallel()

	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Hello, World!")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	detector := NewChangeDetector()

	hash1, err := detector.ComputeHash(testFile)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	if len(hash1) == 0 {
		t.Error("hash is empty")
	}

	// Hash should be consistent (SHA-256 = 32 bytes)
	if len(hash1) != 32 {
		t.Errorf("hash length = %d, want 32", len(hash1))
	}

	// Computing again should give same result
	hash2, err := detector.ComputeHash(testFile)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	if string(hash1) != string(hash2) {
		t.Error("hashes are not consistent")
	}
}

func TestChangeDetector_ComputeHash_CacheHit(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Cached content")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	detector := NewChangeDetector()

	// First call - computes and caches
	hash1, err := detector.ComputeHash(testFile)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	// Second call - should use cache
	hash2, err := detector.ComputeHash(testFile)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	if string(hash1) != string(hash2) {
		t.Error("cached hash differs from original")
	}
}

func TestChangeDetector_ComputeHash_NonExistentFile(t *testing.T) {
	t.Parallel()

	detector := NewChangeDetector()
	hash, err := detector.ComputeHash("/nonexistent/file/path.txt")

	if err != nil {
		t.Errorf("expected no error for nonexistent file, got: %v", err)
	}

	if len(hash) != 0 {
		t.Errorf("expected empty hash for nonexistent file, got length %d", len(hash))
	}
}

func TestChangeDetector_HasChanged(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Original content")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	detector := NewChangeDetector()

	// Get initial hash
	hashBefore, err := detector.ComputeHash(testFile)
	if err != nil {
		t.Fatalf("ComputeHash failed: %v", err)
	}

	// File hasn't changed
	changed, err := detector.HasChanged(testFile, hashBefore)
	if err != nil {
		t.Fatalf("HasChanged failed: %v", err)
	}
	if changed {
		t.Error("file hasn't changed but HasChanged returned true")
	}

	// Modify the file
	newContent := []byte("Modified content")
	if err := os.WriteFile(testFile, newContent, 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}

	// Clear cache to force recomputation
	detector.ClearCache()

	// Now file should be detected as changed
	changed, err = detector.HasChanged(testFile, hashBefore)
	if err != nil {
		t.Fatalf("HasChanged failed: %v", err)
	}
	if !changed {
		t.Error("file has changed but HasChanged returned false")
	}
}

func TestChangeDetector_GetCachedHash(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Cache test")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	detector := NewChangeDetector()

	// No cache initially
	_, found := detector.GetCachedHash(testFile)
	if found {
		t.Error("expected no cached hash initially")
	}

	// Compute hash (caches it)
	hash, _ := detector.ComputeHash(testFile)

	// Now should find in cache
	cachedHash, found := detector.GetCachedHash(testFile)
	if !found {
		t.Error("expected to find cached hash")
	}
	if string(cachedHash) != string(hash) {
		t.Error("cached hash differs from computed hash")
	}
}

func TestChangeDetector_CacheHash(t *testing.T) {
	t.Parallel()

	detector := NewChangeDetector()
	testPath := "/test/path.txt"
	testHash := []byte{1, 2, 3, 4, 5}

	detector.CacheHash(testPath, testHash)

	cached, found := detector.GetCachedHash(testPath)
	if !found {
		t.Error("expected to find cached hash")
	}

	if string(cached) != string(testHash) {
		t.Error("cached hash doesn't match")
	}
}

func TestChangeDetector_ClearExpired(t *testing.T) {
	t.Parallel()

	detector := NewChangeDetector()
	testFile := filepath.Join(t.TempDir(), "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// Compute hash (caches with TTL)
	detector.ComputeHash(testFile)

	// Verify cache exists
	_, found := detector.GetCachedHash(testFile)
	if !found {
		t.Error("expected to find cached hash")
	}

	// Manually expire by setting a very short TTL and waiting
	// Since we can't modify TTL, just test the method doesn't crash
	detector.ClearExpired()
}

func TestChangeDetector_ClearCache(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Clear test")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	detector := NewChangeDetector()

	// Populate cache
	detector.ComputeHash(testFile)

	// Verify cache exists
	_, found := detector.GetCachedHash(testFile)
	if !found {
		t.Error("expected to find cached hash")
	}

	// Clear cache
	detector.ClearCache()

	// Verify cache is gone
	_, found = detector.GetCachedHash(testFile)
	if found {
		t.Error("expected cache to be cleared")
	}
}

func TestHashToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		hash []byte
		want string
	}{
		{
			name: "empty hash",
			hash: []byte{},
			want: "",
		},
		{
			name: "simple hash",
			hash: []byte{0x01, 0x02, 0x03, 0x0a, 0xff},
			want: "0102030aff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashToString(tt.hash)
			if got != tt.want {
				t.Errorf("HashToString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChangeDetector_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Concurrent test")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	detector := NewChangeDetector()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			detector.ComputeHash(testFile)
			detector.GetCachedHash(testFile)
			detector.ClearExpired()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify detector still works
	hash, err := detector.ComputeHash(testFile)
	if err != nil {
		t.Errorf("ComputeHash failed after concurrent access: %v", err)
	}
	if len(hash) == 0 {
		t.Error("hash is empty after concurrent access")
	}
}

func TestChangeDetector_HashConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{"empty", ""},
		{"short", "Hello"},
		{"long", strings.Repeat("a", 10000)},
		{"special", "Hello\nWorld\t!@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.txt")
			content := []byte(tt.content)

			if err := os.WriteFile(testFile, content, 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			detector := NewChangeDetector()

			hash1, err := detector.ComputeHash(testFile)
			if err != nil {
				t.Fatalf("ComputeHash failed: %v", err)
			}

			hash2, err := detector.ComputeHash(testFile)
			if err != nil {
				t.Fatalf("ComputeHash failed: %v", err)
			}

			if string(hash1) != string(hash2) {
				t.Error("hashes are inconsistent")
			}

			// Different content should produce different hash
			detector2 := NewChangeDetector()
			testFile2 := filepath.Join(tmpDir, "test2.txt")
			os.WriteFile(testFile2, []byte(tt.content+"x"), 0644)

			hash3, _ := detector2.ComputeHash(testFile2)
			if string(hash1) == string(hash3) {
				t.Error("different content produced same hash")
			}
		})
	}
}

func TestChangeDetector_HashDifferentFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	content := []byte("Same content")
	os.WriteFile(file1, content, 0644)
	os.WriteFile(file2, content, 0644)

	detector := NewChangeDetector()

	hash1, _ := detector.ComputeHash(file1)
	hash2, _ := detector.ComputeHash(file2)

	// Same content should produce same hash
	if string(hash1) != string(hash2) {
		t.Error("same content in different files produced different hashes")
	}
}
