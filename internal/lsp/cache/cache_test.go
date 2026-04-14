package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// makeDiagnostics creates a test diagnostics slice with the given message.
func makeDiagnostics(msg string) []lsp.Diagnostic {
	return []lsp.Diagnostic{
		{
			Message:  msg,
			Severity: lsp.SeverityError,
		},
	}
}

// ---------------------------------------------------------------------------
// T-002: DiagnosticCache.Get/Set tests
// ---------------------------------------------------------------------------

// TestDiagnosticCache_SetGet_HappyPath verifies that Set then Get returns
// the stored diagnostics.
func TestDiagnosticCache_SetGet_HappyPath(t *testing.T) {
	c := NewDiagnosticCache(5 * time.Second)
	diags := makeDiagnostics("undefined variable")

	c.Set("file:///main.go", 1, diags)

	got, ok := c.Get("file:///main.go", 1)
	if !ok {
		t.Fatal("expected ok=true for valid (uri, version), got false")
	}
	if len(got) != 1 || got[0].Message != "undefined variable" {
		t.Fatalf("unexpected diagnostics: %v", got)
	}
}

// TestDiagnosticCache_Get_WrongVersion verifies that Get returns false when
// the requested version differs from the stored version (REQ-AGG-006).
func TestDiagnosticCache_Get_WrongVersion(t *testing.T) {
	c := NewDiagnosticCache(5 * time.Second)
	c.Set("file:///main.go", 1, makeDiagnostics("error"))

	_, ok := c.Get("file:///main.go", 2) // version mismatch
	if ok {
		t.Fatal("expected ok=false for version mismatch, got true")
	}
}

// TestDiagnosticCache_Get_ExpiredEntry verifies that Get returns false for
// an expired entry (REQ-AGG-004).
func TestDiagnosticCache_Get_ExpiredEntry(t *testing.T) {
	// TTL of 0 means entries expire immediately
	c := NewDiagnosticCache(0)
	c.Set("file:///main.go", 1, makeDiagnostics("error"))

	_, ok := c.Get("file:///main.go", 1)
	if ok {
		t.Fatal("expected ok=false for expired entry, got true")
	}
}

// TestDiagnosticCache_Get_MissingURI verifies that Get returns false for an
// unknown URI.
func TestDiagnosticCache_Get_MissingURI(t *testing.T) {
	c := NewDiagnosticCache(5 * time.Second)

	_, ok := c.Get("file:///unknown.go", 1)
	if ok {
		t.Fatal("expected ok=false for missing URI, got true")
	}
}

// TestDiagnosticCache_Get_ReturnsCopy verifies that the returned slice is a
// copy (mutation does not affect the cached entry).
func TestDiagnosticCache_Get_ReturnsCopy(t *testing.T) {
	c := NewDiagnosticCache(5 * time.Second)
	original := makeDiagnostics("original")
	c.Set("file:///main.go", 1, original)

	got, ok := c.Get("file:///main.go", 1)
	if !ok {
		t.Fatal("expected ok=true")
	}
	// Mutate the returned slice
	got[0].Message = "mutated"

	// Re-fetch and verify the cache is not affected
	got2, ok2 := c.Get("file:///main.go", 1)
	if !ok2 {
		t.Fatal("expected ok=true on second Get")
	}
	if got2[0].Message != "original" {
		t.Fatalf("expected cache to be unaffected by mutation, got %q", got2[0].Message)
	}
}

// ---------------------------------------------------------------------------
// T-003: Invalidate + TTL tests
// ---------------------------------------------------------------------------

// TestDiagnosticCache_Invalidate_RemovesEntry verifies that Invalidate removes
// the entry so subsequent Get returns false (REQ-AGG-005).
func TestDiagnosticCache_Invalidate_RemovesEntry(t *testing.T) {
	c := NewDiagnosticCache(5 * time.Second)
	c.Set("file:///main.go", 1, makeDiagnostics("error"))

	c.Invalidate("file:///main.go")

	_, ok := c.Get("file:///main.go", 1)
	if ok {
		t.Fatal("expected ok=false after Invalidate, got true")
	}
}

