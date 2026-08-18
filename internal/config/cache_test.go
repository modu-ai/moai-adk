package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// requireCacheEnabled pins the config-cache escape hatch (EnvConfigCacheDisabled)
// OFF for the duration of a test. Every cache-feature test below exercises
// LoadWithCache's read/write path directly and therefore MUST run with the
// cache layer active, regardless of any ambient MOAI_CONFIG_CACHE_DISABLED set
// in the environment (e.g. when the whole package is run under that env to
// verify the disabled path does not break non-cache tests). Without this pin,
// the escape hatch would silently bypass the cache and the assertions would
// observe a cache-miss on every call.
func requireCacheEnabled(t *testing.T) {
	t.Helper()
	t.Setenv(EnvConfigCacheDisabled, "")
}

// TestCache_HitSkipsSectionParse (AC-PERF-001) verifies that a cache hit
// serves the cached config WITHOUT re-reading section files. The proof: we
// overwrite the section file's content (changing the effective config value)
// while preserving its mtime AND size, so the fingerprint stays valid. If the
// cache hit path reads the file, it would return the new value; if it serves
// from cache, it returns the old (cached) value.
func TestCache_HitSkipsSectionParse(t *testing.T) {
	requireCacheEnabled(t)
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	// Pad both versions to the same byte count so the size fingerprint matches.
	origContent := "constitution:\n  development_mode: ddd      \n" // 6-char pad
	swapContent := "constitution:\n  development_mode: tdd      \n" // same length

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"), origContent)

	// First load: populates the cache with development_mode=ddd.
	cfg1, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("first LoadWithCache: %v", err)
	}
	if cfg1.Quality.DevelopmentMode != "ddd" {
		t.Fatalf("first load: expected development_mode=ddd, got %q", cfg1.Quality.DevelopmentMode)
	}

	// Overwrite with swapped content (same byte count → same size fingerprint).
	origInfo, _ := os.Stat(filepath.Join(sectionsDir, "quality.yaml"))
	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"), swapContent)
	// Reset mtime to match the original so the fingerprint stays valid.
	_ = os.Chtimes(filepath.Join(sectionsDir, "quality.yaml"),
		origInfo.ModTime(), origInfo.ModTime())

	// Second load: fingerprint matches (same mtime + same size) → cache HIT.
	// If the cache served from the file, it would see "tdd"; from cache → "ddd".
	cfg2, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("second LoadWithCache: %v", err)
	}
	if cfg2.Quality.DevelopmentMode != "ddd" {
		t.Fatalf("cache hit should serve cached ddd, got %q (cache miss re-read the file)", cfg2.Quality.DevelopmentMode)
	}
}

// TestCache_MtimeChangeInvalidates (AC-PERF-002) verifies that advancing a
// section file's mtime invalidates the cache.
func TestCache_MtimeChangeInvalidates(t *testing.T) {
	requireCacheEnabled(t)
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: tdd\n")

	// First load populates the cache with tdd.
	cfg1, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if cfg1.Quality.DevelopmentMode != "tdd" {
		t.Fatalf("expected tdd, got %q", cfg1.Quality.DevelopmentMode)
	}

	// Advance mtime by writing new content + sleeping to ensure mtime changes.
	time.Sleep(20 * time.Millisecond)
	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: ddd\n")

	// Second load should detect mtime change, re-read, and serve ddd.
	cfg2, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("second load: %v", err)
	}
	if cfg2.Quality.DevelopmentMode != "ddd" {
		t.Fatalf("mtime change should invalidate cache → ddd, got %q", cfg2.Quality.DevelopmentMode)
	}
}

// TestCache_SectionDeletionInvalidates (AC-PERF-003) verifies that deleting a
// section file invalidates the cache.
func TestCache_SectionDeletionInvalidates(t *testing.T) {
	requireCacheEnabled(t)
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "feedback.yaml"),
		"feedback:\n  repository: test/repo\n")

	// First load populates the cache with feedback.repository = test/repo.
	cfg1, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if cfg1.Feedback.Repository != "test/repo" {
		t.Fatalf("expected test/repo, got %q", cfg1.Feedback.Repository)
	}

	// Delete feedback.yaml.
	mustRemove(t, filepath.Join(sectionsDir, "feedback.yaml"))

	// Second load should detect deletion (section present at cache-write but now gone).
	cfg2, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("second load: %v", err)
	}
	// Default Repository should be different from "test/repo" (it's the compiled default).
	if cfg2.Feedback.Repository == "test/repo" {
		t.Fatalf("deletion should invalidate cache → default repo, but still got cached test/repo")
	}
}

// TestCache_CorruptFailsOpen (AC-PERF-004) verifies that a corrupt cache file
// fails open silently — falls back to full re-merge, no user-facing error.
func TestCache_CorruptFailsOpen(t *testing.T) {
	requireCacheEnabled(t)
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: tdd\n")

	// First load populates a valid cache.
	_, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}

	// Corrupt the cache file.
	cachePath := cacheFilePath(dir)
	mustWriteFile(t, cachePath, "{this is not valid JSON")

	// Second load should fail open: re-merge from files, no error.
	cfg, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("corrupt cache should fail open, got error: %v", err)
	}
	if cfg.Quality.DevelopmentMode != "tdd" {
		t.Fatalf("corrupt cache fail-open should re-merge → tdd, got %q", cfg.Quality.DevelopmentMode)
	}

	// Cache should be rewritten with valid content.
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read rewritten cache: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("corrupt cache was not overwritten with valid JSON on re-merge")
	}
}

