package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

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

// TestCheckCacheHitRate_EnabledShowsRate (AC-PC-007) verifies that when
// cacheStrategy.enabled == true and a 7-day JSONL window exists, the check emits
// a message matching "Cache hit rate (last 7 days): NN%".
func TestCheckCacheHitRate_EnabledShowsRate(t *testing.T) {
	root := t.TempDir()
	writeCacheYAML(t, root, true)
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

// TestCheckCacheHitRate_DisabledNoRate (AC-PC-007 §3) verifies that when
// cacheStrategy.enabled == false, the check does NOT report a hit-rate line
// (absence is correct).
func TestCheckCacheHitRate_DisabledNoRate(t *testing.T) {
	root := t.TempDir()
	writeCacheYAML(t, root, false)
	seedCacheUsage(t, root, 0)

	check := checkCacheHitRate(root, false)
	if strings.Contains(check.Message, "Cache hit rate") {
		t.Errorf("disabled cache must NOT report hit rate; got: %q", check.Message)
	}
}

// TestCheckCacheHitRate_SingleTurnWarning (K5) verifies the WARN when the
// single-turn session ratio exceeds 10% — message must recommend
// `session_ttl: "off"`.
func TestCheckCacheHitRate_SingleTurnWarning(t *testing.T) {
	root := t.TempDir()
	writeCacheYAML(t, root, true)
	// 1 multi-turn session + 9 single-turn sessions → 9/10 = 90% > 10%.
	seedCacheUsage(t, root, 9)

	check := checkCacheHitRate(root, true)
	if check.Status != CheckWarn {
		t.Errorf("status = %q, want warn (single-turn ratio > 10%%)", check.Status)
	}
	if !strings.Contains(check.Detail, `session_ttl: "off"`) {
		t.Errorf("warning detail must recommend session_ttl: \"off\"; got: %q", check.Detail)
	}
}

// TestCheckCacheHitRate_NoConfigNoRate verifies a project without cache.yaml
// (cache disabled by safe default) does not report a hit-rate line.
func TestCheckCacheHitRate_NoConfigNoRate(t *testing.T) {
	root := t.TempDir()
	check := checkCacheHitRate(root, false)
	if strings.Contains(check.Message, "Cache hit rate") {
		t.Errorf("no cache.yaml must NOT report hit rate; got: %q", check.Message)
	}
}

// TestCheckCacheHitRate_EnabledNoTelemetry verifies that when caching is enabled
// but the JSONL window has no entries, the check reports n/a (not a percentage).
func TestCheckCacheHitRate_EnabledNoTelemetry(t *testing.T) {
	root := t.TempDir()
	writeCacheYAML(t, root, true)
	// No cache-usage.jsonl written → empty window.
	check := checkCacheHitRate(root, false)
	if check.Status != CheckOK {
		t.Errorf("status = %q, want ok (enabled, no telemetry)", check.Status)
	}
	if !strings.Contains(check.Message, "n/a") {
		t.Errorf("empty-window message should be n/a; got: %q", check.Message)
	}
}

// TestCheckCacheHitRate_VerboseDetail verifies the verbose path emits a token
// breakdown detail for the healthy (OK) case.
func TestCheckCacheHitRate_VerboseDetail(t *testing.T) {
	root := t.TempDir()
	writeCacheYAML(t, root, true)
	seedCacheUsage(t, root, 0)
	check := checkCacheHitRate(root, true)
	if check.Status != CheckOK {
		t.Fatalf("status = %q, want ok", check.Status)
	}
	if !strings.Contains(check.Detail, "reads") || !strings.Contains(check.Detail, "creation") {
		t.Errorf("verbose detail should include token breakdown; got: %q", check.Detail)
	}
}

// TestCheckCacheHitRate_DisabledVerboseHint verifies the disabled+verbose path
// surfaces the enablement hint detail.
func TestCheckCacheHitRate_DisabledVerboseHint(t *testing.T) {
	root := t.TempDir()
	writeCacheYAML(t, root, false)
	check := checkCacheHitRate(root, true)
	if !strings.Contains(check.Detail, "cacheStrategy.enabled: true") {
		t.Errorf("disabled+verbose should hint at enabling; got detail: %q", check.Detail)
	}
}

// TestCheckCacheHitRate_TelemetryReadError verifies that an unreadable telemetry
// log (here: the JSONL path is a directory) degrades to OK with a "no telemetry
// yet" message rather than failing the check.
func TestCheckCacheHitRate_TelemetryReadError(t *testing.T) {
	root := t.TempDir()
	writeCacheYAML(t, root, true)
	// Create the JSONL path as a DIRECTORY so os.Open succeeds but reads error.
	jsonlPath := filepath.Join(root, ".moai", "state", "cache-usage.jsonl")
	if err := os.MkdirAll(jsonlPath, 0o755); err != nil {
		t.Fatalf("mkdir jsonl-as-dir: %v", err)
	}
	check := checkCacheHitRate(root, true)
	if check.Status != CheckOK {
		t.Errorf("status = %q, want ok (read error degrades gracefully)", check.Status)
	}
	if strings.Contains(check.Message, "Cache hit rate (last 7 days):") &&
		!strings.Contains(check.Message, "n/a") {
		t.Errorf("read-error path must not report a hit-rate percentage; got: %q", check.Message)
	}
}
