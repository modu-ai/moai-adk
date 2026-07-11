package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// drift_cache.go — HEAD-SHA-keyed drift result cache
// (SPEC-SESSIONSTART-PERF-001 M1, REQ-SSP-004 / REQ-SSP-006).
//
// Drift detection is an ADVISORY, non-blocking check that runs on the
// session-start critical path. Recomputing it on every session start is wasted
// work whenever the repository has not moved, so the computed report is cached
// against the HEAD SHA it was derived from. While HEAD is unchanged, a hit
// returns the cached report without performing ANY git-log work.
//
// Every failure path here is FAIL-OPEN: a missing, unreadable, unparseable or
// stale cache simply causes a recompute. A cache problem must never fail — or
// even degrade — drift detection.

// driftCacheFilename is the cache file stored under the project's .moai/state/
// runtime-state directory (gitignored; the same family as context-usage.json
// and active-sessions.json).
const driftCacheFilename = "drift-cache.json"

// driftCacheFile is the on-disk cache payload.
//
// The cache is keyed ONLY on the HEAD SHA. Uncommitted frontmatter edits do not
// advance HEAD, so a stale-frontmatter window exists between an edit and its
// commit. This is an accepted, documented trade-off: the session-start check is
// advisory, and `moai spec drift --no-cache` (REQ-SSP-006a) is the authoritative
// on-demand path that always recomputes.
type driftCacheFile struct {
	HeadSHA    string        `json:"head_sha"`
	ComputedAt string        `json:"computed_at"`
	Count      int           `json:"count"`
	Records    []DriftRecord `json:"records"`
}

// driftCachePath returns the cache location for a project root.
func driftCachePath(baseDir string) string {
	return filepath.Join(baseDir, ".moai", "state", driftCacheFilename)
}

// loadDriftCache returns the cached report when a valid entry exists for head.
//
// ok=false means "recompute" and is returned for every failure mode: an empty
// head (non-git checkout), an absent or unreadable file, malformed JSON, or a
// cache written against a different HEAD (stale — AC-SSP-022).
func loadDriftCache(baseDir, head string) (*DriftReport, bool) {
	if head == "" {
		return nil, false
	}

	data, err := os.ReadFile(driftCachePath(baseDir))
	if err != nil {
		return nil, false // absent or unreadable — fail open
	}

	var cached driftCacheFile
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, false // corrupt — fail open
	}

	if cached.HeadSHA == "" || cached.HeadSHA != head {
		return nil, false // stale: HEAD advanced since this entry was written
	}

	records := cached.Records
	if records == nil {
		records = []DriftRecord{}
	}

	return &DriftReport{Records: records, Count: cached.Count}, true
}

// saveDriftCache persists report keyed on head.
//
// Best-effort by design: every error is swallowed. A failed cache write costs a
// recompute on the next run, which is strictly better than failing a check whose
// entire purpose is advisory.
func saveDriftCache(baseDir, head string, report *DriftReport) {
	if head == "" || report == nil {
		return
	}

	path := driftCachePath(baseDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}

	payload := driftCacheFile{
		HeadSHA:    head,
		ComputedAt: time.Now().UTC().Format(time.RFC3339),
		Count:      report.Count,
		Records:    report.Records,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(path, data, 0o644)
}