// TestCache_SchemaVersionMismatchFailsOpen (AC-PERF-004, C-EDGE-001) verifies
// that a schema-version mismatch triggers the fail-open re-merge path.
func TestCache_SchemaVersionMismatchFailsOpen(t *testing.T) {
	requireCacheEnabled(t)
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: tdd\n")

	// First load populates a valid cache.
	_, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}

	// Rewrite cache with a wrong schema_version.
	cachePath := cacheFilePath(dir)
	corrupt := map[string]any{
		"schema_version": 99999, // future version
		"written_at":     "2026-08-13T00:00:00Z",
		"fingerprint":    map[string]any{},
		"config":         map[string]any{},
	}
	data, _ := json.Marshal(corrupt)
	mustWriteFile(t, cachePath, string(data))

	// Load should fail open: re-merge, no error, rewrite cache.
	cfg, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("schema mismatch should fail open, got error: %v", err)
	}
	if cfg.Quality.DevelopmentMode != "tdd" {
		t.Fatalf("schema mismatch fail-open should re-merge → tdd, got %q", cfg.Quality.DevelopmentMode)
	}
}

// TestCache_AtomicConcurrentWrite (AC-PERF-008) verifies that concurrent
// cache writes never produce a partially-written file visible to readers.
func TestCache_AtomicConcurrentWrite(t *testing.T) {
	requireCacheEnabled(t)
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: tdd\n")

	const writers = 8
	var wg sync.WaitGroup
	errs := make(chan error, writers)

	// Spawn N concurrent writers — each does a full LoadWithCache (which writes the cache).
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := NewLoader().LoadWithCache(dir)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent LoadWithCache error: %v", err)
		}
	}

	// After all concurrent writes, the cache file must be valid JSON.
	data, err := os.ReadFile(cacheFilePath(dir))
	if err != nil {
		t.Fatalf("read cache after concurrent writes: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("cache file is corrupt (not valid JSON) after concurrent writes")
	}

	// And a subsequent read must produce a valid config.
	cfg, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("post-race read: %v", err)
	}
	if cfg.Quality.DevelopmentMode != "tdd" {
		t.Fatalf("post-race read: expected tdd, got %q", cfg.Quality.DevelopmentMode)
	}
}

// TestCache_LocationUnderStateDir (AC-PERF-009) verifies the cache file lives
// under .moai/state/ and has the fixed name config-cache.json.
func TestCache_LocationUnderStateDir(t *testing.T) {
	requireCacheEnabled(t)
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: tdd\n")

	_, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	resolved := cacheFilePath(dir)
	stateDir := filepath.Join(dir, "state")
	if !strings.HasPrefix(resolved, stateDir) {
		t.Fatalf("cache path %q is not under state dir %q", resolved, stateDir)
	}
	if filepath.Base(resolved) != "config-cache.json" {
		t.Fatalf("cache file name should be config-cache.json, got %q", filepath.Base(resolved))
	}

	if _, err := os.Stat(resolved); err != nil {
		t.Fatalf("cache file not created at expected location: %v", err)
	}
}

// TestCache_DoesNotCreateConfigDir (issue #1568) verifies that a cache-miss
// write does not materialize the config directory as a side effect. Before
// this guard, running any non-trivial moai command in a fresh directory
// created <dir>/.moai/state/config-cache.json via writeCache's MkdirAll, and
// `moai init` then failed its "project already initialized" validation
// against the directory the very same command had just created.
func TestCache_DoesNotCreateConfigDir(t *testing.T) {
	requireCacheEnabled(t)
	base := t.TempDir()
	// The config directory of an uninitialized project: <root>/.moai, absent.
	configDir := filepath.Join(base, "fresh-project", ".moai")

	cfg, err := NewLoader().LoadWithCache(configDir)
	if err != nil {
		t.Fatalf("LoadWithCache on uninitialized dir: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadWithCache returned nil config")
	}

	if _, statErr := os.Stat(configDir); statErr == nil {
		t.Fatalf("LoadWithCache created %s as a side effect; the cache write must be skipped when the config directory does not exist", configDir)
	} else if !os.IsNotExist(statErr) {
		t.Fatalf("stat %s: %v", configDir, statErr)
	}
}

// TestCache_SizeChangeInvalidates (AP-2 defense) verifies that a file whose
// mtime did NOT change but whose size DID is treated as invalid.
func TestCache_SizeChangeInvalidates(t *testing.T) {
	requireCacheEnabled(t)
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: tdd\n")

	// First load populates the cache.
	_, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}

	// Write different-sized content, then reset mtime to match the original
	// (simulating same-mtime-different-size edge case).
	origInfo, _ := os.Stat(filepath.Join(sectionsDir, "quality.yaml"))
	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: ddd\n  coverage_threshold: 0.90\n")
	_ = os.Chtimes(filepath.Join(sectionsDir, "quality.yaml"),
		origInfo.ModTime(), origInfo.ModTime())

	// Second load should detect size change and invalidate.
	cfg, err := NewLoader().LoadWithCache(dir)
	if err != nil {
		t.Fatalf("second load: %v", err)
	}
	if cfg.Quality.DevelopmentMode != "ddd" {
		t.Fatalf("size change should invalidate cache → ddd, got %q", cfg.Quality.DevelopmentMode)
	}
}

// --- helpers ---

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustRemove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove %s: %v", path, err)
	}
}
