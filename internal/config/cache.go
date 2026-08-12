package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config/atomicfile"
)

// configCacheSchemaVersion is the cache file format version. A mismatch
// between this constant and the cache file's schema_version field triggers
// the fail-open re-merge path (REQ-PERF-004 / AC-PERF-004, C-EDGE-001).
// Bump this when the Config struct's serialization changes in a way that
// breaks backward compatibility with prior cache files.
const configCacheSchemaVersion = 1

// cacheFileName is the fixed cache file name under the state directory.
// Fixed name ensures predictable gitignore and cleanup (REQ-PERF-009).
const cacheFileName = "config-cache.json"

// sectionFingerprint records the mtime (nanoseconds) and size (bytes) of a
// single section file at cache-write time. Both fields MUST be recorded to
// detect: mtime change (AC-PERF-002), size change without mtime change
// (AP-2 defense, B-2), and section deletion (AC-PERF-003 — the file's
// absence from the current fingerprint map is the invalidation signal).
type sectionFingerprint struct {
	MTime int64 `json:"mtime_ns"` // UnixNano of the file's ModTime
	Size  int64 `json:"size"`     // file size in bytes
}

// cacheFile is the on-disk JSON format for the config cache. It carries a
// schema_version guard, the per-section fingerprint, a written_at timestamp,
// and the full merged *Config (C-3: no new config format — existing struct
// encoded as JSON).
type cacheFile struct {
	SchemaVersion int                           `json:"schema_version"`
	WrittenAt     time.Time                     `json:"written_at"`
	Fingerprint   map[string]sectionFingerprint `json:"fingerprint"`
	Config        *Config                       `json:"config"`
}

// cacheFilePath returns the path to the config cache file under the state
// directory: <configDir>/state/config-cache.json (REQ-PERF-009).
//
// @MX:ANCHOR: [AUTO] config cache path resolver — predictable location for gitignore + cleanup
// @MX:REASON: REQ-PERF-009; every cache read/write routes through this single function so the location is enforceable
func cacheFilePath(configDir string) string {
	return filepath.Join(filepath.Clean(configDir), "state", cacheFileName)
}

// LoadWithCache reads configuration with a disk cache layer. On a cache hit
// (cache file present + fingerprint matches every section file's current
// mtime AND size + schema_version matches), it returns the cached config
// WITHOUT opening any section file. On a cache miss (absent, corrupt,
// schema-mismatched, or fingerprint-invalidated), it falls through to the
// full re-merge path and writes a fresh cache (REQ-PERF-001..004, 008, 009).
//
// Fail-open invariant (C-1): any cache error — corrupt JSON, unreadable
// file, schema mismatch, filesystem error — falls back to the full Load
// path silently. The cache is an optimization, never a correctness dependency.
func (l *Loader) LoadWithCache(configDir string) (*Config, error) {
	configDir = filepath.Clean(configDir)
	cachePath := cacheFilePath(configDir)
	sectionsDir := filepath.Join(configDir, "config", "sections")

	// Compute the current fingerprint of all section files on disk.
	currentFP, fpErr := computeFingerprint(sectionsDir)
	if fpErr != nil {
		// Fail open: fingerprint computation failure → full load (no cache).
		slog.Debug("config cache: fingerprint computation failed, falling back to full load",
			"sections_dir", sectionsDir, "error", fpErr)
		return l.Load(configDir)
	}

	// Attempt to read the cache file.
	cached, cacheErr := readCache(cachePath)

	// Cache HIT: valid file + matching schema version + matching fingerprint.
	if cacheErr == nil &&
		cached.SchemaVersion == configCacheSchemaVersion &&
		fingerprintValid(cached.Fingerprint, currentFP) {

		// Populate loadedSections from the cache fingerprint so downstream
		// code (validation, env overrides) sees the same section set as a
		// full Load would have produced.
		l.mu.Lock()
		l.loadedSections = fingerprintToLoadedSections(cached.Fingerprint)
		l.mu.Unlock()

		slog.Debug("config cache: HIT (serving cached config)")
		return cached.Config, nil
	}

	// Cache MISS: full re-merge.
	if cacheErr != nil {
		slog.Debug("config cache: MISS (read failed)", "error", cacheErr)
	} else if cached.SchemaVersion != configCacheSchemaVersion {
		slog.Debug("config cache: MISS (schema version mismatch)",
			"cached", cached.SchemaVersion, "current", configCacheSchemaVersion)
	} else {
		slog.Debug("config cache: MISS (fingerprint invalidated)")
	}

	cfg, err := l.Load(configDir)
	if err != nil {
		return nil, err
	}

	// Write cache (fail-open: a write failure is logged and ignored — the
	// cache is an optimization, not a correctness dependency; C-EDGE-003).
	if writeErr := writeCache(cachePath, cfg, currentFP); writeErr != nil {
		slog.Debug("config cache: write failed (non-fatal)",
			"path", cachePath, "error", writeErr)
	}

	return cfg, nil
}

