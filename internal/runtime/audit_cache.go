// Package runtime provides core runtime utilities for MoAI workflow operations.
// Source: SPEC-WF-AUDIT-GATE-001 REQ-WAG-003 / AC-WAG-09
package runtime

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// cacheTTL is the validity period for a cached PASS verdict. REQ-WAG-003.
const cacheTTL = 24 * time.Hour

// CachedEntry holds a persisted PASS verdict for cache lookup.
type CachedEntry struct {
	// AuditAt is the UTC time of the original audit.
	AuditAt time.Time

	// AuditorVersion is the plan-auditor identifier used.
	AuditorVersion string

	// ReportPath is the path to the iteration-scoped audit report.
	ReportPath string

	// PlanArtifactHash is the SHA-256 hash of plan artifacts at audit time.
	PlanArtifactHash string
}

// AuditCache provides plan artifact hashing and 24-hour verdict caching.
//
// The cache key is (specID, planArtifactHash). A cached entry is valid when:
// - verdict == PASS
// - planArtifactHash matches current hash (change invalidates cache)
// - entry age < 24h
//
// @MX:ANCHOR: [AUTO] AuditCache is the primary caching interface for the audit gate
// @MX:REASON: fan_in >= 3 (GateConfig.Invoke, test mocks, audit_cache_test)
type AuditCache interface {
	// ComputeHash computes a combined SHA-256 hash of all plan artifacts in specDir.
	//
	// Hash algorithm: SHA-256 over whitespace-normalized content of each artifact file,
	// sorted by filename, concatenated. OPEN QUESTION Q1 resolution: whitespace-insensitive.
	//
	// Artifact files considered (in sorted order): spec.md, plan.md, acceptance.md, tasks.md.
	ComputeHash(specDir string) (string, error)

	// Lookup returns a cached PASS entry if one exists for (specID, hash) within 24h.
	// Returns (entry, true) on hit, (nil, false) on miss.
	Lookup(specID, hash string, now time.Time) (*CachedEntry, bool)

	// Store saves a PASS verdict to the cache.
	// Only PASS verdicts are cached; other verdicts are not stored.
	Store(specID, hash string, result *AuditResult)
}

// planArtifactNames is the ordered list of plan artifact file names to hash.
// OPEN QUESTION Q1 resolution: whitespace-insensitive, sorted by filename.
var planArtifactNames = []string{
	"acceptance.md",
	"plan.md",
	"spec.md",
	"tasks.md",
}

// InMemoryCache implements AuditCache using an in-memory store.
//
// Suitable for both production (single process lifetime) and tests.
// Each GateConfig holds its own InMemoryCache instance, so cache entries do not
// persist across separate /moai run invocations. For cross-invocation caching,
// the daily report file serves as the durable cache source (see AuditReporter).
type InMemoryCache struct {
	entries map[string]*CachedEntry
}

// NewInMemoryCache creates an empty InMemoryCache.
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{entries: make(map[string]*CachedEntry)}
}

// ComputeHash computes a combined, whitespace-normalized SHA-256 hash
// of the plan artifact files present in specDir.
//
// Missing artifact files are silently skipped (only spec.md is required).
// Content is whitespace-normalized before hashing (runs of whitespace → single space).
func (c *InMemoryCache) ComputeHash(specDir string) (string, error) {
	h := sha256.New()

	for _, name := range planArtifactNames {
		path := filepath.Join(specDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // optional artifact
			}
			return "", fmt.Errorf("read %s: %w", name, err)
		}
		// Whitespace normalization: collapse runs of whitespace to a single space.
		normalized := normalizeWhitespace(string(data))
		// Include filename as separator to prevent hash collisions between files.
		// hash.Hash.Write never errors per io.Writer contract, so the return value is intentionally discarded.
		_, _ = fmt.Fprintf(h, "%s:%s\n", name, normalized)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// normalizeWhitespace collapses any run of whitespace characters to a single space
// and trims leading/trailing whitespace.
func normalizeWhitespace(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

// cacheKey returns the map key for (specID, hash).
func cacheKey(specID, hash string) string {
	return specID + ":" + hash
}

// Lookup checks the in-memory cache for a valid PASS entry.
func (c *InMemoryCache) Lookup(specID, hash string, now time.Time) (*CachedEntry, bool) {
	entry, ok := c.entries[cacheKey(specID, hash)]
	if !ok {
		return nil, false
	}
	if now.Sub(entry.AuditAt) >= cacheTTL {
		// TTL expired — invalidate the entry.
		delete(c.entries, cacheKey(specID, hash))
		return nil, false
	}
	return entry, true
}

// Store saves a PASS audit result to the in-memory cache.
// Non-PASS results are ignored.
func (c *InMemoryCache) Store(specID, hash string, result *AuditResult) {
	if result.Verdict != VerdictPass {
		return
	}
	c.entries[cacheKey(specID, hash)] = &CachedEntry{
		AuditAt:          result.AuditAt,
		AuditorVersion:   result.AuditorVersion,
		ReportPath:       result.ReportPath,
		PlanArtifactHash: hash,
	}
}

// ListSpecArtifactFiles returns the sorted list of plan artifact files found in specDir.
// Used for debugging and testing.
func ListSpecArtifactFiles(specDir string) ([]string, error) {
	var found []string
	for _, name := range planArtifactNames {
		path := filepath.Join(specDir, name)
		if _, err := os.Stat(path); err == nil {
			found = append(found, name)
		}
	}
	sort.Strings(found)
	return found, nil
}

// ValidateReportDir ensures that the plan-audit report directory exists and
// is safe (within projectDir). Creates it if absent.
//
// REQ-WAG-004 / AC-WAG-10: directory auto-creation and path traversal prevention.
func ValidateReportDir(projectDir, reportDir string) error {
	clean := filepath.Clean(reportDir)
	cleanProject := filepath.Clean(projectDir)

	if !strings.HasPrefix(clean, cleanProject+string(filepath.Separator)) &&
		clean != cleanProject {
		return fmt.Errorf("report directory %q is outside project root %q (path traversal prevented)", clean, cleanProject)
	}

	info, err := os.Stat(clean)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("report path %q exists but is not a directory", clean)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("stat %q: %w", clean, err)
	}

	return os.MkdirAll(clean, fs.ModePerm)
}

// SanitizeMarkdown escapes Markdown special characters in a user-supplied string.
//
// Used to sanitize bypass_reason before writing to report files. REQ-WAG-006.
// Escapes: *, _, [, ], (, ), #, `, |
func SanitizeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"*", `\*`,
		"_", `\_`,
		"[", `\[`,
		"]", `\]`,
		"(", `\(`,
		")", `\)`,
		"#", `\#`,
		"`", "\\`",
		"|", `\|`,
	)
	return replacer.Replace(s)
}
