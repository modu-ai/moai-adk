package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
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
// runtime-state directory (gitignored; the same family as context-usage/
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

// CacheHeadSHA resolves the drift cache's KEY for a project root: the HEAD SHA
// the cache is stored against.
//
// It exists so the SessionStart handler's suppression record and the cache
// itself are keyed on the same value resolved the same way — a record keyed on
// a differently-derived head would suppress fills for a cache entry that never
// matches. O(1): one `git rev-parse`, no git-log work.
func CacheHeadSHA(baseDir string) (string, error) {
	return gitHeadSHAAt(baseDir)
}

// CachedDriftCount returns the drift count cached against the current HEAD.
//
// ok=false means "no usable entry for this HEAD" and is returned for every
// failure mode: a non-git checkout, an absent or unreadable cache file,
// malformed JSON, or an entry written against a different HEAD. It performs NO
// git-log work on either path — that is the whole point: the session-start
// advisory resolves from the cache alone, and a miss costs nothing.
func CachedDriftCount(baseDir string) (int, bool) {
	head, err := gitHeadSHAAt(baseDir)
	if err != nil || head == "" {
		return 0, false
	}
	report, ok := loadDriftCache(baseDir, head)
	if !ok || report == nil {
		return 0, false
	}
	return report.Count, true
}

// driftCacheReplaceFn is the final rename step of the cache write, as a seam so
// an interrupted write can be observed directly rather than inferred. Production
// points it at the in-tree atomic primitive.
var driftCacheReplaceFn = atomicfile.Replace

// saveDriftCache persists report keyed on head.
//
// Best-effort by design: every error is swallowed. A failed cache write costs a
// recompute on the next run, which is strictly better than failing a check whose
// entire purpose is advisory.
//
// The write is ATOMIC — temp file in the destination directory, then
// atomicfile.Replace. The out-of-band fill child (the only writer of this file
// that is bounded by a deadline) is bounded at a duration engineered to fire
// near completion, which is exactly when an in-place os.WriteFile of a large
// JSON payload is most likely to be mid-flight. loadDriftCache would fail open
// on the truncated result; a torn file is prevented rather than tolerated.
//
// Replace is the counterpart of atomicfile.Claim and the two are never
// substituted: Replace answers "make this path hold this content" (an existing
// destination is success), Claim answers "am I the one who proceeds" (an
// existing path is failure).
func saveDriftCache(baseDir, head string, report *DriftReport) {
	if head == "" || report == nil {
		return
	}

	path := driftCachePath(baseDir)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
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

	tmp, err := os.CreateTemp(dir, ".drift-cache-*.tmp")
	if err != nil {
		return
	}
	tmpName := tmp.Name()
	// Every failure path below removes the temp file: a residual artefact in
	// .moai/state/ is indistinguishable from state the runtime owns.
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		_ = os.Remove(tmpName)
		return
	}
	if err := driftCacheReplaceFn(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
	}
}
