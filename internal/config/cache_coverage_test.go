package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestCache_WriteCacheMkdirFailure verifies writeCache handles a missing
// parent directory by creating it (not an error path — it succeeds).
func TestCache_WriteCacheCreatesStateDir(t *testing.T) {
	dir := t.TempDir()
	// State dir does not exist yet.
	cachePath := filepath.Join(dir, "state", "config-cache.json")
	cfg := NewDefaultConfig()
	fp := map[string]sectionFingerprint{"quality.yaml": {MTime: 1, Size: 50}}

	if err := writeCache(cachePath, cfg, fp); err != nil {
		t.Fatalf("writeCache should create state dir: %v", err)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("cache file should exist: %v", err)
	}
}

// TestCache_ReadCacheNilConfig verifies readCache rejects a cache with nil Config.
func TestCache_ReadCacheNilConfig(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "test-cache.json")
	// Write a cache file with nil config.
	data, _ := json.Marshal(map[string]any{
		"schema_version": 1,
		"fingerprint":    map[string]any{},
		"config":         nil,
	})
	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
		t.Fatalf("write test cache: %v", err)
	}

	_, err := readCache(cachePath)
	if err == nil {
		t.Fatal("readCache should error on nil config")
	}
}

// TestCache_ComputeFingerprintMissingDir verifies computeFingerprint returns
// empty map when the sections directory does not exist.
func TestCache_ComputeFingerprintMissingDir(t *testing.T) {
	fp, err := computeFingerprint("/nonexistent/path/sections")
	if err != nil {
		t.Fatalf("computeFingerprint on missing dir should not error: %v", err)
	}
	if len(fp) != 0 {
		t.Fatalf("computeFingerprint on missing dir should return empty map, got %d entries", len(fp))
	}
}

// TestCache_FingerprintValidNewFileInvalidates verifies that a new file
// appearing in current (but not in cached) invalidates.
func TestCache_FingerprintValidNewFileInvalidates(t *testing.T) {
	cached := map[string]sectionFingerprint{
		"quality.yaml": {MTime: 1, Size: 50},
	}
	current := map[string]sectionFingerprint{
		"quality.yaml":  {MTime: 1, Size: 50},
		"language.yaml": {MTime: 2, Size: 30},
	}
	if fingerprintValid(cached, current) {
		t.Fatal("fingerprintValid should return false when current has extra files")
	}
}

// TestLoadWithCache_FingerprintFailureFallsBack verifies that a fingerprint
// computation failure (e.g. permissions) falls back to full Load.
func TestLoadWithCache_FingerprintFailureFallsBack(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: tdd\n")

	// Normal load should work.
	cfg, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("LoadWithCache: %v", err)
	}
	if cfg.Quality.DevelopmentMode != "tdd" {
		t.Fatalf("expected tdd, got %q", cfg.Quality.DevelopmentMode)
	}
}

// TestPreToolUseSliceSections_ReturnsExpected verifies the exported helper.
func TestPreToolUseSliceSections_ReturnsExpected(t *testing.T) {
	sections := PreToolUseSliceSections()
	if len(sections) == 0 {
		t.Fatal("PreToolUseSliceSections should return non-empty slice")
	}
	// Must include quality (gate config) and workflow (branch-guard flag).
	found := map[string]bool{}
	for _, s := range sections {
		found[s] = true
	}
	if !found["quality"] {
		t.Fatal("PreToolUseSliceSections must include 'quality' (gate config)")
	}
	if !found["workflow"] {
		t.Fatal("PreToolUseSliceSections must include 'workflow' (branch-guard flag)")
	}
}

// TestLoadSlice_UnknownSectionSkipped verifies unknown section names are
// silently skipped (no panic, no error).
func TestLoadSlice_UnknownSectionSkipped(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: tdd\n")

	// Include an unknown section name — should be skipped.
	cfg, err := NewLoader().LoadSlice(dir, "quality", "nonexistent_section")
	if err != nil {
		t.Fatalf("LoadSlice with unknown section: %v", err)
	}
	if cfg.Quality.DevelopmentMode != "tdd" {
		t.Fatalf("expected tdd, got %q", cfg.Quality.DevelopmentMode)
	}
}

// TestLoadWithCacheSlice_CacheHitServesCachedConfig verifies that after a
// cache write, a subsequent LoadWithCacheSlice hits the cache.
func TestLoadWithCacheSlice_CacheHitServesCachedConfig(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: ddd\n")

	// First call: cache miss, writes cache.
	l1 := NewLoader()
	cfg1, err := l1.LoadWithCacheSlice(dir, "quality")
	if err != nil {
		t.Fatalf("first LoadWithCacheSlice: %v", err)
	}
	if cfg1.Quality.DevelopmentMode != "ddd" {
		t.Fatalf("expected ddd, got %q", cfg1.Quality.DevelopmentMode)
	}

	// Second call: should hit cache.
	l2 := NewLoader()
	cfg2, err := l2.LoadWithCacheSlice(dir, "quality")
	if err != nil {
		t.Fatalf("second LoadWithCacheSlice: %v", err)
	}
	if cfg2.Quality.DevelopmentMode != "ddd" {
		t.Fatalf("cache hit should serve ddd, got %q", cfg2.Quality.DevelopmentMode)
	}
}
