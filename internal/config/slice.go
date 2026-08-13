package config

import (
	"log/slog"
	"path/filepath"
)

// preToolUseSliceSections is the set of section names the PreToolUse handler
// actually consumes: gate config (quality), branch-guard flag (workflow),
// system (log level for autonomy tier), and gate (the quality gate config).
// These are the ONLY sections loaded on a cache miss when the PreToolUse
// slice is requested (REQ-PERF-005).
var preToolUseSliceSections = []string{"quality", "workflow", "system", "gate"}

// sectionLoaderFunc is the signature of a per-section loader method on Loader.
type sectionLoaderFunc func(l *Loader, dir string, cfg *Config)

// sectionLoaders maps a section name to its loader method. Used by LoadSlice
// to load only the named sections.
var sectionLoaders = map[string]sectionLoaderFunc{
	"user":          (*Loader).loadUserSection,
	"language":      (*Loader).loadLanguageSection,
	"quality":       (*Loader).loadQualitySection,
	"git_convention": (*Loader).loadGitConventionSection,
	"git_strategy":  (*Loader).loadGitStrategySection,
	"llm":           (*Loader).loadLLMSection,
	"ralph":         (*Loader).loadRalphSection,
	"state":         (*Loader).loadStateSection,
	"workflow":      (*Loader).loadWorkflowSection,
	"statusline":    (*Loader).loadStatuslineSection,
	"feedback":      (*Loader).loadFeedbackSection,
	"handoff":       (*Loader).loadHandoffSection,
	"archive":       (*Loader).loadArchiveSection,
	"gate":          (*Loader).loadGateSection,
	"system":        (*Loader).loadSystemSection,
	"constitution":  (*Loader).loadConstitutionSection,
	"context":       (*Loader).loadContextSection,
	"interview":     (*Loader).loadInterviewSection,
	"design":        (*Loader).loadDesignSection,
}

// LoadSlice reads ONLY the named configuration sections from disk, applying
// compiled defaults for all other sections. This is the lazy config slice
// (REQ-PERF-005): on a cache miss, the PreToolUse handler requests only the
// sections it consumes rather than the full ~20-section set.
//
// The returned Config has full struct shape — every field is populated (named
// sections from files, unnamed sections from defaults). Downstream code reads
// the Config struct identically regardless of whether Load or LoadSlice was
// used; the difference is purely in how many section files were opened.
//
// sectionNames uses the same naming convention as loadedSections: the section
// name without the .yaml extension (e.g. "quality", not "quality.yaml").
// Unknown section names are silently skipped.
func (l *Loader) LoadSlice(configDir string, sectionNames ...string) (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.loadedSections = make(map[string]bool)
	cfg := NewDefaultConfig()

	sectionsDir := filepath.Join(filepath.Clean(configDir), "config", "sections")

	// Load only the named sections. Unknown names are silently skipped.
	for _, name := range sectionNames {
		loaderFn, ok := sectionLoaders[name]
		if !ok {
			slog.Debug("LoadSlice: unknown section name, skipping", "section", name)
			continue
		}
		loaderFn(l, sectionsDir, cfg)
	}

	return cfg, nil
}

// LoadWithCacheSlice combines the disk cache (M1) with the lazy slice (M2).
// On a cache HIT, the full cached config is served directly at cache-read
// cost (the lazy slice is not exercised). On a cache MISS, only the named
// sections are loaded from disk instead of the full ~20-section set, and a
// fresh cache is written.
//
// The sectionNames parameter selects which sections to load on a cache miss.
// Use preToolUseSliceSections for the PreToolUse handler's slice.
func (l *Loader) LoadWithCacheSlice(configDir string, sectionNames ...string) (*Config, error) {
	configDir = filepath.Clean(configDir)
	cachePath := cacheFilePath(configDir)
	sectionsDir := filepath.Join(configDir, "config", "sections")

	// Compute the current fingerprint of all section files on disk.
	currentFP, fpErr := computeFingerprint(sectionsDir)
	if fpErr != nil {
		slog.Debug("config cache: fingerprint computation failed, falling back to LoadSlice",
			"sections_dir", sectionsDir, "error", fpErr)
		return l.LoadSlice(configDir, sectionNames...)
	}

	// Attempt to read the cache file.
	cached, cacheErr := readCache(cachePath)

	// Cache HIT: valid file + matching schema version + matching fingerprint.
	if cacheErr == nil &&
		cached.SchemaVersion == configCacheSchemaVersion &&
		fingerprintValid(cached.Fingerprint, currentFP) {

		l.mu.Lock()
		l.loadedSections = fingerprintToLoadedSections(cached.Fingerprint)
		l.mu.Unlock()

		slog.Debug("config cache: HIT (serving cached config via slice path)")
		return cached.Config, nil
	}

	// Cache MISS: lazy slice load.
	if cacheErr != nil {
		slog.Debug("config cache: MISS via slice path (read failed)", "error", cacheErr)
	} else if cached.SchemaVersion != configCacheSchemaVersion {
		slog.Debug("config cache: MISS via slice path (schema version mismatch)",
			"cached", cached.SchemaVersion, "current", configCacheSchemaVersion)
	} else {
		slog.Debug("config cache: MISS via slice path (fingerprint invalidated)")
	}

	cfg, err := l.LoadSlice(configDir, sectionNames...)
	if err != nil {
		return nil, err
	}

	// Write cache (fail-open).
	if writeErr := writeCache(cachePath, cfg, currentFP); writeErr != nil {
		slog.Debug("config cache: write failed via slice path (non-fatal)",
			"path", cachePath, "error", writeErr)
	}

	return cfg, nil
}

// PreToolUseSliceSections returns the section names the PreToolUse handler
// consumes. This is the canonical slice for the lazy config path
// (REQ-PERF-005).
func PreToolUseSliceSections() []string {
	return preToolUseSliceSections
}
