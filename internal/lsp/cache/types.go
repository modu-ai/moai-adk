package cache

import (
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// CacheEntry holds a snapshot of diagnostics for a single file at a specific version.
//
// @MX:ANCHOR: [AUTO] CacheEntry — TTL cache entry type referenced directly by DiagnosticCache, Aggregator, tests, and other callers
// @MX:REASON: fan_in >= 3 — DiagnosticCache.Set/Get, Aggregator.GetDiagnostics, and cache tests all use this type
type CacheEntry struct {
	// Diagnostics is the list of diagnostics captured at this version.
	Diagnostics []lsp.Diagnostic

	// Version is the document version at which these diagnostics were captured.
	Version int64

	// ExpiresAt is the absolute time after which this entry is considered stale.
	// A zero value is always considered expired.
	ExpiresAt time.Time
}

// IsExpired reports whether this cache entry has passed its expiry time.
// A zero ExpiresAt is always expired.
func (e *CacheEntry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return true
	}
	return time.Now().After(e.ExpiresAt)
}