// TestDiagnosticCache_Invalidate_NonExistentURI verifies that Invalidate on an
// unknown URI is a no-op (no panic, no error).
func TestDiagnosticCache_Invalidate_NonExistentURI(t *testing.T) {
	c := NewDiagnosticCache(5 * time.Second)
	// Should not panic
	c.Invalidate("file:///nonexistent.go")
}

// TestDiagnosticCache_TTL_ZeroExpiresImmediately verifies that a cache created
// with TTL=0 causes all entries to be expired on the next Get.
func TestDiagnosticCache_TTL_ZeroExpiresImmediately(t *testing.T) {
	c := NewDiagnosticCache(0)
	c.Set("file:///main.go", 1, makeDiagnostics("error"))

	_, ok := c.Get("file:///main.go", 1)
	if ok {
		t.Fatal("expected ok=false for TTL=0 (immediately expired), got true")
	}
}

// ---------------------------------------------------------------------------
// T-004: Start/Stop cleanup goroutine tests
// ---------------------------------------------------------------------------

// TestDiagnosticCache_Start_RemovesExpiredEntries verifies that after Start,
// the cleanup goroutine removes expired entries within the cleanup interval.
func TestDiagnosticCache_Start_RemovesExpiredEntries(t *testing.T) {
	// Use 5ms cleanup interval to make the test fast
	c := NewDiagnosticCache(0, WithCleanupInterval(5*time.Millisecond))
	c.Set("file:///main.go", 1, makeDiagnostics("error"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)
	defer c.Stop()

	// Wait until the cleanup goroutine runs (max 200ms)
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		c.mu.RLock()
		_, exists := c.entries["file:///main.go"]
		c.mu.RUnlock()
		if !exists {
			return // cleanup removed the expired entry
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("cleanup goroutine did not remove expired entry within deadline")
}

// TestDiagnosticCache_Stop_CancelsCleanup verifies that Stop cancels the cleanup
// goroutine promptly.
func TestDiagnosticCache_Stop_CancelsCleanup(t *testing.T) {
	c := NewDiagnosticCache(5*time.Second, WithCleanupInterval(50*time.Millisecond))

	ctx := context.Background()
	c.Start(ctx)

	start := time.Now()
	c.Stop()
	elapsed := time.Since(start)

	if elapsed > 300*time.Millisecond {
		t.Fatalf("Stop took %v, expected < 300ms", elapsed)
	}
}

// TestDiagnosticCache_Start_Idempotent verifies that calling Start twice does
// not start a second goroutine (idempotent).
func TestDiagnosticCache_Start_Idempotent(t *testing.T) {
	c := NewDiagnosticCache(5*time.Second, WithCleanupInterval(50*time.Millisecond))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.Start(ctx)
	c.Start(ctx) // second call — must be no-op
	defer c.Stop()

	// Verify the cache still functions correctly
	c.Set("file:///main.go", 1, makeDiagnostics("error"))
	_, ok := c.Get("file:///main.go", 1)
	if !ok {
		t.Fatal("expected cache to be functional after double Start")
	}
}

// TestDiagnosticCache_Stop_BeforeStart verifies that Stop before Start is a no-op.
func TestDiagnosticCache_Stop_BeforeStart(t *testing.T) {
	c := NewDiagnosticCache(5 * time.Second)
	// Should not panic
	c.Stop()
}

// TestDiagnosticCache_Concurrent_GetSet verifies thread-safety under concurrent
// access from 50 goroutines (REQ-AGG-010).
func TestDiagnosticCache_Concurrent_GetSet(t *testing.T) {
	c := NewDiagnosticCache(5 * time.Second)
	const goroutines = 50

	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			uri := "file:///main.go"
			c.Set(uri, int64(idx), makeDiagnostics("msg"))
			c.Get(uri, int64(idx)) //nolint:errcheck // concurrent access test
		}(i)
	}
	wg.Wait()
	// If we reach here without data race, the test passes.
}
