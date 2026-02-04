package quality

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"
	"time"
)

// ChangeDetector manages file hash-based change detection.
// It caches computed hashes to avoid redundant file reads.
type ChangeDetector struct {
	mu    sync.RWMutex
	cache map[string]*cachedHash
}

// cachedHash represents a cached file hash with expiration time.
type cachedHash struct {
	hash      []byte
	expiresAt time.Time
}

// NewChangeDetector creates a new ChangeDetector with an empty cache.
func NewChangeDetector() *ChangeDetector {
	return &ChangeDetector{
		cache: make(map[string]*cachedHash),
	}
}

// ComputeHash calculates the SHA-256 hash of a file's contents.
// Returns an empty slice and nil error if file doesn't exist (graceful degradation).
func (d *ChangeDetector) ComputeHash(filePath string) ([]byte, error) {
	// Check cache first
	d.mu.RLock()
	cached, exists := d.cache[filePath]
	if exists && time.Now().Before(cached.expiresAt) {
		hashCopy := make([]byte, len(cached.hash))
		copy(hashCopy, cached.hash)
		d.mu.RUnlock()
		return hashCopy, nil
	}
	d.mu.RUnlock()

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - return empty hash without error
			return []byte{}, nil
		}
		return nil, err
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(content)
	hashBytes := hash[:]

	// Cache the result
	d.mu.Lock()
	d.cache[filePath] = &cachedHash{
		hash:      hashBytes,
		expiresAt: time.Now().Add(HashCacheTTL),
	}
	d.mu.Unlock()

	return hashBytes, nil
}

// HasChanged compares the current file hash with a previous hash.
// Returns true if the file is different or doesn't exist.
func (d *ChangeDetector) HasChanged(filePath string, previousHash []byte) (bool, error) {
	currentHash, err := d.ComputeHash(filePath)
	if err != nil {
		return false, err
	}

	// If file doesn't exist, consider it unchanged (graceful)
	if len(currentHash) == 0 && len(previousHash) == 0 {
		return false, nil
	}

	// If lengths differ, hashes are different
	if len(currentHash) != len(previousHash) {
		return true, nil
	}

	// Compare byte by byte
	for i := range currentHash {
		if currentHash[i] != previousHash[i] {
			return true, nil
		}
	}

	return false, nil
}

// GetCachedHash retrieves a cached hash if it exists and hasn't expired.
// Returns the hash and whether it was found in cache.
func (d *ChangeDetector) GetCachedHash(filePath string) ([]byte, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	cached, exists := d.cache[filePath]
	if !exists {
		return nil, false
	}

	if time.Now().After(cached.expiresAt) {
		// Cache expired
		delete(d.cache, filePath)
		return nil, false
	}

	hashCopy := make([]byte, len(cached.hash))
	copy(hashCopy, cached.hash)
	return hashCopy, true
}

// CacheHash stores a hash in the cache with TTL.
func (d *ChangeDetector) CacheHash(filePath string, hash []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()

	hashCopy := make([]byte, len(hash))
	copy(hashCopy, hash)

	d.cache[filePath] = &cachedHash{
		hash:      hashCopy,
		expiresAt: time.Now().Add(HashCacheTTL),
	}
}

// HashToString converts a hash byte slice to a hexadecimal string.
func HashToString(hash []byte) string {
	if len(hash) == 0 {
		return ""
	}
	return hex.EncodeToString(hash)
}

// ClearExpired removes all expired entries from the cache.
func (d *ChangeDetector) ClearExpired() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for path, cached := range d.cache {
		if now.After(cached.expiresAt) {
			delete(d.cache, path)
		}
	}
}

// ClearCache removes all cached hashes.
func (d *ChangeDetector) ClearCache() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cache = make(map[string]*cachedHash)
}
