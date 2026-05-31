// Package state holds runtime-managed telemetry writers/readers under .moai/state/.
//
// SPEC-V3R6-PROMPT-CACHE-001 M3 — cache-usage JSONL writer + reader + aggregation.
//
// This package owns the on-disk representation of Anthropic Prompt Caching
// telemetry. The PostToolUse hook (internal/hook) appends entries here; the
// moai doctor cache metric (internal/cli) aggregates them. Keeping the record
// type and the read/write/aggregate logic in one package lets both M3 and M4
// share a single source of truth for the JSONL schema.
//
// REQ-PC-004: each entry carries cache_creation_input_tokens and
//             cache_read_input_tokens (plus timestamp, session_id, turn, model).
// REQ-PC-006: hit rate = sum(cache_read) / (sum(cache_read) + sum(cache_creation)).
package state

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// CacheUsageFileName is the JSONL telemetry file under .moai/state/.
const CacheUsageFileName = "cache-usage.jsonl"

// CacheUsageEntry is one row of the cache-usage JSONL telemetry log.
//
// The JSON field names use the FULL Anthropic API field names
// (cache_creation_input_tokens / cache_read_input_tokens) so that AC-PC-005
// can grep the persisted entry for both keys. The shorter struct field names
// (CacheCreation / CacheRead) are Go-side ergonomics only.
type CacheUsageEntry struct {
	// Timestamp is the RFC3339 (UTC) time the entry was written.
	Timestamp string `json:"timestamp"`
	// SessionID is the Claude Code session identifier.
	SessionID string `json:"session_id"`
	// Turn is the 1-based turn index within the session.
	Turn int `json:"turn"`
	// CacheCreation is usage.cache_creation_input_tokens from the API response.
	CacheCreation int `json:"cache_creation_input_tokens"`
	// CacheRead is usage.cache_read_input_tokens from the API response.
	CacheRead int `json:"cache_read_input_tokens"`
	// Model is the model identifier (e.g., "claude-sonnet-4-6").
	Model string `json:"model"`
}

// CacheUsageStats is the aggregate of a window of CacheUsageEntry rows.
type CacheUsageStats struct {
	// TotalCacheCreation is the sum of cache_creation_input_tokens.
	TotalCacheCreation int
	// TotalCacheRead is the sum of cache_read_input_tokens.
	TotalCacheRead int
	// HitRate is read/(read+creation) in [0,1]; 0 when both sums are zero.
	HitRate float64
	// TotalSessions is the count of distinct session_id values in the window.
	TotalSessions int
	// SingleTurnSessions is the count of sessions whose max turn == 1.
	SingleTurnSessions int
	// EntryCount is the number of entries aggregated in the window.
	EntryCount int
}

// SingleTurnRatio returns SingleTurnSessions / TotalSessions in [0,1].
// Returns 0 when there are no sessions (avoids divide-by-zero).
func (s CacheUsageStats) SingleTurnRatio() float64 {
	if s.TotalSessions == 0 {
		return 0
	}
	return float64(s.SingleTurnSessions) / float64(s.TotalSessions)
}

// CacheUsageLogPath returns the absolute path to the cache-usage JSONL file
// under projectRoot/.moai/state/. The path is computed with filepath.Join so
// it is correct on every platform.
//
// @MX:ANCHOR: [AUTO] CacheUsageLogPath — sole resolver of the cache-usage JSONL location
// @MX:REASON: fan_in >= 3 expected — the PostToolUse hook writer, the moai doctor aggregator, and the M3/M4 tests all derive the JSONL path through this one helper; a divergent path would split telemetry between writer and reader and silently zero out the hit rate.
func CacheUsageLogPath(projectRoot string) string {
	return filepath.Join(projectRoot, defs.MoAIDir, defs.StateSubdir, CacheUsageFileName)
}

