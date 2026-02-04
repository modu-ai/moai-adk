package ops

import (
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// cacheEntry represents a single cache entry with TTL support.
type cacheEntry struct {
	key       string
	result    GitResult
	expiresAt time.Time
	element   *list.Element
}

// Cache provides an LRU cache with TTL support for Git operation results.
type Cache struct {
	mu         sync.RWMutex
	entries    map[string]*cacheEntry
	lruList    *list.List
	sizeLimit  int
	defaultTTL time.Duration
}

// NewCache creates a new cache with the specified size limit and default TTL.
func NewCache(sizeLimit int, defaultTTL time.Duration) *Cache {
	if sizeLimit <= 0 {
		sizeLimit = 100
	}
	if defaultTTL <= 0 {
		defaultTTL = 60 * time.Second
	}

	return &Cache{
		entries:    make(map[string]*cacheEntry),
		lruList:    list.New(),
		sizeLimit:  sizeLimit,
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a result from the cache.
// Returns the result and true if found and not expired, otherwise returns empty result and false.
func (c *Cache) Get(key string) (GitResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		return GitResult{}, false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		c.removeEntry(entry)
		return GitResult{}, false
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(entry.element)

	// Return a copy with CacheHit set
	result := entry.result
	result.CacheHit = true
	result.Cached = true

	return result, true
}

// Set stores a result in the cache with the specified TTL.
// If ttl is 0, the default TTL is used.
func (c *Cache) Set(key string, result GitResult, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	expiresAt := time.Now().Add(ttl)

	// Check if key already exists
	if entry, exists := c.entries[key]; exists {
		// Update existing entry
		entry.result = result
		entry.expiresAt = expiresAt
		c.lruList.MoveToFront(entry.element)
		return
	}

	// Evict if at capacity
	for c.lruList.Len() >= c.sizeLimit {
		c.evictOldest()
	}

	// Create new entry
	entry := &cacheEntry{
		key:       key,
		result:    result,
		expiresAt: expiresAt,
	}
	entry.element = c.lruList.PushFront(entry)
	c.entries[key] = entry
}

// Clear removes all cache entries for a specific operation type.
// Returns the number of entries removed.
func (c *Cache) Clear(opType GitOperationType) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var toRemove []*cacheEntry
	for _, entry := range c.entries {
		if entry.result.OperationType == opType {
			toRemove = append(toRemove, entry)
		}
	}

	for _, entry := range toRemove {
		c.removeEntry(entry)
	}

	return len(toRemove)
}

// ClearAll removes all cache entries.
// Returns the number of entries removed.
func (c *Cache) ClearAll() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := len(c.entries)
	c.entries = make(map[string]*cacheEntry)
	c.lruList.Init()

	return count
}

// Size returns the current number of entries in the cache.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Stats returns cache statistics.
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	size := len(c.entries)
	utilization := float64(size) / float64(c.sizeLimit)

	return CacheStats{
		Size:        size,
		SizeLimit:   c.sizeLimit,
		Utilization: utilization,
	}
}

// removeEntry removes an entry from the cache (must be called with lock held).
func (c *Cache) removeEntry(entry *cacheEntry) {
	c.lruList.Remove(entry.element)
	delete(c.entries, entry.key)
}

// evictOldest removes the least recently used entry (must be called with lock held).
func (c *Cache) evictOldest() {
	oldest := c.lruList.Back()
	if oldest == nil {
		return
	}

	entry := oldest.Value.(*cacheEntry)
	c.removeEntry(entry)
}

// CleanExpired removes all expired entries from the cache.
// Returns the number of entries removed.
func (c *Cache) CleanExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var toRemove []*cacheEntry

	for _, entry := range c.entries {
		if now.After(entry.expiresAt) {
			toRemove = append(toRemove, entry)
		}
	}

	for _, entry := range toRemove {
		c.removeEntry(entry)
	}

	return len(toRemove)
}

// GenerateCacheKey generates a cache key from the operation context.
// The key is an MD5 hash of the operation type, arguments, working directory, and branch.
func GenerateCacheKey(opType GitOperationType, args []string, workDir, branch string) string {
	data := fmt.Sprintf("%s:%v:%s:%s", opType, args, workDir, branch)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}
