package cache

import (
	"context"
	"sync"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// Option configures a DiagnosticCache at construction time.
type Option func(*DiagnosticCache)

// WithCleanupInterval sets the periodic cleanup interval for expired entries.
// The default is 10 seconds.
func WithCleanupInterval(d time.Duration) Option {
	return func(c *DiagnosticCache) {
		c.cleanupInterval = d
	}
}

// DiagnosticCache is a thread-safe, TTL-based in-memory cache for LSP diagnostics.
// Entries are keyed by file URI and invalidated when the document version changes
// or the TTL expires.
//
// @MX:ANCHOR: [AUTO] DiagnosticCache — core diagnostic cache type shared by Aggregator, tests, and Manager
// @MX:REASON: fan_in >= 3 — Aggregator.GetDiagnostics, cache tests, and Manager integration code all reference this type directly
type DiagnosticCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration

	cleanupInterval time.Duration

	// ctx and cancel manage the cleanup goroutine lifecycle.
	ctx    context.Context    //nolint:containedctx
	cancel context.CancelFunc
}

// NewDiagnosticCache creates a new DiagnosticCache with the given TTL and options.
// A TTL of 0 causes all entries to expire immediately.
func NewDiagnosticCache(ttl time.Duration, opts ...Option) *DiagnosticCache {
	c := &DiagnosticCache{
		entries:         make(map[string]*CacheEntry),
		ttl:             ttl,
		cleanupInterval: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Set stores diagnostics for the given URI at the given version.
// The entry expires after the cache's configured TTL.
//
// @MX:ANCHOR: [AUTO] DiagnosticCache.Set — entry point for storing diagnostic results
// @MX:REASON: fan_in >= 3 — Aggregator, cache tests, and integration tests all call Set
func (c *DiagnosticCache) Set(uri string, version int64, diagnostics []lsp.Diagnostic) {
	copied := make([]lsp.Diagnostic, len(diagnostics))
	copy(copied, diagnostics)

	// A TTL of 0 results in a zero ExpiresAt, which IsExpired treats as always expired.
	var expiresAt time.Time
	if c.ttl > 0 {
		expiresAt = time.Now().Add(c.ttl)
	}

	entry := &CacheEntry{
		Diagnostics: copied,
		Version:     version,
		ExpiresAt:   expiresAt,
	}

	c.mu.Lock()
	c.entries[uri] = entry
	c.mu.Unlock()
}

// Get retrieves diagnostics for the given URI and version.
// Returns (nil, false) when:
//   - no entry exists for the URI
//   - the stored version does not match the requested version (REQ-AGG-006)
//   - the entry has expired (REQ-AGG-004)
//
// On a cache hit, a copy of the diagnostics slice is returned to prevent
// callers from mutating the cached data.
//
// @MX:ANCHOR: [AUTO] DiagnosticCache.Get — core path for cache lookup
// @MX:REASON: fan_in >= 3 — Aggregator.GetDiagnostics, cache tests, and Ralph engine all retrieve results through Get
func (c *DiagnosticCache) Get(uri string, version int64) ([]lsp.Diagnostic, bool) {
	c.mu.RLock()
	entry, ok := c.entries[uri]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}
	if entry.Version != version {
		return nil, false
	}
	if entry.IsExpired() {
		return nil, false
	}

	copied := make([]lsp.Diagnostic, len(entry.Diagnostics))
	copy(copied, entry.Diagnostics)
	return copied, true
}

// GetStale returns the diagnostics for uri regardless of TTL or version.
// It is used for graceful degradation: when the upstream is unavailable,
// any cached value (even expired) is preferred over an empty response.
// Returns (nil, false) when no entry exists for the URI at all.
func (c *DiagnosticCache) GetStale(uri string) ([]lsp.Diagnostic, bool) {
	c.mu.RLock()
	entry, ok := c.entries[uri]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	copied := make([]lsp.Diagnostic, len(entry.Diagnostics))
	copy(copied, entry.Diagnostics)
	return copied, true
}

// Invalidate removes the cache entry for the given URI immediately (REQ-AGG-005).
// If no entry exists for the URI, this is a no-op.
func (c *DiagnosticCache) Invalidate(uri string) {
	c.mu.Lock()
	delete(c.entries, uri)
	c.mu.Unlock()
}

// Start begins the periodic cleanup goroutine that removes expired entries.
// The goroutine runs until the provided context is cancelled or Stop is called.
// Calling Start more than once is idempotent.
//
// @MX:WARN: [AUTO] Start goroutine — long-running background goroutine that evicts expired entries every cleanupInterval
// @MX:REASON: goroutine leak occurs without ctx cancellation. A Start/Stop pair or external ctx cancellation must be guaranteed.
func (c *DiagnosticCache) Start(ctx context.Context) {
	c.mu.Lock()
	if c.cancel != nil {
		// Already started — idempotent
		c.mu.Unlock()
		return
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.mu.Unlock()

	go c.cleanupLoop(c.ctx)
}

// Stop cancels the internal cleanup goroutine started by Start.
// If Start was never called, Stop is a no-op.
func (c *DiagnosticCache) Stop() {
	c.mu.Lock()
	cancel := c.cancel
	c.cancel = nil
	c.ctx = nil
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// cleanupLoop periodically removes expired entries from the cache.
func (c *DiagnosticCache) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.evictExpired()
		}
	}
}

// evictExpired removes all expired entries under the write lock.
func (c *DiagnosticCache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for uri, entry := range c.entries {
		if entry.IsExpired() {
			delete(c.entries, uri)
		}
	}
}
