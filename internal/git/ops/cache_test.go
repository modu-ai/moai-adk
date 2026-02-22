package ops

import (
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	tests := []struct {
		name      string
		sizeLimit int
		ttl       time.Duration
		wantLimit int
		wantTTL   time.Duration
	}{
		{
			name:      "default values",
			sizeLimit: 100,
			ttl:       60 * time.Second,
			wantLimit: 100,
			wantTTL:   60 * time.Second,
		},
		{
			name:      "custom values",
			sizeLimit: 50,
			ttl:       30 * time.Second,
			wantLimit: 50,
			wantTTL:   30 * time.Second,
		},
		{
			name:      "zero size defaults to 100",
			sizeLimit: 0,
			ttl:       60 * time.Second,
			wantLimit: 100,
			wantTTL:   60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCache(tt.sizeLimit, tt.ttl)
			if c == nil {
				t.Fatal("NewCache returned nil")
			}
			if c.sizeLimit != tt.wantLimit {
				t.Errorf("sizeLimit = %d, want %d", c.sizeLimit, tt.wantLimit)
			}
			if c.defaultTTL != tt.wantTTL {
				t.Errorf("defaultTTL = %v, want %v", c.defaultTTL, tt.wantTTL)
			}
		})
	}
}

func TestCache_SetAndGet(t *testing.T) {
	c := NewCache(100, 60*time.Second)

	result := GitResult{
		Success:       true,
		Stdout:        "main",
		OperationType: OpBranch,
	}

	// Set a value
	c.Set("test-key", result, 0)

	// Get the value
	got, hit := c.Get("test-key")
	if !hit {
		t.Fatal("expected cache hit")
	}
	if got.Stdout != result.Stdout {
		t.Errorf("Stdout = %q, want %q", got.Stdout, result.Stdout)
	}
	if !got.CacheHit {
		t.Error("expected CacheHit to be true")
	}
}

func TestCache_GetMiss(t *testing.T) {
	c := NewCache(100, 60*time.Second)

	_, hit := c.Get("nonexistent")
	if hit {
		t.Error("expected cache miss for nonexistent key")
	}
}

