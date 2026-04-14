package cache

import (
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// TestCacheEntry_IsExpired_FutureExpiresAt verifies that an entry with a future
// ExpiresAt is not expired.
func TestCacheEntry_IsExpired_FutureExpiresAt(t *testing.T) {
	entry := &CacheEntry{
		Diagnostics: []lsp.Diagnostic{},
		Version:     1,
		ExpiresAt:   time.Now().Add(10 * time.Second),
	}
	if entry.IsExpired() {
		t.Fatal("expected IsExpired()=false for future ExpiresAt, got true")
	}
}

// TestCacheEntry_IsExpired_PastExpiresAt verifies that an entry with a past
// ExpiresAt is expired.
func TestCacheEntry_IsExpired_PastExpiresAt(t *testing.T) {
	entry := &CacheEntry{
		Diagnostics: []lsp.Diagnostic{},
		Version:     1,
		ExpiresAt:   time.Now().Add(-1 * time.Second),
	}
	if !entry.IsExpired() {
		t.Fatal("expected IsExpired()=true for past ExpiresAt, got false")
	}
}

// TestCacheEntry_IsExpired_ZeroTime verifies that an entry with zero ExpiresAt
// is always considered expired.
func TestCacheEntry_IsExpired_ZeroTime(t *testing.T) {
	entry := &CacheEntry{
		Diagnostics: []lsp.Diagnostic{},
		Version:     1,
		ExpiresAt:   time.Time{}, // zero value
	}
	if !entry.IsExpired() {
		t.Fatal("expected IsExpired()=true for zero ExpiresAt, got false")
	}
}
