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
// @MX:ANCHOR: [AUTO] DiagnosticCache — Aggregator, 테스트, Manager가 공유하는 진단 캐시 핵심 타입
// @MX:REASON: fan_in >= 3 — Aggregator.GetDiagnostics, cache 테스트, Manager 연동 코드가 모두 이 타입을 직접 참조함
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
// @MX:ANCHOR: [AUTO] DiagnosticCache.Set — 진단 결과 저장 진입점
// @MX:REASON: fan_in >= 3 — Aggregator, 캐시 테스트, 통합 테스트가 모두 Set을 호출함
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
// @MX:ANCHOR: [AUTO] DiagnosticCache.Get — 캐시 조회 핵심 경로
// @MX:REASON: fan_in >= 3 — Aggregator.GetDiagnostics, 캐시 테스트, Ralph 엔진이 모두 Get을 통해 결과를 조회함
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
// @MX:WARN: [AUTO] Start 고루틴 — cleanupInterval마다 만료 항목을 삭제하는 장기 실행 백그라운드 고루틴
// @MX:REASON: ctx 취소 없이는 고루틴 누수가 발생함. Start/Stop 쌍 또는 외부 ctx 취소를 반드시 보장해야 함.
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
