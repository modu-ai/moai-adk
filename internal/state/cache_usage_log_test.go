package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestAppendCacheUsage_CreatesStateDirAndWritesBothKeys verifies that
// AppendCacheUsage creates .moai/state/ and that the persisted JSONL line
// carries BOTH the cache_creation_input_tokens and cache_read_input_tokens
// keys (the AC-PC-005 contract — exercised end-to-end via the hook in
// internal/hook, validated at the writer level here).
func TestAppendCacheUsage_CreatesStateDirAndWritesBothKeys(t *testing.T) {
	root := t.TempDir()

	entry := CacheUsageEntry{
		SessionID:     "sess-1",
		Turn:          1,
		CacheCreation: 12450,
		CacheRead:     0,
		Model:         "claude-sonnet-4-6",
	}
	if err := AppendCacheUsage(root, entry); err != nil {
		t.Fatalf("AppendCacheUsage: %v", err)
	}

	path := CacheUsageLogPath(root)
	if path != filepath.Join(root, ".moai", "state", "cache-usage.jsonl") {
		t.Errorf("CacheUsageLogPath = %q, unexpected layout", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read jsonl: %v", err)
	}
	content := string(data)
	for _, key := range []string{"cache_creation_input_tokens", "cache_read_input_tokens", "timestamp", "session_id", "turn", "model"} {
		if !strings.Contains(content, key) {
			t.Errorf("persisted JSONL missing key %q; got: %s", key, content)
		}
	}
	if !strings.HasSuffix(content, "\n") {
		t.Errorf("JSONL line must end with newline; got: %q", content)
	}
}

// TestAppendCacheUsage_PopulatesTimestampWhenEmpty verifies the writer fills an
// RFC3339 UTC timestamp when the entry omits one.
func TestAppendCacheUsage_PopulatesTimestampWhenEmpty(t *testing.T) {
	root := t.TempDir()
	if err := AppendCacheUsage(root, CacheUsageEntry{SessionID: "s", Turn: 1}); err != nil {
		t.Fatalf("AppendCacheUsage: %v", err)
	}
	entries, err := ReadCacheUsage(root)
	if err != nil {
		t.Fatalf("ReadCacheUsage: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}
	if _, perr := time.Parse(time.RFC3339, entries[0].Timestamp); perr != nil {
		t.Errorf("timestamp %q not RFC3339: %v", entries[0].Timestamp, perr)
	}
}

// TestReadCacheUsage_MissingFileReturnsEmpty verifies a fresh project (no log)
// yields an empty slice and nil error.
func TestReadCacheUsage_MissingFileReturnsEmpty(t *testing.T) {
	root := t.TempDir()
	entries, err := ReadCacheUsage(root)
	if err != nil {
		t.Fatalf("ReadCacheUsage on missing file: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("want 0 entries on missing file, got %d", len(entries))
	}
}

// TestReadCacheUsage_SkipsMalformedLines verifies a corrupt row does not abort
// reading of the surrounding valid rows.
func TestReadCacheUsage_SkipsMalformedLines(t *testing.T) {
	root := t.TempDir()
	if err := AppendCacheUsage(root, CacheUsageEntry{SessionID: "s", Turn: 1, CacheCreation: 100}); err != nil {
		t.Fatalf("seed append: %v", err)
	}
	// Inject a malformed line directly.
	path := CacheUsageLogPath(root)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open for malformed inject: %v", err)
	}
	if _, werr := f.WriteString("{ this is not json\n"); werr != nil {
		t.Fatalf("write malformed: %v", werr)
	}
	_ = f.Close()
	if err := AppendCacheUsage(root, CacheUsageEntry{SessionID: "s", Turn: 2, CacheRead: 900}); err != nil {
		t.Fatalf("trailing append: %v", err)
	}

	entries, err := ReadCacheUsage(root)
	if err != nil {
		t.Fatalf("ReadCacheUsage: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("want 2 valid entries (malformed skipped), got %d", len(entries))
	}
}

// TestAggregateCacheUsage_HitRateAndSingleTurn verifies the K1 hit-rate formula
// and the K5 single-turn-session classification over a synthetic window.
func TestAggregateCacheUsage_HitRateAndSingleTurn(t *testing.T) {
	now := time.Date(2026, 5, 30, 12, 0, 0, 0, time.UTC)
	mk := func(sess string, turn, creation, read int, ageHours int) CacheUsageEntry {
		return CacheUsageEntry{
			Timestamp:     now.Add(-time.Duration(ageHours) * time.Hour).Format(time.RFC3339),
			SessionID:     sess,
			Turn:          turn,
			CacheCreation: creation,
			CacheRead:     read,
			Model:         "claude-sonnet-4-6",
		}
	}
	entries := []CacheUsageEntry{
		// multi-turn session A: turn1 creation, turn2 read
		mk("A", 1, 1000, 0, 1),
		mk("A", 2, 0, 9000, 1),
		// single-turn session B: turn1 only
		mk("B", 1, 1000, 0, 2),
		// out-of-window entry (10 days old) — excluded from a 7-day window
		mk("C", 1, 5000, 0, 24*10),
	}

	stats := AggregateCacheUsage(entries, now, 7*24*time.Hour)

	if stats.EntryCount != 3 {
		t.Errorf("EntryCount = %d, want 3 (out-of-window C excluded)", stats.EntryCount)
	}
	if stats.TotalCacheCreation != 2000 {
		t.Errorf("TotalCacheCreation = %d, want 2000", stats.TotalCacheCreation)
	}
	if stats.TotalCacheRead != 9000 {
		t.Errorf("TotalCacheRead = %d, want 9000", stats.TotalCacheRead)
	}
	// hit rate = 9000 / (9000 + 2000) = 0.8181...
	wantHit := 9000.0 / 11000.0
	if diff := stats.HitRate - wantHit; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("HitRate = %v, want %v", stats.HitRate, wantHit)
	}
	if stats.TotalSessions != 2 {
		t.Errorf("TotalSessions = %d, want 2 (A,B in window)", stats.TotalSessions)
	}
	if stats.SingleTurnSessions != 1 {
		t.Errorf("SingleTurnSessions = %d, want 1 (B only)", stats.SingleTurnSessions)
	}
	if r := stats.SingleTurnRatio(); r != 0.5 {
		t.Errorf("SingleTurnRatio = %v, want 0.5", r)
	}
}

// TestAggregateCacheUsage_ZeroDenominatorHitRate verifies hit rate is 0 (not
// NaN) when no tokens were recorded.
func TestAggregateCacheUsage_ZeroDenominatorHitRate(t *testing.T) {
	stats := AggregateCacheUsage(nil, time.Now(), 7*24*time.Hour)
	if stats.HitRate != 0 {
		t.Errorf("HitRate on empty = %v, want 0", stats.HitRate)
	}
	if stats.SingleTurnRatio() != 0 {
		t.Errorf("SingleTurnRatio on empty = %v, want 0", stats.SingleTurnRatio())
	}
}

// TestAggregateCacheUsage_NonPositiveWindowIncludesAll verifies a window <= 0
// includes every entry regardless of timestamp.
func TestAggregateCacheUsage_NonPositiveWindowIncludesAll(t *testing.T) {
	now := time.Now()
	old := now.Add(-365 * 24 * time.Hour).Format(time.RFC3339)
	entries := []CacheUsageEntry{
		{Timestamp: old, SessionID: "x", Turn: 1, CacheRead: 50, CacheCreation: 50},
	}
	stats := AggregateCacheUsage(entries, now, 0)
	if stats.EntryCount != 1 {
		t.Errorf("EntryCount = %d, want 1 (window<=0 includes all)", stats.EntryCount)
	}
}
