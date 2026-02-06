package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_GetSet(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, cachePath string)
		currentVersion string
		ttl            time.Duration
		wantNil        bool
		wantAvailable  bool
	}{
		{
			name: "fresh cache hit",
			setup: func(t *testing.T, cachePath string) {
				t.Helper()
				c := NewCache(cachePath, DefaultCacheTTL)
				entry := &CacheEntry{
					CheckedAt:  time.Now(),
					Available:  true,
					CurrentVer: "v2.0.0",
					LatestInfo: &VersionInfo{Version: "v2.0.1"},
				}
				if err := c.Set(entry); err != nil {
					t.Fatal(err)
				}
			},
			currentVersion: "v2.0.0",
			ttl:            DefaultCacheTTL,
			wantNil:        false,
			wantAvailable:  true,
		},
		{
			name: "expired cache miss",
			setup: func(t *testing.T, cachePath string) {
				t.Helper()
				c := NewCache(cachePath, DefaultCacheTTL)
				entry := &CacheEntry{
					CheckedAt:  time.Now().Add(-2 * time.Hour),
					Available:  true,
					CurrentVer: "v2.0.0",
					LatestInfo: &VersionInfo{Version: "v2.0.1"},
				}
				if err := c.Set(entry); err != nil {
					t.Fatal(err)
				}
			},
			currentVersion: "v2.0.0",
			ttl:            DefaultCacheTTL,
			wantNil:        true,
		},
		{
			name: "version mismatch invalidation",
			setup: func(t *testing.T, cachePath string) {
				t.Helper()
				c := NewCache(cachePath, DefaultCacheTTL)
				entry := &CacheEntry{
					CheckedAt:  time.Now(),
					Available:  false,
					CurrentVer: "v1.9.0",
				}
				if err := c.Set(entry); err != nil {
					t.Fatal(err)
				}
			},
			currentVersion: "v2.0.0",
			ttl:            DefaultCacheTTL,
			wantNil:        true,
		},
		{
			name:           "missing cache file",
			setup:          func(t *testing.T, cachePath string) { t.Helper() },
			currentVersion: "v2.0.0",
			ttl:            DefaultCacheTTL,
			wantNil:        true,
		},
		{
			name: "corrupted cache file",
			setup: func(t *testing.T, cachePath string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(cachePath, []byte("not json"), 0644); err != nil {
					t.Fatal(err)
				}
			},
			currentVersion: "v2.0.0",
			ttl:            DefaultCacheTTL,
			wantNil:        true,
		},
		{
			name: "no update available cached",
			setup: func(t *testing.T, cachePath string) {
				t.Helper()
				c := NewCache(cachePath, DefaultCacheTTL)
				entry := &CacheEntry{
					CheckedAt:  time.Now(),
					Available:  false,
					CurrentVer: "v2.0.0",
				}
				if err := c.Set(entry); err != nil {
					t.Fatal(err)
				}
			},
			currentVersion: "v2.0.0",
			ttl:            DefaultCacheTTL,
			wantNil:        false,
			wantAvailable:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cachePath := filepath.Join(tmpDir, "cache", "update_check.json")

			tt.setup(t, cachePath)

			c := NewCache(cachePath, tt.ttl)
			got := c.Get(tt.currentVersion)

			if tt.wantNil {
				if got != nil {
					t.Errorf("Get() = %+v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("Get() = nil, want non-nil")
			}

			if got.Available != tt.wantAvailable {
				t.Errorf("Get().Available = %v, want %v", got.Available, tt.wantAvailable)
			}
		})
	}
}

func TestCache_DefaultPath(t *testing.T) {
	c := NewCache("", 0)
	if c.path == "" {
		t.Error("NewCache with empty path should use default")
	}
	if c.ttl != DefaultCacheTTL {
		t.Errorf("NewCache with zero ttl should default to %v, got %v", DefaultCacheTTL, c.ttl)
	}
}

func TestCache_SetCreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "deep", "nested", "dir", "cache.json")

	c := NewCache(cachePath, DefaultCacheTTL)
	entry := &CacheEntry{
		CheckedAt:  time.Now(),
		Available:  false,
		CurrentVer: "v1.0.0",
	}

	if err := c.Set(entry); err != nil {
		t.Fatalf("Set() should create directories, got error: %v", err)
	}

	if _, err := os.Stat(cachePath); err != nil {
		t.Errorf("cache file should exist after Set(), got error: %v", err)
	}
}