// computeFingerprint scans the sections directory and records the mtime and
// size of every *.yaml file. Returns an empty map (no error) when the
// directory does not exist — matching Loader.Load's "use defaults" behavior.
func computeFingerprint(sectionsDir string) (map[string]sectionFingerprint, error) {
	entries, err := os.ReadDir(sectionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]sectionFingerprint{}, nil
		}
		return nil, fmt.Errorf("read sections dir %s: %w", sectionsDir, err)
	}

	fp := make(map[string]sectionFingerprint, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fp[entry.Name()] = sectionFingerprint{
			MTime: info.ModTime().UnixNano(),
			Size:  info.Size(),
		}
	}
	return fp, nil
}

// fingerprintValid compares the cached fingerprint against the current one.
// Returns true ONLY when every cached entry still exists with the same mtime
// AND size, AND no new files have appeared (length match). A deleted section
// file (present in cached, absent from current) is an invalidation signal
// equal in force to an mtime change (AC-PERF-003 / REQ-PERF-003).
func fingerprintValid(cached, current map[string]sectionFingerprint) bool {
	// A new file appearing (current has more entries) invalidates.
	if len(current) != len(cached) {
		return false
	}
	for name, c := range cached {
		cur, ok := current[name]
		if !ok {
			// Section file deleted (AC-PERF-003).
			return false
		}
		if c.MTime != cur.MTime {
			return false
		}
		if c.Size != cur.Size {
			// Size changed without mtime change (AP-2 / B-2).
			return false
		}
	}
	return true
}

// readCache reads and parses the cache file. Returns an error on any
// filesystem or parse failure — the caller treats all errors as cache-miss.
func readCache(path string) (*cacheFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read cache %s: %w", path, err)
	}
	var cf cacheFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("parse cache %s: %w", path, err)
	}
	if cf.Config == nil {
		return nil, fmt.Errorf("cache %s: config field is nil", path)
	}
	return &cf, nil
}

// writeCache serializes the config + fingerprint to the cache file
// atomically via atomicfile.Write (write-temp + rename in the same
// directory, NOT /tmp cross-filesystem — AP-3 / REQ-PERF-008).
func writeCache(path string, cfg *Config, fp map[string]sectionFingerprint) error {
	cf := cacheFile{
		SchemaVersion: configCacheSchemaVersion,
		WrittenAt:     time.Now().UTC(),
		Fingerprint:   fp,
		Config:        cfg,
	}
	data, err := json.Marshal(&cf)
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	// Ensure the state directory exists before the atomic write.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	return atomicfile.Write(path, data, 0o644)
}

// fingerprintToLoadedSections converts a fingerprint map (filename → metadata)
// into the loadedSections map that Loader.Load would have produced
// (section-name-without-extension → true). This keeps the cache-hit path
// consistent with the full-load path for downstream code that reads
// loadedSections (validation, env overrides).
func fingerprintToLoadedSections(fp map[string]sectionFingerprint) map[string]bool {
	result := make(map[string]bool, len(fp))
	for filename := range fp {
		name := strings.TrimSuffix(filename, ".yaml")
		result[name] = true
	}
	return result
}
