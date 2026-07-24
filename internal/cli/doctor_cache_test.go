package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/state"
)

// writeCacheYAML writes a cache.yaml with the given enabled flag under root.
func writeCacheYAML(t *testing.T, root string, enabled bool) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir config sections: %v", err)
	}
	body := "cacheStrategy:\n  enabled: " + boolStr(enabled) + "\n  session_ttl: \"1h\"\n  spec_ttl: \"5m\"\n  min_cacheable_tokens: 2048\n"
	if err := os.WriteFile(filepath.Join(dir, "cache.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write cache.yaml: %v", err)
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// seedCacheUsage writes a synthetic 7-day-window JSONL fixture under root.
// 80% hit rate: 2-turn session (creation 1000, read 9000 → hit 9000/10000=0.9),
// plus a single-turn session that drives the K5 ratio.
func seedCacheUsage(t *testing.T, root string, singleTurnSessions int) {
	t.Helper()
	now := time.Now().UTC()
	mk := func(sess string, turn, creation, read int) state.CacheUsageEntry {
		return state.CacheUsageEntry{
			Timestamp:     now.Add(-1 * time.Hour).Format(time.RFC3339),
			SessionID:     sess,
			Turn:          turn,
			CacheCreation: creation,
			CacheRead:     read,
			Model:         "claude-sonnet-4-6",
		}
	}
	// Multi-turn session with a high hit rate.
	if err := state.AppendCacheUsage(root, mk("multi", 1, 2000, 0)); err != nil {
		t.Fatalf("seed multi turn1: %v", err)
	}
	if err := state.AppendCacheUsage(root, mk("multi", 2, 0, 8000)); err != nil {
		t.Fatalf("seed multi turn2: %v", err)
	}
	// Single-turn sessions for the K5 ratio.
	for i := 0; i < singleTurnSessions; i++ {
		sess := "single-" + boolStr(i%2 == 0) + string(rune('a'+i))
		if err := state.AppendCacheUsage(root, mk(sess, 1, 500, 0)); err != nil {
			t.Fatalf("seed single %d: %v", i, err)
		}
	}
}

// TestCheckCacheHitRate_ShowsRate verifies that a populated 7-day JSONL
// window produces a message matching "Cache hit rate (last 7 days): NN%".
func TestCheckCacheHitRate_ShowsRate(t *testing.T) {
	root := t.TempDir()
	seedCacheUsage(t, root, 0)

	check := checkCacheHitRate(root, false)

	re := regexp.MustCompile(`Cache hit rate \(last 7 days\): [0-9]+%`)
	if !re.MatchString(check.Message) {
		t.Errorf("message does not match hit-rate pattern; got: %q", check.Message)
	}
	// 8000 / (8000 + 2000) = 80%.
	if !strings.Contains(check.Message, "80%") {
		t.Errorf("expected 80%% hit rate in message; got: %q", check.Message)
	}
}

// TestCheckCacheHitRate_FlagDoesNotSuppressRate guards the correction: the
// metric reports PostToolUse telemetry regardless of cacheStrategy.enabled.
// The flag once gated a cache_control injector that no production code ever
// called, so gating the metric on it hid real data behind a no-op toggle.
func TestCheckCacheHitRate_FlagDoesNotSuppressRate(t *testing.T) {
	root := t.TempDir()
	writeCacheYAML(t, root, false)
	seedCacheUsage(t, root, 0)

	check := checkCacheHitRate(root, false)
	if !strings.Contains(check.Message, "Cache hit rate") {
		t.Errorf("cacheStrategy.enabled: false must NOT suppress the hit rate; got: %q", check.Message)
	}
	if !strings.Contains(check.Message, "80%") {
		t.Errorf("expected the seeded 80%% hit rate; got: %q", check.Message)
	}
}

// TestCheckCacheHitRate_NoConfigStillReports verifies a project without
// cache.yaml still reports telemetry — the config file is not a precondition.
func TestCheckCacheHitRate_NoConfigStillReports(t *testing.T) {
	root := t.TempDir()
	seedCacheUsage(t, root, 0)

	check := checkCacheHitRate(root, false)
	if !strings.Contains(check.Message, "Cache hit rate") {
		t.Errorf("missing cache.yaml must NOT suppress the hit rate; got: %q", check.Message)
	}
}

// TestCheckCacheHitRate_SingleTurnWarning verifies the WARN raised when the
// single-turn session ratio exceeds 10%. The detail names the cache-write cost
// rather than recommending a session_ttl setting: TTLs are chosen by Claude
// Code, so no moai-side config change can alter them.
func TestCheckCacheHitRate_SingleTurnWarning(t *testing.T) {
	root := t.TempDir()
	// 1 multi-turn session + 9 single-turn sessions → 9/10 = 90% > 10%.
	seedCacheUsage(t, root, 9)

	check := checkCacheHitRate(root, true)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %q, want warn (single-turn ratio > 10%%)", check.Status)
	}
	if !strings.Contains(check.Detail, "single-turn") {
		t.Errorf("warning detail must name the single-turn ratio; got: %q", check.Detail)
	}
	if strings.Contains(check.Detail, "session_ttl") {
		t.Errorf("warning detail must not recommend a no-op session_ttl change; got: %q", check.Detail)
	}
}

// TestCheckCacheHitRate_NoTelemetryReportsNA verifies that an empty JSONL
// window reports n/a rather than a percentage.
func TestCheckCacheHitRate_NoTelemetryReportsNA(t *testing.T) {
	root := t.TempDir()
	// No cache-usage.jsonl written → empty window.
	check := checkCacheHitRate(root, false)
	if check.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok (no telemetry)", check.Status)
	}
	if !strings.Contains(check.Message, "n/a") {
		t.Errorf("empty-window message should be n/a; got: %q", check.Message)
	}
}

// TestCheckCacheHitRate_VerboseDetail verifies the verbose path emits a token
// breakdown detail for the healthy (OK) case.
func TestCheckCacheHitRate_VerboseDetail(t *testing.T) {
	root := t.TempDir()
	seedCacheUsage(t, root, 0)
	check := checkCacheHitRate(root, true)
	if check.Status != uikit.CheckOK {
		t.Fatalf("status = %q, want ok", check.Status)
	}
	if !strings.Contains(check.Detail, "reads") || !strings.Contains(check.Detail, "creation") {
		t.Errorf("verbose detail should include token breakdown; got: %q", check.Detail)
	}
}

// TestCheckCacheHitRate_TelemetryReadError verifies that an unreadable telemetry
// log (here: the JSONL path is a directory) degrades to OK with a "no telemetry
// yet" message rather than failing the check.
func TestCheckCacheHitRate_TelemetryReadError(t *testing.T) {
	root := t.TempDir()
	// Create the JSONL path as a DIRECTORY so os.Open succeeds but reads error.
	jsonlPath := filepath.Join(root, ".moai", "state", "cache-usage.jsonl")
	if err := os.MkdirAll(jsonlPath, 0o755); err != nil {
		t.Fatalf("mkdir jsonl-as-dir: %v", err)
	}
	check := checkCacheHitRate(root, true)
	if check.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok (read error degrades gracefully)", check.Status)
	}
	if strings.Contains(check.Message, "Cache hit rate (last 7 days):") &&
		!strings.Contains(check.Message, "n/a") {
		t.Errorf("read-error path must not report a hit-rate percentage; got: %q", check.Message)
	}
}