func TestCache_TTLExpiration(t *testing.T) {
	c := NewCache(100, 100*time.Millisecond)

	result := GitResult{
		Success: true,
		Stdout:  "test",
	}

	c.Set("expire-key", result, 100*time.Millisecond)

	// Should hit immediately
	_, hit := c.Get("expire-key")
	if !hit {
		t.Error("expected cache hit before expiration")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should miss after expiration
	_, hit = c.Get("expire-key")
	if hit {
		t.Error("expected cache miss after expiration")
	}
}

func TestCache_LRUEviction(t *testing.T) {
	c := NewCache(3, 60*time.Second)

	// Add 3 items
	for i := range 3 {
		c.Set(string(rune('a'+i)), GitResult{Stdout: string(rune('a' + i))}, 0)
	}

	// Access 'a' to make it recently used
	c.Get("a")

	// Add a 4th item, should evict 'b' (least recently used)
	c.Set("d", GitResult{Stdout: "d"}, 0)

	// 'a' should still exist (recently accessed)
	_, hit := c.Get("a")
	if !hit {
		t.Error("expected 'a' to still be in cache")
	}

	// 'b' should be evicted (LRU)
	_, hit = c.Get("b")
	if hit {
		t.Error("expected 'b' to be evicted")
	}

	// 'c' and 'd' should exist
	_, hit = c.Get("c")
	if !hit {
		t.Error("expected 'c' to still be in cache")
	}
	_, hit = c.Get("d")
	if !hit {
		t.Error("expected 'd' to be in cache")
	}
}

func TestCache_Clear(t *testing.T) {
	c := NewCache(100, 60*time.Second)

	// Add items with different operation types
	c.Set("branch-1", GitResult{OperationType: OpBranch}, 0)
	c.Set("branch-2", GitResult{OperationType: OpBranch}, 0)
	c.Set("status-1", GitResult{OperationType: OpStatus}, 0)

	// Clear only branch entries
	cleared := c.Clear(OpBranch)
	if cleared != 2 {
		t.Errorf("Clear returned %d, want 2", cleared)
	}

	// Branch entries should be gone
	_, hit := c.Get("branch-1")
	if hit {
		t.Error("branch-1 should be cleared")
	}

	// Status entry should remain
	_, hit = c.Get("status-1")
	if !hit {
		t.Error("status-1 should still exist")
	}
}

func TestCache_ClearAll(t *testing.T) {
	c := NewCache(100, 60*time.Second)

	c.Set("key1", GitResult{}, 0)
	c.Set("key2", GitResult{}, 0)

	cleared := c.ClearAll()
	if cleared != 2 {
		t.Errorf("ClearAll returned %d, want 2", cleared)
	}

	if c.Size() != 0 {
		t.Errorf("Size after ClearAll = %d, want 0", c.Size())
	}
}

func TestCache_Size(t *testing.T) {
	c := NewCache(100, 60*time.Second)

	if c.Size() != 0 {
		t.Errorf("initial Size = %d, want 0", c.Size())
	}

	c.Set("key1", GitResult{}, 0)
	c.Set("key2", GitResult{}, 0)

	if c.Size() != 2 {
		t.Errorf("Size after 2 adds = %d, want 2", c.Size())
	}
}

func TestCache_Stats(t *testing.T) {
	c := NewCache(100, 60*time.Second)

	c.Set("key1", GitResult{}, 0)
	c.Set("key2", GitResult{}, 0)

	// Get existing key (hit)
	c.Get("key1")
	c.Get("key1")

	// Get non-existing key (miss)
	c.Get("nonexistent")

	stats := c.Stats()
	if stats.Size != 2 {
		t.Errorf("stats.Size = %d, want 2", stats.Size)
	}
	if stats.SizeLimit != 100 {
		t.Errorf("stats.SizeLimit = %d, want 100", stats.SizeLimit)
	}
	if stats.Utilization != 0.02 {
		t.Errorf("stats.Utilization = %f, want 0.02", stats.Utilization)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := NewCache(100, 60*time.Second)
	done := make(chan bool)

	// Concurrent writes
	for i := range 10 {
		go func(i int) {
			key := string(rune('a' + i))
			c.Set(key, GitResult{Stdout: key}, 0)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := range 10 {
		go func(i int) {
			key := string(rune('a' + i))
			c.Get(key)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 20 {
		<-done
	}

	// Should not panic and size should be reasonable
	if c.Size() > 100 {
		t.Errorf("Size = %d, exceeded limit", c.Size())
	}
}

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		opType   GitOperationType
		args     []string
		workDir  string
		branch   string
		wantSame bool
	}{
		{
			name:    "same inputs produce same key",
			opType:  OpBranch,
			args:    []string{"--show-current"},
			workDir: "/test/dir",
			branch:  "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := GenerateCacheKey(tt.opType, tt.args, tt.workDir, tt.branch)
			key2 := GenerateCacheKey(tt.opType, tt.args, tt.workDir, tt.branch)

			if key1 != key2 {
				t.Errorf("same inputs produced different keys: %s vs %s", key1, key2)
			}

			// Different inputs should produce different keys
			key3 := GenerateCacheKey(tt.opType, tt.args, tt.workDir, "develop")
			if key1 == key3 {
				t.Error("different branch should produce different key")
			}
		})
	}
}

func TestCache_UpdateExistingKey(t *testing.T) {
	c := NewCache(100, 60*time.Second)

	// Set initial value
	c.Set("key", GitResult{Stdout: "original"}, 0)

	// Update with new value
	c.Set("key", GitResult{Stdout: "updated"}, 0)

	got, hit := c.Get("key")
	if !hit {
		t.Fatal("expected cache hit")
	}
	if got.Stdout != "updated" {
		t.Errorf("Stdout = %q, want %q", got.Stdout, "updated")
	}

	// Size should still be 1
	if c.Size() != 1 {
		t.Errorf("Size = %d, want 1 (key update should not increase size)", c.Size())
	}
}
