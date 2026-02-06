package update

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultCacheTTL is the default time-to-live for cached update check results.
const DefaultCacheTTL = 1 * time.Hour

// CacheEntry stores the result of an update check for reuse.
type CacheEntry struct {
	CheckedAt  time.Time    `json:"checked_at"`
	LatestInfo *VersionInfo `json:"latest_info,omitempty"`
	Available  bool         `json:"available"`
	CurrentVer string       `json:"current_ver"`
}

// Cache provides file-based caching for update check results.
// It prevents redundant GitHub API calls across sessions.
type Cache struct {
	path string
	ttl  time.Duration
}

// NewCache creates a Cache that stores results at the given path.
// If path is empty, defaults to ~/.moai/cache/update_check.json.
// If ttl is zero, defaults to DefaultCacheTTL.
func NewCache(path string, ttl time.Duration) *Cache {
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		path = filepath.Join(homeDir, ".moai", "cache", "update_check.json")
	}
	if ttl == 0 {
		ttl = DefaultCacheTTL
	}
	return &Cache{path: path, ttl: ttl}
}

// Get returns a cached entry if it is fresh and matches the current version.
// Returns nil (no error) on cache miss, expiration, version mismatch, or corruption.
func (c *Cache) Get(currentVersion string) *CacheEntry {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	// Invalidate if binary version changed since cache was written
	if entry.CurrentVer != currentVersion {
		return nil
	}

	// Invalidate if TTL expired
	if time.Since(entry.CheckedAt) > c.ttl {
		return nil
	}

	return &entry
}

// Set writes a cache entry to disk, creating directories as needed.
// Errors are returned but callers may choose to ignore them.
func (c *Cache) Set(entry *CacheEntry) error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0644); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}

	return nil
}