// AppendCacheUsage appends one CacheUsageEntry as a single JSONL line to the
// cache-usage log under projectRoot. The .moai/state/ directory is created if
// absent. When the entry's Timestamp is empty, it is populated with the current
// UTC time in RFC3339.
//
// The write is append-only (os.O_APPEND); concurrent appends from separate
// processes are serialized by the OS at the syscall level for the small
// single-line payloads this writer produces.
func AppendCacheUsage(projectRoot string, entry CacheUsageEntry) error {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	path := CacheUsageLogPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), defs.DirPerm); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal cache usage entry: %w", err)
	}
	line = append(line, '\n')

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, defs.FilePerm)
	if err != nil {
		return fmt.Errorf("open cache usage log: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("write cache usage entry: %w", err)
	}
	return nil
}

// ReadCacheUsage reads all entries from the cache-usage JSONL log under
// projectRoot. A missing file returns an empty slice and a nil error (a fresh
// project has no telemetry yet). Malformed lines are skipped silently so a
// single corrupt row never blocks aggregation.
func ReadCacheUsage(projectRoot string) ([]CacheUsageEntry, error) {
	path := CacheUsageLogPath(projectRoot)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("open cache usage log: %w", err)
	}
	defer func() { _ = f.Close() }()

	return decodeCacheUsage(f)
}

// decodeCacheUsage decodes JSONL rows from r, skipping malformed lines.
func decodeCacheUsage(r io.Reader) ([]CacheUsageEntry, error) {
	var entries []CacheUsageEntry
	scanner := bufio.NewScanner(r)
	// Raise the line buffer ceiling: a single entry is small, but long-running
	// projects may accumulate wide rows; 1 MiB is a comfortable upper bound.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		raw := scanner.Bytes()
		if len(raw) == 0 {
			continue
		}
		var e CacheUsageEntry
		if err := json.Unmarshal(raw, &e); err != nil {
			// Skip malformed rows — telemetry must degrade gracefully.
			continue
		}
		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("scan cache usage log: %w", err)
	}
	return entries, nil
}

// AggregateCacheUsage computes window statistics over the given entries. Only
// entries whose Timestamp is within [now-window, now] are counted; entries with
// an unparseable Timestamp are treated as in-window (conservative — they are not
// silently dropped). A non-positive window includes every entry.
//
// @MX:ANCHOR: [AUTO] AggregateCacheUsage — sole hit-rate + single-turn-ratio aggregator over the JSONL window
// @MX:REASON: fan_in >= 3 expected — the moai doctor cache metric, AC-PC-007 doctor test, and the M4 aggregation unit test all depend on this single computation; the hit-rate formula (read/(read+creation)) and the single-turn-session definition (max turn per session == 1) are contractual KPI definitions (K1/K5) that must not diverge between writer-time and doctor-time.
func AggregateCacheUsage(entries []CacheUsageEntry, now time.Time, window time.Duration) CacheUsageStats {
	var stats CacheUsageStats
	// maxTurnBySession tracks the highest turn index seen per session, used to
	// classify single-turn sessions (max turn == 1) for the K5 penalty ratio.
	maxTurnBySession := make(map[string]int)

	cutoff := now.Add(-window)
	for _, e := range entries {
		if window > 0 && !inWindow(e.Timestamp, cutoff, now) {
			continue
		}
		stats.EntryCount++
		stats.TotalCacheCreation += e.CacheCreation
		stats.TotalCacheRead += e.CacheRead
		if cur, ok := maxTurnBySession[e.SessionID]; !ok || e.Turn > cur {
			maxTurnBySession[e.SessionID] = e.Turn
		}
	}

	stats.TotalSessions = len(maxTurnBySession)
	for _, maxTurn := range maxTurnBySession {
		if maxTurn <= 1 {
			stats.SingleTurnSessions++
		}
	}

	denom := stats.TotalCacheRead + stats.TotalCacheCreation
	if denom > 0 {
		stats.HitRate = float64(stats.TotalCacheRead) / float64(denom)
	}
	return stats
}

// inWindow reports whether the RFC3339 timestamp ts falls within [cutoff, now].
// An unparseable timestamp returns true (conservative inclusion).
func inWindow(ts string, cutoff, now time.Time) bool {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return true
	}
	return !t.Before(cutoff) && !t.After(now)
}
