package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadSlice_StrictSubset (AC-PERF-005) verifies that LoadSlice reads ONLY
// the named section files and does NOT touch unrelated sections. The proof:
// unrelated section files are absent from disk, and the load still succeeds
// (using defaults for the missing ones), while the named sections are read.
func TestLoadSlice_StrictSubset(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	// Write only the sections in the PreToolUse slice.
	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: ddd\n")
	mustWriteFile(t, filepath.Join(sectionsDir, "workflow.yaml"),
		"workflow:\n  default_mode: autopilot\n")

	// The PreToolUse handler consumes: gate config (quality.yaml),
	// branch-guard flag (workflow.yaml). It does NOT need language.yaml,
	// statusline.yaml, ralph.yaml, etc.
	cfg, err := NewLoader().LoadSlice(dir, "quality", "workflow")
	if err != nil {
		t.Fatalf("LoadSlice: %v", err)
	}

	// Named sections should reflect file content.
	if cfg.Quality.DevelopmentMode != "ddd" {
		t.Fatalf("LoadSlice should read quality.yaml → ddd, got %q", cfg.Quality.DevelopmentMode)
	}

	// Unnamed sections should have compiled defaults (not read from disk).
	// The default DevelopmentMode should be "tdd" (or whatever the compiled
	// default is) — NOT a value from a file (no file exists for them).
	defaultCfg := NewDefaultConfig()
	if cfg.Language.ConversationLanguage != defaultCfg.Language.ConversationLanguage {
		t.Fatalf("LoadSlice should NOT read language.yaml; expected default %q, got %q",
			defaultCfg.Language.ConversationLanguage, cfg.Language.ConversationLanguage)
	}
}

// TestLoadSlice_PopulatesLoadedSections verifies that LoadSlice tracks only
// the sections it actually read.
func TestLoadSlice_PopulatesLoadedSections(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: ddd\n")
	mustWriteFile(t, filepath.Join(sectionsDir, "language.yaml"),
		"language:\n  conversation_language: ko\n")

	l := NewLoader()
	_, err := l.LoadSlice(dir, "quality")
	if err != nil {
		t.Fatalf("LoadSlice: %v", err)
	}

	loaded := l.LoadedSections()
	if !loaded["quality"] {
		t.Fatal("LoadSlice should mark quality as loaded")
	}
	if loaded["language"] {
		t.Fatal("LoadSlice should NOT mark language as loaded (not in slice)")
	}
}

// TestLoadWithCacheSlice_CacheMissUsesSlice (AC-PERF-005 integration) verifies
// that LoadWithCacheSlice, on a cache miss, loads only the named sections
// rather than the full set.
func TestLoadWithCacheSlice_CacheMissUsesSlice(t *testing.T) {
	dir := t.TempDir()
	sectionsDir := filepath.Join(dir, "config", "sections")
	mustMkdir(t, sectionsDir)

	mustWriteFile(t, filepath.Join(sectionsDir, "quality.yaml"),
		"constitution:\n  development_mode: ddd\n")

	// No cache file exists → guaranteed cache miss.
	l := NewLoader()
	cfg, err := l.LoadWithCacheSlice(dir, "quality")
	if err != nil {
		t.Fatalf("LoadWithCacheSlice: %v", err)
	}

	// quality.yaml was read.
	if cfg.Quality.DevelopmentMode != "ddd" {
		t.Fatalf("expected ddd from quality.yaml, got %q", cfg.Quality.DevelopmentMode)
	}

	// A cache file was written despite the slice load.
	cachePath := cacheFilePath(dir)
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("cache file should be written on miss: %v", err)
	}

	// Subsequent call should HIT the cache (full config now cached).
	cfg2, err := NewLoader().LoadWithCacheSlice(dir, "quality")
	if err != nil {
		t.Fatalf("second LoadWithCacheSlice: %v", err)
	}
	if cfg2.Quality.DevelopmentMode != "ddd" {
		t.Fatalf("cache hit should serve ddd, got %q", cfg2.Quality.DevelopmentMode)
	}
}
