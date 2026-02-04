package quality

import (
	"crypto/sha256"
	"os"
	"sync"
)

// ChangeDetector manages file hash-based change detection per REQ-HOOK-060.
// It provides thread-safe hash computation and caching for efficient
// change detection without redundant disk reads.
type ChangeDetector struct {
	mu    sync.RWMutex
	cache map[string][]byte
}

// NewChangeDetector creates a new ChangeDetector with initialized cache.
func NewChangeDetector() *ChangeDetector {
	return &ChangeDetector{
		cache: make(map[string][]byte),
	}
}

// ComputeHash calculates SHA-256 hash of the file at filePath.
// Always computes a fresh hash (no cache check) and caches the result.
// Returns a 32-byte SHA-256 hash.
func (d *ChangeDetector) ComputeHash(filePath string) ([]byte, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(content)
	hashBytes := hash[:]

	// Cache the computed result (make a copy for storage)
	d.mu.Lock()
	defer d.mu.Unlock()
	stored := make([]byte, len(hashBytes))
	copy(stored, hashBytes)
	d.cache[filePath] = stored

	return hashBytes, nil
}

// HasChanged compares the current file hash with the provided previousHash.
// Returns true if the file has changed (hash differs), false if unchanged.
func (d *ChangeDetector) HasChanged(filePath string, previousHash []byte) (bool, error) {
	currentHash, err := d.ComputeHash(filePath)
	if err != nil {
		return false, err
	}

	if len(currentHash) != len(previousHash) {
		return true, nil
	}

	for i := range currentHash {
		if currentHash[i] != previousHash[i] {
			return true, nil
		}
	}

	return false, nil
}

// GetCachedHash retrieves a cached hash for the given file path.
// Returns the hash and true if found, nil and false if not cached.
func (d *ChangeDetector) GetCachedHash(filePath string) ([]byte, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	hash, found := d.cache[filePath]
	if !found {
		return nil, false
	}

	// Return a copy to prevent external modification
	result := make([]byte, len(hash))
	copy(result, hash)
	return result, true
}

// CacheHash stores a hash for the given file path.
// This allows explicit pre-caching of known hash values.
func (d *ChangeDetector) CacheHash(filePath string, hash []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Store a copy to prevent external modification
	stored := make([]byte, len(hash))
	copy(stored, hash)
	d.cache[filePath] = stored
}
